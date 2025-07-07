CREATE TABLE IF NOT EXISTS `comment` (
                                         `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                         `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                                         `video_id` BIGINT UNSIGNED NOT NULL COMMENT '视频ID',
                                         `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '父评论ID，0表示一级评论',
                                         `content` TEXT NOT NULL COMMENT '评论内容',
                                         `is_deleted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除：0-未删除，1-已删除',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '评论时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_video_id` (`video_id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_video_created_at` (`video_id`, `created_at` DESC)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='视频评论表';
