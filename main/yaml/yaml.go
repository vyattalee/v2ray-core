package yaml

import (
	"io"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	"github.com/v2fly/v2ray-core/v4/infra/conf"
	"github.com/v2fly/v2ray-core/v4/infra/conf/serial"
	"github.com/v2fly/v2ray-core/v4/main/confloader"
)

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      "YAML",
		Extension: []string{"yaml"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				cf := &conf.Config{}
				for i, arg := range v {
					newError("Reading config: ", arg).AtInfo().WriteToLog()
					r, err := confloader.LoadConfig(arg)
					common.Must(err)
					//c, err := serial.DecodeJSONConfig(r)
					c, err := serial.DecodeYAMLConfig(r)
					common.Must(err)
					if i == 0 {
						// This ensure even if the muti-yaml parser do not support a setting,
						// It is still respected automatically for the first configure file
						*cf = *c
						continue
					}
					cf.Override(c, arg)
				}
				return cf.Build()
			case io.Reader:
				//return serial.LoadJSONConfig(v)
				return serial.LoadYAMLConfig(v)
			default:
				return nil, newError("unknow type")
			}
		},
	}))
}
