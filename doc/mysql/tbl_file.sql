CREATE TABLE `tbl_file` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `file_sha1` char(40) NOT NULL DEFAULT '' COMMENT '文件hash',
    `file_name` varchar(256) NOT NULL DEFAULT '' COMMENT '文件名',
    `file_size` bigint(20) DEFAULT '0' COMMENT '文件大小',
    `file_addr` varchar(1024) NOT NULL DEFAULT '' COMMENT '文件存储位置',
    `create_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
    `update_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态（可用、禁用、已删除等）',
    `ext1` int(11) DEFAULT '0' COMMENT '备用字段1',
    `ext2` text COMMENT '备用字段2',
    PRIMARY KEY(`id`),
    UNIQUE KEY `idx_file_hash` (`file_sha1`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;