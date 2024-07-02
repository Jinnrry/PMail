package context

import (
	"context"
)

const (
	LogID = "LogID"
)

type Context struct {
	context.Context `json:"-"`
	UserID          int
	UserAccount     string
	UserName        string
	Values          map[string]any
	Lang            string
	IsAdmin         bool
}

func (c *Context) SetValue(key string, value any) {
	if c.Values == nil {
		c.Values = map[string]any{}
	}
	c.Values[key] = value

}

func (c *Context) GetValue(key string) any {
	if c.Values == nil {
		return nil
	}
	return c.Values[key]
}
