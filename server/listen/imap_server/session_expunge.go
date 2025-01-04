package imap_server

import (
	"github.com/Jinnrry/pmail/services/del_email"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/spf13/cast"
)

func (s *serverSession) Expunge(w *imapserver.ExpungeWriter, uids *imap.UIDSet) error {
	if uids == nil {
		return nil
	}

	uidList := []int{}
	for _, uidRange := range *uids {
		if uidRange.Start > 0 && uidRange.Stop > 0 {
			for i := uidRange.Start; i <= uidRange.Stop; i++ {
				uidList = append(uidList, cast.ToInt(uint32(i)))
			}
		}
	}

	err := del_email.DelByUID(s.ctx, uidList)
	if err != nil {
		return &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: err.Error(),
		}
	}

	return nil
}
