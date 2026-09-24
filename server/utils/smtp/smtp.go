// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package smtp implements the Simple Mail Transfer Protocol as defined in RFC 5321.
// It also implements the following extensions:
//
//	8BITMIME  RFC 1652
//	AUTH      RFC 2554
//	STARTTLS  RFC 3207
//
// Additional extensions may be handled by clients.
//
// The smtp package is frozen and is not accepting new features.
// Some external packages provide more functionality. See:
//
//	https://godoc.org/?q=smtp
//
// 在go原始SMTP协议的基础上修复了TLS验证错误、支持了SMTPS协议、 支持自定义HELLO命令的域名信息
package smtp

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"github.com/Jinnrry/pmail/config"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"
)

var NoSupportSTARTTLSError = errors.New("smtp: server doesn't support STARTTLS")
var EOFError = errors.New("EOF")

const (
	outboundSMTPDialTimeout      = 2 * time.Second
	outboundSMTPQuitWriteTimeout = 250 * time.Millisecond
)

// 出站 SMTP 各阶段超时，遵循 RFC 5321 §4.5.3.2 的要求：
// 客户端必须按"每条命令 / 每个数据块"分别计时，而不能对整个邮件事务
// 设置统一 Deadline——否则携带大附件（例如 1GB 附件）的邮件传输
// 必然超过固定时限被误杀。按阶段计时后，整体超时自然随邮件大小线性扩展。
//
// 使用 var 而非 const，便于在不重新编译的情况下调整（RFC 5321 §4.5.3.2 SHOULD）。
var (
	// RFC 5321 §4.5.3.2.1：等待初始 220 问候。
	// 许多服务器在高负载时会延迟发送 220，因此给足时间。
	outboundSMTPGreetingTimeout = 5 * time.Minute
	// RFC 5321 §4.5.3.2.2 / §4.5.3.2.3：单条命令（EHLO/HELO/MAIL/RCPT 等）
	// 等待响应的超时。
	outboundSMTPCommandTimeout = 5 * time.Minute
	// RFC 5321 §4.5.3.2.4：发出 DATA 命令后等待 354 响应的超时。
	outboundSMTPDataStartTimeout = 2 * time.Minute
	// RFC 5321 §4.5.3.2.5：每个数据块写入的超时，随每次 Write 刷新，
	// 因此大附件传输不会被误杀。
	outboundSMTPDataBlockTimeout = 3 * time.Minute
	// RFC 5321 §4.5.3.2.6：正文发送完毕后等待最终 250 的超时。
	// 此时接收方通常正在做入库/投递处理，过早超时会引发重复投递。
	outboundSMTPDataEndTimeout = 10 * time.Minute
)

// A Client represents a client connection to an SMTP server.
type Client struct {
	// Text is the textproto.Conn used by the Client. It is exported to allow for
	// clients to add extensions.
	Text *textproto.Conn
	// keep a reference to the connection so it can be used to create a TLS
	// connection later
	conn net.Conn
	// whether the Client is using TLS
	tls        bool
	serverName string
	// map of supported extensions
	ext map[string]string
	// supported auth mechanisms
	auth       []string
	localName  string // the name to use in HELO/EHLO
	didHello   bool   // whether we've said HELO/EHLO
	helloError error  // the error from the hello
}

// Dial returns a new Client connected to an SMTP server at addr.
// The addr must include a port, as in "mail.example.com:smtp".
func Dial(addr, fromDomain string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, outboundSMTPDialTimeout)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return NewClient(conn, host, fromDomain)
}

// with tls
func DialTls(addr, domain, fromDomain string) (*Client, error) {
	if domain == "" {
		domain = fromDomain
	}

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         domain,
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: outboundSMTPDialTimeout}, "tcp", addr, tlsconfig)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return NewClient(conn, host, fromDomain)
}

// NewClient returns a new Client using an existing connection and host as a
// server name to be used when authenticating.
func NewClient(conn net.Conn, host, fromDomain string) (*Client, error) {
	text := textproto.NewConn(conn)

	// RFC 5321 §4.5.3.2.1：仅对初始 220 问候单独计时，
	// 读取完成后立即清除，不影响后续命令与数据传输阶段。
	if err := conn.SetReadDeadline(time.Now().Add(outboundSMTPGreetingTimeout)); err != nil {
		text.Close()
		return nil, err
	}
	_, _, err := text.ReadResponse(220)
	if err != nil {
		text.Close()
		return nil, err
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		text.Close()
		return nil, err
	}

	localName := "domain.com"

	if fromDomain != "" {
		localName = fromDomain
	} else if config.Instance != nil && config.Instance.Domain != "" {
		localName = config.Instance.Domain
	}

	c := &Client{Text: text, conn: conn, serverName: host, localName: localName}
	_, c.tls = conn.(*tls.Conn)
	return c, nil
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.Text.Close()
}

// hello runs a hello exchange if needed.
func (c *Client) hello() error {
	if !c.didHello {
		c.didHello = true
		err := c.ehlo()
		if err != nil {
			c.helloError = c.helo()
		}
	}
	return c.helloError
}

// Hello sends a HELO or EHLO to the server as the given host name.
// Calling this method is only necessary if the client needs control
// over the host name used. The client will introduce itself as "localhost"
// automatically otherwise. If Hello is called, it must be called before
// any of the other methods.
func (c *Client) Hello(localName string) error {
	if err := validateLine(localName); err != nil {
		return err
	}
	if c.didHello {
		return errors.New("smtp: Hello called after other methods")
	}
	c.localName = localName
	return c.hello()
}

// cmd is a convenience function that sends a command and returns the response
func (c *Client) cmd(expectCode int, format string, args ...any) (int, string, error) {
	return c.cmdWithTimeout(expectCode, outboundSMTPCommandTimeout, format, args...)
}

// cmdWithTimeout 对单条命令的收发应用超时（RFC 5321 §4.5.3.2 要求的按命令计时）。
// 命令结束后立即清除 Deadline，避免影响后续数据传输阶段。
func (c *Client) cmdWithTimeout(expectCode int, timeout time.Duration, format string, args ...any) (int, string, error) {
	if err := c.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return 0, "", err
	}
	defer func() { _ = c.conn.SetDeadline(time.Time{}) }()

	id, err := c.Text.Cmd(format, args...)
	if err != nil {
		return 0, "", err
	}
	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)
	code, msg, err := c.Text.ReadResponse(expectCode)
	return code, msg, err
}

// helo sends the HELO greeting to the server. It should be used only when the
// server does not support ehlo.
func (c *Client) helo() error {
	c.ext = nil
	_, _, err := c.cmd(250, "HELO %s", c.localName)
	return err
}

// ehlo sends the EHLO (extended hello) greeting to the server. It
// should be the preferred greeting for servers that support it.
func (c *Client) ehlo() error {
	_, msg, err := c.cmd(250, "EHLO %s", c.localName)
	if err != nil {
		return err
	}
	ext := make(map[string]string)
	extList := strings.Split(msg, "\n")
	if len(extList) > 1 {
		extList = extList[1:]
		for _, line := range extList {
			k, v, _ := strings.Cut(line, " ")
			ext[k] = v
		}
	}
	if mechs, ok := ext["AUTH"]; ok {
		c.auth = strings.Split(mechs, " ")
	}
	c.ext = ext
	return err
}

// StartTLS sends the STARTTLS command and encrypts all further communication.
// Only servers that advertise the STARTTLS extension support this function.
func (c *Client) StartTLS(config *tls.Config) error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(220, "STARTTLS")
	if err != nil {
		return err
	}
	if config == nil {
		config = &tls.Config{}
	}
	if config.ServerName == "" {
		// Make a copy to avoid polluting argument
		config = config.Clone()
		config.ServerName = c.serverName
	}
	c.conn = tls.Client(c.conn, config)
	c.Text = textproto.NewConn(c.conn)
	c.tls = true
	return c.ehlo()
}

// TLSConnectionState returns the client's TLS connection state.
// The return values are their zero values if StartTLS did
// not succeed.
func (c *Client) TLSConnectionState() (state tls.ConnectionState, ok bool) {
	tc, ok := c.conn.(*tls.Conn)
	if !ok {
		return
	}
	return tc.ConnectionState(), true
}

// Verify checks the validity of an email address on the server.
// If Verify returns nil, the address is valid. A non-nil return
// does not necessarily indicate an invalid address. Many servers
// will not verify addresses for security reasons.
func (c *Client) Verify(addr string) error {
	if err := validateLine(addr); err != nil {
		return err
	}
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(250, "VRFY %s", addr)
	return err
}

// Auth authenticates a client using the provided authentication mechanism.
// A failed authentication closes the connection.
// Only servers that advertise the AUTH extension support this function.
func (c *Client) Auth(a smtp.Auth) error {
	if err := c.hello(); err != nil {
		return err
	}
	encoding := base64.StdEncoding
	mech, resp, err := a.Start(&smtp.ServerInfo{Name: c.serverName, TLS: c.tls, Auth: c.auth})
	if err != nil {
		c.Quit()
		return err
	}
	resp64 := make([]byte, encoding.EncodedLen(len(resp)))
	encoding.Encode(resp64, resp)
	code, msg64, err := c.cmd(0, "AUTH %s %s", mech, resp64)
	for err == nil {
		var msg []byte
		switch code {
		case 334:
			msg, err = encoding.DecodeString(msg64)
		case 235:
			// the last message isn't base64 because it isn't a challenge
			msg = []byte(msg64)
		default:
			err = &textproto.Error{Code: code, Msg: msg64}
		}
		if err == nil {
			resp, err = a.Next(msg, code == 334)
		}
		if err != nil {
			// abort the AUTH
			c.cmd(501, "*")
			c.Quit()
			break
		}
		if resp == nil {
			break
		}
		resp64 = make([]byte, encoding.EncodedLen(len(resp)))
		encoding.Encode(resp64, resp)
		code, msg64, err = c.cmd(0, "%s", string(resp64))
	}
	return err
}

// Mail issues a MAIL command to the server using the provided email address.
// If the server supports the 8BITMIME extension, Mail adds the BODY=8BITMIME
// parameter. If the server supports the SMTPUTF8 extension, Mail adds the
// SMTPUTF8 parameter.
// This initiates a mail transaction and is followed by one or more Rcpt calls.
func (c *Client) Mail(from string) error {
	if err := validateLine(from); err != nil {
		return err
	}
	if err := c.hello(); err != nil {
		return err
	}
	cmdStr := "MAIL FROM:<%s>"
	if c.ext != nil {
		if _, ok := c.ext["8BITMIME"]; ok {
			cmdStr += " BODY=8BITMIME"
		}
		if _, ok := c.ext["SMTPUTF8"]; ok {
			cmdStr += " SMTPUTF8"
		}
	}
	_, _, err := c.cmd(250, cmdStr, from)
	return err
}

// Rcpt issues a RCPT command to the server using the provided email address.
// A call to Rcpt must be preceded by a call to Mail and may be followed by
// a Data call or another Rcpt call.
func (c *Client) Rcpt(to string) error {
	if err := validateLine(to); err != nil {
		return err
	}
	_, _, err := c.cmd(25, "RCPT TO:<%s>", to)
	return err
}

type dataCloser struct {
	c *Client
	io.WriteCloser
}

func (d *dataCloser) Close() error {
	if err := d.WriteCloser.Close(); err != nil {
		return err
	}
	// RFC 5321 §4.5.3.2.6：等待最终 250 响应的超时单独计时。
	// 此时接收方通常正在做入库处理，过早超时会引发重复投递。
	if err := d.c.conn.SetReadDeadline(time.Now().Add(outboundSMTPDataEndTimeout)); err != nil {
		return err
	}
	defer func() { _ = d.c.conn.SetReadDeadline(time.Time{}) }()
	_, _, err := d.c.Text.ReadResponse(250)
	return err
}

// Data issues a DATA command to the server and returns a writer that
// can be used to write the mail headers and body. The caller should
// close the writer before calling any more methods on c. A call to
// Data must be preceded by one or more calls to Rcpt.
func (c *Client) Data() (io.WriteCloser, error) {
	// RFC 5321 §4.5.3.2.4：等待 354 响应的超时单独计时。
	_, _, err := c.cmdWithTimeout(354, outboundSMTPDataStartTimeout, "DATA")
	if err != nil {
		return nil, err
	}
	return &dataCloser{c, &dataBlockWriter{c: c, w: c.Text.DotWriter()}}, nil
}

// dataBlockWriter 在每个数据块写入前刷新写超时（RFC 5321 §4.5.3.2.5：
// 每个数据块单独计时）。因此整体传输时限随邮件大小线性扩展，
// 大附件不会被会话级 Deadline 误杀；而对端挂死时，
// 最后一个数据块也会在超时内失败并释放连接。
type dataBlockWriter struct {
	c *Client
	w io.WriteCloser
}

// outboundSMTPDataBlockSize 是单个数据块的大小。
// 上层可能把整封邮件（含超大附件）一次性传入 Write，
// 这里按固定块切分并逐块刷新超时，确保"每个数据块单独计时"
// 的语义不依赖调用方的写入粒度。
const outboundSMTPDataBlockSize = 32 * 1024

func (d *dataBlockWriter) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		chunk := p
		if len(chunk) > outboundSMTPDataBlockSize {
			chunk = chunk[:outboundSMTPDataBlockSize]
		}
		if err := d.c.conn.SetWriteDeadline(time.Now().Add(outboundSMTPDataBlockTimeout)); err != nil {
			return written, err
		}
		n, err := d.w.Write(chunk)
		written += n
		if err != nil {
			return written, err
		}
		p = p[n:]
	}
	return written, nil
}

func (d *dataBlockWriter) Close() error {
	if err := d.c.conn.SetWriteDeadline(time.Now().Add(outboundSMTPDataBlockTimeout)); err != nil {
		return err
	}
	return d.w.Close()
}

func finishDelivery(c *Client, w io.WriteCloser, msg []byte) error {
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	// The final 250 response to DATA transfers responsibility for the message.
	// Send QUIT with a bounded write, but do not wait for 221: a failed or
	// missing reply after acceptance must not block or trigger redelivery.
	if err := c.conn.SetWriteDeadline(time.Now().Add(outboundSMTPQuitWriteTimeout)); err != nil {
		log.Debugf("Could not bound SMTP QUIT write after message acceptance: %v", err)
	}
	if err := c.Text.PrintfLine("QUIT"); err != nil {
		log.Debugf("SMTP QUIT failed after message acceptance: %v", err)
	}
	return nil
}

func SendMailWithTls(domain string, addr string, a smtp.Auth, from string, fromDomain string, to []string, msg []byte) error {
	if err := validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}
	c, err := DialTls(addr, domain, fromDomain)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.hello(); err != nil {
		return err
	}
	if a != nil && c.ext != nil {
		if _, ok := c.ext["AUTH"]; !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	return finishDelivery(c, w, msg)
}

// SendMail connects to the server at addr, switches to TLS if
// possible, authenticates with the optional mechanism a if possible,
// and then sends an email from address from, to addresses to, with
// message msg.
// The addr must include a port, as in "mail.example.com:smtp".
//
// The addresses in the to parameter are the SMTP RCPT addresses.
//
// The msg parameter should be an RFC 822-style email with headers
// first, a blank line, and then the message body. The lines of msg
// should be CRLF terminated. The msg headers should usually include
// fields such as "From", "To", "Subject", and "Cc".  Sending "Bcc"
// messages is accomplished by including an email address in the to
// parameter but not including it in the msg headers.
//
// The SendMail function and the net/smtp package are low-level
// mechanisms and provide no support for DKIM signing, MIME
// attachments (see the mime/multipart package), or other mail
// functionality. Higher-level packages exist outside of the standard
// library.
// 修复TSL验证问题
func SendMail(domain string, addr string, a smtp.Auth, from string, fromDomain string, to []string, msg []byte) error {

	log.Debugf("SendMail,%s ,%s ,%s ,%s ,%v ", domain, addr, from, fromDomain, to)

	if err := validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}
	c, err := Dial(addr, fromDomain)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.hello(); err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); !ok {
		return NoSupportSTARTTLSError
	}

	var config *tls.Config
	if domain != "" {
		config = &tls.Config{
			ServerName: domain,
		}
	}

	if err = c.StartTLS(config); err != nil {
		return err
	}
	if a != nil && c.ext != nil {
		if _, ok := c.ext["AUTH"]; !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	return finishDelivery(c, w, msg)
}

// SendMailUnsafe 无TLS加密的邮件发送方式
func SendMailUnsafe(domain string, addr string, a smtp.Auth, from string, fromDomain string, to []string, msg []byte) error {
	if err := validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}
	c, err := Dial(addr, fromDomain)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.hello(); err != nil {
		return err
	}

	if a != nil && c.ext != nil {
		if _, ok := c.ext["AUTH"]; !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	return finishDelivery(c, w, msg)
}

// Extension reports whether an extension is support by the server.
// The extension name is case-insensitive. If the extension is supported,
// Extension also returns a string that contains any parameters the
// server specifies for the extension.
func (c *Client) Extension(ext string) (bool, string) {
	if err := c.hello(); err != nil {
		return false, ""
	}
	if c.ext == nil {
		return false, ""
	}
	ext = strings.ToUpper(ext)
	param, ok := c.ext[ext]
	return ok, param
}

// Reset sends the RSET command to the server, aborting the current mail
// transaction.
func (c *Client) Reset() error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(250, "RSET")
	return err
}

// Noop sends the NOOP command to the server. It does nothing but check
// that the connection to the server is okay.
func (c *Client) Noop() error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(250, "NOOP")
	return err
}

// Quit sends the QUIT command and closes the connection to the server.
func (c *Client) Quit() error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(221, "QUIT")
	if err != nil {
		return err
	}
	return c.Text.Close()
}

// validateLine checks to see if a line has CR or LF as per RFC 5321.
func validateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return errors.New("smtp: A line must not contain CR or LF")
	}
	return nil
}
