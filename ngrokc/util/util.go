package util

import (

	"bytes"
	"reflect"
	"encoding/json"
	"encoding/binary"

	"ngrok-client/ngrokc/err"
)

// 代理类型
const (
	PROTOCOL_HTTP = "http"
	PROTOCOL_HTTPS = "https"
)

// 请求
const (
	AUTH_TYPE = "Auth"
	REQ_TUNNEL_TYPE = "ReqTunnel"
	REG_PROXY_TYPE = "RegProxy"
	PING_TYPE = "Ping"
)

// 响应
const (
	AUTH_RESP_TYPE = "AuthResp"
	NEW_TUNNEL_TYPE = "NewTunnel"
	REQ_PROXY_TYPE = "ReqProxy"
	START_PROXY_TYPE = "StartProxy"
	PONG_TYPE = "Pong"
)

// ToLen 从二进制中读出数字
func ToLen(bytes []byte) uint16 {
	return binary.LittleEndian.Uint16(bytes)
}

// LenToBytes 将一个长度数字放入8个byte中
func LenToBytes(length uint16) []byte {
	content := make([]byte, 8)

	binary.LittleEndian.PutUint16(content, length)

	return content
}

// Payload 的结构体
type PayloadStruct struct {
	Payload interface{}
	Type string 
}

// ParseCommuStruct 解析接收到的数据
// 将会返回(resp, type, err)
// resp 是对应响应的结构体(struct), 返回类型是 interface{}
// type 是代表返回的是哪个类型的响应，返回类型是 string
// err 表示函数处理有没有错误，没有错误则为0, 有错误则返回对应错误代码
func ParsePayloadStruct(content []byte) (interface{}, string, int) {

	var resp PayloadStruct

	// 删除无用的空余数据，不然json解析会出错
	content = bytes.Trim(content, "\x00")

	errObj := json.Unmarshal(content, &resp)

	if errObj != nil {
		return nil, resp.Type, err.ERR_UNKNOW_RESP
	}

	payloadType := reflect.TypeOf(resp.Payload)
	
	if payloadType.Kind() == reflect.Map {
		switch resp.Type {
		case AUTH_RESP_TYPE:
			// AuthResp
			var payload AuthResp
			payload.ParseFromMap(resp.Payload.(map[string]interface{}))
			return payload, resp.Type, err.ERR_SUCCESS
		case NEW_TUNNEL_TYPE:
			// NewTunnel
			var payload NewTunnel
			payload.ParseFromMap(resp.Payload.(map[string]interface{}))
			return payload, resp.Type, err.ERR_SUCCESS
		case REQ_PROXY_TYPE:
			// ReqProxy
			var payload ReqProxy
			return payload, resp.Type, err.ERR_SUCCESS
		case START_PROXY_TYPE:
			// StartProxy
			var payload StartProxy
			payload.ParseFromMap(resp.Payload.(map[string]interface{}))
			return payload, resp.Type, err.ERR_SUCCESS
		case PONG_TYPE:
			// Pong
			var payload Pong
			return payload, resp.Type, err.ERR_SUCCESS
		default:
			return nil, resp.Type, err.ERR_UNKNOW_RESP
		}
	} else {
		return nil, resp.Type, err.ERR_UNKNOW_RESP
	}
}

// PayloadStructToBytes 将一个Payload结构转化为二进制，并且在前面8位byte中带上数据的长度
// 将会返回([]byte, err)
// []byte 是生产的数据，前8位byte后面数据的长度
// err 表示函数处理有没有错误，没有错误则为nil, 有错误则返回
func PayloadStructToBytes(payload interface{}, payloadType string) ([]byte, error) {
	var retVal []byte

	payStruct := PayloadStruct{Payload:payload, Type:payloadType}

	content, err := json.Marshal(payStruct)
	
	if err == nil {
		var length = len(content)
		var lenBytes = LenToBytes(uint16(length))
		
		var buf bytes.Buffer
		buf.Write(lenBytes)
		buf.Write(content)

		retVal = buf.Bytes()

		return retVal, nil
	} else {
		
		return nil, err
	}
}