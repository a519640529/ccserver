package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
	"sort"
	"strconv"
	"time"
)

var QueryAllCoinPoolTimeOut = time.Second * 30

const (
	QueryAllCoinPoolTransactParam_ParentNode int = iota
	QueryAllCoinPoolTransactParam_Data
	QueryAllCoinPoolTransactParam_Total
)

type QueryAllCoinPoolTransactHandler struct {
}

func (this *QueryAllCoinPoolTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	//logger.Logger.Trace("QueryAllCoinPoolTransactHandler.OnExcute ")
	for sid, gs := range GameSessMgrSington.servers {
		if gs.srvType == srvlib.GameServerType {
			tnp := &transact.TransNodeParam{
				Tt:     common.TransType_QueryAllCoinPool,
				Ot:     transact.TransOwnerType(srvlib.GameServerType),
				Oid:    sid,
				AreaID: common.GetSelfAreaId(),
				Tct:    transact.TransactCommitPolicy_SelfDecide,
			}
			tNode.StartChildTrans(tnp, ud, QueryAllCoinPoolTimeOut)
		}
	}

	return transact.TransExeResult_Success
}

func (this *QueryAllCoinPoolTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	//logger.Logger.Trace("QueryAllCoinPoolTransactHandler.OnCommit ")
	field := tNode.TransEnv.GetField(QueryAllCoinPoolTransactParam_Data)
	parent := tNode.TransEnv.GetField(QueryAllCoinPoolTransactParam_ParentNode)
	Total := tNode.TransEnv.GetField(QueryAllCoinPoolTransactParam_Total)
	if tParent, ok := parent.(*transact.TransNode); ok {
		resp := webapi.NewResponseBody()
		resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
		if settings, ok := field.(map[string]*common.PlatformStates); ok {
			//根据平台id排序
			//map排序
			keys := []int{}
			for k := range settings {
				id, err := strconv.Atoi(k)
				if err == nil {
					keys = append(keys, id)
				}
			}
			sort.Ints(keys)
			info := []*common.PlatformStates{}
			for _, id := range keys {
				info = append(info, settings[strconv.Itoa(id)])
			}
			resp[webapi.RESPONSE_DATA] = info
		}
		resp[webapi.RESPONSE_TOTAL] = Total
		dataResp := &common.M2GWebApiResponse{}
		dataResp.Body, _ = resp.Marshal()
		tParent.TransRep.RetFiels = dataResp
		tParent.Resume()
	}
	return transact.TransExeResult_Success
}

func (this *QueryAllCoinPoolTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	//logger.Logger.Trace("QueryAllCoinPoolTransactHandler.OnRollBack ")
	return transact.TransExeResult_Success
}
func (this *QueryAllCoinPoolTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	//logger.Logger.Trace("QueryAllCoinPoolTransactHandler.OnChildTransRep ")
	if ud != nil {
		settings := make(map[string]*common.PlatformStates)
		err := netlib.UnmarshalPacketNoPackId(ud.([]byte), &settings)
		if err == nil {
			field := tNode.TransEnv.GetField(QueryAllCoinPoolTransactParam_Data)
			if field == nil {
				tNode.TransEnv.SetField(QueryAllCoinPoolTransactParam_Data, settings)
			} else {
				if arr, ok := field.(map[string]*common.PlatformStates); ok {
					//arr是第一个game返回的值
					for k, newpf := range settings {
						if oldpf, ok := arr[k]; ok {
							for m, newg := range newpf.GamesVal {
								if oldg, gok := oldpf.GamesVal[m]; gok {
									if newg.States > oldg.States {
										oldg.States = newg.States
										oldg.CoinValue = newg.CoinValue
										oldg.LowerLimit = newg.LowerLimit
										oldg.UpperLimit = newg.UpperLimit

									}
								} else {
									//找不到游戏
									arr[k].GamesVal[m] = newg
								}
							}
						} else {
							//找不到pf
							arr[k] = newpf
						}
					}
					tNode.TransEnv.SetField(QueryAllCoinPoolTransactParam_Data, arr)
				}
			}
		}
	}

	return transact.TransExeResult_Success
}

func StartQueryCoinPoolStatesTransact(tParent *transact.TransNode, pageNo, pageSize int32) {
	tnp := &transact.TransNodeParam{
		Tt:     common.TransType_QueryAllCoinPool,
		Ot:     transact.TransOwnerType(common.GetSelfSrvType()),
		Oid:    common.GetSelfSrvId(),
		AreaID: common.GetSelfAreaId(),
	}

	Platforms := PlatformMgrSington.Platforms

	//map排序
	var keys []int
	for k, v := range Platforms {
		//平台是否开启
		if !v.Disable {
			id, err := strconv.Atoi(k)
			if err == nil {
				keys = append(keys, id)
			}
		}
	}
	sort.Ints(keys)

	if pageNo <= 0 {
		pageNo = 1
	}
	end := pageNo * pageSize
	start := end - pageSize
	n := int32(len(keys))
	if start > n || end > n {
		start = 0
		end = 100
	}
	if end > n {
		end = n
	}
	//当前页所有的平台id
	NeedPlatforms := keys[start:end]
	games := make(map[string][]*common.GamesIndex)
	for _, platform := range NeedPlatforms {
		//获取当前所有平台下的gameid
		pf := strconv.Itoa(platform)
		//加载配置
		gps := PlatformMgrSington.GetPlatformGameConfig(pf)
		for _, v := range gps {
			//获取所有开启的游戏
			if v.Status && v.DbGameFree.Id%10 != 4 {
				g := &common.GamesIndex{}
				g.GameFreeId = v.DbGameFree.Id
				g.GroupId = v.GroupId
				games[pf] = append(games[pf], g)
			}
		}
	}
	ud := &common.QueryGames{
		Index: games,
	}
	tNode := transact.DTCModule.StartTrans(tnp, ud, QueryAllCoinPoolTimeOut)
	if tNode != nil {
		tNode.TransEnv.SetField(QueryAllCoinPoolTransactParam_ParentNode, tParent)
		tNode.TransEnv.SetField(QueryAllCoinPoolTransactParam_Total, n)
		tNode.Go(core.CoreObject())
	}
}

func init() {
	transact.RegisteHandler(common.TransType_QueryAllCoinPool, &QueryAllCoinPoolTransactHandler{})
}
