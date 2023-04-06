package main

import (
	"math/rand"
	"strconv"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/gamehall"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
)

var jackpotInterval = time.Hour * common.SCENE_BIGWINHISTORY_TIMEINTERVAL

var JackpotListMgrSington = &JackpotListMgr{
	BigWinHistoryByGameID: make(map[int][]*gamehall.BigWinHistoryInfo),
	jackpotListHandle:     make(map[int]timer.TimerHandle), // 新的爆奖记录生成
}

type JackpotListMgr struct {
	BigWinHistoryByGameID map[int][]*gamehall.BigWinHistoryInfo
	jackpotListHandle     map[int]timer.TimerHandle // 新的爆奖记录生成
}

func (this *JackpotListMgr) AddJackpotList(gameid int, data *gamehall.BigWinHistoryInfo) {
	this.BigWinHistoryByGameID[gameid] = append(this.BigWinHistoryByGameID[gameid], data)
	if len(this.BigWinHistoryByGameID[gameid]) > common.SCENE_BIGWINHISTORY_MAXNUMBER {
		this.BigWinHistoryByGameID[gameid] = this.BigWinHistoryByGameID[gameid][1:]
	}
}

func (this *JackpotListMgr) GetJackpotList(gameid int) []*gamehall.BigWinHistoryInfo {
	if this.BigWinHistoryByGameID[gameid] == nil {
		this.BigWinHistoryByGameID[gameid] = make([]*gamehall.BigWinHistoryInfo, 0)
	}
	return this.BigWinHistoryByGameID[gameid]
}

func genRandTime(sec int, circleTime time.Time) time.Time {
	//随机时间间隔
	rand.Seed(time.Now().UnixNano() + int64(sec))
	interval := rand.Intn(60) + 60*sec // 分钟
	circleTime = circleTime.Add(time.Duration(-interval) * time.Minute)
	s := rand.Intn(60) //随机一个秒数
	circleTime = circleTime.Add(time.Duration(-s) * time.Second)
	return circleTime
}

func genRoomIDAndScore(gameid int) (roomID int64, score int64) {
	// 随机从房间内读取一个场数据
	var scenes = make([]*Scene, 0)
	for _, s := range SceneMgrSington.scenes {
		if s != nil && s.dbGameFree.GetGameId() == int32(gameid) {
			scenes = append(scenes, s)
		}
	}
	if len(scenes) < 1 {
		return
	}
	s := scenes[rand.Intn(len(scenes))]
	jackpot := s.dbGameFree.GetJackpot()
	roomID = int64(s.dbGameFree.GetBaseScore())
	baseScore := int64(jackpot[0]) * roomID
	score = int64(baseScore) + int64(rand.Int31n(int32(baseScore/2)))
	logger.Logger.Infof("genjackpot %v score %v roomID%v baseScore%v", jackpot[0], score, s.dbGameFree.GetBaseScore(), baseScore)
	return
}

// 生成爆奖记录
func (this *JackpotListMgr) GenJackpot(gameid int) {
	// 首次生成初始化爆奖信息
	if len(this.BigWinHistoryByGameID[gameid]) == 0 {
		// 直接从大厅取机器人
		circleTime := time.Now()
		sec := common.SCENE_BIGWINHISTORY_LIMITNUMBER
		for _, p := range PlayerMgrSington.playerMap {
			if len(this.BigWinHistoryByGameID[gameid]) >= common.SCENE_BIGWINHISTORY_LIMITNUMBER {
				break
			}
			if p.IsRob {
				p.RobotRandName()
				genedTime := genRandTime(sec, circleTime).Unix()
				spinid := strconv.FormatInt(int64(p.SnId), 10) // 用户id转换成字符串
				baseBet, priceValue := genRoomIDAndScore(gameid)
				if baseBet == 0 || priceValue == 0 {
					return
				}
				newJackpot := &gamehall.BigWinHistoryInfo{
					SpinID:      spinid,
					CreatedTime: genedTime,
					BaseBet:     baseBet,
					TotalBet:    baseBet,
					PriceValue:  priceValue,
					UserName:    p.Name,
				}
				this.AddJackpotList(gameid, newJackpot)
				sec--
			}
		}
	} else {
		lastRecord := this.BigWinHistoryByGameID[gameid][len(this.BigWinHistoryByGameID[gameid])-1] // 当中奖纪录>10条时，随机时间差, 满足当前时间-最后一次爆奖记录时间 > 随机时间差 时重新生成一条记录
		lastTime := time.Unix(lastRecord.GetCreatedTime(), 0)
		genNewJackpotFlag := lastTime.Add(jackpotInterval).Before(time.Now())
		if genNewJackpotFlag {
			for _, p := range PlayerMgrSington.playerMap {
				if p.IsRob {
					p.RobotRandName()
					genedTime := time.Now().Unix()
					spinid := strconv.FormatInt(int64(p.SnId), 10) // 用户id转换成字符串
					baseBet, priceValue := genRoomIDAndScore(gameid)
					if baseBet == 0 || priceValue == 0 {
						return
					}
					newJackpot := &gamehall.BigWinHistoryInfo{
						SpinID:      spinid,
						CreatedTime: genedTime,
						BaseBet:     baseBet,
						TotalBet:    baseBet,
						PriceValue:  priceValue,
						UserName:    p.Name,
					}
					this.AddJackpotList(gameid, newJackpot)
					break
				}
			}
			this.after(gameid)
		}
	}
}

// AddVirtualJackpot 添加虚拟爆奖记录（名字+用户id, 使用大厅机器人信息）
func (this *JackpotListMgr) AddVirtualJackpot(gameid int, data *gamehall.BigWinHistoryInfo) {
	if len(PlayerMgrSington.playerMap) < 1 {
		logger.Logger.Error("AddVirtualJackpot not found robot")
		return
	}

	for _, p := range PlayerMgrSington.playerMap {
		if p.IsRob {
			p.RobotRandName()
			spinid := strconv.FormatInt(int64(p.SnId), 10) // 用户id转换成字符串
			data.SpinID = spinid
			data.UserName = p.Name
			this.AddJackpotList(gameid, data)
			break
		}
	}
}

func (this *JackpotListMgr) start(gameid int) {
	this.jackpotListHandle[gameid], _ = timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		this.GenJackpot(gameid)
		return true
	}), nil, jackpotInterval, 1)
}

func (this *JackpotListMgr) after(gameid int) {
	interval := jackpotInterval + time.Duration(rand.Intn(2))*time.Hour + time.Duration(rand.Intn(60))*time.Minute + time.Duration(rand.Intn(60))
	this.jackpotListHandle[gameid], _ = timer.AfterTimer(func(h timer.TimerHandle, ud interface{}) bool {
		this.GenJackpot(gameid)

		jackpotList := JackpotListMgrSington.GetJackpotList(gameid)
		msg := this.GetStoCMsg(jackpotList)
		logger.Logger.Infof("jackpotlist timer after gameid(%v) %v", gameid, msg)
		return true
	}, nil, interval)
}

func (this *JackpotListMgr) StopTimer(gameid int) bool {
	return timer.StopTimer(this.jackpotListHandle[gameid])
}

func (this *JackpotListMgr) ResetAfterTimer(gameid int) {
	if this.StopTimer(gameid) {
		this.after(gameid)
	}
}

func (this *JackpotListMgr) GetStoCMsg(jackpotList []*gamehall.BigWinHistoryInfo) *gamehall.SCBigWinHistory {
	pack := &gamehall.SCBigWinHistory{}
	for i := len(jackpotList) - 1; i >= 0; i-- {
		v := jackpotList[i]
		player := &gamehall.BigWinHistoryInfo{
			SpinID:      proto.String(v.GetSpinID()),
			CreatedTime: proto.Int64(v.GetCreatedTime()),
			BaseBet:     proto.Int64(v.GetBaseBet()),
			TotalBet:    proto.Int64(v.GetTotalBet()),
			PriceValue:  proto.Int64(int64(v.GetPriceValue())),
			UserName:    proto.String(v.GetUserName()),
			Cards:       v.GetCards(),
		}
		pack.BigWinHistory = append(pack.BigWinHistory, player)
	}

	//pack := &avengers.SCAvengersBigWinHistory{}
	//for i := len(jackpotList) - 1; i >= 0; i-- {
	//	v := jackpotList[i]
	//	player := &avengers.AvengersBigWinHistoryInfo{
	//		SpinID:      proto.String(v.GetSpinID()),
	//		CreatedTime: proto.Int64(v.GetCreatedTime()),
	//		RoomID:      proto.Int64(v.GetRoomID()),
	//		PriceValue:  proto.Int64(int64(v.GetPriceValue())),
	//		UserName:    proto.String(v.GetUserName()),
	//	}
	//	pack.BigWinHistory = append(pack.BigWinHistory, player)
	//}
	proto.SetDefaults(pack)
	return pack
}
