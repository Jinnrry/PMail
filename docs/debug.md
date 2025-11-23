测试imap协议返回

`openssl s_client  -crlf  -connect imap.xxxx.com:993`
`openssl s_client  -crlf  -connect 127.0.0.1:993`


删除邮件
```bash
A1 LOGIN testCase testCase
A3 SELECT "Deleted Messages" 
A4 UID STORE 1:* +FLAGS.SILENT (\Deleted) 
A5 EXPUNGE
```

搜索邮件
```bash
A1 LOGIN testCase testCase
114 SELECT "INBOX"
115 UID SEARCH 1:5 NOT DELETED
```


执行全部测试用例(linux macos不加sudo没有权限监听小于1024的端口，因此会失败)
`sudo env "PATH=$PATH" make test` 

执行单个测试用例
`go test -v -run ^Test ./services/del_email`