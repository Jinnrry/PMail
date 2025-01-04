package imap_server

import (
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

func (s *serverSession) Search(kind imapserver.NumKind, criteria *imap.SearchCriteria, options *imap.SearchOptions) (*imap.SearchData, error) {
	retList := []*response.UserEmailUIDData{}

	for _, uidSet := range criteria.UID {
		for _, uid := range uidSet {
			res := list.GetUEListByUID(s.ctx, s.currentMailbox, cast.ToInt(uid.Start), cast.ToInt(uid.Stop), nil)
			retList = append(retList, res...)
		}
	}
	ret := &imap.SearchData{}

	if kind == imapserver.NumKindSeq {
		idList := imap.SeqSet{}
		for _, data := range retList {
			log.WithContext(s.ctx).Debugf("Search Seq result: UID: %d  EmailID:%d", data.ID, data.EmailID)
			idList = append(idList, imap.SeqRange{
				Start: cast.ToUint32(data.SerialNumber),
				Stop:  cast.ToUint32(data.SerialNumber),
			})
		}
		ret.All = idList
	} else {
		idList := imap.UIDSet{}
		for _, data := range retList {
			log.WithContext(s.ctx).Debugf("Search UID result: UID: %d  EmailID:%d", data.ID, data.EmailID)

			idList = append(idList, imap.UIDRange{
				Start: imap.UID(data.ID),
				Stop:  imap.UID(data.ID),
			})
		}
		ret.All = idList
	}
	return ret, nil
}
