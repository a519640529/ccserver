package rollpoint

import (
	rule "games.yol.com/win88/gamerule/rollpoint"
	"games.yol.com/win88/gamesrv/base"
)

type RollPointPlayerData struct {
	*base.Player
	gainCoin     int32                     //本局赢取的金币
	CoinPool     map[int32]int32           //押注的金币
	allBetCoin   int64                     //当局总下注筹码
	PushCoinLog  map[int32]map[int32]int32 //押注的记录,位置:押注:数量
	gameing      bool                      //是否参与游戏
	WinRecord    []int                     //最近20局输赢记录
	BetBigRecord []int                     //最近20局下注总额记录
	cGetWin20    int                       //返回玩家最近20局的获胜次数
	cGetBetGig20 int                       //返回玩家最近20局的下注总额
	taxCoin      int64                     //本局税收
	winCoin      int64                     //本局收税前赢的钱
}

func NewRollPointPlayerData(p *base.Player) *RollPointPlayerData {
	return &RollPointPlayerData{
		Player:      p,
		CoinPool:    make(map[int32]int32),
		PushCoinLog: make(map[int32]map[int32]int32),
	}
}

//玩家初始化
func (this *RollPointPlayerData) init() {

	this.Clean()
}

//返回玩家最近20局的获胜次数
func (this *RollPointPlayerData) GetWin20() int {
	Count := 0
	if len(this.WinRecord) > 20 {
		this.WinRecord = append(this.WinRecord[:0], this.WinRecord[1:]...)
	}
	if len(this.WinRecord) > 0 {
		for _, v := range this.WinRecord {
			Count += v
		}
	}
	return Count
}

//返回玩家最近20局的下注总额
func (this *RollPointPlayerData) GetBetGig20() int {
	Count := 0
	if len(this.BetBigRecord) > 20 {
		this.BetBigRecord = append(this.BetBigRecord[:0], this.BetBigRecord[1:]...)
	}
	if len(this.BetBigRecord) > 0 {
		for _, v := range this.BetBigRecord {
			Count += v
		}
	}
	return Count
}

func (this *RollPointPlayerData) Clean() {
	this.gainCoin = 0
	for key, _ := range this.CoinPool {
		this.CoinPool[key] = 0
	}

	for key, _ := range this.PushCoinLog {
		for coin, _ := range this.PushCoinLog[key] {
			this.PushCoinLog[key][coin] = 0
		}
	}

	if this.WinRecord == nil {
		this.WinRecord = append(this.WinRecord, 0)
	}

	if this.BetBigRecord == nil {
		this.BetBigRecord = append(this.BetBigRecord, 0)
	}
	this.gameing = false
	this.UnmarkFlag(base.PlayerState_WaitOp)
	this.taxCoin = 0
	this.winCoin = 0
}
func (this *RollPointPlayerData) PushCoin(flag int32, coin int32) {
	this.CoinPool[flag] = this.CoinPool[flag] + coin
	if pool, ok := this.PushCoinLog[flag]; ok {
		if log, ok := pool[coin]; ok {
			pool[coin] = log + 1
		} else {
			pool[coin] = 1
		}
	} else {
		pool = make(map[int32]int32)
		pool[coin] = 1
		this.PushCoinLog[flag] = pool
	}
}
func (this *RollPointPlayerData) GetTotlePushCoin() int32 {
	var coin int32
	for _, value := range this.CoinPool {
		coin += value
	}
	return coin
}
func (this *RollPointPlayerData) MaxChipCheck(index int64, coin int64, maxBetCoin []int32) (bool, int32) {
	if index < 0 || int32(index) > rule.BetArea_Max {
		return false, -1
	}
	maxBetCoinIndex := rule.AreaIndex2MaxChipIndex[index]
	totleCoin := this.CoinPool[int32(index)]
	if maxBetCoin[maxBetCoinIndex] > 0 && totleCoin+int32(coin) > maxBetCoin[maxBetCoinIndex] {
		return false, maxBetCoin[maxBetCoinIndex]
	}
	return true, 0
}
