package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/service"
)

type Queues struct {
	Quality chan string
	Exports chan string
}

func NewQueues(size int) *Queues {
	if size < 1 {
		size = 8
	}
	return &Queues{Quality: make(chan string, size), Exports: make(chan string, size)}
}

type Manager struct {
	services *service.Services
	queues   *Queues
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewManager(services *service.Services, queues *Queues) *Manager {
	return &Manager{services: services, queues: queues}
}

func (m *Manager) Start(parent context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ctx != nil {
		return
	}
	m.ctx, m.cancel = context.WithCancel(parent)
	m.wg.Add(2)
	go m.qualityLoop()
	go m.exportLoop()
}

func (m *Manager) EnqueueQuality(runID string) error {
	if runID == "" {
		return fmt.Errorf("%w: empty run id", model.ErrInvalid)
	}
	m.queues.Quality <- runID
	return nil
}

func (m *Manager) EnqueueExport(jobID string) error {
	if jobID == "" {
		return fmt.Errorf("%w: empty export id", model.ErrInvalid)
	}
	m.queues.Exports <- jobID
	return nil
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel == nil {
		return
	}
	m.cancel()
	m.wg.Wait()
	m.ctx = nil
	m.cancel = nil
}
