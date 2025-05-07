package main

import (
	"context"
	"fmt"
	"github.com/lanceryou/dtf/client"
	"github.com/lanceryou/dtf/client/saga"
	hs "github.com/lanceryou/dtf/server/http"
	"github.com/lanceryou/dtf/server/tc"
	sagabranch "github.com/lanceryou/dtf/server/tc/saga"
	"github.com/lanceryou/dtf/transaction/store/mysql"
	"github.com/lanceryou/dtf/utils/gorm"
	"log/slog"
	"time"
)

var dsn = "root:@tcp(127.0.0.1:3306)/trans_db?charset=utf8mb4&parseTime=true&loc=Local"
var tm = NewAccount(dsn, ":8000", 100)
var rm = NewAccount(dsn, ":8001", 100)

// TransferNormal 正常转账
// 本地起两个服务模拟转账
// 其中一个作为TM发起分布式事务
func TransferNormal(request string, amount int64, tmAmount, rmAmount int64) {
	trans := saga.NewSaga(client.NewTransaction(client.NewHttpTransaction("http://localhost:8080/transaction"), "saga"))
	req := &TransAmount{RequestId: request, Amount: amount}
	if err := saga.Transaction(context.TODO(), trans, func(ctx context.Context, t *saga.Saga) error {
		// 发起转入
		_, err := t.AddCompensation(ctx, req, "http://localhost:8000/TransCancelIn")
		if err != nil {
			return err
		}
		_, err = tm.TransIn(ctx, req)
		if err != nil {
			return err
		}
		// 转出
		_, err = t.AddCompensation(ctx, req, "http://localhost:8001/TransCancelOut")
		if err != nil {
			return err
		}
		_, err = rm.TransOut(ctx, req)
		return err
	}); err != nil {
		slog.Error("Transaction err=", err)
		return
	}

	if tm.amount != tmAmount || rm.amount != rmAmount {
		slog.Error("transaction some wrong.", "in amount", tm.amount, "out amount", rm.amount)
	}
	slog.Info("transfer amount success.", "in", tm.amount, "out", rm.amount)
}

// TransferAbNormal 模拟都执行成功，业务最后返回失败
func TransferAbNormal(request string, amount int64, tmAmount, rmAmount int64) {
	trans := saga.NewSaga(client.NewTransaction(client.NewHttpTransaction("http://localhost:8080/transaction"), "saga"))
	req := &TransAmount{RequestId: request, Amount: amount}
	if err := saga.Transaction(context.TODO(), trans, func(ctx context.Context, t *saga.Saga) error {
		// 发起转入
		_, err := t.AddCompensation(ctx, req, "http://localhost:8000/TransCancelIn")
		if err != nil {
			return err
		}
		_, err = tm.TransIn(ctx, req)
		if err != nil {
			return err
		}
		// 转出
		_, err = t.AddCompensation(ctx, req, "http://localhost:8001/TransCancelOut")
		if err != nil {
			return err
		}
		// // 转出失败
		_, err = rm.TransOut(ctx, req)
		if err == nil {
			err = fmt.Errorf("somw error")
		}
		return err
	}); err != nil {
		slog.Error("Transaction err", err)
		return
	}

	if tm.amount != tmAmount || rm.amount != rmAmount {
		slog.Error("transaction some wrong.", "in amount", tm.amount, "out amount", rm.amount)
	}
	slog.Info("transfer amount failed.", "in", tm.amount, "out", rm.amount)
}

// TransferAbNormalOne 模拟都执行成功，业务最后返回失败
func TransferAbNormalOne(request string, amount int64, tmAmount, rmAmount int64) {
	trans := saga.NewSaga(client.NewTransaction(client.NewHttpTransaction("http://localhost:8080/transaction"), "saga"))
	req := &TransAmount{RequestId: request, Amount: amount}
	if err := saga.Transaction(context.TODO(), trans, func(ctx context.Context, t *saga.Saga) error {
		// 发起转入
		_, err := t.AddCompensation(ctx, req, "http://localhost:8000/TransCancelIn")
		if err != nil {
			return err
		}
		_, err = tm.TransIn(ctx, req)
		if err != nil {
			return err
		}
		// 转出
		_, err = t.AddCompensation(ctx, req, "http://localhost:8001/TransCancelOut")
		if err != nil {
			return err
		}
		// 转出失败
		_, err = rm.TransOut(ctx, req)
		if err == nil {
			err = fmt.Errorf("somw error")
		}
		return err
	}); err != nil {
		slog.Error("Transaction err", err)
		return
	}

	if tm.amount != tmAmount || rm.amount != rmAmount {
		slog.Error("transaction some wrong.", "in amount", tm.amount, "out amount", rm.amount)
	}
	slog.Info("TransferAbNormalOne transfer amount failed.", "in", tm.amount, "out", rm.amount)
}

func main() {
	coord := tc.NewTC(tc.NewDurationTimer(time.Second), mysql.NewStore(gorm.NewDB(dsn)), 10)
	//coord := tc.NewTC(tc.NewDurationTimer(time.Second), store.NewMemoryStore())
	coord.RegisterBranchExecutor("saga", tc.BranchFunc(sagabranch.HttpSagaBranchExec))
	var server = hs.NewHttpTransaction(coord, &Locker{})
	server.Init()
	go server.Run(":8080")
	time.Sleep(time.Second)
	// 初始化tm和rm都是100
	TransferNormal("1", 10, 110, 90)
	// 模拟失败回滚
	TransferAbNormal("2", 10, 110, 90)
	// 业务正常，但是最后返回失败
	TransferAbNormalOne("3", 9, 110, 90)
	// 正常转账
	TransferNormal("4", 10, 120, 80)
}
