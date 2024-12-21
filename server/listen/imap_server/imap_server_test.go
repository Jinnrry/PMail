package imap_server

import (
	"crypto/tls"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/charset"
	"mime"
	"testing"
	"time"
)

var clientUnLogin *imapclient.Client
var clientLogin *imapclient.Client

func TestMain(m *testing.M) {
	config.Init()
	db.Init("")
	go StarTLS()
	time.Sleep(2 * time.Second)

	options := &imapclient.Options{
		WordDecoder: &mime.WordDecoder{CharsetReader: charset.Reader},
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	var err error
	clientUnLogin, err = imapclient.DialTLS("127.0.0.1:993", options)
	if err != nil {
		panic(err)
	}

	clientLogin, err = imapclient.DialTLS("127.0.0.1:993", options)
	if err != nil {
		panic(err)
	}

	err = clientLogin.Login("testCase", "testCase").Wait()
	if err != nil {
		panic(err)
	}

	m.Run()
}

func TestCapability(t *testing.T) {

	res, err := clientUnLogin.Capability().Wait()
	if err != nil {
		t.Error(err)
	}
	if _, ok := res["IMAP4rev1"]; !ok {
		t.Error("Capability Error")
	}

}

func TestLogin(t *testing.T) {
	err := clientUnLogin.Login("testCase", "testCaseasdfsadf").Wait()
	sErr := err.(*imap.Error)
	if sErr.Code != "AUTHENTICATIONFAILED" {
		t.Error("Login Error")
	}
}

func TestCreate(t *testing.T) {

}
func TestDelete(t *testing.T) {

}
func TestRename(t *testing.T) {

}
func TestList(t *testing.T) {
	res, err := clientUnLogin.List("", "", &imap.ListOptions{}).Collect()

	if err == nil {
		t.Logf("%+v", res)
		t.Error("List Unlogin error")
	}

	res, err = clientLogin.List("", "", &imap.ListOptions{}).Collect()
	if err != nil {
		t.Error(err)
	}
	if len(res) == 0 {
		t.Error("List Error")
	}

	res, err = clientLogin.List("", "*", &imap.ListOptions{}).Collect()
	if err != nil {
		t.Error(err)
	}
	if len(res) == 0 {
		t.Error("List Error")
	}

}
func TestAppend(t *testing.T) {

}
func TestSelect(t *testing.T) {
	res, err := clientUnLogin.Select("INBOX", &imap.SelectOptions{}).Wait()
	if err == nil {
		t.Logf("%+v", res)
		t.Error("Select Unlogin error")
	}

	res, err = clientLogin.Select("INBOX", &imap.SelectOptions{}).Wait()
	if err != nil {
		t.Logf("%+v", res)
		t.Error("Select error")
	}

	if res == nil || res.NumMessages == 0 {
		t.Error("Select Error")
	}

	res, err = clientLogin.Select("Deleted Messages", &imap.SelectOptions{}).Wait()
	if err != nil {
		t.Logf("%+v", res)
		t.Error("Select error")
	}

	if res == nil || res.NumMessages == 0 {
		t.Error("Select Error")
	}

}

func TestStatus(t *testing.T) {
	res, err := clientUnLogin.Status("INBOX", &imap.StatusOptions{}).Wait()
	if err == nil {
		t.Logf("%+v", res)
		t.Error("Select Unlogin error")
	}

	res, err = clientLogin.Status("INBOX", &imap.StatusOptions{}).Wait()
	if err != nil {
		t.Logf("%+v", res)
		t.Error("Select error")
	}

}

func TestFetch(t *testing.T) {
	res2, err := clientLogin.Select("INBOX", &imap.SelectOptions{}).Wait()
	if err != nil {
		t.Logf("%+v", res2)
		t.Error("Fetch error")
	}

	res, err := clientLogin.Fetch(imap.SeqSetNum(1), &imap.FetchOptions{
		Envelope:     true,
		Flags:        true,
		InternalDate: true,
		RFC822Size:   true,
		UID:          true,
		BodySection: []*imap.FetchItemBodySection{
			{
				Specifier: imap.PartSpecifierText,
				Peek:      true,
			},
		},
	}).Collect()
	if err != nil {
		t.Logf("%+v", res)
		t.Error("Fetch error")
	}

	res, err = clientLogin.Fetch(imap.SeqSetNum(1, 2, 3, 4, 5, 6, 7, 8, 9), &imap.FetchOptions{
		Flags: true,
		UID:   true,
	}).Collect()
	if err != nil {
		t.Logf("%+v", res)
		t.Error("Fetch error")
	}

	res, err = clientLogin.Fetch(imap.SeqSetNum(1), &imap.FetchOptions{
		Envelope:     true,
		Flags:        true,
		InternalDate: true,
		RFC822Size:   true,
		UID:          true,
		BodySection: []*imap.FetchItemBodySection{
			{
				Specifier:    imap.PartSpecifierHeader,
				HeaderFields: []string{"subject"},
				Peek:         true,
			},
		},
	}).Collect()
	if err != nil {
		t.Logf("%+v", res)
		t.Error("Fetch error")
	}

	res, err = clientLogin.Fetch(imap.SeqSetNum(1), &imap.FetchOptions{
		UID: true,
		BodySection: []*imap.FetchItemBodySection{
			{
				Specifier: imap.PartSpecifierHeader,
				Peek:      true,
			},
		},
	}).Collect()
	if err != nil {
		t.Logf("%+v", res)
		t.Error("Fetch error")
	}
}
func TestStore(t *testing.T) {

}
func TestClose(t *testing.T) {

}
func TestExpunge(t *testing.T) {

}
func TestExamine(t *testing.T) {

}
func TestSubscribe(t *testing.T) {

}
func TestUnSubscribe(t *testing.T) {

}
func TestLSub(t *testing.T) {

}

func TestCheck(t *testing.T) {

}
func TestSearch(t *testing.T) {

}
func TestCopy(t *testing.T) {

}

func TestNoop(t *testing.T) {
	err := clientLogin.Noop().Wait()
	if err != nil {
		t.Error(err)
	}
}
func TestIDLE(t *testing.T) {

}
func TestUnselect(t *testing.T) {

}

func TestLogout(t *testing.T) {
	err := clientLogin.Logout().Wait()
	if err != nil {
		t.Error(err)
	}
}
