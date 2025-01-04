package imap_server

import (
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/group"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/spf13/cast"
)

func (s *serverSession) Move(w *imapserver.MoveWriter, numSet imap.NumSet, dest string) error {

	var emailList []*response.EmailResponseData

	switch numSet.(type) {
	case imap.SeqSet:
		seqSet := numSet.(imap.SeqSet)
		for _, seq := range seqSet {
			emailList = list.GetEmailListByGroup(s.ctx, s.currentMailbox, list.ImapListReq{
				Star: cast.ToInt(seq.Start),
				End:  cast.ToInt(seq.Stop),
			}, false)
		}
	case imap.UIDSet:
		uidSet := numSet.(imap.UIDSet)
		for _, uid := range uidSet {
			emailList = list.GetEmailListByGroup(s.ctx, s.currentMailbox, list.ImapListReq{
				Star: cast.ToInt(uint32(uid.Start)),
				End:  cast.ToInt(uint32(uid.Stop)),
			}, true)
		}
	}

	var mailIds []int
	for _, email := range emailList {
		mailIds = append(mailIds, email.Id)
	}

	if group.IsDefaultBox(dest) {
		return move2defaultbox(s.ctx, mailIds, dest)
	} else {
		return move2userbox(s.ctx, mailIds, dest)
	}

}

func move2defaultbox(ctx *context.Context, mailIds []int, dest string) error {
	err := group.Move2DefaultBox(ctx, mailIds, dest)
	if err != nil {
		return &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: err.Error(),
		}
	}
	return nil
}

func move2userbox(ctx *context.Context, mailIds []int, dest string) error {
	groupInfo, err := group.GetGroupByFullPath(ctx, dest)
	if err != nil {
		return &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: err.Error(),
		}
	}
	if groupInfo == nil || groupInfo.ID == 0 {
		return &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: "Group not found",
		}
	}

	group.MoveMailToGroup(ctx, mailIds, groupInfo.ID)

	return nil
}
