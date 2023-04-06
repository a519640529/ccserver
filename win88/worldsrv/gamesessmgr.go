package main

import (
	"games.yol.com/win88/protocol/webapi"
	"math"
	"strconv"
	"strings"

	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	"github.com/idealeak/goserver/srvlib/protocol"
)

const (
	ReplayServerType int = 8
	ReplayServerId       = 801
)

var GameSessMgrSington = &GameSessMgr{
	servers:  make(map[int]*GameSession),
	gamesrvs: make(map[int][]*GameSession),
	gates:    make(map[int]*GameSession),
}

type GameSessMgr struct {
	servers  map[int]*GameSession
	gamesrvs map[int][]*GameSession
	gates    map[int]*GameSession
}

// 注册事件
func (this *GameSessMgr) OnRegiste(s *netlib.Session) {
	attr := s.GetAttribute(srvlib.SessionAttributeServerInfo)
	if attr != nil {
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			if srvInfo.GetType() == srvlib.GameServiceType {
				logger.Logger.Warn("(this *GameSessMgr) OnRegiste (GameSrv):", s)
				srvId := int(srvInfo.GetId())
				gs := NewGameSession(srvId, srvlib.GameServerType, s)
				if gs != nil {
					this.servers[srvId] = gs
					data := srvInfo.GetData()
					if data != "" {
						gameids := strings.Split(data, ",")
						for _, id := range gameids {
							if gameid, err := strconv.Atoi(id); err == nil {
								if gss, exist := this.gamesrvs[gameid]; exist {
									gss = append(gss, gs)
									this.gamesrvs[gameid] = gss
								} else {
									this.gamesrvs[gameid] = []*GameSession{gs}
								}
								gs.gameIds = append(gs.gameIds, int32(gameid))
							}
						}
					} else {
						if gss, exist := this.gamesrvs[0]; exist {
							gss = append(gss, gs)
							this.gamesrvs[0] = gss
						} else {
							this.gamesrvs[0] = []*GameSession{gs}
						}
						//gs.gameIds = append(gs.gameIds, 0)
					}
					gs.OnRegiste()
					//尝试创建百人场
					HundredSceneMgrSington.TryCreateRoom()
				}
			} else if srvInfo.GetType() == srvlib.GateServiceType {
				logger.Logger.Warn("(this *GameSessMgr) OnRegiste (GateSrv):", s)
				srvId := int(srvInfo.GetId())
				gs := NewGameSession(srvId, srvlib.GateServerType, s)
				if gs != nil {
					this.gates[srvId] = gs
				}

			}
		}
	}
}

// 注销事件
func (this *GameSessMgr) OnUnregiste(s *netlib.Session) {
	attr := s.GetAttribute(srvlib.SessionAttributeServerInfo)
	if attr != nil {
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			if srvInfo.GetType() == srvlib.GameServiceType {
				logger.Logger.Warn("(this *GameSessMgr) OnUnregiste (GameSrv):", s)
				srvId := int(srvInfo.GetId())
				gs := this.servers[srvId]
				if gs != nil {
					delete(this.servers, srvId)
					data := srvInfo.GetData()
					if data != "" {
						gameids := strings.Split(data, ",")
						for _, id := range gameids {
							if gameid, err := strconv.Atoi(id); err == nil {
								if gss, exist := this.gamesrvs[gameid]; exist {
									cnt := len(gss)
									for j := 0; j < cnt; j++ {
										if gss[j] == gs {
											gss[j] = gss[cnt-1]
											gss = gss[:cnt-1]
											this.gamesrvs[gameid] = gss
											break
										}
									}
								}
							}
						}
					} else {
						if gss, exist := this.gamesrvs[0]; exist {
							cnt := len(gss)
							for j := 0; j < cnt; j++ {
								if gss[j] == gs {
									gss[j] = gss[cnt-1]
									gss = gss[:cnt-1]
									this.gamesrvs[0] = gss
									break
								}
							}
						}
					}
					gs.OnUnregiste()
				}
			} else if srvInfo.GetType() == srvlib.GateServiceType {
				logger.Logger.Warn("(this *GameSessMgr) OnUnregiste (GateSrv):", s)
				LoginStateMgrSington.LogoutAllBySession(s)
				srvId := int(srvInfo.GetId())
				delete(this.gates, srvId)
			}
		}
	}
}
func (this *GameSessMgr) GetGameServerSess(gameid int) []*GameSession {
	return this.gamesrvs[gameid]
}

// 获取最小负载的GameSession
func (this *GameSessMgr) GetMinLoadSess(gameid int) *GameSession {
	minLoad := math.MaxInt32
	loadFactor := 0
	var gs *GameSession
	if gss, exist := this.gamesrvs[gameid]; exist {
		if gss != nil {
			for _, s := range gss {
				if s.state == common.GAME_SESS_STATE_ON {
					loadFactor = s.GetLoadFactor()
					if minLoad > loadFactor {
						minLoad = loadFactor
						gs = s
					}
				}
			}
			if gs != nil {
				return gs
			}
		}
	}
	if gss, exist := this.gamesrvs[0]; exist {
		if gss != nil {
			for _, s := range gss {
				if s.state == common.GAME_SESS_STATE_ON {
					loadFactor = s.GetLoadFactor()
					if minLoad > loadFactor {
						minLoad = loadFactor
						gs = s
					}
				}
			}
			if gs != nil {
				return gs
			}
		}
	}
	return gs
}

func (this *GameSessMgr) GetGameSess(srvId int) *GameSession {
	if gs, exist := this.servers[srvId]; exist {
		return gs
	}
	return nil
}
func (this *GameSessMgr) GetAllGameSess() []*GameSession {
	servers := make([]*GameSession, 0)
	for _, v := range this.servers {
		servers = append(servers, v)
	}
	return servers
}
func (this *GameSessMgr) GetGateSess(srvId int) *GameSession {
	if gs, exist := this.gates[srvId]; exist {
		return gs
	}
	return nil
}
func (this *GameSessMgr) RebindPlayerSnId(oldSnId, newSnId int32) {
	for _, gs := range this.servers {
		gs.RebindPlayerSnId(oldSnId, newSnId)
	}
}

func (this *GameSessMgr) ListServerState(srvId, srvType int) []*webapi.ServerInfo {
	_createGateServerInfo := func(srvId, srvType int, s *GameSession) *webapi.ServerInfo {
		var robNum, playerNum int
		for _, p := range PlayerMgrSington.playerMap {
			if p != nil && p.gateSess == s.Session {
				if p.IsRob {
					robNum++
				} else {
					playerNum++
				}
			}
		}
		si := &webapi.ServerInfo{
			SrvId:     int32(srvId),
			SrvType:   int32(srvType),
			State:     int32(s.state),
			PlayerNum: int32(playerNum),
			RobotNum:  int32(robNum),
			SceneNum:  0,
		}
		attr := s.GetAttribute(srvlib.SessionAttributeServerInfo)
		if attr != nil {
			if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
				si.Data = srvInfo.GetData()
			}
		}
		return si
	}

	_createGameServerInfo := func(srvId, srvType int, s *GameSession) *webapi.ServerInfo {
		var playerNum int
		for _, p := range s.players {
			if !p.IsRob {
				playerNum++
			}
		}
		si := &webapi.ServerInfo{
			SrvId:     int32(srvId),
			SrvType:   int32(srvType),
			State:     int32(s.state),
			PlayerNum: int32(playerNum),
			RobotNum:  int32(len(s.players) - playerNum),
			SceneNum:  int32(len(s.scenes)),
		}
		attr := s.GetAttribute(srvlib.SessionAttributeServerInfo)
		if attr != nil {
			if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
				si.Data = srvInfo.GetData()
			}
		}
		return si
	}
	var datas []*webapi.ServerInfo
	if srvType != 0 {
		switch srvType {
		case srvlib.GateServiceType:
			if srvId != 0 {
				if s, exist := this.gates[srvId]; exist {
					si := _createGateServerInfo(srvId, srvlib.GateServiceType, s)
					datas = append(datas, si)
				}
			}
		case srvlib.GameServiceType:
			if srvId != 0 {
				if s, exist := this.servers[srvId]; exist {
					si := _createGameServerInfo(srvId, srvlib.GameServiceType, s)
					datas = append(datas, si)
				}
			}
		}
	} else {
		if srvId == 0 {
			for sid, s := range this.gates {
				si := _createGateServerInfo(sid, srvlib.GateServiceType, s)
				datas = append(datas, si)
			}
			for sid, s := range this.servers {
				si := _createGameServerInfo(sid, srvlib.GameServiceType, s)
				datas = append(datas, si)
			}
		} else {
			if s, exist := this.gates[srvId]; exist {
				si := _createGateServerInfo(srvId, srvlib.GateServiceType, s)
				datas = append(datas, si)
			}
			if s, exist := this.servers[srvId]; exist {
				si := _createGameServerInfo(srvId, srvlib.GameServiceType, s)
				datas = append(datas, si)
			}
		}
	}

	//worldsrv 自身的信息
	myInfo := &webapi.ServerInfo{
		SrvId:     int32(common.GetSelfSrvId()),
		SrvType:   int32(common.GetSelfSrvType()),
		State:     int32(common.GAME_SESS_STATE_ON),
		PlayerNum: int32(len(PlayerMgrSington.players)),
		RobotNum:  int32(len(PlayerMgrSington.playerSnMap) - len(PlayerMgrSington.players)),
		SceneNum:  int32(len(SceneMgrSington.scenes)),
	}
	if SrvIsMaintaining {
		myInfo.State = int32(common.GAME_SESS_STATE_ON)
	} else {
		myInfo.State = int32(common.GAME_SESS_STATE_OFF)
	}
	//把自己加进去
	datas = append(datas, myInfo)
	return datas
}

func init() {
	srvlib.ServerSessionMgrSington.AddListener(GameSessMgrSington)
}
