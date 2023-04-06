package crash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/crash"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/crash"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
	"math"
	"math/rand"
	"strconv"
	"time"
)

const (
	CRASH_ZONE_BLACK int = iota
	CRASH_ZONE_RED
	CRASH_ZONE_LUCKY
	CRASH_ZONE_MAX
)

const (
	ROBOT_TYPE_RVBRANDOM int = iota
	ROBOT_TYPE_RVBFWIN
	ROBOT_TYPE_RVBIWIN
)

type CrashScene struct {
	base.BaseScene
	*crash.SCCrashRoomInfo
	players map[int32]*CrashPlayer
	//totalBet [CRASH_ZONE_MAX]int64
}

func NewCrashScene(info *crash.SCCrashRoomInfo) *CrashScene {
	s := &CrashScene{
		SCCrashRoomInfo: info,
		players:         make(map[int32]*CrashPlayer),
	}
	s.RobotTypeAIName = make(map[int]string)
	s.Init()
	return s
}

func (s *CrashScene) Init() {

	//for zone, chips := range s.GetTotalChips() {
	//	s.totalBet[zone] += int64(chips)
	//}

	//s.RobotTypeAIName[ROBOT_TYPE_RVBRANDOM] = "crashrandom.json"
	//s.RobotTypeAIName[ROBOT_TYPE_RVBFWIN] = "crashfollowwin.json"
	//s.RobotTypeAIName[ROBOT_TYPE_RVBIWIN] = "crashinvertwin.json"
	//
	//for _, v := range s.RobotTypeAIName {
	//	base.InitTree(v)
	//}

}

func (s *CrashScene) Clear() {
	for _, player := range s.players {
		player.Clear()
	}
	//for i := 0; i < CRASH_ZONE_MAX; i++ {
	//	s.totalBet[i] = 0
	//}

}

func (s *CrashScene) RandPlayerType() int {
	tmpRate := []int{70, 10, 20}
	//获得场景所有机器人的比例，按照比例分派
	tmpNum := []int{}
	for i := 0; i < len(tmpRate); i++ {
		tmpNum = append(tmpNum, 0)
	}

	for _, v := range s.players {
		if v.IsRobot() {
			tmpNum[v.TreeID] += 1
		}
	}

	minNum := float64(99999999)
	minIndex := 0
	for i := 0; i < len(tmpRate); i++ {
		val := float64(tmpNum[i]) / float64(tmpRate[i])
		if val < minNum {
			minNum = val
			minIndex = i
		}
	}

	return minIndex
}

func (s *CrashScene) AddPlayer(p base.Player) {
	if mp, ok := p.(*CrashPlayer); ok {
		s.players[p.GetSnId()] = mp
	}
}

func (s *CrashScene) DelPlayer(snid int32) {
	if p, exist := s.players[snid]; exist && p != nil {
		delete(s.players, snid)
	}
}

func (s *CrashScene) GetPlayerByPos(pos int32) base.Player {
	return nil
}

func (s *CrashScene) GetPlayerBySnid(snid int32) base.Player {
	if p, exist := s.players[snid]; exist {
		return p
	}
	return nil
}

func (this *CrashScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*player.SCPlayerData); ok {
		return this.GetPlayerBySnid(user.GetData().GetSnId())
	}
	return nil
}

func (s *CrashScene) IsFull() bool {
	return false
}

func (s *CrashScene) IsMatchScene() bool {
	return false
}

func (s *CrashScene) IsCoinScene() bool {
	return false
}

var CrashChipWeight = []int64{96, 1, 1, 1, 1}

func (scene *CrashScene) Action(s *netlib.Session, player *CrashPlayer) {
	//if model.GameParamData.UseBevRobot {
	//	return
	//}
	//logger.Logger.Info("(scene *CrashScene) Action ", player.GetSnId())
	if base.StartSessionGameTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if scene.GetState() != int32(rule.CrashSceneStateStake) {
			return true
		}
		if player.down {
			return true
		}
		if time.Now().Before(player.tNextBet) {
			return true
		}

		player.choose = scene.GetMultiple(player.GetSnId())
		if player.choose <= rule.MinMultiple {
			player.choose = rule.MinMultiple + int32(rand.Intn(10))
		}

		chip := int32(0)
		params := scene.GetParamsEx()
		if len(params) != 0 {
			dbGameFree := base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
			if dbGameFree != nil {
				otherParams := dbGameFree.GetOtherIntParams()
				o := common.RandSliceIndexByWight(CrashChipWeight)
				chip = rand.Int31n(otherParams[o])
				if chip <= 0 {
					chip = 100
				}
				//金币不够
				if player.GetCoin() < int64(chip) {
					coin := otherParams[len(otherParams)-1]
					base.ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coin))
					player.tNextBet = time.Now().Add(time.Second * 10)
					player.Coin = proto.Int64(player.GetCoin() + int64(coin))
					return true
				}
			}
		}

		pack := &crash.CSCrashOp{
			OpCode: proto.Int(rule.CrashPlayerOpBet),
			Params: []int64{int64(player.choose), int64(chip)},
		}
		player.multiple = player.choose
		player.betTotal = int64(chip)
		proto.SetDefaults(pack)
		logger.Logger.Infof("%v 玩家下注：%v", player.SnId, pack)
		s.Send(int(crash.CrashPacketID_PACKET_CS_CRASH_PLAYEROP), pack)
		player.down = true
		player.tNextBet = time.Now().Add(time.Duration(rand.Int31n(1000)+500) * time.Millisecond)
		return true
	}), nil, time.Millisecond*200, -1) {
		player.tNextBet = time.Now().Add(time.Duration(rand.Int31n(1000)+500) * time.Millisecond)
	}

}

func (s *CrashScene) GetMultiple(snid int32) int32 {
	//随机生成哈希
	gameHash := Sha256(fmt.Sprintf("%v%v%v", snid, time.Now().UnixNano(), rand.Intn(100)))
	//随机哈希加原子哈希
	sha256str := Sha256(fmt.Sprintf("%v", gameHash))
	//取前13位
	s13 := sha256str[0:13]
	//转16进制到10进制
	h, _ := strconv.ParseInt(s13, 16, 0)
	//2的52次方
	e := math.Pow(2, 52)
	//运算方法
	result := math.Floor((96 * e) / (e - float64(h)))
	if result < 101 {
		result = 0
	}
	if result > 5000 {
		result = 5000
	}
	return int32(result)
}

//Sha256加密
func Sha256(src string) string {
	m := sha256.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}

func (s *CrashScene) Update(ts int64) {
	if model.GameParamData.UseBevRobot {
		for _, mpd := range s.players {
			mpd.UpdateAction(ts)
		}
	}
}
