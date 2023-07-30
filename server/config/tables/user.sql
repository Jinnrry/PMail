CREATE TABLE user
(
    id       INT unsigned AUTO_INCREMENT PRIMARY KEY COMMENT '自增id',
    account  varchar(20) COMMENT '账号登陆名',
    name     varchar(10) COMMENT '用户名',
    password char(32) COMMENT '登陆密码，两次md5加盐，md5(md5(password+"pmail") +"pmail2023")',
    UNIQUE INDEX udx_account ( account )
)COMMENT='登陆信息表'