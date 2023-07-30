package models

type User struct {
	ID       int    `db:"id"`
	Account  string `db:"account"`
	Name     string `db:"name"`
	Password string `db:"password"`
}
