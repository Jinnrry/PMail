package imap_server

import (
	"bytes"
	"fmt"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/detail"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/spf13/cast"
	"mime"
	"strings"
	"time"
)

func (s *serverSession) Fetch(w *imapserver.FetchWriter, numSet imap.NumSet, options *imap.FetchOptions) error {
	switch numSet.(type) {
	case imap.SeqSet:
		seqSet := numSet.(imap.SeqSet)
		for _, seq := range seqSet {
			emailList := list.GetEmailListByGroup(s.ctx, s.currentMailbox, list.ImapListReq{
				Star: cast.ToInt(seq.Start),
				End:  cast.ToInt(seq.Stop),
			}, false)
			write(s.ctx, w, emailList, options)
		}

	case imap.UIDSet:
		uidSet := numSet.(imap.UIDSet)
		for _, uid := range uidSet {
			emailList := list.GetEmailListByGroup(s.ctx, s.currentMailbox, list.ImapListReq{
				Star: cast.ToInt(uint32(uid.Start)),
				End:  cast.ToInt(uint32(uid.Stop)),
			}, true)
			write(s.ctx, w, emailList, options)
		}
	}
	return nil
}

func write(ctx *context.Context, w *imapserver.FetchWriter, emailList []*response.EmailResponseData, options *imap.FetchOptions) {
	for _, email := range emailList {
		writer := w.CreateMessage(cast.ToUint32(email.SerialNumber))
		if options.UID {
			writer.WriteUID(imap.UID(email.UeId))
		}
		if options.RFC822Size {
			emailContent := parsemail.NewEmailFromModel(email.Email).BuildBytes(ctx, false)
			writer.WriteRFC822Size(cast.ToInt64(len(emailContent)))
		}
		if options.Flags {
			if email.IsRead == 1 {
				writer.WriteFlags([]imap.Flag{imap.FlagSeen})
			} else {
				writer.WriteFlags([]imap.Flag{})
			}
		}
		if options.InternalDate {
			writer.WriteInternalDate(email.CreateTime)
		}
		for _, section := range options.BodySection {
			if !section.Peek {
				detail.MakeRead(ctx, email.Id, true)
			}
			emailContent := parsemail.NewEmailFromModel(email.Email).BuildBytes(ctx, false)

			if section.Specifier == imap.PartSpecifierNone || section.Specifier == imap.PartSpecifierText {
				bodyWriter := writer.WriteBodySection(section, cast.ToInt64(len(emailContent)))
				bodyWriter.Write(emailContent)
				bodyWriter.Close()
			}
			if section.Specifier == imap.PartSpecifierHeader {
				var b bytes.Buffer
				parseEmail := parsemail.NewEmailFromModel(email.Email)
				for _, field := range section.HeaderFields {
					switch field {
					case "date":
						fmt.Fprintf(&b, "Date: %s\r\n", email.CreateTime.Format(time.RFC1123Z))
					case "subject":
						fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", email.Subject))
					case "from":
						if email.FromName != "" {
							fmt.Fprintf(&b, "From: %s <%s>\r\n", mime.QEncoding.Encode("utf-8", email.FromName), email.FromAddress)
						} else {
							fmt.Fprintf(&b, "From: %s\r\n", email.FromAddress)
						}
					case "to":
						fmt.Fprintf(&b, "To: %s\r\n", parseEmail.BuildTo2String())
					case "cc":
						if len(parseEmail.Cc) > 0 {
							fmt.Fprintf(&b, "Cc: %s\r\n", parseEmail.BuildCc2String())
						}
					case "message-id":
						fmt.Fprintf(&b, "Message-ID: %s\r\n", fmt.Sprintf("%d@%s", email.Id, config.Instance.Domain))
					case "content-type":
						args := strings.SplitN(string(emailContent), "\r\n", 3)
						fmt.Fprintf(&b, "%s%s\r\n", args[0], args[1])
					}
				}

				bodyWriter := writer.WriteBodySection(section, cast.ToInt64(b.Len()))
				bodyWriter.Write(b.Bytes())
				bodyWriter.Close()
			}

		}
		writer.Close()
	}
}
