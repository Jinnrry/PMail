# PMail

> The current code is not stable, be sure to record the log! Lost letters or letters parsed wrong can find out the
> original content of the mail from the log!

## [中文文档](./README_CN.md)

## Introduction

PMail is a personal email server that pursues a minimal deployment process and extreme resource consumption. It runs on
a single file and contains complete send/receive mail service and web-side mail management functions. Just a server , a
domain name , a line of code , a minute of deployment time , you will be able to build a domain name mailbox of your
own .

Any project related Issue, PR is welcome.At present, the project UI design is ugly, UI interaction experience is poor,
welcome all UI, designers, front-end guidance. Finally, also for this project to solicit a beautiful and lovely Logo!

<img src="./docs/en.gif" alt="Editor" width="800px">

## Features

* Single file operation and easy deployment.

* The binary file is only 15MB and takes up less than 10M of memory during the run.

* Support dkim, spf checksum, [Email Test](https://www.mail-tester.com/) score 10 points if correctly configured.

* Implementing the ACME protocol, the program will automatically obtain and update Let's Encrypt certificates.

## Disadvantages

* At present, only the core function of sending and receiving emails has been completed. Basically, it can only be used
  by a single person, and does not deal with issues related to permission management in the process of multiple users.

* The UI is ugly

# How to run

## 1、Download

[Click Here](https://github.com/Jinnrry/PMail/releases) Download a program file that matches you.

## 2、Run

`double-click to open` Or `execute command to run`

## 3、Configuration

Open `http://127.0.0.1` in your browser or use your server's public IP to visit, then follow the instructions to
configure.

## 4、Email Test

Check if your mailbox has completed all the security configuration. It is recommended to
use [https://www.mail-tester.com/](https://www.mail-tester.com/) for checking.

## 5、 WeChat Message Push

Open the `config/config.json` file in the run directory, edit a few configuration items at the beginning of `weChatPush`
and restart the service.

# For Developer

## Project Framework

1、 FE： vue3+element-plus

The code is in `fe` folder.

2、Server： golang + mysql

The code is in `server` folder.

## Plugin Development

Reference this file. `server/hooks/wechat_push/wechat_push.go`
