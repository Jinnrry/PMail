CREATE TABLE `group`
(
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    name      varchar(10) NOT NULL DEFAULT '',
    parent_id INTEGER     NOT NULL DEFAULT 0,
    user_id   INTEGER     NOT NULL DEFAULT 0
)