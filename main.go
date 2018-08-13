package main

import (
	"syscall"
	"fmt"
	"os"
	"os/signal"
	"flag"
	"encoding/json"
	"ngrok-client/ngrokc/config"
	"ngrok-client/ngrokc/connection"
)

var configFile = flag.String("config", "", "config file path")

// Server config
var serverHostname = flag.String("server_hostname", "", "Server hostname, IP or domain name")
var serverPort = flag.Int("server_port", 0, "server port")
var username = flag.String("user", "", "username to register")
var password = flag.String("password", "", "password of username")

// Http proxy config
var httpHostname = flag.String("http_hostname", "", "Http hostname, IP or domain name, can be null")
var httpSubdomain = flag.String("http_subdomain", "", "Http subdomian name, some server maybe not accept, can be null")
var httpLocalPort = flag.Int("http_local_port", 0, "Local http port")

// Https proxy config
var httpsHostname = flag.String("https_hostname", "", "Https hostname, IP or domain name, can be null")
var httpsSubdomain = flag.String("https_subdomain", "", "Https subdomian name, some server maybe not accept, can be null")
var httpsLocalPort = flag.Int("https_local_port", 0, "Local https port")

var readBufSize = flag.Int("read_buf_size", 0, "")

// parseConfig() 从配置文件和命令行中解析配置，优先选择命令行中的配置
func parseConfig() {

	flag.Parse()

	// 配置文件
	if *configFile != "" {

		parseConfigFile(*configFile, config.CONFIG)

	}

	if *serverHostname != "" {
		config.CONFIG.ServerHostname = *serverHostname
	}

	if *serverPort > 0 {
		config.CONFIG.ServerPort = uint(*serverPort)
	}

	if *username != "" {
		config.CONFIG.User = *username
	}

	if *password != "" {
		config.CONFIG.Password = *password
	}

	if *httpHostname != "" {
		config.CONFIG.HttpHostname = *httpHostname
	}

	if *httpSubdomain != "" {
		config.CONFIG.HttpSubdomain = *httpSubdomain
	}

	if *httpLocalPort > 0 {
		config.CONFIG.HttpLocalPort = uint(*httpLocalPort)
	}

	if *httpsHostname != "" {
		config.CONFIG.HttpsHostname = *httpsHostname
	}

	if *httpsSubdomain != "" {
		config.CONFIG.HttpsSubdomain = *httpsSubdomain
	}

	if *httpsLocalPort > 0 {
		config.CONFIG.HttpsLocalPort = uint(*httpsLocalPort)
	}

	if *readBufSize > 0 {
		config.CONFIG.ReadBufSize = uint(*readBufSize)
	}

}

// parseConfigFile() 从指定的配置文件中读取配置.
func parseConfigFile(filepath string, conf *config.Configuration) {
	file, err := os.Open(filepath)
	
	if err != nil {
		fmt.Println("config file error:" + err.Error())
		return
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&conf)

	if err != nil {
		fmt.Println("config file parse error:" + err.Error())
		return
	}
}

func main() {

	parseConfig()

	var ccon = connection.ControlConnection{}

	// 除了关闭信号
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	go exit(signalChan, &ccon)

	ccon.Init(config.CONFIG.ServerHostname, config.CONFIG.ServerPort, config.CONFIG.User, config.CONFIG.Password)
	ccon.SetHTTPConfig(config.CONFIG.HttpHostname, config.CONFIG.HttpSubdomain, config.CONFIG.HttpAuth, config.CONFIG.HttpLocalPort)
	ccon.SetHTTPSConfig(config.CONFIG.HttpsHostname, config.CONFIG.HttpsSubdomain, config.CONFIG.HttpsAuth, config.CONFIG.HttpsLocalPort)
	fmt.Println(ccon.Service())

}

func exit(signalChan chan os.Signal, ccon *connection.ControlConnection) {

	sign := <- signalChan

	fmt.Println(sign)

	ccon.Close()
}