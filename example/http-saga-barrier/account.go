package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lanceryou/dtf/api/tcpb"
	"github.com/lanceryou/dtf/client/saga"
	"github.com/lanceryou/dtf/utils/gorm"
	"github.com/lanceryou/dtf/utils/http/handler"
	orm "gorm.io/gorm"
	"log/slog"
)

type TransAmount struct {
	RequestId string
	Amount    int64
	Timeout   int
}

func NewAccount(dsn string, addr string, amount int64) *Account {
	acc := &Account{Engine: gin.New(), amount: amount, history: make(map[string]int64)}
	acc.DB = gorm.NewDB(dsn)
	acc.Init()
	go acc.Run(addr)
	return acc
}

type Account struct {
	*gin.Engine
	*orm.DB
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
	err = saga.Barrier(a.DB).Run(r.RequestId, r.RequestId+"-TransIn", saga.Prepare, func() error {
		a.amount += r.Amount
		return nil
	})
	slog.Info("TransIn", "amount", a.amount, "req amount", r.Amount, "err", err)
	return resp, err
}

func (a *Account) TransCancelIn(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	err = saga.Barrier(a.DB).Run(r.RequestId, r.RequestId+"-TransIn", saga.Compensation, func() error {
		a.amount -= r.Amount
		return nil
	})
	slog.Info("TransCancelIn", "amount", a.amount, "req amount", r.Amount, "err", err)
	return resp, nil
}

func (a *Account) TransOut(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	if r.Amount%2 != 0 {
		return resp, fmt.Errorf("amount error:%v", r.Amount)
	}
	// history 类似流水去重表
	err = saga.Barrier(a.DB).Run(r.RequestId, r.RequestId+"-TransOut", saga.Prepare, func() error {
		a.amount -= r.Amount
		return nil
	})
	slog.Info("TransOut", "amount", a.amount, "req amount", r.Amount, "err", err)
	return resp, nil
}

func (a *Account) TransCancelOut(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	err = saga.Barrier(a.DB).Run(r.RequestId, r.RequestId+"-TransOut", saga.Compensation, func() error {
		a.amount += r.Amount
		return nil
	})
	slog.Info("TransCancelOut", "amount", a.amount, "req amount", r.Amount, "err", err)
	return resp, nil
}
