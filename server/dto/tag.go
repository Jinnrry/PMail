package dto

import "encoding/json"

type SearchTag struct {
	Type   int `json:"type"`
	Status int `json:"status"`
}

func (t SearchTag) ToString() string {
	data, _ := json.Marshal(t)
	return string(data)
}
