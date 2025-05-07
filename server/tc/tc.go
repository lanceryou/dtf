package tc

import (
	"context"
	"errors"
	"fmt"
	"github.com/lanceryou/dtf/transaction"
	"github.com/lanceryou/dtf/transaction/store"
	"github.com/lanceryou/dtf/utils/batch"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"time"
)

var (
	InvalidStatus = errors.New("invalid status")
)

func NewTC(rt RunTimer, store store.Store, cnt int) *TC {
	tc := &TC{
		RunTimer: rt,
		Store:    store,
		BE:       make(map[string]BranchExecutor),
	}
	tc.batch = batch.NewStreamBatch(cnt, func(ts []*transaction.GlobalTransaction) error {
		return tc.Store.Save(ts...)
	})
	return tc
}

// TC 全局事务协调者
// 处理事务状态转换
type TC struct {
	RunTimer RunTimer
	Store    store.Store
	BE       map[string]BranchExecutor
	batch    *batch.StreamBatch[*transaction.GlobalTransaction]
}

func (t *TC) RegisterBranchExecutor(transType string, executor BranchExecutor) {
	t.BE[transType] = executor
}

// Prepare 协调者收到prepare请求处理
func (t *TC) Prepare(ctx context.Context, gt *transaction.GlobalTransaction) error {
	return t.changeStatus(gt, transaction.Prepare)
}

// Submit 当前状态表示所有prepare通过，可以执行分支事务
func (t *TC) Submit(ctx context.Context, gid string) error {
	gt, err := t.Store.Get(gid)
	if err != nil {
		return err
	}
	// 假如事务已经进行到abort阶段或者已经完成事务，直接拒绝
	// TODO 这里不处理超时，假如超时业务方自己abort或者等协调者定时执行
	if gt.Status == transaction.Success || gt.Status == transaction.Abort {
		slog.Error("receive submit request, but status invalid,", "status", transaction.StatusM[gt.Status], "gid", gt.Gid)
		return InvalidStatus
	}
	if err = t.changeStatus(gt, transaction.Submit); err != nil {
		return err
	}
	// submit 分支事务
	// TODO 执行分支失败需要考虑 可能超时失败实际成功
	// 考虑直接返回成功， 监控定时执行， 可能需要人工介入
	if err = t.execBranch(ctx, gt); err != nil {
		return err
	}
	// 分支执行成功，假如change失败等待定时执行，不需要返回错误信息
	_ = t.changeStatus(gt, transaction.Success)
	return nil
}

func (t *TC) Abort(ctx context.Context, gid string) error {
	gt, err := t.Store.Get(gid)
	if err != nil {
		return err
	}
	// 已经进入提交或者完成阶段不允许回滚，等待submit完成
	if gt.Status == transaction.Success || gt.Status == transaction.Submit {
		slog.Error("receive abort request, but status invalid,", "status", transaction.StatusM[gt.Status], "gid", gt.Gid)
		return InvalidStatus
	}
	if err = t.changeStatus(gt, transaction.Abort); err != nil {
		return err
	}
	// abort 分支事务
	if err = t.execBranch(ctx, gt); err != nil {
		return err
	}
	_ = t.changeStatus(gt, transaction.Failed)
	return nil
}

func (t *TC) RegisterBranch(bt *transaction.BranchTransaction) error {
	bt.Status = transaction.Prepare
	bt.TransTime = time.Now().Unix()
	slog.Info("RegisterBranch", "gid", bt.Gid, ",branch_id", bt.BranchId, ",status=prepare,time", bt.TransTime, "header", bt.Header)
	return t.Store.RegisterBranch(bt)
}

func (t *TC) changeStatus(gt *transaction.GlobalTransaction, status transaction.Status) error {
	now := time.Now()
	gt.Status = status
	gt.TransTime = now.Unix()
	gt.NextRunTime = t.RunTimer.NextRunTime(now)
	slog.Info("changeStatus", "gid", gt.Gid, ",status", transaction.StatusM[status], ",time", gt.TransTime)
	return t.batch.Run(gt)
}

func (t *TC) execBranch(ctx context.Context, gt *transaction.GlobalTransaction) error {
	exec, ok := t.BE[gt.TransType]
	if !ok {
		return fmt.Errorf("trans type is not support:%s", gt.TransType)
	}
	var eg errgroup.Group
	for _, branch := range gt.Branches {
		bch := branch
		eg.Go(func() error {
			err := exec.Exec(ctx, gt.Status, bch)
			if err != nil {
				slog.Error("branch transaction exec failed", "branch", branch, "err", err)
			}
			return err
		})
	}

	return eg.Wait()
}
