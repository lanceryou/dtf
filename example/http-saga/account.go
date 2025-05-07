package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lanceryou/dtf/api/tcpb"
	"github.com/lanceryou/dtf/utils/http/handler"
	"log/slog"
)

type TransAmount struct {
	RequestId string
	Amount    int64
	Timeout   int
}

func NewAccount(addr string, amount int64) *Account {
	acc := &Account{Engine: gin.New(), amount: amount, history: make(map[string]int64)}
	acc.Init()
	go acc.Run(addr)
	return acc
}

type Account struct {
	*gin.Engine
	history map[string]int64
	amount  int64
}

func (a *Account) Init() {
	a.POST("/TransIn", handler.GinHandler(a.TransIn))
	a.POST("/TransCancelIn", handler.GinHandler(a.TransCancelIn))
	a.POST("/TransOut", handler.GinHandler(a.TransOut))
	a.POST("/TransCancelOut", handler.GinHandler(a.TransCancelOut))
}

func (a *Account) TransIn(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	// 空悬挂处理
	// try 因为超时，导致tc直接执行了cancel
	// 执行后try到达
	// 实际执行一般会需要全局锁控制
	_, ok := a.history[r.RequestId+"-cancel"]
	// 产生空悬挂，返回失败，等待TC执行cancel
	if ok {
		return nil, fmt.Errorf("cancel reach early")
	}
	// history 类似流水去重表
	a.history[r.RequestId+"-try"] = r.Amount
	a.amount += r.Amount
	slog.Info("TransIn", "amount", a.amount, "req amount", r.Amount)
	return resp, nil
}

func (a *Account) TransCancelIn(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	// 处理空回滚之类的场景
	_, ok := a.history[r.RequestId+"-try"]
	// try 没有成功说明之前没有预留资源成功，直接回滚成功
	if !ok {
		return
	}
	a.history[r.RequestId+"-cancel"] = r.Amount
	a.amount -= r.Amount
	slog.Info("TransCancelIn", "amount", a.amount, "req amount", r.Amount)
	return resp, nil
}

func (a *Account) TransOut(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	if r.Amount%2 != 0 {
		return resp, fmt.Errorf("amount error:%v", r.Amount)
	}
	// history 类似流水去重表
	a.history["out-"+r.RequestId+"-try"] = r.Amount
	a.amount -= r.Amount
	slog.Info("TransOut", "amount", a.amount, "req amount", r.Amount)
	return resp, nil
}

func (a *Account) TransCancelOut(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	_, ok := a.history["out-"+r.RequestId+"-try"]
	if !ok {
		return
	}
	a.history["out-"+r.RequestId+"-cancel"] = r.Amount
	a.amount += r.Amount
	slog.Info("TransCancelOut", "amount", a.amount, "req amount", r.Amount)
	return resp, nil
}
