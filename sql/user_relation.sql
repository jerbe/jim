SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

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

SET FOREIGN_KEY_CHECKS = 1;
