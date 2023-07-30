package dto

import (
	"context"
	"pmail/models"
)

const (
	LogID = "LogID"
)

type Context struct {
	context.Context
	UserInfo *models.User
	values   map[string]any
	Lang     string
}

func (c *Context) SetValue(key string, value any) {
	if c.values == nil {
		c.values = map[string]any{}
	}
	c.values[key] = value

}

func (c Context) GetValue(key string) any {
	if c.values == nil {
		return nil
	}
	return c.values[key]
}
