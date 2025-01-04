package imap_server

import (
	"github.com/Jinnrry/pmail/services/group"
	"github.com/emersion/go-imap/v2"
	"github.com/spf13/cast"
)

func (s *serverSession) Status(mailbox string, options *imap.StatusOptions) (*imap.StatusData, error) {
	category := []string{}
	if options.UIDNext {
		category = append(category, "UIDNEXT")
	}
	if options.NumMessages {
		category = append(category, "MESSAGES")
	}
	if options.UIDValidity {
		category = append(category, "UIDVALIDITY")
	}
	if options.NumUnseen {
		category = append(category, "UNSEEN")
	}

	_, data := group.GetGroupStatus(s.ctx, mailbox, category)

	numMessages := cast.ToUint32(data["MESSAGES"])
	numUnseen := cast.ToUint32(data["UNSEEN"])
	numValidity := cast.ToUint32(data["UIDVALIDITY"])
	numUIDNext := cast.ToUint32(data["UIDNEXT"])

	ret := &imap.StatusData{
		Mailbox:     mailbox,
		NumMessages: &numMessages,
		UIDNext:     imap.UID(numUIDNext),
		UIDValidity: numValidity,
		NumUnseen:   &numUnseen,
	}

	return ret, nil
}
