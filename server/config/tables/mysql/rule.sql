CREATE TABLE `rule`
(
    id      INT unsigned AUTO_INCREMENT PRIMARY KEY COMMENT '自增id',
    user_id int          NOT NULL DEFAULT 0 COMMENT '用户id',
    `name`  varchar(255) NOT NULL DEFAULT '' COMMENT '规则名称',
    `value` json         NOT NULL COMMENT '规则内容',
    action  int          not null default 0 comment '执行动作,1已读，2转发，3删除',
    params  varchar(255) not null default '' comment '执行参数',
    sort    int          not null default 0 COMMENT '排序，越大约优先'
) COMMENT '收信规则表'