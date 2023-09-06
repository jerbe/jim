
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for user_relation_invite
-- ----------------------------
DROP TABLE IF EXISTS `user_relation_invite`;
CREATE TABLE `user_relation_invite` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '用户id',
  `target_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '目标用户id',
  `note` varchar(100) NOT NULL DEFAULT '' COMMENT '申请注释',
  `reply` varchar(100) NOT NULL DEFAULT '' COMMENT '回复',
  `status` tinyint(1) unsigned NOT NULL DEFAULT 0 COMMENT '申请状态',
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '更新时间',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  `uq_flag` varchar(50) NOT NULL DEFAULT '0' COMMENT '控制记录唯一值的标记,与user_id,target_id联合',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_flag_idx` (`uq_flag`) USING BTREE,
  KEY `user_target_idx` (`user_id`,`target_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

SET FOREIGN_KEY_CHECKS = 1;
