CREATE TABLE IF NOT EXISTS `favorite` (
                                          `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                          `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                                          `video_id` BIGINT UNSIGNED NOT NULL COMMENT '视频ID',
                                          `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '点赞时间',
                                          `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                          PRIMARY KEY (`id`),
    UNIQUE KEY `idx_user_video` (`user_id`, `video_id`),
    INDEX `idx_video_id` (`video_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户点赞表';
