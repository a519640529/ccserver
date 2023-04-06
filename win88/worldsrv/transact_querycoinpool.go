package main

import (
	webapi2 "games.yol.com/win88/protocol/webapi"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/netlib"
	"time"

	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
)

var QueryCoinPoolTimeOut = time.Second * 30

const (
	QueryCoinPoolTransactParam_ParentNode int = iota
	QueryCoinPoolTransactParam_Data
)

type QueryCoinPoolTransactHandler struct {
}

func (this *QueryCoinPoolTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	//logger.Logger.Trace("QueryCoinPoolTransactHandler.OnExcute ")
	if data, ok := ud.(*common.W2GQueryCoinPool); ok {
		var ids []int32
		var gameType int32
		if data.GroupId != 0 {
			pgg := PlatformGameGroupMgrSington.GetGameGroup(data.GroupId)
			if pgg != nil {
				ids = append(ids, pgg.DbGameFree.Id)
				gameType = pgg.DbGameFree.GetGameType()
			}
		} else {
			ids, gameType = srvdata.DataMgr.GetGameFreeIds(data.GameId, data.GameMode)
		}
		for sid, gs := range GameSessMgrSington.servers {
			if gameType == common.GameType_Mini {
				//迷你类型 并且所需要的gameId在gameIds里边
				if len(gs.gameIds) > 0 && common.InSliceInt32(gs.gameIds, data.GameId) {
					for _, id := range ids {
						gs.DetectCoinPoolSetting(data.Platform, id, data.GroupId)
					}
					tnp := &transact.TransNodeParam{
						Tt:     common.TransType_QueryCoinPool,
						Ot:     transact.TransOwnerType(srvlib.GameServerType),
						Oid:    sid,
						AreaID: common.GetSelfAreaId(),
						Tct:    transact.TransactCommitPolicy_SelfDecide,
					}
					tNode.StartChildTrans(tnp, ud, QueryCoinPoolTimeOut)
				}
			} else if len(gs.gameIds) == 0 {
				//其他类型 gameIds通配
				for _, id := range ids {
					gs.DetectCoinPoolSetting(data.Platform, id, data.GroupId)
				}
				tnp := &transact.TransNodeParam{
					Tt:     common.TransType_QueryCoinPool,
					Ot:     transact.TransOwnerType(srvlib.GameServerType),
					Oid:    sid,
					AreaID: common.GetSelfAreaId(),
					Tct:    transact.TransactCommitPolicy_SelfDecide,
				}
				tNode.StartChildTrans(tnp, ud, QueryCoinPoolTimeOut)
			}
		}
	}
	return transact.TransExeResult_Success
}

func (this *QueryCoinPoolTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	//logger.Logger.Trace("QueryCoinPoolTransactHandler.OnCommit ")
	field := tNode.TransEnv.GetField(QueryCoinPoolTransactParam_Data)
	parent := tNode.TransEnv.GetField(QueryCoinPoolTransactParam_ParentNode)
	if tParent, ok := parent.(*transact.TransNode); ok {
		queryGamePool := &webapi2.SAQueryGamePoolByGameId{
			Tag:             webapi2.TagCode_SUCCESS,
			CoinPoolSetting: field.([]*webapi2.CoinPoolSetting),
		}
		tParent.TransRep.RetFiels = queryGamePool
		tParent.Resume()
	}
	return transact.TransExeResult_Success
}

func (this *QueryCoinPoolTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	//logger.Logger.Trace("QueryCoinPoolTransactHandler.OnRollBack ")
	return transact.TransExeResult_Success
}

func (this *QueryCoinPoolTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	//logger.Logger.Trace("QueryCoinPoolTransactHandler.OnChildTransRep ")
	if ud != nil {
		var userData []*webapi2.CoinPoolSetting
		err := netlib.UnmarshalPacketNoPackId(ud.([]byte), &userData)
		if err == nil {
			field := tNode.TransEnv.GetField(QueryCoinPoolTransactParam_Data)
			if field == nil {
				tNode.TransEnv.SetField(QueryCoinPoolTransactParam_Data, userData)
			} else {
				if arr, ok := field.([]*webapi2.CoinPoolSetting); ok {
					arr = append(arr, userData...)
					tNode.TransEnv.SetField(QueryCoinPoolTransactParam_Data, arr)
				}
			}
		} else {
			logger.Logger.Trace("trascate.OnChildRespWrapper err:", err)
		}
	}

	return transact.TransExeResult_Success
}

func StartQueryCoinPoolTransact(tParent *transact.TransNode, gameid, gamemode int32, platform string, groupId int32) {
	tnp := &transact.TransNodeParam{
		Tt:     common.TransType_QueryCoinPool,
		Ot:     transact.TransOwnerType(common.GetSelfSrvType()),
		Oid:    common.GetSelfSrvId(),
		AreaID: common.GetSelfAreaId(),
	}
	ud := &common.W2GQueryCoinPool{
		GameId:   gameid,
		GameMode: gamemode,
		Platform: platform,
		GroupId:  groupId,
	}
	tNode := transact.DTCModule.StartTrans(tnp, ud, QueryCoinPoolTimeOut)
	if tNode != nil {
		tNode.TransEnv.SetField(QueryCoinPoolTransactParam_ParentNode, tParent)
		tNode.Go(core.CoreObject())
	}
}

func init() {
	transact.RegisteHandler(common.TransType_QueryCoinPool, &QueryCoinPoolTransactHandler{})
}
