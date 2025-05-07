package saga

import (
	"context"
	"github.com/lanceryou/dtf/transaction"
	hc "github.com/lanceryou/dtf/utils/http/client"
	"log/slog"
)

type Empty struct{}

// HttpSagaBranchExec RM 支持http协议
func HttpSagaBranchExec(ctx context.Context, status transaction.Status, branch *transaction.BranchTransaction) error {
	if status != transaction.Abort {
		return nil
	}
	// 只需要处理abort，submit说明全部执行成功，不需要补偿
	// prepare 不需要处理
	var rsp Empty
	client := hc.NewHttpClient()
	err := client.Post(ctx, branch.Header["saga"], branch.Resource, &rsp)
	if err != nil {
		slog.ErrorContext(ctx, "HttpSagaBranchExec failed.", "status", transaction.StatusM[status], "branch", branch, "err", err)
		return err
	}
	slog.InfoContext(ctx, "HttpSagaBranchExec success.", "status", transaction.StatusM[status], "branch", branch)
	return nil
}
