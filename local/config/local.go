package config

type Local struct {
	BindHost      string `yaml:"BindHost"`
	BindPort      uint16 `yaml:"BindPort"`
	RemoteHost    string `yaml:"RemoteHost"`
	RemotePort    uint16 `yaml:"RemotePort"`
	RemoteSSL     bool   `yaml:"RemoteSSL"`
	Salt          string `yaml:"Salt"`
	Username      string `yaml:"Username"`
	Password      string `yaml:"Password"`
	AllowInsecure bool   `yaml:"AllowInsecure"`

	HttpPath      string `yaml:"HttpPath"`
	HttpDomain    string `yaml:"HttpDomain"`
	HttpUserAgent string `yaml:"HttpUserAgent"`
}
