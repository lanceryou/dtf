package client

import (
	"context"
	"github.com/lanceryou/dtf/api/tcpb"
	"github.com/lanceryou/dtf/transaction"
	hc "github.com/lanceryou/dtf/utils/http/client"
)

func NewHttpTransaction(addr string) *HttpTransaction {
	return &HttpTransaction{
		addr:   addr,
		client: hc.NewHttpClient(),
	}
}

// HttpTransaction http client 调用
type HttpTransaction struct {
	addr   string
	client *hc.HttpClient
}

func (h *HttpTransaction) Prepare(ctx context.Context, request *tcpb.TranRequest) (*tcpb.Empty, error) {
	var resp tcpb.Empty
	err := h.client.Post(ctx, h.addr+"/Prepare", request, &resp)
	return &resp, err
}

func (h *HttpTransaction) Submit(ctx context.Context, request *tcpb.TranRequest) (*tcpb.Empty, error) {
	var resp tcpb.Empty
	err := h.client.Post(ctx, h.addr+"/Submit", request, &resp)
	return &resp, err
}

func (h *HttpTransaction) Abort(ctx context.Context, request *tcpb.TranRequest) (*tcpb.Empty, error) {
	var resp tcpb.Empty
	err := h.client.Post(ctx, h.addr+"/Abort", request, &resp)
	return &resp, err
}

func (h *HttpTransaction) RegisterBranch(ctx context.Context, request *tcpb.BranchRequest) (*tcpb.Empty, error) {
	var resp tcpb.Empty
	err := h.client.Post(ctx, h.addr+"/RegisterBranch", request, &resp)
	return &resp, err
}

func (h *HttpTransaction) Protocol() string {
	return "http"
}

var _ transaction.Transaction = &HttpTransaction{}
