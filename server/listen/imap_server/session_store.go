package imap_server

import (
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

	if !array.InArray(imap.FlagSeen, flags.Flags) {
		return nil
	}

	switch numSet.(type) {
	case imap.SeqSet:
		seqSet := numSet.(imap.SeqSet)
		for _, seq := range seqSet {
			emailList := list.GetEmailListByGroup(s.ctx, s.currentMailbox, list.ImapListReq{
				Star: cast.ToInt(seq.Start),
				End:  cast.ToInt(seq.Stop),
			}, false)
			for _, data := range emailList {
				detail.MakeRead(s.ctx, data.Id, flags.Op == imap.StoreFlagsAdd)
			}
		}

	case imap.UIDSet:
		uidSet := numSet.(imap.UIDSet)
		for _, uid := range uidSet {
			emailList := list.GetEmailListByGroup(s.ctx, s.currentMailbox, list.ImapListReq{
				Star: cast.ToInt(uint32(uid.Start)),
				End:  cast.ToInt(uint32(uid.Stop)),
			}, true)
			for _, data := range emailList {
				detail.MakeRead(s.ctx, data.Id, flags.Op == imap.StoreFlagsAdd)
			}
		}
	}
	return nil
}
