package imap_server

import (
	"github.com/Jinnrry/pmail/services/list"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// Search implements the IMAP SEARCH command with full criteria support
// Supports: UID, SeqNum, Date filters, Header search, Body/Text search,
// Flag filters, Size filters, and logical combinations (NOT, OR)
func (s *serverSession) Search(kind imapserver.NumKind, criteria *imap.SearchCriteria, options *imap.SearchOptions) (*imap.SearchData, error) {
	log.WithContext(s.ctx).Debugf("IMAP SEARCH: mailbox=%s, kind=%v, criteria=%+v", s.currentMailbox, kind, criteria)

	// Use the new comprehensive search function
	retList, err := list.SearchEmails(s.ctx, s.currentMailbox, criteria)
	if err != nil {
		log.WithContext(s.ctx).Errorf("IMAP SEARCH error: %v", err)
		return nil, err
	}

	ret := &imap.SearchData{}

	if kind == imapserver.NumKindSeq {
		// Return sequence numbers
		idList := imap.SeqSet{}
		for _, data := range retList {
			idList = append(idList, imap.SeqRange{
				Start: cast.ToUint32(data.SerialNumber),
				Stop:  cast.ToUint32(data.SerialNumber),
			})
		}
		ret.All = idList
		ret.Count = uint32(len(retList))

		// Handle ESEARCH options
		if options != nil {
			if options.ReturnMin && len(retList) > 0 {
				ret.Min = cast.ToUint32(retList[0].SerialNumber)
				for _, data := range retList {
					if cast.ToUint32(data.SerialNumber) < ret.Min {
						ret.Min = cast.ToUint32(data.SerialNumber)
					}
				}
			}
			if options.ReturnMax && len(retList) > 0 {
				ret.Max = cast.ToUint32(retList[0].SerialNumber)
				for _, data := range retList {
					if cast.ToUint32(data.SerialNumber) > ret.Max {
						ret.Max = cast.ToUint32(data.SerialNumber)
					}
				}
			}
		}
	} else {
		// Return UIDs
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

		// Handle ESEARCH options
		if options != nil {
			if options.ReturnMin && len(retList) > 0 {
				ret.Min = cast.ToUint32(retList[0].ID)
				for _, data := range retList {
					if cast.ToUint32(data.ID) < ret.Min {
						ret.Min = cast.ToUint32(data.ID)
					}
				}
			}
			if options.ReturnMax && len(retList) > 0 {
				ret.Max = cast.ToUint32(retList[0].ID)
				for _, data := range retList {
					if cast.ToUint32(data.ID) > ret.Max {
						ret.Max = cast.ToUint32(data.ID)
					}
				}
			}
		}
	}

	log.WithContext(s.ctx).Debugf("IMAP SEARCH result: count=%d", ret.Count)
	return ret, nil
}
