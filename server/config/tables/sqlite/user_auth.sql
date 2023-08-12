CREATE TABLE user_auth
(
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       int ,
    email_account varchar(30)
);

CREATE UNIQUE INDEX udx_uid_ename on user_auth ( user_id, email_account);
CREATE UNIQUE INDEX udx_ename_uid on user_auth ( email_account,user_id );