var lang = {
    "lang": "en",
    "submit": "submit",
    "compose": "compose",
    "new": "new",
    "account": "Account",
    "password": "Password",
    "login": "login",
    "search": "Search Email",
    "inbox": "Inbox",
    "sender": "Sender",
    "title": "Title",
    "date": "Date",
    "to": "To:",
    "cc": "Cc:",
    "sender_desc": "Only the email prefix is required",
    "to_desc": "Recipient's e-mail address",
    "cc_desc": "Cc's e-mail address",
    "send": "send",
    "add_att": "Add Attachment",
    "attachment":"Attachment",
    "err_sender_must": "Sender's email prefix is required!",
    "only_prefix": "Only the email prefix is required!",
    "err_email_format": "Incorrect e-mail address, please check the e-mail format!",
    "err_title_must": "Title is required!",
    "succ_send": "Send Success!",
    "outbox": "outbox",
    "modify_pwd": "modify password",
    "enter_again": "enter again",
    "err_required_pwd": "Please Input Password!",
    "succ": "Success!",
    "err_pwd_diff": "The passwords entered twice do not match!",
    "fail": "Fail!",
    "settings":"Settings",
    "security":"Security"
};



var zhCN = {
    "lang": "zhCn",
    "submit": "提交",
    "compose": "发件",
    "new": "新",
    "account": "用户名",
    "password": "密码",
    "login": "登录",
    "search": "搜索邮件",
    "inbox": "收件箱",
    "sender": "发件人",
    "title": "主题",
    "date": "时间",
    "to": "收件人:",
    "cc": "抄送:",
    "sender_desc": "只需要邮箱前缀",
    "to_desc": "接收人邮件地址",
    "cc_desc": "抄送人邮箱地址",
    "send": "发送",
    "add_att": "添加附件",
    "attachment":"附件",
    "err_sender_must": "发件人邮箱前缀必填",
    "only_prefix": "只需要邮箱前缀",
    "err_email_format": "邮箱地址错误，请检查邮箱格式！",
    "err_title_must": "标题必填！",
    "succ_send": "发送成功!",
    "outbox": "发件箱",
    "modify_pwd": "修改密码",
    "enter_again": "确认密码",
    "err_required_pwd": "请输入密码!",
    "succ": "成功!",
    "err_pwd_diff": "两次输入的密码不一致!",
    "fail": "失败",
    "settings":"设置",
    "security":"安全"
}

switch (navigator.language) {
    case "zh":
        lang = zhCN
        break
    case "zh-CN":
        lang = zhCN
        break
    default:
        break
}



export default lang;