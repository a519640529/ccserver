package common

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"math/rand"
	"os"
	"regexp"

	"encoding/json"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	"math"
)

const (
	RankServerType = 15
)

var seed int64 = 1

func GetSelfSrvType() int {
	return netlib.Config.SrvInfo.Type
}

func GetSelfSrvId() int {
	return netlib.Config.SrvInfo.Id
}

func GetSelfAreaId() int {
	return netlib.Config.SrvInfo.AreaID
}

func GetAccountSrvId() int {
	return srvlib.ServerSessionMgrSington.GetServerId(GetSelfAreaId(), srvlib.AccountServerType)
}

func GetGameSrvId() int {
	return srvlib.ServerSessionMgrSington.GetServerId(GetSelfAreaId(), srvlib.GameServerType)
}
func GetGameSrvIds() []int {
	return srvlib.ServerSessionMgrSington.GetServerIds(GetSelfAreaId(), srvlib.GameServerType)
}
func GetWorldSrvId() int {
	return srvlib.ServerSessionMgrSington.GetServerId(GetSelfAreaId(), srvlib.WorldServerType)
}

func GetRankSrvId() int {
	return srvlib.ServerSessionMgrSington.GetServerId(GetSelfAreaId(), RankServerType)
}

func GetAppId() string {
	return Config.AppId
}

func GetClientSessionId(s *netlib.Session) srvlib.SessionId {
	param := s.GetAttribute(srvlib.SessionAttributeClientSession)
	if sid, ok := param.(srvlib.SessionId); ok {
		return sid
	}
	return srvlib.SessionId(0)
}

func GetRandInt(max int) int {
	seed++
	rand.Seed(seed)
	return rand.Intn(max)
}

func Md5String(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return hex.EncodeToString(h.Sum(nil))
}

func MakeMd5String(strs ...string) string {
	buff := md5.New()
	for _, value := range strs {
		io.WriteString(buff, value)
	}
	return hex.EncodeToString(buff.Sum(nil))
}

func SetIntegerBit(num int32, index int32) int32 {
	return num | (1 << uint(index-1))
}

func GetIntegerBit(num int32, index int32) bool {
	if num&(1<<uint(index-1)) > 0 {
		return true
	} else {
		return false
	}
}

// 校验身份证是否合法
var IDReg, _ = regexp.Compile(`(^[1-9]\d{5}(18|19|([23]\d))\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$)|(^[1-9]\d{5}\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{2}$)`)
var REGEXP_IPRule, _ = regexp.Compile(`^(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[1-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)$`)
var ClubNameRule, _ = regexp.Compile("^[\u4e00-\u9fa5a-zA-Z-z0-9]+$")

func IsValidID(id string) bool {
	if IDReg != nil {
		return IDReg.Match([]byte(id))
	}
	return false
}

func IsValidIP(Ip string) bool {
	const UNKNOWIP = "0.0.0.0"
	if Ip == "" || Ip == UNKNOWIP {
		return false
	}
	if !REGEXP_IPRule.MatchString(Ip) {
		return false
	}
	return true
}

func JsonToStr(v interface{}) string {
	buff, _ := json.Marshal(v)
	return string(buff)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func MinI64(l int64, r int64) int64 {
	if l < r {
		return l
	} else {
		return r
	}
}

func AbsI64(l int64) int64 {
	if l >= 0 {
		return l
	} else {
		return -l
	}
}

// 如果结果为负，说明是玩家亏钱，结果为正，玩家赢钱。
// 图像上是一个分段函数，绝对值极值为1
func GetWinLossRate(win int64, loss int64) float64 {
	ret := float64(0)

	if win > loss {
		ret = float64(loss) / float64(win)
	} else {
		ret = -float64(win) / float64(loss)
	}

	return ret
}

// 得到一个初段慢，高段变快数
func GetSoftMaxNum(cur float64, maxValue float64) float64 {
	if cur > maxValue {
		return maxValue
	}
	return cur * math.Sin((cur/maxValue)*math.Pi/2)
}
