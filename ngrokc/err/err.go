package err

const (
	ERR_SUCCESS = 0

	// 未知的响应
	ERR_UNKNOW_RESP = -100

	// 验证失败
	ERR_AUTH_FAILED = -101

	// ReqTunnel 请求失败, 返回的NewTunnel中含有错误信息
	ERR_NEW_TUNNEL_ERROR = -102

	// 不是客户端代理的URL
	ERR_UNKNOW_PROXY_URL = -103

	// 代理连接，连接本地端口失败
	ERR_CONNECT_LOCAL_FAILED = -104
)

func GetErrMsg(errno int) string {
	switch errno {
	case ERR_SUCCESS:
		return "Success"
	case ERR_UNKNOW_RESP:
		return "Unknow response"
	case ERR_AUTH_FAILED:
		return "Auth failed"
	case ERR_NEW_TUNNEL_ERROR:
		return "New tunnel request error"
	case ERR_UNKNOW_PROXY_URL:
		return "Unknow proxy url"
	case ERR_CONNECT_LOCAL_FAILED:
		return "Ngrok proxy failed to connect local service"
	default:
		return "Unknow error code, please check!!!"
	}
}
