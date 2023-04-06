package main

import (
	"container/list"
	"math/rand"
	"strconv"
	"time"

	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/core/utils"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	server_proto "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	srvproto "github.com/idealeak/goserver/srvlib/protocol"
)

type PlayerStatics struct {
	PlayerCnt       int
	RobotCnt        int
	PlayerGamingCnt int
	RobotGamingCnt  int
	GamingCnt       map[int]int
}

type PlayerPendingData struct {
	sid int64
	ts  int64
}

var PlayerMgrSington = &PlayerMgr{
	playerMap:        make(map[int64]*Player),
	playerSnMap:      make(map[int32]*Player),
	playerAccountMap: make(map[string]*Player),
	playerTokenMap:   make(map[string]*Player),
	playerUnionIdMap: make(map[string]*Player),
	playerLoading:    make(map[string]*PlayerPendingData),
	players:          make([]*Player, 0, 1024),
	playerOfPlatform: make(map[string]map[int32]*Player),
}

type PlayerMgr struct {
	BaseClockSinker
	playerMap        map[int64]*Player
	playerSnMap      map[int32]*Player
	playerAccountMap map[string]*Player
	playerTokenMap   map[string]*Player
	playerUnionIdMap map[string]*Player
	playerLoading    map[string]*PlayerPendingData
	players          []*Player
	playerOfPlatform map[string]map[int32]*Player
}

func (this *PlayerMgr) GetOnlineCount() *PlayerStatics {
	ps := &PlayerStatics{
		GamingCnt: make(map[int]int),
	}
	for _, player := range this.playerMap {
		if player != nil && player.IsOnLine() {
			if !player.IsRob {
				ps.PlayerCnt++
				if player.scene != nil {
					ps.PlayerGamingCnt++
				}
			} else {
				ps.RobotCnt++
				if player.scene != nil {
					ps.RobotGamingCnt++
				}
			}
			if player.scene != nil {
				ps.GamingCnt[player.scene.gameId] = ps.GamingCnt[player.scene.gameId] + 1
			}
		}
	}
	return ps
}

func (this *PlayerMgr) Exist(id int64) bool {
	_, ok := this.playerMap[id]
	return ok
}

func (this *PlayerMgr) IsOnline(snId int32) bool {
	player, ok := this.playerSnMap[snId]
	if ok {
		return player.IsOnLine()
	} else {
		return false
	}
}

func (this *PlayerMgr) ManagePlayer(player *Player) {
	if old, ok := this.playerMap[player.sid]; ok && old != nil {
		logger.Logger.Warnf("(this *PlayerMgr) ManagePlayer [playerMap] found sid=%v player exist snid=%v, mysnid=%v", player.sid, old.SnId, player.SnId)
	}
	this.playerMap[player.sid] = player
	if old, ok := this.playerSnMap[player.SnId]; ok && old != nil {
		logger.Logger.Warnf("(this *PlayerMgr) ManagePlayer [playerSnMap] found player exist snid=%v, mysnid=%v", old.SnId, player.SnId)
	}
	this.playerSnMap[player.SnId] = player
	if old, ok := this.playerAccountMap[player.AccountId]; ok && old != nil {
		logger.Logger.Warnf("(this *PlayerMgr) ManagePlayer [playerAccountMap] found player exist snid=%v, mysnid=%v", old.SnId, player.SnId)
	}
	this.playerAccountMap[player.AccountId] = player
	if player.customerToken != "" {
		if old, ok := this.playerTokenMap[player.customerToken]; ok && old != nil {
			logger.Logger.Warnf("(this *PlayerMgr) ManagePlayer [playerTokenMap] found player exist snid=%v, mysnid=%v", old.SnId, player.SnId)
		}
		this.playerTokenMap[player.customerToken] = player
	}

	//if player.UnionId != "" {
	//	if old, ok := this.playerUnionIdMap[player.UnionId]; ok && old != nil {
	//		logger.Logger.Warnf("(this *PlayerMgr) ManagePlayer [playerUnionIdMap] found player exist snid=%v, mysnid=%v", old.SnId, player.SnId)
	//	}
	//	this.playerUnionIdMap[player.UnionId] = player
	//}

	logger.Logger.Tracef("###%v mount to DBSaver[ManagePlayer]", player.Name)
	if !player.IsRob {
		var found bool
		for i, p := range this.players {
			if p.SnId == player.SnId {
				found = true
				logger.Logger.Warnf("(this *PlayerMgr) ManagePlayer [this.players] found player exist snid=%v", player.SnId)
				this.players[i] = player
				break
			}
		}
		if !found {
			this.players = append(this.players, player)
		}
		//平台玩家管理器
		if pp, exist := this.playerOfPlatform[player.Platform]; exist {
			pp[player.SnId] = player
		} else {
			pp = make(map[int32]*Player)
			pp[player.SnId] = player
			this.playerOfPlatform[player.Platform] = pp
		}
	}
	DbSaver_Inst.RegisterDbSaverTask(player)
}

func (this *PlayerMgr) AddPlayer(id int64, playerInfo *model.PlayerData, s *netlib.Session) bool {
	player := NewPlayer(id, playerInfo, s)
	if player == nil {
		return false
	}

	if id == 0 {
		logger.Logger.Warnf("(this *PlayerMgr) AddPlayer player id == 0:")
		return false
	}

	logger.Logger.Trace("(this *PlayerMgr) AddPlayer Set player ip:", player.Ip)
	this.playerMap[id] = player
	var oldp *Player
	if p, exist := this.playerSnMap[player.SnId]; exist {
		oldp = p
	}
	this.playerSnMap[player.SnId] = player
	this.playerAccountMap[player.AccountId] = player
	if player.customerToken != "" {
		this.playerTokenMap[player.customerToken] = player
	}
	//if player.UnionId != "" {
	//	this.playerUnionIdMap[player.UnionId] = player
	//}
	if !player.IsRob {
		var found bool
		for i, p := range this.players {
			if p.SnId == player.SnId {
				found = true
				logger.Logger.Warnf("(this *PlayerMgr) AddPlayer [this.players] found player exist snid=%v", player.SnId)
				this.players[i] = player
				break
			}
		}
		if !found {
			this.players = append(this.players, player)
		}

		//平台玩家管理器
		if pp, exist := this.playerOfPlatform[player.Platform]; exist {
			pp[player.SnId] = player
		} else {
			pp = make(map[int32]*Player)
			pp[player.SnId] = player
			this.playerOfPlatform[player.Platform] = pp
		}

		logger.Logger.Tracef("###%v mount to DBSaver[AddPlayer]", player.Name)
		if oldp != nil { //删除旧的玩家
			DbSaver_Inst.UnregisteDbSaveTask(oldp)
		}
		DbSaver_Inst.RegisterDbSaverTask(player)
		niceIdMgr.NiceIdCheck(player.SnId)
	} else {
		player.NiceId = niceIdMgr.PopNiceId(player.SnId)
	}
	return true
}

func (this *PlayerMgr) DelPlayer(snid int32) bool {
	if player, ok := this.playerSnMap[snid]; ok && player != nil {
		logger.Logger.Infof("(this *PlayerMgr) DelPlayer(%v)", snid)
		if player != nil {
			player.OnLogouted()
		}
		//if player.UnionId != "" {
		//	delete(this.playerUnionIdMap, player.UnionId)
		//}
		if player.sid != 0 {
			delete(this.playerMap, player.sid)
		}
		delete(this.playerSnMap, player.SnId)
		delete(this.playerAccountMap, player.AccountId)
		if player.customerToken != "" {
			delete(this.playerTokenMap, player.customerToken)
		}
		if player != nil && !player.IsRob {
			index := -1
			for i, p := range this.players {
				if p.SnId == snid {
					index = i
					break
				}
			}
			if index != -1 {
				count := len(this.players)
				if index == 0 {
					this.players = this.players[1:]
				} else if index == count-1 {
					this.players = this.players[:count-1]
				} else {
					arr := this.players[:index]
					arr = append(arr, this.players[index+1:]...)
					this.players = arr
				}
			}

			//平台玩家管理器
			if pp, exist := this.playerOfPlatform[player.Platform]; exist {
				delete(pp, player.SnId)
			}

			//从时间片上摘除掉
			logger.Logger.Tracef("###%v unmount from DBSaver[DelPlayer]", player.Name)
			DbSaver_Inst.UnregisteDbSaveTask(player)
		}
		if player.IsRob {
			niceIdMgr.PushNiceId(player.NiceId)
		}
		return true
	}
	if bag, ok := BagMgrSington.PlayerBag[snid]; ok && bag != nil { // 清空背包
		delete(BagMgrSington.PlayerBag, snid)
	}
	return false
}

func (this *PlayerMgr) DroplinePlayer(p *Player) {
	delete(this.playerMap, p.sid)
}

func (this *PlayerMgr) ReholdPlayer(p *Player, newSid int64, newSess *netlib.Session) {
	if p.sid != 0 {
		delete(this.playerMap, p.sid)
	}

	if newSid == 0 {
		logger.Logger.Errorf("(this *PlayerMgr) ReholdPlayer(snid=%v, new=%v)", p.SnId, newSid)
	}

	p.sid = newSid
	p.gateSess = newSess
	p.state = PlayerState_Online
	this.playerMap[newSid] = p
}

func (this *PlayerMgr) GetPlayer(id int64) *Player {
	if pi, ok := this.playerMap[id]; ok {
		return pi
	}
	return nil
}

func (this *PlayerMgr) GetPlayerBySnId(id int32) *Player {
	if p, ok := this.playerSnMap[id]; ok {
		return p
	}
	return nil
}

// 批量取出玩家信息
func (this *PlayerMgr) GetPlayersBySnIds(ids []int32) []*Player {
	var retPlayers []*Player
	for _, v := range ids {
		if p, ok := this.playerSnMap[v]; ok {
			retPlayers = append(retPlayers, p)
		}
	}
	return retPlayers
}

func (this *PlayerMgr) GetPlayerByAccount(acc string) *Player {
	if p, ok := this.playerAccountMap[acc]; ok {
		return p
	}
	return nil
}

func (this *PlayerMgr) GetPlayerByToken(token string) *Player {
	if p, ok := this.playerTokenMap[token]; ok {
		return p
	}
	return nil
}

func (this *PlayerMgr) GetPlayerByUnionId(unionId string) *Player {
	if p, ok := this.playerUnionIdMap[unionId]; ok {
		return p
	}
	return nil
}

func (this *PlayerMgr) UpdatePlayerToken(p *Player, newToken string) {
	oldToken := p.customerToken
	if oldToken != newToken {
		if oldToken != "" {
			if _, ok := this.playerTokenMap[oldToken]; ok {
				delete(this.playerTokenMap, oldToken)
			}
		}
		if newToken != "" {
			this.playerTokenMap[newToken] = p
			p.customerToken = newToken
		}
	}
}

func (this *PlayerMgr) BroadcastMessage(packetid int, rawpack interface{}) bool {
	sc := &srvproto.BCSessionUnion{
		Bccs: &srvproto.BCClientSession{},
	}
	pack, err := BroadcastMaker.CreateBroadcastPacket(sc, packetid, rawpack)
	if err == nil && pack != nil {
		srvlib.ServerSessionMgrSington.Broadcast(int(srvproto.SrvlibPacketID_PACKET_SS_BROADCAST), pack, common.GetSelfAreaId(), srvlib.GateServerType)
		return true
	}
	return false
}

func (this *PlayerMgr) BroadcastMessageToPlatform(platform string, packetid int, rawpack interface{}) {
	if platform == "" {
		this.BroadcastMessage(packetid, rawpack)
	} else {
		players := this.playerOfPlatform[platform]
		mgs := make(map[*netlib.Session][]*srvproto.MCSessionUnion)
		for _, p := range players {
			if p != nil && p.gateSess != nil && p.IsOnLine() /*&& p.Platform == platform*/ {
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
}

func (this *PlayerMgr) BroadcastMessageToPlatformWithHall(platform string, snid int32, packetid int, rawpack interface{}) {
	if platform == "" {
		this.BroadcastMessage(packetid, rawpack)
	} else {
		player := this.GetPlayerBySnId(snid)
		if player != nil {
			players := this.playerOfPlatform[platform]
			mgs := make(map[*netlib.Session][]*srvproto.MCSessionUnion)
			for _, p := range players {
				if p != nil && p.gateSess != nil && p.IsOnLine() && p.scene == nil {
					if FriendMgrSington.IsShield(p.SnId, snid) {
						continue
					}
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
	}
}

func (this *PlayerMgr) BroadcastMessageToGroup(packetid int, rawpack interface{}, tags []string) bool {
	pack := &server_proto.SSCustomTagMulticast{
		Tags: tags,
	}
	if byteData, ok := rawpack.([]byte); ok {
		pack.RawData = byteData
	} else {
		byteData, err := netlib.MarshalPacket(packetid, rawpack)
		if err == nil {
			pack.RawData = byteData
		} else {
			logger.Logger.Info("PlayerMgr.BroadcastMessageToGroup err:", err)
			return false
		}
	}
	srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_SS_CUSTOMTAG_MULTICAST), pack, common.GetSelfAreaId(), srvlib.GateServerType)
	return true
}

func (this *PlayerMgr) BroadcastMessageToTarget(platform string, target []int32, packetid int, rawpack interface{}) {
	players := this.playerOfPlatform[platform]
	mgs := make(map[*netlib.Session][]*srvproto.MCSessionUnion)
	for _, p := range players {
		if p != nil && p.gateSess != nil && p.IsOnLine() /*&& p.Platform == platform*/ {
			if common.InSliceInt32(target, p.SnId) {
				mgs[p.gateSess] = append(mgs[p.gateSess], &srvproto.MCSessionUnion{
					Mccs: &srvproto.MCClientSession{
						SId: proto.Int64(p.sid),
					},
				})
			}
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

// 感兴趣所有clock event
func (this *PlayerMgr) InterestClockEvent() int {
	return (1 << CLOCK_EVENT_MAX) - 1
}

func (this *PlayerMgr) OnSecTimer() {
	for _, player := range this.players {
		player.OnSecTimer()
	}
}

func (this *PlayerMgr) OnMiniTimer() {
	for _, player := range this.players {
		utils.CatchPanic(func() {
			player.OnMiniTimer()
		})
	}
}

func (this *PlayerMgr) OnHourTimer() {
	for _, player := range this.players {
		utils.CatchPanic(func() {
			player.OnHourTimer()
		})
	}
}

func (this *PlayerMgr) OnDayTimer() {
	for _, player := range this.players {
		utils.CatchPanic(func() {
			player.OnDayTimer(false, true, 1)
		})
	}
}

func (this *PlayerMgr) OnMonthTimer() {
	for _, player := range this.players {
		utils.CatchPanic(func() {
			player.OnMonthTimer()
		})
	}
}

func (this *PlayerMgr) OnWeekTimer() {
	for _, player := range this.players {
		utils.CatchPanic(func() {
			player.OnWeekTimer()
		})
	}
}

func (this *PlayerMgr) OnShutdown() {
	this.SaveAll()
}

func (this *PlayerMgr) SaveAll() {
	count := len(this.players)
	start := time.Now()
	saveCnt := 0
	failCnt := 0
	nochangeCnt := 0
	logger.Logger.Info("===@PlayerMgr.SaveAll BEG@=== TotalCount:", count)
	for i, p := range this.players {
		idx := i + 1
		if p.dirty {
			if model.SavePlayerData(p.PlayerData) {
				logger.Logger.Infof("===@SavePlayerData %v/%v snid:%v coin:%v safebox:%v coinpayts:%v safeboxts:%v gamets:%v save [ok] @=", idx, count, p.SnId, p.Coin, p.SafeBoxCoin, p.CoinPayTs, p.SafeBoxCoinTs, p.GameCoinTs)
				saveCnt++
			} else {
				logger.Logger.Warnf("===@SavePlayerData %v/%v snid:%v coin:%v safebox:%v coinpayts:%v safeboxts:%v gamets:%v save [error]@=", idx, count, p.SnId, p.Coin, p.SafeBoxCoin, p.CoinPayTs, p.SafeBoxCoinTs, p.GameCoinTs)
				failCnt++
			}
		} else {
			logger.Logger.Infof("===@SavePlayerData %v/%v snid:%v coin:%v safebox:%v coinpayts:%v safeboxts:%v gamets:%v nochange [ok]@=", idx, count, p.SnId, p.Coin, p.SafeBoxCoin, p.CoinPayTs, p.SafeBoxCoinTs, p.GameCoinTs)
			nochangeCnt++
		}
	}
	logger.Logger.Infof("===@PlayerMgr.SaveAll END@===, total:%v saveCnt:%v failCnt:%v nochangeCnt:%v take:%v", count, saveCnt, failCnt, nochangeCnt, time.Now().Sub(start))
}

func (this *PlayerMgr) RebindSnId(oldSnId, newSnId int32) {
	if player, exist := this.playerSnMap[oldSnId]; exist {
		if _, exist := this.playerSnMap[newSnId]; !exist {
			delete(this.playerSnMap, oldSnId)
			this.playerSnMap[newSnId] = player
			player.SnId = newSnId
		}
	}

	oldCode := strconv.Itoa(int(oldSnId))
	newCode := strconv.Itoa(int(newSnId))
	for _, p := range this.playerMap {
		if p.currClubId == oldSnId {
			p.currClubId = newSnId
		}

		if p.BeUnderAgentCode == oldCode {
			p.BeUnderAgentCode = newCode
			p.dirty = true
		}
	}
	this.SaveAll()
}

// 黑名单事件
func (this *PlayerMgr) OnAddBlackInfo(blackinfo *BlackInfo) {
	//nothing
	//if blackinfo.Snid > 0 {
	//	if p := this.GetPlayerBySnId(blackinfo.Snid); p != nil {
	//		p.PlayerData.BlacklistType = int32(blackinfo.BlackType)
	//		p.dirty = true
	//		p.Time2Save()
	//	} else {
	//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//			model.UpdatePlayerBlacklistType(blackinfo.Platform, blackinfo.Snid, int32(blackinfo.BlackType))
	//			return nil
	//		}), nil, "PlayerMgrOnAddBlackInfo").Start()
	//	}
	//}
}

func (this *PlayerMgr) OnEditBlackInfo(blackinfo *BlackInfo) {
	//nothing
	//if blackinfo.Snid > 0 {
	//	if p := this.GetPlayerBySnId(blackinfo.Snid); p != nil {
	//		p.PlayerData.BlacklistType = int32(blackinfo.BlackType)
	//		p.dirty = true
	//		p.Time2Save()
	//	} else {
	//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//			model.UpdatePlayerBlacklistType(blackinfo.Platform, blackinfo.Snid, int32(blackinfo.BlackType))
	//			return nil
	//		}), nil, "PlayerMgrOnEditBlackInfo").Start()
	//	}
	//}
}

func (this *PlayerMgr) OnRemoveBlackInfo(blackinfo *BlackInfo) {
	//nothing
	//if blackinfo.Snid > 0 {
	//	if p := this.GetPlayerBySnId(blackinfo.Snid); p != nil {
	//		p.PlayerData.BlacklistType = 0
	//		p.dirty = true
	//		p.Time2Save()
	//	} else {
	//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//			model.UpdatePlayerBlacklistType(blackinfo.Platform, blackinfo.Snid, int32(0))
	//			return nil
	//		}), nil, "PlayerMgrOnRemoveBlackInfo").Start()
	//	}
	//}
}

func (this *PlayerMgr) KickoutByPlatform(name string) {
	for _, p := range this.players {
		if name == "" || p.Platform == name {
			p.Kickout(common.KickReason_Disconnection)
		}
	}
}

func (this *PlayerMgr) UpdateAllPlayerPackageTag(packageTag, platform, channel, promoter string, promoterTree, tagkey int32) int {
	var cnt int
	for _, p := range this.players {
		if p != nil && !p.IsRob {
			if p.PackageID == packageTag {
				p.Platform = platform
				p.Channel = channel
				p.BeUnderAgentCode = promoter
				p.PromoterTree = promoterTree
				//p.TagKey = tagkey
				p.dirty = true
				cnt++
			}
		}
	}
	return cnt
}

func (this *PlayerMgr) LoadRobots() {
	if model.GameParamData.PreLoadRobotCount > 0 {
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			tsBeg := time.Now()
			robots := model.GetRobotPlayers(model.GameParamData.PreLoadRobotCount)
			tsEnd := time.Now()
			logger.Logger.Tracef("GetRobotPlayers take:%v total:%v", tsEnd.Sub(tsBeg), len(robots))
			return robots
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if robots, ok := data.([]*model.PlayerData); ok {
				if robots != nil {
					for i := 0; i < len(robots); i++ {
						if this.GetPlayerBySnId(robots[i].SnId) == nil {
							player := NewPlayer(0, robots[i], nil)
							if player != nil {
								this.playerSnMap[player.SnId] = player
								this.playerAccountMap[player.AccountId] = player
								if player.customerToken != "" {
									this.playerTokenMap[player.customerToken] = player
								}
								//if player.UnionId != "" {
								//	this.playerUnionIdMap[player.UnionId] = player
								//}
							}
						}
					}
				}
			}
		}), "GetRobotPlayers").Start()
	}
}

func (this *PlayerMgr) StartLoading(accid string, sid int64) bool {
	ts := time.Now().Unix()
	if d, exist := this.playerLoading[accid]; exist {
		d.sid = sid
		if ts-d.ts > 300 {
			d.ts = ts
			return false
		}
		return true
	}
	this.playerLoading[accid] = &PlayerPendingData{sid: sid, ts: ts}
	return false
}

func (this *PlayerMgr) EndPlayerLoading(accid string) int64 {
	if d, exist := this.playerLoading[accid]; exist {
		delete(this.playerLoading, accid)
		return d.sid
	}
	return 0
}

func PlayerRankGe(p1, p2 *Player, n int) bool {
	switch n {
	case 0:
		if p1.TotalCoin == p2.TotalCoin {
			return p1.SnId < p2.SnId
		} else {
			return p1.TotalCoin > p2.TotalCoin
		}
	case 1:
		if p1.CoinPayTotal == p2.CoinPayTotal {
			return p1.SnId < p2.SnId
		} else {
			return p1.CoinPayTotal > p2.CoinPayTotal
		}
	case 2:
		if p1.CoinExchangeTotal == p2.CoinExchangeTotal {
			return p1.SnId < p2.SnId
		} else {
			return p1.CoinExchangeTotal > p2.CoinExchangeTotal
		}
	case 3:
		a := p1.Coin + p1.SafeBoxCoin - p1.CoinPayTotal
		b := p2.Coin + p2.SafeBoxCoin - p2.CoinPayTotal
		if a == b {
			return p1.SnId < p2.SnId
		} else {
			return a > b
		}

	}
	return false
}

//	func (this *PlayerMgr) GetRank() map[string][]*model.Rank {
//		ret := make(map[string][]*model.Rank)
//		ls := make(map[string]*list.List)
//
//		platforms := PlatformMgrSington.Platforms
//		for p := range platforms {
//			ret[p] = make([]*model.Rank, 0)
//			ls[p] = list.New()
//		}
//
//		for _, player := range this.players {
//			if player.IsRob {
//				continue
//			}
//
//			p := player.PlayerData.Platform
//			if _, ok := platforms[p]; !ok {
//				continue
//			}
//
//			l := ls[p]
//			for n := l.Front(); n != nil; n = n.Next() {
//				if np, ok := n.Value.(*Player); ok {
//					if PlayerRankGe(player, np) {
//						l.InsertBefore(player, n)
//						goto CHECK
//					}
//				}
//				//else {
//				//	logger.Logger.Warnf("PlayerMgr.GetRank n.Value.(*Player) fail")
//				//	continue
//				//}
//			}
//
//			l.PushBack(player)
//		CHECK:
//			if l.Len() > model.MAX_RANK_COUNT {
//				l.Remove(l.Back())
//			}
//		}
//
//		for p := range platforms {
//			l := ls[p]
//			for n := l.Front(); n != nil; n = n.Next() {
//				if np, ok := n.Value.(*Player); ok {
//					ret[p] = append(ret[p], &model.Rank{
//						SnId:      np.PlayerData.SnId,
//						Name:      np.PlayerData.Name,
//						Head:      np.PlayerData.Head,
//						VIP:       np.PlayerData.VIP,
//						TotalCoin: np.PlayerData.TotalCoin,
//					})
//				}
//			}
//		}
//
//		return ret
//	}
func (this *PlayerMgr) GetAssetRank(platform string) []*model.Rank {
	ret := make([]*model.Rank, 0, model.MAX_RANK_COUNT)
	l := list.New()

	for _, player := range this.players {
		if player.IsRob {
			continue
		}

		if player.PlayerData.Platform != platform {
			continue
		}

		for n := l.Front(); n != nil; n = n.Next() {
			if np, ok := n.Value.(*Player); ok {
				if PlayerRankGe(player, np, 0) {
					l.InsertBefore(player, n)
					goto CHECK
				}
			}
		}

		l.PushBack(player)
	CHECK:
		if l.Len() > model.MAX_RANK_COUNT {
			l.Remove(l.Back())
		}
	}

	for n := l.Front(); n != nil; n = n.Next() {
		if np, ok := n.Value.(*Player); ok {
			ret = append(ret, &model.Rank{
				SnId:      np.PlayerData.SnId,
				Name:      np.PlayerData.Name,
				Head:      np.PlayerData.Head,
				VIP:       np.PlayerData.VIP,
				TotalCoin: np.PlayerData.TotalCoin,
			})
		}
	}

	return ret
}
func (this *PlayerMgr) GetRechargeLists(platform string) []*model.Rank {
	ret := make([]*model.Rank, 0, model.MAX_RANK_COUNT)
	l := list.New()

	for _, player := range this.players {
		if player.IsRob {
			continue
		}

		if player.PlayerData.Platform != platform {
			continue
		}

		for n := l.Front(); n != nil; n = n.Next() {
			if np, ok := n.Value.(*Player); ok {
				if PlayerRankGe(player, np, 1) {
					l.InsertBefore(player, n)
					goto CHECK
				}
			}
		}

		l.PushBack(player)
	CHECK:
		if l.Len() > model.MAX_RANK_COUNT {
			l.Remove(l.Back())
		}
	}

	for n := l.Front(); n != nil; n = n.Next() {
		if np, ok := n.Value.(*Player); ok {
			ret = append(ret, &model.Rank{
				SnId:      np.PlayerData.SnId,
				Name:      np.PlayerData.Name,
				Head:      np.PlayerData.Head,
				VIP:       np.PlayerData.VIP,
				TotalCoin: np.PlayerData.CoinPayTotal,
			})
		}
	}

	return ret
}
func (this *PlayerMgr) GetExchangeLists(platform string) []*model.Rank {
	ret := make([]*model.Rank, 0, model.MAX_RANK_COUNT)
	l := list.New()

	for _, player := range this.players {
		if player.IsRob {
			continue
		}

		if player.PlayerData.Platform != platform {
			continue
		}

		for n := l.Front(); n != nil; n = n.Next() {
			if np, ok := n.Value.(*Player); ok {
				if PlayerRankGe(player, np, 2) {
					l.InsertBefore(player, n)
					goto CHECK
				}
			}
		}

		l.PushBack(player)
	CHECK:
		if l.Len() > model.MAX_RANK_COUNT {
			l.Remove(l.Back())
		}
	}

	for n := l.Front(); n != nil; n = n.Next() {
		if np, ok := n.Value.(*Player); ok {
			ret = append(ret, &model.Rank{
				SnId:      np.PlayerData.SnId,
				Name:      np.PlayerData.Name,
				Head:      np.PlayerData.Head,
				VIP:       np.PlayerData.VIP,
				TotalCoin: np.PlayerData.CoinExchangeTotal,
			})
		}
	}

	return ret
}
func (this *PlayerMgr) GetProfitLists(platform string) []*model.Rank {
	ret := make([]*model.Rank, 0, model.MAX_RANK_COUNT)
	l := list.New()

	for _, player := range this.players {
		if player.IsRob {
			continue
		}

		if player.PlayerData.Platform != platform {
			continue
		}

		for n := l.Front(); n != nil; n = n.Next() {
			if np, ok := n.Value.(*Player); ok {
				if PlayerRankGe(player, np, 3) {
					l.InsertBefore(player, n)
					goto CHECK
				}
			}
		}

		l.PushBack(player)
	CHECK:
		if l.Len() > model.MAX_RANK_COUNT {
			l.Remove(l.Back())
		}
	}

	for n := l.Front(); n != nil; n = n.Next() {
		if np, ok := n.Value.(*Player); ok {
			ret = append(ret, &model.Rank{
				SnId: np.PlayerData.SnId,
				Name: np.PlayerData.Name,
				Head: np.PlayerData.Head,
				VIP:  np.PlayerData.VIP,
				//TotalCoin: np.PlayerData.ProfitCoin,
			})
		}
	}

	return ret
}

func (this *PlayerMgr) DeletePlayerByPlatform(platform string) {
	var dels []*Player
	for _, p := range this.players {
		if p != nil && p.Platform == platform {
			p.Kickout(common.KickReason_Disconnection)
			dels = append(dels, p)
		}
	}

	for _, p := range dels {
		if p != nil {
			p.isDelete = true
			if p.scene == nil {
				this.DelPlayer(p.SnId)
			}
		}
	}
}

func (this *PlayerMgr) StatsOnline() model.PlayerOLStats {
	stats := model.PlayerOLStats{
		PlatformStats: make(map[string]*model.PlayerStats),
		RobotStats: model.PlayerStats{
			InGameCnt: make(map[int32]map[int32]int32),
		},
	}

	for _, p := range this.playerMap {
		if p != nil {
			if p.IsRob {
				pps := &stats.RobotStats
				if pps != nil {
					if p.scene == nil {
						pps.InHallCnt++
					} else {
						if g, exist := pps.InGameCnt[int32(p.scene.gameId)]; exist {
							g[p.scene.dbGameFree.GetId()]++
						} else {
							g := make(map[int32]int32)
							pps.InGameCnt[int32(p.scene.gameId)] = g
							g[p.scene.dbGameFree.GetId()]++
						}
					}
				}
			} else {
				var pps *model.PlayerStats
				var exist bool
				if pps, exist = stats.PlatformStats[p.Platform]; !exist {
					pps = &model.PlayerStats{InGameCnt: make(map[int32]map[int32]int32)}
					stats.PlatformStats[p.Platform] = pps
				}

				if pps != nil {
					if p.scene == nil {
						pps.InHallCnt++
					} else {
						if g, exist := pps.InGameCnt[int32(p.scene.gameId)]; exist {
							g[p.scene.dbGameFree.GetId()]++
						} else {
							g := make(map[int32]int32)
							pps.InGameCnt[int32(p.scene.gameId)] = g
							g[p.scene.dbGameFree.GetId()]++
						}
					}
				}
			}
		}
	}
	return stats
}

func (p *PlayerMgr) UpdateName(snId int32, name string) {
	player := p.GetPlayerBySnId(snId)
	if player == nil {
		return
	}
	player.setName(name)
	player.dirty = true
}

func (p *PlayerMgr) UpdateHead(snId, head int32) {
	player := p.GetPlayerBySnId(snId)
	if player == nil {
		return
	}
	player.Head = head
	//0:男 1:女
	player.Sex = (player.Head%2 + 1) % 2
	player.dirty = true
	player.changeIconTime = time.Now()
}

func (p *PlayerMgr) UpdateHeadOutline(snId, outline int32) {
	player := p.GetPlayerBySnId(snId)
	if player == nil {
		return
	}
	player.HeadOutLine = outline
	player.dirty = true
}

func (p *PlayerMgr) ModifyActSwitchToPlayer(platform string, modify bool) {
	if modify { //活动开关修改了才去更新活动开关
		if players, ok := p.playerOfPlatform[platform]; ok {
			for _, p := range players {
				if p != nil && !p.IsRob {
					p.ModifyActSwitch()
				}
			}
		}
	}
}

/*
推荐好友规则
1.优先判断在线玩家人数N

	（1）N≥20；每次刷新，从在线玩家中随机6个
	（2）N＜20；则填充机器人，保证N=20，每次填充的机器人头像和昵称随机；然后从N中随机6个

2.刷新有CD（暂定20s），刷新过后进入cd
*/
type RecommendFriend struct {
	Snid int32
	Name string
	Head int32
}

func (this *PlayerMgr) RecommendFriendRule(platform string, snid int32) []RecommendFriend {
	if platform == "" {
		return nil
	} else {
		rets := []RecommendFriend{}
		players := this.playerOfPlatform[platform]
		for _, player := range players { //优先真人
			if player.SnId != snid && !FriendMgrSington.IsFriend(snid, player.SnId) {
				ret := RecommendFriend{
					Snid: player.SnId,
					Name: player.Name,
					Head: player.Head,
				}
				rets = append(rets, ret)
				if len(rets) >= 20 {
					break
				}
			}
		}
		if len(rets) < 20 {
			for _, player := range this.playerSnMap { //其次机器人
				if player.IsRob {
					ret := RecommendFriend{
						Snid: player.SnId,
						Name: player.Name,
						Head: player.Head,
					}
					rets = append(rets, ret)
					if len(rets) >= 20 {
						break
					}
				}
			}
		}
		//if len(rets) < 20 { //假数据
		//	needNum := 20 - len(rets)
		//	for i := 0; i < needNum; i++ {
		//		name := "贵宾"
		//		if rand.Int31n(100) < 60 {
		//			pool := srvdata.PBDB_NameMgr.Datas.GetArr()
		//			cnt := int32(len(pool))
		//			if cnt > 0 {
		//				name = pool[rand.Int31n(cnt)].GetName()
		//			}
		//		}
		//		ret := RecommendFriend{
		//			Snid: 99999999,
		//			Name: name,
		//			Head: rand.Int31n(6) + 1,
		//		}
		//		rets = append(rets, ret)
		//	}
		//}
		needIdxs := []int{}
		if rets != nil {
			if len(rets) >= 6 {
				for {
					if len(needIdxs) >= 6 {
						break
					}
					randIdx := rand.Intn(len(rets))
					if !common.InSliceInt(needIdxs, randIdx) {
						needIdxs = append(needIdxs, randIdx)
					}
				}
			} else {
				for i := 0; i < len(rets); i++ {
					needIdxs = append(needIdxs, i)
				}
			}
		}
		ret := []RecommendFriend{}
		for _, idx := range needIdxs {
			ret = append(ret, rets[idx])
		}
		return ret
	}
}

func init() {
	//BlackListMgrSington.RegisteObserver(PlayerMgrSington)
	PlayerSubjectSign.AttachName(PlayerMgrSington)
	PlayerSubjectSign.AttachHead(PlayerMgrSington)
	PlayerSubjectSign.AttachHeadOutline(PlayerMgrSington)
	ClockMgrSington.RegisteSinker(PlayerMgrSington)
}
