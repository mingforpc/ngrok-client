package util

type AuthResp struct {
	Version   string
	MmVersion string
	ClientId  string
	Error     string
}

func (resp *AuthResp) ParseFromMap(content map[string]interface{}) {
	if content != nil {
		
		resp.Version, _ = content["Version"].(string)
		resp.MmVersion, _ = content["MmVersion"].(string)
		resp.ClientId, _ = content["ClientId"].(string)
		resp.Error, _ = content["Error"].(string)
	}
}

type NewTunnel struct {
	ReqId string
	Url string
	Protocol string
	Error string
}

func (resp *NewTunnel) ParseFromMap(content map[string]interface{}) {
	if content != nil {
		
		resp.ReqId, _ = content["ReqId"].(string)
		resp.Url, _ = content["Url"].(string)
		resp.Protocol, _ = content["Protocol"].(string)
		resp.Error, _ = content["Error"].(string)
	}
}

type ReqProxy struct {

}

type StartProxy struct {
	Url string // 将要访问的URL
	ClientAddr string
}

func (resp *StartProxy) ParseFromMap(content map[string]interface{}) {
	if content != nil {
		
		resp.Url, _ = content["Url"].(string)
		resp.ClientAddr, _ = content["ClientAddr"].(string)
	}
}

type Pong struct {

}