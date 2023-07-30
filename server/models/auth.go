package models

type UserAuth struct {
	ID           int    `db:"id"`
	UserID       int    `db:"user_id"`
	EmailAccount string `db:"email_account"`
}
