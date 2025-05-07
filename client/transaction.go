package client

import (
	"context"
	"github.com/lanceryou/dtf/api/tcpb"
	"github.com/lanceryou/dtf/transaction"
	"github.com/lanceryou/dtf/utils/id"
)

type Options struct {
	gen id.Generator
	tb  Branch
}

func (o *Options) apply() {
	if o.gen == nil {
		o.gen = id.GeneratorFunc(id.UUID)
	}

	if o.tb == nil {
		o.tb = NewHttpBranch()
	}
}

type Option func(options *Options)

func WithGenId(gen id.Generator) Option {
	return func(options *Options) {
		options.gen = gen
	}
}

func WithBranch(tb Branch) Option {
	return func(options *Options) {
		options.tb = tb
	}
}

type Transaction struct {
	Branch
	opt       Options
	Gid       string
	transType string
	ts        transaction.Transaction
}

func NewTransaction(ts transaction.Transaction, transType string, ops ...Option) *Transaction {
	var opt Options
	for _, op := range ops {
		op(&opt)
	}
	opt.apply()

	return &Transaction{
		opt:       opt,
		Gid:       opt.gen.ID(),
		transType: transType,
		ts:        ts,
		Branch:    opt.tb,
	}
}

// RegisterBranch 注册分支
// 协议 http, grpc, ... confirm， cancel是需要服务端执行，
// 服务端提供不同协议实现方式
func (t *Transaction) RegisterBranch(ctx context.Context, resource string, branchId string, header map[string]string) error {
	_, err := t.ts.RegisterBranch(ctx, &tcpb.BranchRequest{
		Gid:      t.Gid,
		BranchId: branchId,
		Header:   header,
		Resource: resource,
	})
	return err
}

func (t *Transaction) Prepare(ctx context.Context) error {
	_, err := t.ts.Prepare(ctx, &tcpb.TranRequest{
		Gid:       t.Gid,
		TransType: t.transType,
		Status:    tcpb.Status_Prepare,
	})
	return err
}

func (t *Transaction) Submit(ctx context.Context) error {
	_, err := t.ts.Submit(ctx, &tcpb.TranRequest{
		Gid:       t.Gid,
		TransType: t.transType,
		Status:    tcpb.Status_Submit,
	})
	return err
}

func (t *Transaction) Abort(ctx context.Context, reason error) error {
	_, err := t.ts.Abort(ctx, &tcpb.TranRequest{
		Gid:       t.Gid,
		TransType: t.transType,
		Status:    tcpb.Status_Abort,
		Reason:    reason.Error(),
	})
	return err
}

func (t *Transaction) BranchID() string {
	return t.opt.gen.ID()
}
