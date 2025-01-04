package imap_server

import (
	"github.com/Jinnrry/pmail/consts"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/services/group"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/emersion/go-imap/v2"
	"github.com/spf13/cast"
)

func (s *serverSession) Copy(numSet imap.NumSet, dest string) (*imap.CopyData, error) {

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

	if len(emailList) == 0 {
		return nil, &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: "Email Not Found",
		}
	}

	var err error
	destUid := []int{}
	UIDValidity := 0
	if group.IsDefaultBox(dest) {
		UIDValidity, destUid, err = copy2defaultbox(s.ctx, emailList, dest)
	} else {
		UIDValidity, destUid, err = copy2userbox(s.ctx, emailList, dest)
	}
	data := imap.CopyData{}
	data.UIDValidity = cast.ToUint32(UIDValidity)
	data.DestUIDs = imap.UIDSet{}
	data.SourceUIDs = imap.UIDSet{}
	for _, uid := range destUid {
		data.DestUIDs = append(data.DestUIDs, imap.UIDRange{Start: imap.UID(cast.ToUint32(uid)), Stop: imap.UID(cast.ToUint32(uid))})
	}

	for _, email := range emailList {
		data.SourceUIDs = append(data.SourceUIDs, imap.UIDRange{Start: imap.UID(cast.ToUint32(email.UeId)), Stop: imap.UID(cast.ToUint32(email.UeId))})
	}

	return &data, err
}

func copy2defaultbox(ctx *context.Context, mails []*response.EmailResponseData, dest string) (int, []int, error) {

	var destUid []int
	for _, email := range mails {

		newUe := models.UserEmail{
			UserID:  ctx.UserID,
			EmailID: email.Id,
			IsRead:  email.IsRead,
			GroupId: 0,
		}
		switch dest {
		case "Deleted Messages":
			newUe.Status = consts.EmailStatusDel
		case "INBOX":
			newUe.Status = consts.EmailStatusWait
		case "Sent Messages":
			newUe.Status = consts.EmailStatusSent
		case "Drafts":
			newUe.Status = consts.EmailStatusDrafts
		case "Junk":
			newUe.Status = consts.EmailStatusJunk
		}
		db.Instance.Insert(&newUe)
		destUid = append(destUid, newUe.ID)
	}

	return models.GroupNameToCode[dest], destUid, nil
}

func copy2userbox(ctx *context.Context, mails []*response.EmailResponseData, dest string) (int, []int, error) {
	groupInfo, err := group.GetGroupByFullPath(ctx, dest)
	if err != nil {
		return 0, nil, &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: err.Error(),
		}
	}
	if groupInfo == nil || groupInfo.ID == 0 {
		return 0, nil, &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: "Group not found",
		}
	}
	var destUid []int
	for _, email := range mails {
		newUe := models.UserEmail{
			UserID:  ctx.UserID,
			EmailID: email.Id,
			IsRead:  email.IsRead,
			GroupId: groupInfo.ID,
			Status:  email.Status,
		}
		db.Instance.Insert(&newUe)
		destUid = append(destUid, newUe.ID)
	}

	return groupInfo.ID, destUid, nil
}
