package redvsblack

import (
	"fmt"
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/redvsblack"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/redvsblack"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
	"math/rand"
	"time"
)

const (
	RVSB_ZONE_BLACK int = iota
	RVSB_ZONE_RED
	RVSB_ZONE_LUCKY
	RVSB_ZONE_MAX
)

const (
	ROBOT_TYPE_RVBRANDOM int = iota
	ROBOT_TYPE_RVBFWIN
	ROBOT_TYPE_RVBIWIN
)

type RedVsBlackScene struct {
	base.BaseScene
	*redvsblack.SCRedVsBlackRoomInfo
	players  map[int32]*RedVsBlackPlayer
	totalBet [RVSB_ZONE_MAX]int64
}

func NewRedVsBlackScene(info *redvsblack.SCRedVsBlackRoomInfo) *RedVsBlackScene {
	s := &RedVsBlackScene{
		SCRedVsBlackRoomInfo: info,
		players:              make(map[int32]*RedVsBlackPlayer),
	}
	s.RobotTypeAIName = make(map[int]string)
	s.Init()
	return s
}

func (s *RedVsBlackScene) Init() {

	for zone, chips := range s.GetTotalChips() {
		s.totalBet[zone] += int64(chips)
	}

	s.RobotTypeAIName[ROBOT_TYPE_RVBRANDOM] = "rvbrandom.json"
	s.RobotTypeAIName[ROBOT_TYPE_RVBFWIN] = "rvbfollowwin.json"
	s.RobotTypeAIName[ROBOT_TYPE_RVBIWIN] = "rvbinvertwin.json"

	for _, v := range s.RobotTypeAIName {
		base.InitTree(v)
	}

}

func (s *RedVsBlackScene) Clear() {
	for _, player := range s.players {
		player.Clear()
	}
	for i := 0; i < RVSB_ZONE_MAX; i++ {
		s.totalBet[i] = 0
	}

}

func (s *RedVsBlackScene) RandPlayerType() int {
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

func (s *RedVsBlackScene) AddPlayer(p base.Player) {
	if mp, ok := p.(*RedVsBlackPlayer); ok {
		s.players[p.GetSnId()] = mp
	}
}

func (s *RedVsBlackScene) DelPlayer(snid int32) {
	if p, exist := s.players[snid]; exist && p != nil {
		delete(s.players, snid)
	}
}

func (s *RedVsBlackScene) GetPlayerByPos(pos int32) base.Player {
	return nil
}

func (s *RedVsBlackScene) GetPlayerBySnid(snid int32) base.Player {
	if p, exist := s.players[snid]; exist {
		return p
	}
	return nil
}

func (this *RedVsBlackScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*player.SCPlayerData); ok {
		return this.GetPlayerBySnid(user.GetData().GetSnId())
	}
	return nil
}

func (s *RedVsBlackScene) IsFull() bool {
	return false
}

func (s *RedVsBlackScene) IsMatchScene() bool {
	return false
}

func (s *RedVsBlackScene) IsCoinScene() bool {
	return false
}

var RedVsBlackChipWeight = []int64{50, 30, 10, 8, 2}

func (scene *RedVsBlackScene) Action(s *netlib.Session, player *RedVsBlackPlayer) {
	if model.GameParamData.UseBevRobot {
		return
	}
	//logger.Logger.Info("(scene *RedVsBlackScene) Action ", player.GetSnId())
	if base.StartSessionGameTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if scene.GetState() != int32(rule.RedVsBlackSceneStateStake) {
			return true
		}
		if time.Now().Before(player.tNextBet) {
			return true
		}

		var idx int32
		if player.choose == -1 {
			idx = rand.Int31n(2)
			player.choose = idx
		} else {
			idx = player.choose
			if rand.Int31n(100) < 10 {
				idx = int32(RVSB_ZONE_LUCKY)
			}
		}
		chip := int32(0)
		params := scene.GetParamsEx()
		if len(params) != 0 {
			dbGameFree := base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
			if dbGameFree != nil {
				otherParams := dbGameFree.GetOtherIntParams()
				o := common.RandSliceIndexByWight(RedVsBlackChipWeight)
				chip = otherParams[o]
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

		pack := &redvsblack.CSRedVsBlackOp{
			OpCode: proto.Int(rule.RedVsBlackPlayerOpBet),
			Params: []int64{int64(idx), int64(chip)},
		}
		proto.SetDefaults(pack)
		s.Send(int(redvsblack.RedVsBlackPacketID_PACKET_CS_RVSB_PLAYEROP), pack)
		player.tNextBet = time.Now().Add(time.Duration(rand.Int31n(1000)+500) * time.Millisecond)
		return true
	}), nil, time.Millisecond*200, -1) {
		player.tNextBet = time.Now().Add(time.Duration(rand.Int31n(1000)+500) * time.Millisecond)
	}

}

func (s *RedVsBlackScene) Update(ts int64) {
	if model.GameParamData.UseBevRobot {
		for _, mpd := range s.players {
			mpd.UpdateAction(ts)
		}
	}
}
