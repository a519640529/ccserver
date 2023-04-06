package action

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/gamehall"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSDestroyRoomPacketFactory struct {
}
type CSDestroyRoomHandler struct {
}

func (this *CSDestroyRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSDestroyRoom{}
	return pack
}

func (this *CSDestroyRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSDestroyRoomHandler Process recv ", data)
	p := base.PlayerMgrSington.GetPlayer(sid)
	if p == nil {
		logger.Logger.Warn("CSDestroyRoomHandler p == nil")
		return nil
	}
	scene := p.GetScene()
	if scene == nil {
		logger.Logger.Warn("CSDestroyRoomHandler p.GetScene() == nil")
		return nil
	}
	if !scene.HasPlayer(p) {
		return nil
	}
	if scene.Creator != p.SnId {
		logger.Logger.Warn("CSDestroyRoomHandler s.creator != p.AccountId")
		return nil
	}
	scene.Destroy(true)
	return nil
}

type CSLeaveRoomPacketFactory struct {
}
type CSLeaveRoomHandler struct {
}

func (this *CSLeaveRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSLeaveRoom{}
	return pack
}

func (this *CSLeaveRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSLeaveRoomHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSLeaveRoom); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSLeaveRoomHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warnf("CSLeaveRoomHandler[%v] p.GetScene() == nil", p.SnId)
			return nil
		}
		if msg.GetMode() == 0 && !scene.CanChangeCoinScene(p) {
			logger.Logger.Warnf("CSLeaveRoomHandler[%v][%v] scene.gaming==true", scene.SceneId, p.SnId)
			pack := &gamehall.SCLeaveRoom{
				OpRetCode: gamehall.OpResultCode_Game_OPRC_YourAreGamingCannotLeave_Game,
				RoomId:    proto.Int(scene.SceneId),
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_LEAVEROOM), pack)
			return nil
		}
		if msg.GetMode() == 0 {
			if scene.HasAudience(p) {
				scene.AudienceLeave(p, common.PlayerLeaveReason_Normal)
			} else if scene.HasPlayer(p) {
				scene.PlayerLeave(p, common.PlayerLeaveReason_Normal, false)
			}
		} else {
			pack := &gamehall.SCLeaveRoom{
				Reason:    proto.Int(0),
				OpRetCode: gamehall.OpResultCode_Game_OPRC_Sucess_Game,
				Mode:      msg.Mode,
				RoomId:    proto.Int(scene.SceneId),
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_LEAVEROOM), pack)
			scene.PlayerDropLine(p.SnId)

			//标记上暂离状态
			p.ActiveLeave = true
			p.MarkFlag(base.PlayerState_Online)
			p.MarkFlag(base.PlayerState_Leave)
			p.SyncFlag()
		}
	}
	return nil
}

type CSAudienceLeaveRoomHandler struct {
}

func (this *CSAudienceLeaveRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSAudienceLeaveRoomHandler Process recv ", data)
	p := base.PlayerMgrSington.GetPlayer(sid)
	if p == nil {
		logger.Logger.Warn("CSAudienceLeaveRoomHandler p == nil")
		return nil
	}
	scene := p.GetScene()
	if scene == nil {
		logger.Logger.Warn("CSAudienceLeaveRoomHandler p.GetScene() == nil")
		return nil
	}
	if !scene.HasAudience(p) {
		return nil
	}
	scene.AudienceLeave(p, common.PlayerLeaveReason_Normal)
	return nil
}

type CSForceStartPacketFactory struct {
}
type CSForceStartHandler struct {
}

func (this *CSForceStartPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSForceStart{}
	return pack
}

func (this *CSForceStartHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSForceStartHandler Process recv ", data)
	p := base.PlayerMgrSington.GetPlayer(sid)
	if p == nil {
		logger.Logger.Warn("CSForceStartHandler p == nil")
		return nil
	}
	if p.GetScene() == nil {
		logger.Logger.Warn("CSForceStartHandler p.GetScene() == nil")
		return nil
	}
	if p.Pos != 0 /*p.GetScene().creator != p.SnId*/ { //第1个进房间的玩家
		logger.Logger.Warn("CSForceStartHandler p.GetScene().creator != p.SnId")
		return nil
	}
	if p.GetScene().Gaming {
		logger.Logger.Warn("CSForceStartHandler p.GetScene().gaming==true")
		return nil
	}
	if !p.GetScene().GetScenePolicy().IsCanForceStart(p.GetScene()) {
		logger.Logger.Warn("CSForceStartHandler !p.GetScene().sp.IsCanForceStart(p.GetScene())")
		return nil
	}
	//强制开始
	p.GetScene().GetScenePolicy().ForceStart(p.GetScene())

	p.GetScene().NotifySceneRoundStart(1)

	packClient := &gamehall.SCForceStart{}
	proto.SetDefaults(packClient)
	p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_FORCESTART), packClient)
	return nil
}

//type CSChangeCardPacketFactory struct {
//}
//type CSChangeCardHandler struct {
//}

//func (this *CSChangeCardPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSChangeCard{}
//	return pack
//}

//func (this *CSChangeCardHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSChangeCardHandler Process recv ", data)
//	p := base.PlayerMgrSington.GetPlayer(sid)
//	if p == nil {
//		logger.Logger.Warn("CSChangeCardHandler p == nil")
//		return nil
//	}
//	if p.GetScene() == nil {
//		logger.Logger.Warn("CSChangeCardHandler p.GetScene() == nil")
//		return nil
//	}
//	if !p.GetScene().gaming {
//		logger.Logger.Warn("CSChangeCardHandler p.GetScene().gaming==false")
//		return nil
//	}
//	if p.GMLevel < model.GMACData.ChangeCardLevel || model.GMACData.ChangeCardLevel == 0 {
//		logger.Logger.Warnf("CSChangeCardHandler %v p.GMLevel(%v) < model.GMACData.ChangeCardLevel(%v)", p.GetName(), p.GMLevel, model.GMACData.ChangeCardLevel)
//		return nil
//	}

//	if p.IsMarkFlag(PlayerState_Ting) || p.IsMarkFlag(PlayerState_Auto) {
//		logger.Logger.Warnf("CSChangeCardHandler %v p.IsMarkFlag(PlayerState_Ting) || p.IsMarkFlag(PlayerState_Auto)", p.GetName())
//		return nil
//	}
//	if gmapi, ok := p.extraData.(MahjongPlayerGMInterface); ok {
//		lastCard, newCard := gmapi.ChangeLastCard()
//		if lastCard >= 0 && lastCard < mahjong.MJ_MAX {
//			pack := &protocol.SCChangeCard{
//				Card:    proto.Int32(lastCard),
//				NewCard: proto.Int32(newCard),
//			}
//			proto.SetDefaults(pack)
//			p.SendToClient(int(protocol.AgentGamePacketID_PACKET_SC_CHANGECARD), pack)
//			logger.Logger.Warnf("CSChangeCardHandler %v:gmlvl:%v changcard %v->%v", p.GetName(), p.GMLevel, lastCard, newCard)
//		}
//	}
//	return nil
//}

type CSPlayerSwithFlagPacketFactory struct {
}
type CSPlayerSwithFlagHandler struct {
}

func (this *CSPlayerSwithFlagPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSPlayerSwithFlag{}
	return pack
}

func (this *CSPlayerSwithFlagHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPlayerSwithFlagHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSPlayerSwithFlag); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSPlayerSwithFlagHandler p == nil")
			return nil
		}
		flag := int(msg.GetFlag())
		//flag = 1 << uint(flag)
		//if msg.Mark != nil {
		logger.Logger.Trace("CSPlayerSwithFlagHandler Process recv SnId(%v) Mark is %v", p.SnId, msg.GetMark())
		if msg.GetMark() == 0 { //取消状态
			oldFlag := p.GetFlag()
			if p.IsMarkFlag(flag) {
				p.UnmarkFlag(flag)
			}
			if flag == base.PlayerState_Leave {
				if p.GetScene() != nil {
					//重置下房间状态
					p.GetScene().PlayerReturn(p, true)
				}
			}
			if oldFlag != p.GetFlag() {
				p.SyncFlag()
			}
		} else { //设置状态
			if flag == base.PlayerState_Leave {
				p.ActiveLeave = false //被动暂离
				if p.GetScene() != nil {
					//todo dev fish 字游戏暂时被删除了 所以这里先注释掉
					//if p.GetScene().gameId == common.GameId_HFishing || p.GetScene().gameId == common.GameId_LFishing ||
					//	p.GetScene().gameId == common.GameId_RFishing || p.GetScene().gameId == common.GameId_DFishing ||
					//	p.GetScene().gameId == common.GameId_NFishing || p.GetScene().gameId == common.GameId_TFishing {
					//	p.GetScene().sp.OnPlayerOp(p.GetScene(), p, FishingPlayerOpLeave, []int64{})
					//}
				}
			}
			if !p.IsMarkFlag(flag) {
				p.MarkFlag(flag)
				p.SyncFlag()
			}
		}
		//}
	}
	return nil
}

//type CSLowerRicePacketFactory struct {
//}
//type CSLowerRiceHandler struct {
//}
//
//func (this *CSLowerRicePacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSLowerRice{}
//	return pack
//}
//
//func (this *CSLowerRiceHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSLowerRiceHandler Process recv ", data)
//	if msg, ok := data.(*gamehall.CSLowerRice); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSLowerRiceHandler p == nil")
//			return nil
//		}
//		bp := base.PlayerMgrSington.GetPlayerBySnId(msg.GetSnid())
//		p.DownRice(msg.GetCoin(), bp, p)
//	}
//	return nil
//}

//type CSCDezhouPokerWPUpdatePacketFactory struct {
//}
//type CSCDezhouPokerWPUpdateHandler struct {
//}
//
//func (this *CSCDezhouPokerWPUpdatePacketFactory) CreatePacket() interface{} {
//	pack := &server.CSCDezhouPokerWPUpdate{}
//	return pack
//}

//func (this *CSCDezhouPokerWPUpdateHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSCDezhouPokerWPUpdateHandler Process recv ", data)
//	p := base.PlayerMgrSington.GetPlayer(sid)
//	if p == nil {
//		logger.Logger.Warn("CSCDezhouPokerWPUpdateHandler p == nil")
//		return nil
//	}
//
//	if !p.IsRob {
//		return nil
//	}
//
//	scene := p.GetScene()
//	if scene == nil {
//		logger.Logger.Warn("CSCDezhouPokerWPUpdateHandler p.GetScene() == nil")
//		return nil
//	}
//
//	//转发给真实玩家
//	for _, p := range scene.players {
//		if p != nil && !p.IsRob && p.GMLevel > 0 {
//			p.SendToClient(packetid, data)
//		}
//	}
//	return nil
//}

//type CSCOmahaPokerWPUpdatePacketFactory struct {
//}
//type CSCOmahaPokerWPUpdateHandler struct {
//}
//
//func (this *CSCOmahaPokerWPUpdatePacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSCOmahaPokerWPUpdate{}
//	return pack
//}
//
//func (this *CSCOmahaPokerWPUpdateHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSCOmahaPokerWPUpdateHandler Process recv ", data)
//	p := base.PlayerMgrSington.GetPlayer(sid)
//	if p == nil {
//		logger.Logger.Warn("CSCOmahaPokerWPUpdateHandler p == nil")
//		return nil
//	}
//
//	if !p.IsRob {
//		return nil
//	}
//
//	scene := p.GetScene()
//	if scene == nil {
//		logger.Logger.Warn("CSCOmahaPokerWPUpdateHandler p.GetScene() == nil")
//		return nil
//	}
//
//	//转发给真实玩家
//	for _, p := range scene.players {
//		if p != nil && !p.IsRob && p.GMLevel > 0 {
//			p.SendToClient(packetid, data)
//		}
//	}
//	return nil
//}

func init() {
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_DESTROYROOM), &CSDestroyRoomHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_DESTROYROOM), &CSDestroyRoomPacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_LEAVEROOM), &CSLeaveRoomHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_LEAVEROOM), &CSLeaveRoomPacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_AUDIENCE_LEAVEROOM), &CSAudienceLeaveRoomHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_AUDIENCE_LEAVEROOM), &CSLeaveRoomPacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_FORCESTART), &CSForceStartHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_FORCESTART), &CSForceStartPacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_PLAYER_SWITCHFLAG), &CSPlayerSwithFlagHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_PLAYER_SWITCHFLAG), &CSPlayerSwithFlagPacketFactory{})
}
