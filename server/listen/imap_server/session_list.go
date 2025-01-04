package imap_server

import (
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	log "github.com/sirupsen/logrus"
	"strings"
)

func matchGroup(ctx *context.Context, w *imapserver.ListWriter, basePath, pattern string) {
	var groups []*models.Group
	if basePath == "" && pattern == "*" {
		db.Instance.Table("group").Where("user_id=?", ctx.UserID).Find(&groups)
		//w.WriteList(&imap.ListData{
		//	Attrs:   []imap.MailboxAttr{imap.MailboxAttrNoSelect, imap.MailboxAttrHasChildren},
		//	Delim:   '/',
		//	Mailbox: "[PMail]",
		//})
		w.WriteList(&imap.ListData{
			Attrs:   []imap.MailboxAttr{imap.MailboxAttrHasNoChildren},
			Delim:   '/',
			Mailbox: "INBOX",
		})
		w.WriteList(&imap.ListData{
			Attrs:   []imap.MailboxAttr{imap.MailboxAttrHasNoChildren, imap.MailboxAttrSent},
			Delim:   '/',
			Mailbox: "Sent Messages",
		})
		w.WriteList(&imap.ListData{
			Attrs:   []imap.MailboxAttr{imap.MailboxAttrHasNoChildren, imap.MailboxAttrDrafts},
			Delim:   '/',
			Mailbox: "Drafts",
		})

		w.WriteList(&imap.ListData{
			Attrs:   []imap.MailboxAttr{imap.MailboxAttrHasNoChildren, imap.MailboxAttrTrash},
			Delim:   '/',
			Mailbox: "Deleted Messages",
		})
		w.WriteList(&imap.ListData{
			Attrs:   []imap.MailboxAttr{imap.MailboxAttrHasNoChildren, imap.MailboxAttrJunk},
			Delim:   '/',
			Mailbox: "Junk",
		})
	} else {
		pattern = strings.ReplaceAll(pattern, "/*", "/%")

		db.Instance.Table("group").Where("user_id=? and full_path like ?", ctx.UserID, pattern).Find(&groups)

	}
	for _, group := range groups {

		data := &imap.ListData{
			Attrs:   []imap.MailboxAttr{},
			Mailbox: group.Name,
			Delim:   '/',
		}

		if hasChildren(ctx, group.ID) {
			data.Attrs = append(data.Attrs, imap.MailboxAttrHasChildren)
		}

		data.Mailbox = getLayerName(ctx, group, true)

		w.WriteList(data)

	}

}

func hasChildren(ctx *context.Context, id int) bool {
	var parent []*models.Group
	db.Instance.Table("group").Where("parent_id=?", id).Find(&parent)
	return len(parent) > 0
}
func getLayerName(ctx *context.Context, item *models.Group, allPath bool) string {
	if item.ParentId == 0 {
		return item.Name
	}
	var parent models.Group
	_, _ = db.Instance.Table("group").Where("id=?", item.ParentId).Get(&parent)
	if allPath {
		return getLayerName(ctx, &parent, allPath) + "/" + item.Name
	}
	return getLayerName(ctx, &parent, allPath)
}

func (s *serverSession) List(w *imapserver.ListWriter, ref string, patterns []string, options *imap.ListOptions) error {
	log.WithContext(s.ctx).Debugf("imap server list, ref: %s ,patterns: %s ", ref, patterns)

	if ref == "" && len(patterns) == 0 {
		w.WriteList(&imap.ListData{
			Attrs:   []imap.MailboxAttr{imap.MailboxAttrNoSelect, imap.MailboxAttrHasChildren},
			Delim:   '/',
			Mailbox: "[PMail]",
		})
	}
	for _, pattern := range patterns {
		matchGroup(s.ctx, w, ref, pattern)
	}

	return nil
}
