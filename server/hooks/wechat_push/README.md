## How To Ues / 如何使用


Copy plugin binary file to `/plugins` 

复制插件二进制文件到`/plugins`文件夹

add config.json to `/plugins/config.com` like this:

新建配置文件`/plugins/config.com`，内容如下

```jsonc
{
  "weChatPushAppId": "", // wechat appid
  "weChatPushSecret": "", // weChat  Secret
  "weChatPushTemplateId": "", // weChat TemplateId
  "weChatPushUserId": "", // weChat UserId
}
```

WeChat Message Template :

微信推送模板设置：

Template Title: New Email Notice

模板标题：新邮件提醒

Template Content: {{Content.DATA}}

模板内容：{{Content.DATA}}