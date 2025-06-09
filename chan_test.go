package workerpool_test

import (
	"sync"
	"testing"

	"github.com/danilovaalina/workerpool"
	"github.com/stretchr/testify/require"
)

func TestChanWorkerPool_WithJobLimit(t *testing.T) {
	t.Parallel()

	wp := workerpool.NewChanWorkerPool(workerpool.WithJobLimit(5))
	wp.IncWorkers(4)
	defer wp.CloseAndWait()

	var (
		wg   sync.WaitGroup
		data []int
	)

	wg.Add(100)
	for i := range 100 {
		go wp.Do(func() {
			defer wg.Done()
			data = append(data, i)
		})
	}

	wg.Wait()

	require.Equal(t, 100, len(data))
}

func TestChanWorkerPool_ClosedPool(t *testing.T) {
	t.Parallel()

	wp := workerpool.NewChanWorkerPool()
	wp.IncWorkers(1)
	wp.CloseAndWait()

	err := wp.Do(func() {})
	require.Errorf(t, err, "worker pool closed")
}

func BenchmarkChanWorkerPool_OneWorker(b *testing.B) {
	wp := workerpool.NewChanWorkerPool()
	wp.IncWorkers(1)
	defer wp.CloseAndWait()

	var wg sync.WaitGroup

	b.ResetTimer()
	wg.Add(b.N)
	for range b.N {
		go wp.Do(wg.Done)
	}

	wg.Wait()
}

func BenchmarkChanWorkerPool_FewWorkers(b *testing.B) {
	wp := workerpool.NewChanWorkerPool()
	wp.IncWorkers(4)
	defer wp.CloseAndWait()

	var wg sync.WaitGroup

	b.ResetTimer()
	wg.Add(b.N)
	for range b.N {
		go wp.Do(wg.Done)
	}

	wg.Wait()
}
