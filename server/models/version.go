package models

type Version struct {
	Id   int    `xorm:"id int unsigned not null pk autoincr" json:"id"`
	Info string `xorm:"varchar(255) notnull" json:"info"`
}

func (p *Version) TableName() string {
	return "version"
}
