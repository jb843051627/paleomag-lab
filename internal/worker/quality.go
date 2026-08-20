package worker

import (
	"context"
	"log/slog"
)

func (m *Manager) qualityLoop() {
	defer m.wg.Done()
	logger := slog.Default()
	for {
		select {
		case <-m.ctx.Done():
			return
		case runID := <-m.queues.Quality:
			if _, err := m.services.Measurements.EvaluateQuality(m.ctx, runID); err != nil && !isContextError(m.ctx, err) {
				logger.Error("quality evaluation failed", "run_id", runID, "error", err)
			}
		}
	}
}

func isContextError(ctx context.Context, err error) bool {
	return ctx.Err() != nil || err == context.Canceled || err == context.DeadlineExceeded
}
