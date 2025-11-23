package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/signal"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/spf13/cast"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

var httpClient *http.Client

const TestPort = 17888

var TestHost string = "http://127.0.0.1:" + cast.ToString(TestPort)

func TestMain(m *testing.M) {
	cookeieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	httpClient = &http.Client{Jar: cookeieJar, Timeout: 5 * time.Minute}
	os.Remove("config/config.json")
	os.Remove("config/pmail_temp.db")
	os.Setenv("setup_port", cast.ToString(TestPort))

	go func() {
		main()
	}()
	time.Sleep(5 * time.Second)

	m.Run()

	signal.StopChan <- true
	time.Sleep(3 * time.Second)
}

func TestMaster(t *testing.T) {
	t.Run("TestPort", testPort)
	t.Run("testDataBaseSet", testDataBaseSet)
	t.Run("testPwdSet", testPwdSet)
	t.Run("testDomainSet", testDomainSet)
	t.Run("testDNSSet", testDNSSet)
	cfg, err := config.ReadConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.HttpsEnabled = 2
	cfg.HttpPort = TestPort
	cfg.LogLevel = "debug"
	err = config.WriteConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("testSSLSet", testSSLSet)
	t.Logf("Stop 8 Second for wating restart")
	time.Sleep(8 * time.Second)
	t.Run("testLogin", testLogin)           // 登录管理员账号
	t.Run("testCreateUser", testCreateUser) // 创建3个测试用户
	t.Run("testEditUser", testEditUser)     // 编辑user2，封禁user3
	t.Run("testSendEmail", testSendEmail)
	t.Logf("Stop 8 Second for wating sending")
	time.Sleep(8 * time.Second)
	t.Run("testEmailList", testEmailList)
	t.Run("testGetDetail", testGetEmailDetail)
	t.Run("testDelEmail", testDelEmail)

	t.Run("testSendEmail2User1", testSendEmail2User1)
	t.Run("testSendEmail2User12", testSendEmail2User12)
	t.Run("testSendEmail2User2", testSendEmail2User2)
	t.Run("testSendEmail2User3", testSendEmail2User3)
	time.Sleep(8 * time.Second)

	t.Run("testLoginUser3", testLoginUser3) // 测试登录被封禁账号

	t.Run("testLoginUser2", testLoginUser2) // 测试登录普通账号

	t.Run("testUser2EmailList", testUser2EmailList)

	t.Run("testUser2DelEmail", testUser2DelEmail) // 删除2个人共同拥有的邮件

	// 创建group
	t.Run("testCreateGroup", testCreateGroup)

	// 创建rule
	t.Run("testCreateRule", testCreateRule)

	// 再次发邮件
	t.Run("testMoverEmailSend", testSendEmail2User2ForMove)
	time.Sleep(4 * time.Second)

	t.Run("testMoverEmailSend", testSendEmail2User2ForSpam)
	time.Sleep(3 * time.Second)

	// 生成10封测试邮件
	t.Run("genTestEmailData", genTestEmailData)
	time.Sleep(3 * time.Second)

	// 检查规则执行
	t.Run("testCheckRule", testCheckRule)
	time.Sleep(3 * time.Second)
}

func testCheckRule(t *testing.T) {
	var ue models.UserEmail
	db.Instance.Where("group_id!=0").Get(&ue)
	if ue.GroupId == 0 {
		t.Error("邮件规则执行失败！")
	}
}

func testGetEmailDetail(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/detail", "application/json", strings.NewReader(`{
	"id":1
}`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("GetEmailDetail Error! ", data)
	}

}

func testCreateRule(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/rule/add", "application/json", strings.NewReader(`{
	"name":"Move Group",
	"rules":[{"field":"Subject","type":"contains","rule":"Move"}],
	"action":4,
	"params":"1",
	"sort":1
}`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("CreateRule Api Error!", data)
	}

	ret, err = httpClient.Post(TestHost+"/api/rule/get", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("CreateRule Api Error!", data)
	}
	dt := data.Data.([]any)
	if len(dt) != 1 {
		t.Error("Rule List Is Empty!")
	}

}

func testCreateGroup(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/group/add", "application/json", strings.NewReader(`{
	"name":"TestGroup"
}`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("CreateGroup Api Error!", data)
	}

	ret, err = httpClient.Post(TestHost+"/api/group/list", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("CreateGroup Api Error!", data)
	}
	dt := data.Data.([]any)
	if len(dt) != 4 {
		t.Errorf("Group List Check Error!,response: %+v", data)
	}
}

func testEditUser(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/user/edit", "application/json", strings.NewReader(`{
	"account":"user2",
	"username":"user2New",
	"password":"user2New"
}`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Edit User Api Error!", data)
	}

	ret, err = httpClient.Post(TestHost+"/api/user/edit", "application/json", strings.NewReader(`{
	"account":"user3",
	"disabled": 1
}`))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Edit User Api Error!", data)
	}

}

func testCreateUser(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/user/create", "application/json", strings.NewReader(`{
	"account":"user1",
	"username":"user1",
	"password":"user1"
}`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error(data)
		t.Error("Create User Api Error!")
	}

	ret, err = httpClient.Post(TestHost+"/api/user/create", "application/json", strings.NewReader(`{
	"account":"user2",
	"username":"user2",
	"password":"user2"
}`))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Create User Api Error!")
	}

	ret, err = httpClient.Post(TestHost+"/api/user/create", "application/json", strings.NewReader(`{
	"account":"user3",
	"username":"user3",
	"password":"user3"
}`))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Create User Api Error!")
	}

}

func testPort(t *testing.T) {
	if !portCheck(TestPort) {
		t.Error("port check failed")
	} else {
		t.Log("port check passed")
	}

}

func testDataBaseSet(t *testing.T) {

	// 获取配置
	ret, err := http.Post(TestHost+"/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"database\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("Response %+v", data)
		t.Error("Get Database Config Api Error!")
		return
	}

	argList := flag.Args()

	configData := `
{"action":"set","step":"database","db_type":"sqlite","db_dsn":"./config/pmail_temp.db"}
`

	if array.InArray("mysql", argList) {
		configData = `
{"action":"set","step":"database","db_type":"mysql","db_dsn":"root:githubTest@tcp(mysql:3306)/pmail?parseTime=True"}
`
	} else if array.InArray("postgres", argList) {
		configData = `
{"action":"set","step":"database","db_type":"postgres","db_dsn":"postgres://postgres:githubTest@postgres:5432/pmail?sslmode=disable"}
`
	}

	// 设置配置
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(configData))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get Database Config Api Error!")
	}

	// 获取配置
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"database\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get Database Config Api Error!")
	}
	dt := data.Data.(map[string]interface{})
	if cast.ToString(dt["db_dsn"]) == "" {
		t.Error("Check Database Config Api Error!")
	}

	t.Log("Database Config Api Success!")
}

func testPwdSet(t *testing.T) {

	// 获取配置
	ret, err := http.Post(TestHost+"/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"password\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get Password Config Api Error!")
	}
	// 设置配置
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(`
{"action":"set","step":"password","account":"testCase","password":"testCase"}
`))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Set Password Config Api Error!")
		t.Error(data)
	}

	// 获取配置
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"password\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get Password Config Api Error!")
	}

	if cast.ToString(data.Data) != "testCase" {
		t.Error("Check Password Config Api Error!")
	}

	t.Log("Password Config Api Success!")
}

func testDomainSet(t *testing.T) {
	// 获取配置
	ret, err := http.Post(TestHost+"/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"domain\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get domain Config Api Error!")
	}
	// 设置配置
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(`
{"action":"set","step":"domain","smtp_domain":"test.domain","web_domain":"mail.test.domain"}
`))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Set domain Config Api Error!")
	}

	// 获取配置
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"domain\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get Password Config Api Error!")
	}

	dt := data.Data.(map[string]interface{})

	if cast.ToString(dt["smtp_domain"]) != "test.domain" {
		t.Error("Check domain Config Api Error!")
	}
	if cast.ToString(dt["web_domain"]) != "mail.test.domain" {
		t.Error("Check domain Config Api Error!")
	}
	t.Log("domain Config Api Success!")
}

func testDNSSet(t *testing.T) {
	// 获取配置
	ret, err := http.Post(TestHost+"/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"dns\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get domain Config Api Error!")
	}
	t.Log("DNS Set Success!")
}

func testSSLSet(t *testing.T) {
	// 获取配置
	ret, err := http.Post(TestHost+"/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"ssl\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get domain Config Api Error!")
	}
	// 设置配置
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(`
{"action":"set","step":"ssl","ssl_type":"1","key_path":"./config/ssl/private.key","crt_path":"./config/ssl/public.crt"}
`))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Set domain Config Api Error!")
	}

	t.Log("domain Config Api Success!")
}

func testLogin(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/login", "application/json", strings.NewReader("{\"account\":\"testCase\",\"password\":\"testCase\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Login Api Error!")
	}
	t.Logf("testLogin Success! Response: %+v", data)
}

func testLoginUser2(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/login", "application/json", strings.NewReader("{\"account\":\"user2\",\"password\":\"user2New\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Login User2 Api Error!", data)
	}
	t.Logf("testLoginUser2 Success! Response: %+v", data)
}

func testLoginUser3(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/login", "application/json", strings.NewReader("{\"account\":\"user3\",\"password\":\"user3\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 100 {
		t.Error("Login User3 Api Error!", data)
	}
	t.Logf("testLoginUser3 Success! Response: %+v", data)
}

func testSendEmail(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(`
		{
    "from": {
        "name": "i",
        "email": "i@test.domain"
    },
    "to": [
        {
            "name": "y",
            "email": "y@test.domain"
        }
    ],
    "cc": [
        
    ],
    "subject": "Title",
    "text": "text",
    "html": "<div>text</div>"
}

`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Send Email Api Error!")
	}
	t.Logf("testSendEmail Success! Response: %+v", data)
}

func testSendEmail2User2ForSpam(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(`
		{
    "from": {
        "name": "user2",
        "email": "user2@test.domain"
    },
    "to": [
        {
            "name": "y",
            "email": "admin@test.domain"
        }
    ],
    "cc": [
        
    ],
    "subject": "spam",
    "text": "NeedMove",
    "html": "<div>text</div>"
}

`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Send Email Api Error!")
	}

	t.Logf("testSendEmail2User2ForMove Success! Response: %+v", data)

}

func testSendEmail2User2ForMove(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(`
		{
    "from": {
        "name": "user2",
        "email": "user2@test.domain"
    },
    "to": [
        {
            "name": "y",
            "email": "user2@test.domain"
        }
    ],
    "cc": [
        
    ],
    "subject": "MovePlease",
    "text": "NeedMove",
    "html": "<div>text</div>"
}

`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Send Email Api Error!")
	}

	t.Logf("testSendEmail2User2ForMove Success! Response: %+v", data)

}

func genTestEmailData(t *testing.T) {
	for i := 0; i < 10; i++ {
		ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(fmt.Sprintf(
			`
		{
    "from": {
        "name": "user2",
        "email": "user2@test.domain"
    },
    "to": [
        {
            "name": "admin",
            "email": "admin@test.domain"
        }
    ],
    "cc": [
        
    ],
    "subject": "测试邮件%d",
    "text": "测试邮件%d",
    "html": "<div>测试邮件%d</div>"
}

`, i, i, i)))
		if err != nil {
			t.Error(err)
		}
		data, err := readResponse(ret.Body)
		if err != nil {
			t.Error(err)
		}
		if data.ErrorNo != 0 {
			t.Error("Send Email Api Error!")
		}
		time.Sleep(3 * time.Second)
	}

}

func testSendEmail2User12(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(`
		{
    "from": {
        "name": "i",
        "email": "i@test.domain"
    },
    "to": [
        {
            "name": "y",
            "email": "user1@test.domain"
        },
		{
            "name": "y2",
            "email": "user2@test.domain"
        }
    ],
    "cc": [
        
    ],
    "subject": "HelloUser1User2",
    "text": "HelloUser1User2",
    "html": "<div>HelloUser1User2</div>"
}

`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Send Email Api Error!")
	}

	t.Logf("testSendEmail2User1 Success! Response: %+v", data)
}

func testSendEmail2User1(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(`
		{
    "from": {
        "name": "i",
        "email": "i@test.domain"
    },
    "to": [
        {
            "name": "y",
            "email": "user1@test.domain"
        }
    ],
    "cc": [
        
    ],
    "subject": "HelloUser1",
    "text": "text",
    "html": "<div>text</div>"
}

`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Send Email Api Error!")
	}

	t.Logf("testSendEmail2User1 Success! Response: %+v", data)
}

func testSendEmail2User2(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(`
		{
    "from": {
        "name": "i",
        "email": "i@test.domain"
    },
    "to": [
        {
            "name": "y",
            "email": "user2@test.domain"
        }
    ],
    "cc": [
        
    ],
    "subject": "HelloUser2",
    "text": "text",
    "html": "<div>text</div>"
}

`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Send Email Api Error!")
	}

	t.Logf("testSendEmail2User2 Success! Response: %+v", data)
}

func testSendEmail2User3(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(`
		{
    "from": {
        "name": "i",
        "email": "i@test.domain"
    },
    "to": [
        {
            "name": "y",
            "email": "user3@test.domain"
        }
    ],
    "cc": [
        
    ],
    "subject": "HelloUser3",
    "text": "text",
    "html": "<div>text</div>"
}

`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Send Email Api Error!")
	}

	t.Logf("testSendEmail2User3 Success! Response: %+v", data)

}

func testEmailList(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/list", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get Email List Api Error!")
	}
	dt := data.Data.(map[string]interface{})
	if dt["list"] == nil || len(dt["list"].([]interface{})) == 0 {
		t.Error("Email List Is Empty!")
		return
	}

	lst := dt["list"].([]interface{})
	item := lst[0].(map[string]interface{})
	id := cast.ToInt(item["id"])
	if id == 0 {
		t.Error("Email List Data Error!")
	}

	t.Logf("testEmailList Success! Response: %+v", data)
}

func testUser2EmailList(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/list", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get Email List Api Error!")
	}
	dt := data.Data.(map[string]interface{})

	if dt["list"] == nil || len(dt["list"].([]interface{})) != 2 {
		t.Error("Email List Is Empty!")
	}

	t.Logf("testUser2EmailList Success! Response: %+v", data)

}

func testDelEmail(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/list", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get Email List Api Error!")
	}
	dt := data.Data.(map[string]interface{})
	if len(dt["list"].([]interface{})) == 0 {
		t.Error("Email List Is Empty!")
	}
	lst := dt["list"].([]interface{})
	item := lst[0].(map[string]interface{})
	id := cast.ToInt(item["id"])

	ret, err = httpClient.Post(TestHost+"/api/email/del", "application/json", strings.NewReader(fmt.Sprintf(`{
	"ids":[%d]	
}`, id)))
	if err != nil {
		t.Error(err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Email Delete Api Error!")
	}
	var mail models.UserEmail
	db.Instance.Where("email_id = ?", id).Get(&mail)
	if mail.Status != 3 {
		t.Error("Email Delete Api Error!")
	}

	t.Logf("testDelEmail Success! Response: %+v", data)
}

func testUser2DelEmail(t *testing.T) {
	ret, err := httpClient.Post(TestHost+"/api/email/list", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get Email List Api Error!")
	}
	dt := data.Data.(map[string]interface{})
	if len(dt["list"].([]interface{})) == 0 {
		t.Error("Email List Is Empty!")
	}
	lst := dt["list"].([]interface{})

	for _, item := range lst {
		// 删除两个用户的邮件
		title := cast.ToString(item.(map[string]interface{})["title"])
		id := cast.ToInt(item.(map[string]interface{})["id"])
		if title == "HelloUser1User2" {
			ret, err = httpClient.Post(TestHost+"/api/email/del", "application/json", strings.NewReader(fmt.Sprintf(`{
	"ids":[%d]	
}`, id)))
			if err != nil {
				t.Error(err)
			}
			data, err = readResponse(ret.Body)
			if err != nil {
				t.Error(err)
			}
			if data.ErrorNo != 0 {
				t.Error("Email Delete Api Error!")
			}
			var mails []models.UserEmail
			db.Instance.Where("email_id = ?", id).Find(&mails)
			for _, mail := range mails {
				if mail.Status != 3 && mail.UserID == 3 {
					t.Error("Email Delete Api Error!")
				}
				if mail.UserID != 3 && mail.Status == 3 {
					t.Error("Email Delete Api Error!")
				}
			}

		}

	}

	t.Logf("testDelEmail Success! Response: %+v", data)
}

// portCheck 检查端口是占用
func portCheck(port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(port)))
	if err != nil {
		return true
	}
	defer l.Close()
	return false
}

func readResponse(r io.Reader) (*response.Response, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	ret := &response.Response{}
	err = json.Unmarshal(data, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
