package base

import (
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	rawproto "google.golang.org/protobuf/proto"
	"strconv"
	"time"
)

// 对应到客户端的一个玩家对象.

//私有标记，不同步给客户端
const (
	PlayerState_Online           int = 1 << iota //在线标记 1
	PlayerState_Ready                            //准备标记 2
	PlayerState_WaitOp                           //等待操作标记 3
	PlayerState_Auto                             //托管状态 4
	PlayerState_Check                            //已看牌状态 5
	PlayerState_Fold                             //弃牌状态 6
	PlayerState_WaitNext                         //等待下一局游戏 7
	PlayerState_GameBreak                        //不能继续游戏 8
	PlayerState_Leave                            //暂离状态 9
	PlayerState_EnterSceneFailed                 //进场失败 10
	PlayerState_Max
)

//玩家事件
const (
	PlayerEventEnter    int = iota //进入事件
	PlayerEventLeave               //离开事件
	PlayerEventDropLine            //掉线
	PlayerEventRehold              //重连
	PlayerReturn                   //玩家重新进入房间
	PlayerEventRecharge            //冲值事件
)

//小游戏内加减币重试次数
const PlayerAddCoinDefRetryCnt int = 3

type Player struct {
	model.PlayerData                 //po 持久化对象
	ExtraData        interface{}     //扩展接口
	gateSess         *netlib.Session //所在GateServer的session
	worldSess        *netlib.Session //所在WorldServer的session
	scene            *Scene          //当前所在个Scene
	sid              int64           //对应客户端的sessionId
	gateSid          int64           //对应网关的sessionId
	city             string          //城市
	flag             int             //状态标记
	Pos              int             //当前位置
	dirty            bool            //脏标记
	billed           bool            //是否已经结算过了
	Coin             int64           //金币
	serviceFee       int64           //服务费|税收
	TotalBet         int64           //总下注额（从进房间开始,包含多局游戏的下注）
	GameTimes        int32           //游戏次数
	winTimes         int             //胜利次数
	lostTimes        int             //失败次数
	takeCoin         int64           //携带金币
	ExpectLeaveCoin  int64           //期望离场时的金币[机器人用]
	ExpectGameTime   int32           //期望进行的局数[机器人用]
	curIsWin         int64           //当局输赢   负数：输  0：平局  正数：赢
	currentCoin      int64           //本局结束后剩余
	CurrentBet       int64           //本局下注额
	CurrentTax       int64           //本局税收
	//StartCoin          int64           //本局开始金币
	IsQM               bool      //是否是全民推广用户
	LastOPTimer        time.Time //玩家最后一次操作时间
	ValidCacheBetTotal int64     //有效下注缓存
	isFightRobot       bool      //测试机器人，这种机器人可以用作记录水池数据，方便模拟用户输赢
	dropTime           time.Time //掉线时间
	WhiteLevel         int32
	BlackLevel         int32
	SingleAdjust       *model.PlayerSingleAdjust
}

func NewPlayer(sid int64, data []byte, ws, gs *netlib.Session) *Player {
	p := &Player{
		sid:       sid,
		worldSess: ws,
		gateSess:  gs,
		flag:      PlayerState_Online,
		Pos:       -1,
	}
	if p.init(data) {
		return p
	}

	return nil
}

func (this *Player) init(data []byte) bool {
	if !this.UnmarshalData(data) {
		return false
	}
	this.city = this.City
	if this.GDatas == nil {
		this.GDatas = make(map[string]*model.PlayerGameInfo)
	}
	if this.WBLevel > 0 {
		this.WhiteLevel = this.WBLevel
	} else if this.WBLevel < 0 {
		this.BlackLevel = -this.WBLevel
	}
	return true
}
func (this *Player) MarkFlag(flag int) {
	this.flag |= flag
}
func (this *Player) UnmarkFlag(flag int) {
	this.flag &= ^flag
}
func (this *Player) IsMarkFlag(flag int) bool {
	if (this.flag & flag) != 0 {
		return true
	}
	return false
}
func (this *Player) IsOnLine() bool {
	return this.IsMarkFlag(PlayerState_Online)
}

func (this *Player) IsReady() bool {
	return this.IsMarkFlag(PlayerState_Ready)
}

func (this *Player) IsAuto() bool {
	return this.IsMarkFlag(PlayerState_Auto)
}

func (this *Player) IsGameing() bool {
	return !this.IsMarkFlag(PlayerState_WaitNext) && !this.IsMarkFlag(PlayerState_GameBreak)
}

func (this *Player) SyncFlag(onlysync2me ...bool) {
	//donothing
}

func (this *Player) SyncFlagToWorld() {
	//donothing
}

func (this *Player) SendToClient(packetid int, rawpack interface{}) bool {
	if this.gateSess == nil {
		logger.Logger.Warnf("(this *Player) SendToClient [snid:%v packetid:%v] gatesess == nil ", this.SnId, packetid)
		return false
	}
	if rawpack == nil {
		logger.Logger.Tracef("(this *Player) SendToClient [snid:%v packetid:%v] rawpack == nil ", this.SnId, packetid)
		return false
	}
	if this.scene == nil {
		logger.Logger.Tracef("(this *Player) SendToClient [snid:%v packetid:%v] scene == nil ", this.SnId, packetid)
		return false
	}
	if !this.IsOnLine() {
		logger.Logger.Warnf("(this *Player) SendToClient [snid:%v packetid:%v] Player if offline.", this.SnId, packetid)
		return false
	}
	if this.IsMarkFlag(PlayerState_Leave) {
		logger.Logger.Warnf("(this *Player) SendToClient [snid:%v packetid:%v] Player if leave.", this.SnId, packetid)
		return false
	}

	//包装一下
	//wrapPack := &mngame.SCMNGameDispatcher{
	//	Id: this.GetScene().GetGameFreeId(),
	//}
	//if data, err := netlib.MarshalPacket(packetid, rawpack); err == nil {
	//	wrapPack.Data = data
	//	return common.SendToGate(this.sid, int(mngame.MNGamePacketID_PACKET_SC_MNGAME_DISPATCHER), wrapPack, this.gateSess)
	//}

	return false
}

func (this *Player) Broadcast(packetid int, rawpack interface{}, excludeSid int64) bool {
	if this.scene != nil {
		this.scene.Broadcast(packetid, rawpack.(rawproto.Message), excludeSid)
		return true
	}
	return false
}

func (this *Player) SendToWorld(packetid int, rawpack interface{}) bool {
	if this.worldSess == nil {
		logger.Logger.Tracef("(this *Player) SendToWorld [%v] worldsess == nil ", this.Name)
		return false
	}
	if rawpack == nil {
		logger.Logger.Trace("(this *Player) SendToWorld rawpack == nil ")
		return false
	}

	this.scene.SendToWorld(packetid, rawpack)
	return true
}

func (this *Player) OnEnter(s *Scene) {
	this.scene = s
}

func (this *Player) OnRehold(newSid int64, newSess *netlib.Session) {
	this.sid = newSid
	this.gateSess = newSess
}

func (this *Player) OnDropLine() {
	this.gateSess = nil
	this.sid = 0
}

func (this *Player) OnLeave(reason int) {
}

func (this *Player) MarshalData(gameid int) (d []byte, e error) {
	d, e = netlib.Gob.Marshal(&this.PlayerData)
	logger.Logger.Trace("(this *Player) MarshalData(gameid int)")
	return
}

func (this *Player) UnmarshalData(data []byte) bool {
	err := netlib.Gob.Unmarshal(data, &this.PlayerData)
	if err == nil {
		this.dirty = true
		return true
	} else {
		logger.Logger.Warn("Player.SyncData err:", err)
	}
	return false
}
func (this *Player) UnMarshalSingleAdjustData(data []byte) bool {
	if data == nil {
		return true
	}
	err := netlib.Gob.Unmarshal(data, &this.SingleAdjust)
	if err == nil {
		return true
	} else {
		logger.Logger.Warn("Player.SingleAdjust err:", err)
	}
	return false
}
func (this *Player) IsSingleAdjustPlayer() (*model.PlayerSingleAdjust, bool) {
	if this.SingleAdjust == nil {
		return nil, false
	}
	return this.SingleAdjust, true
}
func (this *Player) AddAdjustCount(gameFreeId int32) {
	sa := this.SingleAdjust
	if sa != nil && sa.GameFreeId == gameFreeId {
		this.SingleAdjust.CurTime++
		//通知world
		pack := &server.GWAddSingleAdjust{
			Platform:   this.Platform,
			SnId:       this.SnId,
			GameFreeId: gameFreeId,
		}
		this.SendToWorld(int(server.SSPacketID_PACKET_GW_ADDSINGLEADJUST), pack)
	}
}
func (this *Player) UpsertSingleAdjust(msg *model.PlayerSingleAdjust) {
	this.SingleAdjust = msg
}
func (this *Player) DeleteSingleAdjust(platform string, gamefreeid int32) {
	c := this.SingleAdjust
	if c != nil && c.Platform == platform && c.GameFreeId == gamefreeid {
		this.SingleAdjust = nil
	}
}
func (this *Player) OnSecTimer() {
}

func (this *Player) OnMiniTimer() {
}

func (this *Player) OnHourTimer() {
}

func (this *Player) OnDayTimer() {
	//在线跨天 数据给昨天，今天置为空
	this.YesterdayGameData = this.TodayGameData
	this.TodayGameData = model.NewPlayerGameCtrlData()
}

func (this *Player) OnMonthTimer() {
}

func (this *Player) OnWeekTimer() {
}

func (this *Player) GetName() string {
	return this.Name
}

func (this *Player) MarkDirty() {
	this.dirty = true
}

func (this *Player) AddCoin(num int64, gainWay int32, notifyC, broadcast bool, oper, remark string, retryCnt int, cb AddCoinCb) {
	if num == 0 {
		logger.Logger.Tracef("金币变动:%v 直接返回", num)
		content := &AddCoinContext{
			Player:    this,
			Coin:      num,
			GainWay:   gainWay,
			Oper:      oper,
			Remark:    remark,
			NotifyCli: notifyC,
			Broadcast: broadcast,
			WriteLog:  true,
			RetryCnt:  retryCnt,
		}
		cb(content, this.Coin, true)
		return
	}
	StartAsyncAddCoinTransact(&AddCoinContext{
		Player:    this,
		Coin:      num,
		GainWay:   gainWay,
		Oper:      oper,
		Remark:    remark,
		NotifyCli: notifyC,
		Broadcast: broadcast,
		WriteLog:  true,
		RetryCnt:  retryCnt,
	}, cb)
}

//保存金币变动日志
//数据用途: 个人房间内牌局账变记录，后台部分报表使用，确保数据计算无误，否则可能影响月底对账
//takeCoin: 牌局结算前玩家身上的金币
//changecoin:  本局玩家输赢的钱，注意是税后
//coin: 结算后玩家当前身上的金币余额
//totalbet: 总下注额
//taxcoin: 本局该玩家产生的税收，这里要包含俱乐部的税
//wincoin: 本局赢取的金币，含税 wincoin==changecoin+taxcoin
//jackpotWinCoin: 从奖池中赢取的金币(拉霸类游戏)
//smallGameWinCoin: 小游戏赢取的金币(拉霸类游戏)
func (this *Player) SaveSceneCoinLog(takeCoin, changecoin, coin, totalbet, taxcoin, wincoin int64, jackpotWinCoin int64, smallGameWinCoin int64) {
	if this.scene != nil {
		if !this.IsRob && !this.scene.Testing && !this.scene.IsMatchScene() { //机器人log排除掉
			var eventType int64 //输赢事件值 默认值为0
			if coin-takeCoin > 0 {
				eventType = 1
			} else if coin-takeCoin < 0 {
				eventType = -1
			}
			log := model.NewSceneCoinLogEx(this.SnId, changecoin, takeCoin, coin, eventType,
				int64(this.scene.DbGameFree.GetBaseScore()), totalbet, int32(this.scene.GameId), this.PlayerData.Ip,
				this.scene.paramsEx[0], this.Pos, this.Platform, this.Channel, this.BeUnderAgentCode, int32(this.scene.SceneId),
				this.scene.DbGameFree.GetGameMode(), this.scene.GetGameFreeId(), taxcoin, wincoin,
				jackpotWinCoin, smallGameWinCoin, this.PackageID)
			if log != nil {
				LogChannelSington.WriteLog(log)
			}
		}
	}
}

func (this *Player) SyncCoin() {

}

func (this *Player) ReportGameEvent(tax, taxex, changeCoin, validbet, validFlow, in, out int64) {
	// 记录玩家 首次参与该场次的游戏时间 游戏次数
	gameFreeId := strconv.Itoa(int(this.scene.DbGameFree.GetId()))
	data, ok := this.GDatas[gameFreeId]
	if !ok {
		data = &model.PlayerGameInfo{FirstTime: time.Now(), Statics: model.PlayerGameStatics{GameTimes: 1}}
		this.GDatas[gameFreeId] = data
	} else {
		if data.Statics.GameTimes <= 0 {
			data.FirstTime = time.Now()
		}
		data.Statics.GameTimes++
	}

	// 记录玩家 首次参与该游戏时间 游戏次数(不区分场次)
	dataGame, ok := this.GDatas[this.scene.KeyGameId]
	if !ok {
		dataGame = &model.PlayerGameInfo{FirstTime: time.Now(), Statics: model.PlayerGameStatics{GameTimes: 1}}
		this.GDatas[this.scene.KeyGameId] = dataGame
	} else {
		if dataGame.Statics.GameTimes <= 0 {
			dataGame.FirstTime = time.Now()
		}
		dataGame.Statics.GameTimes++
	}

	gamingTime := int32(time.Now().Sub(this.scene.GameNowTime).Seconds())
	LogChannelSington.WriteMQData(model.GenerateGameEvent(model.CreatePlayerGameRecEvent(this.SnId, tax, taxex, changeCoin, validbet, validFlow, in, out,
		int32(this.scene.GameId), this.scene.DbGameFree.GetId(), int32(this.scene.GameMode),
		this.scene.GetRecordId(), this.Channel, this.BeUnderAgentCode, this.Platform, this.City, this.DeviceOS,
		this.CreateTime, gamingTime, data.FirstTime.Local(), dataGame.FirstTime.Local(), data.Statics.GameTimes, dataGame.Statics.GameTimes, this.LastLoginTime,
		this.TelephonePromoter, this.DeviceId)))
}

//破产事件
func (this *Player) ReportBankRuptcy(gameId, gameMode, gameFreeId int32) {
	//if !this.IsRob {
	//	d, e := model.MarshalBankruptcyEvent(2, this.SnId, this.TelephonePromoter, this.Channel, this.BeUnderAgentCode, this.Platform, this.City, this.CreateTime, gameId, gameMode, gameFreeId)
	//	if e == nil {
	//		rmd := model.NewInfluxDBData("hj.player_bankruptcy", d)
	//		if rmd != nil {
	//			InfluxDBDataChannelSington.Write(rmd)
	//		}
	//	}
	//}
}

// 汇总玩家该次游戏总产生的税收
// 数据用途: 平台和推广间分账用，确保数据计算无误，
// 注意：该税收不包含俱乐部的抽水
// tax：游戏税收
func (this *Player) AddServiceFee(tax int64) {
	if this.scene == nil || this.scene.Testing { //测试场不统计
		return
	}
	if tax > 0 && !this.IsRob {
		this.serviceFee += tax
	}
}

func (this *Player) GetStaticsData(gameDiff string) (winCoin int64, lostCoin int64) {
	if this.PlayerData.GDatas != nil {
		if data, ok := this.PlayerData.GDatas[gameDiff]; ok {
			winCoin, lostCoin = data.Statics.TotalOut, data.Statics.TotalIn
			return
		}
	}
	return
}
func (this *Player) SaveReportForm(showId, roomType int, keyGameId string, profitCoin, flow int64, validBet int64) {
	//个人报表统计
	if this.TotalGameData == nil {
		this.TotalGameData = make(map[int][]*model.PlayerGameTotal)
	}
	if this.TotalGameData[showId] == nil {
		this.TotalGameData[showId] = []*model.PlayerGameTotal{new(model.PlayerGameTotal)}
	}
	td := this.TotalGameData[showId][len(this.TotalGameData[showId])-1]
	td.ProfitCoin += profitCoin
	td.BetCoin += validBet
	td.FlowCoin += flow
	///////////////最多盈利
	if pgs, exist := this.GDatas[keyGameId]; exist {
		if pgs.Statics.MaxSysOut < profitCoin {
			pgs.Statics.MaxSysOut = profitCoin
		}
	} else {
		this.GDatas[keyGameId] = &model.PlayerGameInfo{Statics: model.PlayerGameStatics{MaxSysOut: profitCoin}}
	}
}

// 个人投入产出汇总，以游戏id为key存储
// 数据用途：计算玩家赔率用，数据确保计算无误，否则可能影响玩家手牌的调控
// key: 游戏ID对应的字符串
// gain：输赢额，注意如果是[正值]这里一定要用税前数据，否则玩家会有数值调控优势
// 如果需要汇总gameid today的数据，可以使用game scene的GetTotalTodayDaliyGameData
func (this *Player) Statics(keyGameId string, keyGameFreeId string, gain int64, isAddNum bool) {
	if this.scene == nil || this.scene.Testing { //测试场|自建房和机器人不统计
		return
	}

	if this.IsRob && !this.scene.IsRobFightGame() {
		return
	}

	if this.TodayGameData.CtrlData == nil {
		this.TodayGameData.CtrlData = make(map[string]*model.PlayerGameStatics)
	}

	var totalIn int64
	var totalOut int64
	if gain > 0 {
		totalOut = gain
	} else {
		totalIn = -gain
	}

	statics := make([]*model.PlayerGameStatics, 0, 4)
	//当天数据统计
	//按场次分
	if data, ok := this.TodayGameData.CtrlData[keyGameFreeId]; ok {
		statics = append(statics, data)
	} else {
		gs := &model.PlayerGameStatics{}
		this.TodayGameData.CtrlData[keyGameFreeId] = gs
		statics = append(statics, gs)
	}
	//按游戏分
	if data, ok := this.TodayGameData.CtrlData[keyGameId]; ok {
		statics = append(statics, data)
	} else {
		data = &model.PlayerGameStatics{}
		this.TodayGameData.CtrlData[keyGameId] = data
		statics = append(statics, data)
	}

	//按游戏场次进行的统计
	if data, ok := this.GDatas[keyGameFreeId]; ok {
		statics = append(statics, &data.Statics)
	} else {
		data = &model.PlayerGameInfo{FirstTime: time.Now(), Statics: model.PlayerGameStatics{}}
		this.GDatas[keyGameFreeId] = data
		statics = append(statics, &data.Statics)
	}
	if data, ok := this.GDatas[keyGameId]; ok {
		statics = append(statics, &data.Statics)
	} else {
		data = &model.PlayerGameInfo{FirstTime: time.Now(), Statics: model.PlayerGameStatics{}}
		this.GDatas[keyGameId] = data
		statics = append(statics, &data.Statics)
	}

	if !this.scene.IsPrivateScene() {
		//增加黑白名单、GM过滤，因为黑白名单过后，会导致玩家体验急剧变化
		needStatic := this.WhiteLevel == 0 && this.WhiteFlag == 0 && this.BlackLevel == 0 && this.GMLevel == 0
		//增加黑白名单过滤，因为黑白名单后，会导致数据出现补偿
		if needStatic {
			for _, data := range statics {
				if data != nil {
					data.TotalIn += totalIn
					data.TotalOut += totalOut
					if isAddNum {
						data.GameTimes++
						if gain > 0 {
							data.WinGameTimes++
						} else if gain < 0 {
							data.LoseGameTimes++
						} else {
							data.DrawGameTimes++
						}
					}
				}
			}
		}
	}
}

func (this *Player) CheckType(gamefreeId, gameId int32) *server.DB_PlayerType {
	types := srvdata.PlayerTypeMgrSington.GetPlayerType(gamefreeId)
	cnt := len(types)
	if cnt > 0 {
		var pgs *model.PlayerGameStatics
		if this.GDatas != nil {
			if d, exist := this.GDatas[strconv.Itoa(int(gameId))]; exist {
				pgs = &d.Statics
			}
		}

		//赔率 产出/投入 万分比
		odds := int64(float64(float64(pgs.TotalOut+1)/float64(pgs.TotalIn+1)) * 10000)
		if odds > 10000000 {
			odds = 10000000
		}
		for i := 0; i < cnt; i++ {
			t := types[i]
			if t != nil {
				if this.CoinPayTotal >= int64(t.GetPayLowerLimit()) && this.CoinPayTotal <= int64(t.GetPayUpperLimit()) &&
					pgs.GameTimes >= int64(t.GetGameTimeLowerLimit()) && pgs.GameTimes <= int64(t.GetGameTimeUpperLimit()) &&
					pgs.TotalIn >= int64(t.GetTotalInLowerLimit()) && pgs.TotalIn <= int64(t.GetTotalInUpperLimit()) &&
					odds >= int64(t.GetOddsLowerLimit()) && odds <= int64(t.GetOddsUpperLimit()) {
					return t
				}
			}
		}
	}
	return nil
}

//计算玩家赔率 产出/投入
func (this *Player) LoseRate(gamefreeId, gameId int32) (rate float64) {
	if this.GDatas != nil {
		if d, exist := this.GDatas[strconv.Itoa(int(gameId))]; exist {
			return float64(float64(d.Statics.TotalOut+1) / float64(d.Statics.TotalIn+1))
		}
	}

	return 1.0
}

//计算玩家赔率 产出/投入
func (this *Player) LoseRateKeyGameid(gameKeyId string) (rate float64) {
	if this.GDatas != nil {
		if d, exist := this.GDatas[gameKeyId]; exist {
			return float64(float64(d.Statics.TotalOut+1) / float64(d.Statics.TotalIn+1))
		}
	}

	return 1.0
}

func (this *Player) BirdPlayerCheck(keyGameId string) bool {
	if model.GameParamData.BirdPlayerFlag == false {
		return false
	}
	if this.IsRob {
		return false
	}
	if this.GDatas == nil {
		return true
	}

	playerDate, ok := this.GDatas[keyGameId]
	if !ok || playerDate == nil {
		return true
	}

	if playerDate.Statics.GameTimes < 10 && playerDate.Statics.TotalOut < 100000 &&
		(playerDate.Statics.TotalOut+100000)/(playerDate.Statics.TotalIn+100000) <= 2 {
		return true
	}
	return false
}

func (this *Player) GetCoin() int64 {
	return this.Coin
}
func (this *Player) SetCoin(coin int64) {
	this.Coin = coin
}

//func (this *Player) GetStartCoin() int64 {
//	return this.StartCoin
//}
//func (this *Player) SetStartCoin(startCoin int64) {
//	this.StartCoin = startCoin
//}
func (this *Player) GetExtraData() interface{} {
	return this.ExtraData
}
func (this *Player) SetExtraData(data interface{}) {
	this.ExtraData = data
}
func (this *Player) GetLastOPTimer() time.Time {
	return this.LastOPTimer
}
func (this *Player) SetLastOPTimer(lastOPTimer time.Time) {
	this.LastOPTimer = lastOPTimer
}
func (this *Player) GetPos() int {
	return this.Pos
}
func (this *Player) SetPos(pos int) {
	this.Pos = pos
}
func (this *Player) GetGameTimes() int32 {
	return this.GameTimes
}
func (this *Player) SetGameTimes(gameTimes int32) {
	this.GameTimes = gameTimes
}
func (this *Player) GetWinTimes() int {
	return this.winTimes
}
func (this *Player) SetWinTimes(winTimes int) {
	this.winTimes = winTimes
}
func (this *Player) GetLostTimes() int {
	return this.lostTimes
}
func (this *Player) SetLostTimes(lostTimes int) {
	this.lostTimes = lostTimes
}
func (this *Player) GetTotalBet() int64 {
	return this.TotalBet
}
func (this *Player) SetTotalBet(totalBet int64) {
	this.TotalBet = totalBet
}
func (this *Player) GetCurrentBet() int64 {
	return this.CurrentBet
}
func (this *Player) SetCurrentBet(currentBet int64) {
	this.CurrentBet = currentBet
}
func (this *Player) GetCurrentTax() int64 {
	return this.CurrentTax
}
func (this *Player) SetCurrentTax(currentTax int64) {
	this.CurrentTax = currentTax
}
func (this *Player) GetSid() int64 {
	return this.sid
}
func (this *Player) SetSid(sid int64) {
	this.sid = sid
}
func (this *Player) GetScene() *Scene {
	return this.scene
}
func (this *Player) GetGateSess() *netlib.Session {
	return this.gateSess
}
func (this *Player) SetGateSess(gateSess *netlib.Session) {
	this.gateSess = gateSess
}
func (this *Player) GetWorldSess() *netlib.Session {
	return this.worldSess
}
func (this *Player) SetWorldSess(worldSess *netlib.Session) {
	this.worldSess = worldSess
}
func (this *Player) GetCurrentCoin() int64 {
	return this.currentCoin
}
func (this *Player) SetCurrentCoin(currentCoin int64) {
	this.currentCoin = currentCoin
}
func (this *Player) GetTakeCoin() int64 {
	return this.takeCoin
}
func (this *Player) SetTakeCoin(takeCoin int64) {
	this.takeCoin = takeCoin
}
func (this *Player) GetFlag() int {
	return this.flag
}
func (this *Player) SetFlag(flag int) {
	this.flag = flag
}
func (this *Player) GetCity() string {
	return this.city
}
func (this *Player) SetCity(city string) {
	this.city = city
}

//////////////////////////////////////////////////
