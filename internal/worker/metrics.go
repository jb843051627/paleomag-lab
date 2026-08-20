package worker

import "sync/atomic"

type Metrics struct {
	qualityProcessed atomic.Uint64
	exportsProcessed atomic.Uint64
	qualityFailed    atomic.Uint64
	exportsFailed    atomic.Uint64
}

func (m *Metrics) QualityProcessed() uint64 { return m.qualityProcessed.Load() }
func (m *Metrics) ExportsProcessed() uint64 { return m.exportsProcessed.Load() }
func (m *Metrics) QualityFailed() uint64    { return m.qualityFailed.Load() }
func (m *Metrics) ExportsFailed() uint64    { return m.exportsFailed.Load() }
