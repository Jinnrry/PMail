package pop3_server

//import (
//	"bytes"
//	"fmt"
//	"github.com/Jinnrry/gopop"
//	"github.com/Jinnrry/pmail/config"
//	"github.com/Jinnrry/pmail/db"
//	"github.com/Jinnrry/pmail/utils/context"
//	"github.com/emersion/go-message/mail"
//	"io"
//	"testing"
//)
//
//func Test_action_Retr(t *testing.T) {
//	config.Init()
//	db.Init("")
//
//	a := action{}
//	session := &gopop.Session{
//		Ctx: &context.Context{
//			UserID: 1,
//		},
//	}
//	got, got1, err := a.Retr(session, 301)
//
//	_, _, _ = got, got1, err
//}
//
//func Test_email(t *testing.T) {
//	var b bytes.Buffer
//
//	// Create our mail header
//	var h mail.Header
//
//	// Create a new mail writer
//	mw, _ := mail.CreateWriter(&b, h)
//
//	// Create a text part
//	tw, _ := mw.CreateInline()
//
//	var html mail.InlineHeader
//
//	html.Header.Set("Content-Transfer-Encoding", "base64")
//	w, _ := tw.CreatePart(html)
//
//	io.WriteString(w, "=")
//
//	w.Close()
//
//	tw.Close()
//
//	fmt.Printf("%s", b.String())
//
//}
