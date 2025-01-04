package imap_server

import (
	"github.com/Jinnrry/pmail/services/group"
	"github.com/emersion/go-imap/v2"
	"strings"
)

func (s *serverSession) Create(mailbox string, options *imap.CreateOptions) error {
	groupPath := strings.Split(mailbox, "/")

	var parentId int
	for _, path := range groupPath {
		newGroup, err := group.CreateGroup(s.ctx, path, parentId)
		if err != nil {
			return &imap.Error{
				Type: imap.StatusResponseTypeNo,
				Text: err.Error(),
			}
		}
		parentId = newGroup.ID
	}

	return nil
}
