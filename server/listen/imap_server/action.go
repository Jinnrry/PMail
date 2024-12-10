package imap_server

import (
	"github.com/Jinnrry/pmail/listen/imap_server/goimap"
	log "github.com/sirupsen/logrus"
)

type action struct{}

func (a action) Create(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "Create", path)
	return nil
}

func (a action) Delete(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "Create", path)
	return nil
}

func (a action) Rename(session *goimap.Session, oldPath, newPath string) error {
	log.Infof("%s,%s,%s", "Create", oldPath, newPath)
	return nil
}

func (a action) List(session *goimap.Session, basePath, template string) ([]string, error) {
	log.Infof("%s,%s,%s", "Create", basePath, template)
	return nil, nil
}

func (a action) Append(session *goimap.Session, item string) error {
	log.Infof("%s,%s", "Create", item)
	return nil
}

func (a action) Select(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "Create", path)
	return nil
}

func (a action) Fetch(session *goimap.Session, mailIds, dataNames string) (string, error) {
	log.Infof("%s,%s,%s", "Fetch", mailIds, dataNames)
	return "", nil
}

func (a action) Store(session *goimap.Session, mailId, flags string) error {
	log.Infof("%s,%s,%s", "Store", mailId, flags)
	return nil
}

func (a action) Close(session *goimap.Session) error {
	log.Infof("%s", "Close")
	return nil
}

func (a action) Expunge(session *goimap.Session) error {
	log.Infof("%s", "Expunge")
	return nil
}

func (a action) Examine(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "Examine", path)
	return nil
}

func (a action) Subscribe(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "Subscribe", path)
	return nil
}

func (a action) UnSubscribe(session *goimap.Session, path string) error {
	log.Infof("%s,%s", "UnSubscribe", path)
	return nil
}

func (a action) LSub(session *goimap.Session, path, mailbox string) ([]string, error) {
	log.Infof("%s,%s,%s", "LSub", path, mailbox)
	return nil, nil
}

func (a action) Status(session *goimap.Session, mailbox, category string) (string, error) {
	log.Infof("%s,%s,%s", "Status", mailbox, category)
	return "", nil
}

func (a action) Check(session *goimap.Session) error {
	log.Infof("%s", "Check")
	return nil
}

func (a action) Search(session *goimap.Session, keyword, criteria string) (string, error) {
	log.Infof("%s,%s,%s", "Search", keyword, criteria)
	return "", nil
}

func (a action) Copy(session *goimap.Session, mailId, mailBoxName string) error {
	log.Infof("%s,%s,%s", "Copy", mailId, mailBoxName)
	return nil
}

func (a action) CapaBility(session *goimap.Session) ([]string, error) {
	log.Infof("%s", "CapaBility")
	return nil, nil
}

func (a action) Noop(session *goimap.Session) error {
	log.Infof("%s", "Noop")
	return nil
}

func (a action) Login(session *goimap.Session, username, password string) error {
	log.Infof("%s,%s,%s", "Login", username, password)
	return nil
}

func (a action) Logout(session *goimap.Session) error {
	log.Infof("%s", "Logout")
	return nil
}

func (a action) Custom(session *goimap.Session, cmd string, args []string) ([]string, error) {
	log.Infof("%s,%+v", cmd, args)
	return nil, nil
}
