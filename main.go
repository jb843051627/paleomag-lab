package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/handler"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
	"github.com/jb843051627/paleomag-lab/internal/service"
	"github.com/jb843051627/paleomag-lab/internal/store"
	"github.com/jb843051627/paleomag-lab/internal/worker"
)

func main() {
	smoke := flag.Bool("smoke-test", false, "run a local workflow smoke test")
	addr := flag.String("addr", envOr("PALEOMAG_ADDR", ":8080"), "HTTP listen address")
	flag.Parse()
	if *smoke {
		if err := runSmoke(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("smoke test passed")
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dbPath := envOr("PALEOMAG_DB", filepath.Join(".", "paleomag.db"))
	db, err := store.Open(ctx, dbPath)
	if err != nil {
		fatal(err)
	}
	defer db.Close()

	repos := repository.NewAll(db)
	services := service.NewAll(repos, clock.Real{})
	queues := worker.NewQueues(32)
	workers := worker.NewManager(services, queues)
	workers.Start(ctx)
	defer workers.Stop()

	server := &http.Server{Addr: *addr, Handler: handler.New(services, workers, db).Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fatal(err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func runSmoke() error {
	tempDir, err := os.MkdirTemp("", "paleomag-smoke-")
	if err != nil {
		return fmt.Errorf("create smoke directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(tempDir, "smoke.db"))
	if err != nil {
		return err
	}
	defer db.Close()
	services := service.NewAll(repository.NewAll(db), clock.Real{})
	specimen, err := services.Specimens.Create(ctx, service.CreateSpecimenInput{Code: "SMOKE-001", Lithology: "basalt", Site: "ridge-7", Collector: "smoke"})
	if err != nil {
		return err
	}
	if _, err := services.Specimens.Orient(ctx, specimen.ID, model.Orientation{Declination: 32, Inclination: 14, Coordinate: "geographic", Source: "sun-compass"}, "smoke"); err != nil {
		return err
	}
	instrument, err := services.Instruments.Register(ctx, service.RegisterInstrumentInput{Name: "VectorMag", Serial: "VM-SMOKE", Kind: "spinner", MinField: 0, MaxField: 500})
	if err != nil {
		return err
	}
	calibration, err := services.Instruments.BeginCalibration(ctx, instrument.ID, "reference-cube", "smoke")
	if err != nil {
		return err
	}
	if _, err := services.Instruments.CompleteCalibration(ctx, calibration.ID, true, [3]float64{0.01, -0.02, 0.03}, "smoke"); err != nil {
		return err
	}
	plan, err := services.Plans.Create(ctx, service.CreatePlanInput{SpecimenID: specimen.ID, Name: "pilot demag", Method: "alternating-field", Operator: "smoke"})
	if err != nil {
		return err
	}
	if plan, err = services.Plans.AddStep(ctx, plan.ID, 0, 20, "zero field", "smoke"); err != nil {
		return err
	}
	if plan, err = services.Plans.AddStep(ctx, plan.ID, 10, 20, "first field", "smoke"); err != nil {
		return err
	}
	if _, err := services.Plans.Prepare(ctx, plan.ID, "smoke"); err != nil {
		return err
	}
	run, err := services.Measurements.StartRun(ctx, service.StartRunInput{SpecimenID: specimen.ID, PlanID: plan.ID, InstrumentID: instrument.ID, CalibrationID: calibration.ID}, "smoke")
	if err != nil {
		return err
	}
	if err := services.Measurements.Record(ctx, model.Measurement{ID: "smoke-m1", RunID: run.ID, Step: 1, Field: 0, X: 1, Y: 2, Z: 3, Intensity: 3.7, Quality: model.QualityPending, MeasuredAt: time.Now()}, "smoke"); err != nil {
		return err
	}
	if err := services.Measurements.Record(ctx, model.Measurement{ID: "smoke-m2", RunID: run.ID, Step: 2, Field: 10, X: 1.1, Y: 2.1, Z: 3.1, Intensity: 3.8, Quality: model.QualityPending, MeasuredAt: time.Now()}, "smoke"); err != nil {
		return err
	}
	if _, _, err := services.Measurements.CompleteRun(ctx, run.ID, "smoke"); err != nil {
		return err
	}
	if _, err := services.Measurements.AcceptRun(ctx, run.ID, "smoke"); err != nil {
		return err
	}
	interpretation, err := services.Interpretations.Generate(ctx, run.ID, "smoke")
	if err != nil {
		return err
	}
	if _, err := services.Interpretations.Submit(ctx, interpretation.ID, "smoke"); err != nil {
		return err
	}
	if _, _, err := services.Reviews.Decide(ctx, interpretation.ID, "reviewer", model.ReviewApprove, "accepted by smoke test"); err != nil {
		return err
	}
	job, err := services.Exports.Request(ctx, specimen.ID, "json", "smoke")
	if err != nil {
		return err
	}
	if _, err := services.Exports.Run(ctx, job.ID); err != nil {
		return err
	}
	return nil
}
