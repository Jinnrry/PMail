package models

type Group struct {
	ID       int    `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	ParentId int    `db:"parent_id" json:"parent_id"`
	UserId   int    `db:"user_id" json:"-"`
}
