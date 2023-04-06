package svc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/task"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	PlayerDBName            = "user"
	PlayerCollName          = "user_playerinfo"
	PlayerDelBackupCollName = "user_playerinfo_del_backup"
	PlayerColError          = errors.New("Playerinfo collection open failed")
)

func PlayerDataCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, PlayerDBName)
	if s != nil {
		c_playerdata, first := s.DB().C(PlayerCollName)
		if first {
			c_playerdata.EnsureIndex(mgo.Index{Key: []string{"accountid"}, Unique: true, Background: true, Sparse: true})
			c_playerdata.EnsureIndex(mgo.Index{Key: []string{"snid"}, Unique: true, Background: true, Sparse: true})
			c_playerdata.EnsureIndex(mgo.Index{Key: []string{"channel"}, Background: true, Sparse: true})
			c_playerdata.EnsureIndex(mgo.Index{Key: []string{"tel"}, Background: true, Sparse: true})
			c_playerdata.EnsureIndex(mgo.Index{Key: []string{"bankaccount"}, Background: true, Sparse: true})
			c_playerdata.EnsureIndex(mgo.Index{Key: []string{"alipayaccount"}, Background: true, Sparse: true})
			c_playerdata.EnsureIndex(mgo.Index{Key: []string{"alipayaccname"}, Background: true, Sparse: true})
			c_playerdata.EnsureIndex(mgo.Index{Key: []string{"bankaccname"}, Background: true, Sparse: true})
		}
		return c_playerdata
	}

	return nil
}

type PlayerDelBackupDataSvc struct {
}

func PlayerDelBackupDataCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, PlayerDBName)
	if s != nil {
		c_playerdata, _ := s.DB().C(PlayerDelBackupCollName)
		return c_playerdata
	}
	return nil
}

type PlayerDataSvc struct {
}

//func (svc *PlayerDataSvc) SavePlayerRebate(pd *model.PlayerData, ret *bool) error {
//	cplayerdata := PlayerDataCollection(pd.Platform)
//	if cplayerdata == nil {
//		return nil
//	}
//	err := cplayerdata.Update(bson.M{"snid": pd.SnId}, bson.D{{"$set", bson.D{{"rebatedata", pd.RebateData}, {"totalconvertibleflow", pd.TotalConvertibleFlow}}}})
//	if err != nil {
//		logger.Logger.Trace("SavePlayerRebate failed:", err)
//		ret.Err = err
//		return err
//	}
//	return nil
//}

// 获取代理信息
func (svc *PlayerDataSvc) GetAgentInfo(args *model.GetAgentInfoArgs, ret *model.PlayerData) error {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return nil
	}
	err := cplayerdata.Find(bson.M{"tel": args.Tel}).One(ret)
	if err != nil {
		logger.Logger.Error("GetAgentInfo failed:", err)
		return err
	}
	if CorrectData(ret) {
	}
	return nil
}
func (svc *PlayerDataSvc) GetPlayerData(args *model.GetPlayerDataArgs, ret *model.PlayerDataRet) error {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return nil
	}

	pd := &model.PlayerData{}
	fileName := fmt.Sprintf("%v.json", args.Acc)
	_, err := os.Stat(fileName)
	if err == nil || os.IsExist(err) {
		pd = model.RestorePlayerData(fileName)
		err = os.Remove(fileName)
	} else {
		if args.Acc != "" {
			err = cplayerdata.Find(bson.M{"accountid": args.Acc}).One(pd)
		}
	}

	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		logger.Logger.Trace("GetPlayerinfo Find failed:", err)

		if !bson.IsObjectIdHex(args.Acc) {
			logger.Logger.Warn("NewPlayer failed: acc is illeage ", args.Acc)
			return nil
		}

		a, err := _AccountSvc.getAccount(args.Plt, args.Acc)
		if err != nil {
			logger.Logger.Warnf("model.GetAccount(%v) failed:%v", args.Acc, err)
			return nil
		}

		id := a.SnId
		if id == 0 {
			id, err = GetOnePlayerIdFromBucket()
			if err != nil {
				logger.Logger.Warn("NewPlayer failed:", err)
				return err
			}
		}

		var dataParams model.PlayerParams
		json.Unmarshal([]byte(a.Params), &dataParams)

		name := dataParams.Name
		if name == "" {
			name = model.GameParamData.GuestDefaultName //fmt.Sprintf("贵宾%v", id)
		}
		if name == "" {
			name = "贵宾"
		}

		pd = model.NewPlayerData(args.Acc, name, id, a.Channel, a.Platform, a.Promoter, a.InviterId, a.PromoterTree, a.Params,
			a.Tel, a.PackegeTag, dataParams.Ip, 0, dataParams.UnionId, a.DeviceInfo, a.SubPromoter, a.TagKey)
		if pd != nil {
			err = cplayerdata.Insert(pd)
			if err != nil {
				logger.Logger.Errorf("GetPlayerinfo Insert err:%v acc:%v snid:%v", err, args.Acc, id)
				return err
			}
			ret.IsNew = true
			ret.Pd = pd
			return nil
		}
		return nil
	}

	//todo 修正玩家的离线时间，上次登录时间，创建时间
	pd.LastLogoutTime = pd.LastLogoutTime.Local()
	pd.LastLoginTime = pd.LastLoginTime.Local()
	pd.CreateTime = pd.CreateTime.Local()

	if CorrectData(pd) {
	}
	ret.Pd = pd
	return nil
}

// 谷歌 facebook 账号数据
func (svc *PlayerDataSvc) CreatePlayerDataByThird(args *model.CreatePlayer, ret *model.PlayerDataRet) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	acc := args.AccId
	pd := &model.PlayerData{}
	fileName := fmt.Sprintf("%v.json", acc)
	_, err = os.Stat(fileName)
	if err == nil || os.IsExist(err) {
		pd = model.RestorePlayerData(fileName)
		err = os.Remove(fileName)
	} else {
		if acc != "" {
			err = cplayerdata.Find(bson.M{"accountid": acc}).One(pd)
		}
	}

	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		logger.Logger.Trace("CreatePlayerDataOnRegister Find failed:", err)

		if !bson.IsObjectIdHex(acc) {
			logger.Logger.Warn("NewPlayer failed: acc is illeage ", acc)
			return
		}
		var a *model.Account
		a, err = _AccountSvc.getAccount(args.Plt, args.AccId)
		if err != nil {
			logger.Logger.Warnf("_AccountSvc.getAccount(%v,%v) failed:%v", args.Plt, args.AccId, err)
			return
		}

		id := a.SnId
		if id == 0 {
			id, err = GetOnePlayerIdFromBucket()
			if err != nil {
				logger.Logger.Warn("NewPlayer failed:", err)
				return
			}
		}

		//name := model.GameParamData.GuestDefaultName //fmt.Sprintf("db%v", id)
		name := args.NickName
		if name == "" {
			name = "贵宾"
		}
		var dataParams model.PlayerParams
		json.Unmarshal([]byte(a.Params), &dataParams)
		pd = model.NewPlayerDataThird(acc, name, args.HeadUrl, id, a.Channel, a.Platform, a.Promoter, a.InviterId,
			a.PromoterTree, a.Params, a.Tel, a.PackegeTag, dataParams.Ip, a.SubPromoter, a.TagKey)
		if pd != nil {
			err = cplayerdata.Insert(pd)
			if err != nil {
				logger.Logger.Trace("CreatePlayerDataOnRegister Insert failed:", err)
				return
			}
			ret.Pd = pd
			ret.IsNew = true
			return
		}
		return
	}
	if CorrectData(pd) {
	}
	ret.Pd = pd
	return nil
}

func (svc *PlayerDataSvc) CreatePlayerDataOnRegister(args *model.PlayerDataArg, ret *model.PlayerDataRet) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	acc := args.AccId
	pd := &model.PlayerData{}
	fileName := fmt.Sprintf("%v.json", acc)
	_, err = os.Stat(fileName)
	if err == nil || os.IsExist(err) {
		pd = model.RestorePlayerData(fileName)
		err = os.Remove(fileName)
	} else {
		if acc != "" {
			err = cplayerdata.Find(bson.M{"accountid": acc}).One(pd)
		}
	}

	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		logger.Logger.Trace("CreatePlayerDataOnRegister Find failed:", err)

		if !bson.IsObjectIdHex(acc) {
			logger.Logger.Warn("NewPlayer failed: acc is illeage ", acc)
			return
		}
		var a *model.Account
		a, err = _AccountSvc.getAccount(args.Plt, args.AccId)
		if err != nil {
			logger.Logger.Warnf("_AccountSvc.getAccount(%v,%v) failed:%v", args.Plt, args.AccId, err)
			return
		}

		id := a.SnId
		if id == 0 {
			id, err = GetOnePlayerIdFromBucket()
			if err != nil {
				logger.Logger.Warn("NewPlayer failed:", err)
				return
			}
		}

		name := model.GameParamData.GuestDefaultName //fmt.Sprintf("db%v", id)
		if name == "" {
			name = "贵宾"
		}
		var dataParams model.PlayerParams
		json.Unmarshal([]byte(a.Params), &dataParams)
		pd = model.NewPlayerData(acc, name, id, a.Channel, a.Platform, a.Promoter, a.InviterId,
			a.PromoterTree, a.Params, a.Tel, a.PackegeTag, dataParams.Ip, int64(args.AddCoin),
			"", a.DeviceInfo, a.SubPromoter, a.TagKey)
		if pd != nil {
			err = cplayerdata.Insert(pd)
			if err != nil {
				logger.Logger.Trace("CreatePlayerDataOnRegister Insert failed:", err)
				return
			}
			ret.Pd = pd
			ret.IsNew = true
			return
		}
		return
	}
	if CorrectData(pd) {
	}
	ret.Pd = pd
	return
}

func (svc *PlayerDataSvc) GetPlayerDataBySnId(args *model.GetPlayerDataBySnIdArgs, ret *model.PlayerDataRet) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return nil
	}

	err = cplayerdata.Find(bson.M{"snid": args.SnId}).One(&ret.Pd)
	if err != nil {
		logger.Logger.Tracef("from %v Get %v player data error:%v", args.Plt, args.SnId, err)
		if args.CreateIfNotExist && err == mgo.ErrNotFound {
			a, err := _AccountSvc.getAccountBySnId(args.Plt, args.SnId)
			if err != nil {
				logger.Logger.Warnf("model.getAccountBySnId(%v) failed:%v", args.SnId, err)
				return err
			}

			var dataParams model.PlayerParams
			json.Unmarshal([]byte(a.Params), &dataParams)

			name := dataParams.Name
			if name == "" {
				name = model.GameParamData.GuestDefaultName //fmt.Sprintf("贵宾%v", id)
			}
			if name == "" {
				name = "贵宾"
			}

			pd := model.NewPlayerData(a.AccountId.Hex(), name, a.SnId, a.Channel, a.Platform, a.Promoter, a.InviterId, a.PromoterTree, a.Params,
				a.Tel, a.PackegeTag, dataParams.Ip, 0, dataParams.UnionId, a.DeviceInfo, a.SubPromoter, a.TagKey)
			if pd != nil {
				err = cplayerdata.Insert(pd)
				if err != nil {
					logger.Logger.Errorf("GetPlayerDataBySnId Insert err:%v acc:%v snid:%v", err, a.AccountId.Hex(), a.SnId)
					return err
				}
				ret.IsNew = true
				ret.Pd = pd
				return nil
			}
		}
		return err
	}

	if args.CorrectData && ret.Pd != nil {
		CorrectData(ret.Pd)
	}

	return nil
}

//func (svc *PlayerDataSvc) UpdateCreateCreateClubNum(snId int32, ret *model.PlayerDataRet) error {
//	cplayerdata := PlayerDataCollection()
//	if cplayerdata == nil {
//		return nil
//	}
//	err := cplayerdata.Update(bson.M{"snid": snId}, bson.D{{"$inc", bson.D{{"createclubnum", 1}}}})
//	if err != nil {
//		logger.Logger.Error("Update player safeboxcoin error:", err)
//		ret.Err = err
//		return err
//	}
//	return nil
//}

func (svc *PlayerDataSvc) GetPlayerDatasBySnIds(args *model.GetPlayerDatasBySnIdsArgs, ret *[]*model.PlayerData) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return nil
	}
	err = cplayerdata.Find(bson.M{"snid": bson.M{"$in": args.SnIds}}).All(ret)
	if err != nil {
		logger.Logger.Tracef("GetPlayerDatasBySnIds(snids=%v) error:%v", args.SnIds, err)
		return
	}

	if args.CorrectData {
		for _, e := range *ret {
			if CorrectData(e) {

			}
		}
	}

	return
}
func (svc *PlayerDataSvc) GetPlayerDataByUnionId(args *model.GetPlayerDataByUnionIdArgs, ret *model.PlayerDataRet) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Find(bson.M{"unionid": args.UnionId}).One(&ret.Pd)
	if err != nil {
		logger.Logger.Tracef("GetPlayerDataByUnionId % player data error:%v", args.UnionId, err)
		return
	}
	if args.CorrectData {
		if CorrectData(ret.Pd) {

		}
	}
	return
}

func (svc *PlayerDataSvc) GetPlayerTel(args *model.GetPlayerTelArgs, ret *model.GetPlayerTelRet) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Find(bson.M{"tel": args.Tel}).Select(bson.M{"snid": 1}).One(ret)
	if err != nil {
		logger.Logger.Warn("model.PlayerData.Find err:", err)
		return
	}
	return
}
func (svc *PlayerDataSvc) GetPlayerCoin(args *model.GetPlayerCoinArgs, ret *model.GetPlayerCoinRet) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Find(bson.M{"snid": args.SnId}).Select(bson.M{"coin": 1, "safeboxcoin": 1}).One(ret)
	if err != nil {
		logger.Logger.Trace("GetPlayerCoin error:", err)
		ret.Err = err
		return
	}
	return
}

func SavePlayerData(pd *model.PlayerData) (err error) {
	cplayerdata := PlayerDataCollection(pd.Platform)
	if cplayerdata == nil {
		return
	}
	if pd != nil {
		model.RecalcuPlayerCheckSum(pd)
		_, err = cplayerdata.Upsert(bson.M{"_id": pd.Id}, pd)
		if err != nil {
			logger.Logger.Errorf("model.SavePlayerData %v err:%v", pd.SnId, err)
			return
		}

		//清理coinWAL
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			RemoveCoinWALByCoinType(pd.Platform, pd.SnId, PayCoinLogType_Coin, pd.CoinPayTs)
			RemoveCoinWALInGame(pd.Platform, pd.SnId, 0, pd.GameCoinTs)
			return nil
		}), nil, "RemoveCoinWAL").StartByFixExecutor("RemoveCoinWAL")
	}
	return
}

/*
 * 保存玩家的全部信息
 */
func (svc *PlayerDataSvc) SavePlayerData(pd *model.PlayerData, ret *bool) (err error) {
	err = SavePlayerData(pd)
	*ret = err == nil
	return
}

func (svc *PlayerDataSvc) RemovePlayer(args *model.RemovePlayerArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return nil
	}
	err = cplayerdata.Remove(bson.M{"snid": args.SnId})
	if err != nil {
		logger.Logger.Info("Remove player failed.")
		return
	}
	*ret = true
	return
}
func (svc *PlayerDataSvc) RemovePlayerByAcc(args *model.RemovePlayerByAccArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return nil
	}
	err = cplayerdata.Remove(bson.M{"accountid": args.Acc})
	if err != nil {
		logger.Logger.Info("Remove player failed.")
		return
	}
	*ret = true
	return
}

/*
 * 检查手机号是否存在
 */
func (svc *PlayerDataSvc) GetPlayerSnid(args *model.GetPlayerSnidArgs, ret *int32) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Find(bson.M{"accountid": args.Acc}).Select(bson.M{"snid": 1}).One(ret)
	if err != nil {
		logger.Logger.Error("svc.GetPlayerSnid is error ", err)
		return
	}
	return
}

/*
 * 检查昵称是否存在
 */
func (svc *PlayerDataSvc) PlayerNickIsExist(args *model.PlayerNickIsExistArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Find(bson.M{"name": args.Name}).One(nil)
	if err != nil {
		logger.Logger.Error("svc.PlayerNickIsExist is error", err)
		return err
	}
	*ret = true
	return
}

/*
 * 检查玩家是否存在
 */
func (svc *PlayerDataSvc) PlayerIsExistBySnId(args *model.PlayerIsExistBySnIdArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Find(bson.M{"snid": args.SnId}).One(nil)
	if err != nil {
		logger.Logger.Error("svc.GetPlayerIsExistBySnId is error", err)
		return
	}

	*ret = true
	return
}

// 修改推广包标识
func (svc *PlayerDataSvc) UpdatePlayerPackageId(args *model.UpdatePackageId, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.PlatformStr)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.M{"packageid": args.Tag,
		"platform": strconv.Itoa(int(args.Platform)), "channel": strconv.Itoa(int(args.Channel)),
		"beunderagentcode": strconv.Itoa(int(args.Promoter)), "inviterid": args.Inviterid, "promotertree": args.PromoterTree}}})
	if err != nil {
		logger.Logger.Error("Update player packageid error:", err)
		return err
	}
	*ret = true
	return nil
}
func (svc *PlayerDataSvc) UpdatePlayerPackageIdByStr(args *model.UpdatePackageId, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.PlatformStr)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.M{"packageid": args.Tag,
		"platform": args.PlatformStr, "channel": args.ChannelStr,
		"beunderagentcode": args.PromoterStr, "inviterid": args.Inviterid, "promotertree": args.PromoterTree}}})
	if err != nil {
		logger.Logger.Trace("UpdatePlayerPackageIdByStr error:", err)
		return err
	}

	*ret = true
	return nil
}

func (svc *PlayerDataSvc) UpdatePlayerPackageIdByAcc(args *model.UpdatePackageId, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.PlatformStr)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"accountid": args.AccId}, bson.D{{"$set", bson.M{"packageid": args.Tag,
		"platform": args.PlatformStr, "channel": args.ChannelStr,
		"beunderagentcode": args.PromoterStr, "inviterid": args.Inviterid, "promotertree": args.PromoterTree}}})
	if err != nil {
		logger.Logger.Trace("UpdatePlayerPackageIdByAcc error:", err)
		return
	}

	*ret = true
	return
}

// 修改平台
func (svc *PlayerDataSvc) UpdatePlayerPlatform(args *model.UpdatePackageId, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.PlatformStr)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set",
		bson.D{{"platform", args.PlatformStr}, {"channel", args.ChannelStr}}}})
	if err != nil {
		logger.Logger.Trace("Update player nick error:", err)
		return
	}

	*ret = true
	return
}

// 修改玩家无级代推广员id
func (svc *PlayerDataSvc) UpdatePlayerPromoterTree(args *model.UpdatePackageId, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.PlatformStr)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"promotertree", args.PromoterTree}}}})
	if err != nil {
		logger.Logger.Trace("UpdatePlayerPromoterTree error:", err)
		return
	}

	*ret = true
	return
}

// 修改玩家全民推广
func (svc *PlayerDataSvc) UpdatePlayerInviteID(args *model.UpdatePackageId, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.PlatformStr)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"inviterid", args.Inviterid}}}})
	if err != nil {
		logger.Logger.Trace("UpdatePlayerInviteID error:", err)
		return
	}

	*ret = true
	return
}

/*
 * 修改昵称
 */
func (svc *PlayerDataSvc) UpdatePlayerNick(args *model.UpdatePlayerInfo, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Update(bson.M{"accountid": args.Acc}, bson.D{{"$set", bson.D{{"name", args.Nick}}}})
	if err != nil {
		logger.Logger.Trace("Update player nick error:", err)
		return
	}

	*ret = true
	return
}

/*
 * 修改黑名单类型
 */
func (svc *PlayerDataSvc) UpdatePlayerBlacklistType(args *model.UpdatePlayerBlacklistTypeArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"blacklisttype", args.BlackListType}}}})
	if err != nil {
		logger.Logger.Trace("Update player blacklisttype error:", err)
		return
	}

	*ret = true
	return
}

/*
 * 修改玩家备注信息
 */
func (svc *PlayerDataSvc) UpdatePlayerMarkInfo(args *model.UpdatePlayerMarkInfoArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"markinfo", args.MarkInfo}}}})
	if err != nil {
		logger.Logger.Warnf("Update player mark error:", err)
		return
	}

	*ret = true
	return
}

/*
 * 修改玩家特殊白名单
 */
func (svc *PlayerDataSvc) UpdatePlayerWhiteFlag(args *model.UpdatePlayerWhiteFlagArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"whiteflag", args.WhiteFlag}}}})
	if err != nil {
		logger.Logger.Warnf("Update player whiteFlag error:", err)
	}

	*ret = true
	return
}

/*
 * 修改玩家支付信息
 */
func (svc *PlayerDataSvc) UpdatePlayerPayAct(args *model.UpdatePlayerPayActArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"payactstate", args.PayActState}}}})
	if err != nil {
		logger.Logger.Warnf("Update player payact error:", err)
		return
	}

	*ret = true
	return
}

/*
 * 修改玩家电销标记
 */
func (svc *PlayerDataSvc) UpdatePlayerTelephonePromoter(args *model.UpdatePlayerTelephonePromoterArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set",
		bson.D{{"telephonepromoter", args.TelephonePromoter}}}})
	if err != nil {
		logger.Logger.Warnf("Update player telephonepromoter error:", err)
		return
	}

	*ret = true
	return
}

/*
 * 修改玩家电销标记
 */
func (svc *PlayerDataSvc) UpdatePlayerTelephoneCallNum(args *model.UpdatePlayerTelephonePromoterArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set",
		bson.D{{"telephonecallnum", args.TelephoneCallNum}}}})
	if err != nil {
		logger.Logger.Warnf("Update player telephonecallnum error:", err)
		return
	}

	*ret = true
	return
}

/*
 * 修改头像
 */
func (svc *PlayerDataSvc) UpdatePlayeIcon(args *model.UpdatePlayeIconArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Update(bson.M{"accountid": args.Acc}, bson.D{{"$set", bson.D{{"head", args.Head}}}})
	if err != nil {
		logger.Logger.Trace("Update player icon error:", err)
		return
	}

	*ret = true
	return
}

/*
 * 修改性别
 */
func (svc *PlayerDataSvc) UpdatePlayeSex(args *model.UpdatePlayeSexArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Update(bson.M{"accountid": args.Acc}, bson.D{{"$set", bson.D{{"sex", args.Sex}}}})
	if err != nil {
		logger.Logger.Trace("Update player sex error:", err)
		return
	}

	*ret = true
	return
}

/*
 * 检查手机号是否存在
 */
func (svc *PlayerDataSvc) PlayerTelIsExist(args *model.PlayerTelIsExistArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Find(bson.M{"tel": args.Tel, "tagkey": args.TagKey}).One(nil)
	if err != nil {
		return
	}

	*ret = true
	return
}

/*
 * 绑定手机号
 */
func (svc *PlayerDataSvc) UpdatePlayerTel(args *model.UpdatePlayerTelArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"tel", args.Tel}}}})
	if err != nil {
		logger.Logger.Warn("UpdatePlayerTel error:", err)
		return
	}
	*ret = true
	return
}

func (svc *PlayerDataSvc) CountAlipayAccountCount(args *model.CountAlipayAccountCountArgs, ret *int) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return
	}
	n, err := cplayerdata.Find(bson.M{"alipayaccount": args.AliPayAccount}).Count()
	if err != nil {
		logger.Logger.Warn("CountAlipayAccountCount error:", err)
		return err
	}
	*ret = n
	return
}

func (svc *PlayerDataSvc) CountBankAlipayNameCount(args *model.CountBankAlipayNameCountArgs, ret *int) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return
	}
	var conds []bson.M
	conds = append(conds, bson.M{"alipayaccname": args.AliPayAccName})
	conds = append(conds, bson.M{"bankaccname": args.BankAccName})
	n, err := cplayerdata.Find(bson.M{"snid": bson.M{"$ne": args.SnId}, "$or": conds}).Count()
	if err != nil {
		logger.Logger.Warn("CountBankAlipayNameCount error:", err)
		return
	}
	*ret = n
	return
}

/*
 * 绑定支付宝账号
 */
func (svc *PlayerDataSvc) UpdatePlayerAlipay(args *model.UpdatePlayerAlipayArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set",
		bson.D{{"alipayaccount", args.AliPayAccount}, {"alipayaccname", args.AliPayAccName}}}})
	if err != nil {
		logger.Logger.Warn("UpdatePlayerAlipay error:", err)
		return
	}
	*ret = true
	return
}
func (svc *PlayerDataSvc) UpdatePlayerAlipayAccount(args *model.UpdatePlayerAlipayArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"alipayaccount", args.AliPayAccount}}}})
	if err != nil {
		logger.Logger.Warn("UpdatePlayerAlipay error:", err)
		return
	}
	*ret = true
	return
}
func (svc *PlayerDataSvc) UpdatePlayerAlipayName(args *model.UpdatePlayerAlipayArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set",
		bson.D{{"alipayaccname", args.AliPayAccName}}}})
	if err != nil {
		return
	}
	*ret = true
	return
}

func (svc *PlayerDataSvc) CountBankAccountCount(args *model.CountBankAccountCountArgs, ret *int) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return
	}
	n, err := cplayerdata.Find(bson.M{"bankaccount": args.BankAccount}).Count()
	if err != nil {
		logger.Logger.Warn("CountBankAccountCount error:", err)
		return
	}
	*ret = n
	return
}

/*
 * 绑定银行账号
 */
func (svc *PlayerDataSvc) UpdatePlayerBank(args *model.UpdatePlayerBankArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return errors.New("user_playerinfo not open")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"bank", args.Bank},
		{"bankaccount", args.BankAccount}, {"bankaccname", args.BankAccName}}}})
	if err != nil {
		logger.Logger.Warn("UpdatePlayerBank error:", err)
		return
	}
	*ret = true
	return
}

/*
 * 修改玩家是否返利
 */
func (svc *PlayerDataSvc) UpdatePlayerIsRebate(args *model.UpdatePlayerIsRebateArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"iscanrebate", args.IsCanRebate}}}})
	if err != nil {
		logger.Logger.Trace("Update player isRebate error:", err)
		return
	}
	*ret = true
	return
}

/*
 * 修改玩家是否可以修改昵称
 */
func (svc *PlayerDataSvc) UpdatePlayerIsStopRename(args *model.UpdatePlayerIsStopRenameArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return errors.New("param err")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"isstoprename", args.IsStopReName}}}})
	if err != nil {
		logger.Logger.Trace("Update player isRename error:", err)
		return
	}
	*ret = true
	return
}

func (svc *PlayerDataSvc) PlayerRebindSnId(args *model.PlayerRebindSnIdArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return errors.New("cplayerdata == nil")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"snid", args.NewSnId}}}})
	if err != nil {
		logger.Logger.Trace("PlayerRebindSnId error:", err)
		return
	}
	*ret = true
	return
}

/*
 * 修改保险箱密码
 */
func (svc *PlayerDataSvc) UpdateSafeBoxPassword(args *model.PlayerSafeBoxArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Update(bson.M{"accountid": args.Acc, "safeboxpassword": args.OldPassWord},
		bson.D{{"$set", bson.D{{"safeboxpassword", args.PassWord}}}})
	if err != nil {
		logger.Logger.Trace("Update player safeboxpassword error:", err)
		return
	}
	*ret = true
	return
}

/*
 * 找回保险箱密码
 */
func (svc *PlayerDataSvc) GetBackSafeBoxPassword(args *model.PlayerSafeBoxArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return errors.New("GetBackSafeBoxPassword error")
	}
	err = cplayerdata.Update(bson.M{"tel": args.Tel, "tagkey": args.TagKey},
		bson.D{{"$set", bson.D{{"safeboxpassword", args.PassWord}}}})
	if err != nil {
		logger.Logger.Trace("Update player safeboxpassword error:", err)
		return
	}
	*ret = true
	return
}

/*
 * 重置保险箱密码
 */
func (svc *PlayerDataSvc) ResetSafeBoxPassword(args *model.PlayerSafeBoxArgs, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return PlayerColError
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{{"safeboxpassword", args.PassWord}}}})
	if err != nil {
		logger.Logger.Trace("Reset player safeboxpassword error:", err)
		return
	}
	*ret = true
	return
}

func (svc *PlayerDataSvc) UpdatePlayerCoin(args *model.PlayerSafeBoxCoin, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return fmt.Errorf("db may be closed")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{
		{"coin", args.Coin},
		{"diamond", args.Diamond},
		{"safeboxcoin", args.SafeBoxCoin},
		{"coinpayts", args.CoinPayTs},
		{"safeboxcoints", args.SafeBoxCoinTs}}}})
	if err != nil {
		logger.Logger.Error("UpdatePlayerCoin error:", err)
		return
	}
	*ret = true
	return
}

func (svc *PlayerDataSvc) UpdatePlayerSetCoin(args *model.PlayerSafeBoxCoin, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return fmt.Errorf("db may be closed")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{
		{"coin", args.Coin}}}})
	if err != nil {
		logger.Logger.Error("svc.UpdatePlayerSetCoin error:", err)
		return
	}
	*ret = true
	return
}

func (svc *PlayerDataSvc) UpdatePlayerLastExchangeTime(args *model.PlayerSafeBoxCoin, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return fmt.Errorf("db may be closed")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{
		{"lastexchangetime", args.LastexChangeTime}}}})
	if err != nil {
		logger.Logger.Error("svc.UpdateLastExchangeTime error:", err)
		return
	}
	*ret = true
	return
}

func (svc *PlayerDataSvc) UpdatePlayerExchageFlow(args *model.PlayerSafeBoxCoin, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return fmt.Errorf("db may be closed")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{
		{"totalconvertibleflow", args.TotalConvertibleFlow},
		{"totalflow", args.TotalFlow}}}})
	if err != nil {
		logger.Logger.Error("svc.UpdatePlayerExchageFlow error:", err)
		return
	}
	*ret = true
	return
}
func (svc *PlayerDataSvc) UpdatePlayerExchageFlowAndOrder(args *model.PlayerSafeBoxCoin, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return fmt.Errorf("db may be closed")
	}
	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bson.D{
		{"totalconvertibleflow", args.TotalConvertibleFlow},
		{"lastexchangeorder", args.LastexChangeOrder}}}})
	if err != nil {
		logger.Logger.Error("svc.UpdatePlayerExchageFlowAndOrder error:", err)
		return
	}
	*ret = true
	return
}

/*
 * 查找用户
 */
func (svc *PlayerDataSvc) FindPlayerList(args *model.PlayerSelect, ret *[]*model.PlayerBaseInfo) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return
	}

	var conds []bson.M
	if args.Id != 0 {
		conds = append(conds, bson.M{"snid": args.Id})
	}
	if len(args.Tel) > 0 {
		conds = append(conds, bson.M{"tel": args.Tel})
	}
	if len(args.NickName) > 0 {
		conds = append(conds, bson.M{"name": args.NickName})
	}
	if args.Coinl > 0 || args.Coinh > 0 {
		conds = append(conds, bson.M{"coin": bson.M{"$gt": args.Coinl, "$lt": args.Coinh}})
	}
	if len(args.Alipay) > 0 {
		conds = append(conds, bson.M{"alipayaccount": args.Alipay})
	}
	if args.Registerl > 0 || args.Registerh > 0 {
		conds = append(conds, bson.M{"createtime": bson.M{"$gt": time.Unix(int64(args.Registerl), 0), "$lt": time.Unix(int64(args.Registerh), 0)}})
	}
	if len(args.Channel) > 0 {
		conds = append(conds, bson.M{"channel": args.Channel})
	}
	selecter := bson.M{"$or": conds}
	err = cplayerdata.Find(selecter).All(&ret)
	if err != nil {
		logger.Logger.Error("svc.FindPlayerList is error", err)
		return nil
	}

	return
}

func (svc *PlayerDataSvc) UpdatePlayerElement(args *model.UpdateElement, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return fmt.Errorf("db may be close")
	}
	playerMap := make(map[string]interface{})
	json.Unmarshal([]byte(args.PlayerMap), &playerMap)

	var o model.PlayerData
	r_t := reflect.TypeOf(o)
	var bsonInfo bson.D
	var docelem bson.DocElem
	for k, v := range playerMap {
		if f, ok := r_t.FieldByName(k); ok {
			if reflect.TypeOf(v).ConvertibleTo(f.Type) {
				docelem.Name = strings.ToLower(k)
				docelem.Value = v
				bsonInfo = append(bsonInfo, docelem)
			} else {
				logger.Logger.Warnf("UpdatePlayerElement Type not fit %v field, get %v", k, v)
			}
		} else {
			logger.Logger.Warnf("UpdatePlayerElement no %v field", k)
		}
	}

	err = cplayerdata.Update(bson.M{"snid": args.SnId}, bson.D{{"$set", bsonInfo}})
	if err != nil {
		logger.Logger.Warnf("UpdatePlayerElement error:%v", err)
		return
	}

	*ret = true
	return
}

func (svc *PlayerDataSvc) GetPlayerBaseInfo(args *model.GetPlayerBaseInfoArgs, ret *model.PlayerBaseInfo) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	err = cplayerdata.Find(bson.M{"snid": args.SnId, "platform": &args.Plt}).One(ret)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Error("svc.GetPlayerBaseInfo is error: ", err)
		return
	}
	return
}

func (svc *PlayerDataSvc) SetBlackWhiteLevel(args *model.BlackWhiteLevel, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return nil
	}
	cond := bson.M{"snid": args.SnId}
	if args.Platform != "" {
		cond["platform"] = args.Platform
	}
	err = cplayerdata.Update(cond, bson.D{{"$set", bson.D{{"wblevel", args.WBLevel},
		{"wbcointotalin", args.WbCoinTotalIn},
		{"wbcointotalout", args.WbCoinTotalOut}, {"wbcoinlimit", args.WbCoinLimit},
		{"wbtime", args.T}, {"wbmaxnum", args.WbMaxNum}}}})
	if err != nil {
		logger.Logger.Info("SetBlackWhiteLevel error ", err)
		return
	}
	*ret = true
	return
}

func (svc *PlayerDataSvc) SetBlackWhiteLevelUnReset(args *model.BlackWhiteLevel, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return nil
	}
	cond := bson.M{"snid": args.SnId}
	err = cplayerdata.Update(cond, bson.D{{"$set", bson.D{{"wblevel", args.WBLevel},
		{"wbcoinlimit", args.WbCoinLimit},
		{"wbtime", args.T}, {"wbmaxnum", args.WbMaxNum}}}})
	if err != nil {
		logger.Logger.Info("SetBlackWhiteLevel error ", err)
		return
	}

	*ret = true
	return
}

func (svc *PlayerDataSvc) UpdateAllPlayerPackageTag(args *model.UpdatePackageId, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.PlatformStr)
	if cplayerdata == nil {
		return nil
	}

	info, err := cplayerdata.UpdateAll(bson.M{"packageid": args.PackageTag}, bson.D{{"$set",
		bson.D{{"platform", args.PlatformStr}, {"channel", args.ChannelStr},
			{"beunderagentcode", args.PromoterStr}, {"promotertree", args.PromoterTree}}}})
	if err != nil {
		logger.Logger.Info("svc.UpdateAllPlayerPackageTag error ", err)
		return
	}
	logger.Logger.Info("svc.UpdateAllPlayerPackageTag result:", info)
	*ret = true
	return
}

func (svc *PlayerDataSvc) SetVipLevel(args *model.SetPlayerAtt, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return nil
	}
	cond := bson.M{"snid": args.SnId}
	err = cplayerdata.Update(cond, bson.D{{"$set", bson.D{{"forcevip", args.VipLevel},
		{"vip", args.VipLevel}}}})
	if err != nil {
		logger.Logger.Info("SetVipLevel error ", err)
		return
	}
	*ret = true
	return
}

func (svc *PlayerDataSvc) SetGMLevel(args *model.SetPlayerAtt, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return nil
	}
	cond := bson.M{"snid": args.SnId}
	err = cplayerdata.Update(cond, bson.D{{"$set", bson.D{{"gmlevel", args.GmLevel}}}})
	if err != nil {
		logger.Logger.Info("SetGMLevel error ", err)
		return
	}
	*ret = true
	return
}
func (svc *PlayerDataSvc) GetSameIpPlayer(args *model.GetSameParamPlayerArgs, ret *[]int32) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}

	//临时结构
	var t []struct {
		SnId int32
	}
	err = cplayerdata.Find(bson.M{"ip": args.Param}).Select(bson.M{"snid": 1}).All(&t)
	if err != nil {
		logger.Logger.Error("svc.GetSameIpPlayer is error", err)
		return
	}
	for _, v := range t {
		*ret = append(*ret, v.SnId)
	}

	fmt.Println("GetSameIpPlayer....", ret)
	return
}
func (svc *PlayerDataSvc) GetSameBankNamePlayer(args *model.GetSameParamPlayerArgs, ret *[]int32) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	//临时结构
	var t []struct {
		SnId int32
	}
	err = cplayerdata.Find(bson.M{"bankaccname": args.Param}).Select(bson.M{"snid": 1}).All(&t)
	if err != nil {
		logger.Logger.Error("svc.GetSameBankNamePlayer is error", err)
		return
	}

	for _, value := range t {
		*ret = append(*ret, value.SnId)
	}
	return
}
func (svc *PlayerDataSvc) GetSameBankCardPlayer(args *model.GetSameParamPlayerArgs, ret *[]int32) (err error) {
	cplayerdata := PlayerDataCollection(args.Plt)
	if cplayerdata == nil {
		return
	}
	//临时结构
	var t []struct {
		SnId int32
	}
	err = cplayerdata.Find(bson.M{"bankaccount": args.Param}).Select(bson.M{"snid": 1}).All(&t)
	if err != nil {
		logger.Logger.Error("svc.GetSameBankCardPlayer is error", err)
		return
	}

	for _, value := range t {
		*ret = append(*ret, value.SnId)
	}
	return
}

func (svc *PlayerDataSvc) GetRobotPlayers(limit int, ret *[]*model.PlayerData) (err error) {
	cplayerdata := PlayerDataCollection(mongo.G_P)
	if cplayerdata == nil {
		return
	}
	data := make([]*model.PlayerData, 0, limit)
	err = cplayerdata.Find(bson.M{"isrob": true}).Limit(limit).All(&ret)
	if err != nil {
		logger.Logger.Error("svc.GetRobotPlayers is error", err)
		return
	}

	*ret = data
	return
}

/*
 * 保存玩家的删除备份全部信息
 */
func (svc *PlayerDelBackupDataSvc) SaveDelBackupPlayerData(pd *model.PlayerData, ret *bool) (err error) {
	tPlayerDelBackup := PlayerDelBackupDataCollection(pd.Platform)
	if tPlayerDelBackup == nil {
		return nil
	}
	if pd != nil {
		model.RecalcuPlayerCheckSum(pd)
		_, err = tPlayerDelBackup.UpsertId(pd.Id, pd)
		if err != nil {
			logger.Logger.Errorf("svc.SaveDelBackupPlayerData %v err:%v", pd.SnId, err)
			return
		}
		*ret = true
		return
	}
	return
}

func (svc *PlayerDataSvc) SetLogicLevel(args *model.LogicInfoArg, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return
	}
	cond := bson.M{"snid": bson.M{"$in": args.SnIds}}
	var info *mgo.ChangeInfo
	if args.LogicLevel == 0 { //清空所有分层信息
		info, err = cplayerdata.UpdateAll(cond, bson.D{{"$set", bson.D{{"logiclevels", []int32{}}}}})
	} else {
		info, err = cplayerdata.UpdateAll(cond, bson.D{{"$addToSet", bson.D{{"logiclevels", args.LogicLevel}}}})
	}
	if err != nil {
		logger.Logger.Error("SetLogicLevel error ", err)
		return
	}
	*ret = true
	logger.Logger.Tracef("SetLogicLevel UpdataAll:%#v", info)
	return
}

func (svc *PlayerDataSvc) ClrLogicLevel(args *model.LogicInfoArg, ret *bool) (err error) {
	cplayerdata := PlayerDataCollection(args.Platform)
	if cplayerdata == nil {
		return nil
	}
	cond := bson.M{"snid": bson.M{"$in": args.SnIds}}
	info, err := cplayerdata.UpdateAll(cond, bson.D{{"$pull", bson.D{{"logiclevels", args.LogicLevel}}}})
	if err != nil {
		logger.Logger.Error("ClrLogicLevel error ", err)
		return err
	}
	*ret = true
	logger.Logger.Tracef("ClrLogicLevel UpdataAll:%#v", info)
	return nil
}

func CorrectData(pd *model.PlayerData) bool {
	var coinTotal int64
	dirty := false
	//金币冲账
	coinlogs, err := GetCoinWALBySnidAndCoinTypeAndGreaterTs(pd.Platform, pd.SnId, PayCoinLogType_Coin, pd.CoinPayTs)
	if err == nil && len(coinlogs) != 0 {
		oldTs := pd.CoinPayTs
		oldCoin := pd.Coin
		var cnt int64

		for i := 0; i < len(coinlogs); i++ {
			cnt = coinlogs[i].Count
			pd.Coin += cnt
			coinTotal += cnt

			if coinlogs[i].LogType == common.GainWay_ShopBuy {
				pd.CoinPayTotal += cnt
				if pd.TodayGameData == nil {
					pd.TodayGameData = model.NewPlayerGameCtrlData()
				}
				pd.TodayGameData.RechargeCoin += cnt //累加当天充值金额
			}
			if coinlogs[i].Ts > pd.CoinPayTs {
				pd.CoinPayTs = coinlogs[i].Ts
			}
			dirty = true
		}

		newTs := pd.CoinPayTs
		newCoin := pd.Coin
		logger.Logger.Warnf("PlayerData(%v) CorrectData before:CoinPayTs=%v before:Coin=%v after:CoinPayTs=%v after:Coin=%v", pd.SnId, oldTs, oldCoin, newTs, newCoin)
	}
	////保险箱冲账
	//boxCoinLogs, err := GetCoinWALBySnidAndCoinTypeAndGreaterTs(pd.Platform, pd.SnId, PayCoinLogType_SafeBoxCoin, pd.SafeBoxCoinTs)
	//if err == nil && len(boxCoinLogs) != 0 {
	//	oldTs := pd.SafeBoxCoinTs
	//	oldCoin := pd.SafeBoxCoin
	//	for i := 0; i < len(boxCoinLogs); i++ {
	//		pd.SafeBoxCoin += boxCoinLogs[i].Count
	//		coinTotal += boxCoinLogs[i].Count
	//		if boxCoinLogs[i].LogType == common.GainWay_ShopBuy {
	//			pd.CoinPayTotal += boxCoinLogs[i].Count
	//			if pd.TodayGameData == nil {
	//				pd.TodayGameData = model.NewPlayerGameCtrlData()
	//			}
	//			pd.TodayGameData.RechargeCoin += boxCoinLogs[i].Count //累加当天充值金额
	//		}
	//		if boxCoinLogs[i].Ts > pd.SafeBoxCoinTs {
	//			pd.SafeBoxCoinTs = boxCoinLogs[i].Ts
	//		}
	//		dirty = true
	//	}
	//
	//	newTs := pd.SafeBoxCoinTs
	//	newCoin := pd.SafeBoxCoin
	//	logger.Logger.Warnf("PlayerData(%v) CorrectData before:SafeBoxCoinTs=%v before:SafeBoxCoin=%v after:SafeBoxCoinTs=%v after:SafeBoxCoin=%v", pd.SnId, oldTs, oldCoin, newTs, newCoin)
	//}
	//
	////比赛入场券冲账
	//ticketLogs, err := GetCoinWALBySnidAndCoinTypeAndGreaterTs(pd.Platform, pd.SnId, PayCoinLogType_Ticket, pd.TicketPayTs)
	//if err == nil && len(ticketLogs) != 0 {
	//	oldTs := pd.TicketPayTs
	//	oldTicket := pd.Ticket
	//	for i := 0; i < len(ticketLogs); i++ {
	//		pd.Ticket += ticketLogs[i].Count
	//		pd.TicketTotal += ticketLogs[i].Count
	//		if ticketLogs[i].Ts > pd.TicketPayTs {
	//			pd.TicketPayTs = ticketLogs[i].Ts
	//		}
	//		dirty = true
	//	}
	//
	//	newTs := pd.TicketPayTs
	//	newTicket := pd.Ticket
	//	logger.Logger.Warnf("PlayerData(%v) CorrectData before:TicketPayTs=%v before:Ticket=%v after:TicketPayTs=%v after:Ticket=%v", pd.SnId, oldTs, oldTicket, newTs, newTicket)
	//}

	//同步游服丢失的金币变化
	if SyncGameCoin(pd, 0, pd.GameCoinTs) {
		dirty = true
	}
	//确保金币不小于0
	if pd.Coin < 0 {
		logger.Logger.Warnf("PlayerData(%v) CorrectData found pd.Coin<0(%v)", pd.SnId, pd.Coin)
		if pd.SafeBoxCoin > 0 {
			pd.SafeBoxCoin += pd.Coin
			if pd.SafeBoxCoin < 0 {
				logger.Logger.Warnf("PlayerData(%v) CorrectData found pd.SafeBoxCoin<0(%v)", pd.SnId, pd.SafeBoxCoin)
				pd.SafeBoxCoin = 0
			}
		}
		pd.Coin = 0
		dirty = true
	}

	if dirty {
		SavePlayerData(pd)
	}

	return dirty
}

func SyncGameCoin(pd *model.PlayerData, sceneid int, enterts int64) bool {
	dirty := false
	//游服金币冲账
	endts := time.Now().UnixNano()
	gamecoinlogs, err := GetCoinWALBySnidAndInGameAndGreaterTs(pd.Platform, pd.SnId, int32(sceneid), enterts)
	if err == nil && len(gamecoinlogs) != 0 {
		oldCoin := pd.Coin
		var cnt int64
		for i := 0; i < len(gamecoinlogs); i++ {
			ts := gamecoinlogs[i].Ts
			if ts >= enterts && ts <= endts {
				cnt = gamecoinlogs[i].Count
				pd.Coin += cnt
				if ts > pd.GameCoinTs {
					pd.GameCoinTs = ts
				}
				dirty = true
			}
		}

		newCoin := pd.Coin
		newTs := pd.GameCoinTs
		logger.Logger.Warnf("PlayerData(%v) SyncGameCoin before:enterts=%v before:Coin=%v after:GameCoinTs=%v after:Coin=%v", pd.SnId, enterts, oldCoin, newTs, newCoin)
	}
	return dirty
}

func init() {
	rpc.Register(&PlayerDataSvc{})
	rpc.Register(&PlayerDelBackupDataSvc{})
}
