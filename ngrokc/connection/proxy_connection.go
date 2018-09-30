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

type ProxyConnection struct {
	ClientId string

	Url        string
	ClientAddr string

	// 远程地址(ip:端口号 / 域名:端口号)
	RemoteAddress string

	// 标记是否关闭链接, true:关闭， false:不关闭
	IsClose bool
	// 链接中其他goroutine的关闭信号
	closed chan bool

	// proxy 连向服务端的连接
	proxyConn net.Conn

	// local 连向本地的连接
	localConn net.Conn

	// 本地服务写缓冲通道
	localWriteChan chan []byte

	// 服务端写缓冲通道
	remoteWriteChan chan []byte

	// 是否已经接收到 StartProxy 正式开始代理
	isStart bool

	// 指向控制链接的指针
	controlConn *ControlConnection
}

// Init(clientId, remoteAddress string, controlConn *ControlConnection) 初始化连接，只是初始化参数，并没有真正连接，
// 需要在Start()之前调用
func (conn *ProxyConnection) Init(clientId, remoteAddress string, controlConn *ControlConnection) {
	conn.ClientId = clientId
	conn.RemoteAddress = remoteAddress

	conn.controlConn = controlConn

	conn.closed = make(chan bool)
}

// connectServ() 连接服务端
func (conn *ProxyConnection) connectServ() error {

	// 无视ssl证书，仅用于测试
	config := &tls.Config{InsecureSkipVerify: true}

	connection, err := tls.Dial("tcp", conn.RemoteAddress, config)

	if err == nil {
		conn.proxyConn = connection
	}

	return err
}

// connectLocal() 连接本地服务
func (conn *ProxyConnection) connectLocal(isSSL bool, port uint) error {

	var connection net.Conn
	var err error

	var address = "127.0.0.1:" + strconv.FormatUint(uint64(port), 10)
	if isSSL {
		// SSL 连接

		// 无视ssl证书，仅用于测试
		config := &tls.Config{InsecureSkipVerify: true}

		connection, err = tls.Dial("tcp", address, config)
	} else {
		// 普通连接
		connection, err = net.Dial("tcp", address)
	}

	if err != nil {
		// TODO: 错误处理
		fmt.Println("connectLocal():" + err.Error())
	} else {
		conn.localConn = connection
	}

	return err
}

// Start() 开始服务
func (conn *ProxyConnection) Start() {

	// 连接服务器失败
	err := conn.connectServ()

	if err != nil {
		fmt.Printf("Failed to connect to server: %s", err)
		conn.Close()
		return
	}

	conn.remoteWriteChan = make(chan []byte, 10)
	conn.localWriteChan = make(chan []byte, 10)

	// 为发送给服务器数据建立单独的goroutine, 通过 remoteWriteChan 缓冲通道交给writeRemote()函数发送数据
	go conn.writeRemote()

	go conn.readRemote()

	// 发送 RegProxy 请求
	regProxy := util.RegProxy{ClientId: conn.ClientId}

	content, err := util.PayloadStructToBytes(regProxy, util.REG_PROXY_TYPE)

	if err != nil {
		// 组装Payload错误
		fmt.Printf("PayloadStructToBytes() Failed in Start(): %s", err)
		conn.Close()
		return
	}

	conn.remoteWriteChan <- content

}

// writeLocal() 链接写函数，通过 localWriteChan 缓冲通道接收要发送给本地的数据，再逐一发送
// 目前设计为执行在一个单独的goroutine中
func (conn *ProxyConnection) writeLocal() {

	for conn.IsClose == false {

		select {
		case _, ok := <-conn.closed:
			if !ok {
				return
			}
		case buf, ok := <-conn.localWriteChan:
			if ok {
				fmt.Printf("writeLocal: %s", buf)
				var n = 0

				for n < len(buf) {

					n, err := conn.localConn.Write(buf)

					if err != nil {
						// TODO: 错误处理
						fmt.Println("writeLocal():" + err.Error())
						conn.Close()
						return
					}

					buf = buf[n:]
				}
			} else {
				return
			}
		}

	}

}

// writeRemote() 链接写函数，通过 remoteWriteChan 缓冲通道接收要发送给服务器的数据，再逐一发送
// 目前设计为执行在一个单独的goroutine中
func (conn *ProxyConnection) writeRemote() {

	for conn.IsClose == false {

		select {
		case _, ok := <-conn.closed:
			if !ok {
				return
			}
		case buf, ok := <-conn.remoteWriteChan:
			if ok {
				fmt.Printf("writeRemote: %s", buf)
				var n = 0

				for n < len(buf) {

					n, err := conn.proxyConn.Write(buf)

					if err != nil {
						// TODO: 错误处理
						fmt.Println("writeRemote():" + err.Error())
						conn.Close()
						return
					}

					buf = buf[n:]
				}
			} else {
				return
			}

		}

	}

}

// readLocal() 从本地服务读取数据
func (conn *ProxyConnection) readLocal() {

	buf := make([]byte, config.CONFIG.ReadBufSize)

	for conn.IsClose == false {

		n, err := conn.localConn.Read(buf)
		fmt.Printf("readLocal: %s", buf[0:n])
		if err != nil {
			// TODO: 错误处理
			fmt.Println("readLocal():" + err.Error())
			conn.Close()
			return
		}

		select {
		case _, ok := <-conn.closed:
			if !ok {
				return
			}
		case conn.remoteWriteChan <- buf[0:n]:
			continue
		}

	}
}

// readRemote() 从服务端读取数据
func (conn *ProxyConnection) readRemote() {

	var cmdBuffer *bytes.Buffer = nil
	var tempBuffer *bytes.Buffer = nil

	// 存储proxy数据的buffer
	var dataBuffer *bytes.Buffer = nil

	var cmdLen uint16 = 0

	for conn.IsClose == false {

		buf := make([]byte, config.CONFIG.ReadBufSize)

		select {
		case _, ok := <-conn.closed:
			if !ok {
				return
			}
		default:

			n, err := conn.proxyConn.Read(buf)

			if err != nil {
				// TODO: 错误处理
				conn.Close()
				fmt.Println("readRemote():" + err.Error())
				return
			}

			if n <= 0 {
				// TODO: 错误处理
				conn.Close()
				fmt.Println("conn.proxyConn.Read() get 0 bytes:" + err.Error())
				return
			}

			if !conn.isStart {
				// 还未接收 StartProxy 命令

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
						// 接收到的数据长度大于命令长度，说明完整的命令后接着可能是代理的数据了
						cmdBuffer = &bytes.Buffer{}
						cmdBuffer.Write(buf[8 : cmdLen+8])

						dataBuffer = &bytes.Buffer{}
						dataBuffer.Write(buf[cmdLen+8 : n])

					} else {
						// 接收到的数据长度少于命令长度，说明该命令不完整，还需要继续获取
						tempBuffer = &bytes.Buffer{}
						tempBuffer.Write(buf[:n])
					}

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

						dataBuffer = &bytes.Buffer{}
						dataBuffer.Write(tempBytes[cmdLen+8 : bufLen])

					}

					cmdLen = 0

				} else {
					// 长度少于8byte,且不是未接收完的数据，需要错误处理
					conn.Close()
				}

				if cmdBuffer != nil {
					// 接收到一条完整的命令

					fmt.Println(cmdBuffer.String())
					conn.dispatch(cmdBuffer.Bytes())
					cmdBuffer = nil
				}

				if dataBuffer != nil {
					fmt.Println(dataBuffer.String())
					conn.localWriteChan <- dataBuffer.Bytes()
					dataBuffer = nil

				}

			} else {
				// 已经接收 StartProxy 命令

				if dataBuffer == nil {
					dataBuffer = &bytes.Buffer{}
				} else {

					if dataBuffer.Len() > 0 {

						conn.localWriteChan <- dataBuffer.Bytes()
					}

					dataBuffer.Reset()
				}

				dataBuffer.Write(buf[0:n])

				// 读写数据，传入本地连接
				conn.localWriteChan <- dataBuffer.Bytes()
			}

		}

	}
}

// dispatch() 解析命令，并将命令分配给各个函数处理
func (conn *ProxyConnection) dispatch(cmdBytes []byte) {
	resp, respType, errno := util.ParsePayloadStruct(cmdBytes)

	if errno != errcode.ERR_SUCCESS {
		// 命令解析错误

		fmt.Printf("util.ParsePayloadStruct() err: %s", errcode.GetErrMsg(errno))

		// 关闭连接
		conn.Close()
		return
	}

	var handlerErr int

	switch respType {
	case util.START_PROXY_TYPE:
		handlerErr = conn.startProxyHandler(resp.(util.StartProxy))
	default:
		// 未知命令，可能版本问题
		handlerErr = errcode.ERR_UNKNOW_RESP
	}

	if handlerErr != errcode.ERR_SUCCESS {
		// 错误处理
		// 关闭连接
		conn.Close()
	}
}

// startProxyHandler() 处理 StartProxy 请求
func (conn *ProxyConnection) startProxyHandler(resp util.StartProxy) int {

	var errnum = errcode.ERR_SUCCESS

	if resp.Url == conn.controlConn.HTTPUrl {
		// 代理HTTP

		err := conn.connectLocal(false, conn.controlConn.HTTPLocalPort)

		if err == nil {
			conn.isStart = true

			go conn.writeLocal()
			go conn.readLocal()
		} else {
			// 连接本地端口失败
			fmt.Println("startProxyHandler() failed to connect local HTTP service:" + err.Error())
			conn.Close()
			errnum = errcode.ERR_CONNECT_LOCAL_FAILED
		}

	} else if resp.Url == conn.controlConn.HTTPSUrl {
		// 代理HTTPS
		err := conn.connectLocal(true, conn.controlConn.HTTPSLocalPort)

		if err == nil {
			conn.isStart = true

			go conn.writeLocal()
			go conn.readLocal()
		} else {
			fmt.Println("startProxyHandler()failed to connect local HTTPS service:" + err.Error())
			conn.Close()
			errnum = errcode.ERR_CONNECT_LOCAL_FAILED
		}

	} else {
		errnum = errcode.ERR_UNKNOW_PROXY_URL
	}

	return errnum
}

// Close()关闭代理连接的方法
func (conn *ProxyConnection) Close() {

	if !conn.IsClose {

		// 关闭closed通道，使得其他goroutine能够知道要关闭连接
		close(conn.closed)

		if conn.localConn != nil {
			conn.localConn.Close()
		}

		if conn.proxyConn != nil {
			conn.proxyConn.Close()
		}

		conn.IsClose = true

		close(conn.localWriteChan)
		close(conn.remoteWriteChan)
	}
}
