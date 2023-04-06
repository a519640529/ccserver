package model

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"games.yol.com/win88/common"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"io"
	"strconv"
	"time"
)

type Account struct {
	AccountId    bson.ObjectId `bson:"_id"`
	SnId         int32         //玩家账号直接在这里生成
	UserName     string        //Service Provider AccountId
	PassWord     string        //昵称密码
	TelPassWord  string        //帐号密码
	BackPassWord string        //备份账号密码
	//VIP            int32         //VIP帐号 等级
	Platform       string    //平台
	Channel        string    //渠道
	Promoter       string    //推广员
	SubPromoter    string    //子推广员
	InviterId      int32     //邀请人ID
	PromoterTree   int32     //推广树ID
	Tel            string    //电话号码
	Params         string    //其他参数
	DeviceOs       string    //系统
	PackegeTag     string    //包标识
	Package        string    //包名 android:包名 ios:bundleid
	DeviceInfo     string    //设备信息
	LoginTimes     int       //登录次数
	State          int64     //冻结到期时间戳
	RegisteTime    time.Time //注册时间
	LastLoginTime  time.Time //最后一次登录时间
	LastLogoutTime time.Time //最后一次登出时间
	Flag           int32     //二次推广用户标记
	Remark         string    //备注信息
	TagKey         int32     //包标识关键字
}

func NewAccount() *Account {
	account := &Account{AccountId: bson.NewObjectId()}
	return account
}

type AccIsExistArg struct {
	AccId       string
	Username    string
	Password    string
	Tel         string
	Platform    string
	Ts          int64
	LoginType   int32
	TagKey      int32
	VerifyToken bool
}
type AccRet struct {
	Acc *Account
	Tag int
}

func AccountCheck(userName, tel, passWord, platform string, ts int64, logintype int32, tagkey int32, verifyToken bool) (*Account, int) {
	if rpcCli == nil {
		return nil, 0
	}

	args := &AccIsExistArg{
		Username:    userName,
		Password:    passWord,
		Tel:         tel,
		Platform:    platform,
		Ts:          ts,
		LoginType:   logintype,
		TagKey:      tagkey,
		VerifyToken: verifyToken,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.AccountCheck", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("error:", err)
		return nil, 0
	}
	return ret.Acc, ret.Tag
}
func AccountIsExist(userName, tel, passWord, platform string, ts int64, logintype int32, tagkey int32, verifyToken bool) (*Account, int) {
	if rpcCli == nil {
		return nil, 0
	}

	args := &AccIsExistArg{
		Username:    userName,
		Password:    passWord,
		Tel:         tel,
		Platform:    platform,
		Ts:          ts,
		LoginType:   logintype,
		TagKey:      tagkey,
		VerifyToken: verifyToken,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.AccountIsExist", args, ret, time.Second*30)
	if err != nil {
		return nil, 0
	}
	return ret.Acc, ret.Tag
}

func AccountTelIsRegiste(tel, platform string, tagkey int32) bool {
	if rpcCli == nil {
		return false
	}

	args := &AccIsExistArg{
		Tel:      tel,
		Platform: platform,
		TagKey:   tagkey,
	}
	var ret bool
	rpcCli.CallWithTimeout("AccountSvc.AccountTelIsRegiste", args, &ret, time.Second*30)
	return ret
}

func InsertAccount(userName, passWord, platform, channel, promoter, params, deviceOs string, inviterId, promoterTree int32,
	packTag, packname, deviceInfo string, subPromoter string, tagkey int32) (*Account, int) {
	if rpcCli == nil {
		return nil, 4
	}

	acc := NewAccount()
	if acc == nil {
		return nil, 4
	}

	tCur := time.Now()
	acc.UserName = userName
	acc.PassWord = passWord
	acc.TelPassWord = passWord
	acc.Platform = platform
	acc.Channel = channel
	acc.Params = params
	acc.SubPromoter = subPromoter
	acc.DeviceOs = deviceOs
	acc.Promoter = promoter
	acc.InviterId = inviterId
	acc.PromoterTree = promoterTree
	acc.LastLoginTime = tCur
	acc.RegisteTime = tCur
	acc.LoginTimes = 1
	acc.PackegeTag = packTag
	acc.Package = packname
	acc.DeviceInfo = deviceInfo
	acc.TagKey = tagkey
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.InsertAccount", acc, ret, time.Second*30)
	if err != nil {
		return nil, 0
	}
	return ret.Acc, ret.Tag
}

func InsertTelAccount(userName, passWord, platform, channel, promoter, params string, inviterId, promoterTree int32, tel, telpassword, packTag, packname, deviceInfo string, tagkey int32) (*Account, int) {
	if rpcCli == nil {
		return nil, 4
	}
	acc := NewAccount()
	if acc == nil {
		return nil, 4
	}

	tCur := time.Now()
	acc.UserName = userName
	acc.PassWord = passWord
	acc.Tel = tel
	acc.TelPassWord = telpassword
	acc.Platform = platform
	acc.Channel = channel
	acc.Promoter = promoter
	acc.Params = params
	acc.InviterId = inviterId
	acc.PromoterTree = promoterTree
	acc.LastLoginTime = tCur
	acc.RegisteTime = tCur
	acc.LoginTimes = 1
	acc.PackegeTag = packTag
	acc.Package = packname
	acc.DeviceInfo = deviceInfo
	acc.TagKey = tagkey
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.InsertTelAccount", acc, ret, time.Second*30)
	if err != nil {
		return nil, 4
	}
	return ret.Acc, ret.Tag
}

func UpgradeAccount(accId, tel, passWord, platform string, tagkey int32) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	args := &AccIsExistArg{
		AccId:    accId,
		Password: passWord,
		Tel:      tel,
		Platform: platform,
		TagKey:   tagkey,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.UpgradeAccount", args, ret, time.Second*30)
	if err != nil {
		return fmt.Errorf("UpgradeAccount err:%v tel:%v accid:%v", err, tel, accId)
	}
	if ret.Acc != nil && ret.Acc.Tel != "" {
		return errors.New("Tel is Bind")
	}
	return nil
}

func LogoutAccount(acc *Account) error {
	if acc == nil {
		return nil
	}
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.LogoutAccount", acc, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

type AccFreezeArg struct {
	AccId      string
	Platform   string
	FreezeTime int
}

func FreezeAccount(plt, accId string, freezeTime int) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &AccFreezeArg{
		AccId:      accId,
		Platform:   plt,
		FreezeTime: freezeTime,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.FreezeAccount", args, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

func UnfreezeAccount(plt, accId string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &AccFreezeArg{
		AccId:    accId,
		Platform: plt,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.UnfreezeAccount", args, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

type AccIdArg struct {
	AccId    string
	Platform string
}

type UserNameArg struct {
	UserName string
	Platform string
}

func GetAccount(plt, accId string) (*Account, error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &AccIdArg{
		AccId:    accId,
		Platform: plt,
	}
	var ret *Account
	err := rpcCli.CallWithTimeout("AccountSvc.GetAccount", args, ret, time.Second*30)
	if err != nil {
		return nil, err
	}
	return ret, err
}

func GetAccountByName(plt, username string) (*Account, error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &UserNameArg{
		UserName: username,
		Platform: plt,
	}
	var ret AccRet
	err := rpcCli.CallWithTimeout("AccountSvc.GetAccountByUserName", args, &ret, time.Second*30)
	if err != nil {
		return nil, err
	}
	return ret.Acc, err
}

// 删除账号
func RemoveAccount(plt, accId string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &AccIdArg{
		AccId:    accId,
		Platform: plt,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.RemoveAccount", args, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

// pwd暂时用明文.需要使用MD5加密。参照task_login.go
func EditAccountPwd(plt, accId, pwd string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &Account{
		AccountId: bson.ObjectIdHex(accId),
		PassWord:  pwd,
		Platform:  plt,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.EditAccountPwd", args, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

func ResetBackAccountPwd(plt, accId string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &AccIdArg{
		AccId:    accId,
		Platform: plt,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.ResetBackAccountPwd", args, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

func GenAccountPwd(pwd string) string {
	//md5(原始密码+AppId)
	raw := fmt.Sprintf("%v%v", pwd, common.GetAppId())
	h := md5.New()
	io.WriteString(h, raw)
	hashsum := hex.EncodeToString(h.Sum(nil))
	return hashsum
}

/*
 * 修改帐号密码
 */
func UpdatePlayerPassword(plt, accId, oldpassword, passWord string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &Account{
		AccountId:    bson.ObjectIdHex(accId),
		PassWord:     passWord,
		BackPassWord: oldpassword,
		Platform:     plt,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.UpdatePlayerPassword", args, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

/*
 * 修改Token帐号密码
 */
func UpdatePlayerTokenPassword(plt, accId, passWord string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &Account{
		AccountId: bson.ObjectIdHex(accId),
		PassWord:  passWord,
		Platform:  plt,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.UpdatePlayerTokenPassword", args, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

/*
 * 找回密码
 */
func GetBackPlayerPassword(tel, passWord, platform string, tagkey int32) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &Account{
		Tel:      tel,
		Platform: platform,
		TagKey:   tagkey,
		PassWord: passWord,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.GetBackPlayerPassword", args, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

func UpdateAccountPlatformInfo(account string, platform, channel, promoter, inviterid, promoterTree int32, packTag string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &Account{
		AccountId:    bson.ObjectIdHex(account),
		Platform:     strconv.Itoa(int(platform)),
		Channel:      strconv.Itoa(int(channel)),
		Promoter:     strconv.Itoa(int(promoter)),
		PromoterTree: promoterTree,
		InviterId:    inviterid,
		PackegeTag:   packTag,
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.UpdateAccountPlatformInfo", args, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

func GetRobotAccounts(limit int) []Account {
	if rpcCli == nil {
		return nil
	}

	var accs []Account
	err := rpcCli.CallWithTimeout("AccountSvc.GetRobotAccounts", limit, &accs, time.Second*30)
	if err != nil {
		return nil
	}
	return accs
}

func SaveToDelBackupAccount(acc *Account) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.SaveToDelBackupAccount", acc, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

func UpdateAccount(acc *Account) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	ret := &AccRet{}
	err := rpcCli.CallWithTimeout("AccountSvc.UpdateAccount", acc, ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}
