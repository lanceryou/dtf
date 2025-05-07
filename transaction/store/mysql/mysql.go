package mysql

import (
	"encoding/json"
	"github.com/lanceryou/dtf/transaction"
	"github.com/lanceryou/dtf/transaction/store"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Store mysql 存储事务
type Store struct {
	db *gorm.DB
}

func (s *Store) Save(ts ...*transaction.GlobalTransaction) error {
	var gs []*GlobalTransaction
	for _, trans := range ts {
		gs = append(gs, GlobalTransactionToEntity(trans))
	}
	return s.db.Table(GlobalTransactionTable).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "gid"}}, // key colume
		UpdateAll: true,
	}).Create(gs).Error
}

func (s *Store) Get(gid string) (*transaction.GlobalTransaction, error) {
	var gs GlobalTransaction
	err := s.db.Table(GlobalTransactionTable).Where("gid = ?", gid).First(&gs).Error
	if err != nil {
		return nil, err
	}
	ts := GlobalTransactionFromEntity(&gs)
	// 加载分支事务
	var bs []*BranchTransaction
	err = s.db.Table(BranchTransactionTable).Where("gid = ?", gid).Find(&bs).Error
	if err != nil {
		return nil, err
	}

	for _, branch := range bs {
		ts.Branches = append(ts.Branches, BranchTransactionFromEntity(branch))
	}
	return ts, nil
}

func (s *Store) Query(cond *store.TransactionCond) ([]*transaction.GlobalTransaction, error) {
	sess := s.db.Table(GlobalTransactionTable)
	if cond.Runtime != 0 {
		sess = sess.Where("next_run_time <= ?", cond.Runtime)
	}

	if len(cond.Status) != 0 {
		sess = sess.Where("status in ?", cond.Status)
	}

	if cond.Limit != 0 {
		sess = sess.Limit(cond.Limit).Offset(cond.Offset)
	}

	var gs []*GlobalTransaction
	if err := sess.Find(&gs).Error; err != nil {
		return nil, err
	}

	var gids []string
	gm := make(map[string]*transaction.GlobalTransaction)
	for _, tr := range gs {
		gids = append(gids, tr.Gid)
		gm[tr.Gid] = GlobalTransactionFromEntity(tr)
	}
	// 加载分支事务
	var bs []*BranchTransaction
	err := s.db.Table(BranchTransactionTable).Where("gid in ?", gids).Find(&bs).Error
	if err != nil {
		return nil, err
	}

	for _, branch := range bs {
		gm[branch.Gid].Branches = append(gm[branch.Gid].Branches, BranchTransactionFromEntity(branch))
	}

	var ret []*transaction.GlobalTransaction
	for _, v := range gm {
		ret = append(ret, v)
	}
	return ret, nil
}

func (s *Store) RegisterBranch(ts *transaction.BranchTransaction) error {
	return s.db.Table(BranchTransactionTable).Save(BranchTransactionToEntity(ts)).Error
}

var _ store.Store = &Store{}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func GlobalTransactionToEntity(ts *transaction.GlobalTransaction) *GlobalTransaction {
	return &GlobalTransaction{
		Gid:         ts.Gid,
		Status:      uint32(ts.Status),
		TransType:   ts.TransType,
		TransTime:   ts.TransTime,
		NextRunTime: ts.NextRunTime,
		Reason:      ts.Reason,
	}
}

func GlobalTransactionFromEntity(en *GlobalTransaction) *transaction.GlobalTransaction {
	return &transaction.GlobalTransaction{
		Gid:         en.Gid,
		Status:      transaction.Status(uint32(en.Status)),
		TransType:   en.TransType,
		TransTime:   en.TransTime,
		NextRunTime: en.NextRunTime,
		Reason:      en.Reason,
	}
}

func BranchTransactionToEntity(bs *transaction.BranchTransaction) *BranchTransaction {
	extra, err := json.Marshal(bs)
	if err != nil {
		panic(err)
	}
	return &BranchTransaction{
		Gid:       bs.Gid,
		BranchId:  bs.BranchId,
		TransTime: bs.TransTime,
		Reason:    bs.Reason,
		Extra:     string(extra),
	}
}

func BranchTransactionFromEntity(en *BranchTransaction) *transaction.BranchTransaction {
	var ret transaction.BranchTransaction
	if err := json.Unmarshal([]byte(en.Extra), &ret); err != nil {
		panic(err)
	}

	ret.Reason = en.Reason
	ret.TransTime = en.TransTime
	return &ret
}
