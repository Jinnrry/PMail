package imap_server

import (
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/spf13/cast"
)

func (s *serverSession) Search(kind imapserver.NumKind, criteria *imap.SearchCriteria, options *imap.SearchOptions) (*imap.SearchData, error) {
	retList := []*response.UserEmailUIDData{}

	if len(criteria.UID) > 0 {
		for _, uidSet := range criteria.UID {
			for _, uid := range uidSet {
				res := list.GetUEListByUID(s.ctx, s.currentMailbox, cast.ToInt(uint32(uid.Start)), cast.ToInt(uint32(uid.Stop)), nil)
				retList = append(retList, res...)
			}
		}
	} else {
		res := list.GetUEListByUID(s.ctx, s.currentMailbox, 0, 0, nil)
		retList = append(retList, res...)
	}

	ret := &imap.SearchData{}

	if kind == imapserver.NumKindSeq {
		idList := imap.SeqSet{}
		for _, data := range retList {
			idList = append(idList, imap.SeqRange{
				Start: cast.ToUint32(data.SerialNumber),
				Stop:  cast.ToUint32(data.SerialNumber),
			})
		}
		ret.All = idList
		ret.Count = uint32(len(retList))
	} else {
		idList := imap.UIDSet{}
		for _, data := range retList {
			idList = append(idList, imap.UIDRange{
				Start: imap.UID(data.ID),
				Stop:  imap.UID(data.ID),
			})
		}
		ret.UID = true
		ret.All = idList
		ret.Count = uint32(len(retList))
	}
	return ret, nil
}
