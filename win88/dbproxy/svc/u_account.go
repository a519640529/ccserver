package svc

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/rpc"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	AccountDBName            = "user"
	AccountCollName          = "user_account"
	AccountDelBackupCollName = "user_account_del_backup"
	ErrAccDBNotOpen          = model.NewDBError(AccountDBName, AccountCollName, model.NOT_OPEN)
)

func AccountCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, AccountDBName)
	if s != nil {
		c, first := s.DB().C(AccountCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"username", "tagkey"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"tel", "tagkey"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func AccountDelBackupCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, AccountDelBackupCollName)
	if s != nil {
		c, _ := s.DB().C(AccountCollName)
		return c
	}
	return nil
}

type AccountSvc struct {
}

// 登录检查
func (svc *AccountSvc) AccountCheck(args *model.AccIsExistArg, ret *model.AccRet) error {
	caccounts := AccountCollection(args.Platform)
	if caccounts == nil {
		ret.Tag = 1
		return ErrAccDBNotOpen
	}
	acc := &model.Account{}
	err := caccounts.Find(bson.M{"username": args.Username, "tagkey": args.TagKey}).One(acc)
	if err != nil {
		if err == mgo.ErrNotFound {
			logger.Logger.Warn("AccountCheck.error:", err)
			//注册
			ret.Tag = 2
			return nil
		} else {
			logger.Logger.Error("AccountCheck.error:", err)
		}
		return err
	}
	raw := fmt.Sprintf("%v%v%v", acc.PassWord, common.GetAppId(), args.Ts)
	h := md5.New()
	io.WriteString(h, raw)
	pwd := hex.EncodeToString(h.Sum(nil))

	rawt := fmt.Sprintf("%v%v%v", acc.TelPassWord, common.GetAppId(), args.Ts)
	ht := md5.New()
	io.WriteString(ht, rawt)
	pwdt := hex.EncodeToString(ht.Sum(nil))

	if pwd != args.Password && args.LoginType == 0 {
		logger.Logger.Warnf("Password is error:%v raw:%v get:%v expect:%v", acc.AccountId, raw, args.Password, pwd)
		ret.Tag = 3
		return nil
	} else if pwdt != args.Password && args.LoginType == 1 {
		logger.Logger.Warnf("Password is error:%v raw:%v get:%v expect:%v", acc.AccountId, raw, args.Password, pwd)
		ret.Tag = 3
		return nil
	} else if args.VerifyToken && args.LoginType == 5 {
		if acc.PassWord != args.Password {
			logger.Logger.Warnf("Password is error:%v raw:%v get:%v expect:%v", acc.AccountId, raw, args.Password, pwd)
			ret.Tag = 3
			return nil
		}
	}

	//freeze
	if acc.State > time.Now().Unix() {
		ret.Tag = 4
		ret.Acc = acc
		return nil
	}

	ret.Tag = 1
	ret.Acc = acc
	return nil
}

func (svc *AccountSvc) AccountIsExist(args *model.AccIsExistArg, ret *model.AccRet) error {
	caccounts := AccountCollection(args.Platform)
	if caccounts == nil {
		ret.Tag = 0
		return ErrAccDBNotOpen
	}
	acc := &model.Account{}
	var err error
	if args.LoginType == 0 {
		//游客登录
		err = caccounts.Find(bson.M{"username": args.Username, "tagkey": args.TagKey}).One(acc)
	} else if args.LoginType == 1 {
		//帐号登录
		//err = caccounts.Find(bson.M{"tel": args.Username, "tagkey": args.TagKey}).One(acc)
		err = caccounts.Find(bson.M{"username": args.Username, "tagkey": args.TagKey}).One(acc)
		if err != nil {
			logger.Logger.Info("Tel or Password is error:", err)
			ret.Tag = 3
			return nil
		}
	} else if args.LoginType == 3 {
		//帐号注册,验证手机号
		err = caccounts.Find(bson.M{"tel": args.Tel, "tagkey": args.TagKey}).One(acc)
		if err != nil {
			//手机号没有注册
		} else {
			//手机号已经注册
			//logger.Logger.Info("Tel is Exist:", err)
			ret.Tag = 8
			return nil
		}
		//帐号注册,验证游客帐号
		err = caccounts.Find(bson.M{"username": args.Username, "tagkey": args.TagKey}).One(acc)
		if err != nil {
			//设备号没有注册
			//logger.Logger.Info("Username is not Exist:", err)
			ret.Tag = 6
			return nil
		} else {
			//设备号已经注册
			//logger.Logger.Info("Username is Exist:", err)
			ret.Tag = 7
			return nil
		}
	} else if args.LoginType == 4 {
		//账号注册 不需要验证手机号
		//帐号注册,验证游客帐号
		err = caccounts.Find(bson.M{"username": args.Username, "tagkey": args.TagKey}).One(acc)
		if err != nil {
			//设备号没有注册
			//logger.Logger.Info("Username is not Exist:", err)
			ret.Tag = 6
			return nil
		} else {
			//设备号已经注册
			//logger.Logger.Info("Username is Exist:", err)
			ret.Tag = 7
			return nil
		}
	} else if args.LoginType == 5 {
		//telegram登录
		err = caccounts.Find(bson.M{"username": args.Username, "tagkey": args.TagKey}).One(acc)
	}

	if err != nil {
		//logger.Logger.Info("Account Is Exist error:", err)
		ret.Tag = 2
		return nil
	}
	raw := fmt.Sprintf("%v%v%v", acc.PassWord, common.GetAppId(), args.Ts)
	h := md5.New()
	io.WriteString(h, raw)
	pwd := hex.EncodeToString(h.Sum(nil))

	rawt := fmt.Sprintf("%v%v%v", acc.TelPassWord, common.GetAppId(), args.Ts)
	ht := md5.New()
	io.WriteString(ht, rawt)
	pwdt := hex.EncodeToString(ht.Sum(nil))

	if pwd != args.Password && args.LoginType == 0 {
		logger.Logger.Warnf("Password is error:%v raw:%v get:%v expect:%v", acc.AccountId, raw, args.Password, pwd)
		ret.Tag = 3
		return nil
	} else if pwdt != args.Password && args.LoginType == 1 {
		logger.Logger.Warnf("Password is error:%v raw:%v get:%v expect:%v", acc.AccountId, raw, args.Password, pwd)
		ret.Tag = 3
		return nil
	} else if args.VerifyToken && args.LoginType == 5 {
		if acc.PassWord != args.Password {
			logger.Logger.Warnf("Password is error:%v raw:%v get:%v expect:%v", acc.AccountId, raw, args.Password, pwd)
			ret.Tag = 3
			return nil
		}
	}

	//freeze
	if acc.State > time.Now().Unix() {
		ret.Tag = 4
		ret.Acc = acc
		return nil
	}

	ret.Tag = 1
	ret.Acc = acc
	return nil
}

func (svc *AccountSvc) AccountTelIsRegiste(args *model.AccIsExistArg, exist *bool) error {
	caccounts := AccountCollection(args.Platform)
	if caccounts != nil {
		err := caccounts.Find(bson.M{"tel": args.Tel, "tagkey": args.TagKey}).One(nil)
		if err == nil {
			*exist = true
		}
	}

	*exist = false
	return nil
}

func (svc *AccountSvc) InsertAccount(args *model.Account, ret *model.AccRet) error {
	caccounts := AccountCollection(args.Platform)
	if caccounts == nil {
		ret.Tag = 4
		return ErrAccDBNotOpen
	}

	id, err := GetOnePlayerIdFromBucket()
	if err == nil {
		args.SnId = id
	}

	err = caccounts.Insert(args)
	if err != nil {
		logger.Logger.Info("InsertAccount error:", err)
		ret.Tag = 4
		return nil
	}

	ret.Acc = args
	ret.Tag = 5
	return nil
}

func (svc *AccountSvc) InsertTelAccount(args *model.Account, ret *model.AccRet) error {
	caccounts := AccountCollection(args.Platform)
	if caccounts == nil {
		ret.Tag = 4
		return ErrAccDBNotOpen
	}

	id, err := GetOnePlayerIdFromBucket()
	if err == nil {
		args.SnId = id
	}

	err = caccounts.Insert(args)
	if err != nil {
		logger.Logger.Info("InsertAccount error:", err)
		ret.Tag = 4
		return nil
	}

	ret.Acc = args
	ret.Tag = 5
	return nil
}

func (svc *AccountSvc) UpgradeAccount(args *model.Account, ret *model.AccRet) error {
	caccounts := AccountCollection(args.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}

	acc := &model.Account{}
	err := caccounts.Find(bson.M{"tel": args.Tel, "tagkey": args.TagKey}).One(acc)
	if err == nil && acc.AccountId != args.AccountId {
		return fmt.Errorf("tel used account %v", acc.AccountId.Hex())
	}
	if acc.Tel != "" {
		return errors.New("Tel is Bind")
	}
	err = caccounts.Update(bson.M{"_id": args.AccountId}, bson.D{{"$set", bson.D{{"tel", args.Tel}, {"telpassword", args.PassWord}}}})
	if err != nil {
		logger.Logger.Info("UpgradeAccount error ", err)
	}

	return err
}

func (svc *AccountSvc) LogoutAccount(acc *model.Account, ret *model.AccRet) error {
	caccounts := AccountCollection(acc.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}
	return caccounts.Update(bson.M{"_id": acc.AccountId}, bson.D{{"$set", bson.D{{"logintimes", acc.LoginTimes}, {"lastlogouttime", acc.LastLogoutTime}}}})
}

func (svc *AccountSvc) FreezeAccount(args *model.AccFreezeArg, ret *model.AccRet) error {
	caccounts := AccountCollection(args.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}
	endTime := time.Now().Unix() + int64(time.Minute.Seconds())*int64(args.FreezeTime)
	return caccounts.Update(bson.M{"_id": bson.ObjectIdHex(args.AccId)}, bson.D{{"$set", bson.D{{"state", endTime}}}})
}

func (svc *AccountSvc) UnfreezeAccount(args *model.AccFreezeArg, ret *model.AccRet) error {
	caccounts := AccountCollection(args.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}
	return caccounts.Update(bson.M{"_id": bson.ObjectIdHex(args.AccId)}, bson.D{{"$set", bson.D{{"state", 0}}}})
}

func (svc *AccountSvc) getAccount(plt, accId string) (*model.Account, error) {
	caccounts := AccountCollection(plt)
	if caccounts == nil {
		return nil, ErrAccDBNotOpen
	}

	var acc model.Account
	err := caccounts.Find(bson.M{"_id": bson.ObjectIdHex(accId)}).One(&acc)
	if err != nil {
		return nil, err
	}

	return &acc, err
}

func (svc *AccountSvc) getAccountBySnId(plt string, snid int32) (*model.Account, error) {
	caccounts := AccountCollection(plt)
	if caccounts == nil {
		return nil, ErrAccDBNotOpen
	}

	var acc model.Account
	err := caccounts.Find(bson.M{"snid": snid}).One(&acc)
	if err != nil {
		return nil, err
	}

	return &acc, err
}

func (svc *AccountSvc) getAccountByUserName(plt, username string) (*model.Account, error) {
	caccounts := AccountCollection(plt)
	if caccounts == nil {
		return nil, ErrAccDBNotOpen
	}

	var acc model.Account
	err := caccounts.Find(bson.M{"username": username}).One(&acc)
	if err != nil {
		return nil, err
	}

	return &acc, err
}
func (svc *AccountSvc) GetAccountByUserName(args *model.UserNameArg, ret *model.AccRet) error {
	var err error
	ret.Acc, err = svc.getAccountByUserName(args.Platform, args.UserName)
	return err
}

func (svc *AccountSvc) GetAccount(args *model.AccIdArg, ret *model.AccRet) error {
	var err error
	ret.Acc, err = svc.getAccount(args.Platform, args.AccId)
	return err
}

// 删除账号
func (svc *AccountSvc) RemoveAccount(args *model.AccIdArg, ret *model.AccRet) error {
	caccounts := AccountCollection(args.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}

	return caccounts.RemoveId(bson.ObjectIdHex(args.AccId))
}

// pwd暂时用明文.需要使用MD5加密。参照task_login.go
func (svc *AccountSvc) EditAccountPwd(acc *model.Account, ret *model.AccRet) error {
	caccounts := AccountCollection(acc.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}
	hashsum := model.GenAccountPwd(acc.PassWord)
	return caccounts.Update(bson.M{"_id": acc.AccountId}, bson.D{{"$set", bson.D{{"backpassword", acc.TelPassWord}, {"telpassword", hashsum}}}})
}

func (svc *AccountSvc) ResetBackAccountPwd(args *model.AccIdArg, ret *model.AccRet) error {
	caccounts := AccountCollection(args.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}

	var acc model.Account
	err := caccounts.Find(bson.M{"_id": bson.ObjectIdHex(args.AccId)}).One(&acc)
	if err != nil {
		return err
	}
	if len(acc.BackPassWord) > 0 {
		return caccounts.Update(bson.M{"_id": acc.AccountId}, bson.D{{"$set", bson.D{{"telpassword", acc.BackPassWord}}}})
	}
	return nil
}

/*
 * 修改帐号密码
 */
func (svc *AccountSvc) UpdatePlayerPassword(acc *model.Account, ret *model.AccRet) error {
	caccounts := AccountCollection(acc.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}

	return caccounts.Update(bson.M{"_id": acc.AccountId, "telpassword": acc.BackPassWord}, bson.D{{"$set", bson.D{{"telpassword", acc.PassWord}}}})
}

/*
 * 修改Token帐号密码
 */
func (svc *AccountSvc) UpdatePlayerTokenPassword(acc *model.Account, ret *model.AccRet) error {
	caccounts := AccountCollection(acc.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}

	return caccounts.Update(bson.M{"_id": acc.AccountId}, bson.D{{"$set", bson.D{{"password", acc.PassWord}}}})
}

/*
 * 找回密码
 */
func (svc *AccountSvc) GetBackPlayerPassword(acc *model.Account, ret *model.AccRet) error {
	caccounts := AccountCollection(acc.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}

	return caccounts.Update(bson.M{"tel": acc.Tel, "tagkey": acc.TagKey}, bson.D{{"$set", bson.D{{"telpassword", acc.PassWord}}}})
}
func (svc *AccountSvc) UpdateAccountPlatformInfo(acc *model.Account, ret *model.AccRet) error {
	caccounts := AccountCollection(acc.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}

	return caccounts.Update(bson.M{"_id": acc.AccountId},
		bson.D{{"$set", bson.M{"platform": acc.Platform,
			"channel": acc.Channel, "promoter": acc.Promoter, "inviterid": acc.InviterId, "packegetag": acc.PackegeTag, "promotertree": acc.PromoterTree}}})
}

func (svc *AccountSvc) GetRobotAccounts(limit int, accs *[]model.Account) error {
	caccounts := AccountCollection(mongo.G_P)
	if caccounts == nil {
		return nil
	}
	*accs = make([]model.Account, 0, limit)
	err := caccounts.Find(bson.M{"platform": common.Platform_Rob}).Limit(limit).All(&accs)
	if err != nil {
		logger.Logger.Info("GetAllRobotAccounts  error:", err)
		return nil
	}
	return nil
}

func (svc *AccountSvc) SaveToDelBackupAccount(acc *model.Account, ret *model.AccRet) error {
	cDelBackup := AccountDelBackupCollection(acc.Platform)
	if cDelBackup == nil {
		return ErrAccDBNotOpen
	}

	_, err := cDelBackup.Upsert(bson.M{"_id": acc.AccountId}, acc)
	if err != nil {
		logger.Logger.Info("InsertDelBackupAccount error:", err)
	}

	return err
}

func (svc *AccountSvc) UpdateAccount(acc *model.Account, ret *model.AccRet) error {
	caccounts := AccountCollection(acc.Platform)
	if caccounts == nil {
		return ErrAccDBNotOpen
	}

	err := caccounts.UpdateId(acc.AccountId, acc)
	if err != nil {
		logger.Logger.Info("UpdateAccount error:", err)
	}

	return err
}

var _AccountSvc = &AccountSvc{}

func init() {
	rpc.Register(_AccountSvc)
}
