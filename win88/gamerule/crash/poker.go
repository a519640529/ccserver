package crash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
	"math"
	"strconv"
	"time"
)

const (
	POKER_CART_CNT int = 10000
)

var cardSeed = time.Now().UnixNano()

type Card struct {
	Hashstr string
	Explode int32
}

type Poker struct {
	buf      [POKER_CART_CNT]*Card
	pos      int
	gameHash string
	wheel    int
}

//Sha256加密
func Sha256(src string) string {
	m := sha256.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}

func HashToMultiple(oldgameHash string, wheel int) int32 {
	//计算当前传来的gameHash
	jsHash := Sha256(fmt.Sprintf("%v%v", oldgameHash, model.GameParamData.AtomGameHash[wheel]))

	//哈希计算
	s13 := jsHash[0:13]
	h, _ := strconv.ParseInt(s13, 16, 0)
	e := math.Pow(2, 52)
	result := math.Floor((96 * e) / (e - float64(h)))
	if result < 101 {
		result = 0
	}
	if result > 10000 {
		result = 10000
	}
	return int32(result)
}

func NewPoker(period, wheel int) *Poker {
	if len(model.GameParamData.InitGameHash)-1 < wheel{
		wheel = 0
	}
	gameHash := model.GameParamData.InitGameHash[wheel] //"ff6c5b1daa1068897377f7a64a762eafda4d225f25bf8e3bb476a7c4f2d10468"
	p := &Poker{}
	p.init(gameHash, period, wheel)
	return p
}

//const salt = `0ead8d98e67a7c9197a6bb0e664bb84adbeb25e4e0db63d2158e48b98a50534d`

func (this *Poker) init(gameHash string, period, wheel int) {
	if this.wheel != wheel {
		this.wheel = wheel
	}

	for i := POKER_CART_CNT - 1; i >= 0; i-- {
		//logger.Logger.Info("gameHash:",gameHash)
		oldgameHash := gameHash

		//生成下一个gmaeHash
		if oldgameHash != "" {
			gameHash = Sha256(fmt.Sprintf("%v", oldgameHash))
		} else {
			gameHash = Sha256(fmt.Sprintf("%v%v", gameHash, model.GameParamData.AtomGameHash[wheel]))
		}
		//logger.Logger.Info("newgameHash:",gameHash)
		this.gameHash = gameHash

		result := HashToMultiple(oldgameHash, wheel)

		//当前哈希对
		this.buf[i] = &Card{
			Hashstr: oldgameHash,
			Explode: int32(result),
		}
		logger.Logger.Infof("curgameHash:%v %v nextgameHash:%v", oldgameHash, result, gameHash)
	}
	//this.Shuffle()
	this.pos = period
}

func (this *Poker) Next() (*Card, int, int) {
	if this.pos >= len(this.buf) {
		if len(model.GameParamData.InitGameHash) > this.wheel+1 {
			this.wheel++
			this.gameHash = model.GameParamData.AtomGameHash[this.wheel]
		}
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.UptIntKVGameData("CrashWheel", int64(this.wheel))
		}), nil, "UptCrashWheelKVGameData").Start()
		//gameHash := Sha256(fmt.Sprintf("%v", time.Now().UnixNano()))
		this.init(this.gameHash, 0, this.wheel)
		this.pos = 0
		c := this.buf[this.pos]
		this.pos++
		return c, this.pos, this.wheel
	}
	c := this.buf[this.pos]
	this.pos++
	return c, this.pos, this.wheel
}
