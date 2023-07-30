# PMail 

> The current code is not stable, be sure to record the log! Lost letters or letters parsed wrong can find out the original content of the mail from the log!

## [中文文档](./README_CN.md)

## Introduction

An extremely lightweight mailbox server designed for personal use scenarios. 

## Features

* Single file operation and easy deployment.

* The binary file is only 15MB and takes up less than 10M of memory during the run.

* Support dkim, spf checksum, [Email Test](https://www.mail-tester.com/) score 10 points if correctly configured.

## Disadvantages

* At present, only the core function of sending and receiving emails has been completed. Basically, it can only be used by a single person, and does not deal with issues related to permission management in the process of multiple users.

* The UI is ugly

# How to run

## 1、Generate DKIM secret key

Generate public and private keys by the dkim-keygen tool of the [go-msgauth](https://github.com/emersion/go-msgauth) project

Put the key in the `config/dkim` directory.

## 2、Set DNS

Add the following records to your domain DNS settings

| type | hostname             | address / value      |
|------|----------------------|----------------------|
| A    | smtp                 | server ip            |
| MX   | _                    | smtp.YourDomain      |
| TXT  | _                    | v=spf1 a mx ~all     |
| TXT  | default._domainkey	  | Your DKIM public key |

## 3、Domain SSL Key

Prepare the certificate of `smtp.YourDomain`, the private key in ".key" format and the public key in ".crt" format

Put the certificate in the `config/ssl` directory.

## 4、Build（or download）

1、installed `nodejs` and `yarn`

2、installed `golang`

3、exec `./build.sh`

## 5、Config

Modify the `config.json` file in the config directory and fill in your secret key and domain information.

Tips:

MySQL database name must is `pmail`, and charset must is `utf8_general_ci`.

Configuration file description ：
```json
{
  "domain": "demo.com", // Your domain
  "dkimPrivateKeyPath": "config/dkim/dkim.priv",  // dkim private key
  "SSLPrivateKeyPath": "config/ssl/private.key",  // ssl private key
  "SSLPublicKeyPath": "config/ssl/public.crt",    // ssl public key
  "mysqlDSN": "username:password@tcp(127.0.0.1:3306)/pmail?parseTime=True&loc=Local", // mysql connect infonation
  "weChatPushAppId": "",  // WeChat public account appid (for new email message push) . If you don't need it, you can make it empty.
  "weChatPushSecret": "",   // WeChat api secret
  "weChatPushTemplateId": "",  // push template id
  "weChatPushUserId": "" // wechat user id
}
```

## 6、Run

exec `pmail` and check port of 25、80.

The webmail service address is http://yourip. Default account is `admin` and password is `admin`

## 7、Email Test

Check if your mailbox has completed all the security configuration. It is recommended to use [https://www.mail-tester.com/](https://www.mail-tester.com/) for checking. 


# For Developer

## Project Framework

1、 FE： vue3+element-plus

The code is in `fe` folder.

2、Server： golang + mysql

The code is in `server` folder.

## Plugin Development

Reference this file. `server/hooks/wechat_push/wechat_push.go`

# What's More

Welcome PR! Welcome Issues! The project need a Logo !