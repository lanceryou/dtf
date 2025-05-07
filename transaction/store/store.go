package store

import (
	"github.com/lanceryou/dtf/transaction"
	"sync"
)

// Store 事务存储
type Store interface {
	Save(transaction ...*transaction.GlobalTransaction) error
	Get(gid string) (*transaction.GlobalTransaction, error)
	Query(*TransactionCond) ([]*transaction.GlobalTransaction, error)
	// RegisterBranch 注册分支事务
	RegisterBranch(transaction *transaction.BranchTransaction) error
}

type TransactionCond struct {
	Runtime int64
	Status  []uint32
	Limit   int
	Offset  int
}

var _ Store = &memoryStore{}

func NewMemoryStore() Store {
	return &memoryStore{ts: make(map[string]*transaction.GlobalTransaction)}
}

type memoryStore struct {
	sync.Mutex
	ts map[string]*transaction.GlobalTransaction
}

func (m *memoryStore) Save(gt ...*transaction.GlobalTransaction) error {
	m.Lock()
	for _, trans := range gt {
		m.ts[trans.Gid] = trans
	}
	m.Unlock()
	return nil
}

func (m *memoryStore) Get(gid string) (*transaction.GlobalTransaction, error) {
	m.Lock()
	defer m.Unlock()
	return m.ts[gid], nil
}

func (m *memoryStore) Query(cond *TransactionCond) ([]*transaction.GlobalTransaction, error) {
	m.Lock()
	defer m.Unlock()
	var result []*transaction.GlobalTransaction
	for _, gt := range m.ts {
		if cond.Runtime != 0 && gt.NextRunTime > cond.Runtime {
			continue
		}
		for _, status := range cond.Status {
			if gt.Status == transaction.Status(status) {
				result = append(result, gt)
			}
		}
	}
	return result, nil
}

func (m *memoryStore) RegisterBranch(bt *transaction.BranchTransaction) error {
	gt, err := m.Get(bt.Gid)
	if err != nil {
		return err
	}

	m.Lock()
	gt.Branches = append(gt.Branches, bt)
	m.Unlock()
	return nil
}
