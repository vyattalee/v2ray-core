package conf

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v4/proxy/socks"
)

type SocksAccount struct {
	Username string `json:"user" yaml:"user"`
	Password string `json:"pass" yaml:"pass"`
}

func (v *SocksAccount) Build() *socks.Account {
	return &socks.Account{
		Username: v.Username,
		Password: v.Password,
	}
}

const (
	AuthMethodNoAuth   = "noauth"
	AuthMethodUserPass = "password"
)

type SocksServerConfig struct {
	AuthMethod string             `json:"auth" yaml:"auth"`
	Accounts   []*SocksAccount    `json:"accounts" yaml:"accounts"`
	UDP        bool               `json:"udp" yaml:"udp"`
	Host       *cfgcommon.Address `json:"ip" yaml:"ip"`
	Timeout    uint32             `json:"timeout" yaml:"timeout"`
	UserLevel  uint32             `json:"userLevel" yaml:"userLevel"`
}

func (v *SocksServerConfig) Build() (proto.Message, error) {
	config := new(socks.ServerConfig)
	switch v.AuthMethod {
	case AuthMethodNoAuth:
		config.AuthType = socks.AuthType_NO_AUTH
	case AuthMethodUserPass:
		config.AuthType = socks.AuthType_PASSWORD
	default:
		// newError("unknown socks auth method: ", v.AuthMethod, ". Default to noauth.").AtWarning().WriteToLog()
		config.AuthType = socks.AuthType_NO_AUTH
	}

	if len(v.Accounts) > 0 {
		config.Accounts = make(map[string]string, len(v.Accounts))
		for _, account := range v.Accounts {
			config.Accounts[account.Username] = account.Password
		}
	}

	config.UdpEnabled = v.UDP
	if v.Host != nil {
		config.Address = v.Host.Build()
	}

	config.Timeout = v.Timeout
	config.UserLevel = v.UserLevel
	return config, nil
}

type SocksRemoteConfig struct {
	Address *cfgcommon.Address `json:"address" yaml:"address"`
	Port    uint16             `json:"port" yaml:"port"`
	Users   []json.RawMessage  `json:"users" yaml:"users"`
}

type SocksClientConfig struct {
	Servers []*SocksRemoteConfig `json:"servers" yaml:"servers"`
}

func (v *SocksClientConfig) Build() (proto.Message, error) {
	config := new(socks.ClientConfig)
	config.Server = make([]*protocol.ServerEndpoint, len(v.Servers))
	for idx, serverConfig := range v.Servers {
		server := &protocol.ServerEndpoint{
			Address: serverConfig.Address.Build(),
			Port:    uint32(serverConfig.Port),
		}
		for _, rawUser := range serverConfig.Users {
			user := new(protocol.User)
			if err := json.Unmarshal(rawUser, user); err != nil {
				return nil, newError("failed to parse Socks user").Base(err).AtError()
			}
			account := new(SocksAccount)
			if err := json.Unmarshal(rawUser, account); err != nil {
				return nil, newError("failed to parse socks account").Base(err).AtError()
			}
			user.Account = serial.ToTypedMessage(account.Build())
			server.User = append(server.User, user)
		}
		config.Server[idx] = server
	}
	return config, nil
}
