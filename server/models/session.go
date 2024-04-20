package models

type Sessions struct {
	Token  string `xorm:"token char(43) not null pk " json:"token"`
	Data   string `xorm:"data blob" json:"data"`
	Expiry int    `xorm:"expiry timestamp index" json:"expiry"`
}

func (p *Sessions) TableName() string {
	return "sessions"
}
