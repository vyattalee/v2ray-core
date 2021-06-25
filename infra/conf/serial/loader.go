package serial

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common/errors"
	"github.com/v2fly/v2ray-core/v4/infra/conf"
	json_reader "github.com/v2fly/v2ray-core/v4/infra/conf/json"
	yaml_reader "github.com/v2fly/v2ray-core/v4/infra/conf/yaml"
)

type offset struct {
	line int
	char int
}

func findOffset(b []byte, o int) *offset {
	if o >= len(b) || o < 0 {
		return nil
	}

	line := 1
	char := 0
	for i, x := range b {
		if i == o {
			break
		}
		if x == '\n' {
			line++
			char = 0
		} else {
			char++
		}
	}

	return &offset{line: line, char: char}
}

// DecodeJSONConfig reads from reader and decode the config into *conf.Config
// syntax error could be detected.
func DecodeJSONConfig(reader io.Reader) (*conf.Config, error) {
	jsonConfig := &conf.Config{}

	jsonContent := bytes.NewBuffer(make([]byte, 0, 10240))
	jsonReader := io.TeeReader(&json_reader.Reader{
		Reader: reader,
	}, jsonContent)
	decoder := json.NewDecoder(jsonReader)

	if err := decoder.Decode(jsonConfig); err != nil {
		var pos *offset
		cause := errors.Cause(err)
		switch tErr := cause.(type) {
		case *json.SyntaxError:
			pos = findOffset(jsonContent.Bytes(), int(tErr.Offset))
		case *json.UnmarshalTypeError:
			pos = findOffset(jsonContent.Bytes(), int(tErr.Offset))
		}
		if pos != nil {
			return nil, newError("failed to read config file at line ", pos.line, " char ", pos.char).Base(err)
		}
		return nil, newError("failed to read config file").Base(err)
	}

	return jsonConfig, nil
}

func LoadJSONConfig(reader io.Reader) (*core.Config, error) {
	jsonConfig, err := DecodeJSONConfig(reader)
	if err != nil {
		return nil, err
	}

	pbConfig, err := jsonConfig.Build()
	if err != nil {
		return nil, newError("failed to parse json config").Base(err)
	}

	return pbConfig, nil
}

// DecodeJSONConfig reads from reader and decode the config into *conf.Config
// syntax error could be detected.
func DecodeYAMLConfig(reader io.Reader) (*conf.Config, error) {
	yamlConfig := &conf.Config{}

	yamlContent := bytes.NewBuffer(make([]byte, 0, 10240))
	yamlReader := io.TeeReader(&yaml_reader.Reader{
		Reader: reader,
	}, yamlContent)
	decoder := yaml.NewDecoder(yamlReader)
	//decoder.SetStrict(true)

	if err := decoder.Decode(yamlConfig); err != nil {
		var pos *offset
		//cause := errors.Cause(err)
		//switch tErr := cause.(type) {
		////case *yaml.SyntaxError:
		////	pos = findOffset(yamlContent.Bytes(), int(tErr.Offset))
		//case *yaml.TypeError:
		//	pos = findOffset(yamlContent.Bytes(), int(0))  //tErr.Offset
		//}
		if pos != nil {
			return nil, newError("failed to read config file at line ", pos.line, " char ", pos.char).Base(err)
		}
		return nil, newError("failed to read config file").Base(err)
	}

	return yamlConfig, nil
}

func LoadYAMLConfig(reader io.Reader) (*core.Config, error) {
	yamlConfig, err := DecodeYAMLConfig(reader)
	if err != nil {
		return nil, err
	}

	pbConfig, err := yamlConfig.Build()
	if err != nil {
		return nil, newError("failed to parse json config").Base(err)
	}

	return pbConfig, nil
}
