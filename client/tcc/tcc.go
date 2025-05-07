package tcc

import (
	"context"
	"encoding/json"
	"github.com/lanceryou/dtf/api/tccpb"
	"github.com/lanceryou/dtf/client"
	"log/slog"
)

func NewTcc(ts *client.Transaction) *Tcc {
	return &Tcc{
		Transaction: ts,
	}
}

// Tcc 分布式事务
type Tcc struct {
	*client.Transaction
}

// Try 业务预留资源， 假如成功进入commit阶段，失败进入abort
func (t *Tcc) Try(ctx context.Context, resource interface{}, header map[string]string) error {
	branchId := t.BranchID()
	data, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	if err = t.RegisterBranch(ctx, string(data), branchId, header); err != nil {
		return err
	}
	// 发起try， 这里采用tcc proto标准接口， 业务也可以自定义自己的try
	return t.ExecBranch(ctx, header["tcc"], &tccpb.TccRequest{
		Gid:      t.Gid,
		BranchId: branchId,
		Payloads: string(data),
		Op:       tccpb.TccOp_Try,
		Header:   header,
	})
}

// Transaction tcc 事务
// 发起prepare 请求
// fn 执行业务逻辑，预扣资源
// 发起submit或者abort
func Transaction(ctx context.Context, tcc *Tcc, fn func(context.Context, *Tcc) error) error {
	if err := tcc.Prepare(ctx); err != nil {
		slog.ErrorContext(ctx, "tcc transaction prepare failed.", "gid", tcc.Gid, "err", err)
		return err
	}
	if err := fn(ctx, tcc); err != nil {
		slog.ErrorContext(ctx, "tcc transaction exec failed.", "gid", tcc.Gid, "err", err)
		return tcc.Abort(ctx, err)
	}

	return tcc.Submit(ctx)
}
