package rollpoint

import (
	rule "games.yol.com/win88/gamerule/rollpoint"
	"games.yol.com/win88/gamesrv/base"
	"math/rand"
	"sort"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/rollpoint"
	"github.com/idealeak/goserver/core/logger"
)

type RollPointSceneData struct {
	*base.Scene                                        //房间信息
	blackBox            *rule.BlackBox                 //骰子
	players             map[int32]*RollPointPlayerData //玩家信息
	seats               []*RollPointPlayerData         //座位信息
	WinLog              []int64                        //开奖记录
	SyncTime            time.Time                      //同步押注金币的时间
	TotalBet            map[int32]map[int32]int64      //总下注数据
	SyncData            map[int32]map[int32]int64      //同步的数据
	ZoneTotal           map[int32]int64                //区域总下注额
	ZoneRobTotal        map[int32]int64                //机器人区域总下注额
	syncCnt             int                            //同步数量
	RollPoint           []int32                        //可选筹码
	NoBanker            int32
	RollPointTargetList []int32
	LastWinerBetPos     []int32
	LastWinerSnid       int32
}
type RollPointSort struct {
	systemCoinOut int64
	point         [rule.RollNum]int32
}

func NewRollPointSceneData(s *base.Scene) *RollPointSceneData {
	return &RollPointSceneData{
		Scene:        s,
		players:      make(map[int32]*RollPointPlayerData),
		SyncData:     make(map[int32]map[int32]int64),
		TotalBet:     make(map[int32]map[int32]int64),
		ZoneTotal:    make(map[int32]int64),
		ZoneRobTotal: make(map[int32]int64),
	}
}

func (this *RollPointSceneData) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
}

func (this *RollPointSceneData) init() bool {
	data := this.GetDBGameFree()
	if data != nil {
		this.RollPoint = data.GetOtherIntParams()
	} else {
		logger.Logger.Error("Init roll coin scene failed,scene type:", this.SceneType)
	}
	//初始化区域和面额数据
	for i := int32(0); i < rule.BetArea_Max; i++ {
		this.TotalBet[i] = make(map[int32]int64)
		this.ZoneTotal[i] = 0
		this.ZoneRobTotal[i] = 0
		for _, coin := range this.RollPoint {
			this.TotalBet[i][coin] = 0
		}
	}
	this.blackBox = rule.CreateBlackBox()

	return true
}

func (this *RollPointSceneData) Clean() {
	for key, _ := range this.TotalBet {
		this.ZoneTotal[key] = 0
		this.ZoneRobTotal[key] = 0
		for coin, _ := range this.TotalBet[key] {
			this.TotalBet[key][coin] = 0
		}
	}
	for _, p := range this.players {
		p.Clean()
	}
	this.LastWinerBetPos = nil
	this.LastWinerSnid = 0
	//重置水池调控标记
	this.SetCpControlled(false)
}

func (this *RollPointSceneData) delPlayer(p *base.Player) {
	if p, exist := this.players[p.SnId]; exist {
		delete(this.players, p.SnId)
	}
}
func (this *RollPointSceneData) CountPlayer() int {
	num := len(this.players)
	return num
}
func (this *RollPointSceneData) OnPlayerLeave(p *base.Player, reason int) {
	this.delPlayer(p)
	this.BroadcastPlayerLeave(p)
	p.UnmarkFlag(base.PlayerState_GameBreak)
	p.SyncFlag(true)
}

func (this *RollPointSceneData) BroadcastPlayerLeave(p *base.Player) {
	scLeavePack := &rollpoint.SCRollPointPlayerNum{
		PlayerNum: proto.Int(this.CountPlayer()),
	}
	proto.SetDefaults(scLeavePack)
	this.Broadcast(int(rollpoint.RPPACKETID_ROLLPOINT_SC_PLAYERNUM), scLeavePack, p.GetSid())
}

func (this *RollPointSceneData) SceneDestroy(force bool) {
	this.Scene.Destroy(force)
}

func (this *RollPointSceneData) GetRobAllCoin() int64 {
	robAllCoin := int64(0)
	for _, value := range this.ZoneRobTotal {
		robAllCoin += value
	}
	return robAllCoin
}
func (this *RollPointSceneData) GetAllCoin() int64 {
	allCoin := int64(0)
	for _, value := range this.ZoneTotal {
		allCoin += value
	}
	return allCoin
}
func (this *RollPointSceneData) LogWinFlag() {
	point := this.blackBox.Point
	flag := point[0] | point[1]<<8 | point[2]<<16
	this.WinLog = append(this.WinLog, int64(flag))
	if len(this.WinLog) > 40 {
		index := len(this.WinLog) - 40
		this.WinLog = this.WinLog[index:]
	}
}
func (this *RollPointSceneData) CacheCoinLog(flag, coin int32, isRob bool) {
	if pool, ok := this.SyncData[flag]; ok {
		pool[coin] = pool[coin] + 1
	} else {
		pool := make(map[int32]int64)
		pool[coin] = 1
		this.SyncData[flag] = pool
	}
	this.ZoneTotal[flag] = this.ZoneTotal[flag] + int64(coin)
	if isRob {
		this.ZoneRobTotal[flag] = this.ZoneRobTotal[flag] + int64(coin)
	}
	this.syncCnt++
}

func (this *RollPointSceneData) SyncCoinLog() {
	if this.syncCnt > 0 {
		pack := &rollpoint.SCRollPointCoinLog{
			Pos: this.LastWinerBetPos,
		}
		for i := int32(0); i < rule.BetArea_Max; i++ {
			if pool, ok := this.SyncData[i]; ok {
				coinLog := &rollpoint.RollPointCoinLog{
					Index: proto.Int32(i),
				}
				for _, coin := range this.RollPoint {
					coinLog.Coins = append(coinLog.Coins, coin, int32(pool[coin]))
				}
				pack.Coins = append(pack.Coins, coinLog)
			}
		}
		proto.SetDefaults(pack)
		this.Broadcast(int(rollpoint.RPPACKETID_ROLLPOINT_SC_COINLOG), pack, 0) //广播给其他用户
		logger.Logger.Trace("SCRollPointCoinLog:", pack)
		for key, pool := range this.SyncData {
			for coin, cnt := range pool {
				//本次缓存应用到总押注中
				if pool2, ok := this.TotalBet[key]; ok {
					pool2[coin] = pool2[coin] + cnt
				} else {
					pool2 := make(map[int32]int64)
					pool2[coin] = cnt
					this.TotalBet[key] = pool2
				}
				pool[coin] = 0
			}
		}
		this.syncCnt = 0
	}
}

//计算出玩家输赢
func (this *RollPointSceneData) CalcuSystemCoinOut() int64 {
	systemOut := int64(0)
	systemIn := int64(0)
	for index, rate := range this.blackBox.Score {
		if this.ZoneTotal[int32(index)] > 0 {
			playerBetCoin := this.ZoneTotal[int32(index)] - this.ZoneRobTotal[int32(index)]
			if rate > 0 {
				systemOut += playerBetCoin * int64(rate)
			} else {
				systemIn += playerBetCoin
			}
		}
	}
	return systemOut - systemIn
}

//取玩家最少赢钱投点
func (this *RollPointSceneData) GetMinPlayerBetPoint() [rule.RollNum]int32 {
	sortArr := []RollPointSort{}
	indexArr := rand.Perm(len(rule.AllScore))
	for _, key := range indexArr {
		value := rule.AllScore[key]
		systemOut := int64(0)
		for index, rate := range value {
			totle := this.ZoneTotal[int32(index)] - this.ZoneRobTotal[int32(index)]
			systemOut += totle * int64(rate)
		}
		sortArr = append(sortArr, RollPointSort{
			systemCoinOut: systemOut,
			point:         rule.AllRoll[key],
		})
	}
	sort.Slice(sortArr, func(i, j int) bool {
		return sortArr[i].systemCoinOut < sortArr[j].systemCoinOut
	})
	if len(sortArr) == 0 {
		return this.blackBox.Point
	}
	return sortArr[rand.Intn(len(sortArr)/3)].point
}

//取玩家赢钱位置投点
func (this *RollPointSceneData) GetMiddlePlayerBetPoint() [rule.RollNum]int32 {
	sortArr := []RollPointSort{}
	indexArr := rand.Perm(len(rule.AllScore))
	for _, key := range indexArr {
		value := rule.AllScore[key]
		systemOut := int64(0)
		for index, rate := range value {
			totle := this.ZoneTotal[int32(index)] - this.ZoneRobTotal[int32(index)]
			systemOut += totle * int64(rate)
		}
		if systemOut > 0 {
			sortArr = append(sortArr, RollPointSort{
				systemCoinOut: systemOut,
				point:         rule.AllRoll[key],
			})
		}
	}
	if len(sortArr) == 0 {
		return this.blackBox.Point
	}
	return sortArr[rand.Intn(len(sortArr))].point
}

//取玩家最大赢钱位置投点
func (this *RollPointSceneData) GetMaxPlayerBetPoint() [rule.RollNum]int32 {
	sortArr := []RollPointSort{}
	indexArr := rand.Perm(len(rule.AllScore))
	for _, key := range indexArr {
		value := rule.AllScore[key]
		systemOut := int64(0)
		for index, rate := range value {
			totle := this.ZoneTotal[int32(index)] - this.ZoneRobTotal[int32(index)]
			systemOut += totle * int64(rate)
		}
		if systemOut > 0 {
			sortArr = append(sortArr, RollPointSort{
				systemCoinOut: systemOut,
				point:         rule.AllRoll[key],
			})
		}
	}
	sort.Slice(sortArr, func(i, j int) bool {
		return sortArr[i].systemCoinOut > sortArr[j].systemCoinOut
	})
	if len(sortArr) == 0 {
		return this.blackBox.Point
	}
	return sortArr[0].point
}
func (this *RollPointSceneData) ChangeCard() int {
	if this.GetRobAllCoin() == this.GetAllCoin() {
		return -1
	}
	sysOut := this.CalcuSystemCoinOut()
	if !base.CoinPoolMgr.IsMaxOutHaveEnough(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId(), -sysOut) {
		this.SetCpControlled(true)
		this.blackBox.Point = this.GetMinPlayerBetPoint()
		this.blackBox.Score = rule.CalcPoint(this.blackBox.Point)
	}
	status, changeRate := base.CoinPoolMgr.GetCoinPoolStatus(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())
	switch status {
	case base.CoinPoolStatus_Low: //库存值 < 库存下限
		if common.RandInt(10000) < changeRate {
			this.SetCpControlled(true)
			this.blackBox.Point = this.GetMinPlayerBetPoint()
			this.blackBox.Score = rule.CalcPoint(this.blackBox.Point)
		}
	case base.CoinPoolStatus_High, base.CoinPoolStatus_TooHigh:
		conBetTotle := int64(0)
		for index, coin := range this.ZoneTotal {
			if coin > 0 {
				conBetTotle += coin - this.ZoneRobTotal[int32(index)]
			}
		}
		if conBetTotle < sysOut {
			return -1
		}
		currCoin := base.CoinPoolMgr.LoadCoin(this.GetGameFreeId(), this.GetPlatform(), this.GetGroupId())
		setting := base.CoinPoolMgr.GetCoinPoolSetting(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())
		if !(this.RandInt(int(setting.GetUpperOffsetLimit())) < (int(currCoin)-int(setting.GetUpperLimit()))*2) {
			return -1
		}
		this.SetCpControlled(true)
		for i := 0; i < 20; i++ {
			this.blackBox.Point = this.GetMiddlePlayerBetPoint()
			this.blackBox.Score = rule.CalcPoint(this.blackBox.Point)
			sysOut := this.CalcuSystemCoinOut()
			if sysOut > 0 && base.CoinPoolMgr.IsMaxOutHaveEnough(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId(), -sysOut) {
				return -1
			}
		}
		this.blackBox.Point = this.GetMinPlayerBetPoint()
		this.blackBox.Score = rule.CalcPoint(this.blackBox.Point)

	default: //库存下限 < 库存值 < 库存上限
		//不处理
	}
	return -1
}
func (this *RollPointSceneData) CalcLastWiner() {
	var winTimesBig *RollPointPlayerData
	winTimesBig = nil
	for _, v := range this.seats {
		if v.IsMarkFlag(base.PlayerState_GameBreak) { //观众
			continue
		}
		if v.cGetWin20 <= 0 { //近20局胜率不能为0
			continue
		}
		if winTimesBig == nil {
			winTimesBig = v
		} else {
			if winTimesBig.cGetWin20 < v.cGetWin20 {
				winTimesBig = v
			}
		}
	}
	if winTimesBig != nil {
		this.LastWinerSnid = winTimesBig.SnId
	}
}
