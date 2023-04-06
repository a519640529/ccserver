package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

const (
	LOTTERY_LOG_MAX = 30
)

var (
	LotteryDBName   = "user"
	LotteryCollName = "user_lottery"
)

// 奖池记录
type GameLotteryLog struct {
	Time int32   //时间
	Nick string  //昵称
	Card []int32 //牌数据
	Kind int32   //牌型
	Coin int32   //获得奖金
}

type GameLottery struct {
	Id     int32             //游戏场次id
	GameId int32             //游戏id
	Value  int64             //彩金数量
	Logs   []*GameLotteryLog //彩金获得记录
}

type Lottery struct {
	Id         bson.ObjectId          `bson:"_id"`
	Dirty      int32                  `bson:"-"`
	Platform   string                 //平台编号
	Lotteries  map[int32]*GameLottery //奖金池
	CreateTime time.Time              //创建日期
	UpdateTime time.Time              //最后更新日期
}

func NewLottery(platform string) *Lottery {
	cl := &Lottery{
		Id:         bson.NewObjectId(),
		Dirty:      1,
		Platform:   platform,
		CreateTime: time.Now(),
		Lotteries:  make(map[int32]*GameLottery),
	}
	return cl
}

func GetAllLottery() (ret []Lottery, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	err = rpcCli.CallWithTimeout("LotterySvc.GetAllLottery", struct{}{}, &ret, time.Second*30)
	return
}

func UpsertLottery(item *Lottery) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	var ret bool
	err = rpcCli.CallWithTimeout("LotterySvc.UpsertLottery", item, &ret, time.Second*30)
	return
}
