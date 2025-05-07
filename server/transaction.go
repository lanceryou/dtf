package server

import (
	"context"
	"github.com/lanceryou/dtf/api/tcpb"
	"github.com/lanceryou/dtf/server/tc"
	"github.com/lanceryou/dtf/transaction"
	"github.com/lanceryou/dtf/transaction/store"
	"github.com/lanceryou/dtf/utils/dlock"
	"github.com/lanceryou/dtf/utils/waitgroup"
	"log/slog"
	"time"
)

func NewTransactionServer(tc *tc.TC, locker dlock.DistLock, duration time.Duration) *TransactionServer {
	server := &TransactionServer{coordinate: tc, locker: locker, duration: duration}
	go server.loop()
	return server
}

// TransactionServer 与具体协议无关的server
// 负责管理定时处理abort或者submit的请求
type TransactionServer struct {
	coordinate *tc.TC
	locker     dlock.DistLock
	duration   time.Duration
}

func (t *TransactionServer) loop() {
	k := time.NewTicker(t.duration)
	for {
		select {
		case <-k.C:
			gts, err := t.coordinate.Store.Query(&store.TransactionCond{
				Runtime: time.Now().Unix(),
				Status:  []uint32{uint32(transaction.Submit), uint32(transaction.Abort), uint32(transaction.Prepare)},
				Limit:   100,
			})
			if err != nil {
				slog.Error("query transactions failed", "err", err)
				continue
			}
			// 考虑prepare状态 超时abort， 可能存在业务已经完成操作但是请求框架超时，业务已经返回失败信息给用户
			// 做好相关监控 可能人工介入
			var wg waitgroup.Group
			for _, gt := range gts {
				trans := gt
				wg.Run(func() {
					if err = t.task(context.TODO(), trans); err != nil {
						slog.Error("task error", "transaction", gt, "err", err)
					}
				})
			}
			wg.Wait()
		}
	}
}

func (t *TransactionServer) task(ctx context.Context, gt *transaction.GlobalTransaction) error {
	err := t.locker.Lock(ctx, gt.Gid, dlock.WithLease())
	if err != nil {
		return err
	}
	// 锁住资源
	defer t.locker.UnLock(ctx, gt.Gid)
	// 补偿
	if gt.Status == transaction.Submit {
		return t.coordinate.Submit(ctx, gt.Gid)
	} else {
		// prepare 或者 abort
		return t.coordinate.Abort(ctx, gt.Gid)
	}
}

func (t *TransactionServer) Prepare(ctx context.Context, r *tcpb.TranRequest) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	err = t.coordinate.Prepare(ctx, &transaction.GlobalTransaction{
		Gid:       r.Gid,
		TransType: r.TransType,
	})
	return
}

func (t *TransactionServer) Submit(ctx context.Context, r *tcpb.TranRequest) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	err = t.coordinate.Submit(ctx, r.Gid)
	return
}

func (t *TransactionServer) Abort(ctx context.Context, r *tcpb.TranRequest) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	err = t.coordinate.Abort(ctx, r.Gid)
	return
}

func (t *TransactionServer) RegisterBranch(ctx context.Context, r *tcpb.BranchRequest) (resp *tcpb.Empty, err error) {
	resp = &tcpb.Empty{}
	err = t.coordinate.RegisterBranch(&transaction.BranchTransaction{
		Gid:      r.Gid,
		BranchId: r.BranchId,
		Header:   r.Header,
		Resource: r.Resource,
	})
	return
}

var _ transaction.Transaction = &TransactionServer{}
