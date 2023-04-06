package main

import (
	"encoding/base64"
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	player_proto "games.yol.com/win88/protocol/player"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/i18n"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"

	"math/rand"
	"os"
)

const (
	PlayerDataFlag_NoInfo  = 1
	PlayerDataFlag_OneInfo = 2
	PlayerDataFlag_Relogin = 3
)

type TaskPlayerData struct {
	*player_proto.CSPlayerData
	*netlib.Session
	Sid         int64
	flag        player_proto.OpResultCode
	newbie      bool
	promoterCfg *PromoterConfig
	promoterID  string
	pi          *model.PlayerData
}

func (t *TaskPlayerData) Call(o *basic.Object) interface{} {
	pi, tf := model.GetPlayerData(t.PlatformTag, t.GetAccId())
	t.pi = pi
	if pi == nil {
		t.flag = player_proto.OpResultCode_OPRC_Login_CreateFailed
		return pi
	}
	t.newbie = tf //新玩家标记

	if pi.InviterId != 0 {
		invitePd := model.GetPlayerBaseInfo(pi.Platform, pi.InviterId)
		if invitePd != nil {
			t.promoterID = invitePd.BeUnderAgentCode
		}
	}

	t.flag = player_proto.OpResultCode_OPRC_Sucess
	return pi
}

func (t *TaskPlayerData) Done(data interface{}, tt *task.Task) {
	if t.flag == player_proto.OpResultCode_OPRC_Sucess {
		pi := t.pi
		tf := t.newbie
		key, err := GetPromoterKey(0, t.promoterID, "")
		if err == nil {
			t.promoterCfg = PromoterMgrSington.GetConfig(key)
		}
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			//新玩家 & inviterid是玩家id & 微信uinonid=""
			if tf && !pi.IsRob {
				//if pi.InviterId > 100000 {
				//	tag, err := webapi.API_PushSpreadLink(pi.SnId, pi.Platform, pi.PackageID, int(pi.InviterId), 1, 0, common.GetAppId())
				//	if err != nil {
				//		logger.Logger.Warnf("API_PushSpreadLink response tag:%v err:%v", tag, err)
				//	}
				//}
				//
				//var promoterTreeIsValid bool
				////无级代理推广员验证
				//if pi.PromoterTree != 0 {
				//	tag, msg := webapi.API_ValidPromoterTree(pi.SnId, pi.PackageID, pi.PromoterTree, common.GetAppId())
				//	if tag != 0 {
				//		logger.Logger.Warnf("API_ValidPromoterTree response tag:%v msg:%v", tag, msg)
				//		pi.PromoterTree = 0
				//		platform, _ := strconv.Atoi(pi.Platform)
				//		//先修改账号数据
				//		err := model.UpdateAccountPlatformInfo(pi.AccountId, int32(platform), 0, 0, pi.InviterId, pi.PromoterTree, pi.PackageID)
				//		if err != nil {
				//			logger.Logger.Warnf("UpdateAccountPlatformInfo err:%v", err)
				//		}
				//		err = model.UpdatePlayerPackageId(pi.SnId, pi.PackageID, int32(platform), 0, 0, pi.InviterId, pi.PromoterTree)
				//		if err != nil {
				//			logger.Logger.Warnf("UpdatePlayerPackageId err:%v", err)
				//		}
				//	} else {
				//		promoterTreeIsValid = true
				//	}
				//}
				//
				////推广关系获取
				//if !model.GameParamData.DonotGetPromoterByIp {
				//	tag, msg, err := webapi.API_PushInviterIp(pi.SnId, pi.InviterId, pi.PromoterTree, pi.SubBeUnderAgentCode,
				//		pi.PackageID, pi.Ip, pi.DeviceOS, common.GetAppId())
				//	if err == nil && tag == 0 {
				//		if msg != nil {
				//			if msg.Tag != "" {
				//				platform := strconv.Itoa(int(msg.Platform))
				//				if pi.Platform != platform {
				//					logger.Logger.Errorf("API_PushInviterIp [WARN!!!]: ip:%v snid:%v PackageID:%v Platform:%v get PackageID:%v Platform:%v", pi.Ip, pi.SnId, pi.PackageID, pi.Platform, msg.Tag, msg.Platform)
				//				} else {
				//					pi.PackageID = msg.Tag
				//					pi.Platform = platform                                  //平台
				//					pi.Channel = strconv.Itoa(int(msg.ChannelId))           //渠道
				//					pi.BeUnderAgentCode = strconv.Itoa(int(msg.PromoterId)) //推广员
				//					if pi.InviterId == 0 && msg.Spreader != 0 {
				//						pi.InviterId = msg.Spreader //全民代邀请人
				//					}
				//					if msg.PromoterTree != 0 && !promoterTreeIsValid {
				//						pi.PromoterTree = msg.PromoterTree //无级代上级
				//					}
				//
				//					if t.promoterCfg != nil {
				//						if t.promoterCfg.IsInviteRoot > 0 {
				//							pi.BeUnderAgentCode = t.promoterID
				//						}
				//					}
				//
				//					//先修改账号数据
				//					err = model.UpdateAccountPlatformInfo(pi.AccountId, msg.Platform, msg.ChannelId, msg.PromoterId, msg.Spreader, msg.PromoterTree, msg.Tag)
				//					if err != nil {
				//						logger.Logger.Warnf("UpdateAccountPlatformInfo err:%v", err)
				//					}
				//					err = model.UpdatePlayerPackageId(pi.SnId, msg.Tag, msg.Platform, msg.ChannelId, msg.PromoterId, msg.Spreader, msg.PromoterTree)
				//					if err != nil {
				//						logger.Logger.Warnf("UpdatePlayerPackageId err:%v", err)
				//					}
				//				}
				//			}
				//		}
				//	} else {
				//		logger.Logger.Warnf("API_PushInviterIp response tag:%v msg:%v err:%v snid:%v packagetag:%v ip:%v ", tag, msg, err, pi.SnId, pi.PackageID, pi.Ip)
				//	}
				//}
				//
				//isBind := 0
				//if pi.Tel != "" {
				//	isBind = 1
				//}
				////首次创建账号事件
				//if !model.GameParamData.IsWebApiUseMQ {
				//	tag, err := webapi.API_PlayerEvent(model.WEBEVENT_LOGIN, pi.Platform, pi.PackageID, pi.SnId, pi.Channel, pi.BeUnderAgentCode, pi.PromoterTree, 1, 1, isBind, common.GetAppId())
				//	if err != nil {
				//		logger.Logger.Warnf("API_PlayerEvent is error tag:%v err:%v", tag, err)
				//	}
				//}

			}
			return pi
		}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			if pbi, ok := data.(*model.PlayerData); ok {
				params := pbi.UpdateParams(t.GetParams())
				sid := PlayerMgrSington.EndPlayerLoading(t.GetAccId())
				if sid != 0 {
					t.Sid = sid
				}
				PlayerMgrSington.AddPlayer(t.Sid, pbi, t.Session)
				player := PlayerMgrSington.GetPlayer(t.Sid)
				if player != nil {

					var temp []byte
					var ee error
					di := t.GetDeviceInfo()
					if di != "" {
						var e common.Encryptor
						e.Init(common.GetAppId(), player.PackageID, int32(t.GetTimeStamp()))
						temp, ee = base64.StdEncoding.DecodeString(di)
						if ee == nil {
							e.Encrypt(temp, len(temp))
						}
					}

					//player.DeviceInfo = string(temp)
					player.DeviceId = t.GetDeviceId()
					player.params = params
					if t.newbie { //新用户赠送金币
						//首次创建账号事件
						isBind := 0
						if pbi.Tel != "" {
							isBind = 1
						}
						LogChannelSington.WriteMQData(model.GeneratePlayerEvent(model.WEBEVENT_LOGIN, pbi.Platform, pbi.PackageID, pbi.SnId, pbi.Channel, pbi.BeUnderAgentCode, pbi.PromoterTree, 1, 1, isBind, common.GetAppId()))

						newbieCoin := player.GetRegisterPrize()
						if newbieCoin > 0 {
							player.AddCoin(int64(newbieCoin), common.GainWay_NewPlayer, "system", "")
							//增加泥码
							player.AddDirtyCoin(0, int64(newbieCoin))
							player.ReportSystemGiveEvent(newbieCoin, common.GainWay_NewPlayer, true)
							player.AddPayCoinLog(int64(newbieCoin), model.PayCoinLogType_Coin, "NewPlayer")
						}
						if player.InviterId > 0 {
							//actRandCoinMgr.OnPlayerInvite(player.Platform, player.InviterId)
						}
					}
					if bi, ok := BlackListMgrSington.CheckLogin(pbi); !ok {
						t.flag = player_proto.OpResultCode_OPRC_InBlackList
						var msg string
						if bi != nil {
							msg = i18n.Tr("cn", "BlackListLimit2Args", pbi.SnId, bi.Id)
						} else {
							msg = i18n.Tr("cn", "BlackListLimit1Args", pbi.SnId)
						}
						common.SendSrvMsg(player, common.SRVMSG_CODE_DEFAULT, msg)
						//黑名单用户也需要调用一下onlogin,否则会导致数据无法刷新
						player.OnLogined()

					} else {
						if player.GMLevel > 2 {
							//伪造IP地址
							player.Ip = fmt.Sprintf("%v.%v.%v.%v", 1+rand.Int31n(255), 1+rand.Int31n(255), 1+rand.Int31n(255), 1+rand.Int31n(255))
						}
						player.dirty = true
						player.OnLogined()
						player.SendPlayerInfo()
					}
					return
				}
			}
		}), "PlayerRegister").Start()

	} else {
		scPlayerData := &player_proto.SCPlayerData{
			OpRetCode: t.flag,
		}
		proto.SetDefaults(scPlayerData)
		common.SendToGate(t.Sid, int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATA), scPlayerData, t.Session)
		logger.Logger.Info("Get playerinfo error:", scPlayerData)

	}
}

type SaveTask struct {
	pi   *model.PlayerData
	snid int32
}

func (t *SaveTask) Call(o *basic.Object) interface{} {
	if !model.SavePlayerData(t.pi) {
		//save 失败先写到json里面
		model.BackupPlayerData(t.pi)
		return false
	}
	return true
}

func (t *SaveTask) Done(data interface{}, tt task.Task) {
	p := PlayerMgrSington.GetPlayerBySnId(t.snid)
	if p != nil && p.CanDelete() {
		PlayerMgrSington.DelPlayer(p.SnId)
	}

	if ok, saved := data.(bool); ok && saved {
		bak := fmt.Sprintf("%v.json", t.pi.AccountId)
		if exist, _ := common.PathExists(bak); exist {
			os.Remove(bak)
		}
	}
}

type TaskChangeNick struct {
	*Player
	nick   string
	result player_proto.OpResultCode
}

func (t *TaskChangeNick) Call(o *basic.Object) interface{} {
	updateResult := model.UpdatePlayerNick(t.Platform, t.AccountId, t.nick)
	if updateResult == 0 {
		t.result = player_proto.OpResultCode_OPRC_Sucess

	} else {
		if updateResult == 1 {
			t.result = player_proto.OpResultCode_OPRC_Error
		} else {
			t.result = player_proto.OpResultCode_OPRC_Login_NameSame
		}
	}
	return nil
}
func (t *TaskChangeNick) Done(data interface{}, tt task.Task) {
	pack := &player_proto.SCChangeNick{
		Nick:      proto.String(t.nick),
		OpRetCode: t.result,
	}
	proto.SetDefaults(pack)
	t.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_CHANGENICK), pack)
	//	if t.result == player_proto.OpResultCode_OPRC_Sucess {
	//		PlayerMgrSington.UpdatePlayerName(t.Name, t.nick)
	//	}

	t.Player.dirty = true
}
