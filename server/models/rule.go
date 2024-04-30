package models

type Rule struct {
	Id     int    `xorm:"id int unsigned not null pk autoincr" json:"id"`
	UserId int    `xorm:"user_id notnull default(0) comment('用户id')" json:"user_id"`
	Name   string `xorm:"name notnull default('') comment('规则名称')" json:"name"`
	Value  string `xorm:"value text comment('规则内容')" json:"value"`
	Action int    `xorm:"action notnull default(0) comment('执行动作,1已读，2转发，3删除')" json:"action"`
	Params string `xorm:"params notnull default('') comment('执行参数')" json:"params"`
	Sort   int    `xorm:"sort notnull default(0) comment('排序，越大约优先')" json:"sort"`
}

func (p *Rule) TableName() string {
	return "rule"
}
