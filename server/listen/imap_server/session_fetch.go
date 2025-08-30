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

// 纯文本 text/plain
func bsTextPlain(size uint32, numLines int64) *imap.BodyStructureSinglePart {
	return &imap.BodyStructureSinglePart{
		Type:     "text",
		Subtype:  "plain",
		Params:   map[string]string{"charset": "utf-8"},
		Encoding: "base64",
		Size:     size, // 按字节数
		Text:     &imap.BodyStructureText{NumLines: numLines},
		Extended: &imap.BodyStructureSinglePartExt{},
	}
}

// HTML text/html
func bsTextHTML(size uint32, numLines int64) *imap.BodyStructureSinglePart {
	return &imap.BodyStructureSinglePart{
		Type:     "text",
		Subtype:  "html",
		Params:   map[string]string{"charset": "utf-8"},
		Encoding: "base64",
		Size:     size,
		Text:     &imap.BodyStructureText{NumLines: numLines},
		Extended: &imap.BodyStructureSinglePartExt{},
	}
}

// 通用附件（传入 MIME，如 "application/pdf"）
func bsAttachment(filename, mime string, size uint32, encoding string) *imap.BodyStructureSinglePart {
	mt, st := "application", "octet-stream"
	if slash := strings.IndexByte(mime, '/'); slash > 0 {
		mt, st = mime[:slash], mime[slash+1:]
	}
	return &imap.BodyStructureSinglePart{
		Type:     mt,
		Subtype:  st,
		Params:   map[string]string{"name": filename}, // 备用名
		ID:       "",                                  // 可填 Content-ID
		Encoding: encoding,                            // 常见 "base64"
		Size:     size,
		Extended: &imap.BodyStructureSinglePartExt{
			Disposition: &imap.BodyStructureDisposition{
				Value:  "attachment",
				Params: map[string]string{"filename": filename}, // 客户端优先用这里
			},
		},
	}
}

// multipart/alternative：text + html
func bsAlternative(text, html *imap.BodyStructureSinglePart) *imap.BodyStructureMultiPart {
	if text == nil && html == nil {
		return &imap.BodyStructureMultiPart{
			Subtype:  "alternative",
			Children: []imap.BodyStructure{},
			Extended: &imap.BodyStructureMultiPartExt{},
		}
	}
	if text == nil {
		return &imap.BodyStructureMultiPart{
			Subtype:  "alternative",
			Children: []imap.BodyStructure{html},
			Extended: &imap.BodyStructureMultiPartExt{}, // 可选：Params/Disposition/Language/Location
		}
	}
	if html == nil {
		return &imap.BodyStructureMultiPart{
			Subtype:  "alternative",
			Children: []imap.BodyStructure{text},
			Extended: &imap.BodyStructureMultiPartExt{}, // 可选：Params/Disposition/Language/Location
		}
	}

	return &imap.BodyStructureMultiPart{
		Subtype:  "alternative",
		Children: []imap.BodyStructure{text, html},
		Extended: &imap.BodyStructureMultiPartExt{}, // 可选：Params/Disposition/Language/Location
	}
}

// multipart/mixed：{ alternative(text+html), attachments... }
func bsMixedWithAttachments(alt *imap.BodyStructureMultiPart, extend bool, atts ...imap.BodyStructure) *imap.BodyStructureMultiPart {
	children := []imap.BodyStructure{alt}
	children = append(children, atts...)
	var ext *imap.BodyStructureMultiPartExt
	if extend {
		ext = &imap.BodyStructureMultiPartExt{}
	}
	return &imap.BodyStructureMultiPart{
		Subtype:  "mixed",
		Children: children,
		Extended: ext,
	}
}

func write(ctx *context.Context, w *imapserver.FetchWriter, emailList []*response.EmailResponseData, options *imap.FetchOptions) {
	for _, email := range emailList {
		writer := w.CreateMessage(cast.ToUint32(email.SerialNumber))

		traEmail := parsemail.NewEmailFromModel(email.Email)

		if options.UID {
			writer.WriteUID(imap.UID(email.UeId))
		}
		if options.BodyStructure != nil {
			var html, text *imap.BodyStructureSinglePart
			if len(traEmail.HTML) > 0 {
				html = bsTextHTML(uint32(len(traEmail.HTML)), int64(bytes.Count(traEmail.HTML, []byte("\n"))+1))
			}

			if len(traEmail.Text) > 0 {
				text = bsTextPlain(uint32(len(traEmail.Text)), int64(bytes.Count(traEmail.Text, []byte("\n"))+1))
			}

			alt := bsAlternative(text, html)

			var attrs []imap.BodyStructure
			for _, attachment := range traEmail.Attachments {
				attrs = append(attrs, bsAttachment(attachment.Filename, attachment.ContentType, uint32(len(attachment.Content)), "base64"))
			}
			bs := bsMixedWithAttachments(alt, options.BodyStructure.Extended, attrs...) // 最终的 BodyStructure（接口值）

			writer.WriteBodyStructure(bs)
		}
		if options.RFC822Size {
			emailContent := traEmail.BuildBytes(ctx, false)
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
			emailContent := traEmail.BuildBytes(ctx, false)

			if section.Specifier == imap.PartSpecifierNone || section.Specifier == imap.PartSpecifierText {
				if len(section.Part) == 2 {
					// 取text部分
					bodyWriter := writer.WriteBodySection(section, cast.ToInt64(len(emailContent)))
					bodyWriter.Write(traEmail.BuildPart(ctx, section.Part))
					bodyWriter.Close()
				} else {
					bodyWriter := writer.WriteBodySection(section, cast.ToInt64(len(emailContent)))
					bodyWriter.Write(emailContent)
					bodyWriter.Close()
				}
			}
			if section.Specifier == imap.PartSpecifierHeader {
				var b bytes.Buffer
				fields := section.HeaderFields

				if fields == nil || len(fields) == 0 {
					fields = []string{
						"date", "subject", "from", "to", "cc", "message-id", "content-type",
					}
				}

				for _, field := range fields {
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
						fmt.Fprintf(&b, "To: %s\r\n", traEmail.BuildTo2String())
					case "cc":
						if len(traEmail.Cc) > 0 {
							fmt.Fprintf(&b, "Cc: %s\r\n", traEmail.BuildCc2String())
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
