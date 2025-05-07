package tccbranch

import (
	"context"
	"fmt"
	"github.com/lanceryou/dtf/api/tccpb"
	"github.com/lanceryou/dtf/api/tcpb"
	"github.com/lanceryou/dtf/transaction"
	hc "github.com/lanceryou/dtf/utils/http/client"
	"log/slog"
)

// HttpTccBranchExec RM 支持http协议
// 默认实现tcc标准接口, 具体参见tcc proto
func HttpTccBranchExec(ctx context.Context, status transaction.Status, branch *transaction.BranchTransaction) error {
	if len(branch.Header) != 1 {
		return fmt.Errorf("tcc branch url is nor standrad. header:%v", branch.Header)
	}
	hkey := map[transaction.Status]tccpb.TccOp{
		transaction.Prepare: tccpb.TccOp_Try,
		transaction.Submit:  tccpb.TccOp_Confirm,
		transaction.Abort:   tccpb.TccOp_Cancel,
	}
	var rsp tcpb.Empty
	client := hc.NewHttpClient()
	err := client.Post(ctx, branch.Header["tcc"], &tccpb.TccRequest{
		Op:       hkey[status],
		Payloads: branch.Resource,
		Gid:      branch.Gid,
		BranchId: branch.BranchId,
	}, &rsp)
	if err != nil {
		slog.ErrorContext(ctx, "HttpTccBranchExec failed.", "status", transaction.StatusM[status], "branch", branch, "err", err)
		return err
	}
	slog.InfoContext(ctx, "HttpTccBranchExec success.", "status", transaction.StatusM[status], "branch", branch)
	return nil
}
