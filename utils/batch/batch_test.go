package batch

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCondBatch_Run(t *testing.T) {
	var wg sync.WaitGroup
	cb := NewStreamBatch[int](2, func(data []int) error {
		time.Sleep(time.Millisecond * 20)
		fmt.Printf("%d\n", data)
		wg.Done()
		return nil
	})

	for i := 0; i < 100; i++ {
		wg.Add(1)
		tmp := i
		go cb.Run(tmp)
	}
	wg.Wait()
}

func BenchmarkStreamBatch_Run(b *testing.B) {
	cb := NewStreamBatch[int](2, func(data []int) error {
		return nil
	})
	for n := 0; n < b.N; n++ {
		cb.Run(n)
	}
}
