package imap_server

import (
	"github.com/Jinnrry/pmail/services/group"
	"github.com/emersion/go-imap/v2"
	"github.com/spf13/cast"
	"strings"
)

func (s *serverSession) Select(mailbox string, options *imap.SelectOptions) (*imap.SelectData, error) {
	if "" == mailbox {
		return nil, &imap.Error{
			Type: imap.StatusResponseTypeBad,
			Text: "mailbox not found",
		}
	}

	paths := strings.Split(mailbox, "/")
	s.currentMailbox = strings.Trim(paths[len(paths)-1], `"`)
	_, data := group.GetGroupStatus(s.ctx, s.currentMailbox, []string{"MESSAGES", "UNSEEN", "UIDNEXT", "UIDVALIDITY"})

	ret := &imap.SelectData{
		Flags:          []imap.Flag{imap.FlagSeen},
		PermanentFlags: []imap.Flag{imap.FlagSeen},
		NumMessages:    cast.ToUint32(data["MESSAGES"]),
		UIDNext:        imap.UID(data["UIDNEXT"]),
		UIDValidity:    cast.ToUint32(data["UIDVALIDITY"]),
	}

	return ret, nil

}
