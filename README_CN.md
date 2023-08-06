# PMail 

> Welcome PR! Welcome Issues! 目前代码并不稳定，一定记录好日志！丢信或者信件解析错误可以从日志中找出邮件原始内容！

## 为什么写这个项目

迫于越来越多的邮件服务商暂停了针对个人的域名邮箱服务（比如QQ邮箱、微软Outlook邮箱），因此考虑自建域名邮箱服务。
但是自建域名邮箱可选的程序并不多，且目标都不是针对个人使用场景设计的。个人服务器一般内存、CPU、硬盘配置都不高，针对公司场景使用的邮箱程序过于臃肿，
白白浪费资源。就拿我自己的服务器来说，我服务器配置为1核512M 10G硬盘，市面上绝大多数邮箱服务器安装上就把磁盘占满了，根本没法正常使用

## 项目优势

### 1、部署简单

使用Go语言编写，支持跨平台，编译后单文件运行，单文件包含完整的前后端代码。修改配置文件，运行即可。

### 2、资源占用极小

编译后二进制文件仅15MB，运行过程中占用内存10M以内。

### 3、安全方面

支持dkim、spf校验。正确配置的情况下，Email Test得分10分。

## 其他

### 不足

1、目前只完成了最核心的收发邮件功能。基本上仅针对单人使用，没有处理多人使用、权限管理相关问题。

2、目前代码并不稳定，可能存在BUG

3、前端UI交互较差


# 如何部署

## 1、生成DKIM 秘钥

```
go install github.com/emersion/go-msgauth/cmd/dkim-keygen@latest
dkim-keygen
```
执行后将得到`dkim.priv`文件，公钥数据会直接输出

生成以后将密钥放到`config/dkim`目录中

## 2、设置域名DNS

添加以下记录到你到域名解析中

| 类型  | 主机记录                | 记录值              |
|-----|---------------------|------------------|
| A   | smtp                | 服务器IP            |
| MX  | _                   | smtp.你的域名        |
| TXT | _                   | v=spf1 a mx ~all |
| TXT | default._domainkey	 | 你生成的DKIM公钥       |

## 3、申请域名证书

准备好 `smtp.你的域名` 的证书，key格式的私钥和crt格式的公钥

放到`config/ssl`目录中

## 4、编译程序（或者直接下载编译好的二进制文件）

1、前端环境：安装好node环境，配置好yarn

2、后端环境：安装最新的golang

3、执行`./build.sh`

## 5、修改配置文件

修改config目录中的`config.json`文件，填入你的秘钥与域名信息

Tips:

MySQL库名必须叫pmail，另外，数据库必须使用utf8_general_ci字符集

配置文件说明：
```json
{
  "domain": "demo.com", // 你的域名
  "dkimPrivateKeyPath": "config/dkim/dkim.priv",  // dkim私钥
  "SSLPrivateKeyPath": "config/ssl/private.key",  // ssl证书私钥
  "SSLPublicKeyPath": "config/ssl/public.crt",    // ssl证书公钥
  "mysqlDSN": "username:password@tcp(127.0.0.1:3306)/pmail?parseTime=True&loc=Local", // mysql连接信息
  "weChatPushAppId": "",  //微信公众号id（用于新消息提醒），没有留空即可
  "weChatPushSecret": "",   // 微信公众号api秘钥
  "weChatPushTemplateId": "",  // 微信公众号推送模板id
  "weChatPushUserId": "" // 微信推送用户id
}
```

## 6、启动

运行`PMail`程序，检查服务器25、80端口正常即可

邮箱后台, http://yourip，默认账号admin，默认密码admin

## 7、邮箱得分测试

建议找一下邮箱测试服务(比如[https://www.mail-tester.com/](https://www.mail-tester.com/))进行邮件得分检测，避免自己某些步骤漏配，导致发件进对方垃圾箱。

## 8、其他说明

邮件是否进对方垃圾箱与程序无关、与你的服务器IP、服务器域名有关。我自己搭建的服务，测试了收发QQ、Gmail、Outlook、163、126均正常，无任何拦截，且不会进垃圾箱。


# 参与开发

## 项目架构

1、前端： vue3+element-plus

前端代码位于`fe`目录中，运行参考`fe`目录中的README文件

2、后端： golang + mysql

后端代码进入`server`文件夹，运行`main.go`文件

## 插件开发

参考微信推送插件`server/hooks/wechat_push/wechat_push.go`

# 最后

欢迎PR! 欢迎Issue！求个Logo！