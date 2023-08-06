CREATE TABLE sessions
(
    token  CHAR(43) PRIMARY KEY,
    data   BLOB         NOT NULL,
    expiry TIMESTAMP(6) NOT NULL,
    KEY    `sessions_expiry_idx` (`expiry`)
)COMMENT='系统session数据表';
