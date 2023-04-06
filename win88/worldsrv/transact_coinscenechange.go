package main

import (
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
)

var CoinSceneChangeTimeOut = time.Second * 10
var CoinSceneChangeTransactParam int

type CoinSceneChangeCtx struct {
	id         int32
	snid       int32
	sceneId    int32
	exclude    []int32
	isAudience bool
	roomId     string
	isClub     bool
	clubId     int32
}

type CoinSceneChangeTransactHandler struct {
}

func (this *CoinSceneChangeTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("CoinSceneChangeTransactHandler.OnExcute ")
	if ctx, ok := ud.(*CoinSceneChangeCtx); ok {
		player := PlayerMgrSington.GetPlayerBySnId(ctx.snid)
		if player != nil && player.scene != nil {
			tnp := &transact.TransNodeParam{
				Tt:     common.TransType_CoinSceneChange,
				Ot:     transact.TransOwnerType(srvlib.GameServerType),
				Oid:    int(player.scene.gameSess.GetSrvId()),
				AreaID: common.GetSelfAreaId(),
				Tct:    transact.TransactCommitPolicy_SelfDecide,
			}
			pack := &common.WGCoinSceneChange{
				SnId:    ctx.snid,
				SceneId: ctx.sceneId,
			}
			tNode.TransEnv.SetField(CoinSceneChangeTransactParam, ud)
			tNode.StartChildTrans(tnp, pack, CoinSceneChangeTimeOut)
		}
	}

	return transact.TransExeResult_Success
}

func (this *CoinSceneChangeTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("CoinSceneChangeTransactHandler.OnCommit ")
	ud := tNode.TransEnv.GetField(CoinSceneChangeTransactParam)
	if ctx, ok := ud.(*CoinSceneChangeCtx); ok {
		player := PlayerMgrSington.GetPlayerBySnId(ctx.snid)
		if player != nil {
			if ctx.isClub {
				//ClubSceneMgrSington.ClearPlayerChanging(player)
			} else {
				CoinSceneMgrSington.ClearPlayerChanging(player)
			}
			op := common.CoinSceneOp_Change
			if ctx.isAudience {
				op = common.CoinSceneOp_AudienceChange
			}
			pack := &hall_proto.SCCoinSceneOp{
				Id:       proto.Int32(ctx.id),
				OpType:   proto.Int32(op),
				OpParams: ctx.exclude,
			}

			//此处需要判定是否在线，异步后，可能玩家已经离线，无法再执行此操作了
			if player.IsOnLine() == false {
				pack := &hall_proto.SCLeaveRoom{
					OpRetCode: hall_proto.OpResultCode_Game_OPRC_Sucess_Game,
					RoomId:    proto.Int32(ctx.sceneId),
				}
				proto.SetDefaults(pack)
				player.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_LEAVEROOM), pack)
				return transact.TransExeResult_Success
			}

			var ret hall_proto.OpResultCode
			if ctx.isAudience {
				noother := false
				ret = CoinSceneMgrSington.AudienceEnter(player, ctx.id, 0, ctx.exclude, true)
				if ret != hall_proto.OpResultCode_OPRC_Sucess { //失败，还进当前桌
					ret = CoinSceneMgrSington.AudienceEnter(player, ctx.id, 0, nil, true)
					if ret == hall_proto.OpResultCode_OPRC_Sucess {
						noother = true
					}
				}
				if ret != hall_proto.OpResultCode_OPRC_Sucess { //失败，直接离场
					pack := &hall_proto.SCLeaveRoom{
						OpRetCode: hall_proto.OpResultCode_Game_OPRC_Sucess_Game,
						RoomId:    proto.Int32(ctx.sceneId),
					}
					proto.SetDefaults(pack)
					player.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_LEAVEROOM), pack)
				}
				if ret == hall_proto.OpResultCode_OPRC_Sucess && noother {
					ret = hall_proto.OpResultCode_OPRC_NoOtherDownTiceRoom
				}
			} else {
				if ctx.isClub {
					//ret = ClubSceneMgrSington.ClubPlayerEnter(player, ctx.clubId, ctx.roomId, int(-1), "", true,
					//	int(ctx.sceneId))
				} else {
					ret = CoinSceneMgrSington.PlayerEnter(player, ctx.id, 0, ctx.exclude, true)
				}
				if !(ret == hall_proto.OpResultCode_OPRC_Sucess || ret == hall_proto.OpResultCode_OPRC_CoinSceneEnterQueueSucc) { //失败，直接离场
					pack := &hall_proto.SCLeaveRoom{
						OpRetCode: hall_proto.OpResultCode_Game_OPRC_Sucess_Game,
						RoomId:    proto.Int32(ctx.sceneId),
					}
					proto.SetDefaults(pack)
					player.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_LEAVEROOM), pack)
				}
			}
			pack.OpCode = ret
			proto.SetDefaults(pack)
			player.SendToClient(int(hall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_OP), pack)
		}
	}
	return transact.TransExeResult_Success
}

func (this *CoinSceneChangeTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("CoinSceneChangeTransactHandler.OnRollBack ")
	ud := tNode.TransEnv.GetField(CoinSceneChangeTransactParam)
	if ctx, ok := ud.(*CoinSceneChangeCtx); ok {
		player := PlayerMgrSington.GetPlayerBySnId(ctx.snid)
		if player != nil {
			if ctx.isClub {
				//ClubSceneMgrSington.ClearPlayerChanging(player)
			} else {
				CoinSceneMgrSington.ClearPlayerChanging(player)
			}
			op := common.CoinSceneOp_Change
			if ctx.isAudience {
				op = common.CoinSceneOp_AudienceChange
			}
			pack := &hall_proto.SCCoinSceneOp{
				Id:       proto.Int32(ctx.id),
				OpType:   proto.Int32(op),
				OpParams: ctx.exclude,
				OpCode:   hall_proto.OpResultCode_OPRC_CoinSceneYouAreGaming,
			}
			proto.SetDefaults(pack)
			player.SendToClient(int(hall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_OP), pack)
		}
	}
	return transact.TransExeResult_Success
}

func (this *CoinSceneChangeTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("CoinSceneChangeTransactHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

func init() {
	transact.RegisteHandler(common.TransType_CoinSceneChange, &CoinSceneChangeTransactHandler{})
}
