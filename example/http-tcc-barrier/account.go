package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lanceryou/dtf/api/tccpb"
	"github.com/lanceryou/dtf/api/tcpb"
	"github.com/lanceryou/dtf/client/tcc"
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
	holdIn  int64 // 在途资金
	holdOut int64 // 冻结资金
	inCnt   int64
	outCnt  int64
}

func (a *Account) Init() {
	a.POST("/TransTryIn", handler.GinHandler(a.TransTryIn))
	a.POST("/TransConfirmIn", handler.GinHandler(a.TransConfirmIn))
	a.POST("/TransCancelIn", handler.GinHandler(a.TransCancelIn))
	a.POST("/TransTryOut", handler.GinHandler(a.TransTryOut))
	a.POST("/TransConfirmOut", handler.GinHandler(a.TransConfirmOut))
	a.POST("/TransCancelOut", handler.GinHandler(a.TransCancelOut))
	a.POST("/TransIn", handler.GinHandler(a.TransIn))
	a.POST("/TransOut", handler.GinHandler(a.TransOut))
}

func (a *Account) TransTryIn(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	a.holdIn += r.Amount // 冻结金额
	return resp, nil
}

func (a *Account) TransConfirmIn(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	// 模拟处理超时，等TC补偿重试
	if r.Timeout != 0 && a.inCnt == 0 {
		slog.Info("TransConfirmIn timeout")
		a.inCnt++
		return nil, fmt.Errorf("TransConfirmIn timeout")
	}
	_, ok := a.history[r.RequestId+"-confirm"]
	// 已经执行过，直接成功
	if ok {
		return
	}
	a.history[r.RequestId+"-confirm"] = a.holdIn
	a.amount += a.holdIn
	a.holdIn = 0
	a.inCnt = 0
	return resp, nil
}

func (a *Account) TransCancelIn(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	a.holdIn = 0 // 清空之前冻结的金额
	return resp, nil
}

func (a *Account) TransTryOut(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	if r.Amount%2 != 0 {
		return resp, fmt.Errorf("amount error:%v", r.Amount)
	}
	a.holdOut += r.Amount
	a.amount -= r.Amount
	return resp, nil
}

func (a *Account) TransConfirmOut(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	// 模拟处理超时，等TC补偿重试
	if r.Timeout != 0 && a.outCnt == 0 {
		a.outCnt++
		slog.Info("TransConfirmOut timeout")
		return nil, fmt.Errorf("TransConfirmOut timeout")
	}
	_, ok := a.history["out-"+r.RequestId+"-confirm"]
	// 已经执行过，直接成功
	if ok {
		return
	}
	// history 类似流水去重表
	a.history["out-"+r.RequestId+"-confirm"] = r.Amount
	a.holdOut = 0
	a.outCnt = 0
	return resp, nil
}

func (a *Account) TransCancelOut(ctx context.Context, r *TransAmount) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	a.amount += a.holdOut
	a.holdOut = 0
	return resp, nil
}

func (a *Account) TransIn(ctx context.Context, r *tccpb.TccRequest) (resp *tcpb.Empty, err error) {
	var req TransAmount
	if err = json.Unmarshal([]byte(r.Payloads), &req); err != nil {
		return nil, err
	}

	switch r.Op {
	case tccpb.TccOp_Try:
		err = tcc.Barrier(a.DB).Run(r.Gid, r.BranchId, uint32(r.Op), func() error {
			resp, err = a.TransTryIn(ctx, &req)
			return err
		})
		return resp, err
	case tccpb.TccOp_Confirm:
		return a.TransConfirmIn(ctx, &req)
	case tccpb.TccOp_Cancel:
		err = tcc.Barrier(a.DB).Run(r.Gid, r.BranchId, uint32(r.Op), func() error {
			resp, err = a.TransCancelIn(ctx, &req)
			return err
		})
		return resp, err
	}
	return nil, fmt.Errorf("invalid op:%v", r.Op)
}

func (a *Account) TransOut(ctx context.Context, r *tccpb.TccRequest) (resp *tcpb.Empty, err error) {
	var req TransAmount
	if err = json.Unmarshal([]byte(r.Payloads), &req); err != nil {
		return nil, err
	}

	switch r.Op {
	case tccpb.TccOp_Try:
		if req.RequestId == "null" {
			return nil, fmt.Errorf("null try")
		}
		err = tcc.Barrier(a.DB).Run(r.Gid, r.BranchId, uint32(r.Op), func() error {
			resp, err = a.TransTryOut(ctx, &req)
			return err
		})
		return resp, err
	case tccpb.TccOp_Confirm:
		return a.TransConfirmOut(ctx, &req)
	case tccpb.TccOp_Cancel:
		err = tcc.Barrier(a.DB).Run(r.Gid, r.BranchId, uint32(r.Op), func() error {
			resp, err = a.TransCancelOut(ctx, &req)
			return err
		})
		return resp, err
	}
	return nil, fmt.Errorf("invalid op:%v", r.Op)
}
