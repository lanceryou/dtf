package saga

import (
	"github.com/lanceryou/dtf/client"
	"gorm.io/gorm"
)

const (
	Prepare = iota + 1
	Compensation
)

func Barrier(db *gorm.DB) *client.Barrier {
	return client.NewBarrier(db, uint32(Prepare), uint32(Compensation), "saga")
}
