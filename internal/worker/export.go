package worker

import "log/slog"

func (m *Manager) exportLoop() {
	defer m.wg.Done()
	logger := slog.Default()
	for {
		select {
		case <-m.ctx.Done():
			return
		case jobID := <-m.queues.Exports:
			if _, err := m.services.Exports.Run(m.ctx, jobID); err != nil {
				logger.Error("export failed", "job_id", jobID, "error", err)
			}
		}
	}
}
