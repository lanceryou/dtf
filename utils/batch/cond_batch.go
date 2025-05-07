package batch

import "sync"

type StreamBatch[T any] struct {
	num int
	mtx sync.Mutex
	fn  func([]T) error
	ds  []*Writer[T]
}

func NewStreamBatch[T any](num int, fn func([]T) error) *StreamBatch[T] {
	return &StreamBatch[T]{num: num, fn: fn}
}

func (c *StreamBatch[T]) Run(d T) (err error) {
	w := NewWriter(d, &c.mtx)
	c.mtx.Lock()
	c.ds = append(c.ds, w)
	for !w.done && c.ds[0] != w {
		w.Wait()
	}
	if w.done {
		c.mtx.Unlock()
		return w.err
	}
	num := len(c.ds)
	if num >= c.num {
		num = c.num
	}
	ws := c.ds[:num]
	c.mtx.Unlock()

	ds := make([]T, len(ws))
	for i, writer := range ws {
		ds[i] = writer.data
	}
	err = c.fn(ds)
	for i := 0; i < len(ws); i++ {
		ws[i].done = true
		ws[i].err = err
		ws[i].Signal()
	}

	c.mtx.Lock()
	c.ds = c.ds[num:]
	if len(c.ds) != 0 {
		c.ds[0].Signal()
	}
	c.mtx.Unlock()
	return err
}

type Writer[T any] struct {
	*sync.Cond
	done bool
	err  error
	data T
}

func NewWriter[T any](d T, mtx *sync.Mutex) *Writer[T] {
	return &Writer[T]{Cond: sync.NewCond(mtx), data: d}
}
