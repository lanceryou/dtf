package client

import (
	"context"
	"github.com/lanceryou/dtf/api/tccpb"
	hc "github.com/lanceryou/dtf/utils/http/client"
)

type Branch interface {
	ExecBranch(ctx context.Context, addr string, r interface{}) error
}

func NewHttpBranch() Branch {
	return &httpBranch{client: hc.NewHttpClient()}
}

type httpBranch struct {
	client *hc.HttpClient
}

func (h *httpBranch) ExecBranch(ctx context.Context, addr string, r interface{}) error {
	var resp tccpb.Empty
	return h.client.Post(ctx, addr, r, &resp)
}

var _ Branch = &httpBranch{}
