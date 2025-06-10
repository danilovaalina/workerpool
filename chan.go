package workerpool

import (
	"errors"
	"sync"
)

type ChanWorkerPoolOption func(*ChanWorkerPool)

func WithJobLimit(limit int) ChanWorkerPoolOption {
	return func(wp *ChanWorkerPool) {
		if limit > 0 {
			wp.jobs = make(chan Job, limit)
		}
	}
}

type ChanWorkerPool struct {
	limit  int
	jobs   chan Job
	done   chan struct{}
	mu     sync.Mutex
	n      int
	closed chan struct{}
	once   sync.Once
}

func NewChanWorkerPool(opts ...ChanWorkerPoolOption) *ChanWorkerPool {
	p := &ChanWorkerPool{
		jobs:   make(chan Job),
		done:   make(chan struct{}),
		closed: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *ChanWorkerPool) IncWorkers(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for range n {
		p.n++
		go p.worker()
	}
}

func (p *ChanWorkerPool) DecWorkers(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.close(n)
}

func (p *ChanWorkerPool) CloseAndWait() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.close(p.n)
}

func (p *ChanWorkerPool) Do(f Job) error {
	select {
	case p.jobs <- f:
		return nil
	case <-p.closed:
		return errors.New("worker pool closed")
	}
}

func (p *ChanWorkerPool) close(n int) {
	for range min(n, p.n) {
		p.n--
		p.done <- struct{}{}
	}

	if p.n == 0 {
		p.once.Do(func() {
			close(p.closed)
		})
	}
}

func (p *ChanWorkerPool) worker() {
	for {
		select {
		case f := <-p.jobs:
			f()
		case <-p.done:
			return
		}
	}
}
