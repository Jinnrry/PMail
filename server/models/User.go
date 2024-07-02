package models

type User struct {
	ID       int    `xorm:"id unsigned int not null pk autoincr"`
	Account  string `xorm:"varchar(20) notnull unique comment('账号登陆名')"`
	Name     string `xorm:"varchar(10) notnull comment('用户名')"`
	Password string `xorm:"char(32) notnull comment('登陆密码，两次md5加盐，md5(md5(password+pmail) +pmail2023)')" json:"-"`
	Disabled int    `xorm:"disabled unsigned int not null default(0) comment('0启用，1禁用')"`
	IsAdmin  int    `xorm:"is_admin unsigned int not null default(0) comment('0不是管理员，1是管理员')"`
}

func (p User) TableName() string {
	return "user"
}
