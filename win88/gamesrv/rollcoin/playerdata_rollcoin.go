package rollcoin

import (
	"bytes"
	"encoding/gob"
	"games.yol.com/win88/gamesrv/base"
	"github.com/idealeak/goserver/core/logger"
)

type RollCoinPlayerData struct {
	*base.Player
	gainCoin     int32                     //本局赢取的金币
	CoinPool     map[int32]int32           //押注的金币
	betTotal     int64                     //当局总下注筹码
	PushCoinLog  map[int32]map[int32]int32 //押注的记录,位置:押注:数量
	BankerTimes  int32                     //坐庄次数
	gameing      bool                      //是否参与游戏
	WinRecord    []int                     //最近20局输赢记录
	BetBigRecord []int                     //最近20局下注总额记录
	cGetWin20    int                       //返回玩家最近20局的获胜次数
	cGetBetGig20 int                       //返回玩家最近20局的下注总额
	taxCoin      int64                     //本局税收
	winCoin      int64                     //本局收税前赢的钱
}

//绑定ai
func (this *RollCoinPlayerData) init() {
	if this.IsLocal && this.IsRob {
		this.AttachAI(&RollCoinPlayerAI{})
	}
}

//返回玩家最近20局的获胜次数
func (this *RollCoinPlayerData) GetWin20() int {
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
func (this *RollCoinPlayerData) GetBetGig20() int {
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

// 押注是记录金币变动，不产生日志
//func (this *RollCoinPlayerData) BetCoinChange(num int64, gainWay int32, notifyC bool, oper, remark string) {
//	if num == 0 {
//		return
//	}
//	if notifyC {
//		pack := &player.SCPlayerCoinChange{
//			AddCoin:  proto.Int64(num),
//			RestCoin: proto.Int64(this.Coin - this.betTotal),
//		}
//		if this.GetScene().IsCoinScene() {
//			pack.SnId = proto.Int(this.GetPos())
//		} else if this.GetScene().IsHundredScene() {
//			pack.SnId = proto.Int32(this.SnId)
//		}
//		proto.SetDefaults(pack)
//		if this.broadcastFlag {
//			this.Broadcast(int(protocol.MmoPacketID_PACKET_SC_PLAYERCOINCHANGE), pack, 0)
//		} else {
//			this.SendToClient(int(protocol.MmoPacketID_PACKET_SC_PLAYERCOINCHANGE), pack)
//		}
//		logger.Logger.Trace("(this *Player) AddCoin SCPlayerCoinChange:", pack)
//	}
//
//}

func NewRollCoinPlayerData(p *base.Player) *RollCoinPlayerData {
	return &RollCoinPlayerData{
		Player:      p,
		CoinPool:    make(map[int32]int32),
		PushCoinLog: make(map[int32]map[int32]int32),
	}
}

type RollCoinPlayerMemData struct {
	GainCoin int32 //本局赢取的金币
}

//序列化
func (this *RollCoinPlayerData) Marshal() ([]byte, error) {
	md := RollCoinPlayerMemData{
		GainCoin: this.gainCoin,
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&md)
	if err != nil {
		logger.Logger.Warnf("(this *RollCoinPlayerData) Marshal() %v gob.Encode err:%v", this.SnId, err)
		return nil, err
	}
	return buf.Bytes(), nil
}

//反序列化
func (this *RollCoinPlayerData) Unmarshal(data []byte, ud interface{}) error {
	md := &RollCoinPlayerMemData{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(md)
	if err != nil {
		logger.Logger.Warnf("(this *RollCoinPlayerData) Unmarshal gob.Decode err:%v", err)
		return err
	}
	this.gainCoin = md.GainCoin
	return nil
}
func (this *RollCoinPlayerData) Clean() {
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
func (this *RollCoinPlayerData) PushCoin(flag int32, coin int32) {
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
func (this *RollCoinPlayerData) GetTotlePushCoin() int32 {
	var coin int32
	for _, value := range this.CoinPool {
		coin += value
	}
	return coin
}
func (this *RollCoinPlayerData) MaxChipCheck(index int64, coin int64, maxBetCoin []int32) (bool, int32) {
	if index < 0 || int(index) > len(maxBetCoin) {
		return false, -1
	}
	if maxBetCoin[index] > 0 && this.CoinPool[int32(index)]+int32(coin) > maxBetCoin[index] {
		return false, maxBetCoin[index]
	}
	return true, 0
}
