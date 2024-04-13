package models

type UserAuth struct {
	ID           int    `xorm:"id int unsigned not null pk autoincr"`
	UserID       int    `xorm:"user_id int not null unique('uid_account') index comment('用户id')"`
	EmailAccount string `xorm:"email_account not null unique('uid_account') index comment('收信人前缀')"`
}

func (p UserAuth) TableName() string {
	return "user_auth"
}
