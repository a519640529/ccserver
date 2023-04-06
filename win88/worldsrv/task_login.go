package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	server_proto "games.yol.com/win88/protocol/server"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/idealeak/goserver/core/logger"

	"games.yol.com/win88/srvdata"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	login_proto "games.yol.com/win88/protocol/login"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

const (
	LoginFlag_HasAccount    = 1
	LoginFlag_NotAccount    = 2
	LoginFlag_PasswordError = 3
	LoginFlag_CreateError   = 4
	LoginFlag_CreateSuccess = 5
)

type TaskLogin struct {
	*login_proto.CSLogin
	*netlib.Session
	Sid            int64
	BackupPromoter string
	flag           login_proto.OpResultCode
	tagkey         int32
	token          string
}

// in task.Worker goroutine
func (t *TaskLogin) Call(o *basic.Object) interface{} {
	veriftoken := false
	if t.token != "" {
		veriftoken = true
	}
	var acc *model.Account
	var retCode int
	if t.LoginType == 6 {
		acc, retCode = model.AccountCheck(t.GetUsername(), t.GetUsername(), t.GetPassword(), t.GetPlatform(), t.GetTimeStamp(), t.GetLoginType(), t.tagkey, veriftoken)
		if retCode == 2 {
			raw := fmt.Sprintf("%v%v", t.GetUsername(), common.GetAppId())
			h := md5.New()
			io.WriteString(h, raw)
			pwd := hex.EncodeToString(h.Sum(nil))
			var err int
			acc, err = model.InsertTelAccount(t.Username, pwd, t.GetPlatform(), t.GetChannel(), t.GetPromoter(), t.GetParams(), t.GetInviterId(), t.GetPromoterTree(), "", pwd, t.GetPlatformTag(), t.GetPackage(), t.DeviceInfo, t.tagkey)
			if err == 5 {
				pi, tf := model.CreatePlayerDataByThird(acc.Platform, acc.AccountId.Hex(), t.Name, t.HeadUrl)
				if pi == nil || tf == false {
					return nil
				}
				retCode = 1
			}
		}
	} else {
		acc, retCode = model.AccountIsExist(t.GetUsername(), t.GetUsername(), t.GetPassword(), t.GetPlatform(), t.GetTimeStamp(), t.GetLoginType(), t.tagkey, veriftoken)
	}
	switch retCode {
	case 1:
		t.flag = login_proto.OpResultCode_OPRC_Sucess
	case 2:
		//新帐号,游客
		raw := fmt.Sprintf("%v%v", t.GetUsername(), common.GetAppId())
		h := md5.New()
		io.WriteString(h, raw)
		pwd := hex.EncodeToString(h.Sum(nil))

		//校验一下客户端传的IP地址，如果不合法就设置为空字符串
		//本来考虑在这里做校验，考虑到为了保存客户端发来的原始信息则补在这里做验证。
		// 在写入playerinfo和登录日志表时（func (this *PlayerData) UpdateParams(params string)）还会再次做验证。
		//var pp = &model.PlayerParams{}
		//if len(t.GetParams()) != 0 {
		//	err := json.Unmarshal([]byte(t.GetParams()), pp)
		//	if err == nil {
		//		if !common.IsValidIP(pp.Ip) {
		//			pp.Ip = ""
		//			b, _ := json.Marshal(pp)
		//			t.Params = proto.String(string(b))
		//		}
		//	}
		//}

		acc, retCode = model.InsertAccount(t.GetUsername(), pwd, t.GetPlatform(), t.GetChannel(), t.GetPromoter(), t.GetParams(),
			t.GetDeviceOs(), t.GetInviterId(), t.GetPromoterTree(), t.GetPlatformTag(), t.GetPackage(), t.GetDeviceInfo(), t.BackupPromoter, t.tagkey)
		if retCode == 5 {
			t.flag = login_proto.OpResultCode_OPRC_Sucess
		} else {
			t.flag = login_proto.OpResultCode_OPRC_Login_CreateAccError
		}
	case 3:
		t.flag = login_proto.OpResultCode_OPRC_LoginPassError
	case 4:
		t.flag = login_proto.OpResultCode_OPRC_AccountBeFreeze
	}

	//二次推广用户处理
	if acc != nil && acc.Flag != 0 {
		acc.Remark = fmt.Sprintf("%v|%v|%v|%v|%v|%v|%v|%v|%v;%v", acc.Package, acc.PackegeTag, acc.Platform, acc.Channel, acc.InviterId, acc.Promoter, acc.PromoterTree, acc.Flag, time.Now().Unix(), acc.Remark)
		acc.Flag = 0
		acc.Platform = t.GetPlatform()
		acc.Package = t.GetPackage()
		acc.PackegeTag = t.GetPlatformTag()
		acc.Channel = t.GetChannel()
		acc.InviterId = t.GetInviterId()
		acc.Promoter = t.GetPromoter()
		acc.PromoterTree = t.GetPromoterTree()
		err := model.UpdateAccount(acc)
		if err != nil {
			logger.Logger.Warnf("UpdateAccount(%v) err:%v", acc, err)
		}

		err = model.UpdatePlayerPackageIdByAcc(acc.AccountId.Hex(), t.GetPlatformTag(), t.GetPlatform(), t.GetChannel(), t.GetPromoter(), t.GetInviterId(), t.GetPromoterTree())
		if err != nil {
			logger.Logger.Warnf("UpdatePlayerPackageIdByAcc(%v) err:%v", acc, err)
		}
	}
	return acc
}

// in laucher goroutine
func (t *TaskLogin) Done(i interface{}, tt task.Task) {
	sclogin := &login_proto.SCLogin{}
	var acc *model.Account
	if v, ok := i.(*model.Account); ok && v != nil {
		acc = v
		sclogin.AccId = proto.String(acc.AccountId.Hex())
		sclogin.SrvTs = proto.Int64(time.Now().Unix())
		if t.flag == login_proto.OpResultCode_OPRC_Sucess {
			//获取平台名称
			plf := v.Platform
			if plf != common.Platform_Rob {
				deviceOs := t.GetDeviceOs()
				packVers := srvdata.GetPackVers(t.GetPlatformTag())
				if deviceOs != "" && packVers != nil {
					if cvers, ok := packVers[deviceOs]; ok {
						sclogin.MinApkVer = proto.Int32(cvers.MinResVer)
						sclogin.LatestApkVer = proto.Int32(cvers.LatestApkVer)
						sclogin.MinApkVer = proto.Int32(cvers.MinApkVer)
						sclogin.LatestResVer = proto.Int32(cvers.LatestResVer)
					}
				}
				gameVers := srvdata.GetPackVers(t.GetPlatformTag())
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
							lgi := &login_proto.LoginGameInfo{
								GameId:  proto.Int32(v.DbGameFree.GetGameId()),
								LogicId: proto.Int32(v.DbGameFree.GetId()),
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
							//sclogin.ThrGameCfg = append(sclogin.ThrGameCfg, &login_proto.LoginThrGameConfig{
							//	LogicId:   proto.Int32(v.LogicId),
							//	LimitCoin: proto.Int32(v.DBGameFree.GetLimitCoin()),
							//})
							sclogin.ThrGameId = append(sclogin.ThrGameId, v.DbGameFree.Id)
						}
					}
				}
			}
		}
		acc.LastLoginTime = time.Now()
		acc.LoginTimes++
	}
	sclogin.OpRetCode = t.flag
	proto.SetDefaults(sclogin)
	common.SendToGate(t.Sid, int(login_proto.LoginPacketID_PACKET_SC_LOGIN), sclogin, t.Session)
	if t.flag == login_proto.OpResultCode_OPRC_AccountBeFreeze {
		//todo:disconnect
		ssDis := &login_proto.SSDisconnect{
			SessionId: proto.Int64(t.Sid),
			Type:      proto.Int32(common.KickReason_Freeze),
		}
		proto.SetDefaults(ssDis)
		t.Send(int(login_proto.GatePacketID_PACKET_SS_DICONNECT), ssDis)
	}

	if t.flag == login_proto.OpResultCode_OPRC_Sucess {
		lss := LoginStateMgrSington.Logined(t.GetUsername(), t.GetPlatform(), t.Sid, acc, t.tagkey)
		if len(lss) != 0 {
			for k, ls := range lss {
				//todo:顶号
				//todo:gate->disconnect
				ssDis := &login_proto.SSDisconnect{
					SessionId: proto.Int64(k),
					Type:      proto.Int32(common.KickReason_OtherLogin),
				}
				proto.SetDefaults(ssDis)
				t.Send(int(server_proto.SSPacketID_PACKET_SS_DICONNECT), ssDis)
				logger.Logger.Warnf("==========顶号 oldsid:%v newsid:%v", k, t.Sid)
				LoginStateMgrSington.Logout(ls)
				//todo:game->rehold
			}
		}
	} else {
		//清理登录状态
		LoginStateMgrSington.LogoutBySid(t.Sid)
	}
}
