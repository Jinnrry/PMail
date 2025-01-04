package response

import "github.com/Jinnrry/pmail/models"

type EmailResponseData struct {
	models.Email `xorm:"extends"`
	IsRead       int8 `json:"is_read"`
	SerialNumber int  `json:"serial_number"`
	UeId         int  `json:"ue_id"`
}

type UserEmailUIDData struct {
	models.UserEmail `xorm:"extends"`
	SerialNumber     int `json:"serial_number"`
}
