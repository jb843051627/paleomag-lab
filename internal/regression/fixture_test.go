package regression

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
	"github.com/jb843051627/paleomag-lab/internal/service"
	"github.com/jb843051627/paleomag-lab/internal/store"
)

type fixture struct {
	t        *testing.T
	ctx      context.Context
	db       *store.DB
	clock    *clock.Fixed
	services *service.Services
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "paleomag-test.db"))
	if err != nil {
		t.Fatal(err)
	}
	fixed := clock.NewFixed(time.Date(2026, 1, 15, 8, 0, 0, 0, time.UTC))
	f := &fixture{t: t, ctx: context.Background(), db: db, clock: fixed, services: service.NewAll(repository.NewAll(db), fixed)}
	t.Cleanup(func() { _ = db.Close() })
	return f
}

func (f *fixture) specimen(code string) model.Specimen {
	f.t.Helper()
	item, err := f.services.Specimens.Create(f.ctx, service.CreateSpecimenInput{Code: code, Lithology: "basalt", Site: "ridge-7", Collector: "operator-a"})
	if err != nil {
		f.t.Fatal(err)
	}
	item, err = f.services.Specimens.Orient(f.ctx, item.ID, model.Orientation{Declination: 25, Inclination: 12, Coordinate: "geographic", Source: "sun-compass"}, "operator-a")
	if err != nil {
		f.t.Fatal(err)
	}
	return item
}

func (f *fixture) instrument(serial string) model.Instrument {
	f.t.Helper()
	item, err := f.services.Instruments.Register(f.ctx, service.RegisterInstrumentInput{Name: "VectorMag", Serial: serial, Kind: "spinner", MinField: 0, MaxField: 1000})
	if err != nil {
		f.t.Fatal(err)
	}
	return item
}

func (f *fixture) calibration(instrumentID string) model.CalibrationRun {
	f.t.Helper()
	item, err := f.services.Instruments.BeginCalibration(f.ctx, instrumentID, "reference-cube", "operator-a")
	if err != nil {
		f.t.Fatal(err)
	}
	item, err = f.services.Instruments.CompleteCalibration(f.ctx, item.ID, true, [3]float64{0.01, -0.01, 0.02}, "operator-a")
	if err != nil {
		f.t.Fatal(err)
	}
	return item
}

func (f *fixture) plan(specimenID, name string, fields ...float64) model.DemagPlan {
	f.t.Helper()
	item, err := f.services.Plans.Create(f.ctx, service.CreatePlanInput{SpecimenID: specimenID, Name: name, Method: "alternating-field", Operator: "operator-a"})
	if err != nil {
		f.t.Fatal(err)
	}
	for _, field := range fields {
		item, err = f.services.Plans.AddStep(f.ctx, item.ID, field, 20, "planned field", "operator-a")
		if err != nil {
			f.t.Fatal(err)
		}
	}
	item, err = f.services.Plans.Prepare(f.ctx, item.ID, "operator-a")
	if err != nil {
		f.t.Fatal(err)
	}
	return item
}

func (f *fixture) readyRun(code string) (model.Specimen, model.Instrument, model.CalibrationRun, model.DemagPlan, model.MeasurementRun) {
	f.t.Helper()
	specimen := f.specimen(code)
	instrument := f.instrument("serial-" + code)
	calibration := f.calibration(instrument.ID)
	plan := f.plan(specimen.ID, "pilot demag", 0, 10, 20)
	run, err := f.services.Measurements.StartRun(f.ctx, service.StartRunInput{SpecimenID: specimen.ID, PlanID: plan.ID, InstrumentID: instrument.ID, CalibrationID: calibration.ID}, "operator-a")
	if err != nil {
		f.t.Fatal(err)
	}
	return specimen, instrument, calibration, plan, run
}

func (f *fixture) record(runID string, step int, field, x, y, z, intensity float64) model.Measurement {
	f.t.Helper()
	item := model.Measurement{ID: "measurement-" + runID + "-" + string(rune('a'+step)), RunID: runID, Step: step, Field: field, X: x, Y: y, Z: z, Intensity: intensity, Quality: model.QualityPending, MeasuredAt: f.clock.Now()}
	if err := f.services.Measurements.Record(f.ctx, item, "operator-a"); err != nil {
		f.t.Fatal(err)
	}
	return item
}

func (f *fixture) approvedInterpretation(code string, acceptRun bool) (model.Specimen, model.MeasurementRun, model.Interpretation) {
	f.t.Helper()
	specimen, _, _, _, run := f.readyRun(code)
	f.record(run.ID, 1, 0, 1, 2, 3, 3.7)
	f.record(run.ID, 2, 10, 1.1, 2.1, 3.1, 3.8)
	if _, _, err := f.services.Measurements.CompleteRun(f.ctx, run.ID, "operator-a"); err != nil {
		f.t.Fatal(err)
	}
	if acceptRun {
		accepted, err := f.services.Measurements.AcceptRun(f.ctx, run.ID, "operator-a")
		if err != nil {
			f.t.Fatal(err)
		}
		run = accepted
	}
	interpretation, err := f.services.Interpretations.Generate(f.ctx, run.ID, "operator-a")
	if err != nil {
		f.t.Fatal(err)
	}
	if _, err := f.services.Interpretations.Submit(f.ctx, interpretation.ID, "operator-a"); err != nil {
		f.t.Fatal(err)
	}
	if _, interpretation, err = f.services.Reviews.Decide(f.ctx, interpretation.ID, "reviewer-a", model.ReviewApprove, "direction is consistent"); err != nil {
		f.t.Fatal(err)
	}
	return specimen, run, interpretation
}
