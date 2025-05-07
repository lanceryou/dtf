package main

import (
	"context"
	"fmt"
	"github.com/lanceryou/dtf/client"
	"github.com/lanceryou/dtf/client/tcc"
	hs "github.com/lanceryou/dtf/server/http"
	"github.com/lanceryou/dtf/server/tc"
	tccbranch "github.com/lanceryou/dtf/server/tc/tcc"
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
	trans := tcc.NewTcc(client.NewTransaction(client.NewHttpTransaction("http://localhost:8080/transaction"), "tcc"))
	if err := tcc.Transaction(context.TODO(), trans, func(ctx context.Context, t *tcc.Tcc) error {
		// 发起转入
		err := t.Try(ctx, &TransAmount{RequestId: request, Amount: amount}, map[string]string{"tcc": "http://localhost:8000/TransIn"})
		if err != nil {
			return err
		}
		// 转出
		return t.Try(ctx, &TransAmount{RequestId: request, Amount: amount}, map[string]string{"tcc": "http://localhost:8001/TransOut"})
	}); err != nil {
		slog.Error("Transaction err", err)
		return
	}

	if tm.amount != tmAmount || rm.amount != rmAmount {
		slog.Error("transaction some wrong.", "in amount", tm.amount, "out amount", rm.amount)
	}
	slog.Info("transfer amount success.", "in", tm.amount, "out", rm.amount)
}

// TransferAbNormal 模拟资源都预留成功，但是最后业务方处理发生错误
func TransferAbNormal(request string, amount int64, tmAmount, rmAmount int64) {
	trans := tcc.NewTcc(client.NewTransaction(client.NewHttpTransaction("http://localhost:8080/transaction"), "tcc"))
	err := tcc.Transaction(context.TODO(), trans, func(ctx context.Context, t *tcc.Tcc) error {
		// 发起转入
		err := t.Try(ctx, &TransAmount{RequestId: request, Amount: amount}, map[string]string{"tcc": "http://localhost:8000/TransIn"})
		if err != nil {
			return err
		}
		// 转出
		err = t.Try(ctx, &TransAmount{RequestId: request, Amount: amount}, map[string]string{"tcc": "http://localhost:8001/TransOut"})
		if err != nil {
			return err
		}
		return fmt.Errorf("exec failed")
	})
	if err != nil {
		slog.Error("Transaction err", err)
		return
	}

	if tm.amount != tmAmount || rm.amount != rmAmount {
		slog.Error("transaction some wrong.", "in amount", tm.amount, "out amount", rm.amount)
	}
	slog.Info("TransferAbNormal transfer amount failed.", "in", tm.amount, "out", rm.amount)
}

// TransferAbNormalOne 模拟部分资源预留成功
func TransferAbNormalOne(request string, amount int64, tmAmount, rmAmount int64) {
	trans := tcc.NewTcc(client.NewTransaction(client.NewHttpTransaction("http://localhost:8080/transaction"), "tcc"))
	err := tcc.Transaction(context.TODO(), trans, func(ctx context.Context, t *tcc.Tcc) error {
		// 发起转入
		err := t.Try(ctx, &TransAmount{RequestId: request, Amount: amount}, map[string]string{"tcc": "http://localhost:8000/TransIn"})
		if err != nil {
			return err
		}
		// 转出
		return t.Try(ctx, &TransAmount{RequestId: request, Amount: amount}, map[string]string{"tcc": "http://localhost:8001/TransOut"})
	})
	if err != nil {
		slog.Error("Transaction err", err)
		return
	}

	if tm.amount != tmAmount || rm.amount != rmAmount {
		slog.Error("transaction some wrong.", "in amount", tm.amount, "out amount", rm.amount)
	}
	slog.Info("TransferAbNormalOne transfer amount failed.", "in", tm.amount, "out", rm.amount)
}

// TransferNormalSubmitTimeout 模拟超时后TC驱动submit
// 业务第一次收到submit直接失败，等第二次收到在执行成功
func TransferNormalSubmitTimeout(request string, amount int64, tmAmount, rmAmount int64) {
	trans := tcc.NewTcc(client.NewTransaction(client.NewHttpTransaction("http://localhost:8080/transaction"), "tcc"))
	_ = tcc.Transaction(context.TODO(), trans, func(ctx context.Context, t *tcc.Tcc) error {
		// 发起转入 confirm 会返回失败
		err := t.Try(ctx, &TransAmount{RequestId: request, Amount: amount, Timeout: 1}, map[string]string{"tcc": "http://localhost:8000/TransIn"})
		if err != nil {
			return err
		}
		// 转出
		return t.Try(ctx, &TransAmount{RequestId: request, Amount: amount}, map[string]string{"tcc": "http://localhost:8001/TransOut"})
	})

	slog.Info("transaction timeout .", "amount", amount, "in amount", tm.amount, "out amount", rm.amount)
	time.Sleep(time.Second * 3)

	if tm.amount != tmAmount || rm.amount != rmAmount {
		slog.Error("transaction some wrong.", "in amount", tm.amount, "out amount", rm.amount)
	}
	slog.Info("TransferNormalTimeout transfer amount success.", "in", tm.amount, "out", rm.amount)
}

// TransferEmptyAbnormal 模拟空回滚
func TransferEmptyAbnormal(request string, amount int64, tmAmount, rmAmount int64) {
	ts := client.NewTransaction(client.NewHttpTransaction("http://localhost:8080/transaction"), "tcc")
	trans := tcc.NewTcc(ts)
	err := tcc.Transaction(context.TODO(), trans, func(ctx context.Context, t *tcc.Tcc) error {
		// 发起转入
		err := t.Try(ctx, &TransAmount{RequestId: request, Amount: amount}, map[string]string{"tcc": "http://localhost:8000/TransIn"})
		if err != nil {
			return err
		}
		// 转出
		return t.Try(ctx, &TransAmount{RequestId: request, Amount: amount}, map[string]string{"tcc": "http://localhost:8001/TransOut"})
	})
	if err != nil {
		slog.Error("Transaction err", err)
		return
	}

	if tm.amount != tmAmount || rm.amount != rmAmount {
		slog.Error("transaction some wrong.", "in amount", tm.amount, "out amount", rm.amount)
	}
	slog.Info("TransferEmptyAbnormal transfer amount failed.", "in", tm.amount, "out", rm.amount)
}

func main() {
	coord := tc.NewTC(tc.NewDurationTimer(time.Second), mysql.NewStore(gorm.NewDB(dsn)), 10)
	//coord := tc.NewTC(tc.NewDurationTimer(time.Second), store.NewMemoryStore())
	coord.RegisterBranchExecutor("tcc", tc.BranchFunc(tccbranch.HttpTccBranchExec))
	var server = hs.NewHttpTransaction(coord, &Locker{})
	server.Init()
	go server.Run(":8080")
	time.Sleep(time.Second)
	// 初始化tm和rm都是100
	TransferNormal("1", 10, 110, 90)
	// TransferAbNormal 模拟资源都预留成功，但是最后业务方处理发生错误
	TransferAbNormal("2", 10, 110, 90)
	// TransferAbNormalOne 模拟部分资源预留成功 框架触发回滚
	TransferAbNormalOne("3", 9, 110, 90)
	// 正常转账
	TransferNormal("4", 10, 120, 80)
	// 部分服务submit超时或者失败，等待tc定时执行
	TransferNormalSubmitTimeout("5", 10, 130, 70)
	// 模拟空回滚
	TransferEmptyAbnormal("null", 10, 130, 70)
}
