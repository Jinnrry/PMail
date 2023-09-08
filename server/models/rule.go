package models

import (
	"pmail/db"
	"pmail/utils/context"
	"pmail/utils/errors"
)

type Rule struct {
	Id     int    `db:"id" json:"id"`
	UserId string `db:"user_id" json:"user_id"`
	Name   string `db:"name" json:"name"`
	Value  string `db:"value" json:"value"`
	Action int    `db:"action" json:"action"`
	Params string `db:"params" json:"params"`
	Sort   int    `db:"sort" json:"sort"`
}

func (p *Rule) Save(ctx *context.Context) error {

	if p.Id > 0 {
		_, err := db.Instance.Exec(db.WithContext(ctx, "update rule set name=? ,value = ? ,action = ?,params = ?,sort = ? where id = ?"), p.Name, p.Value, p.Action, p.Params, p.Sort, p.Id)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	} else {
		_, err := db.Instance.Exec(db.WithContext(ctx, "insert into rule (name,value,user_id,action,params,sort) values (?,?,?,?,?,?)"), p.Name, p.Value, ctx.UserID, p.Action, p.Params, p.Sort)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	}

}
