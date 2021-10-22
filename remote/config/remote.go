package config

type Remote struct {
	BindHost      string            `yaml:"BindHost"`
	BindPort      uint16            `yaml:"BindPort"`
	UseSSL        bool              `yaml:"UseSSL"`
	SNI           string            `yaml:"SNI"`
	CertPath      string            `yaml:"CertPath"`
	KeyPath       string            `yaml:"KeyPath"`
	Salt          string            `yaml:"Salt"`
	Users         map[string]string `yaml:"Users"`
	ValidHttpPath string            `yaml:"HttpPath"`
}

// type User [2]string

// func NewRemote() *Remote {
// 	users := make(map[string]string)
// 	users["user1"] = "pwd1"
// 	users["user2"] = "pwd2"
// 	return &Remote{
// 		BindPort: 3789,
// 		Salt:     "salt",
// 		Users:    users,
// 	}
// }
