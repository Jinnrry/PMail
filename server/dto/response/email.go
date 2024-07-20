package response

import "github.com/Jinnrry/pmail/models"

type EmailResponseData struct {
	models.Email `xorm:"extends"`
	IsRead       int8 `json:"is_read"`
}
