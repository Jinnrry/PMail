package imap_server

import (
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/detail"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/spf13/cast"
)

func (s *serverSession) Store(w *imapserver.FetchWriter, numSet imap.NumSet, flags *imap.StoreFlags, options *imap.StoreOptions) error {

	if flags.Op == imap.StoreFlagsSet {
		return nil
	}

	var emailList []*response.EmailResponseData

	switch numSet.(type) {
	case imap.SeqSet:
		seqSet := numSet.(imap.SeqSet)
		for _, seq := range seqSet {
			res := list.GetEmailListByGroup(s.ctx, s.currentMailbox, list.ImapListReq{
				Star: cast.ToInt(seq.Start),
				End:  cast.ToInt(seq.Stop),
			}, false)
			emailList = append(emailList, res...)
		}

	case imap.UIDSet:
		uidSet := numSet.(imap.UIDSet)
		for _, uid := range uidSet {
			res := list.GetEmailListByGroup(s.ctx, s.currentMailbox, list.ImapListReq{
				Star: cast.ToInt(uint32(uid.Start)),
				End:  cast.ToInt(uint32(uid.Stop)),
			}, true)
			emailList = append(emailList, res...)
		}
	}

	if array.InArray(imap.FlagSeen, flags.Flags) && flags.Op == imap.StoreFlagsAdd {
		for _, data := range emailList {
			detail.MakeRead(s.ctx, data.Id, flags.Op == imap.StoreFlagsAdd)
		}
	}

	if array.InArray(imap.FlagDeleted, flags.Flags) && flags.Op == imap.StoreFlagsAdd {
		for _, data := range emailList {
			s.deleteUidList = append(s.deleteUidList, data.UeId)
		}
	}

	return nil
}
