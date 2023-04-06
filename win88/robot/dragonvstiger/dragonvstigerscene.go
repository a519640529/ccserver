package dragonvstiger

import (
	"fmt"
	"games.yol.com/win88/common"
	. "games.yol.com/win88/gamerule/dragonvstiger"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/dragonvstiger"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
	"math/rand"
	"time"
)

type DragonVsTigerScene struct {
	base.BaseScene
	*dragonvstiger.SCDragonVsTigerRoomInfo
	players         map[int32]*DragonVsTigerPlayer
	totalBet        [DVST_ZONE_MAX]int64
	dbGameFree      *server.DB_GameFree //百人场静态数据
	upBankerListNum int32               //上庄列表数量
	bankerList      map[int32]int64
	isRandNum       map[int32]int32
}

func NewDragonVsTigerScene(info *dragonvstiger.SCDragonVsTigerRoomInfo) *DragonVsTigerScene {
	s := &DragonVsTigerScene{
		SCDragonVsTigerRoomInfo: info,
		players:                 make(map[int32]*DragonVsTigerPlayer),
	}
	s.RobotTypeAIName = make(map[int]string)
	s.Init()

	return s
}

func (s *DragonVsTigerScene) Init() {

	for zone, chips := range s.GetTotalChips() {
		s.totalBet[zone] += chips
	}

	s.RobotTypeAIName[ROBOT_TYPE_DVTRANDOM] = "dvtrandom.json"
	s.RobotTypeAIName[ROBOT_TYPE_DVTFWIN] = "dvtfollowwin.json"
	s.RobotTypeAIName[ROBOT_TYPE_DVTIWIN] = "dvtinvertwin.json"

	for _, v := range s.RobotTypeAIName {
		base.InitTree(v)
	}
	s.isRandNum = make(map[int32]int32)
	s.bankerList = make(map[int32]int64)
	s.upBankerListNum = 0
}

func (s *DragonVsTigerScene) Clear() {
	for _, player := range s.players {
		player.Clear()
	}
	for i := 0; i < DVST_ZONE_MAX; i++ {
		s.totalBet[i] = 0
	}
	s.isRandNum[s.dbGameFree.GetId()] = rand.Int31n(3) + 1
}

func (s *DragonVsTigerScene) RandPlayerType() int {

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
func (s *DragonVsTigerScene) AddPlayer(p base.Player) {
	if mp, ok := p.(*DragonVsTigerPlayer); ok {
		s.players[p.GetSnId()] = mp
	}
}

func (s *DragonVsTigerScene) DelPlayer(snid int32) {
	if p, exist := s.players[snid]; exist && p != nil {
		delete(s.players, snid)
		if _, ok := s.bankerList[snid]; ok {
			delete(s.bankerList, snid)
		}
	}
}

func (s *DragonVsTigerScene) GetPlayerByPos(pos int32) base.Player {
	return nil
}

func (s *DragonVsTigerScene) GetPlayerBySnid(snid int32) base.Player {
	if p, exist := s.players[snid]; exist {
		return p
	}
	return nil
}

func (this *DragonVsTigerScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*player.SCPlayerData); ok {
		return this.GetPlayerBySnid(user.GetData().GetSnId())
	}
	return nil
}

func (s *DragonVsTigerScene) IsFull() bool {
	return false
}

func (s *DragonVsTigerScene) IsMatchScene() bool {
	return false
}

func (s *DragonVsTigerScene) IsCoinScene() bool {
	return false
}

var DragonVsTigerChipWeight = []int64{50, 30, 10, 8, 2}

func (scene *DragonVsTigerScene) Action(s *netlib.Session, player *DragonVsTigerPlayer) {
	if model.GameParamData.UseBevRobot {
		return
	}
	logger.Logger.Trace("(scene *DragonVsTigerScene) Action ", player.GetSnId())

	//pool := []int{DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAW, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAW}
	ok := base.StartSessionGameTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if scene.GetState() != int32(DragonVsTigerSceneStateStake) {
			return true
		}
		if time.Now().Before(player.tNextBet) {
			return true
		}

		var idx int32
		idx = rand.Int31n(2) + 1
		//if player.choose == -1 {
		//	idx = rand.Int31n(2) + 1
		//	player.choose = idx
		//} else {
		//idx = player.choose
		if rand.Int31n(100) <= 5 {
			idx = int32(DVST_ZONE_DRAW)
		}
		//}
		player.choose = idx
		chip := int32(0)
		params := scene.GetParamsEx()
		if len(params) != 0 {
			dbGameFree := base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
			if dbGameFree != nil {
				otherParams := scene.GetOtherIntParams()
				o := common.RandSliceIndexByWight(DragonVsTigerChipWeight)
				chip = otherParams[o]
				if int64(chip) > player.GetCoin()/3 {
					chip = otherParams[0]
				}

				////金币不够
				if player.GetCoin() < int64(chip) {
					n := len(otherParams)
					coin := otherParams[n-1]
					base.ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coin))
					player.tNextBet = time.Now().Add(time.Second * 10)
					player.Coin = proto.Int64(player.GetCoin() + int64(coin))
					return true
				}
			}
		}

		pack := &dragonvstiger.CSDragonVsTiggerOp{
			OpCode: proto.Int(DragonVsTigerPlayerOpBet),
			Params: []int64{int64(idx), int64(chip)},
		}
		proto.SetDefaults(pack)
		s.Send(int(dragonvstiger.DragonVsTigerPacketID_PACKET_CS_DVST_PLAYEROP), pack)
		player.tNextBet = time.Now().Add(time.Duration(rand.Int31n(1000)+500) * time.Millisecond)
		return true
	}), nil, time.Millisecond*200, -1)
	if ok {
		player.tNextBet = time.Now().Add(time.Duration(rand.Int31n(1000)+500) * time.Millisecond)
	}
}

func (this *DragonVsTigerScene) Update(ts int64) {
	if model.GameParamData.UseBevRobot {
		for _, mpd := range this.players {
			mpd.UpdateAction(ts)
		}
	}
	//随机上庄
	if len(this.bankerList) > 0 {
		if this.upBankerListNum < this.isRandNum[this.dbGameFree.GetId()] {
			for snid, ts := range this.bankerList {
				if time.Now().Unix() >= ts {
					player := this.players[snid]
					if player != nil {
						player.Banker()
					}
					delete(this.bankerList, snid)
					break
				}
			}
		}
	}
}
func (this *DragonVsTigerScene) UpBanker(p *DragonVsTigerPlayer) {
	if this.dbGameFree.GetBanker() == 0 {
		return
	}
	if p == nil {
		return
	}
	if p.GetCoin() < int64(this.dbGameFree.GetBanker()) {
		return
	}
	n := this.upBankerListNum
	if n > this.isRandNum[this.dbGameFree.GetId()] {
		return
	}
	bankRate := 100 / (this.dbGameFree.GetSceneType())                   //1,2,3,4->100,50,33,25
	bankRate = int32(float32(bankRate) * float32(float32(bankRate)/100)) ///100,25,10,8
	if rand.Int31n(100) > bankRate {
		return
	}
	if _, ok := this.bankerList[p.GetSnId()]; !ok && len(this.bankerList) < 3 {
		this.bankerList[p.GetSnId()] = time.Now().Add(time.Duration(rand.Intn(20)) * time.Second).Unix()
	}
}
