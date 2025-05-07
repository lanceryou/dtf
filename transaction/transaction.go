package transaction

import (
	"context"
	"github.com/lanceryou/dtf/api/tcpb"
)

// Transaction 事务接口
// client 和 server 分别实现
type Transaction interface {
	Prepare(context.Context, *tcpb.TranRequest) (*tcpb.Empty, error)
	Submit(context.Context, *tcpb.TranRequest) (*tcpb.Empty, error)
	Abort(context.Context, *tcpb.TranRequest) (*tcpb.Empty, error)
	RegisterBranch(context.Context, *tcpb.BranchRequest) (*tcpb.Empty, error)
}

type Status uint32

var StatusM = map[Status]string{
	Prepare: "prepare",
	Submit:  "submit",
	Abort:   "abort",
	Success: "success",
}

const (
	Prepare Status = iota + 1
	Submit
	Abort
	Success // 完成事务
	Failed  // 事务失败
)

// GlobalTransaction 全局事务
type GlobalTransaction struct {
	Gid         string               // 全局事务id
	Status      Status               // 全局事务状态
	TransTime   int64                // 事务发生时间
	TransType   string               // 事务类型 tcc, saga...
	NextRunTime int64                // 下次执行时间
	Reason      string               // 失败原因
	Branches    []*BranchTransaction // 分支事务信息
}

// BranchTransaction 分支事务
type BranchTransaction struct {
	Gid       string            // 全局事务id
	BranchId  string            // branch id
	Status    Status            // 分支事务状态
	TransTime int64             // 事务发生时间
	Reason    string            // 失败原因
	Header    map[string]string // 回调分支信息
	Resource  string
}

// RegisterBranch 注册分支事务
func (t *GlobalTransaction) RegisterBranch(b *BranchTransaction) {
	t.Branches = append(t.Branches, b)
}
