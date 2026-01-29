package imap_server

import (
	"bytes"
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/detail"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/spf13/cast"
)

// userToAddress 将 parsemail.User 转换为 imap.Address
func userToAddress(u *parsemail.User) imap.Address {
	if u == nil {
		return imap.Address{}
	}
	// 解析邮箱地址的 local 和 domain 部分
	mailbox := u.EmailAddress
	host := ""
	if at := strings.LastIndex(u.EmailAddress, "@"); at >= 0 {
		mailbox = u.EmailAddress[:at]
		host = u.EmailAddress[at+1:]
	}
	return imap.Address{
		Name:    u.Name,
		Mailbox: mailbox,
		Host:    host,
	}
}

// usersToAddresses 将用户列表转换为地址列表
func usersToAddresses(users []*parsemail.User) []imap.Address {
	if len(users) == 0 {
		return nil
	}
	addrs := make([]imap.Address, 0, len(users))
	for _, u := range users {
		if u != nil && u.EmailAddress != "" {
			addrs = append(addrs, userToAddress(u))
		}
	}
	if len(addrs) == 0 {
		return nil
	}
	return addrs
}

// buildEnvelope 构建 IMAP ENVELOPE 结构
func buildEnvelope(email *response.EmailResponseData, traEmail *parsemail.Email) *imap.Envelope {
	// From 地址
	var from []imap.Address
	if email.FromAddress != "" {
		mailbox := email.FromAddress
		host := ""
		if at := strings.LastIndex(email.FromAddress, "@"); at >= 0 {
			mailbox = email.FromAddress[:at]
			host = email.FromAddress[at+1:]
		}
		from = []imap.Address{{
			Name:    email.FromName,
			Mailbox: mailbox,
			Host:    host,
		}}
	}

	// Sender (如果没有单独的 sender，使用 from)
	var sender []imap.Address
	if traEmail.Sender != nil && traEmail.Sender.EmailAddress != "" {
		sender = []imap.Address{userToAddress(traEmail.Sender)}
	} else {
		sender = from
	}

	// Reply-To
	var replyTo []imap.Address
	if len(traEmail.ReplyTo) > 0 {
		replyTo = usersToAddresses(traEmail.ReplyTo)
	} else {
		replyTo = from
	}

	// Message-ID
	messageID := fmt.Sprintf("<%d@%s>", email.Id, config.Instance.Domain)

	return &imap.Envelope{
		Date:      email.CreateTime,
		Subject:   email.Subject,
		From:      from,
		Sender:    sender,
		ReplyTo:   replyTo,
		To:        usersToAddresses(traEmail.To),
		Cc:        usersToAddresses(traEmail.Cc),
		Bcc:       usersToAddresses(traEmail.Bcc),
		MessageID: messageID,
		// InReplyTo 和 References 暂不支持
	}
}

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
// 注意：BodyStructureMultiPart 必须至少有一个子元素，否则 go-imap 会 panic
func bsAlternative(text, html *imap.BodyStructureSinglePart) *imap.BodyStructureMultiPart {
	var children []imap.BodyStructure

	if text != nil {
		children = append(children, text)
	}
	if html != nil {
		children = append(children, html)
	}

	// 如果没有任何内容，创建一个空的 text/plain 占位符
	// 这可以避免 go-imap 的 panic: "imap.BodyStructureMultiPart must have at least one child"
	if len(children) == 0 {
		children = append(children, bsTextPlain(0, 0))
	}

	return &imap.BodyStructureMultiPart{
		Subtype:  "alternative",
		Children: children,
		Extended: &imap.BodyStructureMultiPartExt{},
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
		if options.Envelope {
			env := buildEnvelope(email, traEmail)
			writer.WriteEnvelope(env)
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

			// 优先检查是否请求 HEADER.FIELDS
			if len(section.HeaderFields) > 0 || section.Specifier == imap.PartSpecifierHeader {
				var b bytes.Buffer
				fields := section.HeaderFields

				if fields == nil || len(fields) == 0 {
					// 没有指定字段，返回所有常见头部
					fields = []string{
						"date", "subject", "from", "to", "cc", "message-id", "content-type",
					}
				}

				for _, field := range fields {
					fieldLower := strings.ToLower(field)
					switch fieldLower {
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
					case "sender":
						if email.FromName != "" {
							fmt.Fprintf(&b, "Sender: %s <%s>\r\n", mime.QEncoding.Encode("utf-8", email.FromName), email.FromAddress)
						} else {
							fmt.Fprintf(&b, "Sender: %s\r\n", email.FromAddress)
						}
					case "reply-to":
						if len(traEmail.ReplyTo) > 0 && traEmail.ReplyTo[0].EmailAddress != "" {
							if traEmail.ReplyTo[0].Name != "" {
								fmt.Fprintf(&b, "Reply-To: %s <%s>\r\n", mime.QEncoding.Encode("utf-8", traEmail.ReplyTo[0].Name), traEmail.ReplyTo[0].EmailAddress)
							} else {
								fmt.Fprintf(&b, "Reply-To: %s\r\n", traEmail.ReplyTo[0].EmailAddress)
							}
						}
					case "to":
						toStr := traEmail.BuildTo2String()
						if toStr != "" {
							fmt.Fprintf(&b, "To: %s\r\n", toStr)
						}
					case "cc":
						if len(traEmail.Cc) > 0 {
							fmt.Fprintf(&b, "Cc: %s\r\n", traEmail.BuildCc2String())
						}
					case "bcc":
						if len(traEmail.Bcc) > 0 {
							fmt.Fprintf(&b, "Bcc: %s\r\n", traEmail.BuildBcc2String())
						}
					case "message-id":
						fmt.Fprintf(&b, "Message-ID: <%d@%s>\r\n", email.Id, config.Instance.Domain)
					case "content-type":
						args := strings.SplitN(string(emailContent), "\r\n", 3)
						if len(args) >= 2 {
							fmt.Fprintf(&b, "%s%s\r\n", args[0], args[1])
						}
					case "references", "in-reply-to", "thread-topic", "thread-index", "x-priority", "x-mailer", "x-android-message-id":
						// 这些头部我们目前不存储，跳过
					default:
						// 其他未知头部，忽略
					}
				}

				// 添加结束空行
				b.WriteString("\r\n")

				bodyWriter := writer.WriteBodySection(section, cast.ToInt64(b.Len()))
				bodyWriter.Write(b.Bytes())
				bodyWriter.Close()
			} else if section.Specifier == imap.PartSpecifierNone || section.Specifier == imap.PartSpecifierText {
				if len(section.Part) >= 1 {
					// 获取指定 part 的内容
					partContent := traEmail.BuildPart(ctx, section.Part)
					if partContent != nil {
						bodyWriter := writer.WriteBodySection(section, cast.ToInt64(len(partContent)))
						bodyWriter.Write(partContent)
						bodyWriter.Close()
					} else {
						// Part 不存在，返回空
						bodyWriter := writer.WriteBodySection(section, 0)
						bodyWriter.Close()
					}
				} else {
					bodyWriter := writer.WriteBodySection(section, cast.ToInt64(len(emailContent)))
					bodyWriter.Write(emailContent)
					bodyWriter.Close()
				}
			}
		}
		writer.Close()
	}
}
