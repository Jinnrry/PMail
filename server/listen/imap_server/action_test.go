package imap_server

import (
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/goimap"
	"github.com/Jinnrry/pmail/utils/id"
	"testing"
)

func Test_action_List(t *testing.T) {
	config.Init()
	db.Init("")
	db.Instance.ShowSQL(true)

	_, err := db.Instance.Exec("DELETE FROM `group`")

	groupData1 := models.Group{
		ID:     1,
		Name:   "第一层",
		UserId: 1,
	}
	groupData2 := models.Group{
		ID:       2,
		Name:     "第二层",
		UserId:   1,
		ParentId: 1,
	}
	db.Instance.Insert(&groupData1)
	db.Instance.Insert(&groupData2)

	session := goimap.Session{
		Account: "admin",
	}
	tc := &context.Context{}
	tc.UserID = 1
	tc.SetValue(context.LogID, id.GenLogID())
	session.Ctx = tc

	ret, err := action{}.List(&session, "", "")
	if err != nil {
		t.Error(err)
	}
	if len(ret) == 1 && ret[0] == `* LIST (\NoSelect \HasChildren) "/" "[PMail]` {
		t.Logf("%s", ret[0])
	} else {
		t.Errorf("%s", ret)
	}

	ret, err = action{}.List(&session, "", "*")
	if err != nil {
		t.Error(err)
	}
	if len(ret) == 8 && ret[7] == `* LIST (\HasNoChildren) "/" "&eyxOAFxC-/&eyxOjFxC-"` {
		t.Logf("%s", ret[0])
	} else {
		t.Errorf("%s", ret)
	}
}
