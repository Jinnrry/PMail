package context

import (
	"context"
)

const (
	LogID = "LogID"
)

type Context struct {
	context.Context
	UserID      int
	UserAccount string
	UserName    string
	values      map[string]any
	Lang        string
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
