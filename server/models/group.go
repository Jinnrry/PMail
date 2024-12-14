package models

type Group struct {
	ID       int    `xorm:"id int unsigned not null pk autoincr" json:"id"`
	Name     string `xorm:"varchar(10) notnull default('') comment('分组名称')" json:"name"`
	ParentId int    `xorm:"parent_id int unsigned notnull default(0) comment('父分组名称')" json:"parent_id"`
	UserId   int    `xorm:"user_id int unsigned notnull default(0) comment('用户id')" json:"-"`
}

const (
	INBOX   = 2000000000
	Sent    = 2000000001
	Drafts  = 2000000002
	Deleted = 2000000003
	Junk    = 2000000004
)

var GroupNameToCode = map[string]int{
	"INBOX":            INBOX,
	"Sent Messages":    Sent,
	"Drafts":           Drafts,
	"Deleted Messages": Deleted,
	"Junk":             Junk,
}

var GroupCodeToName = map[int]string{
	INBOX:   "INBOX",
	Sent:    "Sent Messages",
	Drafts:  "Drafts",
	Deleted: "Deleted Messages",
	Junk:    "Junk",
}

func (p *Group) TableName() string {
	return "group"
}
