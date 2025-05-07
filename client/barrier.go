package client

import (
	"gorm.io/gorm"
	"log/slog"
)

const (
	BarrierTab = "barrier_transaction_tab"
)

// BarrierModel gid branch_id op 构造唯一键
type BarrierModel struct {
	Id        uint64 `gorm:"column:id;type:bigint(20) unsigned AUTO_INCREMENT;primaryKey;precision:20;scale:0;not null;autoIncrement"`
	Gid       string `gorm:"column:gid;type:varchar(512);size:512;not null"`                // 全局事务id
	BranchId  string `gorm:"column:branch_id;type:varchar(512);size:512;not null"`          // branch id
	Op        uint32 `gorm:"column:op;type:int(10) unsigned;precision:10;scale:0;not null"` // 操作 try, confirm, cancel...
	TransType string `gorm:"column:trans_type;type:varchar(512);size:512;not null"`         // 事务类型 tcc, saga...
}

func NewBarrier(db *gorm.DB, pop uint32, rollop uint32, transType string) *Barrier {
	return &Barrier{db: db, PrepareOp: pop, RollOp: rollop, transType: transType}
}

// Barrier 方便处理空回滚，空悬挂问题
// 空回滚 业务没有收到try（对应TCC），直接收到cancel
// 解决思路 假如当前请求是cancel 但是数据库没有try，直接返回成功
// 空悬挂 业务收到cancel后，try因为超时才到
// 解决思路 当前请求是try，发现已经存在cancel请求，直接返回成功
type Barrier struct {
	db        *gorm.DB
	PrepareOp uint32
	RollOp    uint32
	transType string
}

func (b *Barrier) Run(gid, branchId string, op uint32, fn func() error) error {
	// 需要跑业务先插入一条barrier记录
	return b.db.Transaction(func(tx *gorm.DB) error {
		runnable, err := b.barrierRun(tx, gid, branchId, op)
		if err != nil || !runnable {
			return err
		}
		err = tx.Table(BarrierTab).Create(&BarrierModel{
			Gid:       gid,
			BranchId:  branchId,
			Op:        op,
			TransType: b.transType,
		}).Error
		if err != nil {
			return err
		}
		return fn()
	})
}

func (b *Barrier) barrierRun(tx *gorm.DB, gid, branchId string, op uint32) (bool, error) {
	//  不需要处理的状态
	if op != b.PrepareOp && op != b.RollOp {
		return true, nil
	}
	// 处理空悬挂
	if op == b.PrepareOp {
		return b.handleDangledRequest(tx, gid, branchId, op)
	}

	// 空回滚
	return b.handleEmptyCompensate(tx, gid, branchId, op)
}

func (b *Barrier) handleDangledRequest(tx *gorm.DB, gid, branchId string, op uint32) (bool, error) {
	// 处理空悬挂
	// 查找补偿数据，插入当前数据
	var cnt int64
	err := tx.Table(BarrierTab).Where("gid = ? AND branch_id = ? AND op = ?", gid, branchId, b.RollOp).Count(&cnt).Error
	if err != nil {
		return false, err
	}
	// 存在回滚数据, 说明产生空悬挂 则不需要跑
	slog.Info("barrier handleDangledRequest", "gid", gid, "branch_id", branchId, "op", op, "is_dangle", cnt != 0)
	return cnt == 0, nil
}

func (b *Barrier) handleEmptyCompensate(tx *gorm.DB, gid, branchId string, op uint32) (bool, error) {
	var cnt int64
	err := tx.Table(BarrierTab).Where("gid = ? AND branch_id = ? AND op = ?", gid, branchId, b.PrepareOp).Count(&cnt).Error
	if err != nil {
		slog.Info("barrier handleEmptyCompensate", "gid", gid, "branch_id", branchId, "op", op, "err", err)
		return false, err
	}
	// 数据库没有try操作, 说明产生空回滚
	slog.Info("barrier handleEmptyCompensate", "gid", gid, "branch_id", branchId, "op", op, "is_empty", cnt == 0)
	return cnt != 0, nil
}
