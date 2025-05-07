package base

import (
	"context"
	"encoding/json"
	"github.com/lanceryou/dtf/client"
	"log/slog"
)

func NewBase(ts *client.Transaction) *Base {
	return &Base{Transaction: ts}
}

type Base struct {
	*client.Transaction
}

func (b *Base) Register(ctx context.Context, resource interface{}, name string) error {
	branchId := b.BranchID()
	data, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	return b.RegisterBranch(ctx, string(data), branchId, map[string]string{
		"base": name,
	})
}

func Transaction(ctx context.Context, base *Base, fn func(context.Context, *Base) error) error {
	if err := base.Prepare(ctx); err != nil {
		slog.ErrorContext(ctx, "base transaction prepare failed.", "gid", base.Gid, "err", err)
		return err
	}
	if err := fn(ctx, base); err != nil {
		slog.ErrorContext(ctx, "base transaction exec failed.", "gid", base.Gid, "err", err)
		return base.Abort(ctx, err)
	}

	return base.Submit(ctx)
}
