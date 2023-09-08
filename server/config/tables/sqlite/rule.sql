create table rule
(
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id int,
    name    varchar(255) default '' not null,
    value   json                    not null,
    action  int          default 0  not null,
    params  varchar(255) default '' not null,
    sort    int          default 0  not null
)