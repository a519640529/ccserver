package main

//
//import (
//	"games.yol.com/win88/common"
//	"games.yol.com/win88/proto"
//	"games.yol.com/win88/protocol/mngame"
//	"games.yol.com/win88/protocol/server"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/idealeak/goserver/core/netlib"
//)
//
//type CSMNGameEnterPacketFactory struct {
//}
//type CSMNGameEnterHandler struct {
//}
//
//func (this *CSMNGameEnterPacketFactory) CreatePacket() interface{} {
//	pack := &mngame.CSMNGameEnter{}
//	return pack
//}
//
//func (this *CSMNGameEnterHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSMNGameEnterHandler Process recv ", data)
//	if msg, ok := data.(*mngame.CSMNGameEnter); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p != nil {
//			code := MiniGameMgrSington.PlayerEnter(p, msg.GetId())
//			pack := &mngame.SCMNGameEnter{
//				Id:        msg.GetId(),
//				OpRetCode: code,
//			}
//			proto.SetDefaults(pack)
//			logger.Logger.Tracef("CSMNGameEnterHandler Process recv %v ", pack)
//			p.SendToClient(int(mngame.MNGamePacketID_PACKET_SC_MNGAME_ENTER), pack)
//		}
//	}
//
//	return nil
//}
//
//type CSMNGameLeavePacketFactory struct {
//}
//type CSMNGameLeaveHandler struct {
//}
//
//func (this *CSMNGameLeavePacketFactory) CreatePacket() interface{} {
//	pack := &mngame.CSMNGameLeave{}
//	return pack
//}
//
//func (this *CSMNGameLeaveHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSMNGameLeaveHandler Process recv ", data)
//	if msg, ok := data.(*mngame.CSMNGameLeave); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p != nil {
//			code := MiniGameMgrSington.PlayerLeave(p, msg.GetId())
//			pack := &mngame.SCMNGameLeave{
//				Id:        msg.GetId(),
//				OpRetCode: code,
//				Reason:    int32(common.PlayerLeaveReason_Normal),
//			}
//			proto.SetDefaults(pack)
//			p.SendToClient(int(mngame.MNGamePacketID_PACKET_SC_MNGAME_LEAVE), pack)
//		}
//	}
//
//	return nil
//}
//
//type CSMNGameDispatcherPacketFactory struct {
//}
//type CSMNGameDispatcherHandler struct {
//}
//
//func (this *CSMNGameDispatcherPacketFactory) CreatePacket() interface{} {
//	pack := &mngame.CSMNGameDispatcher{}
//	return pack
//}
//
//func (this *CSMNGameDispatcherHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSMNGameDispatcherHandler Process recv ", data)
//	if msg, ok := data.(*mngame.CSMNGameDispatcher); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p != nil {
//			MiniGameMgrSington.PlayerMsgDispatcher(p, msg)
//		}
//	}
//
//	return nil
//}
//
//type GWPlayerLeaveMiniGamePacketFactory struct {
//}
//type GWPlayerLeaveMiniGameHandler struct {
//}
//
//func (this *GWPlayerLeaveMiniGamePacketFactory) CreatePacket() interface{} {
//	pack := &server.GWPlayerLeaveMiniGame{}
//	return pack
//}
//
//func (this *GWPlayerLeaveMiniGameHandler) Process(s *netlib.Session, packetid int, data interface{}) error {
//	logger.Logger.Trace("GWPlayerLeaveMiniGameHandler Process recv ", data)
//	if msg, ok := data.(*server.GWPlayerLeaveMiniGame); ok {
//		p := PlayerMgrSington.GetPlayerBySnId(msg.SnId)
//		if p != nil {
//			plt := p.GetPlatform()
//			s := MiniGameMgrSington.GetScene(plt, msg.GetGameFreeId())
//			if s != nil {
//				delete(s.players, p.SnId)
//			}
//
//			gamings, ok := MiniGameMgrSington.playerGaming[p.SnId]
//			if ok {
//				delete(gamings, msg.GetGameFreeId())
//			}
//
//			pack := &mngame.SCMNGameLeave{
//				Id:        msg.GetGameFreeId(),
//				Reason:    msg.GetReason(),
//				OpRetCode: mngame.MNGameOpResultCode_MNGAME_OPRC_Sucess,
//			}
//			proto.SetDefaults(pack)
//			p.SendToClient(int(mngame.MNGamePacketID_PACKET_SC_MNGAME_LEAVE), pack)
//		}
//	}
//
//	return nil
//}
//
//func init() {
//	common.RegisterHandler(int(mngame.MNGamePacketID_PACKET_CS_MNGAME_ENTER), &CSMNGameEnterHandler{})
//	netlib.RegisterFactory(int(mngame.MNGamePacketID_PACKET_CS_MNGAME_ENTER), &CSMNGameEnterPacketFactory{})
//
//	common.RegisterHandler(int(mngame.MNGamePacketID_PACKET_CS_MNGAME_LEAVE), &CSMNGameLeaveHandler{})
//	netlib.RegisterFactory(int(mngame.MNGamePacketID_PACKET_CS_MNGAME_LEAVE), &CSMNGameLeavePacketFactory{})
//
//	common.RegisterHandler(int(mngame.MNGamePacketID_PACKET_CS_MNGAME_DISPATCHER), &CSMNGameDispatcherHandler{})
//	netlib.RegisterFactory(int(mngame.MNGamePacketID_PACKET_CS_MNGAME_DISPATCHER), &CSMNGameDispatcherPacketFactory{})
//
//	netlib.RegisterHandler(int(server.SSPacketID_PACKET_GW_PLAYERLEAVE_MINIGAME), &GWPlayerLeaveMiniGameHandler{})
//	netlib.RegisterFactory(int(server.SSPacketID_PACKET_GW_PLAYERLEAVE_MINIGAME), &GWPlayerLeaveMiniGamePacketFactory{})
//}
