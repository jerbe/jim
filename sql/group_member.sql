
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for group_member
-- ----------------------------
DROP TABLE IF EXISTS `group_member`;
CREATE TABLE `group_member` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `group_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '群ID编号',
  `user_id` int(10) unsigned NOT NULL COMMENT '用户ID',
  `role` tinyint(1) unsigned NOT NULL DEFAULT 0 COMMENT '群成员:0-普通成员;1-群主(owner);2-管理员; 一个群只能有一个群主(owner)',
  `speak_status` tinyint(1) unsigned NOT NULL DEFAULT 1 COMMENT '允许发言: 0-禁言,1-允许发言',
  `updater_id` int(10) unsigned NOT NULL COMMENT '更新人ID',
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '最后一次更新时间',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  `creator_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建人,也就是邀请人ID.',
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_idx` (`group_id`,`user_id`) USING BTREE COMMENT '群ID加用户ID唯一索引',
  KEY `fk_user_id` (`user_id`),
  CONSTRAINT `fk_group_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`),
  CONSTRAINT `fk_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

SET FOREIGN_KEY_CHECKS = 1;
