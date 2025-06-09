package workerpool_test

import (
	"sync"
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
		wg   sync.WaitGroup
		data []int
	)

	wg.Add(100)
	for i := range 100 {
		wp.Do(func() {
			defer wg.Done()
			data = append(data, i)
		})
	}

	wg.Wait()

	require.Equal(t, 100, len(data))
}

func TestCondWorkerPool_WithBatchSize(t *testing.T) {
	t.Parallel()

	wp := workerpool.NewCondWorkerPool(workerpool.WithBatchSize(10))
	wp.IncWorkers(2)
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

func BenchmarkCondWorkerPool_OneWorker(b *testing.B) {
	wp := workerpool.NewCondWorkerPool()
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

func BenchmarkCondWorkerPool_OneWorkerWithBatchProcessing(b *testing.B) {
	wp := workerpool.NewCondWorkerPool(workerpool.WithBatchSize(10))
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

func BenchmarkCondWorkerPool_FewWorkers(b *testing.B) {
	wp := workerpool.NewCondWorkerPool()
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

func BenchmarkCondWorkerPool_FewWorkersWithBatchProcessing(b *testing.B) {
	wp := workerpool.NewCondWorkerPool(workerpool.WithBatchSize(10))
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
