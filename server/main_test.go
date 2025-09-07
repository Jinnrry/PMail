package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/signal"
	"github.com/Jinnrry/pmail/utils/array"
	"github.com/spf13/cast"
)

var httpClient *http.Client

const TestPort = 17888

var TestHost string = "http://127.0.0.1:" + cast.ToString(TestPort)

func TestMain(m *testing.M) {
	log.Println("【测试主函数 TestMain】--- 开始执行 ---")

	cookeieJar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("创建 cookie jar 失败: %v", err)
	}

	httpClient = &http.Client{Jar: cookeieJar, Timeout: 5 * time.Minute}

	log.Println("准备测试环境：删除旧的配置文件和数据库...")
	os.Remove("config/config.json")
	os.Remove("config/pmail_temp.db")
	log.Println("旧文件删除完毕。")

	log.Println("在新的 goroutine 中启动 main() 函数...")
	go func() {
		main()
	}()

	// Wait for server to start with more robust health checking
	log.Println("等待服务器启动...")
	log.Printf("测试环境信息: TestHost=%s, TestPort=%d", TestHost, TestPort)

	// 首先检查端口是否被占用 (应该被我们的服务器占用)
	log.Println("检查端口是否被占用...")
	portOccupied := false
	for i := 0; i < 30; i++ {
		if portCheck(TestPort) {
			log.Printf("端口 %d 在 %d 秒后成功被占用!", TestPort, i+1)
			portOccupied = true
			break
		}
		if i%5 == 0 {
			log.Printf("端口 %d 仍然可用... 尝试次数 %d/30", TestPort, i+1)
		}
		time.Sleep(1 * time.Second)
	}
	if !portOccupied {
		log.Fatal("【严重错误】服务器在30秒内未能占用指定端口，启动失败。")
	}

	// 然后检查 HTTP 端点是否可访问
	log.Println("检查 HTTP 端点是否可访问 (/api/ping)...")
	serverReady := false
	client := &http.Client{Timeout: 10 * time.Second}
	for i := 0; i < 90; i++ { // 为容器环境增加超时时间到90秒
		resp, err := client.Get(TestHost + "/api/ping")
		if err == nil {
			if resp.StatusCode == 200 {
				resp.Body.Close()
				log.Printf("服务器在 %d 秒后准备就绪!", i+1)
				serverReady = true
				break
			} else {
				log.Printf("HTTP 请求返回异常状态码: %d", resp.StatusCode)
				resp.Body.Close()
			}
		} else {
			if i%15 == 0 {
				log.Printf("仍在等待服务器响应... 尝试次数 %d/90 (错误: %v)", i+1, err)
			}
		}
		time.Sleep(1 * time.Second)
	}

	if !serverReady {
		log.Fatal("【严重错误】服务器在90秒内未能成功启动并响应。")
	}

	log.Println("服务器已准备就绪，额外等待3秒以确保所有服务初始化完成...")
	time.Sleep(3 * time.Second) // 额外的缓冲时间

	log.Println("--- 开始运行所有测试用例 ---")
	m.Run()
	log.Println("--- 所有测试用例运行结束 ---")

	log.Println("向服务器发送停止信号...")
	signal.StopChan <- true
	time.Sleep(3 * time.Second)
	log.Println("【测试主函数 TestMain】--- 执行完毕 ---")
}

func TestMaster(t *testing.T) {
	t.Log("【Master测试套件】--- 开始 ---")

	t.Run("TestPort", testPort)
	t.Run("testDataBaseSet", testDataBaseSet)
	t.Run("testPwdSet", testPwdSet)
	t.Run("testDomainSet", testDomainSet)
	t.Run("testDNSSet", testDNSSet)

	t.Log("读取并修改配置文件，为后续测试做准备...")
	cfg, err := config.ReadConfig()
	if err != nil {
		t.Fatalf("读取配置文件失败: %v", err)
	}
	cfg.HttpsEnabled = 2
	cfg.HttpPort = TestPort
	cfg.LogLevel = "debug"
	cfg.SMTPPort = 10025
	cfg.IMAPPort = 10143
	cfg.POP3Port = 10110
	cfg.SMTPSPort = 10465
	cfg.IMAPSPort = 10993
	cfg.POP3SPort = 10995
	err = config.WriteConfig(cfg)
	if err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}
	t.Log("配置文件修改并写入成功。")

	t.Run("testSSLSet", testSSLSet)

	t.Logf("暂停8秒，等待服务器根据新配置重启...")
	time.Sleep(8 * time.Second)

	t.Run("testLogin", testLogin)           // 登录管理员账号
	t.Run("testCreateUser", testCreateUser) // 创建3个测试用户
	t.Run("testEditUser", testEditUser)     // 编辑user2，封禁user3
	t.Run("testSendEmail", testSendEmail)   // 发送第一封测试邮件

	t.Logf("暂停8秒，等待邮件发送和处理...")
	time.Sleep(8 * time.Second)

	t.Run("testEmailList", testEmailList)
	t.Run("testGetDetail", testGetEmailDetail)
	t.Run("testDelEmail", testDelEmail)

	t.Run("testSendEmail2User1", testSendEmail2User1)
	t.Run("testSendEmail2User2", testSendEmail2User2)
	t.Run("testSendEmail2User3", testSendEmail2User3)
	t.Log("发送三封邮件给不同用户后，暂停8秒等待处理...")
	time.Sleep(8 * time.Second)

	t.Run("testLoginUser3", testLoginUser3) // 测试登录被封禁账号
	t.Run("testLoginUser2", testLoginUser2) // 测试登录普通账号
	t.Run("testUser2EmailList", testUser2EmailList)

	t.Run("testCreateGroup", testCreateGroup) // 创建group
	t.Run("testCreateRule", testCreateRule)   // 创建rule

	t.Run("testMoverEmailSend", testSendEmail2User2ForMove)
	t.Log("发送用于测试'移动'规则的邮件后，暂停4秒...")
	time.Sleep(4 * time.Second)

	t.Run("testMoverEmailSend", testSendEmail2User2ForSpam)
	t.Log("发送用于测试'垃圾邮件'规则的邮件后，暂停3秒...")
	time.Sleep(3 * time.Second)

	t.Run("genTestEmailData", genTestEmailData)
	t.Log("生成10封测试邮件后，暂停3秒...")
	time.Sleep(3 * time.Second)

	t.Run("testCheckRule", testCheckRule)
	t.Log("检查规则执行情况后，暂停3秒...")
	time.Sleep(3 * time.Second)

	t.Log("【Master测试套件】--- 结束 ---")
}

func testCheckRule(t *testing.T) {
	t.Log("【子测试 testCheckRule】--- 开始检查邮件规则是否被执行 ---")
	var ue models.UserEmail
	db.Instance.Where("group_id != 0").Get(&ue)
	if ue.GroupId == 0 {
		t.Error("【失败】邮件规则执行失败！未找到任何被移动到自定义分组的邮件。")
	} else {
		t.Logf("【成功】邮件规则执行成功！找到邮件ID: %d 被移动到分组ID: %d", ue.EmailId, ue.GroupId)
	}
	t.Log("【子测试 testCheckRule】--- 结束 ---")
}

func testGetEmailDetail(t *testing.T) {
	t.Log("【子测试 testGetEmailDetail】--- 开始 ---")
	payload := `{"id":1}`
	t.Logf("请求URL: %s/api/email/detail, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/email/detail", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("HTTP请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取邮件详情API返回错误! 响应: %+v", data)
	} else {
		t.Logf("【成功】获取邮件详情成功, 响应: %+v", data)
	}
	t.Log("【子测试 testGetEmailDetail】--- 结束 ---")
}

func testCreateRule(t *testing.T) {
	t.Log("【子测试 testCreateRule】--- 开始 ---")
	payload := `{
    "name":"Move Group",
    "rules":[{"field":"Subject","type":"contains","rule":"Move"}],
    "action":4,
    "params":"1",
    "sort":1
}`
	t.Logf("请求URL: %s/api/rule/add, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/rule/add", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("HTTP请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】创建规则API返回错误! 响应: %+v", data)
	} else {
		t.Log("【成功】创建规则成功。")
	}

	t.Log("开始获取规则列表以验证...")
	ret, err = httpClient.Post(TestHost+"/api/rule/get", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Errorf("HTTP请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取规则列表API返回错误! 响应: %+v", data)
	}
	dt, ok := data.Data.([]any)
	if !ok || len(dt) != 1 {
		t.Errorf("【失败】规则列表数量不为1! 实际数量: %d", len(dt))
	} else {
		t.Logf("【成功】获取到规则列表，数量为 %d。", len(dt))
	}
	t.Log("【子测试 testCreateRule】--- 结束 ---")
}

func testCreateGroup(t *testing.T) {
	t.Log("【子测试 testCreateGroup】--- 开始 ---")
	payload := `{"name":"TestGroup"}`
	t.Logf("请求URL: %s/api/group/add, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/group/add", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("HTTP请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】创建分组API返回错误! 响应: %+v", data)
	} else {
		t.Log("【成功】创建分组成功。")
	}

	t.Log("开始获取分组列表以验证...")
	ret, err = httpClient.Post(TestHost+"/api/group/list", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Errorf("HTTP请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取分组列表API返回错误! 响应: %+v", data)
	}
	dt, ok := data.Data.([]any)
	// 默认有3个分组(收件箱，垃圾箱，已发送)，加上新建的1个，总共4个
	if !ok || len(dt) != 4 {
		t.Errorf("【失败】分组列表数量检查错误! 预期4个，实际 %d。响应: %+v", len(dt), data)
	} else {
		t.Logf("【成功】获取到分组列表，数量为 %d。", len(dt))
	}
	t.Log("【子测试 testCreateGroup】--- 结束 ---")
}

func testEditUser(t *testing.T) {
	t.Log("【子测试 testEditUser】--- 开始 ---")
	// 编辑 user2
	payload1 := `{
    "account":"user2",
    "username":"user2New",
    "password":"user2New"
}`
	t.Logf("请求URL: %s/api/user/edit, 请求体: %s", TestHost, payload1)
	ret, err := httpClient.Post(TestHost+"/api/user/edit", "application/json", strings.NewReader(payload1))
	if err != nil {
		t.Errorf("编辑user2请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析编辑user2的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】编辑用户 'user2' API返回错误! 响应: %+v", data)
	} else {
		t.Log("【成功】编辑用户 'user2' 成功。")
	}

	// 封禁 user3
	payload2 := `{
    "account":"user3",
    "disabled": 1
}`
	t.Logf("请求URL: %s/api/user/edit, 请求体: %s", TestHost, payload2)
	ret, err = httpClient.Post(TestHost+"/api/user/edit", "application/json", strings.NewReader(payload2))
	if err != nil {
		t.Errorf("封禁user3请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析封禁user3的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】封禁用户 'user3' API返回错误! 响应: %+v", data)
	} else {
		t.Log("【成功】封禁用户 'user3' 成功。")
	}
	t.Log("【子测试 testEditUser】--- 结束 ---")
}

func testCreateUser(t *testing.T) {
	t.Log("【子测试 testCreateUser】--- 开始 ---")
	users := []string{"user1", "user2", "user3"}
	for _, user := range users {
		payload := fmt.Sprintf(`{"account":"%s", "username":"%s", "password":"%s"}`, user, user, user)
		t.Logf("请求URL: %s/api/user/create, 请求体: %s", TestHost, payload)
		ret, err := httpClient.Post(TestHost+"/api/user/create", "application/json", strings.NewReader(payload))
		if err != nil {
			t.Errorf("创建用户 '%s' 请求失败: %v", user, err)
			continue
		}
		data, err := readResponse(ret.Body)
		if err != nil {
			t.Errorf("读取或解析创建用户 '%s' 的响应失败: %v", user, err)
			continue
		}
		if data.ErrorNo != 0 {
			t.Errorf("【失败】创建用户 '%s' API返回错误! 响应: %+v", user, data)
		} else {
			t.Logf("【成功】创建用户 '%s' 成功。", user)
		}
	}
	t.Log("【子测试 testCreateUser】--- 结束 ---")
}

func testPort(t *testing.T) {
	t.Log("【子测试 testPort】--- 开始检查端口占用情况 ---")
	if !portCheck(TestPort) {
		t.Error("【失败】端口检查失败，端口未被占用。")
	} else {
		t.Log("【成功】端口检查通过，端口已被程序占用。")
	}
	t.Log("【子测试 testPort】--- 结束 ---")
}

func testDataBaseSet(t *testing.T) {
	t.Log("【子测试 testDataBaseSet】--- 开始 ---")

	// 获取配置
	t.Log("步骤1: 获取当前数据库配置...")
	getPayload := `{"action":"get","step":"database"}`
	ret, err := http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(getPayload))
	if err != nil {
		t.Errorf("获取配置请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析获取配置的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取数据库配置API返回错误! 响应: %+v", data)
		return
	}
	t.Logf("获取当前配置成功: %+v", data)

	argList := flag.Args()
	t.Logf("检测到测试启动参数: %v", argList)

	configData := `{"action":"set","step":"database","db_type":"sqlite","db_dsn":"./config/pmail_temp.db"}`
	if array.InArray("mysql", argList) {
		configData = `{"action":"set","step":"database","db_type":"mysql","db_dsn":"root:githubTest@tcp(mysql:3306)/pmail?parseTime=True"}`
		t.Log("检测到 'mysql' 参数，使用MySQL配置。")
	} else if array.InArray("postgres", argList) {
		configData = `{"action":"set","step":"database","db_type":"postgres","db_dsn":"postgres://postgres:githubTest@postgres:5432/pmail?sslmode=disable"}`
		t.Log("检测到 'postgres' 参数，使用PostgreSQL配置。")
	} else {
		t.Log("未检测到特定数据库参数，使用默认的SQLite配置。")
	}

	// 设置配置
	t.Log("步骤2: 设置新的数据库配置...")
	t.Logf("请求体: %s", configData)
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(configData))
	if err != nil {
		t.Errorf("设置配置请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析设置配置的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】设置数据库配置API返回错误! 响应: %+v", data)
	} else {
		t.Log("设置新配置成功。")
	}

	// 再次获取配置以验证
	t.Log("步骤3: 再次获取数据库配置以进行验证...")
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(getPayload))
	if err != nil {
		t.Errorf("验证配置请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析验证配置的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】验证数据库配置API返回错误! 响应: %+v", data)
	}
	dt, ok := data.Data.(map[string]interface{})
	if !ok || cast.ToString(dt["db_dsn"]) == "" {
		t.Error("【失败】验证数据库配置失败，获取到的 DSN 为空!")
	} else {
		t.Logf("【成功】数据库配置API测试通过! 验证DSN为: %s", dt["db_dsn"])
	}
	t.Log("【子测试 testDataBaseSet】--- 结束 ---")
}

func testPwdSet(t *testing.T) {
	t.Log("【子测试 testPwdSet】--- 开始 ---")
	getPayload := `{"action":"get","step":"password"}`
	setPayload := `{"action":"set","step":"password","account":"testCase","password":"testCase"}`

	// 获取配置
	t.Log("步骤1: 获取管理员密码配置（预期为空）...")
	ret, err := http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(getPayload))
	if err != nil {
		t.Errorf("获取密码请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析获取密码的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取管理员密码API返回错误! 响应: %+v", data)
	} else {
		t.Logf("获取密码配置成功: %+v", data)
	}

	// 设置配置
	t.Log("步骤2: 设置新的管理员密码...")
	t.Logf("请求体: %s", setPayload)
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(setPayload))
	if err != nil {
		t.Errorf("设置密码请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析设置密码的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】设置管理员密码API返回错误! 响应: %+v", data)
	} else {
		t.Log("设置新密码成功。")
	}

	// 再次获取配置以验证
	t.Log("步骤3: 再次获取管理员密码以验证...")
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(getPayload))
	if err != nil {
		t.Errorf("验证密码请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析验证密码的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】验证管理员密码API返回错误! 响应: %+v", data)
	}
	if cast.ToString(data.Data) != "testCase" {
		t.Errorf("【失败】验证管理员密码失败! 预期 'testCase'，实际 '%s'", cast.ToString(data.Data))
	} else {
		t.Log("【成功】管理员密码设置和验证通过!")
	}
	t.Log("【子测试 testPwdSet】--- 结束 ---")
}

func testDomainSet(t *testing.T) {
	t.Log("【子测试 testDomainSet】--- 开始 ---")
	getPayload := `{"action":"get","step":"domain"}`
	setPayload := `{"action":"set","step":"domain","smtp_domain":"test.domain","web_domain":"mail.test.domain"}`

	// 获取配置
	t.Log("步骤1: 获取当前域名配置...")
	ret, err := http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(getPayload))
	if err != nil {
		t.Errorf("获取域名请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析获取域名的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取域名配置API返回错误! 响应: %+v", data)
	} else {
		t.Logf("获取域名配置成功: %+v", data)
	}

	// 设置配置
	t.Log("步骤2: 设置新的域名配置...")
	t.Logf("请求体: %s", setPayload)
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(setPayload))
	if err != nil {
		t.Errorf("设置域名请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析设置域名的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】设置域名配置API返回错误! 响应: %+v", data)
	} else {
		t.Log("设置新域名成功。")
	}

	// 再次获取配置以验证
	t.Log("步骤3: 再次获取域名配置以验证...")
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(getPayload))
	if err != nil {
		t.Errorf("验证域名请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析验证域名的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】验证域名配置API返回错误! 响应: %+v", data)
	}
	dt, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Error("【失败】响应数据格式不正确。")
		return
	}
	if cast.ToString(dt["smtp_domain"]) != "test.domain" || cast.ToString(dt["web_domain"]) != "mail.test.domain" {
		t.Errorf("【失败】验证域名配置失败! 预期 smtp_domain='test.domain', web_domain='mail.test.domain', 实际 smtp_domain='%s', web_domain='%s'", dt["smtp_domain"], dt["web_domain"])
	} else {
		t.Log("【成功】域名配置API测试通过!")
	}
	t.Log("【子测试 testDomainSet】--- 结束 ---")
}

func testDNSSet(t *testing.T) {
	t.Log("【子测试 testDNSSet】--- 开始 ---")
	getPayload := `{"action":"get","step":"dns"}`
	t.Logf("请求URL: %s/api/setup, 请求体: %s", TestHost, getPayload)
	ret, err := http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(getPayload))
	if err != nil {
		t.Errorf("获取DNS请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析获取DNS的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取DNS配置API返回错误! 响应: %+v", data)
	} else {
		t.Logf("【成功】获取DNS配置成功! 响应: %+v", data)
	}
	t.Log("【子测试 testDNSSet】--- 结束 ---")
}

func testSSLSet(t *testing.T) {
	t.Log("【子测试 testSSLSet】--- 开始 ---")
	getPayload := `{"action":"get","step":"ssl"}`
	setPayload := `{"action":"set","step":"ssl","ssl_type":"1","key_path":"./config/ssl/private.key","crt_path":"./config/ssl/public.crt"}`

	// 获取配置
	t.Log("步骤1: 获取当前SSL配置...")
	ret, err := http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(getPayload))
	if err != nil {
		t.Errorf("获取SSL请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析获取SSL的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取SSL配置API返回错误! 响应: %+v", data)
	} else {
		t.Logf("获取SSL配置成功: %+v", data)
	}

	// 设置配置
	t.Log("步骤2: 设置新的SSL配置...")
	t.Logf("请求体: %s", setPayload)
	ret, err = http.Post(TestHost+"/api/setup", "application/json", strings.NewReader(setPayload))
	if err != nil {
		t.Errorf("设置SSL请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析设置SSL的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】设置SSL配置API返回错误! 响应: %+v", data)
	} else {
		t.Log("【成功】设置新SSL配置成功。测试完成！")
	}
	t.Log("【子测试 testSSLSet】--- 结束 ---")
}

func testLogin(t *testing.T) {
	t.Log("【子测试 testLogin】--- 开始测试管理员登录 ---")
	payload := `{"account":"testCase","password":"testCase"}`
	t.Logf("请求URL: %s/api/login, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/login", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("登录请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析登录响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】管理员登录API返回错误! 响应: %+v", data)
	} else {
		t.Logf("【成功】管理员登录成功! 响应: %+v", data)
	}
	t.Log("【子测试 testLogin】--- 结束 ---")
}

func testLoginUser2(t *testing.T) {
	t.Log("【子测试 testLoginUser2】--- 开始测试普通用户(user2)登录 ---")
	payload := `{"account":"user2","password":"user2New"}`
	t.Logf("请求URL: %s/api/login, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/login", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("登录请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析登录响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】用户 'user2' 登录API返回错误! 响应: %+v", data)
	} else {
		t.Logf("【成功】用户 'user2' 登录成功! 响应: %+v", data)
	}
	t.Log("【子测试 testLoginUser2】--- 结束 ---")
}

func testLoginUser3(t *testing.T) {
	t.Log("【子测试 testLoginUser3】--- 开始测试被封禁用户(user3)登录 ---")
	payload := `{"account":"user3","password":"user3"}`
	t.Logf("请求URL: %s/api/login, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/login", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("HTTP请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Fatalf("读取或解析响应体失败: %v", err)
	}
	if data.ErrorNo != 100 {
		t.Errorf("【失败】预期被封禁用户登录返回ErrorNo 100，但实际得到 %d。响应: %+v", data.ErrorNo, data)
	} else {
		t.Logf("【成功】被封禁用户 'user3' 登录失败，符合预期! 响应: %+v", data)
	}
	t.Log("【子测试 testLoginUser3】--- 结束 ---")
}

func testSendEmail(t *testing.T) {
	t.Log("【子测试 testSendEmail】--- 开始 ---")
	payload := `{
    "from": {"name": "i", "email": "i@test.domain"},
    "to": [{"name": "y", "email": "y@test.domain"}],
    "cc": [],
    "subject": "Title",
    "text": "text",
    "html": "<div>text</div>"
}`
	t.Logf("请求URL: %s/api/email/send, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("发送邮件请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析发送邮件的响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】发送邮件API返回错误! 响应: %+v", data)
	} else {
		t.Logf("【成功】发送邮件成功! 响应: %+v", data)
	}
	t.Log("【子测试 testSendEmail】--- 结束 ---")
}

func testSendEmail2User2ForSpam(t *testing.T) {
	t.Log("【子测试 testSendEmail2User2ForSpam】--- 开始 ---")
	payload := `{
    "from": {"name": "user2", "email": "user2@test.domain"},
    "to": [{"name": "y", "email": "admin@test.domain"}],
    "subject": "spam",
    "text": "NeedMove",
    "html": "<div>text</div>"
}`
	t.Logf("请求URL: %s/api/email/send, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("发送邮件请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】发送邮件API返回错误! 响应: %+v", data)
	} else {
		t.Logf("【成功】发送邮件(用于Spam规则)成功! 响应: %+v", data)
	}
	t.Log("【子测试 testSendEmail2User2ForSpam】--- 结束 ---")
}

func testSendEmail2User2ForMove(t *testing.T) {
	t.Log("【子测试 testSendEmail2User2ForMove】--- 开始 ---")
	payload := `{
    "from": {"name": "user2", "email": "user2@test.domain"},
    "to": [{"name": "y", "email": "user2@test.domain"}],
    "subject": "MovePlease",
    "text": "NeedMove",
    "html": "<div>text</div>"
}`
	t.Logf("请求URL: %s/api/email/send, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("发送邮件请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】发送邮件API返回错误! 响应: %+v", data)
	} else {
		t.Logf("【成功】发送邮件(用于Move规则)成功! 响应: %+v", data)
	}
	t.Log("【子测试 testSendEmail2User2ForMove】--- 结束 ---")
}

func genTestEmailData(t *testing.T) {
	t.Log("【子测试 genTestEmailData】--- 开始生成10封测试邮件 ---")
	for i := 0; i < 10; i++ {
		payload := fmt.Sprintf(`{
        "from": {"name": "user2", "email": "user2@test.domain"},
        "to": [{"name": "admin", "email": "admin@test.domain"}],
        "subject": "测试邮件%d",
        "text": "测试邮件%d",
        "html": "<div>测试邮件%d</div>"
    }`, i, i, i)
		t.Logf("正在发送第 %d/10 封测试邮件...", i+1)
		ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(payload))
		if err != nil {
			t.Errorf("发送第 %d 封邮件请求失败: %v", i+1, err)
			continue
		}
		data, err := readResponse(ret.Body)
		if err != nil {
			t.Errorf("读取或解析第 %d 封邮件的响应失败: %v", i+1, err)
			continue
		}
		if data.ErrorNo != 0 {
			t.Errorf("【失败】发送第 %d 封邮件API返回错误! 响应: %+v", i+1, data)
		}
		time.Sleep(3 * time.Second)
	}
	t.Log("【子测试 genTestEmailData】--- 10封测试邮件发送完毕 ---")
}

func testSendEmail2User1(t *testing.T) {
	t.Log("【子测试 testSendEmail2User1】--- 开始 ---")
	payload := `{
    "from": {"name": "i", "email": "i@test.domain"},
    "to": [{"name": "y", "email": "user1@test.domain"}],
    "subject": "HelloUser1",
    "text": "text",
    "html": "<div>text</div>"
}`
	t.Logf("请求URL: %s/api/email/send, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("发送邮件请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】发送邮件API返回错误! 响应: %+v", data)
	} else {
		t.Logf("【成功】发送邮件至 'user1' 成功! 响应: %+v", data)
	}
	t.Log("【子测试 testSendEmail2User1】--- 结束 ---")
}

func testSendEmail2User2(t *testing.T) {
	t.Log("【子测试 testSendEmail2User2】--- 开始 ---")
	payload := `{
    "from": {"name": "i", "email": "i@test.domain"},
    "to": [{"name": "y", "email": "user2@test.domain"}],
    "subject": "HelloUser2",
    "text": "text",
    "html": "<div>text</div>"
}`
	t.Logf("请求URL: %s/api/email/send, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("发送邮件请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】发送邮件API返回错误! 响应: %+v", data)
	} else {
		t.Logf("【成功】发送邮件至 'user2' 成功! 响应: %+v", data)
	}
	t.Log("【子测试 testSendEmail2User2】--- 结束 ---")
}

func testSendEmail2User3(t *testing.T) {
	t.Log("【子测试 testSendEmail2User3】--- 开始 ---")
	payload := `{
    "from": {"name": "i", "email": "i@test.domain"},
    "to": [{"name": "y", "email": "user3@test.domain"}],
    "subject": "HelloUser3",
    "text": "text",
    "html": "<div>text</div>"
}`
	t.Logf("请求URL: %s/api/email/send, 请求体: %s", TestHost, payload)
	ret, err := httpClient.Post(TestHost+"/api/email/send", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("发送邮件请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】发送邮件API返回错误! 响应: %+v", data)
	} else {
		t.Logf("【成功】发送邮件至 'user3' 成功! 响应: %+v", data)
	}
	t.Log("【子测试 testSendEmail2User3】--- 结束 ---")
}

func testEmailList(t *testing.T) {
	t.Log("【子测试 testEmailList】--- 开始 ---")
	t.Logf("请求URL: %s/api/email/list", TestHost)
	ret, err := httpClient.Post(TestHost+"/api/email/list", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Errorf("获取邮件列表请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取邮件列表API返回错误! 响应: %+v", data)
	}
	dt, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Error("【失败】响应数据格式不正确。")
		return
	}
	list, ok := dt["list"].([]interface{})
	if !ok || len(list) == 0 {
		t.Error("【失败】邮件列表为空!")
		return
	}
	item, ok := list[0].(map[string]interface{})
	if !ok {
		t.Error("【失败】邮件列表项格式不正确。")
		return
	}
	id := cast.ToInt(item["id"])
	if id == 0 {
		t.Error("【失败】邮件列表数据错误，ID为0!")
	} else {
		t.Logf("【成功】获取邮件列表成功! 列表不为空，第一封邮件ID为 %d。响应: %+v", id, data)
	}
	t.Log("【子测试 testEmailList】--- 结束 ---")
}

func testUser2EmailList(t *testing.T) {
	t.Log("【子测试 testUser2EmailList】--- 开始 ---")
	t.Logf("请求URL: %s/api/email/list", TestHost)
	ret, err := httpClient.Post(TestHost+"/api/email/list", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Errorf("获取邮件列表请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取邮件列表API返回错误! 响应: %+v", data)
	}
	dt, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Error("【失败】响应数据格式不正确。")
		return
	}
	list, ok := dt["list"].([]interface{})
	if !ok || len(list) != 1 {
		t.Errorf("【失败】邮件列表数量不为1! 实际数量: %d", len(list))
	} else {
		t.Logf("【成功】获取用户 'user2' 的邮件列表成功! 数量为1，符合预期。响应: %+v", data)
	}
	t.Log("【子测试 testUser2EmailList】--- 结束 ---")
}

func testDelEmail(t *testing.T) {
	t.Log("【子测试 testDelEmail】--- 开始 ---")
	// 先获取列表找到一个邮件ID
	t.Log("步骤1: 获取邮件列表以找到要删除的邮件ID...")
	ret, err := httpClient.Post(TestHost+"/api/email/list", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Errorf("获取邮件列表请求失败: %v", err)
	}
	data, err := readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】获取邮件列表API返回错误! 响应: %+v", data)
		return
	}
	dt, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Error("【失败】响应数据格式不正确。")
		return
	}
	list, ok := dt["list"].([]interface{})
	if !ok || len(list) == 0 {
		t.Error("【失败】邮件列表为空，无法执行删除操作!")
		return
	}
	item := list[0].(map[string]interface{})
	id := cast.ToInt(item["id"])
	t.Logf("找到要删除的邮件ID: %d", id)

	// 删除邮件
	t.Log("步骤2: 删除邮件...")
	payload := fmt.Sprintf(`{"ids":[%d]}`, id)
	t.Logf("请求URL: %s/api/email/del, 请求体: %s", TestHost, payload)
	ret, err = httpClient.Post(TestHost+"/api/email/del", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Errorf("删除邮件请求失败: %v", err)
	}
	data, err = readResponse(ret.Body)
	if err != nil {
		t.Errorf("读取或解析响应失败: %v", err)
	}
	if data.ErrorNo != 0 {
		t.Errorf("【失败】删除邮件API返回错误! 响应: %+v", data)
	} else {
		t.Log("删除邮件API调用成功。")
	}

	// 验证数据库状态
	t.Log("步骤3: 验证数据库中邮件的状态...")
	var mail models.UserEmail
	db.Instance.Where("email_id = ?", id).Get(&mail)
	if mail.Status != 3 {
		t.Errorf("【失败】数据库验证失败! 邮件ID %d 的状态应为3(已删除)，但实际为 %d", id, mail.Status)
	} else {
		t.Logf("【成功】删除邮件测试成功! 数据库中邮件状态已更新为3。响应: %+v", data)
	}
	t.Log("【子测试 testDelEmail】--- 结束 ---")
}

// portCheck 检查端口是否被占用
func portCheck(port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(port)))
	if err != nil {
		// 端口被占用时，Listen会返回错误
		return true
	}
	// 如果能成功监听，说明端口未被占用，立即关闭
	defer l.Close()
	return false
}

func readResponse(r io.ReadCloser) (*response.Response, error) {
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		log.Printf("【辅助函数 readResponse】读取响应体失败: %v", err)
		return nil, err
	}

	log.Printf("【辅助函数 readResponse】收到的原始响应体: %s", string(data))

	ret := &response.Response{}
	err = json.Unmarshal(data, ret)
	if err != nil {
		log.Printf("【辅助函数 readResponse】JSON反序列化失败: %v", err)
		// 返回一个包含原始数据的错误，方便调试
		return nil, fmt.Errorf("JSON unmarshal error: %w, raw response: %s", err, string(data))
	}
	return ret, nil
}
