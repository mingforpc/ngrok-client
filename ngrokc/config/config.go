package config

type Configuration struct {
	ServerHostname string `json:"server_hostname"`
	ServerPort uint `json:"server_port"`
	User string `json:"user"`
	Password string `json:"password"`

	HttpHostname string `json:"http_hostname"`
	HttpSubdomain string `json:"http_subdomain"`
	HttpAuth string `json:"http_auth"`
	HttpLocalPort uint `json:"http_local_port"`

	HttpsHostname string `json:"https_hostname"`
	HttpsSubdomain string `json:"https_subdomain"`
	HttpsAuth string `json:"https_auth"`
	HttpsLocalPort uint `json:"https_local_port"`

	ReadBufSize uint `json:"read_buf_size"`
}

var CONFIG *Configuration = &Configuration{}