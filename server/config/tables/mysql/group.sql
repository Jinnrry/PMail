CREATE TABLE `group`
(
    id        INT unsigned AUTO_INCREMENT PRIMARY KEY COMMENT '自增id',
    name      varchar(10) NOT NULL DEFAULT '' COMMENT '分组名称',
    user_id   INT unsigned NOT NULL DEFAULT 0 COMMENT '用户id',
    parent_id INT unsigned COMMENT '父级组ID'
)COMMENT='分组信息表'