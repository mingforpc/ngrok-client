package ngrokc

import (
	"fmt"
	"ngrok-client/ngrokc/config"
	"ngrok-client/ngrokc/connection"
	"os"
	"os/signal"
	"syscall"
)

// Ngrok client的启动函数
func Start() {

	// 异常退出时的处理
	defer exceptionPrecess()

	// 配置文件的解析
	config.ParseConfig()

	var ccon = connection.ControlConnection{}

	// 处理关闭信号
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	go exit(signalChan, &ccon)

	// 初始化 control connection
	ccon.Init(config.CONFIG.ServerHostname, config.CONFIG.ServerPort, config.CONFIG.User, config.CONFIG.Password)
	// 设置HTTP的配置
	ccon.SetHTTPConfig(config.CONFIG.HttpHostname, config.CONFIG.HttpSubdomain, config.CONFIG.HttpAuth, config.CONFIG.HttpLocalPort)
	// 设置HTTPS的配置
	ccon.SetHTTPSConfig(config.CONFIG.HttpsHostname, config.CONFIG.HttpsSubdomain, config.CONFIG.HttpsAuth, config.CONFIG.HttpsLocalPort)
	// 开始服务
	err := ccon.Service()

	fmt.Println(err)

}

func exit(signalChan chan os.Signal, ccon *connection.ControlConnection) {

	sign := <-signalChan

	fmt.Println(sign)

	ccon.Close()
}

func exceptionPrecess() {
	p := recover()
	switch p {
	case nil:
	default:
		panic(p)
	}
}
