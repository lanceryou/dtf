package base

import (
	"context"
	"fmt"
	"github.com/lanceryou/dtf/transaction"
	"log/slog"
)

var bh = map[string]func(string) error{}

func RegisterBaseHandler(name string, fn func(string) error) {
	bh[name] = fn
}

func BaseBranchExec(ctx context.Context, status transaction.Status, branch *transaction.BranchTransaction) error {
	// base 只需要处理submit状态， abort 说明本地事务失败直接返回就行
	if status != transaction.Submit {
		return nil
	}

	fn, ok := bh[branch.Header["base"]]
	if !ok {
		slog.ErrorContext(ctx, "BaseBranchExec failed. can not found", "header", branch.Header)
		return fmt.Errorf("header base not found.")
	}

	return fn(branch.Resource)
}
