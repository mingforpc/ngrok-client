package main

import (
	"syscall"
	"fmt"
	"os"
	"os/signal"
	"ngrok-client/ngrokc/config"
	"ngrok-client/ngrokc/connection"
)

func main() {

	config.ParseConfig()

	var ccon = connection.ControlConnection{}

	// 除了关闭信号
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	go exit(signalChan, &ccon)

	ccon.Init(config.CONFIG.ServerHostname, config.CONFIG.ServerPort, config.CONFIG.User, config.CONFIG.Password)
	ccon.SetHTTPConfig(config.CONFIG.HttpHostname, config.CONFIG.HttpSubdomain, config.CONFIG.HttpAuth, config.CONFIG.HttpLocalPort)
	ccon.SetHTTPSConfig(config.CONFIG.HttpsHostname, config.CONFIG.HttpsSubdomain, config.CONFIG.HttpsAuth, config.CONFIG.HttpsLocalPort)
	ccon.Service()

}

func exit(signalChan chan os.Signal, ccon *connection.ControlConnection) {

	sign := <- signalChan

	fmt.Println(sign)

	ccon.Close()
}