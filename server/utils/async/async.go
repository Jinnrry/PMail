package async

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"pmail/dto"
	"runtime/debug"
	"sync"
)

type Callback func()

type Async struct {
	wg        *sync.WaitGroup
	lastError error
	ctx       *dto.Context
}

func New(ctx *dto.Context) *Async {
	return &Async{
		ctx: ctx,
	}
}

func (as *Async) LastError() error {
	return as.lastError
}

func (as *Async) WaitProcess(callback Callback) {
	if as.wg == nil {
		as.wg = &sync.WaitGroup{}
	}
	as.wg.Add(1)
	as.Process(func() {
		defer as.wg.Done()
		callback()
	})
}

func (as *Async) Process(callback Callback) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				as.lastError = as.HandleErrRecover(err)
			}
		}()
		callback()
	}()
}

func (as *Async) Wait() {
	if as.wg == nil {
		return
	}
	as.wg.Wait()
}

// HandleErrRecover panic恢复处理
func (as *Async) HandleErrRecover(err interface{}) (returnErr error) {
	switch err.(type) {
	case error:
		returnErr = err.(error)
	default:
		returnErr = errors.New(cast.ToString(err))
	}

	log.WithContext(as.ctx).Errorf("goroutine panic:%s  \n %s", err, string(debug.Stack()))

	return
}
