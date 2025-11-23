package imap_server

import (
	"crypto/tls"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/array"
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
	err := clientLogin.Create("一级菜单", nil).Wait()
	if err != nil {
		t.Error(err)
	}

	err = clientLogin.Create("一级菜单/二级菜单", nil).Wait()
	if err != nil {
		t.Error(err)
	}

	res, err := clientLogin.List("", "*", nil).Collect()
	if err != nil {
		t.Error(err)
	}
	var mailbox []string
	for _, v := range res {
		mailbox = append(mailbox, v.Mailbox)
	}

	if !array.InArray("一级菜单", mailbox) || !array.InArray("一级菜单/二级菜单", mailbox) {
		t.Error(mailbox)
	}

}

func TestRename(t *testing.T) {

	err := clientLogin.Rename("一级菜单", "主菜单").Wait()
	if err != nil {
		t.Error(err)
	}

	res, err := clientLogin.List("", "*", nil).Collect()
	if err != nil {
		t.Error(err)
	}
	var mailbox []string
	for _, v := range res {
		mailbox = append(mailbox, v.Mailbox)
	}

	if !array.InArray("主菜单", mailbox) {
		t.Error(mailbox)
	}
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

	res, err = clientLogin.List("", "一级菜单/%", &imap.ListOptions{}).Collect()
	if err != nil {
		t.Error(err)
	}
	if len(res) == 0 {
		t.Error("List Error")
	}

	if len(res) != 1 {
		t.Error("List Error")
	}

	res, err = clientLogin.List("", "一级菜单/*", &imap.ListOptions{}).Collect()
	if err != nil {
		t.Error(err)
	}
	if len(res) != 1 {
		t.Error("List Error")
	}
	if len(res) == 0 {
		t.Error("List Error")
	}

}

func TestDelete(t *testing.T) {

	clientLogin.Create("一级菜单/二级菜单", nil).Wait()

	err := clientLogin.Delete("二级菜单").Wait()
	if err != nil {
		t.Error(err)
	}
	res, err := clientLogin.List("", "*", nil).Collect()
	if err != nil {
		t.Error(err)
	}
	var mailbox []string
	for _, v := range res {
		mailbox = append(mailbox, v.Mailbox)
	}

	if array.InArray("二级菜单", mailbox) {
		t.Error(mailbox)
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
	res, err := clientLogin.Store(
		imap.UIDSetNum(1),
		&imap.StoreFlags{
			Op:    imap.StoreFlagsAdd,
			Flags: []imap.Flag{"\\Seen"},
		},
		&imap.StoreOptions{}).Collect()
	if err != nil {
		t.Errorf("%+v", err)
	}
	t.Logf("%+v", res)

}
func TestClose(t *testing.T) {

}
func TestExpunge(t *testing.T) {

	clientLogin.Select("INBOX", &imap.SelectOptions{}).Wait()

	res, err := clientLogin.UIDExpunge(imap.UIDSetNum(1, 2)).Collect()

	if err != nil {
		t.Errorf("%+v", err)
	}
	t.Logf("%+v", res)
	var ues []models.UserEmail
	db.Instance.Table("user_email").Where("id=1 or id=2").Find(&ues)
	if len(ues) > 0 {
		t.Errorf("TestExpunge Error")
	}

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
	clientLogin.Select("INBOX", &imap.SelectOptions{}).Wait()

	res, err := clientLogin.Search(&imap.SearchCriteria{
		UID: []imap.UIDSet{
			[]imap.UIDRange{
				{Start: 1},
			},
			[]imap.UIDRange{
				{Start: 2},
			},
			[]imap.UIDRange{
				{Start: 2, Stop: 5},
			},
		},
	}, &imap.SearchOptions{}).Wait()
	if err != nil {
		t.Errorf("%+v", err)
	}
	t.Logf("%+v", res)
}
func TestMove(t *testing.T) {
	clientLogin.Select("INBOX", &imap.SelectOptions{}).Wait()

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

	if len(res) > 0 {
		uid := res[0].UID

		_, err = clientLogin.Move(imap.UIDSetNum(uid), "Junk").Wait()
		if err != nil {
			t.Errorf("%+v", err)
		}

	}

}

func TestCopy(t *testing.T) {
	clientLogin.Select("INBOX", &imap.SelectOptions{}).Wait()

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

	if len(res) > 0 {
		_, err = clientLogin.Copy(imap.UIDSetNum(res[0].UID), "Junk").Wait()
		if err != nil {
			t.Errorf("%+v", err)
		}
	} else {
		t.Error("No Fetch Result")
	}

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
