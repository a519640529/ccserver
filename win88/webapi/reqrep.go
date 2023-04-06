package webapi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	STATE_OK  = 1
	STATE_ERR = 0
)

const (
	RESPONSE_STATE     = "State"
	RESPONSE_ERRMSG    = "ErrMes"
	RESPONSE_PAGECOUNT = "PageCount"
	RESPONSE_PAGENO    = "PageNo"
	RESPONSE_TOTAL     = "Total"
	RESPONSE_DATA      = "Data"
)

type RequestBody map[string]interface{}

func NewRequestBody(data []byte) (RequestBody, error) {
	m := make(map[string]interface{})
	var err error
	if len(data) > 0 {
		err = json.Unmarshal(data, &m)
	}
	return RequestBody(m), err
}

func (rp RequestBody) GetStr(key string) (string, bool) {
	if val, ok := rp[key]; ok {
		if str, ok := val.(string); ok {
			return str, true
		}

		return fmt.Sprintf("%v", val), false
	}
	return "", false
}

func (rp RequestBody) GetInt(key string) (int, bool) {
	if val, ok := rp[key]; ok {
		if fval, ok := val.(float64); ok {
			return int(fval), true
		}

		if sval, ok := val.(string); ok {
			v, err := strconv.Atoi(sval)
			if err == nil {
				return v, true
			}
		}
	}
	return 0, false
}

func (rp RequestBody) GetInt64(key string) (int64, bool) {
	if val, ok := rp[key]; ok {
		if fval, ok := val.(float64); ok {
			return int64(fval), true
		}

		if sval, ok := val.(string); ok {
			v, err := strconv.ParseInt(sval, 10, 64)
			if err == nil {
				return v, true
			}
		}
	}
	return 0, false
}

func (rp RequestBody) GetFloat32(key string) (float32, bool) {
	if val, ok := rp[key]; ok {
		if fval, ok := val.(float64); ok {
			return float32(fval), true
		}

		if sval, ok := val.(string); ok {
			v, err := strconv.ParseInt(sval, 10, 32)
			if err == nil {
				return float32(v), true
			}
		}
	}
	return 0, false
}

func (rp RequestBody) GetFloat64(key string) (float64, bool) {
	if val, ok := rp[key]; ok {
		if fval, ok := val.(float64); ok {
			return fval, true
		}

		if sval, ok := val.(string); ok {
			v, err := strconv.ParseInt(sval, 10, 64)
			if err == nil {
				return float64(v), true
			}
		}
	}
	return 0, false
}

func (rp RequestBody) GetBool(key string) (bool, bool) {
	if val, ok := rp[key]; ok {
		if bval, ok := val.(bool); ok {
			return bval, true
		}
		if sval, ok := val.(string); ok {
			lows := strings.ToLower(sval)
			if strings.Compare(lows, "true") == 0 {
				return true, true
			} else if strings.Compare(lows, "false") == 0 {
				return false, true
			}
		}
	}
	return false, false
}

func (rp RequestBody) GetData(key string) (interface{}, bool) {
	if val, ok := rp[key]; ok {
		return val, true
	}
	return nil, false
}

func (rp RequestBody) GetRequestBody(key string) (RequestBody, bool) {
	if val, ok := rp[key]; ok {
		if bval, ok := val.(map[string]interface{}); ok {
			return RequestBody(bval), true
		}
	}
	return nil, false
}

type ResponseBody map[string]interface{}

func NewResponseBody() ResponseBody {
	m := make(map[string]interface{})
	return ResponseBody(m)
}

func (rb ResponseBody) Marshal() ([]byte, error) {
	return json.Marshal(rb)
}
