package waitgroup

import (
	"sync"
)

type Group struct {
	sync.WaitGroup
}

// Run 简化协程处理
// eg var wg Wrapper
// wg.Run(func(){
// do
// })
// wg.Run(func(){
// do
// })
// wg.Wait()
func (w *Group) Run(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}

func (w *Group) RunWithRecover(cb func(), recoverHandler func(r interface{})) {
	w.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if recoverHandler != nil {
					recoverHandler(r)
				}
			}
			w.Done()
		}()
		cb()
	}()
}
