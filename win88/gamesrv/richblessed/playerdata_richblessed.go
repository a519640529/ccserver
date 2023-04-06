package richblessed

import (
	"games.yol.com/win88/gamerule/richblessed"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
)

var (
	ElementsParams        = [][]int32{} // 十三个元素
	FreeElementsParams    = [][]int32{}
	FreeLeElementsParams  = []int32{0, 0, 60, 50, 50, 60, 50, 0, 0, 0, 0, 0, 0}
	JACKPOTElementsParams = []int32{20, 30, 100, 150}
)

// RichBlessed
type RichBlessedPlayerData struct {
	*base.Player
	roomid       int32 //房间ID
	result       *richblessed.WinResult
	betIdx       int               //下注索引
	betCoin      int64             //下注金额
	maxbetCoin   int64             //下注金额
	oneBetCoin   int64             //单注
	cpCtx        model.CoinPoolCtx //水池环境
	leaveTime    int32             //离开时间
	winCoin      int64             //本局输赢
	freewinCoin  int64             //免费游戏总输赢
	JackwinCoin  int64             //Jack奖励
	freeTimes    int32             //免费游戏
	addfreeTimes int32             //新增免费游戏
	gameState    int               //当前游戏模式
	taxCoin      int64             //税收
	noWinTimes   int               //没有中奖次数
	JackpotEle   int32             //中奖元素
	test         bool              // 测试
	///当局数值
	nowFreeTimes     int   //当前第几轮免费游戏
	winLineRate      int64 //线倍率
	JackPotRate      int64 //获取奖池的百分比 5% 10% 15%
	freeWinTotalRate int64 //免费游戏赢的倍率
	// winNowJackPotCoin int64 //当局奖池爆奖
	winNowAllRate int64 //当局赢得倍率
}

func (p *RichBlessedPlayerData) init() {
	p.roomid = 0
	p.test = true
	p.result = new(richblessed.WinResult)
}
func (p *RichBlessedPlayerData) Clear() {
	p.gameState = richblessed.Normal
	p.winCoin = 0
	p.taxCoin = 0
	p.JackpotEle = -1
	p.addfreeTimes = 0
	p.result.AllRate = 0
	p.result.FreeNum = 0
	p.result.JackpotEle = -1
	p.result.JackpotRate = 0
}

// 正常游戏 免费游戏
func (p *RichBlessedPlayerData) CreateResult(eleLineAppearRate [][]int32) {
	p.result.CreateLine(eleLineAppearRate)
}

// 正常游戏 免费游戏
func (p *RichBlessedPlayerData) CreateJACKPOT(eleRate []int32) {
	p.result.CreateJACKPOT(eleRate)

}
