package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lanceryou/dtf/server"
	"github.com/lanceryou/dtf/server/tc"
	"github.com/lanceryou/dtf/utils/dlock"
	"github.com/lanceryou/dtf/utils/http/handler"
	"time"
)

func NewHttpTransaction(tc *tc.TC, locker dlock.DistLock) *HttpTransaction {
	return &HttpTransaction{Engine: gin.New(), TransactionServer: server.NewTransactionServer(tc, locker, time.Second)}
}

type HttpTransaction struct {
	*gin.Engine
	*server.TransactionServer
}

func (h *HttpTransaction) Init() {
	h.POST("/transaction/Prepare", handler.GinHandler(h.Prepare))
	h.POST("/transaction/Submit", handler.GinHandler(h.Submit))
	h.POST("/transaction/Abort", handler.GinHandler(h.Abort))
	h.POST("/transaction/RegisterBranch", handler.GinHandler(h.RegisterBranch))
}
