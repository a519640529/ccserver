package base

import (
	server_proto "games.yol.com/win88/protocol/server"
	"time"

	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	"github.com/idealeak/goserver/srvlib/protocol"
)

var PlayerMgrSington = &PlayerMgr{
	playerMap:        make(map[int64]*Player),
	playerSnMap:      make(map[int32]*Player),
	playerAccountMap: make(map[string]*Player),
	//	playerNameMap:    make(map[string]*Player),
}

type PlayerMgr struct {
	playerMap        map[int64]*Player
	playerSnMap      map[int32]*Player
	playerAccountMap map[string]*Player
	//	playerNameMap    map[string]*Player
}

func (this *PlayerMgr) Exist(id int64) bool {
	_, ok := this.playerMap[id]
	return ok
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
	//	if old, ok := this.playerNameMap[player.Name]; ok && old != nil {
	//		logger.Logger.Warnf("(this *PlayerMgr) ManagePlayer [playerNameMap] found player exist snid=%v, mysnid=%v", old.SnId, player.SnId)
	//	}
	//	this.playerNameMap[player.Name] = player
}

func (this *PlayerMgr) AddPlayer(id int64, data []byte, ws, gs *netlib.Session) *Player {
	oldPlayer := this.GetPlayer(id)
	testFlag := false
	if oldPlayer != nil {
		logger.Logger.Warnf("(this *PlayerMgr) AddPlayer found id=%v player exist snid=%v", id, oldPlayer.SnId)
		testFlag = true
		if oldPlayer.scene != nil {
			logger.Logger.Warnf("(this *PlayerMgr) AddPlayer found snid=%v in sceneid=%v", id, oldPlayer.SnId, oldPlayer.scene.SceneId)
			if SceneMgrSington.GetScene(oldPlayer.scene.SceneId) != nil {
				logger.Logger.Warnf("(this *PlayerMgr) AddPlayer found snid=%v in sceneid=%v SceneMgrSington.GetScene(oldPlayer.scene.sceneId) != nil", id, oldPlayer.SnId, oldPlayer.scene.SceneId)
			}
		}
		this.DelPlayer(id)
	}
	player := NewPlayer(id, data, ws, gs)
	if player == nil {
		logger.Logger.Warn("(this *PlayerMgr) NewPlayer player == nil")
		return nil
	}
	if testFlag == true {
		logger.Logger.Warnf("(this *PlayerMgr) AddPlayer new snid=%v", player.SnId)
	}
	//logger.Logger.Trace("(this *PlayerMgr) NewPlayer player = ", player)
	this.playerMap[id] = player
	this.playerSnMap[player.SnId] = player
	this.playerAccountMap[player.AccountId] = player
	//	this.playerNameMap[player.Name] = player
	logger.Logger.Tracef("(this *PlayerMgr) AddPlayer snid:%v ", player.SnId)
	return player
}

func (this *PlayerMgr) AddLocalPlayer(snid int32) *Player {
	oldPlayer := this.GetPlayerBySnId(snid)
	if oldPlayer != nil {
		logger.Logger.Warnf("(this *PlayerMgr) AddLocalPlayer found snid=%v player exist", oldPlayer.SnId)
		if oldPlayer.scene != nil {
			logger.Logger.Warnf("(this *PlayerMgr) AddLocalPlayer found snid=%v in sceneid=%v", oldPlayer.SnId, oldPlayer.scene.SceneId)
			if SceneMgrSington.GetScene(oldPlayer.scene.SceneId) != nil {
				logger.Logger.Warnf("(this *PlayerMgr) AddLocalPlayer found snid=%v in sceneid=%v SceneMgrSington.GetScene(oldPlayer.scene.sceneId) != nil", oldPlayer.SnId, oldPlayer.scene.SceneId)
			}
		}
		this.DelPlayer(oldPlayer.sid)
	}
	player := NewLocalPlayer(snid)
	if player == nil {
		logger.Logger.Warn("(this *PlayerMgr) NewLocalPlayer player == nil")
		return nil
	}
	this.playerSnMap[player.SnId] = player
	this.playerAccountMap[player.AccountId] = player
	logger.Logger.Tracef("(this *PlayerMgr) AddLocalPlayer snid:%v ", player.SnId)
	return player
}

func (this *PlayerMgr) DelPlayer(id int64) bool {
	player := this.GetPlayer(id)
	if player != nil {
		delete(this.playerMap, id)
		delete(this.playerSnMap, player.SnId)
		delete(this.playerAccountMap, player.AccountId)
		//		delete(this.playerNameMap, player.Name)
		logger.Logger.Tracef("(this *PlayerMgr) DelPlayer snid:%v ", player.SnId)
		return true
	}
	return false
}

func (this *PlayerMgr) DeletePlayers(ids ...int64) {
	for _, sid := range ids {
		if p, ok := this.playerMap[sid]; ok {
			delete(this.playerMap, sid)
			if p != nil {
				delete(this.playerSnMap, p.SnId)
				delete(this.playerAccountMap, p.AccountId)
				//				delete(this.playerNameMap, p.Name)
			}
		}
	}
}

func (this *PlayerMgr) DelPlayerBySnId(snid int32) bool {
	player := this.GetPlayerBySnId(snid)
	if player != nil {
		delete(this.playerMap, player.sid)
		delete(this.playerSnMap, player.SnId)
		delete(this.playerAccountMap, player.AccountId)
		//		delete(this.playerNameMap, player.Name)
		logger.Logger.Tracef("(this *PlayerMgr) DelPlayerBySnId snid:%v ", player.SnId)
		return true
	}
	return false
}

func (this *PlayerMgr) ReholdPlayer(oldSid, newSid int64, newSess *netlib.Session) {
	p := this.GetPlayer(oldSid)
	if p != nil {
		delete(this.playerMap, oldSid)
		this.playerMap[newSid] = p
	}
}

func (this *PlayerMgr) GetPlayer(id int64) *Player {
	if pi, ok := this.playerMap[id]; ok {
		return pi
	}
	return nil
}

func (this *PlayerMgr) GetPlayerBySnId(id int32) *Player {
	if pi, ok := this.playerSnMap[id]; ok {
		return pi
	}
	return nil
}

func (this *PlayerMgr) GetPlayerByAccount(acc string) *Player {
	if p, ok := this.playerAccountMap[acc]; ok {
		return p
	}
	return nil
}

//func (this *PlayerMgr) GetPlayerByName(name string) *Player {
//	if p, ok := this.playerNameMap[name]; ok {
//		return p
//	}
//	return nil
//}

func (this *PlayerMgr) BroadcastMessage(packetid int, rawpack interface{}) bool {
	sc := &protocol.BCSessionUnion{
		Bccs: &protocol.BCClientSession{},
	}
	pack, err := BroadcastMaker.CreateBroadcastPacket(sc, packetid, rawpack)
	if err == nil && pack != nil {
		srvlib.ServerSessionMgrSington.Broadcast(int(protocol.SrvlibPacketID_PACKET_SS_BROADCAST), pack, common.GetSelfAreaId(), srvlib.GateServerType)
		return true
	}
	return false
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

func (this *PlayerMgr) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.playerSnMap[oldSnId]; exist {
		delete(this.playerSnMap, oldSnId)
		this.playerSnMap[newSnId] = p
	}
}

////////////////////////////////////////////////////////////////////
/// Module Implement [beg]
////////////////////////////////////////////////////////////////////
func (this *PlayerMgr) ModuleName() string {
	return "playermgr"
}

func (this *PlayerMgr) Init() {
}

func (this *PlayerMgr) Update() {
}

func (this *PlayerMgr) Shutdown() {
	for _, p := range this.playerMap {
		if p.scene != nil {
			p.scene.Pause()
			//			if p.dirty {
			//				data, err := p.MarshalData(p.scene.gameId)
			//				if err == nil {
			//					msgSync := &protocol2.GWPlayerDataSync{
			//						Sid:        proto.Int64(p.sid),
			//						PlayerData: data,
			//						IParams:    p.MarshalIParam(),
			//						SParams:    p.MarshalSParam(),
			//					}
			//					proto.SetDefaults(msgSync)
			//					p.worldSess.Send(int(protocol2.MmoPacketID_PACKET_GW_PLAYERDATASYNC), msgSync)
			//				}
			//			}
		}
	}
	module.UnregisteModule(this)
}
func (this *PlayerMgr) OnDayTimer() {
	//在线跨天 数据给昨天，今天置为空
	for _, p := range this.playerMap {
		if p != nil && !p.IsRob {
			p.OnDayTimer()
		}
	}
}

////////////////////////////////////////////////////////////////////
/// Module Implement [end]
////////////////////////////////////////////////////////////////////

func init() {
	module.RegisteModule(PlayerMgrSington, time.Second, 0)
}
