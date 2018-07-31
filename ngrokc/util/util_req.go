package util

type Auth struct {
	Version   string // protocol version
	MmVersion string // major/minor software version (informational only)
	User      string
	Password  string
	OS        string
	Arch      string
	ClientId  string // empty for new sessions
}

type ReqTunnel struct {
	ReqId string
	Protocol string

	// http/https only
	Hostname string
	Subdomain string
	HttpAuth string

	// tcp only
	RemotePort uint16
}

type RegProxy struct {
	ClientId string
}

type Ping struct {
	
}