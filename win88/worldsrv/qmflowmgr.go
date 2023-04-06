package main

import (
	"games.yol.com/win88/model"
	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
	"time"

	"github.com/idealeak/goserver/core/module"
)

type QMFlowManager struct {
	BaseClockSinker
	playerStatement    map[int32]map[int32]*webapi.PlayerStatementSrc
	offPlayerStatement map[int32]map[int32]*webapi.PlayerStatementSrc
}

var QMFlowMgr = &QMFlowManager{}

func (this *QMFlowManager) ModuleName() string {
	return "QMFlowManager"
}

func (this *QMFlowManager) Init() {

}

func (this *QMFlowManager) Update() {
	this.Save()
}

func (this *QMFlowManager) Shutdown() {
	this.ForceSave()

	module.UnregisteModule(this)
}

// 感兴趣所有clock event
func (this *QMFlowManager) InterestClockEvent() int {
	return 1 << CLOCK_EVENT_HOUR
}

func (this *QMFlowManager) OnHourTimer() {
	this.AnsySave()
}

// 增加玩家流水统计
// 数据用途: 流水返利，全民代理佣金结算，确保参数计算无误
// totalWin: 总赢取钱(>=0)
// totalLost: 总输的钱(>=0)
func (this *QMFlowManager) AddPlayerStatement(snid, isBind, totalWin, totalLost, gamefreeID int32, platform,
	packageId string, dbGameFree *server_proto.DB_GameFree) {
	if model.GameParamData.CloseQMThr {
		return
	}

	if dbGameFree == nil {
		return
	}

	if totalWin == 0 && totalLost == 0 {
		return
	}
	/*
		if dbGameFree.GetPlayerWaterRate() == 0 {
			return
		}
	*/
	if this.playerStatement == nil {
		this.playerStatement = make(map[int32]map[int32]*webapi.PlayerStatementSrc)
	}
	var gpmap map[int32]*webapi.PlayerStatementSrc
	if gp, exist := this.playerStatement[gamefreeID]; exist {
		gpmap = gp
	} else {
		gpmap = make(map[int32]*webapi.PlayerStatementSrc)
	}

	if ps, exist := gpmap[snid]; exist {
		ps.TotalWin += float64(totalWin*dbGameFree.GetPlayerWaterRate()) / 100
		ps.TotalLose += float64(totalLost*dbGameFree.GetPlayerWaterRate()) / 100
		ps.TotaSrclWin += float64(totalWin)
		ps.TotaSrclLose += float64(totalLost)
	} else {
		ps := &webapi.PlayerStatementSrc{
			Platform:     platform,
			PackageId:    packageId,
			SnId:         snid,
			IsBind:       isBind,
			TotalWin:     float64(totalWin*dbGameFree.GetPlayerWaterRate()) / 100,
			TotalLose:    float64(totalLost*dbGameFree.GetPlayerWaterRate()) / 100,
			TotaSrclWin:  float64(totalWin),
			TotaSrclLose: float64(totalLost),
		}
		gpmap[snid] = ps
	}

	this.playerStatement[gamefreeID] = gpmap
}

// 增加玩家流水统计
// 数据用途: 流水返利，全民代理佣金结算，确保参数计算无误
// totalWin: 总赢取钱(>=0)
// totalLost: 总输的钱(>=0)
func (this *QMFlowManager) AddOffPlayerStatement(snid, totalWin, totalLost, gamefreeID int32, dbGameFree *server_proto.DB_GameFree) {
	if model.GameParamData.CloseQMThr {
		return
	}
	if dbGameFree == nil {
		return
	}
	if totalWin == 0 && totalLost == 0 {
		return
	}
	/*
		if dbGameFree.GetPlayerWaterRate() == 0 {
			return
		}
	*/
	if this.offPlayerStatement == nil {
		this.offPlayerStatement = make(map[int32]map[int32]*webapi.PlayerStatementSrc)
	}
	var gpmap map[int32]*webapi.PlayerStatementSrc
	if gp, exist := this.offPlayerStatement[gamefreeID]; exist {
		gpmap = gp
	} else {
		gpmap = make(map[int32]*webapi.PlayerStatementSrc)
	}

	if ps, exist := gpmap[snid]; exist {
		ps.TotalWin += float64(totalWin*dbGameFree.GetPlayerWaterRate()) / 100
		ps.TotalLose += float64(totalLost*dbGameFree.GetPlayerWaterRate()) / 100
		ps.TotaSrclWin += float64(totalWin)
		ps.TotaSrclLose += float64(totalLost)
	} else {
		ps := &webapi.PlayerStatementSrc{
			SnId:         snid,
			TotalWin:     float64(totalWin*dbGameFree.GetPlayerWaterRate()) / 100,
			TotalLose:    float64(totalLost*dbGameFree.GetPlayerWaterRate()) / 100,
			TotaSrclWin:  float64(totalWin),
			TotaSrclLose: float64(totalLost),
		}
		gpmap[snid] = ps
	}

	this.offPlayerStatement[gamefreeID] = gpmap
}

// 异步保存，用户每个小时的定时处理
func (this *QMFlowManager) AnsySave() {
	//先查找一下不在线的玩家，看是否在线,先保存一次
	tempSave := this.offPlayerStatement
	this.offPlayerStatement = nil

	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		for _, v := range tempSave {
			for m, n := range v {
				baseInfo := model.GetPlayerBaseInfo(n.Platform, m)
				if baseInfo != nil {
					n.Platform = baseInfo.Platform
					n.PackageId = baseInfo.PackageID
					isBind := int32(0)
					if baseInfo.Tel != "" {
						isBind = 1
					}
					n.IsBind = isBind
				}
			}
		}
		return nil
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		for k, v := range tempSave {
			for m, n := range v {
				if n.Platform != "" {
					isQuMin := false
					//pt := PlatformMgrSington.GetPackageTag(n.PackageId)
					//if pt != nil && pt.SpreadTag == 1 {
					//	isQuMin = true
					//}
					if !isQuMin && model.GameParamData.QMOptimization {
						//不是全民包，不发送对应的数据
						delete(v, m)
						continue
					}
					if this.playerStatement == nil {
						this.playerStatement = make(map[int32]map[int32]*webapi.PlayerStatementSrc)
					}
					var gpmap map[int32]*webapi.PlayerStatementSrc
					if gp, exist := this.playerStatement[k]; exist {
						gpmap = gp
					} else {
						gpmap = make(map[int32]*webapi.PlayerStatementSrc)
					}

					if ps, exist := gpmap[m]; exist {
						ps.TotalWin += n.TotalWin
						ps.TotalLose += n.TotalLose
						ps.TotaSrclWin += n.TotaSrclWin
						ps.TotaSrclLose += n.TotaSrclLose
					} else {
						ps := &webapi.PlayerStatementSrc{
							SnId:         n.SnId,
							TotalWin:     n.TotalWin,
							TotalLose:    n.TotalLose,
							TotaSrclWin:  n.TotaSrclWin,
							TotaSrclLose: n.TotaSrclLose,
							PackageId:    n.PackageId,
							Platform:     n.Platform,
							IsBind:       n.IsBind,
						}
						gpmap[m] = ps
					}

					this.playerStatement[k] = gpmap
					delete(v, m)
				} else {
					//没有找到对应的用户信息，可能是数据库出现问题，重新插入数据
					logger.Logger.Errorf("save player qm err:%v,%v,%v,%v,%v,%v", n.SnId, n.TotaSrclLose, n.TotaSrclWin,
						n.TotalLose, n.TotalWin, m)
					if this.offPlayerStatement == nil {
						this.offPlayerStatement = make(map[int32]map[int32]*webapi.PlayerStatementSrc)
					}
					var gpmap map[int32]*webapi.PlayerStatementSrc
					if gp, exist := this.offPlayerStatement[k]; exist {
						gpmap = gp
					} else {
						gpmap = make(map[int32]*webapi.PlayerStatementSrc)
					}

					if ps, exist := gpmap[m]; exist {
						ps.TotalWin += n.TotalWin
						ps.TotalLose += n.TotalLose
						ps.TotaSrclWin += n.TotaSrclWin
						ps.TotaSrclLose += n.TotaSrclLose
					} else {
						ps := &webapi.PlayerStatementSrc{
							SnId:         n.SnId,
							TotalWin:     n.TotalWin,
							TotalLose:    n.TotalLose,
							TotaSrclWin:  n.TotaSrclWin,
							TotaSrclLose: n.TotaSrclLose,
							PackageId:    n.PackageId,
							Platform:     n.Platform,
							IsBind:       n.IsBind,
						}
						gpmap[m] = ps
					}

					this.offPlayerStatement[k] = gpmap
				}
			}
		}
	}), "QMFlowManagerAnsySave").Start()
}

// 强制保存，在关闭时调用
func (this *QMFlowManager) ForceSave() {
	//先查找一下不在线的玩家，看是否在线,先保存一次
	tempSave := this.offPlayerStatement
	this.offPlayerStatement = nil
	for _, v := range tempSave {
		for m, n := range v {
			baseInfo := model.GetPlayerBaseInfo(n.Platform, m)
			if baseInfo != nil {
				n.Platform = baseInfo.Platform
				n.PackageId = baseInfo.PackageID
				isBind := int32(0)
				if baseInfo.Tel != "" {
					isBind = 1
				}
				n.IsBind = isBind
			}
		}
	}

	for k, v := range tempSave {
		for m, n := range v {
			if n.Platform != "" {
				isQuMin := false
				//pt := PlatformMgrSington.GetPackageTag(n.PackageId)
				//if pt != nil && pt.SpreadTag == 1 {
				//	isQuMin = true
				//}
				if !isQuMin && model.GameParamData.QMOptimization {
					//不是全民包，不发送对应的数据
					delete(v, m)
					continue
				}
				if this.playerStatement == nil {
					this.playerStatement = make(map[int32]map[int32]*webapi.PlayerStatementSrc)
				}
				var gpmap map[int32]*webapi.PlayerStatementSrc
				if gp, exist := this.playerStatement[k]; exist {
					gpmap = gp
				} else {
					gpmap = make(map[int32]*webapi.PlayerStatementSrc)
				}

				if ps, exist := gpmap[m]; exist {
					ps.TotalWin += n.TotalWin
					ps.TotalLose += n.TotalLose
					ps.TotaSrclWin += n.TotaSrclWin
					ps.TotaSrclLose += n.TotaSrclLose
				} else {
					ps := &webapi.PlayerStatementSrc{
						SnId:         n.SnId,
						TotalWin:     n.TotalWin,
						TotalLose:    n.TotalLose,
						TotaSrclWin:  n.TotaSrclWin,
						TotaSrclLose: n.TotaSrclLose,
						PackageId:    n.PackageId,
						Platform:     n.Platform,
						IsBind:       n.IsBind,
					}
					gpmap[m] = ps
				}

				this.playerStatement[k] = gpmap
				delete(v, m)
			} else {
				//没有找到对应的用户信息，可能是数据库出现问题，重新插入数据
				logger.Logger.Errorf("save player qm err:%v,%v,%v,%v,%v,%v", n.SnId, n.TotaSrclLose, n.TotaSrclWin,
					n.TotalLose, n.TotalWin, m)
			}
		}
	}

	logger.Logger.Tracef("ready save flow record:%v", len(this.playerStatement))
	//for gamefreeId, info := range this.playerStatement {
	//	var datas []*webapi.PlayerStatement
	//
	//	for _, ps := range info {
	//
	//		tps := &webapi.PlayerStatement{
	//			Platform:     ps.Platform,
	//			IsBind:       ps.IsBind,
	//			PackageId:    ps.PackageId,
	//			TotalWin:     int32(ps.TotalWin),
	//			TotalLose:    int32(ps.TotalLose),
	//			TotaSrclLose: int32(ps.TotaSrclLose),
	//			TotaSrclWin:  int32(ps.TotaSrclWin),
	//			SnId:         ps.SnId,
	//		}
	//		datas = append(datas, tps)
	//	}
	//
	//	QPT := int(model.GameParamData.SpreadAccountQPT)
	//	//每X条一次，避免GET请求头超长
	//	for len(datas) >= QPT {
	//		d := datas[:QPT]
	//		LogChannelSington.WriteMQData(model.GenerateSpreadAccount(common.GetAppId(), gamefreeId, d))
	//	}
	//
	//	if len(datas) != 0 {
	//		d := datas[:]
	//		LogChannelSington.WriteMQData(model.GenerateSpreadAccount(common.GetAppId(), gamefreeId, d))
	//	}
	//}
	logger.Logger.Tracef("save all qmflow ok")
	this.playerStatement = nil
}

func (this *QMFlowManager) Save() {
	//先查找一下不在线的玩家，看是否在线,先保存一次
	for k, v := range this.offPlayerStatement {
		for m, n := range v {
			player := PlayerMgrSington.GetPlayerBySnId(m)
			if player != nil {
				isQuMin := false
				//pt := PlatformMgrSington.GetPackageTag(player.PackageID)
				//if pt != nil && pt.SpreadTag == 1 {
				//	isQuMin = true
				//}
				if !isQuMin && model.GameParamData.QMOptimization {
					//不是全民包，不发送对应的数据
					delete(v, m)
					continue
				}

				isBind := int32(0)
				if player.Tel != "" {
					isBind = 1
				}
				if this.playerStatement == nil {
					this.playerStatement = make(map[int32]map[int32]*webapi.PlayerStatementSrc)
				}
				var gpmap map[int32]*webapi.PlayerStatementSrc
				if gp, exist := this.playerStatement[k]; exist {
					gpmap = gp
				} else {
					gpmap = make(map[int32]*webapi.PlayerStatementSrc)
				}

				if ps, exist := gpmap[m]; exist {
					ps.TotalWin += n.TotalWin
					ps.TotalLose += n.TotalLose
					ps.TotaSrclWin += n.TotaSrclWin
					ps.TotaSrclLose += n.TotaSrclLose
				} else {
					ps := &webapi.PlayerStatementSrc{
						SnId:         n.SnId,
						TotalWin:     n.TotalWin,
						TotalLose:    n.TotalLose,
						TotaSrclWin:  n.TotaSrclWin,
						TotaSrclLose: n.TotaSrclLose,
						PackageId:    player.PackageID,
						Platform:     player.Platform,
						IsBind:       isBind,
					}
					gpmap[m] = ps
				}

				this.playerStatement[k] = gpmap
				delete(v, m)
			}
		}
	}

	//for gamefreeId, info := range this.playerStatement {
	//	var datas []*webapi.PlayerStatement
	//
	//	for _, ps := range info {
	//
	//		tps := &webapi.PlayerStatement{
	//			Platform:     ps.Platform,
	//			IsBind:       ps.IsBind,
	//			PackageId:    ps.PackageId,
	//			TotalWin:     int32(ps.TotalWin),
	//			TotalLose:    int32(ps.TotalLose),
	//			TotaSrclLose: int32(ps.TotaSrclLose),
	//			TotaSrclWin:  int32(ps.TotaSrclWin),
	//			SnId:         ps.SnId,
	//		}
	//		datas = append(datas, tps)
	//
	//	}
	//
	//	QPT := int(model.GameParamData.SpreadAccountQPT)
	//	//每X条一次，避免GET请求头超长
	//	for len(datas) >= QPT {
	//		d := datas[:QPT]
	//		LogChannelSington.WriteMQData(model.GenerateSpreadAccount(common.GetAppId(), gamefreeId, d))
	//		datas = datas[QPT:]
	//	}
	//
	//	//把剩下的记录在推一遍
	//	if len(datas) != 0 {
	//		d := datas[:]
	//		LogChannelSington.WriteMQData(model.GenerateSpreadAccount(common.GetAppId(), gamefreeId, d))
	//	}
	//}

	this.playerStatement = nil
}

func init() {
	module.RegisteModule(QMFlowMgr, time.Minute*3, 0)
	ClockMgrSington.RegisteSinker(QMFlowMgr)
}
