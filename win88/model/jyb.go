package model

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

// 礼包码
var (
	chars = []rune{
		'4', '1', '7', '2', '3', '6', '5', '8', '9',
		'q', 'n', 'c', 'r', 't', 'm', 'a', 'd', 'e', 'h', 'j', 'l', 'f', 'b', 'g',
		'Q', 'N', 'C', 'R', 'T', 'M', 'A', 'D', 'E', 'H', 'J', 'L', 'F', 'B', 'G',
	}
	divider            = 'i' // 分割标识(区分补位，应该是chars里面没出现的字符)
	charLen            = uint64(len(chars))
	charIndexMap       = make(map[rune]int, len(chars)) //  初始化就完成写入 不存在并发读写问题
	Keystart     int64 = 100000                         // 单兑换码上限十万

)

var (
	ErrJYBCode       = errors.New("CodeExist")
	ErrJYBRpc        = errors.New("Rpcisnil")
	ErrJYBPlCode     = errors.New("PlayCodeExist")    // 已经兑换过该礼包
	ErrJybISReceive  = errors.New("jyb is receive")   // 该兑换码已被使用
	ErrJybTsTimeErr  = errors.New("jyb ts is not")    // 该兑换码已过期
	ErrJybIsNotExist = errors.New("jyb is not exist") // 请输入正确的兑换码
)

func init() {
	for i, char := range chars {
		charIndexMap[char] = i
	}
}

type JybInfoAward struct {
	Item    []*Item // 道具
	Coin    int64   // 金币
	Diamond int64   // 钻石
}

type JybInfos struct {
	Jybs []*JybInfo
}

type JybInfo struct {
	JybId     bson.ObjectId    `bson:"_id"` // 礼包ID
	Platform  string           //平台
	Name      string           // 礼包名称
	CodeType  int32            // 礼包类型 1 通用 2 特殊
	StartTime int64            // 开始时间 Unix
	EndTime   int64            // 结束时间
	Content   string           // 礼包内容
	Max       int32            // 总个数
	Receive   int32            // 领取个数
	CodeStart int64            // Code生成起始
	Code      map[string]int32 // 礼包码
	Award     *JybInfoAward    // 礼包内东西
	CreateTs  int64
}

// JybKey 自增
type JybKey struct {
	KeyId    bson.ObjectId `bson:"_id"`
	Keyint   int64
	Plakeyid string            // keystart+Platform
	GeCode   map[string]string // 通用礼包
}

// 验证玩家是否领取礼包  可能存在map并发问题
type JybUserInfo struct {
	JybUserId bson.ObjectId    `bson:"_id"`
	SnId      int32            // 玩家id
	Platform  string           //平台
	JybInfos  map[string]int32 //已领取礼包 web和world使用 注意线程安全
}

type GetJybInfoArgs struct {
	Id       string //  礼包ID
	Plt      string //  平台
	UseCode  string //  礼包码
	SnId     int32  // 玩家id
	CodeType int32
}

type InitJybInfoArgs struct {
	Plt string //  平台
}

type VerifyUpJybInfoArgs struct {
	UseCode   string //  礼包码
	CodeStart int64  // Code生成起始
	Plt       string // 平台
	SnId      int32  // 玩家id
	CodeType  int32
}

type CreateJyb struct {
	*JybInfo
	Codelen int32
	Num     uint64
}

// NewJybUserInfo 初始化
func NewJybKey(plt string) *JybKey {
	jk := &JybKey{KeyId: bson.NewObjectId(), Keyint: Keystart, Plakeyid: fmt.Sprintf("%d_%s", Keystart, plt)}
	return jk
}

// sync.Map  new(sync.Map)
// NewJybUserInfo 初始化
func NewJybUserInfo(sid int32, plt string) *JybUserInfo {
	jUserInfo := &JybUserInfo{JybUserId: bson.NewObjectId(), SnId: sid, Platform: plt, JybInfos: make(map[string]int32)}
	return jUserInfo
}

func NewJybInfo(plt, name, content, code string, startTs, endTs int64, max, codetype int32, award *JybInfoAward) *JybInfo {
	jyb := &JybInfo{
		JybId:     bson.NewObjectId(),
		Platform:  plt,
		Name:      name,
		Content:   content,
		CodeType:  codetype,
		StartTime: startTs,
		EndTime:   endTs,
		Max:       max,
		Receive:   0,
		Code:      make(map[string]int32),
		Award:     award,
		CreateTs:  time.Now().Unix(),
	}
	if codetype == 1 {
		jyb.Code[code] = 1
	}
	return jyb
}

func NewJybCode(jyb *JybInfo, codelen int32, num uint64) {
	cl := int(codelen)
	num += uint64(jyb.CodeStart)
	for i := uint64(jyb.CodeStart); i != num; i++ {
		key := Id2Code(i, cl, true)
		jyb.Code[key] = 1
	}
}

func CreateJybInfo(args *CreateJyb) error {
	if rpcCli == nil {
		return ErrJYBRpc
	}
	ret := &JybInfo{}
	err := rpcCli.CallWithTimeout("JybSvc.CreateJybItem", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("CreateJybInfo err:%v Id:%v ", err, args.JybId)
		return err
	}
	//args.JybInfo = ret // 获取code
	return nil
}

func UpgradeJybUser(args *GetJybInfoArgs) *JybInfo {

	if rpcCli == nil {
		return nil
	}

	ret := &JybInfo{}
	err := rpcCli.CallWithTimeout("JybUserSvc.UpgradeJybUser", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("UpgradeJybUser err:%v Id:%v ", err, args.UseCode)
		return nil
	}

	return ret
}

func VerifyUpJybInfo(args *VerifyUpJybInfoArgs) (*JybInfo, error) {

	if rpcCli == nil {
		return nil, ErrJYBRpc
	}

	ret := &JybInfo{}
	err := rpcCli.CallWithTimeout("JybUserSvc.VerifyUpJybUser", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Errorf("VerifyUpJybUser err:%v Id:%v ", err, args.UseCode)
		return nil, err
	}

	return ret, err
}

func DelJybInfo(args *GetJybInfoArgs) error {

	if rpcCli == nil {
		return nil
	}

	ret := false
	err := rpcCli.CallWithTimeout("JybSvc.DelJyb", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("DelJybInfo err:%v Id:%v ", err, args.Id)
	}

	return err
}

func InitJybItem(args *InitJybInfoArgs) *JybInfos {

	if rpcCli == nil {
		return nil
	}

	ret := &JybInfos{}
	err := rpcCli.CallWithTimeout("JybSvc.InitJybItem", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("InitJybItem err:%v Id:%v ", err, args.Plt)
		return nil
	}

	return ret
}

func reverseSlice(strSlice []rune) []rune {
	sLen := len(strSlice)
	for i := 0; i < sLen/2; i++ {
		strSlice[i], strSlice[sLen-i-1] = strSlice[sLen-i-1], strSlice[i]
	}
	return strSlice
}

// Id2Code 把id转为兑换码
func Id2Code(id uint64, codeMixLength int, isRandomFix bool) string {
	code := make([]rune, 0, codeMixLength)
	for id/charLen > 0 {
		code = append(code, chars[id%charLen])
		id /= charLen
	}

	code = append(code, chars[id%charLen]) // 处理未除尽的余数
	// slice.ReverseSlice(code)
	reverseSlice(code)
	fixLen := codeMixLength - len(code) // 需要补码的长度
	if fixLen > 0 {
		rand.Seed(time.Now().UnixNano())
		code = append(code, divider)
		for i := 0; i < fixLen-1; i++ {
			// 每次固定，如果需要变的话，后面补码的内容可以改变
			if isRandomFix {
				code = append(code, chars[rand.Intn(int(charLen))])
			} else {
				code = append(code, chars[i])
			}
		}
	}
	return string(code)
}

func Code2Id(code string) (uint64, error) {
	if len(code) == 0 {
		return 0, nil
	}
	var id uint64 = 0
	codeRuneList := []rune(code)
	for i := range codeRuneList {
		// 如果是补码标志直接退出
		if codeRuneList[i] == divider {
			break
		}
		charIndex, ok := charIndexMap[codeRuneList[i]]
		if !ok {
			return 0, errors.New("code有误，解码失败")
		}
		if i > 0 {
			id = id*charLen + uint64(charIndex)
		} else {
			id = uint64(charIndex)
		}
	}
	return id, nil
}
