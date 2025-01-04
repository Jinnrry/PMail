package goimap

import (
	"context"
	"net"
	"time"
)

type Status int8

const (
	UNAUTHORIZED Status = 1
	AUTHORIZED   Status = 2
	SELECTED     Status = 3
	LOGOUT       Status = 4
)

type Session struct {
	Status      Status
	Account     string
	DeleteIds   []int64
	Ctx         context.Context
	Conn        net.Conn
	InTls       bool
	AliveTime   time.Time
	CurrentPath string //当前选择的文件夹
	IN_IDLE     bool   // 是否处在IDLE中
}
