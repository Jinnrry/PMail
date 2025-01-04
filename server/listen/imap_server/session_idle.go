package imap_server

import (
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/spf13/cast"
	"sync"
)

var userConnects sync.Map

func (s *serverSession) Idle(w *imapserver.UpdateWriter, stop <-chan struct{}) error {
	connects, ok := userConnects.Load(s.ctx.UserID)
	logId := cast.ToString(s.ctx.GetValue(context.LogID))

	if !ok {

		connects = map[string]*imapserver.UpdateWriter{
			logId: w,
		}
		userConnects.Store(s.ctx.UserID, connects)
	} else {
		connects := connects.(map[string]*imapserver.UpdateWriter)
		if _, ok := connects[logId]; !ok {
			connects[logId] = w
			userConnects.Store(s.ctx.UserID, connects)
		}
	}

	go func() {
		<-stop
		userConnects.Delete(logId)
	}()

	return nil
}

func IdleNotice(ctx *context.Context, userId int, email *models.Email) error {
	if userId == 0 || email == nil || email.Id == 0 {
		return nil
	}

	connects, ok := userConnects.Load(userId)
	if ok {
		connects := connects.(map[string]*imapserver.UpdateWriter)
		for _, connect := range connects {
			connect.WriteNumMessages(1)
		}
	}
	return nil
}
