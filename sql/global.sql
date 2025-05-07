CREATE TABLE IF NOT EXISTS `global_transaction_tab` (
    `id`                   BIGINT UNSIGNED     NOT NULL AUTO_INCREMENT  comment '自增id',
    `gid`                  VARCHAR(512)        NOT NULL COMMENT '全局事务id',
    `status`               INT UNSIGNED        NOT NULL COMMENT '全局事务状态',
    `trans_time`           INT UNSIGNED        NOT NULL comment '事务时间',
    `trans_type`           VARCHAR(512)        NOT NULL COMMENT '事务类型',
    `next_run_time`        INT UNSIGNED        NOT NULL COMMENT '全局事务状态',
    `reason`               VARCHAR(512)        NOT NULL COMMENT '失败原因',
    `update_time`          timestamp           NOT NULL default  current_timestamp on update current_timestamp comment '更新时间',
    `create_time`          timestamp           NOT NULL default  current_timestamp comment '创建时间',
    PRIMARY KEY (`id`),
    Unique KEY (`gid`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC COMMENT '全局事务表';

CREATE TABLE IF NOT EXISTS `branch_transaction_tab` (
    `id`                   BIGINT UNSIGNED     NOT NULL AUTO_INCREMENT  comment '自增id',
    `gid`                  VARCHAR(512)        NOT NULL COMMENT '全局事务id',
    `branch_id`            VARCHAR(512)        NOT NULL COMMENT '分支事务id',
    `trans_time`           INT UNSIGNED        NOT NULL comment '事务时间',
    `extra`                TEXT                NOT NULL COMMENT '事务',
    `reason`               VARCHAR(512)        NOT NULL COMMENT '失败原因',
    `update_time`          timestamp           NOT NULL default  current_timestamp on update current_timestamp comment '更新时间',
    `create_time`          timestamp           NOT NULL default  current_timestamp comment '创建时间',
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC COMMENT '分支事务表';

CREATE TABLE IF NOT EXISTS `barrier_transaction_tab` (
    `id`                   BIGINT UNSIGNED     NOT NULL AUTO_INCREMENT  comment '自增id',
    `gid`                  VARCHAR(256)        NOT NULL COMMENT '全局事务id',
    `branch_id`            VARCHAR(256)        NOT NULL COMMENT '分支事务id',
    `op`                   VARCHAR(64)         NOT NULL comment '操作',
    `trans_type`           VARCHAR(64)         NOT NULL comment '事务类型',
    `update_time`          timestamp           NOT NULL default  current_timestamp on update current_timestamp comment '更新时间',
    `create_time`          timestamp           NOT NULL default  current_timestamp comment '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY (gid, branch_id, op)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC COMMENT '事务表';