package consts

const (
	// EmailTypeSend 发信
	EmailTypeSend int8 = 1
	// EmailTypeReceive 收信
	EmailTypeReceive int8 = 0

	//EmailStatusWait 0未发送
	EmailStatusWait int8 = 0

	//EmailStatusSent 1已发送
	EmailStatusSent int8 = 1

	//EmailStatusFail 2发送失败
	EmailStatusFail int8 = 2

	//EmailStatusDel 3删除
	EmailStatusDel int8 = 3
)
