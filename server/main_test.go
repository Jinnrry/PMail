package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"pmail/db"
	"pmail/dto/response"
	"pmail/models"
	"pmail/services/setup"
	"pmail/signal"
	"strconv"
	"strings"
	"testing"
	"time"
)

var httpClient *http.Client

func TestMain(m *testing.M) {

	fmt.Println("!!!!!TestMain!!!!!!!!")

	cookeieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	httpClient = &http.Client{Jar: cookeieJar, Timeout: 5 * time.Second}
	os.Remove("config/config.json")
	os.Remove("config/pmail_temp.db")
	go main()
	time.Sleep(3 * time.Second)

	m.Run()
	fmt.Println("!!!!!TestMain!!!!!!!!")
	time.Sleep(5 * time.Second)
	stop()
	signal.RestartChan <- false
	fmt.Println("!!!!!TestMain!!!!!!!!")

}

func TestMaster(t *testing.T) {
	fmt.Println("!!!!!TestMaster!!!!!!!!")

	t.Run("TestPort", testPort)
	t.Run("testDataBaseSet", testDataBaseSet)
	t.Run("testPwdSet", testPwdSet)
	t.Run("testDomainSet", testDomainSet)
	t.Run("testDNSSet", testDNSSet)
	cfg, err := setup.ReadConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.HttpsEnabled = 2
	err = setup.WriteConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("testSSLSet", testSSLSet)
	time.Sleep(3 * time.Second)
	t.Run("testLogin", testLogin)
	t.Run("testSendEmail", testSendEmail)
	time.Sleep(3 * time.Second)
	t.Run("testEmailList", testEmailList)
	t.Run("testDelEmail", testDelEmail)
}

func testPort(t *testing.T) {
	if !portCheck(80) {
		t.Error("port check failed")
	}
	t.Log("port check passed")
}

func testDataBaseSet(t *testing.T) {

	// 获取配置
	ret, err := http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"database\"}"))
	if err != nil {
		t.Error(err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Error(err)
	}
	if data.ErrorNo != 0 {
		t.Error("Get Database Config Api Error!")
	}
	// 设置配置
	ret, err = http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader(`
{"action":"set","step":"database","db_type":"sqlite","db_dsn":"./config/pmail_temp.db"}
`))
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
	ret, err = http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"database\"}"))
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
	if cast.ToString(dt["db_dsn"]) != "./config/pmail_temp.db" {
		t.Error("Check Database Config Api Error!")
	}

	t.Log("Database Config Api Success!")
}

func testPwdSet(t *testing.T) {

	// 获取配置
	ret, err := http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"password\"}"))
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
	ret, err = http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader(`
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
	}

	// 获取配置
	ret, err = http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"password\"}"))
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
	ret, err := http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"domain\"}"))
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
	ret, err = http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader(`
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
	ret, err = http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"domain\"}"))
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
	ret, err := http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"dns\"}"))
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
}

func testSSLSet(t *testing.T) {
	// 获取配置
	ret, err := http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader("{\"action\":\"get\",\"step\":\"ssl\"}"))
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
	ret, err = http.Post("http://127.0.0.1/api/setup", "application/json", strings.NewReader(`
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
	ret, err := httpClient.Post("http://127.0.0.1/api/login", "application/json", strings.NewReader("{\"account\":\"testCase\",\"password\":\"testCase\"}"))
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
}

func testSendEmail(t *testing.T) {
	ret, err := httpClient.Post("http://127.0.0.1/api/email/send", "application/json", strings.NewReader(`
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
}

func testEmailList(t *testing.T) {
	ret, err := httpClient.Post("http://127.0.0.1/api/email/list", "application/json", strings.NewReader(`{}`))
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
}

func testDelEmail(t *testing.T) {
	ret, err := httpClient.Post("http://127.0.0.1/api/email/list", "application/json", strings.NewReader(`{}`))
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

	ret, err = httpClient.Post("http://127.0.0.1/api/email/del", "application/json", strings.NewReader(fmt.Sprintf(`{
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
	var mail models.Email
	db.Instance.Where("id = ?", id).Get(&mail)
	if mail.Status != 3 {
		t.Error("Email Delete Api Error!")
	}

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
