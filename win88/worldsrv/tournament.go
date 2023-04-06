package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/protocol/tournament"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/core/timer"
	srvproto "github.com/idealeak/goserver/srvlib/protocol"
	"math/rand"
	"strconv"
	"time"
)

// tournament

var TournamentMgr = &Tournament{
	GameMatchDateList:  make(map[string]map[int32]*webapi_proto.GameMatchDate), //比赛配置
	matches:            make(map[int32]map[int32]*TmMatch),
	signUpPlayers:      make(map[int32]*SignInfo),
	players:            make(map[int32]map[int32]*MatchContext),
	copyPlayers:        make(map[int32]map[int32][]*MatchContext),
	rankPlayers:        make(map[int32]map[int32][]*MatchContext),
	roundOverPlayerNum: make(map[int32]map[int32]int32),
	finalPerRank:       make(map[int32][]*PerRankInfo),
}

type Tournament struct {
	GameMatchDateList map[string]map[int32]*webapi_proto.GameMatchDate

	matches map[int32]map[int32]*TmMatch //key:比赛配置Id,比赛顺序序号

	signUpPlayers map[int32]*SignInfo //锦标赛报名的玩家 key:比赛配置Id

	players            map[int32]map[int32]*MatchContext   //比赛玩家 比赛顺序序号,snid
	copyPlayers        map[int32]map[int32][]*MatchContext //比赛玩家 比赛顺序序号,round
	rankPlayers        map[int32]map[int32][]*MatchContext //比赛玩家为了每轮积分排名用 比赛顺序序号,round
	roundOverPlayerNum map[int32]map[int32]int32           //本轮比赛完的玩家人数
	finalPerRank       map[int32][]*PerRankInfo            //本场比赛最后排名,每淘汰一位记录一位，最后记录决赛玩家
}

type PerRankInfo struct {
	Name   string
	SnId   int32
	RankId int32
	Grade  int32
}

type SignInfo struct {
	signup   map[int32]*TmPlayer //玩家Id
	Platform string
	GameId   int32
	MaxCnt   int
}

// 玩家比赛结束 更新积分
func (this *Tournament) UpdateMatchInfo(p *Player, sortId, grade, isWin int32) {
	logger.Logger.Trace("=========== UpdateMatchInfo ==============")
	if _, ok := this.players[sortId]; !ok {
		this.players[sortId] = make(map[int32]*MatchContext)
	}
	if mtp, ok := this.players[sortId][p.SnId]; ok {
		mtp.grade = grade
		if this.roundOverPlayerNum[sortId] == nil {
			this.roundOverPlayerNum[sortId] = make(map[int32]int32)
		}
		this.roundOverPlayerNum[sortId][mtp.round]++
		if mtp.record == nil {
			mtp.record = make(map[int32]int32)
		}
		if _, ok := mtp.record[isWin]; !ok {
			mtp.record[isWin] = 0
		}
		mtp.record[isWin]++
		//轮数增加
		mtp.round++
		mtp.gaming = false
		if this.copyPlayers[sortId] == nil {
			this.copyPlayers[sortId] = make(map[int32][]*MatchContext)
		}
		var mc MatchContext
		mc = *mtp
		this.copyPlayers[sortId][mtp.round] = append(this.copyPlayers[sortId][mtp.round], &mc)
		if this.rankPlayers[sortId] == nil {
			this.rankPlayers[sortId] = make(map[int32][]*MatchContext)
		}
		this.rankPlayers[sortId][mtp.round] = append(this.rankPlayers[sortId][mtp.round], &mc)
		logger.Logger.Tracef("========snid(%v)   grade(%v)   mtp(%v)============", p.SnId, grade, mtp)
		this.NextRoundStart(sortId, mtp)
	}
}

func (this *Tournament) CreatePlayerMatchContext(p *Player, m *TmMatch, seq int) *MatchContext {
	mc := NewMatchContext(p, m, 1000, seq)
	if mc != nil {
		if this.players[m.SortId] == nil {
			this.players[m.SortId] = make(map[int32]*MatchContext)
		}
		this.players[m.SortId][p.SnId] = mc
		p.matchCtx = mc
		return mc
	}
	return nil
}

// 更新配置
func (this *Tournament) UpdateData(init bool, cfgs *webapi_proto.GameMatchDateList) {
	//logger.Logger.Trace("(this *Tournament) UpdateData:", cfgs)
	if cfgs.Platform == "0" {
		return
	}
	gmds := make(map[int32]*webapi_proto.GameMatchDate)
	for _, v := range cfgs.List {
		gmds[v.Id] = v
	}
	oldgmds := this.GameMatchDateList[cfgs.Platform]
	//旧配置关闭踢人 --旧配置关闭不影响已经开始的只取消报名未开始的
	for _, v := range oldgmds {
		gmd := gmds[v.Id]
		if (gmd != nil && v.MatchSwitch == 1 && gmd.MatchSwitch != v.MatchSwitch) || //配置关闭
			(gmd != nil && !this.IsTimeRange(gmd)) || //或者不在时间段内
			gmd == nil { //或者删除
			signInfo := this.signUpPlayers[v.Id]
			if signInfo != nil && len(signInfo.signup) > 0 {
				this.CancelSignUpAll(v.Id)
			}
		}
	}
	//新配置增加更新
	for _, v := range gmds {
		oldgmd := oldgmds[v.Id]
		if oldgmd == nil || v.MatchSwitch == 1 {
			var gameId int32
			gf := srvdata.PBDB_GameFreeMgr.GetData(v.GameFreeId)
			if gf != nil {
				gameId = gf.GameId
			}
			this.signUpPlayers[v.Id] = &SignInfo{
				signup:   make(map[int32]*TmPlayer),
				Platform: cfgs.Platform,
				GameId:   gameId,
				MaxCnt:   int(v.MatchNumebr),
			}
		}
		if v.MatchSwitch == 2 || !this.IsTimeRange(v) {
			this.CancelSignUpAll(v.Id)
			if v.MatchSwitch == 2 {
				delete(this.signUpPlayers, v.Id)
			}
		}
	}
	this.GameMatchDateList[cfgs.Platform] = gmds

	if !init {
		//通知平台玩家数据更新
		pack := TournamentMgr.GetSCTMInfosPack(cfgs.Platform)
		proto.SetDefaults(pack)
		logger.Logger.Trace("SCTMInfos++++++++++++:", pack)
		PlayerMgrSington.BroadcastMessageToPlatform(cfgs.Platform, int(tournament.TOURNAMENTID_PACKET_TM_SCTMInfos), pack)
	}
}

// 比赛开关
func (this *Tournament) MatchSwitch(platform string, tmId int32) bool {
	if list, ok := this.GameMatchDateList[platform]; ok {
		if gmd, ok1 := list[tmId]; ok1 {
			return gmd.MatchSwitch == 1
		}
	}
	return false
}

// 比赛配置
func (this *Tournament) GetMatchInfo(platform string, tmId int32) *webapi_proto.GameMatchDate {
	if list, ok := this.GameMatchDateList[platform]; ok {
		if gmd, ok1 := list[tmId]; ok1 {
			return gmd
		}
	}
	return nil
}

// 周几
func (this *Tournament) getWeekNum(t time.Time) int {
	strWeek := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	week := t.Weekday()
	for i, s := range strWeek {
		if week.String() == s {
			return i
		}
	}
	return 0
}

// 判断是否在比赛时间段内
func (this *Tournament) IsTimeRange(gmd *webapi_proto.GameMatchDate) bool {
	if gmd != nil {
		if gmd.MatchType == 2 { //冠军赛
			tNow := time.Now()
			switch gmd.MatchTimeType { // 0无时效 1重复时间段 2一次性时间段
			case 0:
				return true
			case 1:
				nowWeek := this.getWeekNum(tNow)
				hms := int32(tNow.Hour()*10000 + tNow.Minute()*100 + tNow.Second())
				if gmd.MatchTimeWeek != nil && len(gmd.MatchTimeWeek) > 0 {
					for _, week := range gmd.MatchTimeWeek {
						if nowWeek == int(week) {
							if hms >= gmd.MatchTimeStartHMS && hms <= gmd.MatchTimeEndHMS {
								return true
							}
						}
					}
				}
			case 2:
				if gmd.MatchTimeStamp != nil && len(gmd.MatchTimeStamp) > 1 {
					startStamp := gmd.MatchTimeStamp[0]
					endStamp := gmd.MatchTimeStamp[1]
					if tNow.Unix() >= startStamp && tNow.Unix() <= endStamp {
						return true
					}
				}
			}
		} else { //锦标赛没有时间限制
			return true
		}
	}
	return false
}

// 是否过期
func (this *Tournament) IsOutTime(gmd *webapi_proto.GameMatchDate) bool {
	if gmd == nil || gmd.MatchSwitch != 1 {
		return true
	}
	if gmd.MatchType == 2 { //冠军赛的一次性时间段才有过期可能
		switch gmd.MatchTimeType { // 0无时效 1重复时间段 2一次性时间段
		case 2:
			tNow := time.Now()
			if gmd.MatchTimeStamp != nil && len(gmd.MatchTimeStamp) > 1 {
				endStamp := gmd.MatchTimeStamp[1]
				if tNow.Unix() >= endStamp { //当前时间大于截止时间
					return true
				}
			}
		}
	}
	return false
}

// 比赛配置数据
func (this *Tournament) GetAllMatchInfo(platform string) map[int32]*webapi_proto.GameMatchDate {
	if list, ok := this.GameMatchDateList[platform]; ok {
		return list
	}
	return nil
}

// 判断是否在比赛中
func (this *Tournament) IsMatching(snid int32) bool {
	if this.players != nil {
		for _, m := range this.players {
			if m != nil {
				for mSnid, _ := range m {
					if mSnid == snid {
						return true
					}
				}
			}
		}
	}
	return false
}

// 判断是否在匹配中
func (this *Tournament) IsMatchWaiting(snid int32) (bool, int32) {
	if this.signUpPlayers != nil {
		for tmid, info := range this.signUpPlayers {
			if info.signup != nil {
				for sid, _ := range info.signup {
					if sid == snid {
						return true, tmid
					}
				}
			}
		}
	}
	return false, 0
}

// 发送比赛配置数据
func (this *Tournament) GetSCTMInfosPack(platform string) *tournament.SCTMInfos {
	pack := &tournament.SCTMInfos{}
	matchInfo := this.GetAllMatchInfo(platform)
	if matchInfo != nil {
		for id, info := range matchInfo {
			if info.MatchSwitch == 1 && !this.IsOutTime(info) {
				tMInfo := &tournament.TMInfo{
					Id:                proto.Int32(id),
					GameFreeId:        info.GameFreeId,
					MatchType:         info.MatchType,
					MatchName:         info.MatchName,
					MatchNumebr:       info.MatchNumebr,
					MatchSwitch:       info.MatchSwitch,
					SignupCostCoin:    info.SignupCostCoin,
					SignupCostDiamond: info.SignupCostDiamond,
					MatchTimeType:     info.MatchTimeType,
					MatchTimeStartHMS: info.MatchTimeStartHMS,
					MatchTimeEndHMS:   info.MatchTimeEndHMS,
					TitleURL:          info.TitleURL,
					AwardShow:         info.AwardShow,
					Rule:              info.Rule,
				}
				if info.MatchTimeWeek != nil && len(info.MatchTimeWeek) > 0 {
					for _, week := range info.MatchTimeWeek {
						tMInfo.MatchTimeWeek = append(tMInfo.MatchTimeWeek, week)
					}
				}
				if info.MatchTimeStamp != nil && len(info.MatchTimeStamp) > 0 {
					for _, stamp := range info.MatchTimeStamp {
						tMInfo.MatchTimeStamp = append(tMInfo.MatchTimeStamp, stamp)
					}
				}
				if info.MatchPromotion != nil {
					for _, mp := range info.MatchPromotion {
						tMInfo.MatchPromotion = append(tMInfo.MatchPromotion, mp)
					}
				}
				if info.Award != nil {
					for _, award := range info.Award {
						miAward := &tournament.MatchInfoAward{
							Coin:      award.Coin,
							Diamond:   award.Diamond,
							UpLimit:   award.UpLimit,
							DownLimit: award.DownLimit,
						}
						if award.ItemId != nil {
							for _, itemInfo := range award.ItemId {
								a := &tournament.ItemInfo{
									ItemId:  itemInfo.ItemId,
									ItemNum: itemInfo.ItemNum,
									Name:    itemInfo.Name,
								}
								miAward.ItemInfo = append(miAward.ItemInfo, a)
							}
						}
						tMInfo.Award = append(tMInfo.Award, miAward)
					}
				}
				if info.SignupCostItem != nil && info.SignupCostItem.ItemNum > 0 {
					signupCost := &tournament.ItemInfo{
						ItemId:  info.SignupCostItem.ItemId,
						ItemNum: info.SignupCostItem.ItemNum,
						Name:    info.SignupCostItem.Name,
					}
					tMInfo.SignupCostItem = signupCost
				}
				pack.TMInfo = append(pack.TMInfo, tMInfo)
			}
		}
	}
	return pack
}

// 报名
func (this *Tournament) SignUp(tmId int32, p *Player) (bool, int32) {
	logger.Logger.Trace("(this *Tournament) SignUp:", tmId, p.SnId)
	//报名费
	if !p.IsRob {
		//0成功 1重复报名 2比赛没有开启 3道具不足 4不在报名时间段 5金币不足 6钻石不足
		ok, code := this.SignUpCost(tmId, p, true)
		if !ok {
			return false, code
		}
	}
	var signInfo *SignInfo
	if _, ok := this.signUpPlayers[tmId]; !ok {
		platform := p.Platform
		if p.IsRob {
			platform = "1"
		}
		gmd := this.GetMatchInfo(platform, tmId)
		if gmd != nil {
			gcf := srvdata.PBDB_GameFreeMgr.GetData(gmd.GameFreeId)
			signInfo = &SignInfo{
				signup:   make(map[int32]*TmPlayer),
				Platform: platform,
				GameId:   gcf.GameId,
			}
			this.signUpPlayers[tmId] = signInfo
		}
	} else {
		signInfo = this.signUpPlayers[tmId]
	}
	if _, ok := signInfo.signup[p.SnId]; !ok {
		n := len(signInfo.signup) + 1
		signInfo.signup[p.SnId] = &TmPlayer{SnId: p.SnId, IsRob: p.IsRob, seq: n}
		if p.IsRob {
			logger.Logger.Trace("Ai 报名.............", n, " p.SnId: ", p.SnId)
		} else {
			logger.Logger.Trace("真人 报名.............", n, " p.SnId: ", p.SnId)
		}
	} else {
		return false, 1
	}

	this.SyncSignNum(tmId)

	return true, 0
}

// 报名费用 0成功 1重复报名 2比赛没有开启 3道具不足 4不在报名时间段 5金币不足 6钻石不足
func (this *Tournament) SignUpCost(tmId int32, p *Player, cost bool) (bool, int32) {
	logger.Logger.Trace("报名费用 SignUpCost: ", tmId, " snid: ", p.SnId, " cost: ", cost)
	if p == nil {
		return false, 3
	}
	gmd := this.GetMatchInfo(p.Platform, tmId)
	if gmd != nil && !p.IsRob { //真人费用检测
		if gmd.SignupCostCoin > 0 {
			logger.Logger.Trace("比赛场报名消耗金币", cost, gmd.SignupCostCoin)
			if cost {
				if p.Coin < gmd.SignupCostCoin { //金币不足
					logger.Logger.Trace("金币不足")
					return false, 5
				} else {
					p.AddCoin(-gmd.SignupCostCoin, common.GainWay_MatchSignup, "system", gmd.MatchName+"-报名消耗")
				}
			} else {
				p.AddCoin(gmd.SignupCostCoin, common.GainWay_MatchSignup, "system", gmd.MatchName+"-报名退还")
			}
		}
		if gmd.SignupCostDiamond > 0 {
			logger.Logger.Trace("比赛场报名消耗钻石", cost, gmd.SignupCostDiamond)
			if cost {
				if p.Diamond < gmd.SignupCostDiamond { //钻石不足
					logger.Logger.Trace("钻石不足")
					return false, 6
				} else {
					p.AddDiamond(-gmd.SignupCostDiamond, common.GainWay_MatchSignup, "system", gmd.MatchName+"-报名消耗")
				}
			} else {
				p.AddDiamond(gmd.SignupCostDiamond, common.GainWay_MatchSignup, "system", gmd.MatchName+"-报名退还")
			}
		}
		if gmd.SignupCostItem != nil && gmd.SignupCostItem.ItemNum > 0 {
			//背包数据处理
			logger.Logger.Trace("比赛场报名消耗道具", cost, gmd.SignupCostItem.ItemNum)
			item := BagMgrSington.GetBagItemById(p.SnId, gmd.SignupCostItem.ItemId)
			if item != nil {
				if cost {
					if item.ItemNum < gmd.SignupCostItem.ItemNum {
						logger.Logger.Trace("道具不足")
						return false, 3
					} else {
						item.ItemNum -= gmd.SignupCostItem.ItemNum
						p.dirty = true
						BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemConsume, item.ItemId, item.Name, gmd.SignupCostItem.ItemNum, gmd.MatchName+"-报名消耗")
						BagMgrSington.SyncBagData(p, item.ItemId)
					}
				} else {
					item.ItemNum += gmd.SignupCostItem.ItemNum
					p.dirty = true
					BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemObtain, item.ItemId, item.Name, gmd.SignupCostItem.ItemNum, gmd.MatchName+"-报名退还")
					BagMgrSington.SyncBagData(p, item.ItemId)
				}
			} else {
				logger.Logger.Trace("道具不足")
				return false, 3
			}
		}
	}
	return true, 0
}

func (this *Tournament) CheckStart(tmId int32) {
	if signInfo, ok := this.signUpPlayers[tmId]; ok {
		n := len(signInfo.signup)
		matchInfo := this.GetMatchInfo(signInfo.Platform, tmId)
		if matchInfo != nil && matchInfo.MatchNumebr == int32(n) {
			logger.Logger.Trace("TournamentMgr.CheckStart: ", tmId)
			hasReal := false
			canStart := true
			for _, v := range signInfo.signup {
				p := PlayerMgrSington.GetPlayerBySnId(v.SnId)
				if p == nil { //人不在
					canStart = false
					this.CancelSignUp(tmId, v.SnId, true)
					break
				}
				if p.scene != nil { //人在游戏内
					canStart = false
					this.CancelSignUp(tmId, v.SnId, true)
					break
				}
				if !v.IsRob {
					hasReal = true
					break
				}
			}
			if canStart {
				//有真人
				if hasReal {
					//人满 开始比赛
					this.Start(tmId, matchInfo)
				} else {
					this.CancelSignUpAll(tmId)
				}
			}
		}
	}
}

func (this *Tournament) CancelSignUpAll(tmId int32) {
	logger.Logger.Trace("CancelSignUpAll", tmId)
	if this.signUpPlayers[tmId] == nil {
		return
	}
	for _, tmp := range this.signUpPlayers[tmId].signup {
		this.CancelSignUp(tmId, tmp.SnId, false)
	}
}

// 取消报名
func (this *Tournament) CancelSignUp(tmid, snid int32, isSync bool) {
	if this.signUpPlayers[tmid] == nil {
		return
	}
	signInfo := this.signUpPlayers[tmid]
	if _, ok := signInfo.signup[snid]; ok {
		p := PlayerMgrSington.GetPlayerBySnId(snid)
		if p != nil && p.scene == nil {
			//退费
			if !p.IsRob {
				this.SignUpCost(tmid, p, false)
			}
			delete(this.signUpPlayers[tmid].signup, snid)
			//通知取消报名
			pack := &tournament.SCSignRace{
				OpCode:  1,
				RetCode: 0,
			}
			proto.SetDefaults(pack)
			if !p.IsRob {
				logger.Logger.Trace("真人取消报名: ", pack)
			}
			p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCSignRace), pack)
		}
	}
	if isSync {
		this.SyncSignNum(tmid)
	}
}

// 通知报名的赛场人数变动
func (this *Tournament) SyncSignNum(tmid int32) {
	var n int
	var maxN int
	if this.signUpPlayers[tmid] != nil {
		n = len(this.signUpPlayers[tmid].signup)
		maxN = this.signUpPlayers[tmid].MaxCnt
	}
	if n > maxN {
		n = maxN
	}
	pack := &tournament.SCSyncSignNum{
		SignNum:    proto.Int(n),
		MaxSignNum: proto.Int(maxN),
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("SCSyncSignNum:", pack)
	this.BroadcastMessage(tmid, int(tournament.TOURNAMENTID_PACKET_TM_SCSyncSignNum), pack)
}

func (this *Tournament) BroadcastMessage(tmid int32, packetid int, rawpack interface{}) {
	mgs := make(map[*netlib.Session][]*srvproto.MCSessionUnion)
	if this.signUpPlayers[tmid] == nil {
		return
	}
	for _, tmp := range this.signUpPlayers[tmid].signup {
		p := PlayerMgrSington.GetPlayerBySnId(tmp.SnId)
		if p != nil && p.gateSess != nil && p.IsOnLine() && p.scene == nil && !p.IsRob {
			mgs[p.gateSess] = append(mgs[p.gateSess], &srvproto.MCSessionUnion{
				Mccs: &srvproto.MCClientSession{
					SId: proto.Int64(p.sid),
				},
			})
		}
	}
	for gateSess, v := range mgs {
		if gateSess != nil && len(v) != 0 {
			pack, err := MulticastMaker.CreateMulticastPacket(packetid, rawpack, v...)
			if err == nil {
				proto.SetDefaults(pack)
				gateSess.Send(int(srvproto.SrvlibPacketID_PACKET_SS_MULTICAST), pack)
			}
		}
	}
}

// 开赛
func (this *Tournament) Start(tmId int32, matchInfo *webapi_proto.GameMatchDate) {
	signInfo := this.signUpPlayers[tmId]
	sortId := int32(time.Now().Nanosecond())
	tm := &TmMatch{
		TMId:     tmId,
		SortId:   sortId,
		gmd:      matchInfo,
		Platform: signInfo.Platform,
	}
	//获取游戏配置
	platform := PlatformMgrSington.GetPlatform(tm.Platform)
	if platform != nil {
		gameCfg := platform.PltGameCfg.GetGameCfg(matchInfo.GameFreeId)
		if gameCfg != nil {
			tm.dbGameFree = gameCfg.DbGameFree
		} else {
			tm.dbGameFree = srvdata.PBDB_GameFreeMgr.GetData(matchInfo.GameFreeId)
		}
	}
	tm.CopyMap(signInfo.signup)
	if this.matches[tmId] == nil {
		this.matches[tmId] = make(map[int32]*TmMatch)
	}
	this.matches[tmId][sortId] = tm
	tm.Start()
	this.signUpPlayers[tmId].signup = make(map[int32]*TmPlayer)
}

// 下一轮是否开始
func (this *Tournament) NextRoundStart(sortId int32, mtp *MatchContext) {
	logger.Logger.Tracef("================NextRoundStart==========当前第 %v 轮============", mtp.round)
	if _, ok := this.copyPlayers[sortId]; ok {
		if mcs, ok1 := this.copyPlayers[sortId][mtp.round]; ok1 {
			gmd := mtp.tm.gmd
			//需要晋级的人数
			promotionNum1 := gmd.MatchPromotion[mtp.round-1]
			promotionNum := gmd.MatchPromotion[mtp.round]
			//通知晋级信息：0.晋级等待匹配 1.失败退出 2.等待判断是否晋级
			pack := &tournament.SCPromotionInfo{}
			if promotionNum != 1 {
				n := int32(len(mcs))
				//非决赛淘汰后开始配桌
				logger.Logger.Trace("非决赛开始淘汰晋级")
				outNum := promotionNum1 - promotionNum
				//已经晋级的人数减去一桌之后  剩余人数还能够满足本轮淘汰
				logger.Logger.Trace("n: ", n, " outNum: ", outNum)
				if n-4 >= outNum {
					//提前晋级的开始凑桌
					MatchContextSlice(mcs).Sort(false)
					//挑选出晋级的玩家
					meIn := false //自己晋级
					for k, v := range mcs {
						if mtp.p.SnId == v.p.SnId {
							meIn = true
						}
						logger.Logger.Trace("排序之后=========== ", k, v.p.SnId, v.round, v.seq, v.grade)
					}
					mct := []*MatchContext{}
					finals := false
					for i := 0; i < len(mcs)-int(outNum); i++ {
						var mc MatchContext
						mc = *mcs[i]
						this.sendPromotionInfo(mcs[i], sortId, 0, false, false) //晋级
						mc.rank = mcs[i].rank
						mct = append(mct, &mc)
						logger.Logger.Trace("======凑桌==========mc=================", mc)
						if !finals && mc.round == int32(len(gmd.MatchPromotion)-2) {
							finals = true
						}
					}
					mcs = mcs[len(mct):]
					this.copyPlayers[sortId][mtp.round] = this.copyPlayers[sortId][mtp.round][len(mct):]
					willOut := false
					if promotionNum1 == this.roundOverPlayerNum[sortId][mtp.round-1] {
						//最后一个人打完了，确定要淘汰的人
						willOut = true
					} else {
						if !meIn { //自己暂时没晋级
							this.sendPromotionInfo(mtp, sortId, 2, false, false) //待定
						}
					}
					if this.finalPerRank[sortId] == nil {
						this.finalPerRank[sortId] = []*PerRankInfo{}
					}
					isOver := false
					for k, v := range this.copyPlayers[sortId][mtp.round] {
						logger.Logger.Trace("凑桌之后剩余===2======== ", k, v.p.SnId, v.round, v.seq, v.grade)
						if willOut {
							this.sendPromotionInfo(v, sortId, 1, true, false) //淘汰
							//把淘汰的玩家记录在排行榜
							pri := &PerRankInfo{
								Name:   v.p.Name,
								SnId:   v.p.SnId,
								RankId: pack.RankId,
								Grade:  v.grade,
							}
							this.finalPerRank[sortId] = append(this.finalPerRank[sortId], pri)
							//真人被淘汰，如果剩下的都是机器人，比赛解散
							if !v.p.IsRob {
								if this.players[sortId] != nil {
									hasReal := false
									for snid, context := range this.players[sortId] {
										if !context.p.IsRob && !this.isOut(sortId, snid) { // 有真人没有淘汰
											hasReal = true
											break
										}
									}
									//没有真人比赛解散
									if !hasReal {
										isOver = true
										logger.Logger.Trace("没有真人比赛解散")
										this.StopMatch(mtp.tm.TMId, sortId)
									}
								}
							}
						}
					}
					if !isOver {
						timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
							MatchSceneMgrSington.NewRoundStart(mtp.tm, mct, finals, mtp.round+1)
							return true
						}), nil, time.Second*7, 1)
					}
				} else {
					this.sendPromotionInfo(mtp, sortId, 2, false, false) //待定
				}
			} else {
				if len(mcs) == 4 {
					MatchContextSlice(mcs).Sort(true)
					if this.finalPerRank[sortId] == nil {
						this.finalPerRank[sortId] = []*PerRankInfo{}
					}
					for _, mc := range mcs {
						pri := &PerRankInfo{
							Name:   mc.p.Name,
							SnId:   mc.p.SnId,
							RankId: mc.rank,
							Grade:  mc.grade,
						}
						this.finalPerRank[sortId] = append(this.finalPerRank[sortId], pri) //把决赛的玩家记录在排行榜
						outCode := int32(1)
						if mc.rank == 1 {
							outCode = 0
						}
						this.sendPromotionInfo(mc, sortId, outCode, true, true) //晋级
					}
					logger.Logger.Trace("比赛结束!!! ")
					// 比赛结束
					this.StopMatch(mtp.tm.TMId, sortId)
				}
			}
		}
	}
}

// 发送晋级信息
func (this *Tournament) sendPromotionInfo(mc *MatchContext, sortId, outCode int32, isOver, isFinals bool) {
	if mc == nil {
		return
	}
	rankId := this.getRank(sortId, mc.round, mc.p.SnId, isFinals)
	mc.rank = rankId
	//通知晋级信息：0.晋级等待匹配 1.失败退出 2.等待判断是否晋级
	pack := &tournament.SCPromotionInfo{}
	pack.RetCode = outCode
	pack.Round = mc.round
	pack.RankId = rankId
	pack.RoundCoin = 100 //暂时用不到先写死
	pack.Record = mc.record
	if mc.tm != nil {
		pack.MatchId = mc.tm.TMId
	}
	if mc.tm != nil && mc.tm.gmd != nil {
		pack.MatchPromotion = mc.tm.gmd.GetMatchPromotion()
		if !mc.p.IsRob && isOver { //真人发奖
			if mc.tm.gmd.Award != nil {
				for _, award := range mc.tm.gmd.Award {
					if rankId >= award.UpLimit && rankId <= award.DownLimit { //上下限是反的，我也是醉了
						rankAward := &tournament.RankAward{
							Coin:    proto.Int64(award.Coin),
							Diamond: proto.Int64(award.Diamond),
						}
						if award.ItemId != nil {
							for _, info := range award.ItemId {
								item := &tournament.ItemInfo{
									ItemId:  proto.Int32(info.ItemId),
									ItemNum: proto.Int32(info.ItemNum),
									Name:    proto.String(info.Name),
								}
								rankAward.ItemInfo = append(rankAward.ItemInfo, item)
							}
						}
						pack.RankAward = rankAward
					}
				}
			}
		}
	}

	outStr := "晋级"
	switch outCode {
	case 1:
		outStr = "淘汰"
	case 2:
		outStr = "待定"
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("sendPromotionInfo: ", outStr, " snid: ", mc.p.SnId, " pack: ", pack)
	ok := mc.p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCPromotionInfo), pack)
	if ok && !mc.p.IsRob && isOver {
		if pack.RankAward != nil {
			if pack.RankAward.Coin != 0 { //金币
				mc.p.AddCoin(pack.RankAward.Coin, common.GainWay_MatchSystemSupply, "system", mc.tm.gmd.MatchName+"排名奖励")
			}
			if pack.RankAward.Diamond != 0 { //钻石
				mc.p.AddDiamond(pack.RankAward.Diamond, common.GainWay_MatchSystemSupply, "system", mc.tm.gmd.MatchName+"排名奖励")
			}
			if pack.RankAward.ItemInfo != nil {
				for _, info := range pack.RankAward.ItemInfo {
					if info.ItemNum > 0 {
						item := &Item{
							ItemId:  info.ItemId,
							ItemNum: info.ItemNum,
						}
						BagMgrSington.AddJybBagInfo(mc.p, []*Item{item})
						data := srvdata.PBDB_GameItemMgr.GetData(item.ItemId)
						if data != nil {
							BagMgrSington.RecordItemLog(mc.p.Platform, mc.p.SnId, ItemObtain, item.ItemId, data.Name, item.ItemNum, mc.tm.gmd.MatchName+"排名奖励")
						}
					}
				}
			}
		}
	}
	if isOver && mc.tm != nil { // 自己比赛结束
		pack := &tournament.SCTMStop{
			MatchId: proto.Int32(mc.tm.TMId),
		}
		proto.SetDefaults(pack)
		logger.Logger.Trace("SCTMStop:", pack)
		mc.p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCTMStop), pack)
		if !mc.p.IsRob {
			this.CheckAddMatchSeasonLv(mc)
		}
		delete(this.players[sortId], mc.p.SnId)
	}
}

// 更新段位
func (this *Tournament) CheckAddMatchSeasonLv(mc *MatchContext) {
	if mc == nil {
		return
	}
	platform := mc.p.Platform
	if platform == Default_Platform {
		return
	}
	rank := mc.rank
	maxPlayerNum := mc.tm.gmd.MatchNumebr
	upLine := maxPlayerNum * 33 / 100
	downLine := maxPlayerNum * 67 / 100
	snid := mc.p.SnId
	ms := MatchSeasonMgrSington.GetMatchSeason(mc.p.SnId)
	msid := MatchSeasonMgrSington.GetMatchSeasonId(platform)
	if msid == nil {
		MatchSeasonMgrSington.UpdateMatchSeasonId(platform)
	}
	if ms == nil {
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			ret, err := model.QueryMatchSeasonBySnid(platform, snid)
			if err != nil {
				return nil
			}
			return ret
		}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			ret := data.(*model.MatchSeason)
			dirty := false
			seasonId := int32(1)
			if msid != nil {
				seasonId = msid.SeasonId
			}
			if ret == nil {
				player := PlayerMgrSington.GetPlayerBySnId(snid)
				if player != nil {
					ret = model.NewMatchSeason(player.Platform, snid, player.Name, seasonId, 1)
					dirty = true
				} else {
					logger.Logger.Error("CSTMSeasonInfoHandler error: player == nil or  msi == nil", snid)
				}
			} else {
				if ret.SeasonId < seasonId { //不同赛季段位继承
					num := seasonId - ret.SeasonId
					finalLv := ret.Lv
					for i := 0; i < int(num); i++ { //继承几次
						if i == int(num)-1 { //上个赛季
							ret.LastLv = finalLv
						}
						finalLv = MatchSeasonMgrSington.MatchSeasonInherit(finalLv)
					}
					ret.Lv = finalLv
					ret.SeasonId = seasonId
					ret.IsAward = false
					dirty = true
				}
			}
			ms = MatchSeasonMgrSington.exchangeModel2Cache(ret)
			ms.dirty = dirty
			MatchSeasonMgrSington.SetMatchSeason(ms)
			logger.Logger.Tracef("UpdateMatchSeasonLv==1==rank:%v downLine:%v upLine:%v ", rank, downLine, upLine)
			if rank <= upLine { //加分
				MatchSeasonMgrSington.UpdateMatchSeasonLv(mc.p, 1)
			} else if rank >= downLine && ms.Lv > 75 { //白银以上才触发减分
				MatchSeasonMgrSington.UpdateMatchSeasonLv(mc.p, -1)
			} else {
				MatchSeasonMgrSington.UpdateMatchSeasonLv(mc.p, 0)
			}
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
		logger.Logger.Tracef("UpdateMatchSeasonLv==2==rank:%v downLine:%v upLine:%v ", rank, downLine, upLine)
		if rank <= upLine { //加分
			MatchSeasonMgrSington.UpdateMatchSeasonLv(mc.p, 1)
		} else if rank >= downLine && ms.Lv > 75 { //白银以上才触发减分
			MatchSeasonMgrSington.UpdateMatchSeasonLv(mc.p, -1)
		} else {
			MatchSeasonMgrSington.UpdateMatchSeasonLv(mc.p, 0)
		}
	}
}

// 比赛结束
func (this *Tournament) StopMatch(tmId, sortId int32) {
	//房间清理
	if this.matches[tmId] != nil && this.matches[tmId][sortId] != nil {
		this.matches[tmId][sortId].Stop()
		delete(this.matches[tmId], sortId)
	}
	//数据清理
	if this.players[sortId] != nil {
		for _, context := range this.players[sortId] {
			if context.p.IsRob {
				////通知客户端比赛结束(主要是机器人取消标记)
				pack := &tournament.SCTMStop{}
				proto.SetDefaults(pack)
				logger.Logger.Trace("SCTMStop:", pack)
				context.p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCTMStop), pack)
			}
		}
		this.players[sortId] = nil
	}
	if this.copyPlayers[sortId] != nil {
		this.copyPlayers[sortId] = nil
	}
	if this.rankPlayers[sortId] != nil {
		this.rankPlayers[sortId] = nil
	}
	if this.roundOverPlayerNum[sortId] != nil {
		this.roundOverPlayerNum[sortId] = nil
	}
	if this.finalPerRank[sortId] != nil {
		this.finalPerRank[sortId] = nil
	}
}

func (this *Tournament) getRank(sortId, round, snid int32, isFinals bool) int32 {
	if _, ok := this.rankPlayers[sortId]; ok {
		if rps, ok1 := this.rankPlayers[sortId][round]; ok1 {
			MatchContextSlice(rps).Sort(isFinals)
			for _, rp := range rps {
				if rp.p.SnId == snid {
					return rp.rank
				}
			}
		}
	}
	return 0
}

func (this *Tournament) isOut(sortId, snid int32) bool {
	out := false
	if this.finalPerRank[sortId] != nil {
		for _, info := range this.finalPerRank[sortId] {
			if info.SnId == snid {
				out = true
			}
		}
	}
	return out
}

func (this *Tournament) ModuleName() string {
	return "Tournament"
}

// 初始化
func (this *Tournament) Init() {
	logger.Logger.Trace("(this *Tournament) Init()")
}

func (this *Tournament) Update() {
	for tmId, v := range this.signUpPlayers {
		n := v.MaxCnt - len(v.signup)
		if n > 2 {
			n = rand.Intn(2) + 1
		} else if n > 0 {
			n = rand.Intn(n) + 1
		}
		if n > 0 && v.GameId != 0 {
			this.InviteRobot(tmId, v.GameId, v.Platform, n)
		}
	}
}

// 邀请Ai
func (this *Tournament) InviteRobot(tmId, gameId int32, platform string, num int) {
	pack := &server.WGInviteMatchRob{
		Platform:  proto.String(platform),
		MatchId:   proto.Int32(tmId),
		RobNum:    proto.Int(num),
		NeedAwait: proto.Bool(true),
	}
	SceneMgrSington.SendToGame(int(gameId), int(server.SSPacketID_PACKET_WG_INVITEMATCHROB), pack)
}
func (this *Tournament) Shutdown() {
}
func init() {
	module.RegisteModule(TournamentMgr, time.Second, 0)
}
