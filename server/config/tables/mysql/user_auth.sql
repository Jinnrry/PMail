CREATE TABLE user_auth
(
    id            INT unsigned AUTO_INCREMENT PRIMARY KEY COMMENT '自增id',
    user_id       int COMMENT '用户id',
    email_account varchar(30) COMMENT '收件人前缀',
    UNIQUE INDEX udx_uid_ename ( user_id, email_account),
    UNIQUE INDEX udx_ename_uid ( email_account,user_id )
)COMMENT='登陆信息表'