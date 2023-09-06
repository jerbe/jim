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

-- ----------------------------
-- Table structure for groups
-- ----------------------------
DROP TABLE IF EXISTS `groups`;
CREATE TABLE `groups` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT '组名',
  `max_member` int(10) unsigned NOT NULL DEFAULT 100 COMMENT '群最大人数',
  `owner_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '该群的群主',
  `speak_status` tinyint(1) unsigned NOT NULL DEFAULT 1 COMMENT '发言状态:0-禁言,1-可发言',
  `creator_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  `updater_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '最后更新人ID',
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  KEY `id` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for user_relation
-- ----------------------------
DROP TABLE IF EXISTS `user_relation`;
CREATE TABLE `user_relation` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_a_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '用户A ID, 用户A一定比用户B小',
  `user_b_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '用户B ID, 用户B一定比用户A大',
  `status` tinyint(1) unsigned NOT NULL COMMENT '好友状态.\n0: A跟B都不是好友;\n1: A对B删除好友;\n2: B对A删除好友;\n3: 双方为好友.',
  `block_status` tinyint(1) NOT NULL COMMENT '拉黑状态.\n0: A跟B都拉黑了;\n1: A对B拉黑;\n2: B对A拉黑;\n3: 双方未拉黑.',
  `remark_on_a` varchar(30) NOT NULL DEFAULT '' COMMENT 'B对A的备注',
  `remark_on_b` varchar(30) NOT NULL DEFAULT '' COMMENT 'A对B的备注',
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '最后更新时间',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_uq_idx` (`user_a_id`,`user_b_id`) USING BTREE COMMENT '双方好友唯一标识',
  KEY `fk_user_b_b` (`user_b_id`),
  CONSTRAINT `fk_user_a_b` FOREIGN KEY (`user_a_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_user_b_b` FOREIGN KEY (`user_b_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

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
