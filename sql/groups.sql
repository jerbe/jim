SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

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

SET FOREIGN_KEY_CHECKS = 1;
