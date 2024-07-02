package dto

import "encoding/json"

type SearchTag struct {
	Type    int8 `json:"type"`     // -1 不限
	Status  int8 `json:"status"`   // -1 不限
	GroupId int  `json:"group_id"` // -1 不限
}

func (t SearchTag) ToString() string {
	data, _ := json.Marshal(t)
	return string(data)
}
