package config

const (
	SOCKS5 = "socks5"
	HTTP   = "http"
)

type Local struct {
	ProxyType     string     `yaml:"ProxyType"`
	BindHost      string     `yaml:"BindHost"`
	BindPort      uint16     `yaml:"BindPort"`
	RemoteHost    string     `yaml:"RemoteHost"`
	RemotePort    uint16     `yaml:"RemotePort"`
	RemoteSSL     bool       `yaml:"RemoteSSL"`
	GeoDomain     *GeoDomain `yaml:"GeoDomain"`
	Salt          string     `yaml:"Salt"`
	Username      string     `yaml:"Username"`
	Password      string     `yaml:"Password"`
	AllowInsecure bool       `yaml:"AllowInsecure"`

	HttpPath      string `yaml:"HttpPath"`
	HttpDomain    string `yaml:"HttpDomain"`
	HttpUserAgent string `yaml:"HttpUserAgent"`
}

type GeoDomain struct {
	DirectIfNotInRules bool   `yaml:"DirectIfNotInRules"`
	GfwPath            string `yaml:"GfwPath"`
	DirectPath         string `yaml:"DirectPath"`
}
