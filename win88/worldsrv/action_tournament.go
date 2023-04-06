package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/tournament"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
	"sort"
	"strconv"
	"time"
)

// 比赛场信息
type CSTMInfoPacketFactory struct {
}
type CSTMInfoHandler struct {
}

func (this *CSTMInfoPacketFactory) CreatePacket() interface{} {
	pack := &tournament.CSTMInfo{}
	return pack
}

func (this *CSTMInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSTMInfoHandler Process recv ", data)
	if _, ok := data.(*tournament.CSTMInfo); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warnf("CSTMInfo p == nil.")
			return nil
		}
		pack := TournamentMgr.GetSCTMInfosPack(p.Platform)
		proto.SetDefaults(pack)
		logger.Logger.Trace("SCTMInfos++++++++++++:", pack)
		p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCTMInfos), pack)
	}
	return nil
}

// 排行榜
type CSTMRankListPacketFactory struct {
}
type CSTMRankListHandler struct {
}

func (this *CSTMRankListPacketFactory) CreatePacket() interface{} {
	pack := &tournament.CSTMRankList{}
	return pack
}

func (this *CSTMRankListHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSTMRankListHandler Process recv ", data)
	if msg, ok := data.(*tournament.CSTMRankList); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warnf("CSTMRankList p == nil.")
			return nil
		}
		pack := &tournament.SCTMRankList{
			TMId:      msg.TMId,
			TimeRange: "9.1-9.30",
			TMRank: []*tournament.TMRank{
				{RankId: 1, RankName: "rankNo.1", WinnerNum: 5},
				{RankId: 2, RankName: "rankNo.2", WinnerNum: 4},
				{RankId: 3, RankName: "rankNo.3", WinnerNum: 2}},
		}
		proto.SetDefaults(pack)
		logger.Logger.Trace("CSTMRankList:", pack)
		p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCTMRankList), pack)
	}
	return nil
}

// 报名
type CSSignRacePacketFactory struct {
}
type CSSignRaceHandler struct {
}

func (this *CSSignRacePacketFactory) CreatePacket() interface{} {
	pack := &tournament.CSSignRace{}
	return pack
}

func (this *CSSignRaceHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSSignRaceHandler Process recv ", data)
	if msg, ok := data.(*tournament.CSSignRace); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warnf("CSSignRace p == nil.")
			return nil
		}
		if p.scene != nil {
			logger.Logger.Warnf("CSSignRace p.scene != nil.")
			return nil
		}
		pack := &tournament.SCSignRace{}
		if msg.GetOpCode() == 0 {
			//报名
			platform := p.Platform
			if p.IsRob {
				platform = "1"
			}
			info := TournamentMgr.GetMatchInfo(platform, msg.TMId)
			if info != nil && TournamentMgr.MatchSwitch(platform, msg.TMId) && TournamentMgr.IsTimeRange(info) {
				isWaiting, _ := TournamentMgr.IsMatchWaiting(p.SnId)
				if TournamentMgr.IsMatching(p.SnId) || isWaiting {
					pack.RetCode = 1 //重复报名
					logger.Logger.Infof("player(%v) IsMatching.", p.SnId)
				} else {
					ok, code := TournamentMgr.SignUp(msg.TMId, p)
					if !ok {
						logger.Logger.Infof("player(%v) match(%v) SignUp is fail.", p.SnId, msg.TMId)
						pack.RetCode = code //0成功 1重复报名 2比赛没有开启 3道具不足 4不在报名时间段 5金币不足 6钻石不足
					}
				}
			} else {
				logger.Logger.Infof("match(%v) is not open.", msg.TMId)
				pack.RetCode = 4 //4不在报名时间段
				if !TournamentMgr.MatchSwitch(platform, msg.TMId) {
					pack.RetCode = 2 //比赛没有开启
				}
				TournamentMgr.CancelSignUpAll(msg.TMId)
			}
		} else {
			if TournamentMgr.IsMatching(p.SnId) {
				logger.Logger.Infof("player(%v) IsMatching.", p.SnId)
			} else {
				//取消报名
				TournamentMgr.CancelSignUp(msg.TMId, p.SnId, true)
			}
			return nil
		}
		proto.SetDefaults(pack)
		logger.Logger.Trace("SCSignRace:", pack)
		signSucc := p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCSignRace), pack)
		if msg.GetOpCode() == 0 && pack.RetCode == 0 && signSucc {
			TournamentMgr.CheckStart(msg.TMId)
		}
	}
	return nil
}

// 赛季信息
type CSTMSeasonInfoPacketFactory struct {
}
type CSTMSeasonInfoHandler struct {
}

func (this *CSTMSeasonInfoPacketFactory) CreatePacket() interface{} {
	pack := &tournament.CSTMSeasonInfo{}
	return pack
}

func (this *CSTMSeasonInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSTMSeasonInfoHandler Process recv ", data)
	if _, ok := data.(*tournament.CSTMSeasonInfo); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warnf("CSTMSeasonInfoHandler p == nil.")
			return nil
		}
		if p.Platform == Default_Platform {
			logger.Logger.Warnf("CSTMSeasonInfoHandler Platform == Default_Platform.")
			return nil
		}
		msid := MatchSeasonMgrSington.GetMatchSeasonId(p.Platform)
		if msid == nil {
			logger.Logger.Warnf("CSTMSeasonInfoHandler msid == nil.")
			return nil
		}
		send := func(ms *MatchSeason) {
			pack := &tournament.SCTMSeasonInfo{
				Id:              msid.SeasonId,
				SeasonTimeStamp: []int64{msid.StartStamp, msid.EndStamp},
				Lv:              ms.Lv,
				LastLv:          ms.LastLv,
				IsAward:         ms.IsAward,
			}
			proto.SetDefaults(pack)
			logger.Logger.Trace("CSTMSeasonInfoHandler:", pack)
			p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCTMSeasonInfo), pack)
		}
		snid := p.SnId
		ms := MatchSeasonMgrSington.GetMatchSeason(snid)
		if ms == nil {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				ret, err := model.QueryMatchSeasonBySnid(p.Platform, snid)
				if err != nil {
					return nil
				}
				return ret
			}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
				var ret *model.MatchSeason
				dirty := false
				if data == nil || data.(*model.MatchSeason) == nil { //新数据
					player := PlayerMgrSington.GetPlayerBySnId(snid)
					if player != nil {
						ret = model.NewMatchSeason(player.Platform, snid, player.Name, msid.SeasonId, 1)
						dirty = true
					} else {
						logger.Logger.Trace("CSTMSeasonInfoHandler error: player==nil ", snid)
					}
				} else {
					ret = data.(*model.MatchSeason)
					if ret.SeasonId < msid.SeasonId { //不同赛季段位继承
						num := msid.SeasonId - ret.SeasonId
						finalLv := ret.Lv
						for i := 0; i < int(num); i++ { //继承几次
							if i == int(num)-1 { //上个赛季
								ret.LastLv = finalLv
							}
							finalLv = MatchSeasonMgrSington.MatchSeasonInherit(finalLv)
						}
						ret.Lv = finalLv
						ret.SeasonId = msid.SeasonId
						ret.IsAward = false
						ret.UpdateTs = time.Now().Unix()
						dirty = true
					}
				}
				ms = MatchSeasonMgrSington.exchangeModel2Cache(ret)
				ms.dirty = dirty
				MatchSeasonMgrSington.SetMatchSeason(ms)
				send(ms)
			})).StartByFixExecutor("SnId:" + strconv.Itoa(int(snid)))
		} else {
			if ms.SeasonId < msid.SeasonId { //不同赛季段位继承
				num := msid.SeasonId - ms.SeasonId
				finalLv := ms.Lv
				for i := 0; i < int(num); i++ { //继承几次
					if i == int(num)-1 { //上个赛季
						ms.LastLv = finalLv
					}
					finalLv = MatchSeasonMgrSington.MatchSeasonInherit(finalLv)
				}
				ms.Lv = finalLv
				ms.SeasonId = msid.SeasonId
				ms.IsAward = false
				ms.UpdateTs = time.Now().Unix()
				ms.dirty = true
				MatchSeasonMgrSington.SetMatchSeason(ms)
			}
			send(ms)
		}
	}
	return nil
}

// 赛季排行榜
type CSTMSeasonRankPacketFactory struct {
}
type CSTMSeasonRankHandler struct {
}

func (this *CSTMSeasonRankPacketFactory) CreatePacket() interface{} {
	pack := &tournament.CSTMSeasonRank{}
	return pack
}

func (this *CSTMSeasonRankHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSTMSeasonRankHandler Process recv ", data)
	if _, ok := data.(*tournament.CSTMSeasonRank); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warnf("CSTMSeasonRankHandler p == nil.")
			return nil
		}
		platform := p.Platform
		if platform == Default_Platform {
			logger.Logger.Warnf("CSTMSeasonRankHandler Platform == Default_Platform.")
			return nil
		}
		pack := &tournament.SCTMSeasonRank{}
		tmpMsrs := []*MatchSeasonRank{}
		msr := MatchSeasonRankMgrSington.GetMatchSeasonRank(platform)
		if msr != nil {
			for _, ms := range msr {
				sr := &MatchSeasonRank{
					SnId: ms.SnId,
					Name: ms.Name,
					Lv:   ms.Lv,
				}
				tmpMsrs = append(tmpMsrs, sr)
			}
			logger.Logger.Trace("真人排行：", len(tmpMsrs))
		}
		robotmsr := MatchSeasonRankMgrSington.GetRobotMatchSeasonRank(platform)
		if robotmsr != nil {
			for _, ms := range robotmsr {
				sr := &MatchSeasonRank{
					SnId: ms.SnId,
					Name: ms.Name,
					Lv:   ms.Lv,
				}
				tmpMsrs = append(tmpMsrs, sr)
			}
			logger.Logger.Trace("真人机器人排行：", len(tmpMsrs))
		}
		if tmpMsrs != nil && len(tmpMsrs) > 0 {
			sort.Slice(tmpMsrs, func(i, j int) bool {
				return tmpMsrs[i].Lv > tmpMsrs[j].Lv
			})
			if len(tmpMsrs) > model.GameParamData.MatchSeasonRankMaxNum {
				tmpMsrs = append(tmpMsrs[:model.GameParamData.MatchSeasonRankMaxNum])
			}
			for i := 0; i < len(tmpMsrs); i++ {
				ms := tmpMsrs[i]
				sr := &tournament.SeasonRank{
					Snid: ms.SnId,
					Name: ms.Name,
					Lv:   ms.Lv,
					Rank: int32(i) + 1,
				}
				pack.ReasonRanks = append(pack.ReasonRanks, sr)
			}
		}
		proto.SetDefaults(pack)
		logger.Logger.Trace("CSTMSeasonRankHandler:", pack)
		p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCTMSeasonRank), pack)
	}
	return nil
}

// 领取赛季奖励
type CSTMSeasonAwardPacketFactory struct {
}
type CSTMSeasonAwardHandler struct {
}

func (this *CSTMSeasonAwardPacketFactory) CreatePacket() interface{} {
	pack := &tournament.CSTMSeasonAward{}
	return pack
}

func (this *CSTMSeasonAwardHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSTMSeasonAwardHandler Process recv ", data)
	if msg, ok := data.(*tournament.CSTMSeasonAward); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warnf("CSTMSeasonAwardHandler p == nil.")
			return nil
		}
		if p.Platform == Default_Platform {
			logger.Logger.Warnf("CSTMSeasonInfoHandler Platform == Default_Platform.")
			return nil
		}
		lv := msg.GetLv()
		logger.Logger.Trace("CSTMSeasonAwardHandler lv: ", lv)
		pack := &tournament.SCTMSeasonAward{
			Lv:   lv,
			Code: 1,
		}
		ms := MatchSeasonMgrSington.GetMatchSeason(p.SnId)
		msi := MatchSeasonMgrSington.GetMatchSeasonId(p.Platform)
		if ms != nil && msi != nil {
			if !ms.IsAward && ms.LastLv == lv && msi.SeasonId > 1 { //领取上赛季奖励
				for _, v := range srvdata.PBDB_GamMatchLVMgr.Datas.GetArr() {
					if v.Star != nil && len(v.Star) > 1 {
						startStar := v.Star[0]
						endStar := v.Star[1]
						if lv >= startStar && lv <= endStar { //匹配段位
							pack.Code = 0
							MatchSeasonMgrSington.UpdateMatchSeasonAward(p.SnId)
							switch v.AwardType1 {
							case 1: //金币
								p.AddCoin(int64(v.Number1), common.GainWay_MatchSeason, "system", "赛季奖励")
							case 2: //钻石
								p.AddDiamond(int64(v.Number1), common.GainWay_MatchSeason, "system", "赛季奖励")
							case 3: //道具
								if v.Number1 > 0 {
									item := &Item{
										ItemId:  v.AwardId1,
										ItemNum: v.Number1,
									}
									BagMgrSington.AddJybBagInfo(p, []*Item{item})
									itemData := srvdata.PBDB_GameItemMgr.GetData(item.ItemId)
									if itemData != nil {
										BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemObtain, item.ItemId, itemData.Name, item.ItemNum, "赛季奖励")
									}
								}
							}
							switch v.AwardType2 {
							case 1: //金币
								p.AddCoin(int64(v.Number2), common.GainWay_MatchSeason, "system", "赛季奖励")
							case 2: //钻石
								p.AddDiamond(int64(v.Number2), common.GainWay_MatchSeason, "system", "赛季奖励")
							case 3: //道具
								if v.Number2 > 0 {
									item := &Item{
										ItemId:  v.AwardId2,
										ItemNum: v.Number2,
									}
									BagMgrSington.AddJybBagInfo(p, []*Item{item})
									itemData := srvdata.PBDB_GameItemMgr.GetData(item.ItemId)
									if itemData != nil {
										BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemObtain, item.ItemId, itemData.Name, item.ItemNum, "赛季奖励")
									}
								}
							}
							switch v.AwardType3 {
							case 1: //金币
								p.AddCoin(int64(v.Number3), common.GainWay_MatchSeason, "system", "赛季奖励")
							case 2: //钻石
								p.AddDiamond(int64(v.Number3), common.GainWay_MatchSeason, "system", "赛季奖励")
							case 3: //道具
								if v.Number3 > 0 {
									item := &Item{
										ItemId:  v.AwardId3,
										ItemNum: v.Number3,
									}
									BagMgrSington.AddJybBagInfo(p, []*Item{item})
									itemData := srvdata.PBDB_GameItemMgr.GetData(item.ItemId)
									if itemData != nil {
										BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemObtain, item.ItemId, itemData.Name, item.ItemNum, "赛季奖励")
									}
								}
							}
							break
						}
					}
				}
			}
		}
		proto.SetDefaults(pack)
		logger.Logger.Trace("SCTMSeasonAward:", pack)
		p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCTMSeasonAward), pack)
	}
	return nil
}

func init() {
	common.RegisterHandler(int(tournament.TOURNAMENTID_PACKET_TM_CSTMInfo), &CSTMInfoHandler{})
	netlib.RegisterFactory(int(tournament.TOURNAMENTID_PACKET_TM_CSTMInfo), &CSTMInfoPacketFactory{})

	common.RegisterHandler(int(tournament.TOURNAMENTID_PACKET_TM_CSTMRankList), &CSTMRankListHandler{})
	netlib.RegisterFactory(int(tournament.TOURNAMENTID_PACKET_TM_CSTMRankList), &CSTMRankListPacketFactory{})

	common.RegisterHandler(int(tournament.TOURNAMENTID_PACKET_TM_CSSignRace), &CSSignRaceHandler{})
	netlib.RegisterFactory(int(tournament.TOURNAMENTID_PACKET_TM_CSSignRace), &CSSignRacePacketFactory{})

	common.RegisterHandler(int(tournament.TOURNAMENTID_PACKET_TM_CSTMSeasonInfo), &CSTMSeasonInfoHandler{})
	netlib.RegisterFactory(int(tournament.TOURNAMENTID_PACKET_TM_CSTMSeasonInfo), &CSTMSeasonInfoPacketFactory{})
	common.RegisterHandler(int(tournament.TOURNAMENTID_PACKET_TM_CSTMSeasonRank), &CSTMSeasonRankHandler{})
	netlib.RegisterFactory(int(tournament.TOURNAMENTID_PACKET_TM_CSTMSeasonRank), &CSTMSeasonRankPacketFactory{})
	common.RegisterHandler(int(tournament.TOURNAMENTID_PACKET_TM_CSTMSeasonAward), &CSTMSeasonAwardHandler{})
	netlib.RegisterFactory(int(tournament.TOURNAMENTID_PACKET_TM_CSTMSeasonAward), &CSTMSeasonAwardPacketFactory{})
}
