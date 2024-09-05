CREATE TABLE `tbl_user_token` (
    `id` INT(11) NOT NULL AUTO_INCREMENT,
    `user_name` varchar(64) NOT NULL DEFAULT '' COMMENT '用户名',
    `user_token` char(40) NOT NULL DEFAULT '' COMMENT 'token',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_username` (`user_name`)
)  ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;