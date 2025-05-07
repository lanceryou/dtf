package saga

import (
	"context"
	"encoding/json"
	"github.com/lanceryou/dtf/client"
	"log/slog"
)

func NewSaga(ts *client.Transaction) *Saga {
	return &Saga{
		Transaction: ts,
	}
}

// Saga saga 一般是正向和补偿一起构成
// 假如正向成功，则需要注册补偿到事务框架
type Saga struct {
	*client.Transaction
}

// AddCompensation 添加补偿信息
func (s *Saga) AddCompensation(ctx context.Context, resource interface{}, compensation string) (string, error) {
	branchId := s.BranchID()
	data, err := json.Marshal(resource)
	if err != nil {
		return branchId, err
	}
	err = s.RegisterBranch(ctx, string(data), branchId, map[string]string{
		"saga": compensation,
	})
	return branchId, err
}

func Transaction(ctx context.Context, saga *Saga, fn func(context.Context, *Saga) error) error {
	if err := saga.Prepare(ctx); err != nil {
		slog.ErrorContext(ctx, "saga transaction prepare failed.", "gid", saga.Gid, "err", err)
		return err
	}
	if err := fn(ctx, saga); err != nil {
		slog.ErrorContext(ctx, "saga transaction exec failed.", "gid", saga.Gid, "err", err)
		return saga.Abort(ctx, err)
	}

	return saga.Submit(ctx)
}
