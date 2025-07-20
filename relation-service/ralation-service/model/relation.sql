CREATE TABLE IF NOT EXISTS `relation` (
                                          `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                          `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                                          `to_user_id` BIGINT UNSIGNED NOT NULL COMMENT '关注用户的ID',
                                          `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '关注时间',
                                          `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                          `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
                                          PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_user_to_user` (`user_id`, `to_user_id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_to_user_id` (`to_user_id`),
    INDEX `idx_deleted_at` (`deleted_at`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户关系表';
