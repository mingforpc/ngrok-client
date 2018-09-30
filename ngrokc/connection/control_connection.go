package connection

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"ngrok-client/ngrokc/config"
	errcode "ngrok-client/ngrokc/err"
	"ngrok-client/ngrokc/util"
	"strconv"
)

type ControlConnection struct {
	ServerDomain string // 域名或者IP
	ServerPort   uint
	User         string
	Password     string
	Arch         string
	ClientId     string

	// 分配的HTTP信息
	HTTPHostname  string
	HTTPSubdomain string
	HTTPAuth      string
	HTTPLocalPort uint
	// 服务器返回的HTTP URL
	HTTPUrl string

	// 分配的HTTPS信息
	HTTPSHostname  string
	HTTPSSubdomain string
	HTTPSAuth      string
	HTTPSLocalPort uint
	// 服务器返回的HTTPS URL
	HTTPSUrl string

	// 是否在断开控制连接后，退出
	ExitWithDisconnect bool

	// 是否已经初始化
	initialized bool

	// 标记是否关闭链接, true:关闭， false:不关闭
	IsClose bool
	// 连接中其他goroutine的关闭信号
	closed chan bool

	// 控制的tcp连接
	conn net.Conn

	// 写缓冲通道
	writeChan chan []byte
}

// Init(domain string, port uint, user, password string) 初始化控制链接参数
func (conn *ControlConnection) Init(domain string, port uint, user, password string) {
	conn.ServerDomain = domain
	conn.ServerPort = port
	conn.User = user
	conn.Password = password

	conn.ExitWithDisconnect = false

	conn.initialized = true
}

// SetHTTPConfig() 设置HTTP代理的配置
func (conn *ControlConnection) SetHTTPConfig(hostname, subdomain, auth string, port uint) {
	conn.HTTPHostname = hostname
	conn.HTTPSubdomain = subdomain
	conn.HTTPAuth = auth
	conn.HTTPLocalPort = port
}

// SetHTTPSConfig() 设置HTTPS代理的配置
func (conn *ControlConnection) SetHTTPSConfig(hostname, subdomain, auth string, port uint) {
	conn.HTTPSHostname = hostname
	conn.HTTPSSubdomain = subdomain
	conn.HTTPSAuth = auth
	conn.HTTPSLocalPort = port
}

// Service() 开始连接，如果失败返回error，该函数阻塞
func (conn *ControlConnection) Service() error {
	if !conn.initialized {
		panic("Should Init first!")
	}

	err := conn.connect()

	if err != nil {
		return err
	}

	// 初始化写数据的缓冲通道
	conn.writeChan = make(chan []byte, 10)

	// 为发送数据建立单独的goroutine， 通过writeChan缓冲通道交给write函数发送数据
	go conn.write()

	auth := util.Auth{Version: "1.0.0", MmVersion: "1", User: conn.User, Password: conn.Password, OS: "!", Arch: "1", ClientId: ""}

	content, err := util.PayloadStructToBytes(auth, util.AUTH_TYPE)

	conn.writeChan <- content

	conn.readHandler()

	return err
}

// connect() 创建链接，并将net.Conn 赋值给对象的conn
func (connection *ControlConnection) connect() error {

	address := connection.ServerDomain + ":" + strconv.FormatUint(uint64(connection.ServerPort), 10)

	// 无视ssl证书，仅用于测试
	config := &tls.Config{InsecureSkipVerify: true}

	conn, err := tls.Dial("tcp", address, config)

	if err == nil {
		connection.conn = conn
	}

	return err
}

// write() 链接写函数，通过 writeChan 缓冲通道接收要发送给服务器的数据，再逐一发送
// 目前设计为执行在一个单独的goroutine中
func (conn *ControlConnection) write() {

	for conn.IsClose == false {

		select {
		case _, ok := <-conn.closed:
			if !ok {
				return
			}
		case buf, ok := <-conn.writeChan:
			if ok {
				var n = 0

				for n < len(buf) {

					n, err := conn.conn.Write(buf)

					if err != nil {
						// TODO: 错误处理
						fmt.Println("write():" + err.Error())

						conn.Close()
					}

					buf = buf[n:]
				}
			} else {
				return
			}
		}
	}

}

// readHandler() 从socket中读取数据，并解析，处理各个事件
func (conn *ControlConnection) readHandler() {

	var cmdBuffer *bytes.Buffer = nil
	var tempBuffer *bytes.Buffer = nil
	var cmdLen uint16 = 0

	for conn.IsClose == false {
		fmt.Println(config.CONFIG.ReadBufSize)
		buf := make([]byte, config.CONFIG.ReadBufSize)

		n, err := conn.conn.Read(buf)

		if err != nil {
			// TODO: 错误处理
			fmt.Println("readHandler():" + err.Error())

			conn.Close()
		}

		if n <= 0 {
			// TODO: 错误处理
			conn.Close()
		}

		if n > 8 && tempBuffer == nil {
			// 新的命令

			cmdLenBytes := buf[:8]
			// 获取到命令的长度
			cmdLen = util.ToLen(cmdLenBytes)

			if uint16(n-8) == cmdLen {
				// 接收到的数据长度刚好等于命令长度
				cmdBuffer = &bytes.Buffer{}
				cmdBuffer.Write(buf[8:n])

			} else if uint16(n-8) > cmdLen {
				// 接收到的数据长度大于命令长度，说明完整的命令后接着有其他命令
				cmdBuffer = &bytes.Buffer{}
				cmdBuffer.Write(buf[8 : cmdLen+8])

				tempBuffer = &bytes.Buffer{}
				tempBuffer.Write(buf[cmdLen+8 : n])

			} else {
				// 接收到的数据长度少于命令长度，说明该命令不完整，还需要继续获取
				tempBuffer = &bytes.Buffer{}
				tempBuffer.Write(buf[:n])
			}

			// 重置命令长度
			cmdLen = 0

		} else if tempBuffer != nil {
			// 未接收完的命令

			// 先将所有数据写入临时缓存
			tempBuffer.Write(buf[:n])

			tempBytes := tempBuffer.Bytes()

			cmdLenBytes := tempBytes[:8]
			// 获取到命令的长度
			cmdLen = util.ToLen(cmdLenBytes)

			// 缓存中数据的长度
			bufLen := tempBuffer.Len()

			if uint16(bufLen-8) == cmdLen {
				// 接收到的数据长度刚好等于命令长度
				cmdBuffer = &bytes.Buffer{}
				cmdBuffer.Write(tempBytes[8:bufLen])

				tempBuffer = nil

			} else if uint16(n-8) > cmdLen {
				// 接收到的数据长度大于命令长度，说明完整的命令后接着有其他命令
				cmdBuffer = &bytes.Buffer{}
				cmdBuffer.Write(tempBytes[8 : cmdLen+8])

				tempBuffer.Reset()
				tempBuffer.Write(tempBytes[cmdLen+8 : bufLen])

			}

			cmdLen = 0

		} else {
			// 长度少于8byte,且不是未接收完的数据，需要错误处理

			fmt.Println("readHandler(): buf less than 8 byte and not uncompleted data!")
			// 命令出错，关闭连接
			conn.Close()
		}

		if cmdBuffer != nil {
			// 接收到一条完整的命令

			fmt.Println(cmdBuffer.String())
			conn.dispatch(cmdBuffer.Bytes())
			cmdBuffer = nil
		}

	}

}

// dispatch() 解析命令，并将命令分配给各个函数处理
func (conn *ControlConnection) dispatch(cmdBytes []byte) {
	resp, respType, errno := util.ParsePayloadStruct(cmdBytes)

	if errno != errcode.ERR_SUCCESS {
		// 命令解析错误
		fmt.Printf("dispatch() ParsePayloadStruct errno: %d, err msg: %s", errno, errcode.GetErrMsg(errno))

		// TODO: 命令出错，是否该断开连接？
		// 目前先断开 control 连接处理
		conn.Close()
		return
	}

	var handlerErr int

	switch respType {
	case util.AUTH_RESP_TYPE:
		handlerErr = conn.authRespHandler(resp.(util.AuthResp))
	case util.NEW_TUNNEL_TYPE:
		handlerErr = conn.newTunnelHandler(resp.(util.NewTunnel))
	case util.REQ_PROXY_TYPE:
		handlerErr = conn.reqProxyHandler(resp.(util.ReqProxy))
	// case util.START_PROXY_TYPE:
	// 	handlerErr = conn.startProxyHandler(resp.(util.StartProxy))
	case util.PONG_TYPE:
		handlerErr = conn.pongHandler(resp.(util.Pong))
	default:
		// 未知命令，可能版本问题
		handlerErr = errcode.ERR_UNKNOW_RESP
	}

	if handlerErr != errcode.ERR_SUCCESS {
		// 错误处理

		conn.Close()
	}
}

// authRespHandler()处理AuthResp的响应函数
func (conn *ControlConnection) authRespHandler(resp util.AuthResp) int {

	// TODO: 以后更新时，可能要判断服务器版本，现在先忽略Version和MmVersion

	if resp.Error != "" || resp.ClientId == "" {
		// 返回的错误信息(Error)不为 "" 或者 服务端没有返回ClientId
		return errcode.ERR_AUTH_FAILED
	}

	conn.ClientId = resp.ClientId

	// HTTP 的 ReqTunnel 请求
	if conn.HTTPLocalPort > 0 {
		// 需要代理的http连接本地端口，当大于0时表示需要代理连接
		reqTunnel := util.ReqTunnel{ReqId: "", Protocol: util.PROTOCOL_HTTP, Hostname: conn.HTTPHostname, Subdomain: conn.HTTPSubdomain, HttpAuth: "", RemotePort: 0}

		byteData, err := util.PayloadStructToBytes(reqTunnel, util.REQ_TUNNEL_TYPE)

		if err != nil {
			// TODO: 错误处理
			fmt.Println("authRespHandler():" + err.Error())
			return errcode.ERR_PAYLOAD_TO_BYTES
		}

		// 将请求放入发送缓存队列
		conn.writeChan <- byteData
	}

	// HTTPS 的 ReqTunnel 请求
	if conn.HTTPSLocalPort > 0 {
		// 需要代理的https连接本地端口，当大于0时表示需要代理连接

		reqTunnel := util.ReqTunnel{ReqId: "", Protocol: util.PROTOCOL_HTTPS, Hostname: conn.HTTPSHostname, Subdomain: conn.HTTPSSubdomain, HttpAuth: "", RemotePort: 0}

		byteData, err := util.PayloadStructToBytes(reqTunnel, util.REQ_TUNNEL_TYPE)

		if err != nil {
			// TODO: 错误处理
			fmt.Println("authRespHandler():" + err.Error())
			return errcode.ERR_PAYLOAD_TO_BYTES
		}

		// 将请求放入发送缓存队列
		conn.writeChan <- byteData
	}

	return errcode.ERR_SUCCESS
}

// newTunnelHandler()处理NewTunnel的响应函数
func (conn *ControlConnection) newTunnelHandler(resp util.NewTunnel) int {

	if resp.Error != "" {
		// 返回信息中Error不为"""
		fmt.Println("resp.Error:" + resp.Error)
		return errcode.ERR_NEW_TUNNEL_ERROR
	}

	switch resp.Protocol {
	case util.PROTOCOL_HTTP:
		conn.HTTPUrl = resp.Url
	case util.PROTOCOL_HTTPS:
		conn.HTTPSUrl = resp.Url
	default:
		// 不支持的协议
	}
	return errcode.ERR_SUCCESS
}

// reqProxyHandller()处理ReqProxy的响应函数
func (conn *ControlConnection) reqProxyHandler(resp util.ReqProxy) int {

	address := conn.ServerDomain + ":" + strconv.FormatUint(uint64(conn.ServerPort), 10)

	proxyConn := ProxyConnection{}
	proxyConn.Init(conn.ClientId, address, conn)

	go proxyConn.Start()

	return errcode.ERR_SUCCESS
}

// pongHandler()处理Pong的响应函数
func (conn *ControlConnection) pongHandler(resp util.Pong) int {

	return errcode.ERR_SUCCESS
}

// Close() 关闭连接
func (conn *ControlConnection) Close() {

	if !conn.IsClose {

		// 关闭closed通道，使得其他goroutine能够知道要关闭连接
		close(conn.closed)

		conn.IsClose = true

		err := conn.conn.Close()

		fmt.Println("Close():" + err.Error())

		close(conn.writeChan)
	}

}
