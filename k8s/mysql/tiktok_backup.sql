-- MySQL dump 10.13  Distrib 8.0.36, for Linux (x86_64)
--
-- Host: localhost    Database: tiktok
-- ------------------------------------------------------
-- Server version	8.0.36

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Current Database: `tiktok`
--

-- CREATE DATABASE /*!32312 IF NOT EXISTS*/ `tiktok` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;

-- USE `tiktok`;

--
-- Table structure for table `comment`
--

DROP TABLE IF EXISTS `comment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `comment` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT 'ID',
  `video_id` bigint unsigned NOT NULL COMMENT 'ID',
  `parent_id` bigint unsigned DEFAULT '0' COMMENT 'ID0',
  `content` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0-1-',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_video_id` (`video_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_video_created_at` (`video_id`,`created_at` DESC)
) ENGINE=InnoDB AUTO_INCREMENT=29482856371716100 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `comment`
--

LOCK TABLES `comment` WRITE;
/*!40000 ALTER TABLE `comment` DISABLE KEYS */;
INSERT INTO `comment` VALUES (27123354045513731,23359767250468865,26426378077339650,1,'创建成功了',0,'2025-07-07 02:46:18','2025-07-07 02:46:18'),(28900449444691971,23359767250468865,26426378077339650,1,'创建成功了，test',0,'2025-07-19 09:00:10','2025-07-19 09:00:10'),(28900948248100867,23359767250468865,26426378077339650,1,'创建成功了，test',0,'2025-07-19 09:05:07','2025-07-19 09:05:07'),(29033629334110211,23359767250468865,26426378077339650,1,'创建成功了，test2',0,'2025-07-20 07:03:11','2025-07-20 07:03:11'),(29035360407257091,23359767250468865,26426378077339650,1,'创建成功了，jaeger',0,'2025-07-20 07:20:23','2025-07-20 07:20:23'),(29035372990169091,23359767250468865,26426378077339650,1,'创建成功了，jaeger',0,'2025-07-20 07:20:30','2025-07-20 07:20:30'),(29448161021919235,23359767250468865,26426378077339650,1,'创建成功了111',0,'2025-07-23 03:41:11','2025-07-23 03:41:11'),(29482856371716099,23359767250468865,26426378077339650,1,'创建成功了111',0,'2025-07-23 09:25:51','2025-07-23 09:25:51');
/*!40000 ALTER TABLE `comment` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `favorite`
--

DROP TABLE IF EXISTS `favorite`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `favorite` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT 'ID',
  `video_id` bigint unsigned NOT NULL COMMENT 'ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_video` (`user_id`,`video_id`),
  KEY `idx_video_id` (`video_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `favorite`
--

LOCK TABLES `favorite` WRITE;
/*!40000 ALTER TABLE `favorite` DISABLE KEYS */;
INSERT INTO `favorite` VALUES (1,23359767250468865,23359767250468865,'2025-07-04 08:44:50','2025-07-04 08:44:50');
/*!40000 ALTER TABLE `favorite` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `relation`
--

DROP TABLE IF EXISTS `relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `relation` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT 'ID',
  `to_user_id` bigint unsigned NOT NULL COMMENT 'ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_user_to_user` (`user_id`,`to_user_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_to_user_id` (`to_user_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `relation`
--

LOCK TABLES `relation` WRITE;
/*!40000 ALTER TABLE `relation` DISABLE KEYS */;
INSERT INTO `relation` VALUES (1,23359767250468865,23360216309432321,'2025-07-15 02:29:17','2025-07-15 02:29:17',NULL),(2,23359767250468865,28145202665357313,'2025-07-15 02:30:17','2025-07-15 02:30:17',NULL),(3,23359767250468865,28171285146107905,'2025-07-15 10:14:30','2025-07-15 10:14:30',NULL),(4,23359767250468865,28172297852420097,'2025-07-15 10:24:31','2025-07-15 10:24:31',NULL),(5,23359767250468865,28172499816546305,'2025-07-16 02:17:59','2025-07-16 02:17:59',NULL),(6,23359767250468865,28172656146644993,'2025-07-16 02:25:35','2025-07-16 02:25:35',NULL);
/*!40000 ALTER TABLE `relation` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL COMMENT 'ID',
  `username` varchar(64) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `avatar` varchar(255) DEFAULT NULL COMMENT 'URL',
  `background_image` varchar(255) DEFAULT NULL,
  `signature` varchar(255) DEFAULT NULL,
  `follow_count` int unsigned NOT NULL DEFAULT '0',
  `follower_count` int unsigned NOT NULL DEFAULT '0',
  `work_count` int unsigned NOT NULL DEFAULT '0',
  `favorite_count` int unsigned NOT NULL DEFAULT '0',
  `total_favorited` int unsigned NOT NULL DEFAULT '0',
  `tags` json DEFAULT NULL COMMENT 'AI',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '10',
  `extra` json DEFAULT NULL,
  `reserved1` varchar(255) DEFAULT NULL COMMENT '1',
  `reserved2` varchar(255) DEFAULT NULL COMMENT '2',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (23359767250468865,'ysh','$2a$10$Bl1XPeQwPTIPoKJyePJ53Oylv.YpvCFZZVY.XNKRh8AHSdJBQA6u2','','','',6,0,0,0,0,'[]',1,'{}','','','2025-06-11 03:38:26','2025-07-16 10:25:36',NULL),(23360216309432321,'yk','$2a$10$pYgh0HwUKmrg4CLOi9L/MOtbY6DpGsJ./uQM3dtGjQkKRXtv9Gcvu','','','',0,1,0,0,0,'[]',1,'{}','','','2025-06-11 03:42:53','2025-07-15 10:29:18',NULL),(28145202665357313,'root1','$2a$10$zsiC.5AdKI6ipkL9nzU2Iu.uVB2Wwak3cQDvTME6ZgUVKnVOMro3u','','','',0,1,0,0,0,'[]',1,'{}','','','2025-07-14 03:57:28','2025-07-15 10:30:18',NULL),(28169117664018433,'root2','$2a$10$wApzQm4ENvZcro8NJZkRfe/8GuEVJyoe6gHAZ6nf7o/nb2/7.Kr8y','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 07:55:02','2025-07-14 07:55:02',NULL),(28171131483586561,'root3','$2a$10$5ede1m644pmS8k1rl6yfgectDW2qWl/D/yxoVhO5DtZe9sRaqVnMy','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 08:15:02','2025-07-14 08:15:02',NULL),(28171285146107905,'root4','$2a$10$aTZc1mZ2c11lFWOi8YBtfOIJywuWhAMW4CkEL6Vo.chNnN9sc9v5.','','','',0,1,0,0,0,'[]',1,'{}','','','2025-07-14 08:16:34','2025-07-15 18:14:31',NULL),(28172297852420097,'root5','$2a$10$n8SAV5aZLRPljXNzs7P/Mu1pjrUB6Zxj1MiKHgywfwb4Pky3KMG4O','','','',0,1,0,0,0,'[]',1,'{}','','','2025-07-14 08:26:38','2025-07-15 18:24:31',NULL),(28172499816546305,'root6','$2a$10$pfyDzynzN1mtkm8StP1be..DHV58r6pG8y/fD/srt/4qRiXjgjddy','','','',0,1,0,0,0,'[]',1,'{}','','','2025-07-14 08:28:38','2025-07-16 10:18:00',NULL),(28172656146644993,'root7','$2a$10$RyPqB0fi9Ppu605aomObHe3fjwEQ2PkKnbXZEo4HSpa1iXsy14xby','','','',0,1,0,0,0,'[]',1,'{}','','','2025-07-14 08:30:11','2025-07-16 10:25:36',NULL),(28173033919217665,'root8','$2a$10$iPhQg3DmOUg56LRqBkq6FeKniQXi.i0x344AQLm9VJbl5Sx3xo3/a','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 08:33:56','2025-07-14 08:33:56',NULL),(28175025760632833,'root9','$2a$10$lgEgGfrsD6oKaRILEbaZLeRSqEnaR2FlK2OKdSsALsVK7azkENo/S','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 08:53:44','2025-07-14 08:53:44',NULL),(28175197961977857,'root10','$2a$10$YSSZr.wJXaGPeme9LJwpheuRz3dyCw.An5D22Zeal66xfV75NUNnm','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 08:55:26','2025-07-14 08:55:26',NULL),(28175284733739009,'root11','$2a$10$914KZh1vRKIzskpNNQmfquQ3qtmBM/dHqC5Rlf19r.X.Me7KiRjnK','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 08:56:18','2025-07-14 08:56:18',NULL),(28177867670028289,'root12','$2a$10$Weu1.Or7k0R2.xEtP25xk.Ax0VF.QlPayOahadta9adcjhXWpWjuu','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 09:21:57','2025-07-14 09:21:57',NULL),(28178164911964161,'root13','$2a$10$jtWYTqDWIf87p7odXnsTNOYkcmGXpS0zDwuwupLcLhkU4e0JyitJW','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 09:24:55','2025-07-14 09:24:55',NULL),(28178660695474177,'root14','$2a$10$MLv8c.2NGMh14i5sVzhvHO2j27r/V75qU9Wdb7wrFKDz4j0AqX4ka','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 09:29:50','2025-07-14 09:29:50',NULL),(28179167803604993,'root15','$2a$10$W4C3ISsd0xRrAiwMzg.cUOwVEg/8YjWZm6IpIHWKk0gURlwD5TMoW','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 09:34:52','2025-07-14 09:34:52',NULL),(28179197264396289,'root16','$2a$10$KN4uJr.7mS0OfuTxQh9BmucNqfioeobhtIhA5hZb8RMONWS7gFkSu','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 09:35:10','2025-07-14 09:35:10',NULL),(28179742289035265,'root17','$2a$10$RFZPWmChSKDEg/bCt004N.sUg5o7MkQgg3CqxOFIZsRo1wKTgQwGi','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 09:40:35','2025-07-14 09:40:35',NULL),(28179825436917761,'root18','$2a$10$8wLaCEj5DtEAT6PkIhVdDeIkbT4FkReDqBAZGjQyoBieY7grJQpoe','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 09:41:24','2025-07-14 09:41:24',NULL),(28180044094373889,'root19','$2a$10$gXOpH4vYe7fDW9Wt6AeyMel7XRq2SKWmAw6L5LkfdNgU5dk1Zb2ZS','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 09:43:35','2025-07-14 09:43:35',NULL),(28180132191535105,'root20','$2a$10$YdL5xK8Io5G4QCQdgATPYuZBLYvpep4vxS1.UCHwEMpbzSPLSpSLa','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 09:44:27','2025-07-14 09:44:27',NULL),(28181459705528321,'root21','$2a$10$jV38hXKxOwh1hROyAOaL6.3LHqL4UPpjvxYK2hW.e.52FGpnltqmO','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 09:57:38','2025-07-14 09:57:38',NULL),(28181728862404609,'root22','$2a$10$pEmLdL6PKnx83K7C5LbmTuxEl0WQ6F1IXjGMvDdOEbOUow1VcfgmW','','','',0,0,0,0,0,'[]',1,'{}','','','2025-07-14 10:00:19','2025-07-14 10:00:19',NULL);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `videos`
--

DROP TABLE IF EXISTS `videos`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `videos` (
  `id` bigint unsigned NOT NULL COMMENT 'IDID',
  `user_id` bigint unsigned NOT NULL COMMENT 'ID',
  `play_url` text NOT NULL,
  `cover_url` text NOT NULL,
  `title` varchar(255) NOT NULL,
  `description` text,
  `duration` float DEFAULT '0',
  `tags` text,
  `favorite_cnt` int DEFAULT '0',
  `comment_cnt` int DEFAULT '0',
  `share_cnt` int DEFAULT '0',
  `collect_cnt` int DEFAULT '0',
  `is_public` tinyint(1) DEFAULT '1' COMMENT '01',
  `audit_status` tinyint DEFAULT '1' COMMENT '012',
  `is_original` tinyint(1) DEFAULT '1' COMMENT '10',
  `source_url` text,
  `transcode_status` tinyint DEFAULT '1' COMMENT '012',
  `video_width` int DEFAULT NULL,
  `video_height` int DEFAULT NULL,
  `biz_ext` json DEFAULT NULL,
  `reserved_1` varchar(255) DEFAULT NULL COMMENT '1',
  `reserved_2` varchar(255) DEFAULT NULL COMMENT '2',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `delete_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_user_time` (`user_id`,`created_at` DESC),
  KEY `idx_public_audit_time` (`is_public`,`audit_status`,`created_at` DESC),
  KEY `idx_favorite_cnt` (`favorite_cnt` DESC),
  KEY `idx_comment_cnt` (`comment_cnt` DESC),
  KEY `idx_share_cnt` (`share_cnt` DESC),
  KEY `idx_audit_status` (`audit_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `videos`
--

LOCK TABLES `videos` WRITE;
/*!40000 ALTER TABLE `videos` DISABLE KEYS */;
INSERT INTO `videos` VALUES (26426378077339650,23359767250468865,'http://127.0.0.1:9000/video-files/video/23359767250468865/test.mp4','','yjj的video','yjj光膀子',0,'',0,8,0,0,1,1,1,'',1,0,0,'{}','','','2025-07-02 07:22:28','2025-07-23 09:25:51',NULL),(28433855086067714,23359767250468865,'http://127.0.0.1:9000/video-files/video/23359767250468865/微信图片_20250716114200.jpg','','ysh旅游','ysh旅游',0,'',0,0,0,0,1,1,1,'',1,0,0,'{}','','','2025-07-16 03:44:58','2025-07-16 03:44:58',NULL),(29158139110621186,23359767250468865,'111111','','yshyd','yshyd',0,'',0,0,0,0,1,1,1,'',1,0,0,'{}','','','2025-07-21 03:40:05','2025-07-21 03:40:05',NULL);
/*!40000 ALTER TABLE `videos` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-07-25  9:01:32
