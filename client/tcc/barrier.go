package tcc

import (
	"github.com/lanceryou/dtf/api/tccpb"
	"github.com/lanceryou/dtf/client"
	"gorm.io/gorm"
)

func Barrier(db *gorm.DB) *client.Barrier {
	return client.NewBarrier(db, uint32(tccpb.TccOp_Try), uint32(tccpb.TccOp_Cancel), "tcc")
}
