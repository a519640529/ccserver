package base

import (
	"games.yol.com/win88/common"
	b3core "github.com/magicsea/behavior3go/core"

	"games.yol.com/win88/proto"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	server_proto "games.yol.com/win88/protocol/server"
)

var NilPlayer Player = nil

const (
	PlayerState_Wait     = 1
	PlayerState_PreCoin  = 2
	PlayerState_Gameing  = 3
	PlayerState_Recharge = 4
)

type Player interface {
	Clear()
	GetSnId() int32
	GetPos() int32
	GetFlag() int32
	GetCoin() int64
	SetFlag(flag int32)
	GetLastOp() int32
	SetLastOp(op int32)
	MarkFlag(flag int32)
	UnmarkFlag(flag int32)
	IsMarkFlag(flag int32) bool
	IsOnLine() bool
	IsReady() bool
	IsSceneOwner() bool
	IsRobot() bool
	UpdateCards(cards []int32)
	UpdateBasePlayers(gameId string, basePlayer *server_proto.RobotData)
	GetBasePlayers(gameId string) *server_proto.RobotData
	LeaveGameMsg(snid int32) bool
	GetOutLimitCoin() int
	GetGameCount() int
	GetLastWinOrLoss() int
	GetTakeCoin() int64
}

//百人辅助数据
type HBasePlayers struct {
	redayBetCoin int64
	trueBetCoin  int64
	BetCoinList  []int64
	Trend20      []int32
}

type BasePlayers struct {
	BasePlayer    map[string]*server_proto.RobotData
	GameCount     int
	LastWinOrLoss int //0 平局  1 赢取  -1 失败
	TakeCoin      int64
	BlackData     *b3core.Blackboard
	TreeID        int
	Scene         interface{}
}

func (this *BasePlayers) UpdateBasePlayers(gameId string, basePlayer *server_proto.RobotData) {
	if this.BasePlayer == nil {
		this.BasePlayer = make(map[string]*server_proto.RobotData)
	}
	this.BasePlayer[gameId] = basePlayer
}
func (this *BasePlayers) GetBasePlayers(gameId string) *server_proto.RobotData {
	if this.BasePlayer == nil {
		this.BasePlayer = make(map[string]*server_proto.RobotData)
	}
	if this.BasePlayer[gameId] == nil {
		this.BasePlayer[gameId] = &server_proto.RobotData{}
	}
	return this.BasePlayer[gameId]
}

func (this *BasePlayers) GetLastWinOrLoss() int {
	return this.LastWinOrLoss
}

func (this *BasePlayers) GetGameCount() int {
	return this.GameCount
}

func (this *BasePlayers) GetTakeCoin() int64 {
	return this.TakeCoin
}

func (this *BasePlayers) GetOutLimitCoin() int {
	return -1
}

func (this *BasePlayers) SendMsg(snid int32, packetid int, data interface{}) bool {
	s := PlayerMgrSington.GetPlayerSession(snid)
	if s != nil {
		return s.Send(packetid, data)
	}
	return false
}

func (this *BasePlayers) LeaveGameMsg(snid int32) bool {
	s := PlayerMgrSington.GetPlayerSession(snid)
	if s != nil {
		pack := &hall_proto.CSLeaveRoom{
			Mode: proto.Int32(0),
		}
		proto.SetDefaults(pack)
		return s.Send(int(hall_proto.GameHallPacketID_PACKET_CS_LEAVEROOM), pack)
	}
	return false
}

func CheckBetCoinChip(betCoin []int32, needBetCoin, maxCoin int64) int64 {
	//先生成需要押注的金额，对needBetCoin 进行转化
	if needBetCoin > maxCoin {
		needBetCoin = maxCoin
	}

	for i := 1; i < len(betCoin); i++ {
		chip := int64(betCoin[i-1])
		if needBetCoin < int64(betCoin[i]) {
			needBetCoin -= needBetCoin % chip
			break
		}
	}

	if needBetCoin <= 0 {
		return 0
	}

	if needBetCoin > int64(betCoin[len(betCoin)-1]) {
		if common.RandInt(10000) < 5000 {
			needBetCoin -= needBetCoin % int64(betCoin[len(betCoin)-2])
		} else {
			needBetCoin -= needBetCoin % int64(betCoin[len(betCoin)-1])
		}
	}
	return needBetCoin
}

//生成百人游戏下注的列表
func CreateBetCoinList(betCoin []int32, needBetCoin int64) ([]int64, int64) {
	var ret []int64

	//开始生成结果
	//选择最优下注，从最大开始寻找
	lastValue := needBetCoin
	if needBetCoin == 0 {
		return ret, 0
	}

	for {
		if lastValue <= 0 {
			break
		}
		for i := len(betCoin) - 1; i >= 0; i-- {
			chip := int64(betCoin[i])
			//拆分多次投注，避免一次投注太大
			if lastValue >= chip && (i == 0 || common.RandInt(10000) > 6000) {
				ret = append(ret, chip)
				lastValue -= chip
				break
			}
		}
		if lastValue < int64(betCoin[0]) {
			break
		}
	}

	return ret, needBetCoin
}
