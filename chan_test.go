package workerpool_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/danilovaalina/workerpool"
	"github.com/stretchr/testify/require"
)

func TestChanWorkerPool(t *testing.T) {
	t.Parallel()

	wp := workerpool.NewChanWorkerPool()
	wp.IncWorkers(4)
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

func TestChanWorkerPool_WithJobLimit(t *testing.T) {
	t.Parallel()

	wp := workerpool.NewChanWorkerPool(workerpool.WithJobLimit(100))
	wp.IncWorkers(4)
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

func TestChanWorkerPool_ClosedPool(t *testing.T) {
	t.Parallel()

	wp := workerpool.NewChanWorkerPool()
	wp.IncWorkers(1)
	wp.CloseAndWait()

	err := wp.Do(func() {})
	require.Errorf(t, err, "worker pool closed")
}

func BenchmarkChanWorkerPool(b *testing.B) {
	tests := []struct {
		name  string
		n     int
		limit int
	}{
		{"OneWorker", 1, 0},
		{"OneWorkerWithJobLimit100", 1, 100},
		{"OneWorkerWithJobLimit10000", 1, 10000},
		{"OneWorkerWithJobLimit100000", 1, 100000},
		{"FewWorkers", 4, 0},
		{"FewWorkersWithJobLimit100", 4, 100},
		{"FewWorkersWithJobLimit10000", 4, 10000},
		{"FewWorkersWithJobLimit100000", 4, 100000},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			wp := workerpool.NewChanWorkerPool(workerpool.WithJobLimit(test.limit))
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
