package action

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/gamehall"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSCoinSceneOpPacketFactory struct {
}
type CSCoinSceneOpHandler struct {
}

func (this *CSCoinSceneOpPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSCoinSceneOp{}
	return pack
}

func (this *CSCoinSceneOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	if msg, ok := data.(*gamehall.CSCoinSceneOp); ok {
		logger.Logger.Trace("CSCoinSceneOpHandler ", msg)
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSCoinSceneOpHandler p == nil", data)
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSCoinSceneOpHandler p.scene == nil")
			return nil
		}
		if !scene.IsCoinScene() {
			return nil
		}
		switch msg.GetOpType() { //离开
		case common.CoinSceneOp_Leave:
			if !scene.HasPlayer(p) {
				return nil
			}
			if scene.CanChangeCoinScene(p) {
				scene.PlayerLeave(p, common.PlayerLeaveReason_Normal, true)
				return nil
			} else {
				//离开失败
				pack := &gamehall.SCCoinSceneOp{
					OpCode:   gamehall.OpResultCode_OPRC_YourAreGamingCannotLeave,
					Id:       msg.Id,
					OpType:   msg.OpType,
					OpParams: msg.OpParams,
				}
				proto.SetDefaults(pack)
				p.SendToClient(int(gamehall.CoinSceneGamePacketID_PACKET_SC_COINSCENE_OP), pack)
			}
		case common.CoinSceneOp_AudienceLeave:
			if scene.HasPlayer(p) {
				if scene.CanChangeCoinScene(p) {
					scene.PlayerLeave(p, common.PlayerLeaveReason_Normal, true)
					return nil
				}
			}
			if !scene.HasAudience(p) {
				return nil
			}
			if scene.CanChangeCoinScene(p) {
				scene.AudienceLeave(p, common.PlayerLeaveReason_Normal)
				return nil
			}
		}
		return nil
	}
	return nil
}

//奔驰宝马的操作
//type CSRollCoinOpPacketFactory struct {
//}
//type CSRollCoinOpHandler struct {
//}
//
//func (this *CSRollCoinOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSRollCoinOp{}
//	return pack
//}
//func (this *CSRollCoinOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSRollCoinOpHandler Process recv ", data)
//	if CSRollCoinOp, ok := data.(*protocol.CSRollCoinOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSRollCoinOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSRollCoinOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_RollCoin {
//			logger.Logger.Error("CSRollCoinOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(CSRollCoinOp.GetOpCode()), CSRollCoinOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//深林舞会的操作
//type CSRollColorOpPacketFactory struct {
//}
//type CSRollColorOpHandler struct {
//}
//
//func (this *CSRollColorOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSRollColorOp{}
//	return pack
//}
//func (this *CSRollColorOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSRollColorOpHandler Process recv ", data)
//	if CSRollColorOp, ok := data.(*protocol.CSRollColorOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSRollColorOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSRollColorOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_RollColor {
//			logger.Logger.Error("CSRollColorOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(CSRollColorOp.GetOpCode()), CSRollColorOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//老虎机的操作
//type CSRollLineOpPacketFactory struct {
//}
//type CSRollLineOpHandler struct {
//}
//
//func (this *CSRollLineOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSRollLineOp{}
//	return pack
//}
//func (this *CSRollLineOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSRollColorOpHandler Process recv ", data)
//	if csRollLineOp, ok := data.(*protocol.CSRollLineOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSRollLineOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSRollLineOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_RollLine {
//			logger.Logger.Error("CSRollLineOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csRollLineOp.GetOpCode()), csRollLineOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//世界杯的操作
//type CSRollTeamOpPacketFactory struct {
//}
//type CSRollTeamOpHandler struct {
//}
//
//func (this *CSRollTeamOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSRollTeamOp{}
//	return pack
//}
//func (this *CSRollTeamOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSRollColorOpHandler Process recv ", data)
//	if csRollTeamOp, ok := data.(*protocol.CSRollTeamOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSRollTeamOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSRollTeamOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_RollTeam {
//			logger.Logger.Error("CSRollTeamOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csRollTeamOp.GetOpCode()), csRollTeamOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//
//type CSPaodekuaiOpPacketFactory struct {
//}
//type CSPaodekuaiOpHandler struct {
//}
//
//func (this *CSPaodekuaiOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSPaodekuaiOp{}
//	return pack
//}
//
//func (this *CSPaodekuaiOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSPaodekuaiOpHandler Process recv ", data)
//	if csPaodekuaiOp, ok := data.(*protocol.CSPaodekuaiOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSPaodekuaiOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSPaodekuaiOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_PaoDeKuai {
//			logger.Logger.Error("CSPaodekuaiOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if p.IsMarkFlag(PlayerState_Auto) && (int(csPaodekuaiOp.GetOpCode()) == PaodekuaiPlayerOpDropCard ||
//			int(csPaodekuaiOp.GetOpCode()) == PaodekuaiPlayerOpDropCard) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csPaodekuaiOp.GetOpCode()), csPaodekuaiOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//推饼操作
//type CSTuibingOpPacketFactory struct {
//}
//type CSTuibingOpHandler struct {
//}
//
//func (this *CSTuibingOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSTuibingOp{}
//	return pack
//}
//
//func (this *CSTuibingOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSTuibingOpHandler Process recv ", data)
//	if CSTuibingOp, ok := data.(*protocol.CSTuibingOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSTuibingOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSTuibingOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_Tuibing {
//			logger.Logger.Error("CSTuibingOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(CSTuibingOp.GetOpCode()), CSTuibingOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//水果机
//type CSFruitMachineOpPacketFactory struct {
//}
//type CSFruitMachineOpHandler struct {
//}
//
//func (this *CSFruitMachineOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSFruitMachineOp{}
//	return pack
//}
//
//func (this *CSFruitMachineOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSFruitMachineOpHandler Process recv ", data)
//	if csFruitMachineOp, ok := data.(*protocol.CSFruitMachineOp); ok {
//
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		//p.MarkFlag(PlayerState_Online)
//		if p == nil {
//			logger.Logger.Warn("CSFruitMachineOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSFruitMachineOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_FruitMachine {
//			logger.Logger.Error("CSFruitMachineOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csFruitMachineOp.GetOpCode()), csFruitMachineOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//足球英豪
type CSFootBallHeroesOpPacketFactory struct {
}
type CSFootBallHeroesOpHandler struct {
}

//func (this *CSFootBallHeroesOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSFootBallHeroesOp{}
//	return pack
//}
//
//func (this *CSFootBallHeroesOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSFootBallHeroesOpHandler Process recv ", data)
//	if csFootBallHeroesOp, ok := data.(*protocol.CSFootBallHeroesOp); ok {
//
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		//p.MarkFlag(PlayerState_Online)
//		if p == nil {
//			logger.Logger.Warn("CSFootBallHeroesOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSFootBallHeroesOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_FootBallHeroes {
//			logger.Logger.Error("CSFootBallHeroesOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csFootBallHeroesOp.GetOpCode()), csFootBallHeroesOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//绝地求生
//type CSIslandSurvivalOpPacketFactory struct {
//}
//type CSIslandSurvivalOpHandler struct {
//}
//
//func (this *CSIslandSurvivalOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSIslandSurvivalOp{}
//	return pack
//}
//
//func (this *CSIslandSurvivalOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSIslandSurvivalOpHandler Process recv ", data)
//	if csIslandSurvivalOp, ok := data.(*protocol.CSIslandSurvivalOp); ok {
//
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		//p.MarkFlag(PlayerState_Online)
//		if p == nil {
//			logger.Logger.Warn("CSIslandSurvivalOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSIslandSurvivalOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_IslandSurvival {
//			logger.Logger.Error("CSIslandSurvivalOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csIslandSurvivalOp.GetOpCode()), csIslandSurvivalOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//女赌神
//type CSGoddessOpPacketFactory struct {
//}
//type CSGoddessOpHandler struct {
//}
//
//func (this *CSGoddessOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSGoddessOp{}
//	return pack
//}
//
//func (this *CSGoddessOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSGoddessOpHandler Process recv ", data)
//	if csGoddessOp, ok := data.(*protocol.CSGoddessOp); ok {
//
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		//p.MarkFlag(PlayerState_Online)
//		if p == nil {
//			logger.Logger.Warn("CSGoddessOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSGoddessOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_Goddess {
//			logger.Logger.Error("CSGoddessOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csGoddessOp.GetOpCode()), csGoddessOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//水浒传
type CSWaterMarginOpPacketFactory struct {
}
type CSWaterMarginOpHandler struct {
}

//func (this *CSWaterMarginOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSWarterMarginOp{}
//	return pack
//}
//
//func (this *CSWaterMarginOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSWaterMarginOpHandler Process recv ", data)
//	if csWaterMarginOp, ok := data.(*protocol.CSWarterMarginOp); ok {
//
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		//p.MarkFlag(PlayerState_Online)
//		if p == nil {
//			logger.Logger.Warn("CSWaterMarginOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSWaterMarginOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_WarterMargin {
//			logger.Logger.Error("CSWaterMarginOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csWaterMarginOp.GetOpCode()), csWaterMarginOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//赢三张
//type CSWinThreePlayerOpPacketFactory struct {
//}
//type CSWinThreePlayerOpHandler struct {
//}
//
//func (this *CSWinThreePlayerOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSWinThreePlayerOp{}
//	return pack
//}
//func (this *CSWinThreePlayerOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSWinThreePlayerOpHandler Process recv ", data)
//	if msg, ok := data.(*protocol.CSWinThreePlayerOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSWinThreePlayerOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSWinThreePlayerOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_WinThree {
//			logger.Logger.Error("CSWinThreePlayerOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(msg.GetOpCode()), msg.GetOpParam())
//		}
//		return nil
//	}
//	return nil
//}

//龙虎斗
//type CSDragonVsTigerOpPacketFactory struct {
//}
//type CSDragonVsTigerOpHandler struct {
//}
//
//func (this *CSDragonVsTigerOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSDragonVsTiggerOp{}
//	return pack
//}
//
//func (this *CSDragonVsTigerOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSDragonVsTigerOpHandler Process recv ", data)
//	if csDragonVsTigerOp, ok := data.(*protocol.CSDragonVsTiggerOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSDragonVsTigerOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSDragonVsTigerOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_DragonVsTiger {
//			logger.Logger.Error("CSDragonVsTigerOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csDragonVsTigerOp.GetOpCode()), csDragonVsTigerOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//红黑大战
//type CSRedVsBlackOpPacketFactory struct {
//}
//type CSRedVsBlackOpHandler struct {
//}
//
//func (this *CSRedVsBlackOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSRedVsBlackOp{}
//	return pack
//}
//
//func (this *CSRedVsBlackOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if csRedVsBlackOp, ok := data.(*protocol.CSRedVsBlackOp); ok {
//		//logger.Logger.Trace("CSRedVsBlackOpHandler Process recv ", data, sid)
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSRedVsBlackOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSRedVsBlackOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_RedVsBlack {
//			logger.Logger.Error("CSRedVsBlackOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			logger.Logger.Warn("CSRedVsBlackOpHandler !scene.HasPlayer(p)")
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csRedVsBlackOp.GetOpCode()), csRedVsBlackOp.GetParams())
//		}
//
//		return nil
//	}
//	return nil
//}

//
//type CSDoudizhuOpPacketFactory struct {
//}
//type CSDoudizhuOpHandler struct {
//}
//
//func (this *CSDoudizhuOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSDoudizhuOp{}
//	return pack
//}
//
//func (this *CSDoudizhuOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSDoudizhuOpHandler Process recv ", data)
//	if csDoudizhuOp, ok := data.(*protocol.CSDoudizhuOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSDoudizhuOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSDoudizhuOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_DouDiZhu {
//			logger.Logger.Error("CSDoudizhuOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		//if scene.mp != nil && !scene.mp.CanOp(scene, p, int(csDoudizhuOp.GetOpCode()), csDoudizhuOp.GetParams()) {
//		//	logger.Logger.Warn("CSDoudizhuOpHandler scene.mp.CanOp==false")
//		//	return nil
//		//}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csDoudizhuOp.GetOpCode()), csDoudizhuOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//德州扑克
//type CSDezhouPokerPlayerOpPacketFactory struct {
//}
//type CSDezhouPokerPlayerOpHandler struct {
//}
//
//func (this *CSDezhouPokerPlayerOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSDezhouPokerPlayerOp{}
//	return pack
//}
//func (this *CSDezhouPokerPlayerOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSDezhouPokerPlayerOpHandler Process recv ", data)
//	if msg, ok := data.(*protocol.CSDezhouPokerPlayerOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSDezhouPokerPlayerOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSDezhouPokerPlayerOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_DezhouPoker {
//			logger.Logger.Error("CSDezhouPokerPlayerOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(msg.GetOpCode()), msg.GetOpParam())
//		}
//		return nil
//	}
//	return nil
//}

//type CSClassicBallFightOpPacketFactory struct {
//}
//type CSClassicBallFightOpHandler struct {
//}
//
//func (this *CSClassicBallFightOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSClassicBallFightOp{}
//	return pack
//}
//
//func (this *CSClassicBallFightOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSClassicBallFightOpHandler Process recv ", data)
//	if csClassicBallFightOp, ok := data.(*protocol.CSClassicBallFightOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSClassicBallFightOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSClassicBallFightOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_Bullfight {
//			logger.Logger.Error("CSClassicBallFightOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csClassicBallFightOp.GetOpCode()), csClassicBallFightOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//二人麻将
//type CSTwoMahjongOpFactory struct {
//}
//type CSTwoMahjongOpHandler struct {
//}
//
//func (this *CSTwoMahjongOpFactory) CreatePacket() interface{} {
//	pack := &protocol.CSTwoMahjongPlayerOp{}
//	return pack
//}
//
//func (this *CSTwoMahjongOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if cSTwoMahjongPlayerOp, ok := data.(*protocol.CSTwoMahjongPlayerOp); ok {
//		//logger.Logger.Trace("CSRedVsBlackOpHandler Process recv ", data, sid)
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSRedVsBlackOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSRedVsBlackOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_Mahjong_ErRen {
//			logger.Logger.Error("CSRedVsBlackOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			logger.Logger.Warn("CSRedVsBlackOpHandler !scene.HasPlayer(p)")
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(cSTwoMahjongPlayerOp.GetOpCode()), cSTwoMahjongPlayerOp.Params)
//		}
//
//		return nil
//	}
//	return nil
//}

//百人牛牛
//type CSHundredBullOpPacketFactory struct {
//}
//type CSHundredBullOpHandler struct {
//}
//
//func (this *CSHundredBullOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSHundredBullOp{}
//	return pack
//}
//
//func (this *CSHundredBullOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSHundredBullOpHandler Process recv ", data)
//	if csHundredBullOp, ok := data.(*protocol.CSHundredBullOp); ok {
//		/*if csHundredBullOp.GetOpCode() == 101 { //玩家离开
//			if len(csHundredBullOp.GetParams()) > 0 {
//				snid := csHundredBullOp.GetParams()
//				p := base.PlayerMgrSington.GetPlayerBySnId(snid[0])
//				scene := p.scene
//				//scene.sp.OnPlayerOp(scene, p, HundredSceneOp_Leave, csHundredBullOp.GetParams())
//				scene.PlayerLeave(p, PlayerLeaveReason_Normal, true)
//			}
//			return nil
//		}
//		if csHundredBullOp.GetOpCode() == 102 { //观众进入
//			if len(csHundredBullOp.GetParams()) > 0 {
//				snid := csHundredBullOp.GetParams()
//				p := base.PlayerMgrSington.GetPlayerBySnId(snid[0])
//
//				p.MarkFlag(PlayerState_GameBreak)
//				p.SyncFlag()
//			}
//			return nil
//		}*/
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		//p.MarkFlag(PlayerState_Online)
//		if p == nil {
//			logger.Logger.Warn("CSHundredBullOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSHundredBullOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_HundredBull {
//			logger.Logger.Error("CSHundredBullOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csHundredBullOp.GetOpCode()), csHundredBullOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//十三水
//type CSThirteenWaterPlayerOpPacketFactory struct {
//}
//type CSThirteenWaterPlayerOpHandler struct {
//}
//
//func (this *CSThirteenWaterPlayerOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSThirteenWaterPlayerOp{}
//	return pack
//}
//func (this *CSThirteenWaterPlayerOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSThirteenWaterPlayerOpHandler Process recv ", data)
//	if msg, ok := data.(*protocol.CSThirteenWaterPlayerOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSThirteenWaterPlayerOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSThirteenWaterPlayerOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_ThirteenWater {
//			logger.Logger.Error("CSThirteenWaterPlayerOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(msg.GetOpCode()), msg.GetOpParam())
//		}
//		return nil
//	}
//	return nil
//}

//百家乐
//type CSBaccaratOpPacketFactory struct {
//}
//type CSBaccaratOpHandler struct {
//}
//
//func (this *CSBaccaratOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSBaccaratOp{}
//	return pack
//}
//
//func (this *CSBaccaratOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSDragonVsTigerOpHandler Process recv ", data)
//	if csBaccaratOp, ok := data.(*protocol.CSBaccaratOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSBaccaratOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSBaccaratOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_Baccarat {
//			logger.Logger.Error("CSBaccaratOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csBaccaratOp.GetOpCode()), csBaccaratOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

// CSTenHalfOpPacketFactory 十点半
//type CSTenHalfOpPacketFactory struct {
//}
//type CSTenHalfOpHandler struct {
//}
//
//func (this *CSTenHalfOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSTenHalfOp{}
//	return pack
//}
//
//func (this *CSTenHalfOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSDragonVsTigerOpHandler Process recv ", data)
//	if csTenhalfOp, ok := data.(*protocol.CSTenHalfOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSTenHalfOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSTenHalfOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_TenHalf {
//			logger.Logger.Error("CSTenHalfOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(csTenhalfOp.GetOpCode()), csTenhalfOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//
//type CSBlackJackOpPacketFactory struct {
//}
//type CSBlackJackOpHandler struct {
//}
//
//func (this *CSBlackJackOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSBlackJackOP{}
//	return pack
//}
//
//func (this *CSBlackJackOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSBlackJackOpHandler Process recv ", data)
//	if op, ok := data.(*protocol.CSBlackJackOP); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSBlackJackOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSBlackJackOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_BlackJack {
//			logger.Logger.Error("CSBlackJackOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) && !scene.HasAudience(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//梭哈
//type CSFiveCardStudOpPacketFactory struct {
//}
//type CSFiveCardStudOpHandler struct {
//}
//
//func (this *CSFiveCardStudOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSFiveCardStudPlayerOp{}
//	return pack
//}
//
//func (this *CSFiveCardStudOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSFiveCardStudOpHandler Process recv ", data)
//	if op, ok := data.(*protocol.CSFiveCardStudPlayerOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSFiveCardStudOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSFiveCardStudOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_FiveCardStud {
//			logger.Logger.Error("CSFiveCardStudOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetOpParam())
//		}
//		return nil
//	}
//	return nil
//}

//type CSGanDengYanOpPacketFactory struct {
//}
//type CSGanDengYanOpHandler struct {
//}
//
//func (this *CSGanDengYanOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSGanDengYanOP{}
//	return pack
//}
//
//func (this *CSGanDengYanOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSGanDengYanOpHandler Process recv ", data)
//	if op, ok := data.(*protocol.CSGanDengYanOP); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSGanDengYanOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSGanDengYanOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_GanDengYan {
//			logger.Logger.Error("CSGanDengYanOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			params := make([]int64, len(op.GetParams()))
//			for k, v := range op.GetParams() {
//				params[k] = int64(v)
//			}
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), params)
//		}
//		return nil
//	}
//	return nil
//}

//飞禽走兽操作
//type CSRollAnimalsOpPacketFactory struct {
//}
//type CSRollAnimalsOpHandler struct {
//}
//
//func (this *CSRollAnimalsOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSRollAnimalsOp{}
//	return pack
//}
//func (this *CSRollAnimalsOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSRollAnimalsOpHandler Process recv ", data)
//	if CSRollAnimalsOp, ok := data.(*protocol.CSRollAnimalsOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSRollAnimalsOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSRollAnimalsOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_RollAnimals {
//			logger.Logger.Error("CSRollAnimalsOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(CSRollAnimalsOp.GetOpCode()), CSRollAnimalsOp.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//type CSRollPointPacketFactory struct {
//}
//type CSRollPointHandler struct {
//}
//
//func (this *CSRollPointPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSRollPointOp{}
//	return pack
//}
//
//func (this *CSRollPointHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSRollPointOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSRollPointHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSRollPointHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

// 九线拉王
//type CSNineLineKingOpPacketFactory struct {
//}
//type CSNineLineKingOpHandler struct {
//}
//
//func (this *CSNineLineKingOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSNineLineKingOp{}
//	return pack
//}
//
//func (this *CSNineLineKingOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSNineLineKingOpHandler Process recv ", data)
//	if op, ok := data.(*protocol.CSNineLineKingOp); ok {
//
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		//p.MarkFlag(PlayerState_Online)
//		if p == nil {
//			logger.Logger.Warn("CSNineLineKingOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSNineLineKingOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_NineLineKing {
//			logger.Logger.Error("CSNineLineKingOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//type CSHundredZJHOpPacketFactory struct {
//}
//type CSHundredZJHOpHandler struct {
//}
//
//func (this *CSHundredZJHOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSHundredZJHOp{}
//	return pack
//}
//
//func (this *CSHundredZJHOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSHundredZJHOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSHundredZJHOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSHundredZJHOpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//type CSHundred28OpPacketFactory struct {
//}
//type CSHundred28OpHandler struct {
//}
//
//func (this *CSHundred28OpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSHundred28GOp{}
//	return pack
//}
//
//func (this *CSHundred28OpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSHundred28GOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSHundred28OpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSHundred28OpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//type CSSanGongOpPacketFactory struct {
//}
//type CSSanGongOpHandler struct {
//}
//
//func (this *CSSanGongOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSSanGongOp{}
//	return pack
//}
//
//func (this *CSSanGongOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSSanGongOpHandler Process recv ", data)
//	if op, ok := data.(*protocol.CSSanGongOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSSanGongOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSSanGongOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameID_SanGong {
//			logger.Logger.Error("CSSanGongOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//血战麻将
//type CSBloodMahjongOpPacketFactory struct {
//}
//type CSBloodMahjongOpHandler struct {
//}
//
//func (this *CSBloodMahjongOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSBloodMahjongPlayerOp{}
//	return pack
//}
//
//func (this *CSBloodMahjongOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	//logger.Logger.Trace("CSBloodMahjongOpHandler Process recv ", data)
//	if op, ok := data.(*protocol.CSBloodMahjongPlayerOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSBloodMahjongOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSBloodMahjongOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_BloodMahjong {
//			logger.Logger.Error("CSBloodMahjongOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetOpParam())
//		}
//		return nil
//	}
//	return nil
//}

//百人二八杠
//type CSHundred28GOpPacketFactory struct {
//}
//type CSHundred28GOpHandler struct {
//}
//
//func (this *CSHundred28GOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSHundred28GOp{}
//	return pack
//}
//
//func (this *CSHundred28GOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSHundred28GOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSHundred28GOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSHundred28GOpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//type CSPushingCoinOpPacketFactory struct {
//}
//
//func (this *CSPushingCoinOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSPushingCoinOP{}
//	return pack
//}
//
//type CSPushingCoinOpHandler struct {
//}
//
//func (this *CSPushingCoinOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSPushingCoinOP); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSPushingCoinOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSPushingCoinOpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOperate(scene, p, op)
//		}
//		return nil
//	}
//	return nil
//}

//  寻宝乐园
//type CSHuntingOpPacketFactory struct {
//}
//type CSHuntingOpHandler struct {
//}
//
//func (this *CSHuntingOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSHuntingOp{}
//	return pack
//}
//
//func (this *CSHuntingOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSHuntingOpHandler Process recv ", data)
//	if op, ok := data.(*protocol.CSHuntingOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSHuntingOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSHuntingOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_Hunting {
//			logger.Logger.Error("CSHuntingOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//  对战三公
//type CSSanGongPVPPlayerOpPacketFactory struct {
//}
//type CSSanGongPVPPlayerOpHandler struct {
//}
//
//func (this *CSSanGongPVPPlayerOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSSanGongPVPPlayerOp{}
//	return pack
//}
//
//func (this *CSSanGongPVPPlayerOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSSanGongPVPPlayerOpHandler Process recv ", data)
//	if op, ok := data.(*protocol.CSSanGongPVPPlayerOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSSanGongPVPPlayerOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSSanGongPVPPlayerOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_SanGongPVP {
//			logger.Logger.Error("CSSanGongPVPPlayerOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

// 德州牛仔
//type CSHundredDZNZOpPacketFactory struct {
//}
//type CSHundredDZNZOpHandler struct {
//}
//
//func (this *CSHundredDZNZOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSHundredDZNZOp{}
//	return pack
//}
//
//func (this *CSHundredDZNZOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSHundredDZNZOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSHundredDZNZOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSHundredDZNZOpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//奥马哈
//type CSOmahaPokerPlayerOpPacketFactory struct {
//}
//type CSOmahaPokerPlayerOpHandler struct {
//}
//
//func (this *CSOmahaPokerPlayerOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSOmahaPokerPlayerOp{}
//	return pack
//}
//func (this *CSOmahaPokerPlayerOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSOmahaPokerPlayerOpHandler Process recv ", data)
//	if op, ok := data.(*protocol.CSOmahaPokerPlayerOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSOmahaPokerPlayerOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSOmahaPokerPlayerOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_OmahaPoker {
//			logger.Logger.Error("CSOmahaPokerPlayerOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetOpParam())
//		}
//		return nil
//	}
//	return nil
//}

/////////////////////////////////////////////////////
//鱼虾蟹
////////////////////////////////////////////////////
//type CSHundredYXXOpPacketFactory struct {
//}
//type CSHundredYXXOpHandler struct {
//}
//
//func (this *CSHundredYXXOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSHundredYXXOp{}
//	return pack
//}
//
//func (this *CSHundredYXXOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSHundredYXXOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSHundredYXXOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSHundredYXXOpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

//  对战三公暗牌
//type CSSanGongHidePVPPlayerOpPacketFactory struct {
//}
//type CSSanGongHidePVPPlayerOpHandler struct {
//}
//
//func (this *CSSanGongHidePVPPlayerOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSSanGongHidePVPPlayerOp{}
//	return pack
//}
//
//func (this *CSSanGongHidePVPPlayerOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSSanGongHidePVPPlayerOpHandler Process recv ", data)
//	if op, ok := data.(*protocol.CSSanGongHidePVPPlayerOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSSanGongHidePVPPlayerOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSSanGongHidePVPPlayerOpHandler p.scene == nil")
//			return nil
//		}
//		if scene.gameId != common.GameId_SanGongHidePVP {
//			logger.Logger.Error("CSSanGongHidePVPPlayerOpHandler gameId Error ", scene.gameId)
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

/////////////////////////////////////////////////////
//财运之神
////////////////////////////////////////////////////
//type CSFortuneZhiShenOpPacketFactory struct {
//}
//type CSFortuneZhiShenOpHandler struct {
//}
//
//func (this *CSFortuneZhiShenOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSFortuneZhiShenOp{}
//	return pack
//}
//
//func (this *CSFortuneZhiShenOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSFortuneZhiShenOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSFortuneZhiShenOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSFortuneZhiShenOpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

/////////////////////////////////////////////////////
//经典777
////////////////////////////////////////////////////
//type CSClassic777OpPacketFactory struct {
//}
//type CSClassic777OpHandler struct {
//}
//
//func (this *CSClassic777OpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSClassic777Op{}
//	return pack
//}
//
//func (this *CSClassic777OpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSClassic777Op); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSClassic777OpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSClassic777OpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

/////////////////////////////////////////////////////
//金鼓齐鸣
////////////////////////////////////////////////////
//type CSGoldDrumOpPacketFactory struct {
//}
//type CSGoldDrumOpHandler struct {
//}
//
//func (this *CSGoldDrumOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSGoldDrumOp{}
//	return pack
//}
//
//func (this *CSGoldDrumOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSGoldDrumOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSGoldDrumOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSGoldDrumOpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

/////////////////////////////////////////////////////
//金福报喜
////////////////////////////////////////////////////
//type CSGoldBlessOpPacketFactory struct {
//}
//type CSGoldBlessOpHandler struct {
//}
//
//func (this *CSGoldBlessOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSGoldBlessOp{}
//	return pack
//}
//
//func (this *CSGoldBlessOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSGoldBlessOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSGoldBlessOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSGoldBlessOpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

/////////////////////////////////////////////////////
//发发发
////////////////////////////////////////////////////
//type CSClassic888OpPacketFactory struct {
//}
//type CSClassic888OpHandler struct {
//}
//
//func (this *CSClassic888OpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSClassic888Op{}
//	return pack
//}
//
//func (this *CSClassic888OpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSClassic888Op); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSClassic888OpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSClassic888OpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

/////////////////////////////////////////////////////
//无尽宝藏
////////////////////////////////////////////////////
//type CSEndlessTreasureOpPacketFactory struct {
//}
//type CSEndlessTreasureOpHandler struct {
//}
//
//func (this *CSEndlessTreasureOpPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSEndlessTreasureOp{}
//	return pack
//}
//
//func (this *CSEndlessTreasureOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if op, ok := data.(*protocol.CSEndlessTreasureOp); ok {
//		p := base.PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSEndlessTreasureOpHandler p == nil")
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSEndlessTreasureOpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.HasPlayer(p) {
//			return nil
//		}
//		if scene.sp != nil {
//			scene.sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
//		}
//		return nil
//	}
//	return nil
//}

////////////////////////////////////////////////////////////////////////////
func init() {
	common.RegisterHandler(int(gamehall.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), &CSCoinSceneOpHandler{})
	netlib.RegisterFactory(int(gamehall.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), &CSCoinSceneOpPacketFactory{})
	//百家乐
	//common.RegisterHandler(int(protocol.BaccaratPacketID_PACKET_CS_BACCARAT_PLAYEROP), &CSBaccaratOpHandler{})
	//netlib.RegisterFactory(int(protocol.BaccaratPacketID_PACKET_CS_BACCARAT_PLAYEROP), &CSBaccaratOpPacketFactory{})
	//德州扑克
	//common.RegisterHandler(int(protocol.DZPKPacketID_PACKET_CS_DEZHOUPOKER_OP), &CSDezhouPokerPlayerOpHandler{})
	//netlib.RegisterFactory(int(protocol.DZPKPacketID_PACKET_CS_DEZHOUPOKER_OP), &CSDezhouPokerPlayerOpPacketFactory{})
	//百人牛牛
	//common.RegisterHandler(int(protocol.HundredBullPacketID_PACKET_CS_PLAYEROP), &CSHundredBullOpHandler{})
	//netlib.RegisterFactory(int(protocol.HundredBullPacketID_PACKET_CS_PLAYEROP), &CSHundredBullOpPacketFactory{})
	// 21点
	//common.RegisterHandler(int(protocol.BlackJackPacketID_CS_PLAYER_OPERATE), &CSBlackJackOpHandler{})
	//netlib.RegisterFactory(int(protocol.BlackJackPacketID_CS_PLAYER_OPERATE), &CSBlackJackOpPacketFactory{})
	//梭哈
	//common.RegisterHandler(int(protocol.FiveCardStudPacketID_PACKET_CS_FiveCardStud_OP), &CSFiveCardStudOpHandler{})
	//netlib.RegisterFactory(int(protocol.FiveCardStudPacketID_PACKET_CS_FiveCardStud_OP), &CSFiveCardStudOpPacketFactory{})
	// 骰宝
	//common.RegisterHandler(int(protocol.RPPACKETID_ROLLPOINT_CS_OP), &CSRollPointHandler{})
	//netlib.RegisterFactory(int(protocol.RPPACKETID_ROLLPOINT_CS_OP), &CSRollPointPacketFactory{})
	// 百人金花
	//common.RegisterHandler(int(protocol.HundredZJHPacketID_PACKET_CS_HZJH_PLAYEROP), &CSHundredZJHOpHandler{})
	//netlib.RegisterFactory(int(protocol.HundredZJHPacketID_PACKET_CS_HZJH_PLAYEROP), &CSHundredZJHOpPacketFactory{})
	// 28
	//common.RegisterHandler(int(protocol.Hundred28GPacketID_PACKET_CS_H28G_PLAYEROP), &CSHundred28OpHandler{})
	//netlib.RegisterFactory(int(protocol.Hundred28GPacketID_PACKET_CS_H28G_PLAYEROP), &CSHundred28OpPacketFactory{})
	//飞禽走兽的操作
	//common.RegisterHandler(int(protocol.RAPACKETID_ROLLANIMALS_CS_OP), &CSRollAnimalsOpHandler{})
	//netlib.RegisterFactory(int(protocol.RAPACKETID_ROLLANIMALS_CS_OP), &CSRollAnimalsOpPacketFactory{})
	// 三公
	//common.RegisterHandler(int(protocol.SanGongPacketID_CS_SGPLAYEROP), &CSSanGongOpHandler{})
	//netlib.RegisterFactory(int(protocol.SanGongPacketID_CS_SGPLAYEROP), &CSSanGongOpPacketFactory{})
	// 德州牛仔
	//common.RegisterHandler(int(protocol.HundredDZNZPacketID_PACKET_CS_HDZNZ_PLAYEROP), &CSHundredDZNZOpHandler{})
	//netlib.RegisterFactory(int(protocol.HundredDZNZPacketID_PACKET_CS_HDZNZ_PLAYEROP), &CSHundredDZNZOpPacketFactory{})
	// 奥马哈
	//common.RegisterHandler(int(protocol.OMHPKPacketID_PACKET_CS_OMAHAPOKER_OP), &CSOmahaPokerPlayerOpHandler{})
	//netlib.RegisterFactory(int(protocol.OMHPKPacketID_PACKET_CS_OMAHAPOKER_OP), &CSOmahaPokerPlayerOpPacketFactory{})
	// 鱼虾蟹
	//common.RegisterHandler(int(protocol.HundredYXXPacketID_PACKET_CS_HYXX_PLAYEROP), &CSHundredYXXOpHandler{})
	//netlib.RegisterFactory(int(protocol.HundredYXXPacketID_PACKET_CS_HYXX_PLAYEROP), &CSHundredYXXOpPacketFactory{})
}
