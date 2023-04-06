package fishing

import (
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	server_proto "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/task"
	"os"
	"strconv"
	"time"
)

type FishingPlayerEvent struct {
	FishId    int32
	Event     int32
	Power     int32
	Ts        int64
	DropCoin  int32
	ExtFishId []int32
}
type FishingPlayerData struct {
	*base.Player
	scene         *base.Scene //玩家当前所在场景
	leavetime     int32
	power         int32
	CoinCache     int64
	bullet        map[int32]*Bullet
	BulletLimit   [BULLETLIMIT]int64
	fishEvent     map[string]*FishingPlayerEvent
	LostCoin      int64
	LostCoinCache int64
	AgentParam    int32   //因代码改动不完全，所以机器人该字段表示代理玩家的id；玩家该字段表示其代理的其中一个机器人的id
	RobotSnIds    []int32 //真实玩家绑定的机器人数组
	AutoFishing   int32
	SelTarget     int32
	TargetFish    int32
	FireRate      int32
	EventTs       int64
	FishPoolKey   string
	taxCoin       float64
	sTaxCoin      float64 //单次记录税率
	enterTime     time.Time
	// start   捕鱼玩法优化第一版,添加炮台定义类型,免费炮台炮弹的数量
	SelVip       int32 //玩家选择的vip炮等级
	powerType    int32 //玩家炮台的类型
	FreePowerNum int32 //免费炮台炮弹的数量
	// end
	winCoin          int64
	Odds             int
	realOdds         int //  玩家的实际赔率
	logBulletHitNums int32
	logFishCount     map[int64]*model.FishCoinNum
	lockFishCount    map[int32]int32
	jackpotCoin      float64 // 用来收集不足1的金币
	Prana            float64 // 蓄能能量
	PranaPercent     int32   // 蓄能能量百分比
	FishCoinLoss     int32
	MaxCoin          int64 // 游戏场最大金币
	MinCoin          int64 // 游戏场最少金币
	ExtraCoin        int64 // 除捕鱼外额外获得的金币
	EnterCoin        int64 // 进场金币 统计使用
	TestHitNum       int64 //测试碰撞次数
}

/*
更新免费炮台相关状态
*/
func (this *FishingPlayerData) UpdateFreePowerState() {
	this.powerType = FreePowerType
	this.FreePowerNum = 100
}

/*
更新成普通炮台相关状态
*/
func (this *FishingPlayerData) UpdateNormalPowerState() {
	this.powerType = NormalPowerType
	this.FreePowerNum = 0
}

/*
更新成钻头炮台相关状态
*/
func (this *FishingPlayerData) UpdateBitPowerState() {
	this.powerType = BitPowerType
}

func (this *FishingPlayerData) MakeLogKey(fishTemplateId, power int32) int64 {
	return common.MakeI64(fishTemplateId, power)
}

func (this *FishingPlayerData) SplitLogKey(key int64) (int32, int32) {
	return common.LowI32(key), common.HighI32(key)
}

func (this *FishingPlayerData) init(s *base.Scene) {
	//this.Player.extraData = this
	this.Player.SetExtraData(this)
	this.LostCoin = 0
	this.LostCoinCache = 0
	this.power = 1
	this.AgentParam = 0
	this.bullet = make(map[int32]*Bullet)
	this.BulletLimit = [BULLETLIMIT]int64{}
	this.fishEvent = make(map[string]*FishingPlayerEvent)
	this.logBulletHitNums = 0
	this.lockFishCount = make(map[int32]int32)
	this.logFishCount = make(map[int64]*model.FishCoinNum)
	this.jackpotCoin = 0.0
	this.TestHitNum = 0
	this.scene = s
	//初始化炮倍
	powers := s.GetDBGameFree().GetOtherIntParams()
	if len(powers) != 0 {
		this.power = powers[0]
	}
	if this.GDatas == nil {
		this.GDatas = make(map[string]*model.PlayerGameInfo)
	}
	if !s.GetTesting() && !this.IsRob {
		key := s.KeyGamefreeId
		var pgd *model.PlayerGameInfo
		if data, exist := this.GDatas[key]; !exist {
			pgd = new(model.PlayerGameInfo)
			this.GDatas[key] = pgd
		} else {
			pgd = data
		}
		if pgd != nil {
			//参数确保
			for i := len(pgd.Data); i < GDATAS_HPFISHING_MAX; i++ {
				pgd.Data = append(pgd.Data, 0)
			}
			this.Prana = float64(pgd.Data[GDATAS_HPFISHING_PRANA])
			//this.SelVip = int32(pgd.Data[GDATAS_FISHING_SELVIP])
		}
		////////////////////////
		var npgd *model.PlayerGameInfo
		if ndata, exist := this.GDatas[s.KeyGameId]; !exist {
			npgd = new(model.PlayerGameInfo)
			this.GDatas[s.KeyGameId] = npgd
		} else {
			npgd = ndata
		}
		if npgd != nil {
			//参数确保
			for i := len(npgd.Data); i < GDATAS_HPFISHING_MAX; i++ {
				npgd.Data = append(npgd.Data, 0)
			}
			this.SelVip = int32(npgd.Data[GDATAS_FISHING_SELVIP])
		}
	}
	this.Clean()
}
func (this *FishingPlayerData) SetSelVip(keyGameId string) {
	if pgd, ok := this.GDatas[keyGameId]; ok {
		pgd.Data[GDATAS_FISHING_SELVIP] = int64(this.SelVip)
	}
}

func (this *FishingPlayerData) Clean() {
	this.bullet = make(map[int32]*Bullet)
	this.BulletLimit = [BULLETLIMIT]int64{}
	this.fishEvent = make(map[string]*FishingPlayerEvent)
	this.logBulletHitNums = 0
	this.bullet = make(map[int32]*Bullet)
	this.logFishCount = make(map[int64]*model.FishCoinNum)
}
func (this *FishingPlayerData) CoinCheck(power int32) bool {
	return this.CoinCache >= int64(power)
}
func (this *FishingPlayerData) CurrentCoin() int64 {
	return this.CoinCache
}

func (this *FishingPlayerData) GetTodayGameData(gameId string) *model.PlayerGameStatics {
	if this.TodayGameData == nil {
		this.TodayGameData = model.NewPlayerGameCtrlData()
	}
	if this.TodayGameData.CtrlData == nil {
		this.TodayGameData.CtrlData = make(map[string]*model.PlayerGameStatics)
	}
	if _, ok := this.TodayGameData.CtrlData[gameId]; !ok {
		this.TodayGameData.CtrlData[gameId] = &model.PlayerGameStatics{}
	}
	return this.TodayGameData.CtrlData[gameId]
}

/*
设置当天数据
*/
func (this *FishingPlayerData) SetTodayGameDate(gameId string, playerGameStatics *model.PlayerGameStatics) {
	this.TodayGameData.CtrlData[gameId] = playerGameStatics
}

/*
获取昨日得当天数据集合
*/
func (this *FishingPlayerData) GetYestDayGameData(gameId string) *model.PlayerGameStatics {
	if this.YesterdayGameData == nil {
		this.YesterdayGameData = &model.PlayerGameCtrlData{}
	}
	if this.YesterdayGameData.CtrlData == nil {
		this.YesterdayGameData.CtrlData = make(map[string]*model.PlayerGameStatics)
	}
	if _, ok := this.YesterdayGameData.CtrlData[gameId]; !ok {
		this.YesterdayGameData.CtrlData[gameId] = &model.PlayerGameStatics{}
	}
	return this.YesterdayGameData.CtrlData[gameId]
}

func (this *FishingPlayerData) SaveDetailedLog(s *base.Scene) {
	if len(this.logFishCount) == 0 {
		return
	}
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		if sceneEx.GetTesting() && this.IsRob {
			this.logFishCount = make(map[int64]*model.FishCoinNum)
			return
		}
		totalin := int64(0)
		totalout := int64(0)
		var fd model.FishDetiel
		logBulletCount := make(map[int32]int32)
		fcn := make([]model.FishCoinNum, 0)
		for k, v := range this.logFishCount {
			FishTemplateId, power := this.SplitLogKey(k)
			fcn = append(fcn, model.FishCoinNum{ID: FishTemplateId, Power: power, Num: v.Num, Coin: v.Coin, HitNum: v.HitNum})
			totalout += int64(v.Coin)
			logBulletCount[power] += v.HitNum
		}
		fd.HitInfo = &fcn
		bt := make([]model.BulletLevelTimes, 0)
		for k, v := range logBulletCount {
			bt = append(bt, model.BulletLevelTimes{Level: k, Times: v})
			totalin += int64(k * v)
		}
		fd.BulletInfo = &bt
		fp := &model.FishPlayerData{
			UserId:   this.SnId,
			UserIcon: this.Head,
			TotalIn:  totalin,
			TotalOut: totalout,
			CurrCoin: this.CoinCache,
		}
		// 捕鱼不需要个人信息里的战绩
		//win := totalout - totalin
		//var isWin int32
		//if win > 0 {
		//	isWin = 1
		//} else if win < 0 {
		//	isWin = -1
		//}
		//sceneEx.SaveFriendRecord(this.SnId, isWin)
		fd.PlayData = fp
		info, err := model.MarshalGameNoteByFISH(&fd)
		if err == nil {
			logid, _ := model.AutoIncGameLogId()
			validFlow := totalin + totalout
			validBet := common.AbsI64(totalin - totalout)
			param := base.GetSaveGamePlayerListLogParam(this.Platform, this.Channel, this.BeUnderAgentCode, this.PackageID, logid,
				this.InviterId, totalin, totalout, int64(this.sTaxCoin), 0, totalin,
				totalout, validFlow, validBet, sceneEx.IsPlayerFirst(this.Player), false)
			sceneEx.SaveGamePlayerListLog(this.SnId, param)
			sceneEx.SaveGameDetailedLog(logid, info, &base.GameDetailedParam{})
		}

		pack := &server_proto.GWFishRecord{
			GameFreeId: proto.Int32(sceneEx.GetDBGameFree().GetId()),
			SnId:       proto.Int32(this.SnId),
		}
		for _, v := range this.logFishCount {
			fishRecord := &server_proto.FishRecord{
				FishId: proto.Int32(v.ID),
				Count:  proto.Int32(v.Num),
			}
			pack.FishRecords = append(pack.FishRecords, fishRecord)
		}
		if len(pack.FishRecords) > 0 {
			this.SendToWorld(int(server_proto.SSPacketID_PACKET_GW_FISHRECORD), pack)
		}

		diffLostCoin := this.LostCoin - this.LostCoinCache
		this.LostCoinCache = this.LostCoin
		gain := totalout - totalin
		this.Statics(s.KeyGameId, s.KeyGamefreeId, gain, true)
		if diffLostCoin > 0 {
			playerBet := &server_proto.PlayerBet{
				SnId:       proto.Int32(this.SnId),
				Bet:        proto.Int64(totalin),
				Gain:       proto.Int64(gain),
				Tax:        proto.Int64(int64(this.sTaxCoin)),
				Coin:       proto.Int64(this.Coin),
				GameCoinTs: proto.Int64(this.GameCoinTs),
			}
			gwPlayerBet := &server_proto.GWPlayerBet{
				SceneId:    proto.Int(sceneEx.SceneId),
				GameFreeId: proto.Int32(sceneEx.GetDBGameFree().GetId()),
				RobotGain:  proto.Int64(-(this.CoinCache - this.GetCoin())),
			}
			gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
			proto.SetDefaults(gwPlayerBet)
			sceneEx.SendToWorld(int(server_proto.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
		}
	}
	this.sTaxCoin = 0
	this.logBulletHitNums = 0
	this.logFishCount = make(map[int64]*model.FishCoinNum)
}

func (this *FishingPlayerData) SetMaxCoin() {
	if this.CoinCache > this.MaxCoin {
		this.MaxCoin = this.CoinCache
	}
}

func (this *FishingPlayerData) SetMinCoin() {
	if this.CoinCache < this.MinCoin || this.MinCoin == 0 {
		this.MinCoin = this.CoinCache
	}
}

func (this *FishingPlayerData) SaveFishingLog(curCoin int64, gameid string) {
	data := this.GDatas[gameid]
	log := fmt.Sprintf("%v,%v,%v,%v,%v\n", this.CurrentCoin(), data.Statics.TotalIn, data.Statics.TotalOut, curCoin, base.GetCoinPoolMgr().GetTax())
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		fileName := fmt.Sprintf("fishdata-%v.csv", this.SnId)
		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
		defer file.Close()
		if err != nil {
			file, err = os.Create(fileName)
			if err != nil {
				return err
			}
		}
		file.WriteString(log)
		return nil
	}), nil, "SaveFishingLog").StartByFixExecutor("SaveFishingLog")
}

// NewStatics .b
func (this *FishingPlayerData) NewStatics(betCoin, gain int64) {
	if this.scene == nil || this.scene.GetTesting() || this.IsRob { //测试场和机器人不统计
		return
	}
	if betCoin == 0 && gain == 0 {
		return
	}
	// start 如果当前处于免费炮的情况,NewStatics 记录消费为0
	if this.powerType == FreePowerType {
		betCoin = 0
	}
	// end

	//黑白名单不参与投入产出统计，影响自己和他人体验
	if this.WhiteLevel != 0 || this.WhiteFlag != 0 || this.BlackLevel != 0 || this.GMLevel > 0 {
		return
	}

	fishlogger.Tracef("============snid=%v betCoin=%v  gain=%v ", this.SnId, betCoin, gain)

	keyGlobal := fmt.Sprintf("%v_%v", this.scene.GetPlatform(), this.scene.GetGameFreeId())
	if base.SysProfitCoinMgr.SysPfCoin != nil {
		if base.SysProfitCoinMgr.SysPfCoin.ProfitCoin == nil {
			base.SysProfitCoinMgr.SysPfCoin.ProfitCoin = make(map[string]*model.SysCoin)
		}
		var syscoin *model.SysCoin
		if data, exist := base.SysProfitCoinMgr.SysPfCoin.ProfitCoin[keyGlobal]; !exist {
			syscoin = new(model.SysCoin)
			base.SysProfitCoinMgr.SysPfCoin.ProfitCoin[keyGlobal] = syscoin
		} else {
			syscoin = data
		}
		syscoin.PlaysBet += betCoin
		syscoin.SysPushCoin += gain
		fishlogger.Tracef("============SysProfitCoinMgr key=%v PlaysBet:= %v SysPushCoin= %v ", keyGlobal, syscoin.PlaysBet, syscoin.SysPushCoin)
	}

	keyPlayer := strconv.Itoa(int(this.scene.GetGameFreeId()))
	var pgd *model.PlayerGameInfo
	if d, exist := this.GDatas[keyPlayer]; exist {
		FishGDataLen(len(d.Data), d)
		pgd = d
	} else {
		pgd = &model.PlayerGameInfo{
			Data: make([]int64, GDATAS_HPFISHING_MAX, GDATAS_HPFISHING_MAX),
		}
		this.GDatas[keyPlayer] = pgd
	}

	low := pgd.Data[GDATAS_HPFISHING_ALLBET]
	high := pgd.Data[GDATAS_HPFISHING_ALLBET64]
	allBet := common.MakeI64(int32(low), int32(high)) + betCoin

	pgd.Data[GDATAS_HPFISHING_ALLBET], pgd.Data[GDATAS_HPFISHING_ALLBET64] = common.LowAndHighI64(allBet)
	pgd.Data[GDATAS_HPFISHING_CHANGEBET] += int64(gain)

	fishlogger.Tracef("============snid=%v total fish betCoin:= %v  gain=%v ", this.SnId, allBet,
		pgd.Data[GDATAS_HPFISHING_CHANGEBET])
}

// GetAllBet .
func (this *FishingPlayerData) GetAllBet(key string) int64 {
	ret := int64(1)
	if d, exist := this.GDatas[key]; exist {
		FishGDataLen(len(d.Data), d)
		ret = common.MakeI64(int32(d.Data[GDATAS_HPFISHING_ALLBET]), int32(d.Data[GDATAS_HPFISHING_ALLBET64]))
	}
	return ret
}

// GetAllChangeBet .
func (this *FishingPlayerData) GetAllChangeBet(key string) int64 {
	ret := int64(0)
	if d, exist := this.GDatas[key]; exist {
		FishGDataLen(len(d.Data), d)
		ret = int64(d.Data[GDATAS_HPFISHING_CHANGEBET])
	}
	return ret
}

// 计算个人赔率和个人限制系数
func (this *FishingPlayerData) GetPlayerOdds(gameid string, ctroRate int32, fishlevel int32) (float64, float64) {
	if data, ok := this.GDatas[gameid]; ok {
		//总产出初始值 = 100*（1-调节频率）*100 （分）	初级场
		//总产出初始值 = 1000*（1-调节频率）*100 （分）	中级场
		//总产出初始值 = 10000*（1-调节频率）*100 （分）	高级场
		//总投入初始值 = 100*100 （分）	初级场
		//总投入初始值 = 1000*100 （分）	中级场
		//总投入初始值 = 10000*100 （分）	高级场
		initBaseValue := int64(10000) //1万分
		if fishlevel == 2 {
			initBaseValue = 100000
		} else if fishlevel == 3 {
			initBaseValue = 1000000
		}
		totalInValue := initBaseValue + data.Statics.TotalIn
		totalOutValue := initBaseValue*(10000-int64(ctroRate))/10000 + data.Statics.TotalOut

		//个人限制系数
		ratio := 1.0
		if fishlevel == 1 && totalOutValue-totalInValue >= 20000 {
			ratio = 0.5
		} else if fishlevel == 2 && totalOutValue-totalInValue >= 100000 {
			ratio = 0.5
		} else if fishlevel == 3 && totalOutValue-totalInValue >= 500000 {
			ratio = 0.5
		}

		return float64(totalOutValue) / float64(totalInValue), ratio
	} else {
		fishlogger.Errorf("player.GDatas[%v] is %v", gameid, this.GDatas[gameid])
		return 0, 0
	}

}

var FishGDataLen = func(flen int, pgd *model.PlayerGameInfo) {
	if flen < GDATAS_HPFISHING_MAX {
		for i := flen; i < GDATAS_HPFISHING_MAX; i++ {
			pgd.Data = append(pgd.Data, 0)
		}
	}
}

// 玩家离场时，返还场中还未发生碰撞的子弹
func (this *FishingPlayerData) RetBulletCoin(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		fishlogger.Tracef("(this *FishingPlayerData) RetBulletCoin, player:=%v retBulletNum=%v", this.SnId, len(this.bullet))
		fishID, coin, taxc := int32(0), int32(0), int64(0)
		for _, v := range this.bullet {
			coin += v.Power
			taxc += int64(float64(sceneEx.GetDBGameFree().GetTaxRate()) / 10000 * float64(v.Power))
		}
		sceneEx.RetBulletCoin(this, fishID, coin, taxc, false) // 合并后发送
		this.bullet = make(map[int32]*Bullet)
		this.BulletLimit = [BULLETLIMIT]int64{}
		this.fishEvent = make(map[string]*FishingPlayerEvent)
		this.bullet = make(map[int32]*Bullet)
		this.logFishCount = make(map[int64]*model.FishCoinNum)
	}
}
