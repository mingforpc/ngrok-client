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
)