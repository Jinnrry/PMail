CREATE TABLE user
(
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    account  varchar(20),
    name     varchar(10),
    password char(32)
);
CREATE UNIQUE INDEX udx_account on user (account);