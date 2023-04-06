package caothap

import (
	"encoding/json"
	proto_gamehall "games.yol.com/win88/protocol/gamehall"
	"math/rand"
	"time"

	"games.yol.com/win88/gamerule/caothap"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/proto"
	//proto_caothap "games.yol.com/win88/protocol/caothap"
	proto_server "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
)

type CaoThapJackpot struct {
	createdTime time.Time
	userName    string
	priceValue  int64
	roomID      int64
}

type CaoThapSceneData struct {
	*base.Scene                                      //房间信息
	players             map[int32]*CaoThapPlayerData //玩家信息
	jackpot             *base.XSlotJackpotPool       //奖池
	jackpotNoticeHandle timer.TimerHandle            //奖池金额通知
	jackpotNoticeTime   time.Time                    //上一次通知奖池的时间
	jackpotTime         time.Time                    //上一次奖池变化的时间
	lastJackpotValue    int64                        //上一次奖池变化时的值
	turnID              int32                        // 操作标识

}

func NewCaoThapSceneData(s *base.Scene) *CaoThapSceneData {
	return &CaoThapSceneData{
		Scene:   s,
		players: make(map[int32]*CaoThapPlayerData),
		//billedData: &proto_caothap.GameBilledData{},
	}
}

func (this *CaoThapSceneData) SaveData(force bool) {
}

func (this *CaoThapSceneData) OnPlayerLeave(p *base.Player, reason int) {
	if p, exist := this.players[p.SnId]; exist {
		delete(this.players, p.SnId)
	}
}

func (this *CaoThapSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

func (this *CaoThapSceneData) init() bool {
	if this.DbGameFree == nil {
		return false
	}

	otherIntParams := this.DbGameFree.GetOtherIntParams()
	params := this.DbGameFree.GetJackpot()
	this.jackpot = &base.XSlotJackpotPool{}
	this.jackpot.InitJackPot(len(otherIntParams))
	for i := 0; i < len(otherIntParams); i++ {
		baseBet := otherIntParams[i]
		this.jackpot.JackpotFund[i] = int64(params[caothap.CAOTHAP_JACKPOT_InitJackpot] * baseBet)
	}

	str := base.XSlotsPoolMgr.GetPool(this.GetGameFreeId(), this.Platform)
	if str != "" {
		jackpot := &base.XSlotJackpotPool{}
		err := json.Unmarshal([]byte(str), jackpot)
		if err == nil {
			this.jackpot = jackpot
		}
	}

	if this.jackpot != nil {
		base.XSlotsPoolMgr.SetPool(this.GetGameFreeId(), this.Platform, this.jackpot)
	}
	//this.lastJackpotValue = this.jackpot.JackpotFund
	this.jackpotTime = time.Now()
	this.turnID = caothap.TURNID_FIRSTINIT
	this.AfterTimer()
	return true
}

type CaoThapSpinResult struct {
	CardTypeValue int64 // 牌型分
	JackpotValue  int64 // 爆奖分
	PrizeValue    int64 // 赢分
	IsJackpot     bool  // 是否爆奖
	WinRate       int32 // 中奖倍率
	TotalTaxScore int64 // 总税收（牌型分+爆奖分）
}

func (this *CaoThapSceneData) broadcastJackpot() {
	pack := &proto_gamehall.BroadcastGameJackpot{
		GameFreeId: proto.Int32(this.DbGameFree.Id),
	}
	pack.JackpotFund = append(pack.JackpotFund, this.jackpot.JackpotFund...)
	proto.SetDefaults(pack)
	this.Broadcast(int(proto_gamehall.HundredScenePacketID_PACKET_BD_GAMEJACKPOT), pack, -1)
	//if this.lastJackpotValue != this.jackpot.JackpotFund {
	//	this.lastJackpotValue = this.jackpot.JackpotFund
	//
	//	pack := &proto_gamehall.BroadcastGameJackpot{
	//		JackpotFund: proto.Int64(this.jackpot.JackpotFund),
	//		GameFreeId:  proto.Int32(this.DbGameFree.Id),
	//	}
	//	proto.SetDefaults(pack)
	//	//this.Broadcast(int(proto_caothap.CaoThapPacketID_PACKET_SC_CAOTHAP_GAMEJACKPOT), pack, -1)
	//	this.Broadcast(int(proto_gamehall.HundredScenePacketID_PACKET_BD_GAMEJACKPOT), pack, -1)
	//	//
	//	//msg := &proto_server.GWGameJackpot{
	//	//	SceneId:     proto.Int32(int32(this.SceneId)),
	//	//	JackpotFund: proto.Int64(this.jackpot.JackpotFund),
	//	//}
	//	//proto.SetDefaults(msg)
	//	//logger.Logger.Infof("GWGameJackpot gameFreeID %v %v", this.DbGameFree.GetId(), msg)
	//	//this.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_GAMEJACKPOT), msg)
	//	//this.BroadcastMessageToPlatform(int(proto_gamehall.HundredScenePacketID_PACKET_BD_GAMEJACKPOT), pack)
	//}
}

func (this *CaoThapSceneData) BroadcastJackpot() {
	// 重置奖池变化的时间
	this.jackpotTime = time.Now()
	this.broadcastJackpot()
	////距上次通知时间较长时 直接发送新通知
	//if time.Now().After(this.jackpotNoticeTime.Add(jackpotNoticeInterval + time.Millisecond*100)) {
	//	this.jackpotNoticeTime = time.Now()
	//	this.broadcastJackpot()
	//	return
	//}
	////避免通知太频繁
	//if this.jackpotNoticeHandle != timer.TimerHandle(0) {
	//	return
	//}
	//this.jackpotNoticeHandle, _ = timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
	//	this.jackpotNoticeHandle = timer.TimerHandle(0)
	//	this.jackpotNoticeTime = time.Now()
	//	this.broadcastJackpot()
	//	return true
	//}), nil, jackpotNoticeInterval, 1)
}

func (this *CaoThapSceneData) AfterTimer() {
	var timeInterval = time.Second * 30 // 时间间隔
	timer.AfterTimer(func(h timer.TimerHandle, ud interface{}) bool {
		if this.jackpotTime.Add(timeInterval).Before(time.Now()) {
			this.PushVirtualDataToWorld()
		}
		this.AfterTimer()
		return true
	}, nil, timeInterval)
}

func (this *CaoThapSceneData) PushVirtualDataToWorld() {
	var isVirtualData bool
	otherIntParams := this.DbGameFree.GetOtherIntParams()
	var jackpotParam = this.DbGameFree.GetJackpot() // 奖池参数
	var jackpotInit = int64(jackpotParam[caothap.CAOTHAP_JACKPOT_InitJackpot] * otherIntParams[0])
	if jackpotInit > 0 && rand.Int31n(100000) < int32(this.jackpot.JackpotFund[0]*10/jackpotInit) { // 保留一位小数位
		isVirtualData = true
	}
	if isVirtualData {
		// 推送最新开奖记录到world
		msg := &proto_server.GWGameNewBigWinHistory{
			SceneId: proto.Int32(int32(this.SceneId)),
			BigWinHistory: &proto_server.BigWinHistoryInfo{
				CreatedTime:   proto.Int64(time.Now().Unix()),
				BaseBet:       proto.Int64(int64(this.DbGameFree.GetBaseScore())),
				TotalBet:      proto.Int64(int64(this.DbGameFree.GetBaseScore())),
				PriceValue:    proto.Int64(this.jackpot.JackpotFund[0]),
				IsVirtualData: proto.Bool(isVirtualData),
			},
		}
		this.jackpot.JackpotFund[0] = jackpotInit // 仅初始化奖池金额
		proto.SetDefaults(msg)
		logger.Logger.Infof("GWGameNewBigWinHistory gameFreeID %v %v", this.DbGameFree.GetId(), msg)
		this.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), msg)
	}
}
