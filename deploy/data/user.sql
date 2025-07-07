CREATE DATABASE tiktok DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE `users` (
                         `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY COMMENT '用户ID，雪花算法生成',
                         `username` VARCHAR(64) NOT NULL UNIQUE COMMENT '用户名，唯一',
                         `password_hash` VARCHAR(255) NOT NULL COMMENT '加密后的密码',
                         `avatar` VARCHAR(255) DEFAULT NULL COMMENT '用户头像URL',
                         `background_image` VARCHAR(255) DEFAULT NULL COMMENT '个人页背景图',
                         `signature` VARCHAR(255) DEFAULT NULL COMMENT '个性签名',
                         `follow_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关注数',
                         `follower_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '粉丝数',
                         `work_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '作品数',
                         `favorite_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '喜欢数',
                         `total_favorited` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '获赞总数',
                         `tags` JSON DEFAULT NULL COMMENT '标签（AI画像使用）',
                         `status` TINYINT NOT NULL DEFAULT 1 COMMENT '账号状态：1正常，0封禁',
                         `extra` JSON DEFAULT NULL COMMENT '扩展字段，预留给未来功能',
                         `reserved1` VARCHAR(255) DEFAULT NULL COMMENT '预留字段1',
                         `reserved2` VARCHAR(255) DEFAULT NULL COMMENT '预留字段2',
                         `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                         `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                         `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户基础表';
