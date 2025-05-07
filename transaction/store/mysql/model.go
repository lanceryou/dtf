package mysql

const (
	GlobalTransactionTable = "global_transaction_tab"
	BranchTransactionTable = "branch_transaction_tab"
)

type GlobalTransaction struct {
	Id          uint64 `gorm:"column:id;type:bigint(20) unsigned AUTO_INCREMENT;primaryKey;precision:20;scale:0;not null;autoIncrement"`
	Gid         string `gorm:"column:gid;type:varchar(512);size:512;not null"`                           // 全局事务id
	Status      uint32 `gorm:"column:status;type:int(10) unsigned;precision:10;scale:0;not null"`        // 全局事务状态
	TransTime   int64  `gorm:"column:trans_time;type:int(10) unsigned;precision:10;scale:0;not null"`    // 事务发生时间
	TransType   string `gorm:"column:trans_type;type:varchar(512);size:512;not null"`                    // 事务类型 tcc, saga...
	NextRunTime int64  `gorm:"column:next_run_time;type:int(10) unsigned;precision:10;scale:0;not null"` // 下次执行时间
	Reason      string `gorm:"column:reason;type:varchar(512);size:512;not null"`                        // 失败原因
}

type BranchTransaction struct {
	Id        uint64 `gorm:"column:id;type:bigint(20) unsigned AUTO_INCREMENT;primaryKey;precision:20;scale:0;not null;autoIncrement"`
	Gid       string `gorm:"column:gid;type:varchar(512);size:512;not null"`                        // 全局事务id
	BranchId  string `gorm:"column:branch_id;type:varchar(512);size:512;not null"`                  // branch id
	TransTime int64  `gorm:"column:trans_time;type:int(10) unsigned;precision:10;scale:0;not null"` // 事务发生时间
	Reason    string `gorm:"column:reason;type:varchar(512);size:512;not null"`                     // 失败原因
	Extra     string `gorm:"column:extra;type:varchar(512);size:512;not null"`                      // 序列化后的分支信息
}
