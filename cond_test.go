package workerpool_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/danilovaalina/workerpool"
	"github.com/stretchr/testify/require"
)

func TestCondWorkerPool(t *testing.T) {
	t.Parallel()

	wp := workerpool.NewCondWorkerPool()
	wp.IncWorkers(2)
	defer wp.CloseAndWait()

	var (
		wg    sync.WaitGroup
		value atomic.Int64
	)

	wg.Add(100)
	for range 100 {
		wp.Do(func() {
			defer wg.Done()
			value.Add(1)
		})
	}

	wg.Wait()

	require.Equal(t, int64(100), value.Load())
}

func TestCondWorkerPool_WithBatchSize(t *testing.T) {
	t.Parallel()

	wp := workerpool.NewCondWorkerPool(workerpool.WithBatchSize(10))
	wp.IncWorkers(2)
	defer wp.CloseAndWait()

	var (
		wg    sync.WaitGroup
		value atomic.Int64
	)

	wg.Add(100)
	for range 100 {
		go wp.Do(func() {
			defer wg.Done()
			value.Add(1)
		})
	}

	wg.Wait()

	require.Equal(t, int64(100), value.Load())
}

func TestCondWorkerPool_ClosedPool(t *testing.T) {
	t.Parallel()

	wp := workerpool.NewChanWorkerPool()
	wp.IncWorkers(1)
	wp.CloseAndWait()

	err := wp.Do(func() {})
	require.Errorf(t, err, "worker pool closed")
}

func BenchmarkCondWorkerPool(b *testing.B) {
	tests := []struct {
		name string
		n    int
		size int
	}{
		{"OneWorker", 1, 0},
		{"OneWorkerWithBatchProcessing", 1, 10},
		{"FewWorkers", 4, 0},
		{"FewWorkersWithBatchProcessing", 4, 10},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			wp := workerpool.NewCondWorkerPool(workerpool.WithBatchSize(test.size))
			wp.IncWorkers(test.n)
			defer wp.CloseAndWait()

			var wg sync.WaitGroup

			b.ResetTimer()
			wg.Add(b.N)
			for range b.N {
				go wp.Do(wg.Done)
			}

			wg.Wait()
		})
	}
}
