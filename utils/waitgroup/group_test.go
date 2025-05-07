package waitgroup

import (
	"errors"
	"testing"
)

func TestGroup(t *testing.T) {
	var wg Group

	// 测试 Run 方法
	wg.Run(func() {
		t.Log("Running task 1")
	})
	wg.Run(func() {
		t.Log("Running task 2")
	})

	// 等待所有任务完成
	wg.Wait()

	// 测试 RunWithRecover 方法
	wg.RunWithRecover(func() {
		panic(errors.New("simulated panic"))
	}, func(r interface{}) {
		t.Logf("Recovered from panic: %v", r)
	})

	// 等待所有任务完成
	wg.Wait()
}
