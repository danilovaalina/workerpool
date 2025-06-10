package workerpool

import (
	"container/list"
	"errors"
	"sync"
)

type CondWorkerPoolOption func(*CondWorkerPool)

func WithBatchSize(size int) CondWorkerPoolOption {
	return func(wp *CondWorkerPool) {
		if size > 0 {
			wp.batchSize = size
		}
	}
}

type CondWorkerPool struct {
	batchSize int
	cond      *sync.Cond
	n         int
	queue     *list.List
	closed    bool
	closeN    int
}

func NewCondWorkerPool(opts ...CondWorkerPoolOption) *CondWorkerPool {
	p := &CondWorkerPool{
		batchSize: 1,
		cond:      sync.NewCond(&sync.Mutex{}),
		queue:     list.New(),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *CondWorkerPool) Do(f Job) error {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	if p.closed {
		return errors.New("worker pool closed")
	}

	p.queue.PushBack(f)

	p.cond.Signal()

	return nil
}

func (p *CondWorkerPool) CloseAndWait() {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.closed = true
	p.closeN = 0
	p.cond.Broadcast()

	for p.n > 0 {
		p.cond.Wait()
	}
}

func (p *CondWorkerPool) IncWorkers(n int) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	for range n {
		p.n++
		go p.worker()
	}
}

func (p *CondWorkerPool) DecWorkers(n int) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.closeN = min(p.closeN+n, p.n)
	p.cond.Broadcast()
}

func (p *CondWorkerPool) worker() {
	for {
		p.cond.L.Lock()

		for p.queue.Len() == 0 {
			if p.closed {
				p.n--
				p.cond.Broadcast()
				p.cond.L.Unlock()
				return
			}

			if p.closeN > 0 {
				p.n--
				p.closeN--
				p.cond.L.Unlock()
				return
			}

			p.cond.Wait()
		}

		var jobs []Job
		for range min(p.batchSize, p.queue.Len()) {
			elem := p.queue.Front()
			job := elem.Value.(Job)
			p.queue.Remove(elem)
			jobs = append(jobs, job)
		}

		p.cond.L.Unlock()

		for _, job := range jobs {
			job()
		}
	}
}
