package tc

import (
	"context"
	"github.com/lanceryou/dtf/transaction"
)

type BranchExecutor interface {
	Exec(ctx context.Context, status transaction.Status, branch *transaction.BranchTransaction) error
}

type BranchFunc func(ctx context.Context, status transaction.Status, branch *transaction.BranchTransaction) error

func (f BranchFunc) Exec(ctx context.Context, status transaction.Status, branch *transaction.BranchTransaction) error {
	return f(ctx, status, branch)
}
