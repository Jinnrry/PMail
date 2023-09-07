package dto

import (
	"encoding/json"
	"pmail/models"
)

type RuleType int

// 1已读，2转发，3删除
var (
	READ    RuleType = 1
	FORWARD RuleType = 2
	DELETE  RuleType = 3
	MOVE    RuleType = 4
)

type Rule struct {
	Id     int      `json:"id"`
	Name   string   `json:"name"`
	Rules  []*Value `json:"rules"`
	Action RuleType `json:"action"`
	Params string   `json:"params"`
	Sort   int      `json:"sort"`
}

type Value struct {
	Field string `json:"field"`
	Type  string `json:"type"`
	Rule  string `json:"rule"`
}

func (p *Rule) Decode(data *models.Rule) *Rule {
	json.Unmarshal([]byte(data.Value), &p.Rules)
	p.Id = data.Id
	p.Name = data.Name
	p.Action = RuleType(data.Action)
	p.Sort = data.Sort
	p.Params = data.Params
	return p
}

func (p *Rule) Encode() *models.Rule {
	v, _ := json.Marshal(p.Rules)
	ret := &models.Rule{
		Id:     p.Id,
		Name:   p.Name,
		Value:  string(v),
		Action: int(p.Action),
		Sort:   p.Sort,
		Params: p.Params,
	}
	return ret
}
