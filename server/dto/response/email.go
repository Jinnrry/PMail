package response

import "pmail/models"

type EmailResponseData struct {
	models.Email `xorm:"extends"`
	IsRead       int8 `json:"is_read"`
}
