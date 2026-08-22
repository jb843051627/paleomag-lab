package regression

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jb843051627/paleomag-lab/internal/handler"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
	"github.com/jb843051627/paleomag-lab/internal/service"
	"github.com/jb843051627/paleomag-lab/internal/store"
	"github.com/jb843051627/paleomag-lab/internal/worker"
)

func TestBug01_CalibrationLookupPreservesNotFound(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-001")
	instrument := f.instrument("SER-001")
	plan := f.plan(specimen.ID, "missing calibration", 0, 10)
	_, err := f.services.Measurements.StartRun(f.ctx, service.StartRunInput{SpecimenID: specimen.ID, PlanID: plan.ID, InstrumentID: instrument.ID, CalibrationID: "calibration-missing"}, "operator-a")
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected not-found chain, got %v", err)
	}
}

func TestBug02_EmptyPlanDoesNotPanicDuringRunStart(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-002")
	instrument := f.instrument("SER-002")
	calibration := f.calibration(instrument.ID)
	plan, err := f.services.Plans.Create(f.ctx, service.CreatePlanInput{SpecimenID: specimen.ID, Name: "empty protocol", Method: "thermal", Operator: "operator-a"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.SQL.Exec(`UPDATE demag_plans SET status=? WHERE id=?`, model.PlanReady, plan.ID); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("run start panicked: %v\n%s", recovered, debug.Stack())
		}
	}()
	_, err = f.services.Measurements.StartRun(f.ctx, service.StartRunInput{SpecimenID: specimen.ID, PlanID: plan.ID, InstrumentID: instrument.ID, CalibrationID: calibration.ID}, "operator-a")
	if !errors.Is(err, model.ErrInvalid) && !errors.Is(err, model.ErrState) {
		t.Fatalf("expected a domain error, got %v", err)
	}
}

func TestBug03_NextStepDoesNotExposePlanSlice(t *testing.T) {
	now := time.Now().UTC()
	plan := model.NewDemagPlan("plan-003", "specimen-003", "sequence", "alternating-field", "operator-a", now)
	if err := plan.AddStep(model.DemagStep{ID: "step-003", Field: 10, Temperature: 20}, now); err != nil {
		t.Fatal(err)
	}
	next := plan.NextStep()
	if next == nil {
		t.Fatal("expected a pending step")
	}
	next.Field = 999
	if plan.Steps[0].Field != 10 {
		t.Fatalf("plan step changed through returned pointer: %.2f", plan.Steps[0].Field)
	}
}

type cancelAfterOneContext struct {
	context.Context
	calls atomic.Int32
}

func (c *cancelAfterOneContext) Err() error {
	if c.calls.Add(1) > 1 {
		return context.Canceled
	}
	return nil
}

func TestBug04_QualityEvaluationHonorsCancellation(t *testing.T) {
	f := newFixture(t)
	_, _, _, _, run := f.readyRun("PM-004")
	f.record(run.ID, 1, 0, 1, 2, 3, 3.7)
	f.record(run.ID, 2, 10, 1.1, 2.1, 3.1, 3.8)
	ctx := &cancelAfterOneContext{Context: context.Background()}
	_, err := f.services.Measurements.EvaluateQuality(ctx, run.ID)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	items, err := f.services.Measurements.List(f.ctx, model.MeasurementFilter{RunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if item.Quality != model.QualityPending {
			t.Fatalf("cancelled evaluation changed quality to %s", item.Quality)
		}
	}
}

func TestBug05_TransactionFailureRollsBackSpecimen(t *testing.T) {
	f := newFixture(t)
	expected := errors.New("abort workflow")
	err := store.WithTx(f.ctx, f.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO specimens(id, code, lithology, site, collector, status, notes, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, "specimen-005", "PM-005", "basalt", "ridge-7", "operator-a", model.SpecimenRegistered, "", time.Now().UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano))
		if err != nil {
			return err
		}
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected original transaction error, got %v", err)
	}
	if _, err := f.services.Specimens.Get(f.ctx, "specimen-005"); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("rolled back specimen is still visible: %v", err)
	}
}

func TestBug06_ConcurrentWorkerStartIsIdempotent(t *testing.T) {
	f := newFixture(t)
	queues := worker.NewQueues(4)
	manager := worker.NewManager(f.services, queues)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			manager.Start(ctx)
		}()
	}
	group.Wait()
	manager.Stop()
}

func TestBug07_FailedExportIsPersistedAsFailed(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-007")
	job, err := f.services.Exports.Request(f.ctx, specimen.ID, "json", "operator-a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.services.Exports.Run(f.ctx, job.ID)
	if !errors.Is(err, model.ErrState) {
		t.Fatalf("expected report state error, got %v", err)
	}
	stored, err := f.services.Exports.Get(f.ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != model.ExportFailed || stored.Error == "" {
		t.Fatalf("failed export was not persisted: %+v", stored)
	}
}

func TestBug08_ReportWithoutApprovalDoesNotPanic(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-008")
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("report generation panicked: %v\n%s", recovered, debug.Stack())
		}
	}()
	_, err := f.services.Exports.BuildReport(f.ctx, specimen.ID, "json")
	if !errors.Is(err, model.ErrState) {
		t.Fatalf("expected approval state error, got %v", err)
	}
}

func TestBug09_TimelineSortDoesNotMutateCaller(t *testing.T) {
	f := newFixture(t)
	entries := []service.TimelineEntry{{EntityID: "a", At: "2026-01-03"}, {EntityID: "b", At: "2026-01-01"}, {EntityID: "c", At: "2026-01-02"}}
	got := f.services.Insights.SortTimeline(entries)
	if got[0].EntityID != "b" || entries[0].EntityID != "a" {
		t.Fatalf("timeline sorting changed caller data: got=%v original=%v", got, entries)
	}
}

func TestBug10_WorkerStopReturnsWhenQueuesAreIdle(t *testing.T) {
	f := newFixture(t)
	manager := worker.NewManager(f.services, worker.NewQueues(2))
	ctx, cancel := context.WithCancel(context.Background())
	manager.Start(ctx)
	done := make(chan struct{})
	go func() { manager.Stop(); close(done) }()
	select {
	case <-done:
		cancel()
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatalf("worker stop did not return\n%s", debug.Stack())
	}
}

func TestBug11_ReportSnapshotPropagatesCancellation(t *testing.T) {
	f := newFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := f.services.Reports.Snapshot(ctx, "missing-specimen")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestBug12_ConcurrentSpecimenCreationKeepsIDsUnique(t *testing.T) {
	f := newFixture(t)
	ids := make(chan string, 8)
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			item, err := f.services.Specimens.Create(f.ctx, service.CreateSpecimenInput{Code: "PM-012-" + string(rune('A'+index)), Lithology: "basalt", Site: "ridge-7", Collector: "operator-a"})
			if err != nil {
				t.Errorf("create specimen %d: %v", index, err)
				return
			}
			ids <- item.ID
		}(i)
	}
	group.Wait()
	close(ids)
	seen := map[string]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate specimen id %s", id)
		}
		seen[id] = true
	}
	if len(seen) != 8 {
		t.Fatalf("expected eight specimens, got %d", len(seen))
	}
}

func TestBug13_CompleteRejectedPlanDoesNotMutateSteps(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-013")
	plan := f.plan(specimen.ID, "prepared only", 0, 10)
	if _, err := f.services.Plans.Complete(f.ctx, plan.ID, "operator-a"); !errors.Is(err, model.ErrState) {
		t.Fatalf("expected state error, got %v", err)
	}
	stored, err := f.services.Plans.Get(f.ctx, plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range stored.Steps {
		if step.Status != model.StepPending {
			t.Fatalf("rejected completion changed step %d to %s", step.Sequence, step.Status)
		}
	}
}

func TestBug14_InvalidMeasurementStepDoesNotPanic(t *testing.T) {
	f := newFixture(t)
	_, _, _, _, run := f.readyRun("PM-014")
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("measurement step lookup panicked: %v\n%s", recovered, debug.Stack())
		}
	}()
	err := f.services.Measurements.Record(f.ctx, model.Measurement{ID: "measurement-014", RunID: run.ID, Step: 2, Field: 10, X: 1, Y: 2, Z: 3, Intensity: 3.7, Quality: model.QualityPending, MeasuredAt: f.clock.Now()}, "operator-a")
	if err != nil {
		t.Fatalf("valid second step was rejected: %v", err)
	}
}

func TestBug15_ExpiredCalibrationCannotStartRun(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-015")
	instrument := f.instrument("SER-015")
	calibration := f.calibration(instrument.ID)
	plan := f.plan(specimen.ID, "expired calibration", 0, 10)
	f.clock.Advance(48 * time.Hour)
	_, err := f.services.Measurements.StartRun(f.ctx, service.StartRunInput{SpecimenID: specimen.ID, PlanID: plan.ID, InstrumentID: instrument.ID, CalibrationID: calibration.ID}, "operator-a")
	if !errors.Is(err, model.ErrCalibration) {
		t.Fatalf("expected calibration error, got %v", err)
	}
}

func TestBug16_InvalidCalibrationReferenceDoesNotBusyInstrument(t *testing.T) {
	f := newFixture(t)
	instrument := f.instrument("SER-016")
	if _, err := f.services.Instruments.BeginCalibration(f.ctx, instrument.ID, "", "operator-a"); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("expected invalid reference error, got %v", err)
	}
	stored, err := f.services.Instruments.Get(f.ctx, instrument.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != model.InstrumentReady {
		t.Fatalf("instrument was left in %s after invalid request", stored.Status)
	}
}

func TestBug17_PlanCannotStartTwoRuns(t *testing.T) {
	f := newFixture(t)
	specimen, instrument, calibration, plan, _ := f.readyRun("PM-017")
	_, err := f.services.Measurements.StartRun(f.ctx, service.StartRunInput{SpecimenID: specimen.ID, PlanID: plan.ID, InstrumentID: instrument.ID, CalibrationID: calibration.ID}, "operator-a")
	if !errors.Is(err, model.ErrState) {
		t.Fatalf("expected duplicate run state error, got %v", err)
	}
}

func TestBug18_EmptyPlanCannotProduceNextField(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-018")
	plan, err := f.services.Plans.Create(f.ctx, service.CreatePlanInput{SpecimenID: specimen.ID, Name: "no fields", Method: "thermal", Operator: "operator-a"})
	if err != nil {
		t.Fatal(err)
	}
	trend := model.TrendSummary{Points: []model.TrendPoint{{Step: 1, Field: 0, Intensity: 2}}, IntensityLoss: 0.1, Stable: false}
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("empty plan recommendation panicked: %v\n%s", recovered, debug.Stack())
		}
	}()
	_, err = f.services.Analysis.RecommendNextField(f.ctx, plan.ID, trend)
	if !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("expected invalid empty plan error, got %v", err)
	}
}

func TestBug19_ProtocolDoesNotAppendToExistingDraft(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-019")
	plan, err := f.services.Plans.Create(f.ctx, service.CreatePlanInput{SpecimenID: specimen.ID, Name: "existing step", Method: "alternating-field", Operator: "operator-a"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err = f.services.Plans.AddStep(f.ctx, plan.ID, 0, 20, "existing field", "operator-a")
	if err != nil {
		t.Fatal(err)
	}
	protocol := model.DemagProtocol{Name: "thermal pilot", Method: "thermal", Temperature: 40, Fields: []float64{20, 40}, Rationale: "separate temperature ladder"}
	protocolService := f.services.Protocols
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("protocol application panicked: %v", recovered)
		}
	}()
	_, err = protocolService.ApplyToPlan(f.ctx, plan.ID, protocol, "operator-a")
	if !errors.Is(err, model.ErrConflict) {
		t.Fatalf("expected existing-step conflict, got %v", err)
	}
}

func TestBug20_ReportRequiresAcceptedRun(t *testing.T) {
	f := newFixture(t)
	specimen, _, _ := f.approvedInterpretation("PM-020", false)
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("report validation panicked: %v", recovered)
		}
	}()
	_, err := f.services.Reports.Snapshot(f.ctx, specimen.ID)
	if !errors.Is(err, model.ErrState) {
		t.Fatalf("expected accepted-run state error, got %v", err)
	}
}

func TestBug21_BadMeasurementDoesNotAnchorQualitySeries(t *testing.T) {
	policy := model.DefaultQualityPolicy()
	items := []model.Measurement{
		{ID: "bad", Step: 1, Field: 0, X: 0, Y: 0, Z: 0, Intensity: 1},
		{ID: "good", Step: 2, Field: 10, X: 1, Y: 1, Z: 1, Intensity: 1},
	}
	results := model.AssessSeries(items, policy)
	if len(results) != 2 || results[0].Flag != model.QualityBad {
		t.Fatalf("unexpected first assessment: %+v", results)
	}
	if results[1].ComparedTo != "" {
		t.Fatalf("second point compared against rejected point %q", results[1].ComparedTo)
	}
}

func TestBug22_ReportRetainsWorkflowWarnings(t *testing.T) {
	f := newFixture(t)
	specimen, _, interpretation := f.approvedInterpretation("PM-022", true)
	if interpretation.Status != model.InterpretationApproved {
		t.Fatalf("interpretation is not approved: %s", interpretation.Status)
	}
	snapshot, err := f.services.Reports.Snapshot(f.ctx, specimen.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.HasWarnings() {
		t.Fatal("expected archive warning in report snapshot")
	}
}

func TestBug23_ArchiveRequiresApprovedInterpretationState(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-023")
	if _, err := f.services.Specimens.Archive(f.ctx, specimen.ID, "operator-a"); !errors.Is(err, model.ErrState) {
		t.Fatalf("expected archive state error, got %v", err)
	}
}

func TestBug24_PreparedPlanRejectsAdditionalStep(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-024")
	plan := f.plan(specimen.ID, "prepared plan", 0, 10)
	_, err := f.services.Plans.AddStep(f.ctx, plan.ID, 20, 20, "late step", "operator-a")
	if !errors.Is(err, model.ErrState) {
		t.Fatalf("expected prepared-plan state error, got %v", err)
	}
}

func TestBug25_ExportFormatIsCaseInsensitive(t *testing.T) {
	f := newFixture(t)
	specimen, _, _ := f.approvedInterpretation("PM-025", true)
	job, err := f.services.Exports.Request(f.ctx, specimen.ID, "CSV", "operator-a")
	if err != nil {
		t.Fatal(err)
	}
	if job.Format != "csv" {
		t.Fatalf("format was not normalized: %q", job.Format)
	}
}

func TestBug26_ExportFailurePreservesDomainError(t *testing.T) {
	f := newFixture(t)
	specimen := f.specimen("PM-026")
	job, err := f.services.Exports.Request(f.ctx, specimen.ID, "json", "operator-a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.services.Exports.Run(f.ctx, job.ID)
	if !errors.Is(err, model.ErrState) {
		t.Fatalf("expected state error after export failure, got %v", err)
	}
}

func TestBug27_OrientationWritesAuditEvent(t *testing.T) {
	f := newFixture(t)
	item, err := f.services.Specimens.Create(f.ctx, service.CreateSpecimenInput{Code: "PM-027", Lithology: "basalt", Site: "ridge-7", Collector: "operator-a"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.services.Specimens.Orient(f.ctx, item.ID, model.Orientation{Declination: 40, Inclination: -10, Coordinate: "geographic", Source: "sun-compass"}, "operator-a"); err != nil {
		t.Fatal(err)
	}
	event, err := f.services.Audits.Latest(f.ctx, "specimen", item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if event.Action != "oriented" || event.After == "" {
		t.Fatalf("orientation audit is incomplete: %+v", event)
	}
}

func TestBug28_MissingSpecimenReturnsHTTPNotFound(t *testing.T) {
	f := newFixture(t)
	api := handler.New(f.services, worker.NewManager(f.services, worker.NewQueues(2)), f.db)
	request := httptest.NewRequest(http.MethodGet, "/api/specimens/missing", nil)
	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestBug29_WorkQueueReportsCapacityInsteadOfBlocking(t *testing.T) {
	f := newFixture(t)
	queues := worker.NewQueues(1)
	manager := worker.NewManager(f.services, queues)
	if err := manager.EnqueueQuality("run-029-a"); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		close(started)
		if err := manager.EnqueueQuality("run-029-b"); !errors.Is(err, model.ErrQueueFull) {
			t.Errorf("expected queue-full error, got %v", err)
		}
	}()
	<-started
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("second enqueue did not report capacity")
	}
}

func TestBug30_SpecimenSurvivesDatabaseRestart(t *testing.T) {
	f := newFixture(t)
	item := f.specimen("PM-030")
	path := f.db.Path
	if err := f.db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Open(f.ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	stored, err := service.NewAll(repository.NewAll(reopened), f.clock).Specimens.Get(f.ctx, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Code != item.Code || stored.Orientation == nil || stored.Status != model.SpecimenOriented {
		t.Fatalf("reopened specimen is incomplete: %+v", stored)
	}
}
