package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	login_proto "games.yol.com/win88/protocol/login"
	player_proto "games.yol.com/win88/protocol/player"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/i18n"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

type CSLoginPacketFactory struct {
}

type CSLoginHandler struct {
}

func (this *CSLoginPacketFactory) CreatePacket() interface{} {
	pack := &login_proto.CSLogin{}
	return pack
}

func (this *CSLoginHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSLoginHandler Process recv ", data)
	if csl, ok := data.(*login_proto.CSLogin); ok {
		sendSCLogin := func(code login_proto.OpResultCode) {
			sclogin := &login_proto.SCLogin{
				OpRetCode: code,
			}
			proto.SetDefaults(sclogin)
			common.SendToGate(sid, int(login_proto.LoginPacketID_PACKET_SC_LOGIN), sclogin, s)
		}
		sendSCDisconnect := func(code int32) {
			ssDis := &login_proto.SSDisconnect{
				SessionId: proto.Int64(sid),
				Type:      proto.Int32(code),
			}
			proto.SetDefaults(ssDis)
			s.Send(int(login_proto.GatePacketID_PACKET_SS_DICONNECT), ssDis)
		}

		platform, channel, promoter, _, tagkey := PlatformMgrSington.GetPlatformByPackageTag(csl.GetPlatformTag())
		//测试token登录
		//csl.Token = proto.String(`mkFzhIuEhpGAjGiDQVlBVE9UU1JQVVdWU0FLQW+AkpKWjpGDQVlBgVZTVYBXV1aEg1CFU09PUIRRWIJRgViFVIKBUYOAgFdBS0FvgIKKgIaEk4CGQVlBQUtBZJePiJGEg0FZUFVUUlRWU1ZYUpwhIQ==`)
		token := ""
		if csl.GetToken() != "" {
			token = csl.GetToken()

			err, tu := common.VerifyTokenAes(token)
			if err != nil {
				sendSCLogin(login_proto.OpResultCode_OPRC_Error)
				sendSCDisconnect(common.KickReason_Freeze)
				return nil
			}
			csl.Username = proto.String(tu.TelegramId)
			csl.Password = proto.String(tu.Password)
			csl.PlatformTag = proto.String(tu.Packagetag)
			csl.LoginType = proto.Int32(5)
			platformid, perr := strconv.Atoi(csl.GetPlatformTag())
			if perr != nil {
				sendSCLogin(login_proto.OpResultCode_OPRC_Error)
				sendSCDisconnect(common.KickReason_Freeze)
				return nil
			}
			platform = int32(platformid)
		}
		if csl.GetUsername() == "" || csl.GetPassword() == "" {
			sendSCLogin(login_proto.OpResultCode_OPRC_Error)
			sendSCDisconnect(common.KickReason_Freeze)
			return nil
		}
		//替换为后台拿到的配置数据，不再使用客户端的发送数据
		backupPromoter := ""

		if csl.GetChannel() != common.Channel_Rob {
			if platform != 0 {
				csl.Platform = proto.String(strconv.Itoa(int(platform)))
			} else {
				csl.Platform = proto.String(Default_Platform)
				platform = Default_PlatformInt
			}
			if channel != 0 {
				csl.Channel = proto.String(strconv.Itoa(int(channel)))
			} else {
				csl.Channel = proto.String("")
			}
			if len(csl.GetPromoter()) <= 0 {
				if promoter != 0 {
					csl.Promoter = proto.String(strconv.Itoa(int(promoter)))
				} else {
					csl.Promoter = proto.String("")
				}
			} else {
				backupPromoter = csl.GetPromoter()
			}
			//InviterId 和 PromoterTree客户端是从剪贴板上读取出来的，所以不以包上的为准
		}

		pt := PlatformMgrSington.GetPlatform(strconv.Itoa(int(platform)))
		if pt == nil || pt.Disable {
			sendSCLogin(login_proto.OpResultCode_OPRC_SceneServerMaintain)
			sendSCDisconnect(common.KickReason_Freeze)
			return nil
		}

		logger.Logger.Tracef("platform:%v,channel:%v,promoter%v.", platform, channel, promoter)
		//是否正在维护
		if model.GameParamData.SrvMaintain && SrvIsMaintaining {
			inWhiteList := false
			for i := 0; i < len(model.GMACData.WhiteList); i++ {
				if model.GMACData.WhiteList[i] == csl.GetUsername() {
					inWhiteList = true
					break
				}
			}
			//排除白名单里的玩家
			if !inWhiteList {
				sendSCLogin(login_proto.OpResultCode_OPRC_SceneServerMaintain)
				sendSCDisconnect(common.KickReason_Freeze)
				return nil
			}
		}

		//检查版本号
		if model.GameParamData.VerifyClientVersion {
			deviceOs := csl.GetDeviceOs()
			packVers := srvdata.GetPackVers(csl.GetPlatformTag())
			if deviceOs != "" && packVers != nil {
				if cvers, ok := packVers[deviceOs]; ok {
					if csl.GetResVer() < cvers.MinResVer {
						sendSCLogin(login_proto.OpResultCode_OPRC_YourResVerIsLow)
						sendSCDisconnect(common.KickReason_ResLow)
						return nil
					}

					if csl.GetApkVer() < cvers.MinApkVer {
						sendSCLogin(login_proto.OpResultCode_OPRC_YourAppVerIsLow)
						sendSCDisconnect(common.KickReason_AppLow)
						return nil
					}
				}
			}
		}
		//
		state := LoginStateMgrSington.GetLoginStateOfSid(sid)
		if state != nil {
			logger.Logger.Warnf("CSLoginHandler relogining (%v) repeated (%v) ", csl.GetUsername(), sid)
			return nil
		}

		username := csl.GetUsername()
		password := csl.GetPassword()
		if LoginStateMgrSington.IsLogining(username, csl.GetPlatform(), tagkey) {
			logger.Logger.Warnf("CSLoginHandler logining (%v) disconnect current(%v) ", csl.GetUsername(), sid)
			//todo:disconnect
			sendSCDisconnect(common.KickReason_Logining)
			return nil
		}
		if csl.GetLoginType() != 5 {
			raw := fmt.Sprintf("%v%v%v%v%v", username, password, csl.GetTimeStamp(), csl.GetParams(), common.GetAppId())
			h := md5.New()
			io.WriteString(h, raw)
			hashsum := hex.EncodeToString(h.Sum(nil))
			if hashsum != csl.GetSign() {
				logger.Logger.Tracef("ClientSessionAttribute_State hashsum not fit!!! get:%v expect:%v rawstr:%v", csl.GetSign(), hashsum, raw)
				sendSCLogin(login_proto.OpResultCode_OPRC_Error)
				sendSCDisconnect(common.KickReason_CheckCodeErr)
				return nil
			}
		}

		var temp []byte
		var ee error
		di := csl.GetDeviceInfo()
		if false && pt.NeedDeviceInfo && di == "" { //maybe res ver is low
			sendSCLogin(login_proto.OpResultCode_OPRC_YourResVerIsLow)
			sendSCDisconnect(common.KickReason_ResLow)
			return nil
		}

		if di != "" {
			var e common.Encryptor
			e.Init(common.GetAppId(), csl.GetPlatformTag(), int32(csl.GetTimeStamp()))
			temp, ee = base64.StdEncoding.DecodeString(di)
			if ee != nil {
				sendSCLogin(login_proto.OpResultCode_OPRC_YourResVerIsLow)
				sendSCDisconnect(common.KickReason_Freeze)
				return nil
			}

			e.Encrypt(temp, len(temp))

			if model.GameParamData.ValidDeviceInfo && !json.Valid(temp) {
				sendSCLogin(login_proto.OpResultCode_OPRC_YourResVerIsLow)
				sendSCDisconnect(common.KickReason_Freeze)
				return nil
			}
		}
		if len(csl.DeviceId) != 0 {
			if bi, ok := BlackListMgrSington.CheckDeviceInBlack(csl.DeviceId, BlackState_Login, csl.GetPlatform()); ok {
				if bi != nil {
					common.SendToGate(sid, int(player_proto.PlayerPacketID_PACKET_SC_SRVMSG), common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, i18n.Tr("cn", "BlackListLimit1Args", bi.Id)), s)
				}
				sendSCLogin(login_proto.OpResultCode_OPRC_YourResVerIsLow)
				sendSCDisconnect(common.KickReason_Freeze)
				return nil
			}
		}
		//deviceInfo := string(temp)
		csl.DeviceInfo = proto.String(string(temp))
		clog := &model.ClientLoginInfo{
			LoginType:    csl.GetLoginType(),
			ApkVer:       csl.GetApkVer(),
			ResVer:       csl.GetResVer(),
			InviterId:    csl.GetInviterId(),
			PromoterTree: csl.GetPromoterTree(),
			UserName:     csl.GetUsername(),
			PlatformTag:  csl.GetPlatformTag(),
			Promoter:     csl.GetPromoter(),
			Sid:          sid,
		}

		if LoginStateMgrSington.StartLogin(username, csl.GetPlatform(), sid, s, clog, tagkey) {
			tl := &TaskLogin{CSLogin: csl, Session: s, Sid: sid, BackupPromoter: backupPromoter, tagkey: tagkey, token: token}
			t := task.New(nil, tl, tl, "TaskLogin")
			if b := t.StartByExecutor(username); !b {
				logger.Logger.Trace("login task lauch failed")
				//todo:disconnect
				sendSCDisconnect(common.KickReason_TaskErr)
			}
		} else { //用缓存信息做登录
			sclogin := &login_proto.SCLogin{}
			name := fmt.Sprintf("%v_%v_%v", username, csl.GetPlatform(), tagkey)
			ls := LoginStateMgrSington.GetLoginStateByName(name)
			if ls != nil {
				acc := ls.acc
				if acc == nil {
					sendSCLogin(login_proto.OpResultCode_OPRC_Error)
					sendSCDisconnect(common.KickReason_CantFindAcc)
					//清理登录状态
					LoginStateMgrSington.LogoutBySid(sid)
					return nil
				}
				//freeze
				if acc.State > time.Now().Unix() {
					sendSCLogin(login_proto.OpResultCode_OPRC_AccountBeFreeze)
					sendSCDisconnect(common.KickReason_Freeze)
					//清理登录状态
					LoginStateMgrSington.LogoutBySid(sid)
					return nil
				}

				pwdIsErr := true
				switch csl.GetLoginType() {
				case 0: //游客登录
					if acc.UserName == csl.GetUsername() && acc.Platform == csl.GetPlatform() && acc.TagKey == tagkey {
						raw := fmt.Sprintf("%v%v%v", acc.PassWord, common.GetAppId(), csl.GetTimeStamp())
						h := md5.New()
						io.WriteString(h, raw)
						pwd := hex.EncodeToString(h.Sum(nil))
						if pwd != csl.GetPassword() {
							pwdIsErr = true
						} else {
							pwdIsErr = false
						}
					}
				case 1: //帐号登录
					if acc.UserName == csl.GetUsername() && acc.Platform == csl.GetPlatform() && acc.TagKey == tagkey {
						raw := fmt.Sprintf("%v%v%v", acc.TelPassWord, common.GetAppId(), csl.GetTimeStamp())
						ht := md5.New()
						io.WriteString(ht, raw)
						pwd := hex.EncodeToString(ht.Sum(nil))
						if pwd != csl.GetPassword() {
							pwdIsErr = true
						} else {
							pwdIsErr = false
						}
					}
				case 5:
					if acc.UserName == csl.GetUsername() && acc.Platform == csl.GetPlatform() && acc.TagKey == tagkey {
						if acc.PassWord != csl.GetPassword() {
							pwdIsErr = true
						} else {
							pwdIsErr = false
						}
					}
				}

				//密码错误
				if pwdIsErr {
					if csl.GetChannel() != common.Channel_Rob {
						//try from db login
						tl := &TaskLogin{CSLogin: csl, Session: s, Sid: sid, token: token}
						t := task.New(nil, tl, tl, "TaskLogin")
						if b := t.StartByExecutor(username); !b {
							logger.Logger.Trace("login task lauch failed")
							//todo:disconnect
							sendSCDisconnect(common.KickReason_DBLoadAcc)
						}
						return nil
					}
					sendSCLogin(login_proto.OpResultCode_OPRC_LoginPassError)
					//清理登录状态
					LoginStateMgrSington.LogoutBySid(sid)
					return nil
				}

				sclogin.AccId = proto.String(acc.AccountId.Hex())
				sclogin.SrvTs = proto.Int64(time.Now().Unix())
				//获取平台名称
				plf := acc.Platform
				if plf != common.Platform_Rob {
					deviceOs := csl.GetDeviceOs()
					packVers := srvdata.GetPackVers(csl.GetPlatformTag())
					if deviceOs != "" && packVers != nil {
						if cvers, ok := packVers[deviceOs]; ok {
							sclogin.MinApkVer = proto.Int32(cvers.MinResVer)
							sclogin.LatestApkVer = proto.Int32(cvers.LatestApkVer)
							sclogin.MinApkVer = proto.Int32(cvers.MinApkVer)
							sclogin.LatestResVer = proto.Int32(cvers.LatestResVer)
						}
					}
					gameVers := srvdata.GetPackVers(csl.GetPlatformTag())
					if gameVers != nil {
						for k, v := range gameVers {
							token := strings.Split(k, ",")
							if len(token) == 2 && token[1] == deviceOs {
								if gameId, err := strconv.Atoi(token[0]); err == nil {
									sclogin.SubGameVer = append(sclogin.SubGameVer, &login_proto.GameVer{
										GameId:       proto.Int(gameId),
										MinApkVer:    proto.Int32(v.MinApkVer),
										LatestApkVer: proto.Int32(v.LatestApkVer),
										MinResVer:    proto.Int32(v.MinResVer),
										LatestResVer: proto.Int32(v.LatestResVer),
									})
								}
							}
						}
					}

					//加载配置
					gps := PlatformMgrSington.GetPlatformGameConfig(plf)
					for _, v := range gps {
						if v.Status {
							if v.DbGameFree.GetGameRule() != 0 {
								//为了防止一个协议数据量太大，所以简化成开启的游戏id

								lgi := &login_proto.LoginGameInfo{
									GameId:  proto.Int32(v.DbGameFree.GameId),
									LogicId: proto.Int32(v.DbGameFree.Id),
								}
								if v.DbGameFree.GetLottery() != 0 { //彩金池
									//_, gl := LotteryMgrSington.FetchLottery(plf, v.DBGameFree.GetId(), v.DBGameFree.GetGameId())
									//if gl != nil {
									//	lgi.LotteryCoin = proto.Int64(gl.Value)
									//}
								}
								sclogin.GameInfo = append(sclogin.GameInfo, lgi)
							} else {
								//为了防止一个协议数据量太大，所以简化成开启的游戏id
								//sclogin.ThrGameCfg = append(sclogin.ThrGameCfg, &proto_login.LoginThrGameConfig{
								//	LogicId:   proto.Int32(v.LogicId),
								//	LimitCoin: proto.Int32(v.DBGameFree.GetLimitCoin()),
								//})
								sclogin.ThrGameId = append(sclogin.ThrGameId, v.DbGameFree.Id)
							}
						}
					}
				}
				acc.LastLoginTime = time.Now()
				acc.LoginTimes++
				sclogin.OpRetCode = login_proto.OpResultCode_OPRC_Sucess
				proto.SetDefaults(sclogin)
				common.SendToGate(sid, int(login_proto.LoginPacketID_PACKET_SC_LOGIN), sclogin, s)

				lss := LoginStateMgrSington.Logined(csl.GetUsername(), csl.GetPlatform(), sid, acc, tagkey)
				if len(lss) != 0 {
					for k, ls := range lss {
						//todo:顶号
						//todo:gate->disconnect
						ssDis := &login_proto.SSDisconnect{
							SessionId: proto.Int64(k),
							Type:      proto.Int32(common.KickReason_OtherLogin),
						}
						proto.SetDefaults(ssDis)
						s.Send(int(login_proto.GatePacketID_PACKET_SS_DICONNECT), ssDis)
						logger.Logger.Warnf("==========顶号 oldsid:%v newsid:%v", k, sid)
						LoginStateMgrSington.Logout(ls)
						//todo:game->rehold
					}
				}
			}
		}
	}
	return nil
}

type CSCustomServicePacketFactory struct {
}

type CSCustomServiceHandler struct {
}

func (this *CSCustomServicePacketFactory) CreatePacket() interface{} {
	pack := &login_proto.CSCustomService{}
	return pack
}

func (this *CSCustomServiceHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSCustomService Process recv ", data)
	if _, ok := data.(*login_proto.CSCustomService); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSCustomServiceHandler p == nil")
			return nil
		}
		var Url string
		var customType int32
		flag := int32(0)
		platform := PlatformMgrSington.GetPlatform(p.Platform)
		if platform != nil {
			Url = platform.ServiceUrl
			customType = platform.CustomType
			if platform.ServiceFlag {
				flag = 1
			}
		}
		pack := &login_proto.SCCustomService{
			Url:        proto.String(Url),
			OpenFlag:   proto.Int32(flag),
			CustomType: proto.Int32(customType),
		}
		p.SendToClient(int(login_proto.LoginPacketID_PACKET_SC_CUSTOMSERVICE), pack)
		logger.Logger.Trace("SCCustomService:", pack)
	}
	return nil
}

type CSPlatFormPacketFactory struct {
}

type CSPlatFormHandler struct {
}

func (this *CSPlatFormPacketFactory) CreatePacket() interface{} {
	pack := &login_proto.CSPlatFormConfig{}
	return pack
}

func (this *CSPlatFormHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPlatFormHandler Process recv ", data)
	if _, ok := data.(*login_proto.CSPlatFormConfig); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSPlatFormHandler p == nil")
			return nil
		}

		//platformId, _, _ := PlatformMgrSington.GetPlatformByPackageTag(msg.GetPlatformTag())
		platform := PlatformMgrSington.GetPlatform(p.Platform)
		if platform == nil {
			scPlatForm := &login_proto.SCPlatFormConfig{
				OpRetCode: login_proto.OpResultCode_OPRC_Error,
			}
			proto.SetDefaults(scPlatForm)
			p.SendToClient(int(login_proto.LoginPacketID_PACKET_SC_PLATFORMCFG), scPlatForm)
			return nil
		}

		scPlatForm := &login_proto.SCPlatFormConfig{
			Platform:               proto.String(p.Platform),
			OpRetCode:              login_proto.OpResultCode_OPRC_Sucess,
			UpgradeAccountGiveCoin: proto.Int32(p.GetUpdateAccPrize()),
			VipRange:               platform.VipRange,
			ExchangeMin:            proto.Int32(platform.ExchangeMin),
			ExchangeLimit:          proto.Int32(platform.ExchangeLimit),
			OtherParams:            proto.String(platform.OtherParams),
			SpreadConfig:           proto.Int32(platform.SpreadConfig),
			ExchangeTax:            proto.Int32(platform.ExchangeTax),
			ExchangeFlow:           proto.Int32(GetExchangeFlow(p.PlayerData)),
			ExchangeBankMax:        proto.Int32(platform.ExchangeBankMax),
			ExchangeAlipayMax:      proto.Int32(platform.ExchangeAlipayMax),
			ExchangeMultiple:       proto.Int32(platform.ExchangeMultiple),
		}
		rebateTask := RebateInfoMgrSington.rebateTask[p.Platform]
		if rebateTask != nil {
			scPlatForm.Rebate = &login_proto.RebateCfg{
				RebateSwitch:   proto.Bool(rebateTask.RebateSwitch),
				ReceiveMode:    proto.Int32(int32(rebateTask.ReceiveMode)),
				NotGiveOverdue: proto.Int32(int32(rebateTask.NotGiveOverdue)),
			}
		}
		if platform.ClubConfig != nil { //俱乐部配置
			scPlatForm.Club = &login_proto.ClubCfg{
				IsOpenClub:              proto.Bool(platform.ClubConfig.IsOpenClub),
				CreationCoin:            proto.Int64(platform.ClubConfig.CreationCoin),
				IncreaseCoin:            proto.Int64(platform.ClubConfig.IncreaseCoin),
				ClubInitPlayerNum:       proto.Int32(platform.ClubConfig.ClubInitPlayerNum),
				CreateClubCheckByManual: proto.Bool(platform.ClubConfig.CreateClubCheckByManual),
				EditClubNoticeByManual:  proto.Bool(platform.ClubConfig.EditClubNoticeByManual),
				CreateRoomAmount:        proto.Int64(platform.ClubConfig.CreateRoomAmount),
				GiveCoinRate:            platform.ClubConfig.GiveCoinRate,
			}
		}

		proto.SetDefaults(scPlatForm)
		p.SendToClient(int(login_proto.LoginPacketID_PACKET_SC_PLATFORMCFG), scPlatForm)
	}
	return nil
}

// 公告信息
type CSBulletionInfoPacketFactory struct {
}

type CSBulletionInfoHandler struct {
}

func (this *CSBulletionInfoPacketFactory) CreatePacket() interface{} {
	pack := &login_proto.CSBulletionInfo{}
	return pack
}
func (this *CSBulletionInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSBulletionInfoHandler Process recv ", data)
	if msg, ok := data.(*login_proto.CSBulletionInfo); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Trace("CSBulletionInfoHandler p == nil ")
			return nil
		}
		pf, _, _, _, _ := PlatformMgrSington.GetPlatformByPackageTag(msg.GetPlatformTag())
		platform := strconv.Itoa(int(pf))
		i := 0
		var bulletList []*login_proto.Bulletion
		for _, v := range BulletMgrSington.BulletMsgList {
			if v != nil && platform == v.Platform && v.State == 1 {
				bulletList = append(bulletList, &login_proto.Bulletion{})
				bulletList[i].Id = proto.Int32(v.Id)
				bulletList[i].NoticeTitle = proto.String(v.NoticeTitle)
				bulletList[i].NoticeContent = proto.String(v.NoticeContent)
				bulletList[i].UpdateTime = proto.String(v.UpdateTime)
				bulletList[i].Sort = proto.Int32(v.Sort)

				i++
			}
		}

		var rawpack = &login_proto.SCBulletionInfo{
			Id:            proto.Int(0),
			BulletionList: bulletList,
		}
		proto.SetDefaults(rawpack)
		p.SendToClient(int(login_proto.LoginPacketID_PACKET_SC_BULLETIONINFO), rawpack)

	}
	return nil
}

// 招商信息 CSCustomerInfoList
type CSCustomerInfoListPacketFactory struct {
}

type CSCustomerInfoListHandler struct {
}

func (this *CSCustomerInfoListPacketFactory) CreatePacket() interface{} {
	pack := &login_proto.CSCustomerInfoList{}
	return pack
}
func (this *CSCustomerInfoListHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSCustomerInfoListHandler Process recv ", data)
	if _, ok := data.(*login_proto.CSCustomerInfoList); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Trace("CSBulletionInfoHandler p == nil ")
			return nil
		}
		var customerList []*login_proto.Customer
		i := 0
		for _, v := range CustomerMgrSington.CustomerMsgList {
			if v != nil && v.Platform == p.Platform && v.Status == 1 {
				customerList = append(customerList, &login_proto.Customer{})
				customerList[i].Id = proto.Int32(v.Id)
				customerList[i].Nickname = proto.String(v.Nickname)
				customerList[i].Headurl = proto.String(v.Headurl)
				customerList[i].Ext = proto.String(v.Ext)
				customerList[i].WeixinAccount = proto.String(v.Weixin_account)
				customerList[i].QqAccount = proto.String(v.Qq_account)
				i++
			}
		}
		var rawpack = &login_proto.SCCustomerInfoList{
			CustomerList: customerList,
		}
		proto.SetDefaults(rawpack)
		p.SendToClient(int(login_proto.LoginPacketID_PACKET_SC_CUSTOMERINFOLIST), rawpack)

	}
	return nil
}

type CSVerifyTypePacketFactory struct {
}

type CSVerifyTypeHandler struct {
}

func (this *CSVerifyTypePacketFactory) CreatePacket() interface{} {
	pack := &login_proto.CSVerifyType{}
	return pack
}
func (this *CSVerifyTypeHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSVerifyTypeHandler Process recv ", data)
	if msg, ok := data.(*login_proto.CSVerifyType); ok {
		pf, _, _, _, _ := PlatformMgrSington.GetPlatformByPackageTag(msg.GetPlatformTag())
		platform := PlatformMgrSington.GetPlatform(strconv.Itoa(int(pf)))
		pack := &login_proto.SCVerifyType{}
		if platform == nil {
			pack.OpRetCode = login_proto.OpResultCode_OPRC_Error
		} else {
			pack.VerifyType = proto.Int32(platform.VerifyCodeType)
			pack.OpRetCode = login_proto.OpResultCode_OPRC_Sucess
			if platform.VerifyCodeType != common.CodeTypeNo && !reTelRule.MatchString(msg.GetTel()) {
				pack := &login_proto.SCVerifyType{
					OpRetCode: login_proto.OpResultCode_OPRC_TelError,
				}
				common.SendToGate(sid, int(login_proto.LoginPacketID_PACKET_SC_VERIFYTYPE), pack, s)
				return nil
			}
		}
		common.SendToGate(sid, int(login_proto.LoginPacketID_PACKET_SC_VERIFYTYPE), pack, s)
	}
	return nil
}

type CSRegisterVerifyTypePacketFactory struct {
}

type CSRegisterVerifyTypeHandler struct {
}

func (this *CSRegisterVerifyTypePacketFactory) CreatePacket() interface{} {
	pack := &login_proto.CSRegisterVerifyType{}
	return pack
}
func (this *CSRegisterVerifyTypeHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSRegisterVerifyTypeHandler Process recv ", data)
	if msg, ok := data.(*login_proto.CSRegisterVerifyType); ok {
		pf, _, _, _, _ := PlatformMgrSington.GetPlatformByPackageTag(msg.GetPlatformTag())
		platform := PlatformMgrSington.GetPlatform(strconv.Itoa(int(pf)))
		pack := &login_proto.SCRegisterVerifyType{}
		if platform == nil {
			pack.OpRetCode = login_proto.OpResultCode_OPRC_Error
		} else {
			if platform.RegisterVerifyCodeSwitch {
				pack.VerifyType = proto.Int32(common.CodeTypeNo)
			} else {
				pack.VerifyType = proto.Int32(platform.VerifyCodeType)
			}
			pack.OpRetCode = login_proto.OpResultCode_OPRC_Sucess
		}
		common.SendToGate(sid, int(login_proto.LoginPacketID_PACKET_SC_REGISTERVERIFYTYPE), pack, s)
	}
	return nil
}

// 获取三方游戏配置
type CSGetThrGameCfgPacketFactory struct {
}

type CSGetThrGameCfgHandler struct {
}

func (this *CSGetThrGameCfgPacketFactory) CreatePacket() interface{} {
	pack := &login_proto.CSGetThrGameCfg{}
	return pack
}
func (this *CSGetThrGameCfgHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGetThrGameCfgHandler Process recv ", data)
	if msg, ok := data.(*login_proto.CSGetThrGameCfg); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSGetThrGameCfgHandler p == nil")
			return nil
		}
		pack := &login_proto.SCGetThrGameCfg{}
		plf := msg.GetPlatform()
		//加载配置
		gps := PlatformMgrSington.GetPlatformGameConfig(plf)
		for _, v := range gps {
			if v.Status {
				if v.DbGameFree.GetGameRule() == 0 {
					pack.ThrGameCfg = append(pack.ThrGameCfg, &login_proto.LoginThrGameConfig{
						LogicId:   proto.Int32(v.DbGameFree.Id),
						LimitCoin: proto.Int32(v.DbGameFree.GetLimitCoin()),
					})
				}
			}
		}
		p.SendToClient(int(login_proto.LoginPacketID_PACKET_SC_GETTHRGAMECFG), pack)
	}
	return nil
}

func init() {
	//登录
	common.RegisterHandler(int(login_proto.LoginPacketID_PACKET_CS_LOGIN), &CSLoginHandler{})
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_CS_LOGIN), &CSLoginPacketFactory{})
	//客服地址
	common.RegisterHandler(int(login_proto.LoginPacketID_PACKET_CS_CUSTOMSERVICE), &CSCustomServiceHandler{})
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_CS_CUSTOMSERVICE), &CSCustomServicePacketFactory{})
	//平台配置信息
	common.RegisterHandler(int(login_proto.LoginPacketID_PACKET_CS_PLATFORMCFG), &CSPlatFormHandler{})
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_CS_PLATFORMCFG), &CSPlatFormPacketFactory{})
	//公告信息
	common.RegisterHandler(int(login_proto.LoginPacketID_PACKET_CS_BULLETIONINFO), &CSBulletionInfoHandler{})
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_CS_BULLETIONINFO), &CSBulletionInfoPacketFactory{})
	//招商信息
	common.RegisterHandler(int(login_proto.LoginPacketID_PACKET_CS_CUSTOMERINFOLIST), &CSCustomerInfoListHandler{})
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_CS_CUSTOMERINFOLIST), &CSCustomerInfoListPacketFactory{})

	// 获取验证码配置
	common.RegisterHandler(int(login_proto.LoginPacketID_PACKET_CS_VERIFYTYPE), &CSVerifyTypeHandler{})
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_CS_VERIFYTYPE), &CSVerifyTypePacketFactory{})

	// 获取登录验证码类型
	common.RegisterHandler(int(login_proto.LoginPacketID_PACKET_CS_REGISTERVERIFYTYPE), &CSRegisterVerifyTypeHandler{})
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_CS_REGISTERVERIFYTYPE), &CSRegisterVerifyTypePacketFactory{})

	//获取三方游戏配置
	common.RegisterHandler(int(login_proto.LoginPacketID_PACKET_CS_GETTHRGAMECFG), &CSGetThrGameCfgHandler{})
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_CS_GETTHRGAMECFG), &CSGetThrGameCfgPacketFactory{})
}
