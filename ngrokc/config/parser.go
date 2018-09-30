package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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

var readBufSize = flag.Int("read_buf_size", 0, "Socket read buffer size")

// 最大Proxy连接数限制
var maxProxyCount = flag.Int64("max_proxy_count", 10, "Proxy connection max count")

// parseConfigFile() 从指定的配置文件中读取配置.
func ParseConfigFile(filepath string, conf *Configuration) {
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

// parseConfig() 从配置文件和命令行中解析配置，优先选择命令行中的配置
func ParseConfig() {

	flag.Parse()

	// 配置文件
	if *configFile != "" {

		ParseConfigFile(*configFile, CONFIG)

	}

	if *serverHostname != "" {
		CONFIG.ServerHostname = *serverHostname
	}

	if *serverPort > 0 {
		CONFIG.ServerPort = uint(*serverPort)
	}

	if *username != "" {
		CONFIG.User = *username
	}

	if *password != "" {
		CONFIG.Password = *password
	}

	if *httpHostname != "" {
		CONFIG.HttpHostname = *httpHostname
	}

	if *httpSubdomain != "" {
		CONFIG.HttpSubdomain = *httpSubdomain
	}

	if *httpLocalPort > 0 {
		CONFIG.HttpLocalPort = uint(*httpLocalPort)
	}

	if *httpsHostname != "" {
		CONFIG.HttpsHostname = *httpsHostname
	}

	if *httpsSubdomain != "" {
		CONFIG.HttpsSubdomain = *httpsSubdomain
	}

	if *httpsLocalPort > 0 {
		CONFIG.HttpsLocalPort = uint(*httpsLocalPort)
	}

	if *readBufSize > 0 {
		CONFIG.ReadBufSize = uint(*readBufSize)
	}

	if CONFIG.MaxProxyCount <= 0 {
		CONFIG.MaxProxyCount = *maxProxyCount
	}

}
