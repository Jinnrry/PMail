package signal

// InitChan 控制初始化流程结束
var InitChan = make(chan bool)

// RestartChan 控制程序重启
var RestartChan = make(chan bool)

// StopChan 控制程序结束
var StopChan = make(chan bool)
