package common

import (
	"encoding/json"
	"github.com/idealeak/goserver/core"
)

var Config = Configuration{}

type Configuration struct {
	AppId     string
	SrvId     string
	IsDevMode bool
}

func (this *Configuration) Name() string {
	return "common"
}

func (this *Configuration) Init() error {
	return nil
}

func (this *Configuration) Close() error {
	return nil
}

func init() {
	core.RegistePackage(&Config)
	core.RegistePackage(&CustomConfig)
}

var CustomConfig = make(CustomConfiguration)

type CustomConfiguration map[string]interface{}

func (this *CustomConfiguration) Name() string {
	return "costum"
}

func (this *CustomConfiguration) Init() error {
	return nil
}

func (this *CustomConfiguration) Close() error {
	return nil
}

func (this *CustomConfiguration) GetString(key string) string {
	if v, exist := (*this)[key]; exist {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}

func (this *CustomConfiguration) GetStrings(key string) (strs []string) {
	if v, exist := (*this)[key]; exist {
		if vals, ok := v.([]interface{}); ok {
			for _, s := range vals {
				if str, ok := s.(string); ok {
					strs = append(strs, str)
				}
			}
			return
		}
	}
	return
}

func (this *CustomConfiguration) GetCustomCfgs(key string) (strs []*CustomConfiguration) {
	if v, exist := (*this)[key]; exist {
		if vals, ok := v.([]interface{}); ok {
			for _, s := range vals {
				if data, ok := s.(map[string]interface{}); ok {
					var pkg *CustomConfiguration
					modelBuff, _ := json.Marshal(data)
					err := json.Unmarshal(modelBuff, &pkg)
					if err == nil {
						strs = append(strs, pkg)
					}
				}
			}
			return
		}
	}
	return
}

func (this *CustomConfiguration) GetInts(key string) (strs []int) {
	if v, exist := (*this)[key]; exist {
		if vals, ok := v.([]interface{}); ok {
			for _, s := range vals {
				if str, ok := s.(float64); ok {
					strs = append(strs, int(str))
				}
			}
			return
		}
	}
	return
}
func (this *CustomConfiguration) GetInt(key string) int {
	if v, exist := (*this)[key]; exist {
		if val, ok := v.(float64); ok {
			return int(val)
		}
	}
	return 0
}

func (this *CustomConfiguration) GetBool(key string) bool {
	if v, exist := (*this)[key]; exist {
		if val, ok := v.(bool); ok {
			return val
		}
	}
	return false
}
