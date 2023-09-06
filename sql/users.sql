SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL DEFAULT '' COMMENT '用户名',
  `password_hash` varchar(100) NOT NULL DEFAULT '' COMMENT '密码',
  `nickname` varchar(50) NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar` varchar(100) NOT NULL DEFAULT '' COMMENT '头像地址',
  `birth_date` date DEFAULT NULL COMMENT '生日',
  `online_status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '0:离线,1:在线,2:离开,3:请勿打扰,4:隐身',
  `status` tinyint(1) unsigned NOT NULL DEFAULT 0 COMMENT '0:禁用,1:启用,2:已删除',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username_uidx` (`username`) USING BTREE COMMENT '用户名唯一索引',
  KEY `nickname_idx` (`nickname`) USING BTREE COMMENT '用户昵称索引'
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

SET FOREIGN_KEY_CHECKS = 1;
