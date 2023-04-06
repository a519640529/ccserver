package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/gamehall"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSCoinSceneGetPlayerNumPacketFactory struct {
}
type CSCoinSceneGetPlayerNumHandler struct {
}

func (this *CSCoinSceneGetPlayerNumPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSCoinSceneGetPlayerNum{}
	return pack
}

func (this *CSCoinSceneGetPlayerNumHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSCoinSceneGetPlayerNumHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSCoinSceneGetPlayerNum); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil {
			nums := CoinSceneMgrSington.GetPlayerNums(p, msg.GetGameId(), msg.GetGameModel())
			pack := &gamehall.SCCoinSceneGetPlayerNum{
				Nums: nums,
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(gamehall.CoinSceneGamePacketID_PACKET_SC_COINSCENE_GETPLAYERNUM), pack)
		}
	}

	return nil
}

type CSCoinSceneOpPacketFactory struct {
}
type CSCoinSceneOpHandler struct {
}

func (this *CSCoinSceneOpPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSCoinSceneOp{}
	return pack
}

func (this *CSCoinSceneOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSCoinSceneOpHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSCoinSceneOp); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil {
			var ret gamehall.OpResultCode
			pack := &gamehall.SCCoinSceneOp{
				Id:     msg.Id,
				OpType: msg.OpType,
			}

			oldPlatform := p.Platform
			switch msg.GetOpType() {
			case common.CoinSceneOp_Enter:
				//pt := PlatformMgrSington.GetPackageTag(p.PackageID)
				//if pt != nil && pt.IsForceBind == 1 {
				//	if p.BeUnderAgentCode == "" || p.BeUnderAgentCode == "0" {
				//		ret = gamehall.OpResultCode_OPRC_MustBindPromoter
				//		goto done
				//	}
				//}

				//已经在房间里了直接返回
				if p.scene != nil {
					logger.Logger.Warnf("CSCoinSceneOpHandler CoinSceneOp_Enter found snid:%v had in scene:%v gameid:%v", p.SnId, p.scene.sceneId, p.scene.gameId)
					p.ReturnScene(false)
					return nil
				}

				var roomId int32
				params := msg.GetOpParams()
				if len(params) != 0 {
					roomId = params[0]
					platformName := CoinSceneMgrSington.GetPlatformBySceneId(int(roomId))
					if p.IsRob {
						p.Platform = platformName
					} else if p.GMLevel > 0 && p.Platform == platformName { //允许GM直接按房间ID进场
						roomId = params[0]
					}
				}
				if len(msg.GetPlatform()) > 0 && p.IsRob {
					p.Platform = msg.GetPlatform()
				}
				//检测房间状态是否开启
				gps := PlatformMgrSington.GetGameConfig(p.Platform, msg.GetId())
				if gps == nil {
					ret = gamehall.OpResultCode_OPRC_RoomHadClosed
					goto done
				}

				dbGameFree := gps.DbGameFree
				if dbGameFree == nil {
					ret = gamehall.OpResultCode_OPRC_RoomHadClosed
					goto done
				}

				if len(params) != 0 && (p.GMLevel > 0 || dbGameFree.GetCreateRoomNum() != 0) { //允许GM|或者可选房间的游戏直接按房间ID进场
					s := SceneMgrSington.GetScene(int(params[0]))
					if s != nil && s.groupId == gps.GroupId {
						roomId = params[0]
					}
				}

				if dbGameFree.GetLimitCoin() != 0 && int64(dbGameFree.GetLimitCoin()) > p.Coin {
					ret = gamehall.OpResultCode_OPRC_CoinNotEnough
					goto done
				}

				if dbGameFree.GetMaxCoinLimit() != 0 && int64(dbGameFree.GetMaxCoinLimit()) < p.Coin && !p.IsRob {
					ret = gamehall.OpResultCode_OPRC_CoinTooMore
					goto done
				}

				//检查游戏次数限制
				if !p.IsRob {
					todayData, _ := p.GetDaliyGameData(int(dbGameFree.GetId()))
					if dbGameFree.GetPlayNumLimit() != 0 &&
						todayData != nil &&
						todayData.GameTimes >= int64(dbGameFree.GetPlayNumLimit()) {
						ret = gamehall.OpResultCode_OPRC_RoomGameTimes
						goto done
					}
				}
				excludeSceneIds := p.lastSceneId[msg.GetId()]
				//临时修改
				var NewSceneIds = make([]int32, len(excludeSceneIds))
				copy(NewSceneIds, excludeSceneIds)
				if len(NewSceneIds) > 1 {
					NewSceneIds = NewSceneIds[len(NewSceneIds)-1:]
				}
				//--------------------------
				ret = CoinSceneMgrSington.PlayerEnter(p, msg.GetId(), roomId, NewSceneIds, false)
				if p.scene != nil {
					pack.OpParams = append(pack.OpParams, int32(p.scene.sceneId))
					//TODO 有房间还进入失败，尝试returnroom
					if ret != gamehall.OpResultCode_OPRC_Sucess {
						p.ReturnScene(false)
					}
				}
			case common.CoinSceneOp_Leave:
				ret = CoinSceneMgrSington.PlayerTryLeave(p, msg.GetId(), false)
				if gamehall.OpResultCode_OPRC_OpYield == ret {
					return nil
				}
			case common.CoinSceneOp_Change:
				if p.scene == nil {
					ret = gamehall.OpResultCode_OPRC_RoomHadClosed
					goto done
				}
				var exclude = int32(p.scene.sceneId)
				params := msg.GetOpParams()
				if len(params) != 0 {
					exclude = params[0]
				}
				if p.scene.IsPrivateScene() {
					//if ClubSceneMgrSington.PlayerInChanging(p) {
					//	return nil
					//}
					//ret = ClubSceneMgrSington.PlayerTryChange(p, msg.GetId(), []int32{exclude}, false)
				} else {
					if CoinSceneMgrSington.PlayerInChanging(p) { //换桌中
						return nil
					}
					excludeSceneIds := p.lastSceneId[msg.GetId()]
					if exclude != 0 {
						excludeSceneIds = append(excludeSceneIds, exclude)
					}
					ret = CoinSceneMgrSington.PlayerTryChange(p, msg.GetId(), excludeSceneIds, false)
				}

			case common.CoinSceneOp_AudienceEnter:
				var roomId int32
				params := msg.GetOpParams()
				if len(params) != 0 && !p.IsRob {
					roomId = params[0]
				}
				ret = CoinSceneMgrSington.AudienceEnter(p, msg.GetId(), roomId, nil, false)
			case common.CoinSceneOp_AudienceLeave:
				ret = CoinSceneMgrSington.PlayerTryLeave(p, msg.GetId(), true)
				if gamehall.OpResultCode_OPRC_OpYield == ret {
					return nil
				}
			case common.CoinSceneOp_AudienceChange:
				var exclude int32
				if p.scene != nil {
					exclude = int32(p.scene.sceneId)
				}
				params := msg.GetOpParams()
				if len(params) != 0 {
					exclude = params[0]
				}
				if CoinSceneMgrSington.PlayerInChanging(p) { //换桌中
					return nil
				}
				excludeSceneIds := p.lastSceneId[msg.GetId()]
				if exclude != 0 {
					excludeSceneIds = append(excludeSceneIds, exclude)
				}
				ret = CoinSceneMgrSington.PlayerTryChange(p, msg.GetId(), excludeSceneIds, true)
			case common.CoinSceneOP_Server:
				if p.scene == nil {
					ret = gamehall.OpResultCode_OPRC_RoomHadClosed
					goto done
				}
				gameFreeId := p.scene.dbGameFree.GetId()
				gameConfig := PlatformMgrSington.GetGameConfig(p.Platform, gameFreeId)
				if gameConfig != nil && gameConfig.DbGameFree.GetMatchMode() == 1 {
					return nil
				}
				var exclude = int32(p.scene.sceneId)
				params := msg.GetOpParams()
				if len(params) != 0 {
					exclude = params[0]
				}
				if p.scene.IsPrivateScene() {
					//if ClubSceneMgrSington.PlayerInChanging(p) {
					//	return nil
					//}
					//ret = ClubSceneMgrSington.PlayerTryChange(p, msg.GetId(), []int32{exclude}, false)
				} else {
					if CoinSceneMgrSington.PlayerInChanging(p) { //换桌中
						return nil
					}
					excludeSceneIds := p.lastSceneId[msg.GetId()]
					if exclude != 0 {
						excludeSceneIds = append(excludeSceneIds, exclude)
					}
					ret = CoinSceneMgrSington.PlayerTryChange(p, msg.GetId(), excludeSceneIds, false)
				}
			}
		done:
			//机器人要避免身上的平台标记被污染
			if p.IsRob {
				if !(ret == gamehall.OpResultCode_OPRC_Sucess ||
					ret == gamehall.OpResultCode_OPRC_CoinSceneEnterQueueSucc) {
					p.Platform = oldPlatform
				}
			}
			pack.OpCode = ret
			proto.SetDefaults(pack)
			p.SendToClient(int(gamehall.CoinSceneGamePacketID_PACKET_SC_COINSCENE_OP), pack)
			if msg.GetOpType() == common.CoinSceneOp_Enter && ret == gamehall.OpResultCode_OPRC_Sucess && p.scene != nil {
				gameName := p.scene.dbGameFree.GetName() + p.scene.dbGameFree.GetTitle()
				ActMonitorMgrSington.SendActMonitorEvent(ActState_Game, p.SnId, p.Name, p.Platform,
					0, 0, gameName, 0)
			}
		}
	}
	return nil
}

type CSCoinSceneListRoomPacketFactory struct {
}
type CSCoinSceneListRoomHandler struct {
}

func (this *CSCoinSceneListRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSCoinSceneListRoom{}
	return pack
}

func (this *CSCoinSceneListRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSCoinSceneListRoomHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSCoinSceneListRoom); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil {
			if !CoinSceneMgrSington.ListRooms(p, msg.GetId()) {
				pack := &gamehall.SCCoinSceneListRoom{
					Id: msg.Id,
				}
				proto.SetDefaults(pack)
				p.SendToClient(int(gamehall.CoinSceneGamePacketID_PACKET_SC_COINSCENE_LISTROOM), pack)
			}
		}
	}

	return nil
}

func init() {
	common.RegisterHandler(int(gamehall.CoinSceneGamePacketID_PACKET_CS_COINSCENE_GETPLAYERNUM), &CSCoinSceneGetPlayerNumHandler{})
	netlib.RegisterFactory(int(gamehall.CoinSceneGamePacketID_PACKET_CS_COINSCENE_GETPLAYERNUM), &CSCoinSceneGetPlayerNumPacketFactory{})

	common.RegisterHandler(int(gamehall.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), &CSCoinSceneOpHandler{})
	netlib.RegisterFactory(int(gamehall.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), &CSCoinSceneOpPacketFactory{})

	common.RegisterHandler(int(gamehall.CoinSceneGamePacketID_PACKET_CS_COINSCENE_LISTROOM), &CSCoinSceneListRoomHandler{})
	netlib.RegisterFactory(int(gamehall.CoinSceneGamePacketID_PACKET_CS_COINSCENE_LISTROOM), &CSCoinSceneListRoomPacketFactory{})
}
