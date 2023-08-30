package smtp_server

import (
	"bytes"
	"encoding/json"
	"github.com/mileusna/spf"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/netip"
	"pmail/db"
	"pmail/dto/parsemail"
	"pmail/hooks"
	"pmail/utils/async"
	"strings"
	"time"
)

func (s *Session) Data(r io.Reader) error {
	emailData, err := io.ReadAll(r)
	if err != nil {
		log.Error("邮件内容无法读取", err)
		return err
	}

	as1 := async.New(nil)
	for _, hook := range hooks.HookList {
		if hook == nil {
			continue
		}
		as1.WaitProcess(func(hk any) {
			hk.(hooks.EmailHook).ReceiveParseBefore(emailData)
		}, hook)
	}
	as1.Wait()

	log.Infof("邮件原始内容: %s", emailData)

	var dkimStatus, SPFStatus bool

	// DKIM校验
	dkimStatus = parsemail.Check(bytes.NewReader(emailData))

	email := parsemail.NewEmailFromReader(bytes.NewReader(emailData))

	if err != nil {
		log.Fatalf("邮件内容解析失败！ Error : %v \n", err)
	}

	SPFStatus = spfCheck(s.RemoteAddress.String(), email.Sender, email.Sender.EmailAddress)

	var dkimV, spfV int8
	if dkimStatus {
		dkimV = 1
	}
	if SPFStatus {
		spfV = 1
	}

	as2 := async.New(nil)
	for _, hook := range hooks.HookList {
		if hook == nil {
			continue
		}
		as2.WaitProcess(func(hk any) {
			hk.(hooks.EmailHook).ReceiveParseAfter(email)
		}, hook)
	}
	as2.Wait()

	sql := "INSERT INTO email (send_date, subject, reply_to, from_name, from_address, `to`, bcc, cc, text, html, sender, attachments,spf_check, dkim_check, create_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = db.Instance.Exec(sql,
		email.Date,
		email.Subject,
		json2string(email.ReplyTo),
		email.From.Name,
		email.From.EmailAddress,
		json2string(email.To),
		json2string(email.Bcc),
		json2string(email.Cc),
		email.Text,
		email.HTML,
		json2string(email.Sender),
		json2string(email.Attachments),
		spfV,
		dkimV,
		time.Now())

	if err != nil {
		log.Println("mysql insert error:", err.Error())
	}

	return nil
}

func json2string(d any) string {
	by, _ := json.Marshal(d)
	return string(by)
}

func spfCheck(remoteAddress string, sender *parsemail.User, senderString string) bool {
	//spf校验
	ipAddress, _ := netip.ParseAddrPort(remoteAddress)

	ip := net.ParseIP(ipAddress.Addr().String())
	if ip.IsPrivate() {
		return true
	}

	tmp := strings.Split(sender.EmailAddress, "@")
	if len(tmp) < 2 {
		return false
	}

	res := spf.CheckHost(ip, tmp[1], senderString, "")

	if res == spf.None || res == spf.Pass {
		// spf校验通过
		return true
	}
	return false
}
