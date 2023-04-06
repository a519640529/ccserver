package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"time"
	"unicode/utf8"

	"net/url"
	"regexp"
	"strconv"

	"encoding/json"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	gamehall_proto "games.yol.com/win88/protocol/gamehall"
	player_proto "games.yol.com/win88/protocol/player"
	"games.yol.com/win88/srvdata"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/i18n"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

type CSPlayerDataPacketFactory struct {
}
type CSPlayerDataHandler struct {
}

func (this *CSPlayerDataPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSPlayerData{}
	return pack
}

var reTelRule, _ = regexp.Compile(`^(1[3|4|5|6|7|8|9][0-9]\d{4,8})$`)

func (this *CSPlayerDataHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPlayerDataHandler Process recv ", data)
	if cspl, ok := data.(*player_proto.CSPlayerData); ok {
		ls := LoginStateMgrSington.GetLoginStateOfSid(sid)
		if ls == nil {
			logger.Logger.Warnf("CSPlayerDataHandler sid:%v not login 1!!!", sid)
			return nil
		}
		if ls.als == nil {
			logger.Logger.Warnf("CSPlayerDataHandler sid:%v not login 2!!!", sid)
			return nil
		}
		if ls.als.acc == nil {
			logger.Logger.Warnf("CSPlayerDataHandler sid:%v not login 3!!!", sid)
			return nil
		}

		if cspl.GetAccId() != ls.als.acc.AccountId.Hex() {
			cspl.AccId = proto.String(ls.als.acc.AccountId.Hex())
			logger.Logger.Warnf("CSPlayerDataHandler player(%v) try to hit db!!!", cspl.GetAccId())
		}

		player := PlayerMgrSington.GetPlayerByAccount(cspl.GetAccId())
		if player != nil {
			//Send logout packet to client
			if player.state != PlayerState_Offline {
				if player.sid != 0 && player.sid != sid {
					//Kick the exist player disconnect
					player.Kickout(common.KickReason_OtherLogin)
				}
			}

			//给玩家发送三方余额状态
			statePack := &gamehall_proto.SCThridGameBalanceUpdateState{}
			if player.thridBalanceReqIsSucces {
				statePack.OpRetCode = gamehall_proto.OpResultCode_Game_OPRC_Sucess_Game
			} else {
				statePack.OpRetCode = gamehall_proto.OpResultCode_Game_OPRC_Error_Game
			}
			player.SendRawToClientIncOffLine(sid, s, int(gamehall_proto.GameHallPacketID_PACKET_SC_THRIDGAMEBALANCEUPDATESTATE), statePack)

			var temp []byte
			var ee error
			di := cspl.GetDeviceInfo()

			pt := PlatformMgrSington.GetPlatform(player.Platform)
			if !player.IsRob && pt == nil {
				scPlayerData := &player_proto.SCPlayerData{
					OpRetCode: player_proto.OpResultCode_OPRC_Error,
				}
				logger.Logger.Trace("player no platform:", player.Platform, player.SnId)
				proto.SetDefaults(scPlayerData)
				player.SendRawToClientIncOffLine(sid, s, int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATA), scPlayerData)
				//Kick the exist player disconnect
				player.Kickout(common.KickReason_Freeze)
				return nil
			}

			if !player.IsRob && pt.NeedDeviceInfo && di == "" { //maybe res ver is low
				scPlayerData := &player_proto.SCPlayerData{
					OpRetCode: player_proto.OpResultCode_OPRC_LoginFailed,
				}
				logger.Logger.Trace("player no deviceinfo:", player.SnId)

				proto.SetDefaults(scPlayerData)
				player.SendRawToClientIncOffLine(sid, s, int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATA), scPlayerData)
				//Kick the exist player disconnect
				player.Kickout(common.KickReason_Freeze)
				return nil
			}

			if !player.IsRob && di != "" && cspl.GetPlatformTag() != "" {
				var e common.Encryptor
				e.Init(common.GetAppId(), cspl.GetPlatformTag(), int32(cspl.GetTimeStamp()))
				temp, ee = base64.StdEncoding.DecodeString(di)
				if ee == nil {
					e.Encrypt(temp, len(temp))
				} else {
					scPlayerData := &player_proto.SCPlayerData{
						OpRetCode: player_proto.OpResultCode_OPRC_NotLogin,
					}
					logger.Logger.Trace("player no decode deviceinfo:", player.SnId)

					proto.SetDefaults(scPlayerData)
					player.SendRawToClientIncOffLine(sid, s, int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATA), scPlayerData)
					//Kick the exist player disconnect
					player.Kickout(common.KickReason_Freeze)
					return nil
				}

				if !player.IsRob && model.GameParamData.ValidDeviceInfo && !json.Valid(temp) {
					scPlayerData := &player_proto.SCPlayerData{
						OpRetCode: player_proto.OpResultCode_OPRC_NotLogin,
					}
					logger.Logger.Trace("player no json deviceinfo:", player.SnId, temp)
					proto.SetDefaults(scPlayerData)
					player.SendRawToClientIncOffLine(sid, s, int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATA), scPlayerData)
					//Kick the exist player disconnect
					player.Kickout(common.KickReason_Freeze)
					return nil
				}
			}

			//deviceInfo := string(temp)

			//rehold player
			PlayerMgrSington.ReholdPlayer(player, sid, s)
			player = PlayerMgrSington.GetPlayer(sid)
			if player != nil {
				player.params = player.UpdateParams(cspl.GetParams())
				//player.DeviceInfo = deviceInfo
				player.DeviceId = cspl.GetDeviceId()
				player.dirty = true
				if player.GMLevel > 2 {
					//伪造IP地址
					player.Ip = fmt.Sprintf("%v.%v.%v.%v", 1+rand.Int31n(255), 1+rand.Int31n(255), 1+rand.Int31n(255), 1+rand.Int31n(255))
				}
				if !player.IsRob {
					if bi, ok := BlackListMgrSington.CheckLogin(player.PlayerData); !ok {
						var msg string
						if bi != nil {
							msg = i18n.Tr("cn", "BlackListLimit2Args", player.SnId, bi.Id)
						} else {
							msg = i18n.Tr("cn", "BlackListLimit1Args", player.SnId)
						}
						common.SendSrvMsg(player, common.SRVMSG_CODE_DEFAULT, msg)
						scPlayerData := &player_proto.SCPlayerData{
							OpRetCode: player_proto.OpResultCode_OPRC_InBlackList,
						}
						proto.SetDefaults(scPlayerData)
						player.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATA), scPlayerData)
						//Kick the exist player disconnect
						player.Kickout(common.KickReason_Freeze)
						return nil
					}
				}
				player.SendPlayerInfo()
				player.OnRehold()
				//断线重连拉去玩家信息时，判断玩家是否在游戏内
				if player.scene != nil {
					player.SendGameConfig(int32(player.scene.gameId), player.Platform, player.Channel)
				}
			}
			return nil
		} else {
			if !PlayerMgrSington.StartLoading(cspl.GetAccId(), sid) {
				PlayerCacheMgrSington.UncacheInvalidPlayerId(ls.als.acc.SnId)
				PlayerCacheMgrSington.Get(ls.als.acc.Platform, ls.als.acc.SnId, func(pd *PlayerCacheItem, async, isnew bool) {
					send := func(code player_proto.OpResultCode) {
						scPlayerData := &player_proto.SCPlayerData{
							OpRetCode: code,
						}
						proto.SetDefaults(scPlayerData)
						common.SendToGate(sid, int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATA), scPlayerData, s)
					}

					sid := PlayerMgrSington.EndPlayerLoading(cspl.GetAccId())
					if pd == nil {
						send(player_proto.OpResultCode_OPRC_Login_CreateFailed)
						return
					}

					var promoterID string
					var promoterCfg *PromoterConfig
					f := func() {
						params := pd.UpdateParams(cspl.GetParams())
						PlayerMgrSington.AddPlayer(sid, pd.PlayerData, s)
						player := PlayerMgrSington.GetPlayer(sid)
						if player != nil {
							if promoterID != "" {
								key, err := GetPromoterKey(0, promoterID, "")
								if err == nil {
									promoterCfg = PromoterMgrSington.GetConfig(key)
									if promoterCfg != nil && promoterCfg.IsInviteRoot > 0 {
										player.BeUnderAgentCode = promoterID
										player.dirty = true
									}
								}
							}

							var temp []byte
							var ee error
							di := cspl.GetDeviceInfo()
							if di != "" {
								var e common.Encryptor
								e.Init(common.GetAppId(), player.PackageID, int32(cspl.GetTimeStamp()))
								temp, ee = base64.StdEncoding.DecodeString(di)
								if ee == nil {
									e.Encrypt(temp, len(temp))
								}
							}

							//player.DeviceInfo = string(temp)
							player.DeviceId = cspl.GetDeviceId()
							player.params = params
							if isnew { //新用户赠送金币
								//首次创建账号事件
								isBind := 0
								if pd.Tel != "" {
									isBind = 1
								}
								LogChannelSington.WriteMQData(model.GeneratePlayerEvent(model.WEBEVENT_LOGIN, pd.Platform, pd.PackageID, pd.SnId, pd.Channel, pd.BeUnderAgentCode, pd.PromoterTree, 1, 1, isBind, common.GetAppId()))

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
							if !player.IsRob {
								if bi, ok := BlackListMgrSington.CheckLogin(pd.PlayerData); !ok {
									var msg string
									if bi != nil {
										msg = i18n.Tr("cn", "BlackListLimit2Args", pd.SnId, bi.Id)
									} else {
										msg = i18n.Tr("cn", "BlackListLimit1Args", pd.SnId)
									}
									common.SendSrvMsg(player, common.SRVMSG_CODE_DEFAULT, msg)
									//黑名单用户也需要调用一下onlogin,否则会导致数据无法刷新
									player.OnLogined()
									send(player_proto.OpResultCode_OPRC_InBlackList)
									return
								}
							}

							if player.GMLevel > 2 {
								//伪造IP地址
								player.Ip = fmt.Sprintf("%v.%v.%v.%v", 1+rand.Int31n(255), 1+rand.Int31n(255), 1+rand.Int31n(255), 1+rand.Int31n(255))
							}
							player.dirty = true
							player.OnLogined()
							player.SendPlayerInfo()
							return
						}
					}
					if pd.InviterId != 0 {
						PlayerCacheMgrSington.Get(pd.Platform, pd.InviterId, func(inviter *PlayerCacheItem, async, isnew bool) {
							promoterID = inviter.BeUnderAgentCode
							f()
						}, false)
					} else {
						f()
					}
				}, true)
			}
		}
	}
	return nil
}

type CSQueryPlayerPacketFactory struct {
}
type CSQueryPlayerHandler struct {
}

func (this *CSQueryPlayerPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSQueryPlayer{}
	return pack
}

func (this *CSQueryPlayerHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSQueryPlayerHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSQueryPlayer); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSQueryPlayerHandler p == nil")
			return nil
		}
		destSnId := msg.GetSnId()
		send := func(f *Friend) {
			pack := &player_proto.SCQueryPlayer{
				SnId:    proto.Int32(f.SnId),
				Name:    proto.String(f.Name),
				Head:    proto.Int32(f.Head),
				Sex:     proto.Int32(f.Sex),
				Coin:    proto.Int64(f.Coin),
				Diamond: proto.Int64(f.Diamond),
			}
			//VCard
			item := BagMgrSington.GetBagItemById(f.SnId, VCard)
			if item != nil {
				pack.VCard = proto.Int64(int64(item.ItemNum))
			}
			//IsFriend
			pack.IsFriend = FriendMgrSington.IsFriend(p.SnId, f.SnId)
			//IsShield
			pack.IsShield = FriendMgrSington.IsShield(p.SnId, f.SnId)
			//Role
			roleInfo := f.Roles
			if roleInfo != nil {
				for id, level := range roleInfo {
					role := &player_proto.RoleOrPet{
						Id:    id,
						Level: level,
					}
					roleData := PetMgrSington.GetIntroductionByModId(id)
					if roleData != nil {
						role.Name = roleData.Name
					}
					pack.Roles = append(pack.Roles, role)
				}
			}
			//Pet
			petInfo := f.Pets
			if petInfo != nil {
				for id, level := range petInfo {
					pet := &player_proto.RoleOrPet{
						Id:    id,
						Level: level,
					}
					petData := PetMgrSington.GetIntroductionByModId(id)
					if petData != nil {
						pet.Name = petData.Name
					}
					pack.Pets = append(pack.Pets, pet)
				}
			}
			proto.SetDefaults(pack)
			ok := p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_QUERYPLAYER), pack)
			logger.Logger.Trace("SCQueryPlayerHandler ok: ", ok, " pack: ", pack)
		}

		player := PlayerMgrSington.GetPlayerBySnId(destSnId)
		if player != nil && player.IsRob {
			pack := &player_proto.SCQueryPlayer{
				SnId:     proto.Int32(player.SnId),
				Name:     proto.String(player.Name),
				Head:     proto.Int32(player.Head),
				Sex:      proto.Int32(player.Sex),
				Coin:     proto.Int64(player.Coin),
				Diamond:  proto.Int64(player.Diamond),
				IsFriend: proto.Bool(false),
				VCard:    proto.Int64(0),
				IsShield: proto.Bool(false),
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_QUERYPLAYER), pack)
			logger.Logger.Trace("SCQueryPlayerHandler ok: ", ok, " pack: ", pack)
		} else { // 真人去friend取信息
			friend := FriendMgrSington.GetFriendBySnid(destSnId)
			if friend != nil {
				//修正下金币钻石当前数据
				destPlayer := PlayerMgrSington.GetPlayerBySnId(destSnId)
				if destPlayer != nil {
					if destPlayer.Coin != friend.Coin {
						friend.Coin = destPlayer.Coin
					}
					if destPlayer.Diamond != friend.Diamond {
						friend.Diamond = destPlayer.Diamond
					}
				}
				send(friend)
			} else {
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					ret, err := model.QueryFriendBySnid(p.Platform, destSnId)
					if err != nil {
						return nil
					}
					return ret
				}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
					if data != nil {
						ret := data.(*model.Friend)
						if ret != nil {
							f := FriendMgrSington.ExchangeModelFriend2Cache(ret)
							send(f)
						}
					}
				})).StartByFixExecutor("GetFriendBySnid")
			}
		}
	}
	return nil
}

func _formatImageCodeKey(tel string, tagkey int32) string {
	return fmt.Sprintf("%v_%v_imagecode", tel, tagkey)
}
func _formatGeeCodeKey(tel, tag string, tagkey int32) string {
	return fmt.Sprintf("%v_%v_%v_geecode", tel, tag, tagkey)
}

// 获取图片验证码
type CSGetImageVerifyCodePacketFactory struct {
}
type CCSGetImageVerifyCodeHandler struct {
}

func (this *CSGetImageVerifyCodePacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSGetImgageVerifyCode{}
	return pack
}

func (this *CCSGetImageVerifyCodeHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CCSGetImageVerifyCodeHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSGetImgageVerifyCode); ok {
		tel := msg.GetTel()
		sendPack := func(code player_proto.OpResultCode, data string) {
			pack := &player_proto.SCGetImgageVerifyCode{
				OpRetCode: code,
				ImageData: proto.String(data),
			}
			proto.SetDefaults(pack)
			common.SendToGate(sid, int(player_proto.PlayerPacketID_PACKET_SC_GETIMAGEVERIFYCODE), pack, s)
		}

		if !reTelRule.MatchString(tel) {
			sendPack(player_proto.OpResultCode_OPRC_TelError, "")
			return nil
		}

		//替换为后台拿到的配置数据，不再使用客户端的发送数据
		_, _, _, _, tagkey := PlatformMgrSington.GetPlatformByPackageTag(msg.GetPlatformTag())

		var err error
		var imgVerifyMsg *webapi.ImgVerifyMsg
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			imgVerifyMsg, err = webapi.API_GetImgVerify(common.GetAppId(), tel)
			return err
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil && imgVerifyMsg != nil {
				key := _formatImageCodeKey(tel, tagkey)
				CacheMemory.Put(key, imgVerifyMsg.Code, 60)
				logger.Logger.Infof("@CacheMemory.Put(key:%v, val:%v)", key, imgVerifyMsg.Code)
				sendPack(player_proto.OpResultCode_OPRC_Sucess, imgVerifyMsg.ImgBase)
			} else {
				sendPack(player_proto.OpResultCode_OPRC_Error, "")
			}
		}), "API_GetImgVerify").Start()
	}
	return nil
}

// 获取验证码
type CSVerificationCodePlayerPacketFactory struct {
}
type CSVerificationCodePlayerHandler struct {
}

func (this *CSVerificationCodePlayerPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSPlayerVerificationCode{}
	return pack
}

func (this *CSVerificationCodePlayerHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSVerificationCodePlayerHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSPlayerVerificationCode); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSVerificationCodePlayerHandler p == nil", sid)
		}
		tel := msg.GetTel()
		//替换为后台拿到的配置数据，不再使用客户端的发送数据
		pf, _, _, _, tagkey := PlatformMgrSington.GetPlatformByPackageTag(msg.GetPlatformTag())
		platform := strconv.Itoa(int(pf))

		sendPack := func(code player_proto.OpResultCode) {
			pack := &player_proto.SCPlayerVerificationCode{
				OpRetCode:        code,
				VerificationCode: proto.Int32(0),
			}
			proto.SetDefaults(pack)
			common.SendToGate(sid, int(player_proto.PlayerPacketID_PACKET_SC_VERIFICATIONCODE), pack, s)
		}
		if platform == "" {
			logger.Logger.Error("platform is nil")
			return nil
		}

		if !reTelRule.MatchString(tel) {
			sendPack(player_proto.OpResultCode_OPRC_TelError)
			return nil
		}
		plt := PlatformMgrSington.GetPlatform(platform)
		if plt == nil {
			logger.Logger.Error("platform is not exist.")
			return nil
		}
		if plt.VerifyCodeType == common.CodeTypeStr {
			key := _formatImageCodeKey(msg.GetTel(), tagkey)
			imgcode := CacheMemory.Get(key)
			if imgcode != msg.GetImgCode() {
				logger.Logger.Warnf("@CacheMemory.Get(key:%v) ImageCode not fit!get %v expect %v", tel, msg.GetImgCode(), imgcode)
				sendPack(player_proto.OpResultCode_OPRC_ImageVerifyCodeFailed)
				return nil
			}
		} else if plt.VerifyCodeType == common.CodeTypeHuaKuai {
			key := _formatGeeCodeKey(tel, msg.GetPlatformTag(), tagkey)
			geecode := CacheMemory.Get(key)
			if code, ok := geecode.(bool); ok && code != true {
				logger.Logger.Warnf("@CacheMemory.Get(key:%v) GeeCode not fit! expect %v, but not", key, geecode)
				sendPack(player_proto.OpResultCode_OPRC_ImageVerifyCodeFailed)
				return nil
			}
		}
		code := CacheMemory.Get(fmt.Sprintf("%v_%v", msg.GetTel(), tagkey))
		if code != nil {
			sendPack(player_proto.OpResultCode_OPRC_Frequently)
			return nil
		}

		var ee error
		var temp []byte
		di := msg.GetDeviceInfo()
		if plt.NeedDeviceInfo && di == "" { //maybe res ver is low
			sendPack(player_proto.OpResultCode_OPRC_YourResVerIsLow)
			return nil
		}

		if di != "" {
			var e common.Encryptor
			e.Init(common.GetAppId(), msg.GetPlatformTag(), int32(msg.GetTs()))
			temp, ee = base64.StdEncoding.DecodeString(di)
			if ee != nil {
				sendPack(player_proto.OpResultCode_OPRC_YourResVerIsLow)
				return nil
			}
			e.Encrypt(temp, len(temp))

			if model.GameParamData.ValidDeviceInfo && !json.Valid(temp) {
				sendPack(player_proto.OpResultCode_OPRC_YourResVerIsLow)
				return nil
			}

			//TODO 校验设备相关黑名单
			deviceInfo := string(temp)
			if bi, ok := BlackListMgrSington.CheckDeviceInBlack(deviceInfo, BlackState_Login, platform); ok {
				sendPack(player_proto.OpResultCode_OPRC_InBlackList)
				if bi != nil {
					common.SendToGate(sid, int(player_proto.PlayerPacketID_PACKET_SC_SRVMSG), common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, i18n.Tr("cn", "BlackListLimit1Args", bi.Id)), s)
				}
				return nil
			}
		}

		telkey := fmt.Sprintf("%v_%v", msg.GetTel(), tagkey)
		if model.GameParamData.FakeVerifyCode != "" {
			//TODO test tel code
			CacheMemory.Put(telkey, model.GameParamData.FakeVerifyCode, 60)
			sendPack(player_proto.OpResultCode_OPRC_Sucess)
			//TODO test tel code
			return nil
		}

		vcodestr := common.RandSmsCode()
		//先设置注册码，防止多次注册
		CacheMemory.Put(telkey, vcodestr, 60)
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			snid := int32(0)
			errstr := ""
			if p == nil {
				//判断是否是注册
				opCode := msg.GetOpCode()
				if opCode != 1 {
					snid, errstr = model.GetPlayerTel(tel, platform, tagkey)
					if snid == 0 {
						return errstr
					}
				}
			} else {
				snid = p.SnId
				platform = p.Platform
			}
			return webapi.API_SendSms(common.GetAppId(), snid, tel, vcodestr, platform)
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				CacheMemory.Put(telkey, vcodestr, 60)
				sendPack(player_proto.OpResultCode_OPRC_Sucess)
			} else {
				logger.Logger.Info("API_SendSms ", msg.GetTel(), data)
				sendPack(player_proto.OpResultCode_OPRC_Error)
			}
		}), "API_SendSms").Start()
	}
	return nil
}

// 获取邀请码奖励
type CSInviteCodePlayerPacketFactory struct {
}
type CSInviteCodePlayerHandler struct {
}

func (this *CSInviteCodePlayerPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSPlayerInviteCode{}
	return pack
}

func (this *CSInviteCodePlayerHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSInviteCodePlayerHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSPlayerInviteCode); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSInviteCodePlayerHandler p == nil")
			return nil
		}
		inviteCode := msg.GetCode()
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {

			var ret webapi.RequestBody
			var err error
			data := make(url.Values)
			data["MemberId"] = []string{strconv.Itoa(int(p.SnId))}
			data["Key"] = []string{inviteCode}

			ret, err = webapi.API_OP("/api/ExchangeCode/MemberExchangeCode", data)
			logger.Logger.Trace("webapi返回数据: ", ret)
			if err != nil {
				return nil
			}
			if ret == nil {
				return nil
			}

			state, _ := ret.GetInt("State")
			if state == 0 {
				return nil
			}

			ret_data, _ := ret.GetRequestBody("Data")
			if ret_data == nil {
				return nil
			}
			coin, _ := ret_data.GetInt64("Price")
			if coin <= 0 {
				return nil
			}

			//发送邮件
			var otherParams []int32
			newMsg := model.NewMessage("", p.SnId, "系统", p.SnId, model.MSGTYPE_INVITECODE, p.Name, inviteCode, coin, 0,
				model.MSGSTATE_UNREAD, time.Now().Unix(), 0, "", otherParams, p.Platform, 0) // 0 代表主大厅
			err = model.InsertMessage(p.Platform, newMsg)
			if err == nil {

				//执行成功后通知后端
				data := make(url.Values)
				data["Snid"] = []string{strconv.Itoa(int(p.SnId))}
				data["Code"] = []string{inviteCode}
				data["State"] = []string{strconv.Itoa(1)}
				_, err_res := webapi.API_OP("/api/ExchangeCode/MemberExchangeCodeSuccess", data)
				if err_res != nil {
					logger.Logger.Error("/api/ExchangeCode/MemberExchangeCodeSuccess err ", err_res)
				}

				return newMsg
			} else {
				//执行失败后通知后端
				data := make(url.Values)
				data["Snid"] = []string{strconv.Itoa(int(p.SnId))}
				data["Code"] = []string{inviteCode}
				data["State"] = []string{strconv.Itoa(0)}
				_, err_res := webapi.API_OP("/api/ExchangeCode/MemberExchangeCodeSuccess", data)
				if err_res != nil {
					logger.Logger.Error("/api/ExchangeCode/MemberExchangeCodeSuccess err ", err_res)
				}

				return nil
			}
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			retCode := player_proto.OpResultCode_OPRC_Sucess
			if newMsg, OK := data.(*model.Message); OK {
				p.AddMessage(newMsg)
			} else {
				retCode = player_proto.OpResultCode_OPRC_Error
			}

			//错误提示
			pack := &player_proto.SCPlayerInviteCode{
				OpRetCode: retCode,
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_INVITECODE), pack)

		}), "InviteCode").Start()

	}
	return nil
}

// 修改昵称
type CSPlayerChangeNickPacketFactory struct {
}
type CSPlayerChangeNickHandler struct {
}

func (this *CSPlayerChangeNickPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSChangeNick{}
	return pack
}

func (this *CSPlayerChangeNickHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	//logger.Logger.Trace("CSPlayerChangeNickHandler Process recv ", data)
	//if msg, ok := data.(*player_proto.CSChangeNick); ok {
	//	//logger.Logger.Trace("修改昵称，获得IP",s.RemoteAddr())
	//	p := PlayerMgrSington.GetPlayer(sid)
	//	if p == nil {
	//		logger.Logger.Warn("CSPlayerChangeNickHandler p == nil")
	//		return nil
	//	}
	//
	//	sendPack := func(code *player_proto.OpResultCode) {
	//		pack := &player_proto.SCChangeNick{
	//			OpRetCode: code,
	//			Nick:      msg.Nick,
	//		}
	//		proto.SetDefaults(pack)
	//		p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_CHANGENICK), pack)
	//	}
	//
	//	nick := msg.GetNick()
	//	//昵称不能为空
	//	if nick == "" {
	//		sendPack(player_proto.OpResultCode_OPRC_NickIsNull)
	//		return nil
	//	}
	//
	//	if !utf8.ValidString(nick) {
	//		sendPack(player_proto.OpResultCode_OPRC_NickIsIllegal)
	//		return nil
	//	}
	//	//昵称已占用，昵称允许重复
	//	//if model.GetPlayerNickIsExist(nick) {
	//	//	sendPack(player_proto.OpResultCode_OPRC_NickIsExist)
	//	//	return nil
	//	//}
	//
	//	//昵称超出长度限制
	//	if len(nick) > 21 {
	//		sendPack(player_proto.OpResultCode_OPRC_NickIsTooLen)
	//		return nil
	//	}
	//
	//	if p.IsStopRename == 1 {
	//		sendPack(player_proto.OpResultCode_OPRC_NickIsCantRename)
	//		return nil
	//	}
	//
	//	//TODO 校验昵称的合法性
	//	//var agentName, agentKey, thirdPlf = model.GetDgConfigByPlatform(p.Platform)
	//	var agentDgName, agentDgKey, thirdDgPlf = model.OnlyGetDgConfigByPlatform(p.Platform)
	//	var agentHboName, agentHboKey, thirdHboPlf = model.OnlyGetHboConfigByPlatform(p.Platform)
	//	//为了兼容以前的数据
	//	var hboGame, hboPass, dgGame, dgPass string
	//	if len(p.StoreHboGame) > 0 {
	//		hboGame = p.StoreHboGame
	//		hboPass = p.StoreHboPass
	//	}
	//	if len(p.StoreDgGame) > 0 {
	//		dgGame = p.StoreDgGame
	//		dgPass = p.StoreDgPass
	//	}
	//	if len(p.DgGame) > 0 && strings.Contains(p.DgGame, "dg") {
	//		dgGame = p.DgGame
	//		dgPass = p.DgPass
	//	}
	//	if len(p.DgGame) > 0 && strings.Contains(p.DgGame, "hbo") {
	//		hboGame = p.DgGame
	//		hboPass = p.DgPass
	//	}
	//	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//		code := model.UpdatePlayerNick(p.AccountId, msg.GetNick())
	//		if code == 0 {
	//
	//			if len(hboGame) > 0 && len(agentHboName) > 0 {
	//				err, codeid, _, _ := webapi.API_DgUpdate(thirdHboPlf, common.GetAppId(),
	//					hboGame, hboPass, nick, agentHboName, agentHboKey)
	//				if err != nil {
	//					logger.Logger.Error("Update hbo name error:", err)
	//				}
	//				if codeid != 0 {
	//					logger.Logger.Error("Update hbo code:", codeid)
	//				}
	//			}
	//
	//			if len(dgGame) > 0 && len(agentDgName) > 0 {
	//				err, codeid, _, _ := webapi.API_DgUpdate(thirdDgPlf, common.GetAppId(),
	//					dgGame, dgPass, nick, agentDgName, agentDgKey)
	//				if err != nil {
	//					logger.Logger.Error("Update dg name error:", err)
	//				}
	//				if codeid != 0 {
	//					logger.Logger.Error("Update dg code:", codeid)
	//				}
	//			}
	//
	//		}
	//		return code
	//	}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
	//		if i, ok := data.(int); ok {
	//			if i == 0 {
	//				PlayerSubjectSign.UpdateName(p.SnId, nick)
	//				sendPack(player_proto.OpResultCode_OPRC_Sucess)
	//			} else {
	//				sendPack(player_proto.OpResultCode_OPRC_Error)
	//			}
	//		}
	//	}), "UpdatePlayerNick").Start()
	//}
	return nil
}

// 修改头像
type CSPlayerChangeIconPacketFactory struct {
}
type CSPlayerChangeIconHandler struct {
}

func (this *CSPlayerChangeIconPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSPlayerChangeIcon{}
	return pack
}

func (this *CSPlayerChangeIconHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPlayerChangeIconHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSPlayerChangeIcon); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSPlayerChangeIconHandler p == nil")
			return nil
		}

		sendPack := func(code player_proto.OpResultCode) {
			pack := &player_proto.SCPlayerChangeIcon{
				OpRetCode: code,
				Icon:      msg.Icon,
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_CHANGEICON), pack)
		}

		if int32(time.Now().Sub(p.changeIconTime)/time.Second) > 30 {
			PlayerSubjectSign.UpdateHead(p.SnId, msg.GetIcon())
			sendPack(player_proto.OpResultCode_OPRC_Sucess)
		} else {
			sendPack(player_proto.OpResultCode_OPRC_Frequently)
		}
	}
	return nil
}

// 修改头像框
type CSPlayerChangeHeadOutLinePacketFactory struct {
}
type CSPlayerChangeHeadOutLineHandler struct {
}

func (this *CSPlayerChangeHeadOutLinePacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSPlayerChangeHeadOutLine{}
	return pack
}

func (this *CSPlayerChangeHeadOutLineHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPlayerChangeHeadOutLineHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSPlayerChangeHeadOutLine); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSPlayerChangeHeadOutLineHandler p == nil")
			return nil
		}

		sendPack := func(code player_proto.OpResultCode) {
			pack := &player_proto.SCPlayerChangeHeadOutLine{
				OpRetCode:   code,
				HeadOutLine: msg.HeadOutLine,
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_HEADOUTLINE), pack)
		}

		headOutline := msg.GetHeadOutLine()
		can := false
		for vip := int32(0); vip <= p.VIP; vip++ {
			dbVip := srvdata.PBDB_VIPMgr.GetData(vip)
			if dbVip != nil {
				if common.InSliceInt32(dbVip.GetRewardOutlineID(), headOutline) {
					can = true
					break
				}
			}
		}
		if !can {
			sendPack(player_proto.OpResultCode_OPRC_IconError)
			return nil
		}
		if msg.GetHeadOutLine() != 0 {
			PlayerSubjectSign.UpdateHeadOutline(p.SnId, headOutline)
			sendPack(player_proto.OpResultCode_OPRC_Sucess)
		} else {
			sendPack(player_proto.OpResultCode_OPRC_Frequently)
		}
	}
	return nil
}

// 修改性别
type CSPlayerChangeSexPacketFactory struct {
}
type CSPlayerChangeSexHandler struct {
}

func (this *CSPlayerChangeSexPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSPlayerChangeSex{}
	return pack
}

func (this *CSPlayerChangeSexHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPlayerChangeSexHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSPlayerChangeSex); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSPlayerChangeSexHandler p == nil")
			return nil
		}

		sendPack := func(code player_proto.OpResultCode) {
			pack := &player_proto.SCPlayerChangeSex{
				OpRetCode: code,
				Sex:       msg.Sex,
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_CHANGESEX), pack)
		}
		if msg.GetSex() < 0 || msg.GetSex() > 2 {
			sendPack(player_proto.OpResultCode_OPRC_SexError)
			return nil
		}

		p.Sex = msg.GetSex()
		p.dirty = true
		sendPack(player_proto.OpResultCode_OPRC_Sucess)
	}
	return nil
}

// 帐号绑定，帐号密码找回，保险箱密码找回
type CSUpgradeAccountPacketFactory struct {
}
type CSUpgradeAccountHandler struct {
}

func (this *CSUpgradeAccountPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSUpgradeAccount{}
	return pack
}

func (this *CSUpgradeAccountHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSUpgradeAccountHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSUpgradeAccount); ok {
		p := PlayerMgrSington.GetPlayer(sid)

		sendPack := func(code player_proto.OpResultCode) {
			pack := &player_proto.SCUpgradeAccount{
				OpRetCode:  code,
				Tel:        msg.Tel,
				ChangeType: msg.ChangeType,
			}
			proto.SetDefaults(pack)
			//p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_UPGRADEACCOUNT), pack)
			common.SendToGate(sid, int(player_proto.PlayerPacketID_PACKET_SC_UPGRADEACCOUNT), pack, s)
		}

		//验证手机号
		tel := msg.GetTel()
		if !reTelRule.MatchString(tel) {
			sendPack(player_proto.OpResultCode_OPRC_TelError)
			return nil
		}
		platformId, _, _, _, tagkey := PlatformMgrSington.GetPlatformByPackageTag(msg.GetPlatformTag())
		platformID := strconv.Itoa(int(platformId))
		platform := PlatformMgrSington.GetPlatform(platformID)
		if msg.GetChangeType() == 0 {
			if !platform.RegisterVerifyCodeSwitch {
				code := CacheMemory.Get(fmt.Sprintf("%v_%v", tel, tagkey))
				if code == nil {
					sendPack(player_proto.OpResultCode_OPRC_VerificationCodeError)
					return nil
				}
				if verifyCode, ok := code.(string); ok && verifyCode != "" {
					if verifyCode != msg.GetVerificationCode() {
						sendPack(player_proto.OpResultCode_OPRC_VerificationCodeError)
						return nil
					}
				}
			}
		} else {
			if platform.VerifyCodeType != common.CodeTypeNo {
				code := CacheMemory.Get(fmt.Sprintf("%v_%v", tel, tagkey))
				if code == nil {
					sendPack(player_proto.OpResultCode_OPRC_VerificationCodeError)
					return nil
				}
				if verifyCode, ok := code.(string); ok && verifyCode != "" {
					if verifyCode != msg.GetVerificationCode() {
						sendPack(player_proto.OpResultCode_OPRC_VerificationCodeError)
						return nil
					}
				}
			}
		}

		//验证密码
		if msg.GetPassword() == "" || len(msg.GetPassword()) < 32 {
			sendPack(player_proto.OpResultCode_OPRC_UpgradeAccount_PasswordIllegal)
			return nil
		}

		var upgradeaccountcount int32
		//升级账号
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			//绑定手机
			if msg.GetChangeType() == 0 {
				if p != nil {
					//手机号是否已经注册
					if model.PlayerTelIsExist(msg.GetTel(), platformID, tagkey) {
						sendPack(player_proto.OpResultCode_OPRC_TelIsExist)
						return errors.New("TelIsExist is Error")
					}
					//帐号绑定
					//更新Account数据
					err := model.UpgradeAccount(p.AccountId, msg.GetTel(), msg.GetPassword(), platformID, tagkey)
					if err != nil {
						return err
					}

					//当前IP和日期已经赠送的金币总额
					upgradeaccountcount = model.GetUpgradeAccountCoinLogsByIPAndDate(p.RegIp, p.Platform, time.Now())
					//更新玩家信息表
					return model.UpdatePlayerTel(p.Platform, p.SnId, msg.GetTel())
				} else {
					//账号密码找回
					return model.GetBackPlayerPassword(msg.GetTel(), msg.GetPassword(), platformID, tagkey)
				}
			}
			//设置保险箱密码
			/*if msg.GetChangeType() == 1 {
				raw := fmt.Sprintf("111111%v", common.GetAppId())
				h := md5.New()
				io.WriteString(h, raw)
				pwd := hex.EncodeToString(h.Sum(nil))
				return model.UpdateSafeBoxPassword(p.AccountId, pwd, msg.GetPassword())
			}*/
			//帐号密码找回
			if msg.GetChangeType() == 2 {
				//账号密码找回
				return model.GetBackPlayerPassword(msg.GetTel(), msg.GetPassword(), platformID, tagkey)
			}
			//保险箱密码找回
			if msg.GetChangeType() == 3 {
				if p != nil {
					p.SafeBoxPassword = msg.GetPassword()
				}
				return model.GetBackSafeBoxPassword(msg.GetTel(), msg.GetPassword(), platformID, tagkey)
			}
			return fmt.Errorf("GetChangeType is Error:%v", msg.GetChangeType())
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				if msg.GetChangeType() == 0 || msg.GetChangeType() == 2 {
					als := LoginStateMgrSington.GetLoginStateByTelAndPlatform(msg.GetTel(), platformID)
					if als != nil && als.acc != nil {
						als.acc.Tel = msg.GetTel()
						als.acc.TelPassWord = msg.GetPassword()
					}
				}
				if p != nil {
					oldTel := p.Tel
					p.Tel = msg.GetTel()
					if msg.GetChangeType() == 0 {
						if !p.IsRob {
							//用户绑定事件
							if oldTel == "" { //首次绑定账号
								p.UpgradeTime = time.Now()
								p.ReportBindPhoneEvent()
								//p.ReportWebEvent(model.WEBEVENT_UPGRADEACC, 0, 0, 1)

								//升级账号赠送
								upgradeAccountGiveCoin := p.GetUpdateAccPrize()
								if upgradeAccountGiveCoin > 0 && !p.layered[common.ActId_UpgradeAccount] {
									if model.GameParamData.UpgradeAccountGiveCoinLimit != 0 && upgradeaccountcount < model.GameParamData.UpgradeAccountGiveCoinLimit {
										p.AddCoin(int64(upgradeAccountGiveCoin), common.GainWay_UpgradeAccount, "", "")
										//增加泥码
										p.AddDirtyCoin(0, int64(upgradeAccountGiveCoin))
										p.AddPayCoinLog(int64(upgradeAccountGiveCoin), model.PayCoinLogType_Coin, "system")
										p.ReportSystemGiveEvent(upgradeAccountGiveCoin, common.GainWay_UpgradeAccount, true)
									} else {
										upgradeAccountGiveCoin = 0
										sendPack(player_proto.OpResultCode_OPRC_Account_IP_TooManyReg)
									}
								}

								//记录赠送日志
								task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
									model.InsertUpgradeAccountCoinLog(p.RegIp, time.Now(), upgradeAccountGiveCoin, p.SnId, p.Channel, p.Platform, p.BeUnderAgentCode, p.PackageID, p.City, p.InviterId)
									return nil
								}), nil, "InsertUpgradeAccountCoinLog").StartByFixExecutor("InsertUpgradeAccountCoinLog")
								//正式用户的登录红包
								//actRandCoinMgr.OnRandOnlineCoin(p)
							}
						}
					}
				}
				sendPack(player_proto.OpResultCode_OPRC_Sucess)
			} else {
				logger.Logger.Warnf("UpgradeAccount err:%v", data)
				sendPack(player_proto.OpResultCode_OPRC_Error)
			}
		}), "UpgradeAccount").StartByExecutor(msg.GetTel())
	}

	return nil
}

// 绑定支付宝
type CSBindAlipayPacketFactory struct {
}
type CSBindAlipayHandler struct {
}

func (this *CSBindAlipayPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSBindAlipay{}
	return pack
}

func (this *CSBindAlipayHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSBindAlipayHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSBindAlipay); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSBindAlipayHandler p == nil")
			return nil
		}

		sendPack := func(code player_proto.OpResultCode) {
			pack := &player_proto.SCBindAlipay{
				OpRetCode:     code,
				AlipayAccount: msg.AlipayAccount,
				AlipayAccName: msg.AlipayAccName,
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_BINDALIPAY), pack)
		}
		platform := PlatformMgrSington.GetPlatform(p.Platform)
		if platform != nil {
			if !platform.IsMarkFlag(int32(Bind_Alipay)) {
				sendPack(player_proto.OpResultCode_OPRC_BindAlipay_PlatformError)
				return nil
			}
		}
		alipayAccount := msg.GetAlipayAccount()
		if alipayAccount == "" {
			sendPack(player_proto.OpResultCode_OPRC_BindAlipay_AccountEmpty)
			return nil
		}

		if !utf8.ValidString(alipayAccount) {
			sendPack(player_proto.OpResultCode_OPRC_BindAlipay_AccountIllegal)
			return nil
		}

		alipayAccName := msg.GetAlipayAccName()
		if alipayAccName == "" {
			sendPack(player_proto.OpResultCode_OPRC_BindAlipay_AccNameEmpty)
			return nil
		}

		if !utf8.ValidString(alipayAccName) {
			sendPack(player_proto.OpResultCode_OPRC_BindAlipay_AccNameIllegal)
			return nil
		}

		if platform.NeedSameName {
			if p.BankAccName != "" {
				if p.BankAccName != alipayAccName {
					sendPack(player_proto.OpResultCode_OPRC_BankAndAli_NotSame)
					return nil
				}
			}
		}

		if msg.GetPassword() == "" || len(msg.GetPassword()) < 32 {
			sendPack(player_proto.OpResultCode_OPRC_Safebox_PasswordIllegal)
			return nil
		}

		raw := fmt.Sprintf("%v%v%v", p.SafeBoxPassword, common.GetAppId(), msg.GetTimeStamp())
		h := md5.New()
		io.WriteString(h, raw)
		hashsum := hex.EncodeToString(h.Sum(nil))
		if hashsum != msg.GetPassword() {
			sendPack(player_proto.OpResultCode_OPRC_Safebox_PasswordIllegal)
			return nil
		}

		var limitCount int32
		if platform != nil && platform.PerBankNoLimitAccount != 0 {
			limitCount = platform.PerBankNoLimitAccount
		}
		var prohibit bool
		var isLimitNum bool
		//绑定支付宝账号
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			cnt := model.CountAlipayAccountCount(p.Platform, msg.GetAlipayAccount())
			if limitCount > 0 && cnt >= int(limitCount) { //没有事务，存在容差，认为可以容忍
				prohibit = true
				return nil
			}
			if platform.PerBankNoLimitName > 0 {
				cnt := model.CountBankAlipayNameCount(p.Platform, alipayAccName, p.SnId)
				if cnt >= int(platform.PerBankNoLimitName) {
					isLimitNum = true
					return nil
				}
			}

			err := model.UpdatePlayerAlipay(p.Platform, p.SnId, alipayAccount, alipayAccName)
			if err == nil {
				model.NewBankBindLog(p.SnId, p.Platform, model.BankBindLogType_Ali, msg.GetAlipayAccName(),
					msg.GetAlipayAccount(), 1)
			}
			return err
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				if prohibit {
					sendPack(player_proto.OpResultCode_OPRC_BindAlipay_CountLimit)
				} else if isLimitNum {
					sendPack(player_proto.OpResultCode_OPRC_BindBankAlipay_NameCountLimit)
				} else {
					p.AlipayAccount = msg.GetAlipayAccount()
					p.AlipayAccName = msg.GetAlipayAccName()
					//用户绑定支付宝事件
					p.ReportBindAlipayEvent()
					sendPack(player_proto.OpResultCode_OPRC_Sucess)
				}
			} else {
				sendPack(player_proto.OpResultCode_OPRC_Error)
			}
		}), "UpdatePlayerAlipay").StartByExecutor(strconv.Itoa(int(p.SnId)))

	}
	return nil
}

// 绑定银行卡
type CSBindBankPacketFactory struct {
}
type CSBindBankHandler struct {
}

func (this *CSBindBankPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSBindBank{}
	return pack
}

func (this *CSBindBankHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSBindBankHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSBindBank); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSBindBankHandler p == nil")
			return nil
		}

		sendPack := func(code player_proto.OpResultCode) {
			pack := &player_proto.SCBindBank{
				OpRetCode:   code,
				Bank:        msg.Bank,
				BankAccount: msg.BankAccount,
				BankAccName: msg.BankAccName,
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_BINDBANK), pack)
		}
		platform := PlatformMgrSington.GetPlatform(p.Platform)
		if platform != nil {
			if !platform.IsMarkFlag(int32(Bind_BackCard)) {
				sendPack(player_proto.OpResultCode_OPRC_BindBank_PlatformError)
				return nil
			}
		}
		bank := msg.GetBank()
		if bank == "" {
			sendPack(player_proto.OpResultCode_OPRC_BindBank_NameEmpty)
			return nil
		}

		if !utf8.ValidString(bank) {
			sendPack(player_proto.OpResultCode_OPRC_BindBank_NameIllegal)
			return nil
		}

		bankAccount := msg.GetBankAccount()
		if bankAccount == "" {
			sendPack(player_proto.OpResultCode_OPRC_BindBank_AccountEmpty)
			return nil
		}

		if !utf8.ValidString(bankAccount) {
			sendPack(player_proto.OpResultCode_OPRC_BindBank_AccountIllegal)
			return nil
		}

		bankAccName := msg.GetBankAccName()
		if bankAccName == "" {
			sendPack(player_proto.OpResultCode_OPRC_BindBank_AccNameEmpty)
			return nil
		}

		if !utf8.ValidString(bankAccName) {
			sendPack(player_proto.OpResultCode_OPRC_BindBank_AccNameIllegal)
			return nil
		}

		if platform.NeedSameName {
			if p.AlipayAccName != "" {
				if p.AlipayAccName != bankAccName {
					sendPack(player_proto.OpResultCode_OPRC_BankAndAli_NotSame)
					return nil
				}
			}
		}

		if msg.GetPassword() == "" || len(msg.GetPassword()) < 32 {
			sendPack(player_proto.OpResultCode_OPRC_Safebox_PasswordIllegal)
			return nil
		}

		raw := fmt.Sprintf("%v%v%v", p.SafeBoxPassword, common.GetAppId(), msg.GetTimeStamp())
		h := md5.New()
		io.WriteString(h, raw)
		hashsum := hex.EncodeToString(h.Sum(nil))
		if hashsum != msg.GetPassword() {
			sendPack(player_proto.OpResultCode_OPRC_Safebox_PasswordIllegal)
			return nil
		}

		if p.BankAccount != "" {
			logger.Logger.Errorf("player had bank,now modify:%v", p.SnId)
		}

		var limitCount int32
		if platform != nil && platform.PerBankNoLimitAccount != 0 {
			limitCount = platform.PerBankNoLimitAccount
		}
		var prohibit bool
		var isLimitNum bool
		//绑定银行账号
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			cnt := model.CountBankAccountCount(p.Platform, bankAccount)
			if limitCount > 0 && cnt >= int(limitCount) { //没有事务，存在容差，认为可以容忍
				prohibit = true
				return nil
			}
			if platform.PerBankNoLimitName > 0 {
				cnt := model.CountBankAlipayNameCount(p.Platform, bankAccName, p.SnId)
				if cnt >= int(platform.PerBankNoLimitName) {
					isLimitNum = true
					return nil
				}
			}

			err := model.UpdatePlayerBank(p.Platform, p.SnId, bank, bankAccount, bankAccName)
			if err == nil {
				model.NewBankBindLog(p.SnId, p.Platform, model.BankBindLogType_Bank, bankAccName, bankAccount, 1)
			}
			return err
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				if prohibit {
					sendPack(player_proto.OpResultCode_OPRC_BindBank_CountLimit)
				} else if isLimitNum {
					sendPack(player_proto.OpResultCode_OPRC_BindBankAlipay_NameCountLimit)
				} else {
					p.Bank = bank
					p.BankAccount = bankAccount
					p.BankAccName = bankAccName
					sendPack(player_proto.OpResultCode_OPRC_Sucess)
				}
			} else {
				sendPack(player_proto.OpResultCode_OPRC_Error)
			}
		}), "UpdatePlayerBank").StartByExecutor(strconv.Itoa(int(p.SnId)))
	}
	return nil
}

// 修改密码
type CSChangePasswordPacketFactory struct {
}
type CSChangePasswordHandler struct {
}

func (this *CSChangePasswordPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSChangePassword{}
	return pack
}

func (this *CSChangePasswordHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSChangePasswordHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSChangePassword); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSChangePasswordHandler p == nil")
			return nil
		}

		sendPack := func(code player_proto.OpResultCode) {
			pack := &player_proto.SCChangePassword{
				OpRetCode:  code,
				ChangeType: msg.ChangeType,
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_CHANGEPASSWORD), pack)
		}

		//修改类型不匹配
		if msg.GetChangeType() < 0 || msg.GetChangeType() > 2 {
			sendPack(player_proto.OpResultCode_OPRC_Error)
			return nil
		}

		//旧密码
		if msg.GetOldPassword() == "" || len(msg.GetOldPassword()) < 32 {
			if msg.GetChangeType() == 0 {
				sendPack(player_proto.OpResultCode_OPRC_UpgradeAccount_PasswordIllegal)
			} else {
				sendPack(player_proto.OpResultCode_OPRC_Safebox_PasswordIllegal)
			}
			return nil
		}

		//新密码
		if msg.GetNewPassword() == "" || len(msg.GetNewPassword()) < 32 {
			if msg.GetChangeType() == 0 {
				sendPack(player_proto.OpResultCode_OPRC_UpgradeAccount_PasswordIllegal)
			} else {
				sendPack(player_proto.OpResultCode_OPRC_Safebox_PasswordIllegal)
			}
			return nil
		}

		if msg.GetNewPassword() == msg.GetOldPassword() && msg.GetChangeType() != 2 {
			sendPack(player_proto.OpResultCode_OPRC_PasswordEqual)
			return nil
		}

		switch msg.GetChangeType() { //0：帐号密码  1：保险箱密码  2：设置保险箱密码
		case 0:
			//修改帐号密码
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpdatePlayerPassword(p.Platform, p.AccountId, msg.GetOldPassword(), msg.GetNewPassword())
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data == nil {
					als := LoginStateMgrSington.GetLoginStateByAccId(p.AccountId)
					if als != nil && als.acc != nil {
						als.acc.TelPassWord = msg.GetNewPassword()
					}
					sendPack(player_proto.OpResultCode_OPRC_Sucess)
				} else {
					sendPack(player_proto.OpResultCode_OPRC_Error)
				}
			}), "UpdatePlayerPassword").StartByExecutor(strconv.Itoa(int(p.SnId)))
		case 1:
			//验证保险箱密码

			if p.SafeBoxPassword != msg.GetOldPassword() {
				sendPack(player_proto.OpResultCode_OPRC_Safebox_PasswordIllegal)
				return nil
			}
			//修改帐号密码
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpdateSafeBoxPassword(p.Platform, p.AccountId, msg.GetOldPassword(), msg.GetNewPassword())
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if i, ok := data.(int); ok {
					if i == 0 {
						p.SafeBoxPassword = msg.GetNewPassword()
						sendPack(player_proto.OpResultCode_OPRC_Sucess)
					} else {
						sendPack(player_proto.OpResultCode_OPRC_Error)
					}
				}
			}), "UpdatePlayerPassword").StartByExecutor(strconv.Itoa(int(p.SnId)))
		case 2:
			//验证保险箱密码
			/*raw := fmt.Sprintf("%v%v%v", p.SafeBoxPassword, common.GetAppId(), msg.GetTimeStamp())
			h := md5.New()
			io.WriteString(h, raw)
			hashsum := hex.EncodeToString(h.Sum(nil))
			if hashsum != msg.GetOldPassword() {
				sendPack(player_proto.OpResultCode_OPRC_Safebox_PasswordIllegal)
				return nil
			}*/
			raw := fmt.Sprintf("%v%v", model.DEFAULT_PLAYER_SAFEBOX_PWD, common.GetAppId())
			h := md5.New()
			io.WriteString(h, raw)
			pwd := hex.EncodeToString(h.Sum(nil))

			//修改帐号密码
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpdateSafeBoxPassword(p.Platform, p.AccountId, pwd, msg.GetNewPassword())
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if i, ok := data.(int); ok {
					if i == 0 {
						p.SafeBoxPassword = msg.GetNewPassword()
						sendPack(player_proto.OpResultCode_OPRC_Sucess)
					} else {
						sendPack(player_proto.OpResultCode_OPRC_Error)
					}
				}
			}), "UpdatePlayerPassword").StartByExecutor(strconv.Itoa(int(p.SnId)))
		}
		//sendPack(player_proto.OpResultCode_OPRC_Sucess)
	}
	return nil
}

// 操作保险箱
type CSPlayerSafeBoxPacketFactory struct {
}
type CSPlayerSafeBoxHandler struct {
}

func (this *CSPlayerSafeBoxPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSPlayerOpCoin{}
	return pack
}

func (this *CSPlayerSafeBoxHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPlayerSafeBoxHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSPlayerOpCoin); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSPlayerSafeBoxHandler p == nil")
			return nil
		}
		if p.scene != nil {
			logger.Logger.Warn("CSPlayerSafeBoxHandler p.scene != nil")
			return nil
		}
		sendPack := func(code player_proto.OpResultCode) {
			pack := &player_proto.SCPlayerOpCoin{
				OpRetCode:   code,
				Op:          msg.Op,
				Coin:        proto.Int64(p.Coin),
				SafeBoxCoin: proto.Int64(p.SafeBoxCoin),
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_PLAYEROPCOIN), pack)
		}
		if msg.GetCoin() <= 0 {
			sendPack(player_proto.OpResultCode_OPRC_Error)
			return nil
		}
		snid := p.SnId
		beforeCoin := p.Coin
		beforeBank := p.SafeBoxCoin
		afterCoin := p.Coin
		afterBank := p.SafeBoxCoin
		if msg.GetOp() == model.SafeBoxLogType_Save {
			//验证金币
			if p.Coin < msg.GetCoin() {
				sendPack(player_proto.OpResultCode_OPRC_CoinNotEnough)
				return nil
			}
			//计算日志数据
			afterCoin = p.Coin - msg.GetCoin()
			afterBank = p.SafeBoxCoin + msg.GetCoin()
			//操作数据
			p.SafeBoxCoin += msg.GetCoin()
			p.AddCoin(-msg.GetCoin(), common.GainWay_SafeBoxSave, "system", "保险箱存入")
			p.dirty = true
		} else if msg.GetOp() == model.SafeBoxLogType_TakeOut {
			//验证金币
			if p.SafeBoxCoin < msg.GetCoin() {
				sendPack(player_proto.OpResultCode_OPRC_CoinNotEnough)
				return nil
			}
			//验证密码
			if msg.GetPassword() == "" || len(msg.GetPassword()) < 32 {
				sendPack(player_proto.OpResultCode_OPRC_Safebox_PasswordIllegal)
				return nil
			}
			hashsum := common.MakeMd5String(p.SafeBoxPassword, common.GetAppId(), strconv.Itoa(int(msg.GetTimeStamp())))
			if hashsum != msg.GetPassword() {
				sendPack(player_proto.OpResultCode_OPRC_SafeBoxPasswordError)
				return nil
			}
			//计算日志数据
			afterCoin = p.Coin + msg.GetCoin()
			afterBank = p.SafeBoxCoin - msg.GetCoin()
			//操作数据
			p.SafeBoxCoin -= msg.GetCoin()
			p.AddCoin(msg.GetCoin(), common.GainWay_SafeBoxTakeOut, "system", "保险箱取出")
			p.dirty = true
		}

		p.SendDiffData()
		sendPack(player_proto.OpResultCode_OPRC_Sucess)
		var oper string
		var logs []*model.PayCoinLog
		var coinPayts int64
		var safePayts int64
		billNo := time.Now().UnixNano()
		if msg.GetOp() == model.SafeBoxLogType_Save {
			log := model.NewPayCoinLog(billNo, p.SnId, msg.GetCoin(), 0, "system", model.PayCoinLogType_SafeBoxCoin, 0)
			if log != nil {
				safePayts = log.TimeStamp
				logs = append(logs, log)
			}
			billNo += 1
			log = model.NewPayCoinLog(billNo, p.SnId, -msg.GetCoin(), 0, "system", model.PayCoinLogType_Coin, 0)
			if log != nil {
				coinPayts = log.TimeStamp
				logs = append(logs, log)
			}
			oper = "保险箱存入"
		} else if msg.GetOp() == model.SafeBoxLogType_TakeOut {
			log := model.NewPayCoinLog(billNo, p.SnId, -msg.GetCoin(), 0, "system", model.PayCoinLogType_SafeBoxCoin, 0)
			if log != nil {
				safePayts = log.TimeStamp
				logs = append(logs, log)
			}
			billNo += 1
			log = model.NewPayCoinLog(billNo, p.SnId, msg.GetCoin(), 0, "system", model.PayCoinLogType_Coin, 0)
			if log != nil {
				coinPayts = log.TimeStamp
				logs = append(logs, log)
			}
			oper = "保险箱取出"
		}
		isNeedRollBack := true
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {

			err := model.InsertPayCoinLogs(p.Platform, logs...)
			if err != nil {
				logger.Logger.Trace("InsertPayCoinLogs err:", err)
				return err
			}
			err, _ = model.InsertSafeBox(p.SnId, msg.GetCoin(), beforeBank, afterBank, beforeCoin, afterCoin,
				msg.GetOp(), time.Now(), p.Ip, oper, p.Platform, p.Channel, p.BeUnderAgentCode)
			if err != nil {
				isNeedRollBack = false
			}
			return err
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data != nil && isNeedRollBack {
				//如果保存失败，并不会回滚玩家的数据，主要原因时无法回滚，如果回滚就需要先扣后补，因为只是在玩家身上流转，
				//所以可以可以不考虑回滚问题
				logger.Logger.Errorf("Player %v box coin op log error.", snid)
			} else {
				p.SetPayTs(coinPayts)
				p.SetSafeBoxPayTs(safePayts)
				p.dirty = true
			}
		}), oper).StartByExecutor(strconv.Itoa(int(p.SnId)))
	}
	return nil
}

// 读取保险箱记录
type CSPlayerSafeBoxCoinLogPacketFactory struct {
}
type CSPlayerSafeBoxCoinLogHandler struct {
}

func (this *CSPlayerSafeBoxCoinLogPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSGetSafeBoxCoinLog{}
	return pack
}

func (this *CSPlayerSafeBoxCoinLogHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("GetSafeBoxCoinLog Process recv ", data)
	if _, ok := data.(*player_proto.CSGetSafeBoxCoinLog); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("GetSafeBoxCoinLog p == nil")
			return nil
		}

		pack := &player_proto.SCGetSafeBoxCoinLog{}

		sendPack := func(code player_proto.OpResultCode) {
			pack.OpRetCode = code
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_GETSAFEBOXCOINLOG), pack)
		}

		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			if sblist, err := model.GetSafeBoxs(p.Platform, p.SnId); err == nil {
				return sblist
			} else {
				sendPack(player_proto.OpResultCode_OPRC_Error)
				return nil
			}
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if safeboxrec, ok := data.([]model.SafeBoxRec); ok {
				for i := 0; i < len(safeboxrec); i++ {
					safeBoxInfo := &player_proto.SafeBoxCoinLog{
						LogType:     proto.Int32(safeboxrec[i].LogType),
						OPCoin:      proto.Int64(safeboxrec[i].Count),
						OPCoinFront: proto.Int64(safeboxrec[i].BeforeSafeBox),
						OPCoinLast:  proto.Int64(safeboxrec[i].AfterSafeBox),
						Ts:          proto.Int64(safeboxrec[i].Time.Unix()),
					}
					pack.Logs = append(pack.Logs, safeBoxInfo)
				}
				sendPack(player_proto.OpResultCode_OPRC_Sucess)
			} else {
				sendPack(player_proto.OpResultCode_OPRC_Error)
			}
		}), "GetSafeBoxCoinLog").StartByExecutor(strconv.Itoa(int(p.SnId)))
	}
	return nil
}

//
////读取游戏记录
//type CSPlayerGameCoinLogPacketFactory struct {
//}
//type CSPlayerGameCoinLogHandler struct {
//}
//
//func (this *CSPlayerGameCoinLogPacketFactory) CreatePacket() interface{} {
//	pack := &player_proto.CSGetGameCoinLog{}
//	return pack
//}
//
//func (this *CSPlayerGameCoinLogHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSPlayerGameCoinLogHandler Process recv ", data)
//	if _, ok := data.(*player_proto.CSGetGameCoinLog); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSPlayerGameCoinLogHandler p == nil")
//			return nil
//		}
//		//model.InsertGameRecList(p.SnId, 0, 20, p.entercoin, p.Coin, p.enterts, time.Now())
//		pack := &player_proto.SCGetGameCoinLog{}
//		sendPack := func(code player_proto.OpResultCode) {
//			pack.OpRetCode = code
//			proto.SetDefaults(pack)
//			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_GETGAMECOINLOG), pack)
//		}
//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//			if glist, err := model.GetGameRecLogs(p.SnId); err == nil {
//				return glist
//			} else {
//				sendPack(player_proto.OpResultCode_OPRC_Error)
//				return nil
//			}
//		}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//			if gamelist, ok := data.([]model.GameRecLog); ok {
//				for i := 0; i < len(gamelist); i++ {
//					gameInfo := &player_proto.GameCoinLog{
//						GameId:     proto.Int32(gamelist[i].GameId),
//						EnterCount: proto.Int64(gamelist[i].EnterCount),
//						LeaveCount: proto.Int64(gamelist[i].LeaveCount),
//						EnterTs:    proto.Int64(gamelist[i].EnterTs.Unix()),
//						LeaveTs:    proto.Int64(gamelist[i].LeaveTs.Unix()),
//					}
//					pack.Logs = append(pack.Logs, gameInfo)
//				}
//				sendPack(player_proto.OpResultCode_OPRC_Sucess)
//			} else {
//				sendPack(player_proto.OpResultCode_OPRC_Error)
//			}
//		}), "GetSafeBoxCoinLog").StartByExecutor(p.AccountId)
//	}
//	return nil
//}

// 注册帐号
type CSPlayerRegisterPacketFactory struct {
}
type CSPlayerRegisterHandler struct {
}

func (this *CSPlayerRegisterPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSRegister{}
	return pack
}

func (this *CSPlayerRegisterHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPlayerRegisterHandler Process recv ", data)
	if t, ok := data.(*player_proto.CSRegister); ok {
		sendPack := func(code player_proto.OpResultCode) {
			pack := &player_proto.SCRegister{
				OpRetCode: code,
			}
			proto.SetDefaults(pack)
			//p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_UPGRADEACCOUNT), pack)
			common.SendToGate(sid, int(player_proto.PlayerPacketID_PACKET_SC_REGISTER), pack, s)
		}

		//替换为后台拿到的配置数据，不再使用客户端的发送数据,
		platformID, channel, promoter, _, tagkey := PlatformMgrSington.GetPlatformByPackageTag(t.GetPlatformTag())
		var loginType int32
		//手机号注册，验证参数
		if t.GetRegistType() == 0 {
			loginType = 3
			//检验手机号合法
			tel := t.GetTel()
			if !reTelRule.MatchString(tel) {
				sendPack(player_proto.OpResultCode_OPRC_TelError)
				return nil
			}

			//验证验证码
			platform := PlatformMgrSington.GetPlatform(strconv.Itoa(int(platformID)))
			if !platform.RegisterVerifyCodeSwitch {
				if platform.VerifyCodeType != common.CodeTypeNo { // 不使用验证码注册
					code := CacheMemory.Get(fmt.Sprintf("%v_%v", tel, tagkey))
					if code == nil {
						sendPack(player_proto.OpResultCode_OPRC_VerificationCodeError)
						return nil
					}
					if verifyCode, ok := code.(string); ok && verifyCode != "" {
						if verifyCode != t.GetVerificationCode() {
							sendPack(player_proto.OpResultCode_OPRC_VerificationCodeError)
							return nil
						}
					}
				}
			}

			//验证密码
			if t.GetTelPassword() == "" || len(t.GetTelPassword()) < 32 {
				sendPack(player_proto.OpResultCode_OPRC_UpgradeAccount_PasswordIllegal)
				return nil
			}
			if t.GetPassword() == "" || len(t.GetPassword()) < 32 {
				sendPack(player_proto.OpResultCode_OPRC_UpgradeAccount_PasswordIllegal)
				return nil
			}
		} else {
			loginType = 4
			if t.GetUsername() == "" {
				sendPack(player_proto.OpResultCode_OPRC_UserNameError)
				return nil
			}
			//验证密码
			if t.GetTelPassword() == "" || len(t.GetTelPassword()) < 32 {
				sendPack(player_proto.OpResultCode_OPRC_UpgradeAccount_PasswordIllegal)
				return nil
			}
		}

		backupPromoter := ""

		strPlatform := strconv.Itoa(int(platformID))
		if t.GetChannel() != common.Channel_Rob {
			t.Platform = proto.String(strPlatform)
			t.Channel = proto.String(strconv.Itoa(int(channel)))
			//因为后台需要，修改此方法，修改为客户端传递的数据
			if len(t.GetPromoter()) <= 0 {
				t.Promoter = proto.String(strconv.Itoa(int(promoter)))
			} else {
				backupPromoter = t.GetPromoter()
			}
		}

		amount := int32(0)
		plt := PlatformMgrSington.GetPlatform(strPlatform)
		//if plt != nil {
		//	amount += plt.NewAccountGiveCoin
		//	amount += plt.UpgradeAccountGiveCoin
		//}

		var temp []byte
		var ee error
		di := t.GetDeviceInfo()
		if plt.NeedDeviceInfo && di == "" { //maybe res ver is low
			sendPack(player_proto.OpResultCode_OPRC_YourResVerIsLow)
			return nil
		}

		if di != "" {
			var e common.Encryptor
			e.Init(common.GetAppId(), t.GetPlatformTag(), int32(t.GetTimeStamp()))
			temp, ee = base64.StdEncoding.DecodeString(di)
			if ee != nil {
				sendPack(player_proto.OpResultCode_OPRC_YourResVerIsLow)
				return nil
			}

			e.Encrypt(temp, len(temp))

			if model.GameParamData.ValidDeviceInfo && !json.Valid(temp) {
				sendPack(player_proto.OpResultCode_OPRC_YourResVerIsLow)
				return nil
			}
		}

		deviceInfo := string(temp)
		if bi, ok := BlackListMgrSington.CheckDeviceInBlack(deviceInfo, BlackState_Login, t.GetPlatform()); ok {
			sendPack(player_proto.OpResultCode_OPRC_InBlackList)
			if bi != nil {
				common.SendToGate(sid, int(player_proto.PlayerPacketID_PACKET_SC_SRVMSG), common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, i18n.Tr("cn", "BlackListLimit1Args", bi.Id)), s)
			}
			return nil
		}

		var pi *model.PlayerData
		var acc *model.Account
		var promoterCfg *PromoterConfig
		var invitePromoterID string
		var tf bool
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			_, err := model.AccountIsExist(t.GetUsername(), t.GetTel(), t.GetPassword(), t.GetPlatform(), t.GetTimeStamp(), loginType, tagkey, false)
			if err == 8 {
				//手机号已经注册
				return player_proto.OpResultCode_OPRC_TelIsRegister
			}
			if err == 7 {
				//账号密码已经注册
				return player_proto.OpResultCode_OPRC_Login_CreateFailed
			}
			if err == 6 || err == 7 {
				//新帐号,游客
				raw := fmt.Sprintf("%v%v", t.GetUsername(), common.GetAppId())
				h := md5.New()
				io.WriteString(h, raw)
				pwd := hex.EncodeToString(h.Sum(nil))
				acc = &model.Account{}

				if err == 6 {
					//可以帐号密码登录，也可以游客登录
					acc, err = model.InsertTelAccount(t.GetUsername(), pwd, t.GetPlatform(), t.GetChannel(), t.GetPromoter(), t.GetParams(), t.GetInviterId(), t.GetPromoterTree(), t.GetTel(), t.GetTelPassword(), t.GetPlatformTag(), t.GetPackage(), deviceInfo, tagkey)
				} else if err == 7 {
					//只能帐号密码登录,设备号+手机号+验证码+时间戳
					username := t.GetUsername() + t.GetTel() + t.GetVerificationCode() + strconv.FormatInt(t.GetTimeStamp(), 10)
					//username := t.GetUsernames()
					acc, err = model.InsertTelAccount(username, pwd, t.GetPlatform(), t.GetChannel(), t.GetPromoter(), t.GetParams(), t.GetInviterId(), t.GetPromoterTree(), t.GetTel(), t.GetTelPassword(), t.GetPlatformTag(), t.GetPackage(), deviceInfo, tagkey)
				}
				if err != 5 {
					return player_proto.OpResultCode_OPRC_Error
				}
				//生成玩家数据

				pi, tf = model.CreatePlayerDataOnRegister(acc.Platform, acc.AccountId.Hex(), amount)
				if pi == nil || tf == false {
					return player_proto.OpResultCode_OPRC_Error
				}
			} else {
				logger.Logger.Tracef("CSPlayerRegister Player err:%v", err)
				return player_proto.OpResultCode_OPRC_Error
			}

			if pi != nil && pi.InviterId != 0 {
				invitePd := model.GetPlayerBaseInfo(pi.Platform, pi.InviterId)
				if invitePd != nil {
					invitePromoterID = invitePd.BeUnderAgentCode
				}
			}

			return player_proto.OpResultCode_OPRC_Sucess
		}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			if code, ok := data.(player_proto.OpResultCode); ok {
				if code != player_proto.OpResultCode_OPRC_Sucess {
					sendPack(code)
					return
				}
			}

			key, err := GetPromoterKey(0, invitePromoterID, "")
			if err == nil {
				promoterCfg = PromoterMgrSington.GetConfig(key)
			}

			//func (this *Player) ReportBindPhoneEvent(amount, tag int32)
			{
				//LogChannelSington.WriteMQData(model.GenerateBindEvent(model.CreatePlayerBindPhoneEvent(
				//	pi.SnId, pi.Channel, pi.BeUnderAgentCode, t.GetPlatform(), pi.City, pi.DeviceOS, pi.CreateTime,
				//	pi.TelephonePromoter)))
			}

			//清理掉cache中的无效id
			PlayerCacheMgrSington.UncacheInvalidPlayerId(pi.SnId)

			//记录赠送日志
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {

				if pi.InviterId > 100000 {
					tag, err := webapi.API_PushSpreadLink(pi.SnId, pi.Platform, pi.PackageID, int(pi.InviterId), 1, 0, common.GetAppId())
					if err != nil {
						logger.Logger.Warnf("API_PushSpreadLink response tag:%v err:%v", tag, err)
					}
				}

				var promoterTreeIsValid bool
				//无级代理推广员验证
				if pi.PromoterTree != 0 {
					tag, msg := webapi.API_ValidPromoterTree(pi.SnId, pi.PackageID, pi.PromoterTree, common.GetAppId())
					if tag != 0 {
						logger.Logger.Warnf("API_ValidPromoterTree response tag:%v msg:%v", tag, msg)
						pi.PromoterTree = 0
						platform, _ := strconv.Atoi(pi.Platform)
						//先修改账号数据
						err := model.UpdateAccountPlatformInfo(pi.AccountId, int32(platform), 0, 0, pi.InviterId, pi.PromoterTree, pi.PackageID)
						if err != nil {
							logger.Logger.Warnf("UpdateAccountPlatformInfo err:%v", err)
						}
						err = model.UpdatePlayerPackageId(pi.SnId, pi.PackageID, int32(platform), 0, 0, pi.InviterId, pi.PromoterTree)
						if err != nil {
							logger.Logger.Warnf("UpdatePlayerPackageId err:%v", err)
						}
					} else {
						promoterTreeIsValid = true
					}
				}

				//推广关系获取
				if !model.GameParamData.DonotGetPromoterByIp {
					tag, msg, err := webapi.API_PushInviterIp(pi.SnId, pi.InviterId, pi.PromoterTree, backupPromoter,
						pi.PackageID, pi.Ip, pi.DeviceOS, common.GetAppId())
					if err == nil {
						if msg != nil {
							if msg.Tag != "" {
								pi.PackageID = msg.Tag
								pi.Platform = strconv.Itoa(int(msg.Platform))           //平台
								pi.Channel = strconv.Itoa(int(msg.ChannelId))           //渠道
								pi.BeUnderAgentCode = strconv.Itoa(int(msg.PromoterId)) //推广员
								if pi.InviterId == 0 && msg.Spreader != 0 {
									pi.InviterId = msg.Spreader //全民代邀请人
								}
								if msg.PromoterTree != 0 && !promoterTreeIsValid {
									pi.PromoterTree = msg.PromoterTree //无级代上级
								}
								if promoterCfg != nil {
									if promoterCfg.IsInviteRoot > 0 {
										pi.BeUnderAgentCode = invitePromoterID
									}
								}
								//先修改账号数据
								err = model.UpdateAccountPlatformInfo(pi.AccountId, msg.Platform, msg.ChannelId, msg.PromoterId, msg.Spreader, msg.PromoterTree, msg.Tag)
								if err != nil {
									logger.Logger.Warnf("UpdateAccountPlatformInfo err:%v", err)
								}
								err = model.UpdatePlayerPackageId(pi.SnId, msg.Tag, msg.Platform, msg.ChannelId, msg.PromoterId, msg.Spreader, msg.PromoterTree)
								if err != nil {
									logger.Logger.Warnf("UpdatePlayerPackageId err:%v", err)
								}
							}
						}
					} else {
						logger.Logger.Warnf("API_PushInviterIp response tag:%v msg:%v err:%v", tag, msg, err)
					}
				}

				return nil
			}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {

				//需要先获得关系才能给与升级的奖励，因为现在数据和推广员绑定到了一起
				var cfgInfo *PromoterConfig
				key, err := GetPromoterKey(pi.PromoterTree, pi.BeUnderAgentCode, pi.Channel)
				if err == nil {
					cfgInfo = PromoterMgrSington.GetConfig(key)
				}

				amount += GGetRegisterPrize(plt, cfgInfo)
				amount += GGetUpdateAccPrize(plt, cfgInfo)

				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					if amount > 0 {
						//当前IP和日期已经赠送的金币总额
						var dataParams model.PlayerParams
						json.Unmarshal([]byte(acc.Params), &dataParams)
						upgradeaccountcount := model.GetUpgradeAccountCoinLogsByIPAndDate(dataParams.Ip, strPlatform, time.Now())
						if model.GameParamData.UpgradeAccountGiveCoinLimit != 0 && upgradeaccountcount >= model.GameParamData.UpgradeAccountGiveCoinLimit {
							amount = 0
							sendPack(player_proto.OpResultCode_OPRC_Account_IP_TooManyReg)
						}
					}
					model.InsertUpgradeAccountCoinLog(pi.RegIp, time.Now(), amount, pi.SnId, pi.Channel, pi.Platform, pi.BeUnderAgentCode, pi.PackageID, pi.City, pi.InviterId)

					if amount > 0 {
						//更新账号金币
						model.UpdatePlayerSetCoin(pi.Platform, pi.SnId, pi.Coin+int64(amount))
					}
					return nil
				}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
					tag := common.GainWay_UpgradeAccount
					if amount > 0 {
						log := model.NewCoinLogEx(pi.SnId, int64(amount), pi.Coin+int64(amount), pi.SafeBoxCoin,
							pi.Ver, int32(tag), 0, "", "", pi.Platform, pi.Channel, pi.BeUnderAgentCode, 0, pi.PackageID, 0)
						if log != nil {
							LogChannelSington.WriteLog(log)
						}
						ReportSystemGiveEvent(pi, amount, int32(tag), true)
					}
					//调整返回在此处
					sendPack(player_proto.OpResultCode_OPRC_Sucess)
				}), "PlayerRegister").Start()

			}), "InsertUpgradeAccountCoinLog").StartByFixExecutor("InsertUpgradeAccountCoinLog")

			////创建账号事件
			//pi.ReportWebEvent(model.WEBEVENT_LOGIN, 1, 1, 1)
			////绑定账号事件
			//pi.ReportWebEvent(model.WEBEVENT_UPGRADEACC, 1, 1, 1)

			//}), "PlayerRegister").StartByExecutor(strconv.Itoa(int(acc.SnId)))
		}), "PlayerRegister").StartByExecutor(t.GetUsername())
	}
	return nil
}

type CSWebAPIPlayerPassPacketFactory struct {
}

type CSWebAPIPlayerPassHandler struct {
}

func (this *CSWebAPIPlayerPassPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSWebAPIPlayerPass{}
	return pack
}

func (this *CSWebAPIPlayerPassHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSWebAPIPlayerPassHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSWebAPIPlayerPass); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSWebAPIPlayerPassHandler p == nil")
			return nil
		}
		opCode := player_proto.OpResultCode_OPRC_Sucess
		errString := ""
		var err error
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			errString, err = webapi.API_PlayerPass(p.SnId, p.Platform, p.Channel, p.BeUnderAgentCode, msg.GetApiName(), msg.GetParams(), common.GetAppId(), p.LogicLevels)
			if err != nil {
				logger.Logger.Errorf("API_PlayerPass error:%v api:%v params:%v", err, msg.GetApiName(), msg.GetParams())
				opCode = player_proto.OpResultCode_OPRC_Error
				return nil
			}
			return err
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			pack := &player_proto.SCWebAPIPlayerPass{
				OpRetCode: opCode,
				ApiName:   msg.ApiName,
				CBData:    msg.CBData,
				Response:  proto.String(errString),
			}
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_WEBAPI_PLAYERPASS), pack)
			logger.Logger.Trace("CSWebAPIPlayerPass:", pack)
		}), "API_PlayerPass").Start()
	}
	return nil
}

type CSWebAPISystemPassPacketFactory struct {
}

type CSWebAPISystemPassHandler struct {
}

func (this *CSWebAPISystemPassPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSWebAPISystemPass{}
	return pack
}

func (this *CSWebAPISystemPassHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSWebAPISystemPassHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSWebAPISystemPass); ok {
		opCode := player_proto.OpResultCode_OPRC_Sucess
		errString := ""
		var err error
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {

			errString, err = webapi.API_SystemPass(msg.GetApiName(), msg.GetParams(), common.GetAppId())
			if err != nil {
				logger.Logger.Errorf("API_SystemPass error:%v apiname=%v params=%v", err, msg.GetApiName(), msg.GetParams())
				opCode = player_proto.OpResultCode_OPRC_Error
				return nil
			}
			return err
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			pack := &player_proto.SCWebAPISystemPass{
				OpRetCode: opCode,
				ApiName:   msg.ApiName,
				CBData:    msg.CBData,
				Response:  proto.String(errString),
			}
			common.SendToGate(sid, int(player_proto.PlayerPacketID_PACKET_SC_WEBAPI_SYSTEMPASS), pack, s)
			logger.Logger.Trace("CSWebAPISystemPass:", pack)
		}), "API_SystemPass").Start()
	}
	return nil
}

type CSSpreadBindPacketFactory struct {
}

type CSSpreadBindHandler struct {
}

func (this *CSSpreadBindPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSSpreadBind{}
	return pack
}

func (this *CSSpreadBindHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSSpreadBindHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSSpreadBind); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSSpreadBindHandler p == nil")
			return nil
		}

		sendPack := func(opCode player_proto.OpResultCode, parentId int32) {
			pack := &player_proto.SCSpreadBind{
				OpRetCode: opCode,
				ParentId:  msg.ParentId,
			}
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_SPREADBIND), pack)
			return
		}

		parentId := msg.GetParentId()
		parent := PlayerMgrSington.GetPlayerBySnId(parentId)
		if parent != nil {
			if parent.Platform != p.Platform {
				sendPack(player_proto.OpResultCode_OPRC_InviterIdNotExist, 0)
				return nil
			}
			if parent.Tel == "" {
				sendPack(player_proto.OpResultCode_OPRC_InviterNoBind, 0)
				return nil
			}
		}

		if model.GameParamData.InvitePromoterBind && p.BeUnderAgentCode != "" && p.BeUnderAgentCode != "0" {
			sendPack(player_proto.OpResultCode_OPRC_HadSpreadInviterId, 0)
			return nil
		}

		if p.InviterId != 0 {
			sendPack(player_proto.OpResultCode_OPRC_HadSpreadInviterId, 0)
			return nil
		}
		var promoterCfg *PromoterConfig
		upPromoterID := ""
		opCode := player_proto.OpResultCode_OPRC_Sucess
		tag := int32(0)
		pAgentCode := ""
		PlayerCacheMgrSington.Get(p.Platform, parentId, func(ppd *PlayerCacheItem, asyn, isnew bool) {
			if ppd == nil || ppd.Platform != p.Platform {
				opCode = player_proto.OpResultCode_OPRC_InviterIdNotExist
				sendPack(opCode, parentId)
				return
			}

			if ppd.Tel == "" {
				opCode = player_proto.OpResultCode_OPRC_InviterNoBind
				sendPack(opCode, parentId)
				return
			}

			pAgentCode = ppd.BeUnderAgentCode
			key, err := GetPromoterKey(0, pAgentCode, "")
			if err == nil {
				upPromoterID = pAgentCode
				promoterCfg = PromoterMgrSington.GetConfig(key)
			}

			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				tag, err = webapi.API_PushSpreadLink(p.SnId, p.Platform, p.PackageID, int(parentId), 1, 1, common.GetAppId())
				if (err != nil || tag != 0) && !common.Config.IsDevMode {
					logger.Logger.Errorf("API_PushSpreadLink error: %v tag: %v", err, tag)
					if tag == 101 { //闭环
						opCode = player_proto.OpResultCode_OPRC_SpreadBindClosedLoop
					} else {
						opCode = player_proto.OpResultCode_OPRC_SpreadBindFailed
					}
					return nil
				} else {
					p.InviterId = parentId
					p.dirty = true
					if promoterCfg != nil {
						if promoterCfg.IsInviteRoot > 0 {
							p.BeUnderAgentCode = upPromoterID
						}
					}

					err = model.UpdatePlayerPackageIdByStr(p.SnId, p.PackageID, p.Platform, p.Channel, p.BeUnderAgentCode, p.InviterId, p.PromoterTree)
				}
				return err
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if opCode == player_proto.OpResultCode_OPRC_Sucess {
					//actRandCoinMgr.OnPlayerInvite(p.Platform, parentId)
					//actRandCoinMgr.tempBind[p.SnId] = time.Now().Unix()

				}
				sendPack(opCode, parentId)
			}), "API_PushSpreadLink").Start()

		}, false)
	}
	return nil
}

type CSBindPromoterPacketFactory struct {
}

type CSBindPromoterHandler struct {
}

func (this *CSBindPromoterPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSBindPromoter{}
	return pack
}

func (this *CSBindPromoterHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSBindPromoterHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSBindPromoter); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSBindPromoterHandler p == nil")
			return nil
		}

		sendPack := func(opCode player_proto.OpResultCode, parentId int32) {
			pack := &player_proto.SCBindPromoter{
				OpRetCode: opCode,
				Promoter:  msg.Promoter,
			}
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_BINDPROMOTER), pack)
			return
		}

		promoter := msg.GetPromoter()

		info, ok := PlatformMgrSington.PromoterList[promoter]
		if !ok {
			sendPack(player_proto.OpResultCode_OPRC_NoPromotor, 0)
			return nil
		}

		plt := PlatformMgrSington.GetPlatform(p.Platform)
		if plt == nil {
			sendPack(player_proto.OpResultCode_OPRC_NoPlatform, 0)
			return nil
		}

		if !plt.IsCanUserBindPromoter {
			sendPack(player_proto.OpResultCode_OPRC_CantUserBind, 0)
			return nil
		}

		if p.BeUnderAgentCode != "" && p.BeUnderAgentCode != "0" {
			sendPack(player_proto.OpResultCode_OPRC_PromoterHasBind, 0)
			return nil
		}

		if promoter == "" || promoter == "0" {
			sendPack(player_proto.OpResultCode_OPRC_CantUserBind, 0)
			return nil
		}

		if info.Platform != p.Platform {
			sendPack(player_proto.OpResultCode_OPRC_PlatformNoPromoter, 0)
			return nil
		}
		if plt.UserBindPromoterPrize > 0 && !p.layered[common.ActId_PromoterBind] {
			p.AddCoin(int64(plt.UserBindPromoterPrize), common.GainWay_PromoterBind, "system", promoter)
			p.ReportSystemGiveEvent(plt.UserBindPromoterPrize, common.GainWay_PromoterBind, true)
		}

		p.BeUnderAgentCode = promoter
		p.PackageID = info.Tag
		p.dirty = true
		sendPack(player_proto.OpResultCode_OPRC_Sucess, 0)

		p.SendPlatformCanUsePromoterBind()
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			_, _ = webapi.API_PushInvitePromoter(p.SnId, promoter, common.GetAppId())
			return nil
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {

		}), "API_PushInviteLinkPromoter").Start()

	}
	return nil
}

type CSGenCustomerTokenPacketFactory struct {
}

type CSGenCustomerTokenHandler struct {
}

func (this *CSGenCustomerTokenPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSGenCustomerToken{}
	return pack
}

func (this *CSGenCustomerTokenHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGenCustomerTokenHandler Process recv ", data)
	if _, ok := data.(*player_proto.CSGenCustomerToken); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSGenCustomerTokenHandler p == nil")
			return nil
		}

		token := p.GenCustomerToken()
		PlayerMgrSington.UpdatePlayerToken(p, token)
		pack := &player_proto.SCGenCustomerToken{Token: proto.String(token)}
		proto.SetDefaults(pack)
		p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_GENCUSTOMTOKEN), pack)
	}
	return nil
}

type CSCustomerNewMsgAckPacketFactory struct {
}

type CSCustomerNewMsgAckHandler struct {
}

func (this *CSCustomerNewMsgAckPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSCustomerNewMsgAck{}
	return pack
}

func (this *CSCustomerNewMsgAckHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSCustomerNewMsgAckHandler Process recv ", data)
	//if msg, ok := data.(*player_proto.CSCustomerNewMsgAck); ok {
	//	p := PlayerMgrSington.GetPlayer(sid)
	//	if p == nil {
	//		logger.Logger.Warn("CSCustomerNewMsgAckHandler p == nil")
	//		return nil
	//	}
	//
	//	CustomerOfflineMsgMgrSington.OnOfflineMsgAck(p.SnId, msg.GetMsgIds())
	//}
	return nil
}

type CSIosInstallStablePacketFactory struct {
}

type CSIosInstallStableHandler struct {
}

func (this *CSIosInstallStablePacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSIosInstallStable{}
	return pack
}

func (this *CSIosInstallStableHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSIosInstallStableHandler Process recv ", data)
	//if _, ok := data.(*player_proto.CSIosInstallStable); ok {
	//	p := PlayerMgrSington.GetPlayer(sid)
	//	if p == nil {
	//		logger.Logger.Warn("CSIosInstallStableHandler p == nil")
	//		return nil
	//	}
	//
	//	if p.CoinPayTotal < int64(model.GameParamData.IosStablePrizeMinRecharge) { //不满足最低充值需求
	//		return nil
	//	}
	//	if p.DeviceOS != "ios" { //不是苹果系统，直接忽略
	//		return nil
	//	}
	//	if p.IosStableState != model.IOSSTABLESTATE_NIL { //未标注过才行
	//		return nil
	//	}
	//
	//	p.IosStableState = model.IOSSTABLESTATE_MARK
	//	p.dirty = true
	//}
	return nil
}

// 查询中奖
type CSFishJackpotPacketFactory struct {
}
type CSFishJackpotHandler struct {
}

func (this *CSFishJackpotPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSJackpotList{}
	return pack
}
func (this *CSFishJackpotHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSFishJackpotHandler Process recv ", data)
	if _, ok := data.(*player_proto.CSJackpotList); ok {
		player := PlayerMgrSington.GetPlayer(sid)
		if player == nil {
			logger.Logger.Warn("CSFishJackpotHandler p == nil")
			return nil
		}

		pack := &player_proto.SCJackpotList{}

		datas := FishJackListMgr.GetJackInfo(player.Platform)
		for _, v := range datas {
			info := &player_proto.FishJackpotInfo{
				Name: proto.String(v.Name),
				Coin: proto.Int64(v.JackpotCoin),
				Type: proto.Int32(v.JackpotType),
				Ts:   proto.Int64(v.Ts),
			}
			pack.JackpotList = append(pack.JackpotList, info)
		}
		logger.Logger.Trace("CSFishJackpotHandler Pb  ", pack)
		player.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_FISHJACKPOTCOIN), pack)

		return nil
	}
	return nil
}

// 查询奖池金额
type CSFishJackpotCoinPacketFactory struct {
}
type CSFishJackpotCoinHandler struct {
}

func (this *CSFishJackpotCoinPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSFishJackpotCoin{}
	return pack
}
func (this *CSFishJackpotCoinHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSFishJackpotCoinHandler Process recv ", data)
	if _, ok := data.(*player_proto.CSFishJackpotCoin); ok {
		player := PlayerMgrSington.GetPlayer(sid)
		if player == nil {
			logger.Logger.Warn("CSFishJackpotCoinHandler p == nil")
			return nil
		}

		pack := &player_proto.SCFishJackpotCoin{}
		if datas, exist := FishJackpotCoinMgr.Jackpot[player.Platform]; exist {
			pack.Coin = proto.Int64(datas + 234982125)
			//pack.Coin = proto.Int64(datas)
		} else {
			pack.Coin = proto.Int64(234982125) // 假数据
			//pack.Coin = proto.Int64(0) // 假数据
		}
		logger.Logger.Trace("CSFishJackpotCoinHandler Pb  ", pack)
		player.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_FISHJACKPOTDATA), pack)

		return nil
	}
	return nil
}

// 全民是否打开
type CSGetSpreadIsOpenPacketFactory struct {
}
type CSGetSpreadIsOpenHandler struct {
}

func (this *CSGetSpreadIsOpenPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSGetSpreadLWIsOpen{}
	return pack
}

func (this *CSGetSpreadIsOpenHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGetSpreadIsOpenHandler Process recv ", data)
	if _, ok := data.(*player_proto.CSGetSpreadLWIsOpen); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSGetSpreadIsOpenHandler p == nil")
			return nil
		}
		plt := PlatformMgrSington.GetPlatform(p.Platform)
		if plt == nil {

			return nil
		}

		pack := &player_proto.SCGetSpreadLWIsOpen{}
		if plt.SpreadWinLose {
			pack.IsOpen = proto.Int32(1)
		} else {
			pack.IsOpen = proto.Int32(0)
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_GetSpreadLWIsOpen), pack)

		//sendPack(player_proto.OpResultCode_OPRC_Sucess)
	}
	return nil
}

// 玩家设置 音乐音效客户端修改 目前功能只有兑换码
type CSPlayerSettingPacketFactory struct {
}
type CSPlayerSettingHandler struct {
}

func (this *CSPlayerSettingPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSPlayerSetting{}
	return pack
}

func (this *CSPlayerSettingHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPlayerSettingHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSPlayerSetting); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSPlayerSettingHandler p == nil")
			return nil
		}

		pack := &player_proto.SCPlayerSetting{
			OpRetCode: player_proto.OpResultCode_OPRC_Jyb_CodeErr,
		}

		if msg.PackageCode != "" {
			args := &model.VerifyUpJybInfoArgs{

				UseCode: msg.PackageCode,
				Plt:     p.Platform,
				SnId:    p.SnId,
			}
			if len(msg.PackageCode) < 12 { // 通用
				args.CodeType = 1
			} else {
				key, err := model.Code2Id(msg.PackageCode)
				if err != nil || key < uint64(model.Keystart) {
					pack.OpRetCode = player_proto.OpResultCode_OPRC_Jyb_CodeErr
					proto.SetDefaults(pack)
					p.SendToClient(int(player_proto.PlayerPacketID_PACKET_ALL_SETTING), pack)
					return nil
				}

				start := uint64(model.Keystart)
				args.CodeStart = int64(key / start * start) // 过滤
				args.CodeType = 2
			}
			BagMgrSington.VerifyUpJybInfo(p, args)

		} else {
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_ALL_SETTING), pack)
		}
	}
	return nil
}

func init() {
	//用户信息
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_PLAYERDATA), &CSPlayerDataHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_PLAYERDATA), &CSPlayerDataPacketFactory{})
	//查看别人信息
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_QUERYPLAYER), &CSQueryPlayerHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_QUERYPLAYER), &CSQueryPlayerPacketFactory{})
	//获取图片验证码
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_GETIMAGEVERIFYCODE), &CCSGetImageVerifyCodeHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_GETIMAGEVERIFYCODE), &CSGetImageVerifyCodePacketFactory{})
	//获取验证码
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_VERIFICATIONCODE), &CSVerificationCodePlayerHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_VERIFICATIONCODE), &CSVerificationCodePlayerPacketFactory{})
	//修改昵称信息
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_CHANGENICK), &CSPlayerChangeNickHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_CHANGENICK), &CSPlayerChangeNickPacketFactory{})
	//修改头像信息
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_CHANGEICON), &CSPlayerChangeIconHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_CHANGEICON), &CSPlayerChangeIconPacketFactory{})
	//修改头像框信息
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_HEADOUTLINE), &CSPlayerChangeHeadOutLineHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_HEADOUTLINE), &CSPlayerChangeHeadOutLinePacketFactory{})
	//修改性别信息
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_CHANGESEX), &CSPlayerChangeSexHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_CHANGESEX), &CSPlayerChangeSexPacketFactory{})
	//升级账号
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_UPGRADEACCOUNT), &CSUpgradeAccountHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_UPGRADEACCOUNT), &CSUpgradeAccountPacketFactory{})
	//绑定支付宝账号
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_BINDALIPAY), &CSBindAlipayHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_BINDALIPAY), &CSBindAlipayPacketFactory{})
	//绑定银行卡账号
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_BINDBANK), &CSBindBankHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_BINDBANK), &CSBindBankPacketFactory{})
	//更改密码
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_CHANGEPASSWORD), &CSChangePasswordHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_CHANGEPASSWORD), &CSChangePasswordPacketFactory{})
	//保险箱存取
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_PLAYEROPCOIN), &CSPlayerSafeBoxHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_PLAYEROPCOIN), &CSPlayerSafeBoxPacketFactory{})
	//读取保险箱记录
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_GETSAFEBOXCOINLOG), &CSPlayerSafeBoxCoinLogHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_GETSAFEBOXCOINLOG), &CSPlayerSafeBoxCoinLogPacketFactory{})
	//读取游戏记录
	//common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_GETGAMECOINLOG), &CSPlayerGameCoinLogHandler{})
	//netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_GETGAMECOINLOG), &CSPlayerGameCoinLogPacketFactory{})
	//注册帐号
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_REGISTER), &CSPlayerRegisterHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_REGISTER), &CSPlayerRegisterPacketFactory{})
	//获取邀请码
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_INVITECODE), &CSInviteCodePlayerHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_INVITECODE), &CSInviteCodePlayerPacketFactory{})
	//玩家API透传
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_WEBAPI_PLAYERPASS), &CSWebAPIPlayerPassHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_WEBAPI_PLAYERPASS), &CSWebAPIPlayerPassPacketFactory{})
	//系统API透传
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_WEBAPI_SYSTEMPASS), &CSWebAPISystemPassHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_WEBAPI_SYSTEMPASS), &CSWebAPISystemPassPacketFactory{})
	//绑定推广关系
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_SPREADBIND), &CSSpreadBindHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_SPREADBIND), &CSSpreadBindPacketFactory{})
	//绑定推广员
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_BINDPROMOTER), &CSBindPromoterHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_BINDPROMOTER), &CSBindPromoterPacketFactory{})
	//生成客服会话token
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_GENCUSTOMTOKEN), &CSGenCustomerTokenHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_GENCUSTOMTOKEN), &CSGenCustomerTokenPacketFactory{})
	//客服离线消息接收ack
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_CUSTOMNEWMSGACK), &CSCustomerNewMsgAckHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_CUSTOMNEWMSGACK), &CSCustomerNewMsgAckPacketFactory{})
	//ios稳定版升级标注
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_IOSINSTALLSTABLE), &CSIosInstallStableHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_IOSINSTALLSTABLE), &CSIosInstallStablePacketFactory{})
	//查询中奖
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_FISHJACKPOTCOIN), &CSFishJackpotHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_FISHJACKPOTCOIN), &CSFishJackpotPacketFactory{})
	//查询奖池金额
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_FISHJACKPOTDATA), &CSFishJackpotCoinHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_FISHJACKPOTDATA), &CSFishJackpotCoinPacketFactory{})
	//查询客损是否打开
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_GetSpreadLWIsOpen), &CSGetSpreadIsOpenHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_GetSpreadLWIsOpen), &CSGetSpreadIsOpenPacketFactory{})
	//玩家设置
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_ALL_SETTING), &CSPlayerSettingHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_ALL_SETTING), &CSPlayerSettingPacketFactory{})
}
