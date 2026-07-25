package config

type DNSConfig struct {
	Servers       []string         `json:"servers"`
	QueryStrategy DNSQueryStrategy `json:"query-strategy"`
	Mode          DNSMode          `json:"mode"`
}

type DNSQueryStrategy string

const (
	DNSUseIP     DNSQueryStrategy = "UseIP"
	DNSUseIPv4   DNSQueryStrategy = "UseIPv4"
	DNSUseIPv6   DNSQueryStrategy = "UseIPv6"
	DNSUseSystem DNSQueryStrategy = "UseSystem"
)

type DNSMode string

const (
	DNSModeSystem DNSMode = "system"
	DNSModeProxy  DNSMode = "proxy"
	DNSModeDirect DNSMode = "direct"
)

func DefaultDNSConfig() DNSConfig {
	return DNSConfig{
		Servers:       []string{"localhost"},
		QueryStrategy: DNSUseSystem,
		Mode:          DNSModeSystem,
	}
}
