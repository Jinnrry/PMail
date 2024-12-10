package goimap

type Action interface {
	Create(session *Session, path string) error                         // 创建邮箱
	Delete(session *Session, path string) error                         // 删除邮箱
	Rename(session *Session, oldPath, newPath string) error             // 重命名邮箱
	List(session *Session, basePath, template string) ([]string, error) // 浏览邮箱
	Append(session *Session, item string) error                         // 上传邮件
	Select(session *Session, path string) error                         // 选择邮箱
	/*
		读取邮件的文本信息，且仅用于显示的目的。
			ALL：只返回按照一定格式的邮件摘要，包括邮件标志、RFC822.SIZE、自身的时间和信封信息。IMAP客户机能够将标准邮件解析成这些信息并显示出来。
			BODY：只返回邮件体文本格式和大小的摘要信息。IMAP客户机可以识别这些细腻，并向用户显示详细的关于邮件的信息。其实是一些非扩展的BODYSTRUCTURE的信息。
			FAST：只返回邮件的一些摘要，包括邮件标志、RFC822.SIZE、和自身的时间。
			FULL：同样的还是一些摘要信息，包括邮件标志、RFC822.SIZE、自身的时间和BODYSTRUCTURE的信息。
			BODYSTRUCTUR：是邮件的[MIME-IMB]的体结构。这是服务器通过解析[RFC-2822]头中的[MIME-IMB]各字段和[MIME-IMB]头信息得出来的。包括的内容有：邮件正文的类型、字符集、编码方式等和各附件的类型、字符集、编码方式、文件名称等等。
			ENVELOPE：信息的信封结构。是服务器通过解析[RFC-2822]头中的[MIME-IMB]各字段得出来的，默认各字段都是需要的。主要包括：自身的时间、附件数、收件人、发件人等。
			FLAGS：此邮件的标志。
			INTERNALDATE：自身的时间。
			RFC822.SIZE：邮件的[RFC-2822]大小
			RFC822.HEADER：在功能上等同于BODY.PEEK[HEADER]，
			RFC822：功能上等同于BODY[]。
			RFC822.TEXT：功能上等同于BODY[TEXT]
			UID：返回邮件的UID号，UID号是唯一标识邮件的一个号码。
			BODY[section] <<partial>>：返回邮件的中的某一指定部分，返回的部分用section来表示，section部分包含的信息通常是代表某一部分的一个数字或者是下面的某一个部分：HEADER, HEADER.FIELDS, HEADER.FIELDS.NOT, MIME, and TEXT。如果section部分是空的话，那就代表返回全部的信息，包括头信息。
			BODY[HEADER]返回完整的文件头信息。
			BODY[HEADER.FIELDS ()]：在小括号里面可以指定返回的特定字段。
			BODY[HEADER.FIELDS.NOT ()]：在小括号里面可以指定不需要返回的特定字段。
			BODY[MIME]：返回邮件的[MIME-IMB]的头信息，在正常情况下跟BODY[HEADER]没有区别。
			BODY[TEXT]：返回整个邮件体，这里的邮件体并不包括邮件头。
			**/
	Fetch(session *Session, mailIds, dataNames string) (string, error)
	Store(session *Session, mailId, flags string) error            // STORE 命令用于修改指定邮件的属性，包括给邮件打上已读标记、删除标记
	Close(session *Session) error                                  // 关闭文件夹
	Expunge(session *Session) error                                // 删除已经标记为删除的邮件，释放服务器上的存储空间
	Examine(session *Session, path string) error                   // 只读方式打开邮箱
	Subscribe(session *Session, path string) error                 // 活动邮箱列表中增加一个邮箱
	UnSubscribe(session *Session, path string) error               // 活动邮箱列表中去掉一个邮箱
	LSub(session *Session, path, mailbox string) ([]string, error) // 显示那些使用SUBSCRIBE命令设置为活动邮箱的文件
	/*
		@category:
			MESSAGES	邮箱中的邮件总数
			RECENT	邮箱中标志为\RECENT的邮件数
			UIDNEXT	可以分配给新邮件的下一个UID
			UIDVALIDITY	邮箱的UID有效性标志
			UNSEEN	邮箱中没有被标志为\UNSEEN的邮件数
	*/
	Status(session *Session, mailbox, category string) (string, error) // 查询邮箱的当前状态
	Check(session *Session) error                                      // sync数据
	Search(session *Session, keyword, criteria string) (string, error) // 命令可以根据搜索条件在处于活动状态的邮箱中搜索邮件，然后显示匹配的邮件编号
	Copy(session *Session, mailId, mailBoxName string) error           // 把邮件从一个邮箱复制到另一个邮箱
	CapaBility(session *Session) ([]string, error)                     // 返回IMAP服务器支持的功能列表
	Noop(session *Session) error                                       // 什么都不做，连接保活
	Login(session *Session, username, password string) error           // 登录
	Logout(session *Session) error                                     // 注销登录
	Custom(session *Session, cmd string, args []string) ([]string, error)
}
