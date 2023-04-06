package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/activity"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core/logger"
	"strconv"
	"time"
)

var ActSignMgrSington = &ActSignMgr{
	SignConfigs: make(map[int]*server.DB_ActSign),
}

type ActSignMgr struct {
	SignConfigs map[int]*server.DB_ActSign
}

func (this *ActSignMgr) Init() {
	if this.SignConfigs == nil {
		this.SignConfigs = make(map[int]*server.DB_ActSign)
	}
	for _, v := range srvdata.PBDB_ActSignMgr.Datas.GetArr() {
		this.SignConfigs[int(v.Id)] = v
	}
}

func (this *ActSignMgr) GetConfig(id int) *server.DB_ActSign {
	signConfig, ok := this.SignConfigs[id]
	if ok {
		return signConfig
	}
	return nil
}

func (this *ActSignMgr) OnPlayerLogin(player *Player) error {
	return this.RefixedPlayerData(player)
}

func (this *ActSignMgr) OnDayChanged(player *Player) error {
	//跨天不需要
	//this.RefixedPlayerData(player)
	//this.SendSignDataToPlayer(player)
	return nil
}

func (this *ActSignMgr) RefixedPlayerData(player *Player) error {
	if player.IsRob {
		return nil
	}
	if player.SignData == nil {
		player.SignData = &model.SignData{
			SignIndex:       0,
			LastSignTickets: 0,
		}
	}
	return nil
}

func (this *ActSignMgr) SendSignDataToPlayer(player *Player) {
	if player.IsRob {
		return
	}
	pack := &activity.SCSignData{}
	//已经领取第几个
	pack.SignCount = proto.Int(player.SignData.SignIndex)
	if player.SignData.LastSignTickets != 0 {
		lastSignTime := time.Unix(player.SignData.LastSignTickets, 0)
		dayDiff := int32(common.DiffDay(time.Now(), lastSignTime))
		if dayDiff == 0 {
			pack.TodaySign = proto.Int32(1)
		} else {
			pack.TodaySign = proto.Int32(0)
		}
	} else {
		pack.TodaySign = proto.Int32(0)
	}
	proto.SetDefaults(pack)
	player.SendToClient(int(activity.ActSignPacketID_PACKET_SCSignData), pack)
	logger.Logger.Trace("SCSignData: ", pack)
}

func (this *ActSignMgr) CanSign(player *Player, signIndex int) activity.OpResultCode_ActSign {
	signConfig := this.GetConfig(signIndex)
	if signConfig == nil {
		return activity.OpResultCode_ActSign_OPRC_Activity_Sign_Error
	}

	if player.SignData.LastSignTickets != 0 {
		lastSignTime := time.Unix(player.SignData.LastSignTickets, 0)
		dayDiff := int32(common.DiffDay(time.Now(), lastSignTime))
		if dayDiff == 0 {
			if player.SignData.SignIndex == signIndex {
				return activity.OpResultCode_ActSign_OPRC_Activity_Sign_Repeat
			} else {
				return activity.OpResultCode_ActSign_OPRC_Activity_Sign_Config_Day_Error
			}
		}

		if player.SignData.SignIndex != (signIndex - 1) {
			return activity.OpResultCode_ActSign_OPRC_Activity_Sign_Config_Day_Error
		}
	} else {
		if signIndex != 1 {
			return activity.OpResultCode_ActSign_OPRC_Activity_Sign_Config_Day_Error
		}
	}

	return activity.OpResultCode_ActSign_OPRC_Activity_Sign_Sucess
}

func (this *ActSignMgr) Sign(player *Player, signIndex int, signType int32) activity.OpResultCode_ActSign {
	errCode := this.CanSign(player, signIndex)
	if errCode != activity.OpResultCode_ActSign_OPRC_Activity_Sign_Sucess {
		return errCode
	}

	signConfig := this.GetConfig(signIndex)
	if signConfig == nil {
		return activity.OpResultCode_ActSign_OPRC_Activity_Sign_Error
	}

	player.SignData.LastSignTickets = time.Now().Unix()
	player.SignData.SignIndex = signIndex

	logger.Logger.Info("签到成功: ", signConfig)
	grade := signConfig.Grade
	switch signType {
	case 0: //普通签到
	case 1: //双倍签到
		grade *= 2
	}
	switch signConfig.Type {
	case 1: //金币
		player.AddCoin(int64(grade), common.GainWay_ActSign, strconv.Itoa(int(signIndex)), time.Now().Format("2006-01-02 15:04:05"))
	case 2: //钻石
		player.AddDiamond(int64(grade), common.GainWay_ActSign, strconv.Itoa(int(signIndex)), time.Now().Format("2006-01-02 15:04:05"))
	case 3: //道具
		item := &Item{
			ItemId:  signConfig.Item_Id,
			ItemNum: grade,
		}
		BagMgrSington.AddJybBagInfo(player, []*Item{item})
		data := srvdata.PBDB_GameItemMgr.GetData(item.ItemId)
		if data != nil {
			BagMgrSington.RecordItemLog(player.Platform, player.SnId, ItemObtain, item.ItemId, data.Name, item.ItemNum, "14日签到获得")
		}
	}
	return activity.OpResultCode_ActSign_OPRC_Activity_Sign_Sucess
}

func init() {
	mgo.SetStats(true)
	RegisteParallelLoadFunc("14日签到", func() error {
		ActSignMgrSington.Init()
		return nil
	})
}
