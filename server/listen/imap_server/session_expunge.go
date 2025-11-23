package imap_server

import (
	"github.com/Jinnrry/pmail/services/del_email"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/spf13/cast"
	"log/slog"
)

func (s *serverSession) Expunge(w *imapserver.ExpungeWriter, uids *imap.UIDSet) error {
	if uids == nil && len(s.deleteUidList) == 0 {
		return nil
	}
	uidList := []int{}

	if uids != nil {
		for _, uidRange := range *uids {
			if uidRange.Start > 0 && uidRange.Stop > 0 {
				for i := uidRange.Start; i <= uidRange.Stop; i++ {
					uidList = append(uidList, cast.ToInt(uint32(i)))
				}
			}
		}
	}

	if len(s.deleteUidList) > 0 {
		uidList = append(uidList, s.deleteUidList...)
	}

	if len(uidList) == 0 {
		return nil
	}

	slog.Debug("DeleteUidList:", slog.Any("uidList", uidList))

	err := del_email.DelByUID(s.ctx, uidList)
	s.deleteUidList = []int{}
	if err != nil {
		return &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: err.Error(),
		}
	}

	return nil
}
