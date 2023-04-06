package rollcoin

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/rollcoin"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
	"io/ioutil"
	"math"
	"time"
)

type RollCoinSceneData struct {
	*base.Scene                                      //房间信息
	players            map[int32]*RollCoinPlayerData //玩家信息
	seats              []*RollCoinPlayerData         //座位信息
	WinLog             []int64                       //开奖记录
	WinFlg             int32                         //开奖结果
	WinIndex           int32                         //开奖结果
	Banker             *RollCoinPlayerData           //玩家坐庄
	BankerList         []int32                       //申请坐庄列表
	BankCoinLimit      int                           //上庄金币限制
	CoinLimit          int                           //进入房间金币限制
	SyncTime           time.Time                     //同步押注金币的时间
	TotalBet           map[int32]map[int32]int64     //总下注数据
	SyncData           map[int32]map[int32]int64     //同步的数据
	ZoneTotal          map[int32]int64               //区域总下注额
	ZoneBetIsFull      map[int32]bool                //区域总下注是否已经满
	ZoneRobTotal       map[int32]int64               //机器人区域总下注额
	syncCnt            int                           //同步数量
	Rollcoin           []int32
	NoBanker           int32
	by                 func(p, q *RollCoinPlayerData) bool
	godid              int32 //神算子
	downBanker         bool  //是否下局下庄
	RollCoinTargetList []int32
}

//Len()
func (s *RollCoinSceneData) Len() int {
	return len(s.seats)
}

//Less():输赢记录将有高到底排序
func (s *RollCoinSceneData) Less(i, j int) bool {
	//return s.seats[i].GetBetGig20() > s.seats[j].GetBetGig20()
	return s.by(s.seats[j], s.seats[i])
}

//Swap()
func (s *RollCoinSceneData) Swap(i, j int) {
	s.seats[i], s.seats[j] = s.seats[j], s.seats[i]
}

func NewRollCoinSceneData(s *base.Scene) *RollCoinSceneData {
	return &RollCoinSceneData{
		Scene:         s,
		players:       make(map[int32]*RollCoinPlayerData),
		SyncData:      make(map[int32]map[int32]int64),
		TotalBet:      make(map[int32]map[int32]int64),
		ZoneTotal:     make(map[int32]int64),
		ZoneRobTotal:  make(map[int32]int64),
		ZoneBetIsFull: make(map[int32]bool),
	}
}

func (this *RollCoinSceneData) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
}

func (this *RollCoinSceneData) init() bool {
	this.WithLocalAI = true
	this.BindAIMgr(&RollCoinSceneAIMgr{})

	data := this.DbGameFree
	this.RollCoinTargetList = []int32{7, 3, 6, 2, 5, 1, 4, 0, 7, 3, 6, 2, 5, 1, 4, 0, 7, 3, 6, 2, 5, 1, 4, 0, 7, 3, 6, 2, 5, 1, 4, 0}
	if data != nil {
		this.BankCoinLimit = int(data.GetBanker())
		this.CoinLimit = int(data.GetBetLimit())
		this.Rollcoin = data.GetOtherIntParams()
	} else {
		logger.Logger.Error("Init roll coin scene failed,scene type:", this.KeyGamefreeId)
	}

	base.SystemChanceMgrEx.Reload()
	logger.Logger.Trace("RollCoinSceneData init")

	//初始化区域和面额数据
	for _, id := range base.SystemChanceMgrEx.RollCoinIds {
		this.TotalBet[id] = make(map[int32]int64)
		this.ZoneTotal[id] = 0
		this.ZoneRobTotal[id] = 0
		for _, coin := range this.Rollcoin {
			this.TotalBet[id][coin] = 0
		}
	}
	return true
}

func (this *RollCoinSceneData) Clean() {
	for key, _ := range this.TotalBet {
		this.ZoneTotal[key] = 0
		this.ZoneBetIsFull[key] = false
		this.ZoneRobTotal[key] = 0
		for coin, _ := range this.TotalBet[key] {
			this.TotalBet[key][coin] = 0
		}
	}
	for _, p := range this.players {
		p.Clean()
	}
	this.WinFlg = 0
	this.WinIndex = 0
	//重置水池调控标记
	this.CpControlled = false
}

func (this *RollCoinSceneData) delPlayer(p *base.Player) {
	if p, exist := this.players[p.SnId]; exist {
		delete(this.players, p.SnId)
	}
}
func (this *RollCoinSceneData) CountPlayer() int {
	num := len(this.players)
	if num >= 21 {
		//truePlayerCount := int32(len(this.players))
		dbGame := this.DbGameFree
		if dbGame != nil {
			//获取fake用户数量//todo
			//correctNum := dbGame.GetCorrectNum()
			//correctRate := dbGame.GetCorrectRate()
			//fakePlayerCount := correctNum + truePlayerCount*correctRate/100 + dbGame.GetDeviation()
			//num = int(truePlayerCount + fakePlayerCount + rand.Int31n(truePlayerCount))
		}
	}
	return num
}
func (this *RollCoinSceneData) OnPlayerLeave(p *base.Player, reason int) {
	this.delPlayer(p)
	this.BroadcastPlayerLeave(p)
	p.UnmarkFlag(base.PlayerState_GameBreak)
	p.SyncFlag(true)
	//离开游戏房间时，会自动取消上庄
	index := -1
	for key, value := range this.BankerList {
		if value == p.SnId {
			index = key
		}
	}
	if index != -1 {
		this.BankerList = append(this.BankerList[:index], this.BankerList[index+1:]...)
		pack := &rollcoin.SCRollCoinBankerList{}
		pack.Insert = proto.Bool(false)
		for _, value := range this.BankerList {
			banker := this.players[value]
			if banker == nil {
				continue
			}
			pack.List = append(pack.List, &rollcoin.RollCoinPlayer{
				SnId:        proto.Int32(banker.SnId),
				Name:        proto.String(banker.Name),
				Sex:         proto.Int32(banker.Sex),
				Head:        proto.Int32(banker.Head),
				Coin:        proto.Int64(banker.GetCoin() - banker.betTotal),
				HeadOutLine: proto.Int32(banker.HeadOutLine),
				VIP:         proto.Int32(banker.VIP),
				City:        proto.String(banker.GetCity()),
			})
		}
		proto.SetDefaults(pack)
		this.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_BANKERLIST), pack, 0)
	}
}

func (this *RollCoinSceneData) BroadcastPlayerLeave(p *base.Player) {
	scLeavePack := &rollcoin.SCRollCoinPlayerNum{
		PlayerNum: proto.Int(this.CountPlayer()),
	}
	proto.SetDefaults(scLeavePack)
	this.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_PLAYERNUM), scLeavePack, p.GetSid())
}

func (this *RollCoinSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

func (this *RollCoinSceneData) GetNormalPos() int32 {

	//计算区域开奖概率
	area := make(map[int32]int32)
	for _, id := range base.SystemChanceMgrEx.RollCoinIds {
		rate := base.SystemChanceMgrEx.RollCoinRate[id]
		if rate == 0 {
			area[id] += 0
		} else {
			area[id] += int32(float64(100000) / float64(rate))
		}
	}
	//赋值区域开奖概率
	base.SystemChanceMgrEx.RollCoinChance = nil
	for key, value := range area {
		base.SystemChanceMgrEx.RollCoinChance = append(base.SystemChanceMgrEx.RollCoinChance, int64(key), int64(value))
	}

	rates := []int64{}
	for key, value := range area {
		rates = append(rates, int64(key), int64(value))
	}

	winFlag := int32(common.RandItemByWight(rates))
	this.WinFlg = winFlag

	return winFlag
}

func (this *RollCoinSceneData) GetWinFlg() int32 {
	//玩家庄处理
	areapos := this.BankerProess()
	if areapos != -1 {
		this.WinFlg = areapos
		return areapos
	}

	rateArray := make([]int64, len(base.SystemChanceMgrEx.RollCoinRate))
	for k, v := range base.SystemChanceMgrEx.RollCoinRate {
		rateArray[k] = int64(v)

	}

	minRate := common.GetMinCommonRate(rateArray) * 10000
	totalRate := int32(0)
	//计算每个区域的开奖权重
	//动态返奖率
	wf := base.CoinPoolMgr.GetRTP(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())

	//计算区域开奖概率
	area := make(map[int32]int32)
	for _, id := range base.SystemChanceMgrEx.RollCoinIds {
		failcoin := int64(0)
		wincoin := int64(0)
		//计算每个区域所有玩家总产出与总投入 总产出/总投入
		for _, player := range this.players {
			if !player.IsRob && player.CoinPool[id] != 0 {
				info := this.GetTotalTodayDaliyGameData(this.DbGameFree.GetGameDif(), player.Player)
				if info != nil {
					failcoin += info.TotalIn
					wincoin += info.TotalOut
				}
			}
		}

		if model.NormalParamData.IsCloseAddRtp == 0 {
			rate := base.SystemChanceMgrEx.RollCoinRate[id]
			setting := base.CoinPoolMgr.GetCoinPoolSetting(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())

			wf := float64(wincoin+300000) / float64(failcoin*int64(10000-setting.GetCoinPoolMode())/10000+300000)

			if rate == 0 {
				area[id] = 0
			} else {
				wf = float64(model.NormalParamData.NormalAddBase) /
					(1.0 + math.Pow(float64(model.NormalParamData.NormalAddBase-1.0), wf))
				area[id] = int32(float64(minRate)*wf) / rate
			}

			if area[id] < 0 {
				area[id] = 0
			}
			totalRate += area[id]
		} else {
			rate := base.SystemChanceMgrEx.RollCoinRate[id]
			if rate == 0 {
				area[id] += 0
			} else {
				area[id] += int32(float64(minRate) * wf / float64(rate))
				totalRate += area[id]
			}
		}
	}
	//赋值区域开奖概率
	base.SystemChanceMgrEx.RollCoinChance = nil
	for key, value := range area {
		base.SystemChanceMgrEx.RollCoinChance = append(base.SystemChanceMgrEx.RollCoinChance, int64(key), int64(value))
	}
	//未中奖区域
	if int32(minRate-int64(totalRate)) > 0 {
		area[int32(len(base.SystemChanceMgrEx.RollCoinRate))] = int32(minRate - int64(totalRate))
	}

	rates := []int64{}
	for key, value := range area {
		rates = append(rates, int64(key), int64(value))
	}

	winFlag := int32(common.RandItemByWight(rates))

	if winFlag == int32(len(base.SystemChanceMgrEx.RollCoinRate)) {
		if this.Banker != nil && !this.Banker.IsRob {
			winFlag = this.BankerProess()
			if winFlag == -1 {
				winFlag = int32(this.GetMinPlayerBetPos()) //玩家要输钱
			}
			this.CpControlled = true
		} else {
			winFlag = int32(this.GetMinPlayerBetPos()) //玩家要输钱
			this.CpControlled = true
		}
	}

	this.WinFlg = winFlag

	return winFlag
}

func (this *RollCoinSceneData) BankerProess() int32 {
	//玩家庄时处理
	areapos := []int32{}
	if this.Banker != nil && !this.Banker.IsRob { //玩家庄
		for _, id := range base.SystemChanceMgrEx.RollCoinIds {
			RobCoin := this.ZoneRobTotal[id] * int64(base.SystemChanceMgrEx.RollCoinRate[id])
			if RobCoin > this.GetRobAllCoin() {
				areapos = append(areapos, id)
			}
		}
		wf := 0
		if d, exist := this.Banker.GDatas[this.KeyGamefreeId]; exist {
			wf = int(float32(d.Statics.TotalOut+1) / float32(d.Statics.TotalIn+1) * 100)
		}
		if wf >= 95 {
			if len(areapos) > 0 && common.RandInt(100) <= 60 {
				return areapos[common.RandInt(len(areapos))]
			}
		} else {
			robpos := int32(this.GetMinRobBetPos())
			if robpos != -1 && common.RandInt(100) <= 60 {
				return robpos
			}
		}
	}
	return -1
}

func (this *RollCoinSceneData) BankerMinProess() int32 {
	//玩家庄时处理
	areapos := []int32{}
	if this.Banker != nil && !this.Banker.IsRob { //玩家庄
		for _, id := range base.SystemChanceMgrEx.RollCoinIds {
			RobCoin := this.ZoneRobTotal[id] * int64(base.SystemChanceMgrEx.RollCoinRate[id])
			if RobCoin > this.GetRobAllCoin() {
				areapos = append(areapos, id)
			}
		}

		robpos := int32(this.GetMinRobBetPos())
		if robpos != -1 && common.RandInt(100) <= 60 {
			return robpos
		}

	}
	return -1
}

func (this *RollCoinSceneData) GetRobAllCoin() int64 {
	robAllCoin := int64(0)
	for _, id := range base.SystemChanceMgrEx.RollCoinIds {
		robAllCoin += this.ZoneRobTotal[id]
	}
	return robAllCoin
}

func (this *RollCoinSceneData) LogWinFlag() {
	this.WinLog = append(this.WinLog, int64(this.WinFlg))
	if len(this.WinLog) > 25 {
		this.WinLog = this.WinLog[1:]
	}
}
func (this *RollCoinSceneData) TryChangeBanker() {
	if len(this.BankerList) > 0 { //列表里有人
		if this.Banker != nil { //原来有庄
			this.Banker.BankerTimes = 0
		}
		this.Banker = this.players[this.BankerList[0]]
		this.BankerList = this.BankerList[1:]
	} else { //列表里没人
		if this.Banker != nil { //原来有庄
			this.Banker.BankerTimes = 0
			this.Banker = nil
		} else { //原来也没庄
			this.Banker = nil
		}
	}
	pack := &rollcoin.SCRollCoinBanker{
		OpRetCode: rollcoin.OpResultCode_OPRC_Sucess,
	}
	if this.Banker != nil {
		this.Banker.BankerTimes = 1
		pack.Banker = &rollcoin.RollCoinPlayer{
			SnId:        proto.Int32(this.Banker.SnId),
			Name:        proto.String(this.Banker.Name),
			Sex:         proto.Int32(this.Banker.Sex),
			Head:        proto.Int32(this.Banker.Head),
			HeadOutLine: proto.Int32(this.Banker.HeadOutLine),
			Coin:        proto.Int64(this.Banker.Coin),
			//Params:      this.Banker.Params,
			Flag:        proto.Int(this.Banker.GetFlag()),
			VIP:         proto.Int32(this.Banker.VIP),
			BankerTimes: proto.Int32(this.Banker.BankerTimes),
			City:        proto.String(this.Banker.GetCity()),
		}
		this.NoBanker = 0
	} else {
		this.NoBanker++
	}
	proto.SetDefaults(pack)
	this.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_BANKER), pack, 0)
	logger.Logger.Trace("SCRollCoinBanker:", pack)
	//if this.NoBanker > int32(common.RandInt(0, 3)) {
	//	robot := []*RollCoinPlayerData{}
	//	for _, value := range this.players {
	//		if value.IsRob {
	//			robot = append(robot, value)
	//		}
	//	}
	//	if len(robot) > 0 {
	//		banker := robot[common.RandInt(len(robot))]
	//		banker.AddCoin(int64(this.BankCoinLimit), common.GainWay_ByPMCmd, false, banker.GetName(),
	//			this.GetSceneName())
	//		this.BankerList = append(this.BankerList, banker.SnId)
	//		this.NoBanker = 0
	//	}
	//}
}
func (this *RollCoinSceneData) BankerListDelCheck() {
	delArr := []int32{}
	for _, id := range this.BankerList {
		banker, ok := this.players[id]
		if !ok {
			delArr = append(delArr, id)
			continue
		}
		if banker.Coin < int64(this.BankCoinLimit) {
			delArr = append(delArr, id)
		}
	}
	if len(delArr) > 0 {
		for _, id := range delArr {
			index := -1
			for key, value := range this.BankerList {
				if value == id {
					index = key
				}
			}
			if index != -1 {
				this.BankerList = append(this.BankerList[:index], this.BankerList[index+1:]...)
			}
		}
	}
}

func (this *RollCoinSceneData) InBankerList(snid int32) bool {
	for _, value := range this.BankerList {
		if value == snid {
			return true
		}
	}
	return false
}

func (this *RollCoinSceneData) CacheCoinLog(flag, coin int32, isRob bool) {
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
func (this *RollCoinSceneData) CheckBetIsFull() bool {
	if this.Banker == nil {
		return false
	}
	for _, isFull := range this.ZoneBetIsFull {
		if !isFull {
			return false
		}
	}
	return true
}
func (this *RollCoinSceneData) SyncCoinLog() {
	if this.syncCnt > 0 {
		pack := &rollcoin.SCRollCoinCoinLog{}
		for _, id := range base.SystemChanceMgrEx.RollCoinIds {
			coinLog := &rollcoin.PushCoinLog{}
			var sumTickBet = int64(0)
			if pool, ok := this.SyncData[id]; ok {
				for coinType, coinNum := range pool {
					coinLog.Coin = append(coinLog.Coin, coinType)
					coinLog.Num = append(coinLog.Num, int32(coinNum))
					sumTickBet += coinNum * int64(coinType)
				}
			}
			pack.CoinPool = append(pack.CoinPool, coinLog)
			pack.TotalBet = append(pack.TotalBet, int64(this.ZoneTotal[id]))
			pack.TickBet = append(pack.TickBet, sumTickBet)
		}
		proto.SetDefaults(pack)
		this.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_COINLOG), pack, 0) //广播给其他用户
		logger.Logger.Trace("SCRollCoinCoinLog:", pack)
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
	this.SyncTime = time.Now()
}

//计算出玩家输赢
func (this *RollCoinSceneData) CalcuSystemCoinOut(winFlag int, billed bool) int64 {
	//计算系统产出
	systemOut := int64(0) //统计系统产出的金币

	if this.Banker != nil && !this.Banker.IsRob { //机器人庄
		for _, id := range base.SystemChanceMgrEx.RollCoinIds {
			if winFlag == int(id) {
				systemOut -= this.ZoneRobTotal[id] * int64(base.SystemChanceMgrEx.RollCoinRate[winFlag])
				systemOut += this.ZoneRobTotal[id]
			} else {
				if !billed {
					systemOut += this.ZoneRobTotal[id]
				}
			}
		}
	} else { //玩家庄
		for _, id := range base.SystemChanceMgrEx.RollCoinIds {
			playerCoin := this.ZoneTotal[id] - this.ZoneRobTotal[id]
			if winFlag == int(id) {
				systemOut += playerCoin * int64(base.SystemChanceMgrEx.RollCoinRate[winFlag])
				systemOut -= playerCoin
			} else {
				if !billed {
					systemOut -= playerCoin
				}
			}
		}
	}
	return systemOut
}

//取玩家最大赢钱位置
func (this *RollCoinSceneData) GetMaxPlayerBetPos(moremax bool) int {
	//计算区域开奖概率
	area := make(map[int32]int32)
	negativeCheck := false
	for _, id := range base.SystemChanceMgrEx.RollCoinIds {
		failcoin := int64(0)
		wincoin := int64(0)
		playernum := int32(0)
		//计算每个区域所有玩家总产出与总投入
		for _, player := range this.players {
			if !player.IsRob && player.CoinPool[id] != 0 {
				info := this.GetTotalTodayDaliyGameData(this.DbGameFree.GetGameDif(), player.Player)
				if info != nil {
					failcoin += info.TotalIn
					wincoin += info.TotalOut
					if info.TotalIn < 0 || info.TotalOut < 0 {
						negativeCheck = true
					}
				}
				playernum++
			}
		}

		if failcoin != -1 && wincoin != -1 {
			rate := base.SystemChanceMgrEx.RollCoinRate[id]
			setting := base.CoinPoolMgr.GetCoinPoolSetting(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())

			wf := float64(wincoin+300000) / float64(failcoin*int64(10000-setting.GetCoinPoolMode())/10000+300000)

			if rate == 0 {
				area[id] = 0
			} else {
				wf = float64(model.NormalParamData.MaxAddBase) /
					(1 + math.Pow(float64(model.NormalParamData.MaxAddBase-1), wf))
				area[id] = int32(float64(100000)*wf) / rate
			}

			if area[id] < 0 {
				area[id] = 0
				negativeCheck = true
			}
		}
	}
	//赋值区域开奖概率
	base.SystemChanceMgrEx.RollCoinChance = nil
	for key, value := range area {
		base.SystemChanceMgrEx.RollCoinChance = append(base.SystemChanceMgrEx.RollCoinChance, int64(key), int64(value))
	}
	if len(base.SystemChanceMgrEx.RollCoinChance) == 0 {
		return -1
	}
	if negativeCheck {
		type RollCoinSceneDataErr struct {
			RollCoinSceneData
			Player   map[int32]*RollCoinPlayerData
			RollData []int64
		}
		errData := RollCoinSceneDataErr{
			RollCoinSceneData: *this,
			Player:            this.players,
			RollData:          base.SystemChanceMgrEx.RollCoinChance,
		}
		buff, err := json.Marshal(errData)
		if err == nil {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				fileName := "rollcoin-memdata-" + time.Now().Format("2006-01-02-15-04-05") + ".json"
				return ioutil.WriteFile(fileName, buff, 0)
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			}), "RollCoinSceneErrorLog").Start()
		}
	}
	return int(common.RandItemByWight(base.SystemChanceMgrEx.RollCoinChance))
}

//取机器人最大赢钱位置
func (this *RollCoinSceneData) GetMaxRobBetPos() int {
	maxbet := int64(0)
	maxpos := -1
	for _, id := range base.SystemChanceMgrEx.RollCoinIds {
		robCoin := this.ZoneRobTotal[id] * int64(base.SystemChanceMgrEx.RollCoinRate[id])
		if maxbet < robCoin && robCoin != 0 {
			maxbet = robCoin
			maxpos = int(id)
		}
	}
	return maxpos
}

//取玩家最少赢钱位置
func (this *RollCoinSceneData) GetMinPlayerBetPos() int {
	//每个区域下注金币*区域倍数
	mCoin := make(map[int32]int64)
	//每个区域下注人数
	mPlayerNum := make(map[int32]int32)
	for _, id := range base.SystemChanceMgrEx.RollCoinIds {
		playerCoin := (this.ZoneTotal[id] - this.ZoneRobTotal[id]) * int64(base.SystemChanceMgrEx.RollCoinRate[id])
		mCoin[id] = playerCoin

		for _, player := range this.players {
			if !player.IsRob && player.CoinPool[id] != 0 {
				mPlayerNum[id]++
			}
		}
	}
	vs := base.SortMapByValue(mCoin)
	maxnum := 1
	num := 0
	//计算区域开奖概率
	//赋值区域开奖概率
	base.SystemChanceMgrEx.RollCoinChance = nil
	lastValue := int64(-1)
	for _, v := range vs {
		playerCoin := v.Value
		if lastValue != playerCoin {
			num++
			if num > maxnum {
				break
			}
		}
		lastValue = playerCoin
		base.SystemChanceMgrEx.RollCoinChance = append(base.SystemChanceMgrEx.RollCoinChance, int64(v.Key), playerCoin)

	}
	if len(base.SystemChanceMgrEx.RollCoinChance) == 0 {
		return int(this.GetNormalPos())
	}
	return int(common.RandItemByAvg(base.SystemChanceMgrEx.RollCoinChance))
}

//取机器人最少赢钱位置
func (this *RollCoinSceneData) GetMinRobBetPos() int {
	maxbet := int64(math.MaxInt64)
	maxpos := -1
	for _, id := range base.SystemChanceMgrEx.RollCoinIds {
		robCoin := this.ZoneRobTotal[id] * int64(base.SystemChanceMgrEx.RollCoinRate[id])
		if maxbet > robCoin && robCoin != 0 {
			maxbet = robCoin
			maxpos = int(id)
		}
	}
	return maxpos
}

func (this *RollCoinSceneData) ChangeCard(re int) int {
	sysOut := this.CalcuSystemCoinOut(re, false)

	if !base.CoinPoolMgr.IsMaxOutHaveEnough(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId(), -sysOut) {
		isFind := false
		for loop := 0; loop < 20; loop++ {
			winflag := int(this.GetWinFlg())
			re = winflag
			sysOut = this.CalcuSystemCoinOut(winflag, false)
			if base.CoinPoolMgr.IsMaxOutHaveEnough(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId(), -sysOut) {
				isFind = true
				break
			}
		}
		if !isFind {
			if this.Banker != nil && !this.Banker.IsRob {
				ret := int(this.BankerProess())
				if ret == -1 {
					return this.GetMinPlayerBetPos() //玩家要输钱
				}
			} else {
				return this.GetMinPlayerBetPos() //玩家要输钱
			}
		}
	}
	status, changeRate := base.CoinPoolMgr.GetCoinPoolStatus(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())
	//status = 2
	//changeRate = 1 //只要达到限制就进入调控
	switch status {
	case base.CoinPoolStatus_Normal:
		//如果“当前库存值-开奖区域奖励倍数X此区域押注额<库存下线”，那么本局游戏，系统从“区域奖励倍数X区域押注金额”最小的两个中随机选择一个开奖；
		curCoin := base.CoinPoolMgr.LoadCoin(this.GetGameFreeId(), this.GetPlatform(), this.GetGroupId())
		playerCoin := this.ZoneTotal[int32(re)] - this.ZoneRobTotal[int32(re)]
		rate := base.SystemChanceMgrEx.RollCoinRate[re]
		setting := base.CoinPoolMgr.GetCoinPoolSetting(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())
		if (curCoin - playerCoin*int64(rate)) < int64(setting.GetLowerLimit()) {
			if this.Banker != nil && !this.Banker.IsRob {
				ret := int(this.BankerProess())
				if ret == -1 {
					return this.GetMinPlayerBetPos() //玩家要输钱
				}
			} else {
				return this.GetMinPlayerBetPos() //玩家要输钱
			}
		}

	case base.CoinPoolStatus_Low: //库存值 < 库存下限
		if common.RandInt(10000) < changeRate {
			if this.CalcuSystemCoinOut(re, false) > 0 {
				if this.Banker != nil && !this.Banker.IsRob {
					ret := int(this.BankerMinProess())
					if ret == -1 {
						return this.GetMinPlayerBetPos() //玩家要输钱
					}
				} else {
					return this.GetMinPlayerBetPos() //玩家要输钱
				}
			}
		}
	case base.CoinPoolStatus_TooHigh: //库存>库存上限+偏移量
		fallthrough
	case base.CoinPoolStatus_High: //库存上限 < 库存值 < 库存上限+偏移量
		if common.RandInt(10000) < changeRate {
			if this.Banker != nil && !this.Banker.IsRob {
				return int(this.BankerProess())
			} else {
				winFlag := this.GetMaxPlayerBetPos(false) //玩家赢钱
				this.WinFlg = int32(winFlag)
				if !base.CoinPoolMgr.IsMaxOutHaveEnough(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId(), -sysOut) {
					isFind := false
					for loop := 0; loop < 20; loop++ {
						winflag := int(this.GetMaxPlayerBetPos(false))
						this.WinFlg = int32(winFlag)
						sysOut = this.CalcuSystemCoinOut(winflag, false)
						if base.CoinPoolMgr.IsMaxOutHaveEnough(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId(), -sysOut) {
							isFind = true
							break
						}
					}
					if !isFind {
						if this.Banker != nil && !this.Banker.IsRob {
							ret := int(this.BankerProess())
							if ret == -1 {
								return this.GetMinPlayerBetPos() //玩家要输钱
							}
						} else {
							return this.GetMinPlayerBetPos() //玩家要输钱
						}
					}
				}

			}
		}
		/*
			if common.RandInt(10000) < changeRate && this.CoinPoolCanOut() {
				systmCoinOut := this.CalcuSystemCoinOut(re, false)
				setting := base.CoinPoolMgr.GetCoinPoolSetting(this.platform, this.gamefreeId, this.groupId)
				if systmCoinOut < 0 && (-systmCoinOut) <= int64(setting.GetMaxOutValue()) {
					if this.Banker != nil && !this.Banker.IsRob {
						return int(this.BankerProess())
					} else {
						return this.GetMaxPlayerBetPos(true)
					}
				}
			}*/
	default: //CoinPoolStatus_Normal//库存下限 < 库存值 < 库存上限
		//不处理
	}
	return int(this.WinFlg)
}

type RollCoinSceneMemData struct {
	WinLog []int64
}

//序列化
func (this *RollCoinSceneData) Marshal() ([]byte, error) {
	md := RollCoinSceneMemData{
		WinLog: this.WinLog,
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&md)
	if err != nil {
		logger.Logger.Warnf("(this *RollCoinSceneData) Marshal() gob.Encode err:%v", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

//反序列化
func (this *RollCoinSceneData) Unmarshal(data []byte, ud interface{}) error {
	md := &RollCoinSceneMemData{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(md)
	if err != nil {
		logger.Logger.Warnf("(this *RollCoinSceneData) Unmarshal() gob.Encode err:%v", err)
		return err
	} else {
		logger.Logger.Trace("RollCoinSceneData unmarshal data:", *md)
	}
	if scece, ok := ud.(*base.Scene); ok {
		for _, value := range scece.Players {
			this.players[value.SnId] = value.ExtraData.(*RollCoinPlayerData)
			logger.Logger.Trace("RollCoinPlayerData:", *(this.players[value.SnId]))
		}
	}
	if scene, ok := ud.(*base.Scene); ok {
		this.Scene = scene
	}
	return nil
}

//发送庄家列表数据给机器人
func (this *RollCoinSceneData) SendRobotUpBankerList() {
	pack := &rollcoin.SCRollCoinBankerList{
		Count: proto.Int(len(this.BankerList)),
	}
	this.RobotBroadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_BANKERLIST), pack)
}
