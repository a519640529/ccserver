package hundredyxx

import (
	rule "games.yol.com/win88/gamerule/hundredyxx"
	"games.yol.com/win88/protocol/hundredyxx"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

const (
	HYXX_BESTWINPOS = 6 //神算子位置
	HYXX_SELFPOS    = 7 //自己的位置
	HYXX_BANKERPOS  = 8 //庄家位置
	HYXX_OLPOS      = 9 //其他在线玩家的位置
)

type SCHundredYXXRoomInfoPacketFactory struct {
}

type SCHundredYXXRoomInfoHandler struct {
}

func (this *SCHundredYXXRoomInfoPacketFactory) CreatePacket() interface{} {
	pack := &hundredyxx.SCHundredYXXRoomInfo{}
	return pack
}

func (this *SCHundredYXXRoomInfoHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	me := base.GetUser(s)
	logger.Logger.Tracef("(this *SCHundredYXXRoomInfoHandler) Process [%v].", me.GetData().GetSnId())
	if msg, ok := pack.(*hundredyxx.SCHundredYXXRoomInfo); ok {
		isNew := false
		scene := base.SceneMgrSington.GetScene(msg.GetRoomId())
		if scene == nil {
			sceneEx := NewHundredYXXScene(msg)
			base.SceneMgrSington.AddScene(sceneEx)
			params := msg.GetParamsEx()
			if len(params) != 0 {
				sceneEx.dbGameFree = base.SceneMgrSington.GetSceneDBGameFree(sceneEx.GetRoomId(), params[0])
			}
			scene = sceneEx
			isNew = true
		} else {
			if sceneEx, ok := scene.(*HundredYXXScene); ok {
				sceneEx.UpdateInfo(msg)
			}
		}
		if scene != nil {
			s.SetAttribute(base.SessionAttributeScene, scene)
			if !isNew {
				user := base.GetUser(s)
				pd := msg.GetPlayers()
				for _, ppd := range pd {
					if ppd.GetSnId() == user.GetData().GetSnId() {
						p := NewHundredYXXPlayer(ppd)
						scene.AddPlayer(p)
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCHundredYXXRoomInfoHandler package data error.")
	}
	return nil
}

type SCHundredYXXOpPacketFactory struct {
}

type SCHundredYXXOpHandler struct {
}

func (this *SCHundredYXXOpPacketFactory) CreatePacket() interface{} {
	pack := &hundredyxx.SCHundredYXXOp{}
	return pack
}

func (this *SCHundredYXXOpHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	me := base.GetUser(s)
	logger.Logger.Tracef("(this *SCHundredYXXOpHandler) Process [%v].", me.GetData().GetName())
	if msg, ok := pack.(*hundredyxx.SCHundredYXXOp); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*HundredYXXScene); ok {
			switch int(msg.GetOpCode()) {
			case rule.HundredYXXPlayerOpUpBanker:
				if msg.GetOpRetCode() != hundredyxx.OpResultCode_OPRC_Sucess {
					delete(scene.tryUpBanker, me.GetData().GetSnId())
				}
			}
		}
	}
	return nil
}

type SCHundredYXXRoomStatePacketFactory struct {
}

type SCHundredYXXRoomStateHandler struct {
}

func (this *SCHundredYXXRoomStatePacketFactory) CreatePacket() interface{} {
	pack := &hundredyxx.SCHundredYXXRoomState{}
	return pack
}

func (this *SCHundredYXXRoomStateHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	me := base.GetUser(s)
	logger.Logger.Tracef("(this *SCHundredYXXRoomStateHandler) Process [%v].", me.GetData().GetSnId())
	if msg, ok := pack.(*hundredyxx.SCHundredYXXRoomState); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*HundredYXXScene); ok {
			scene.State = msg.State
			p := scene.GetMe(s)
			if mee, ok := p.(*HundredYXXPlayer); ok && mee != nil {
				switch int(msg.GetState()) {
				case rule.HundredYXXSceneStateSendCard:
					//if !scene.loadingBankList {
					//	scene.loadingBankList = true
					//	pack := &hundredyxx.CSHundredYXXOp{
					//		OpCode: proto.Int32(HundredYXXPlayerOpUpList),
					//	}
					//	proto.SetDefaults(pack)
					//	s.Send(int(hundredyxx.HundredYXXPacketID_PACKET_CS_HYXX_PLAYEROP), pack)
					//}
				case rule.HundredYXXSceneStateStake:
					if mee.GetSnId() == scene.GetBankerId() {
						return nil
					}
					mee.Action(s, scene)
				case rule.HundredYXXSceneStateOpenCard:
					mee.Clear()
				}
			}
		}
	} else {
		logger.Logger.Error("SCHundredYXXRoomState package data error.")
	}
	return nil
}

type SCHundredYXXGameBilledPacketFactory struct {
}

type SCHundredYXXGameBilledHandler struct {
}

func (this *SCHundredYXXGameBilledPacketFactory) CreatePacket() interface{} {
	pack := &hundredyxx.SCHundredYXXGameBilled{}
	return pack
}

func (this *SCHundredYXXGameBilledHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	me := base.GetUser(s)
	logger.Logger.Tracef("(this *SCHundbullFinalBilledHandler) Process [%v].", me.GetData().GetSnId())
	if msg, ok := pack.(*hundredyxx.SCHundredYXXGameBilled); ok {
		logger.Logger.Trace(msg)
		pds := msg.GetPlayerData()
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*HundredYXXScene); ok {
			scene.Clear(msg.GetNumOfGame())
			player := scene.GetMe(s)
			if player != nil {
				if playerEx, ok := player.(*HundredYXXPlayer); ok {
					for _, pd := range pds {
						if pd.GetPlayerID() == playerEx.GetSnId() {
							playerEx.Coin = pd.Coin
							playerEx.Pos = pd.Pos
						}
					}
					playerEx.BetTotal = 0
				}
			}
		}
	} else {
		logger.Logger.Error("SCHundbullFinalBilledHandler package data error.")
	}
	return nil
}

type SCHundredYXXOpenDicePacketFactory struct {
}

type SCHundredYXXOpenDiceHandler struct {
}

func (this *SCHundredYXXOpenDicePacketFactory) CreatePacket() interface{} {
	pack := &hundredyxx.SCHundredYXXOpenDice{}
	return pack
}

func (this *SCHundredYXXOpenDiceHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	me := base.GetUser(s)
	logger.Logger.Tracef("(this *SCHundredYXXOpenDiceHandler) Process [%v].", me.GetData().GetSnId())
	if msg, ok := pack.(*hundredyxx.SCHundredYXXOpenDice); ok {
		logger.Logger.Trace(msg)
		scene := base.GetScene(s)
		if scene != nil {
			if sceneEx, ok := scene.(*HundredYXXScene); ok {
				sceneEx.DicePoints = msg.DicePoints
			}
		}
	} else {
		logger.Logger.Error("SCHundredYXXOpenDice package data error.")
	}
	return nil
}

type SCHundredYXXBetPacketFactory struct {
}

type SCHundredYXXBetHandler struct {
}

func (this *SCHundredYXXBetPacketFactory) CreatePacket() interface{} {
	pack := &hundredyxx.SCHundredYXXBet{}
	return pack
}

func (this *SCHundredYXXBetHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	me := base.GetUser(s)
	logger.Logger.Tracef("(this *SCHundredYXXBetHandler) Process [%v].", me.GetData().GetSnId())
	if msg, ok := pack.(*hundredyxx.SCHundredYXXBet); ok {
		logger.Logger.Trace(msg)
		me := base.GetUser(s)
		if msg.GetSnId() == me.GetData().GetSnId() {
			scene := base.GetScene(s)
			if scene != nil {
				player := scene.GetMe(s)
				if player != nil {
					if playerEx, ok := player.(*HundredYXXPlayer); ok {
						playerEx.BetTotal = msg.BetTotal
						playerEx.Coin = msg.Coin
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCHundredYXXOpenDice package data error.")
	}
	return nil
}

type SCHundredYXXSyncChipPacketFactory struct {
}

type SCHundredYXXSyncChipHandler struct {
}

func (this *SCHundredYXXSyncChipPacketFactory) CreatePacket() interface{} {
	pack := &hundredyxx.SCHundredYXXSyncChip{}
	return pack
}

func (this *SCHundredYXXSyncChipHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	if msg, ok := pack.(*hundredyxx.SCHundredYXXSyncChip); ok {
		logger.Logger.Trace(msg)
	} else {
		logger.Logger.Error("SCHundredYXXSyncChip package data error.")
	}
	return nil
}

type SCHundredYXXSeatsPacketFactory struct {
}

type SCHundredYXXSeatsHandler struct {
}

func (this *SCHundredYXXSeatsPacketFactory) CreatePacket() interface{} {
	pack := &hundredyxx.SCHundredYXXSeats{}
	return pack
}

func (this *SCHundredYXXSeatsHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	if msg, ok := pack.(*hundredyxx.SCHundredYXXSeats); ok {
		logger.Logger.Trace(msg)
		scene := base.GetScene(s)
		if scene != nil {
			player := scene.GetMe(s)
			if player != nil {
				if playerEx, ok := player.(*HundredYXXPlayer); ok {
					for _, data := range msg.GetData() {
						if data.GetSnId() == playerEx.GetSnId() {
							playerEx.Pos = data.Pos
							playerEx.Coin = data.Coin
							break
						}
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCHundredYXXSeats package data error.")
	}
	return nil
}

type SCHundredYXXUpListPacketFactory struct {
}

type SCHundredYXXUpListHandler struct {
}

func (this *SCHundredYXXUpListPacketFactory) CreatePacket() interface{} {
	pack := &hundredyxx.SCHundredYXXUpList{}
	return pack
}

func (this *SCHundredYXXUpListHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	if msg, ok := pack.(*hundredyxx.SCHundredYXXUpList); ok {
		logger.Logger.Trace(msg)
		scene := base.GetScene(s)
		if scene != nil {
			if sceneEx, ok := scene.(*HundredYXXScene); ok {
				sceneEx.waitingBankerNum = msg.GetCount()
				sceneEx.tryUpBanker = make(map[int32]*HundredYXXPlayer)
			}
		}
	} else {
		logger.Logger.Error("SCHundredYXXUpList package data error.")
	}
	return nil
}

func init() {
	//SCHundredYXXRoomInfo
	netlib.RegisterHandler(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_ROOMINFO), &SCHundredYXXRoomInfoHandler{})
	netlib.RegisterFactory(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_ROOMINFO), &SCHundredYXXRoomInfoPacketFactory{})
	//SCHundredYXXOp
	netlib.RegisterHandler(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_PLAYEROP), &SCHundredYXXOpHandler{})
	netlib.RegisterFactory(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_PLAYEROP), &SCHundredYXXOpPacketFactory{})
	//SCHundredYXXRoomState
	netlib.RegisterHandler(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_ROOMSTATE), &SCHundredYXXRoomStateHandler{})
	netlib.RegisterFactory(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_ROOMSTATE), &SCHundredYXXRoomStatePacketFactory{})
	//SCHundredYXXGameBilled
	netlib.RegisterHandler(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_GAMEBILLED), &SCHundredYXXGameBilledHandler{})
	netlib.RegisterFactory(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_GAMEBILLED), &SCHundredYXXGameBilledPacketFactory{})
	//SCHundredYXXOpenDice
	netlib.RegisterHandler(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_OPENDICE), &SCHundredYXXOpenDiceHandler{})
	netlib.RegisterFactory(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_OPENDICE), &SCHundredYXXOpenDicePacketFactory{})
	//SCHundredYXXBet
	netlib.RegisterHandler(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_BET), &SCHundredYXXBetHandler{})
	netlib.RegisterFactory(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_BET), &SCHundredYXXBetPacketFactory{})
	//SCHundredYXXSyncChip
	netlib.RegisterHandler(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_SYNCCHIP), &SCHundredYXXSyncChipHandler{})
	netlib.RegisterFactory(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_SYNCCHIP), &SCHundredYXXSyncChipPacketFactory{})
	//SCHundredYXXSeats
	netlib.RegisterHandler(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_SEATS), &SCHundredYXXSeatsHandler{})
	netlib.RegisterFactory(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_SEATS), &SCHundredYXXSeatsPacketFactory{})
	//SCHundredYXXUpList
	netlib.RegisterHandler(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_UPLIST), &SCHundredYXXUpListHandler{})
	netlib.RegisterFactory(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_UPLIST), &SCHundredYXXUpListPacketFactory{})
}
