package models

type Group struct {
	ID       int    `xorm:"id int unsigned not null pk autoincr" json:"id"`
	Name     string `xorm:"varchar(10) notnull default('') comment('分组名称')" json:"name"`
	ParentId int    `xorm:"parent_id int unsigned notnull default(0) comment('父分组名称')" json:"parent_id"`
	UserId   int    `xorm:"user_id int unsigned notnull default(0) comment('用户id')" json:"-"`
}

func (p *Group) TableName() string {
	return "group"
}
