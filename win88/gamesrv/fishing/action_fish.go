package fishing

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	fishing_proto "games.yol.com/win88/protocol/fishing"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type ErrorString struct {
	code string
}

func (this *ErrorString) Error() string {
	return this.code
}

//装载完毕
type CSLoadCompletePacketFactory struct {
}
type CSLoadCompleteHandler struct {
}

func (this *CSLoadCompletePacketFactory) CreatePacket() interface{} {
	pack := &fishing_proto.CSLoadComplete{}
	return pack
}
func (this *CSLoadCompleteHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	_, ok := data.(*fishing_proto.CSLoadComplete)
	if ok == false {
		return &ErrorString{code: "CSLoadComplete serialize error."}
	}
	player := base.PlayerMgrSington.GetPlayer(sid)
	if player != nil {
		if player.GetScene() != nil {
			//fsd := player.GetScene().GetExtraData().(*FishingSceneData)
			//if fsd != nil {
			//	fsd.SyncFish(player)
			//}
		}
	}
	return nil
}

//捕鱼
type CSFishingOpPacketFactory struct {
}
type CSFishingOpHandler struct {
}

func (this *CSFishingOpPacketFactory) CreatePacket() interface{} {
	pack := &fishing_proto.CSFishingOp{}
	return pack
}
func (this *CSFishingOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	//logger.Logger.Trace("CSFishingOpHandler Process recv ", data)
	if msg, ok := data.(*fishing_proto.CSFishingOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSFishingOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSFishingOpHandler p.GetScene() == nil")
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		if scene.GetScenePolicy() != nil {
			scene.GetScenePolicy().OnPlayerOp(scene, p, int(msg.GetOpCode()), msg.GetParams())
		}
		return nil
	}
	return nil
}

//捕鱼
type CSFishViewPacketFactory struct {
}
type CSFishViewHandler struct {
}

func (this *CSFishViewPacketFactory) CreatePacket() interface{} {
	pack := &fishing_proto.CSFishView{}
	return pack
}
func (this *CSFishViewHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSFishViewHandler Process recv ", data)
	if msg, ok := data.(*fishing_proto.CSFishView); ok {
		var player *base.Player
		snid := msg.GetSnId()
		if snid == 0 {
			player = base.PlayerMgrSington.GetPlayer(sid)
			if player == nil {
				logger.Logger.Warn("CSFishViewHandler GetPlayer p == nil")
				return nil
			}
		} else {
			player = base.PlayerMgrSington.GetPlayerBySnId(snid)
			if player == nil {
				logger.Logger.Warn("CSFishViewHandler GetPlayerBySnId p == nil")
				return nil
			}
			if !player.IsRob && player != base.PlayerMgrSington.GetPlayer(sid) {
				logger.Logger.Warn("CSFishViewHandler !player.IsRob && player!=base.PlayerMgrSington.GetPlayer(sid)", player.SnId)
				return nil
			}
		}

		if player.GetScene() != nil {
			if fsd, ok := player.GetScene().GetExtraData().(*FishingSceneData); ok && fsd != nil {
				fp := fsd.players[player.SnId]
				if fp != nil {
					fsd.PushEventFish(fp, msg.GetSign(), msg.GetFishs(), msg.GetEventFish())
				}
			}
		}
		return nil
	}
	return nil
}

//瞄准
type CSFishTargetPacketFactory struct {
}
type CSFishTargetHandler struct {
}

func (this *CSFishTargetPacketFactory) CreatePacket() interface{} {
	pack := &fishing_proto.CSFishTarget{}
	return pack
}
func (this *CSFishTargetHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSFishTargetHandler Process recv ", data)
	if msg, ok := data.(*fishing_proto.CSFishTarget); ok {
		var player *base.Player
		if msg.GetRobotId() != 0 {
			player = base.PlayerMgrSington.GetPlayerBySnId(msg.GetRobotId())
		}
		if player == nil {
			logger.Logger.Warn("CSFishViewHandler robot == nil")
			return nil
			//player = base.PlayerMgrSington.GetPlayer(sid)
		}
		if player == nil {
			logger.Logger.Warn("CSFishViewHandler p == nil")
			return nil
		}
		if player.GetScene() != nil {
			if playerEx, ok := player.GetExtraData().(*FishingPlayerData); ok {
				playerEx.TargetFish = msg.GetFishId()
			}
			pack := &fishing_proto.SCFishTarget{
				FishId: proto.Int32(msg.GetFishId()),
				SnId:   proto.Int32(player.SnId),
			}
			player.GetScene().Broadcast(int(fishing_proto.FIPacketID_FISHING_SC_FISHTARGET), pack, 0)
		}
		return nil
	}
	return nil
}

//查看打死鱼
type CSFishLookLockPacketFactory struct {
}
type CSFishLookLockHandler struct {
}

func (this *CSFishLookLockPacketFactory) CreatePacket() interface{} {
	pack := &fishing_proto.CSLookLockFish{}
	return pack
}
func (this *CSFishLookLockHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSFishLookLockHandler Process recv ", data)
	if _, ok := data.(*fishing_proto.CSLookLockFish); ok {
		player := base.PlayerMgrSington.GetPlayer(sid)
		if player == nil {
			logger.Logger.Warn("CSFishLookLockHandler p == nil")
			return nil
		}
		if player.GetScene() != nil {
			if playerEx, ok := player.GetExtraData().(*FishingPlayerData); ok {

				pack := &fishing_proto.SCLookLockFish{}
				for id, v := range playerEx.lockFishCount {
					pack.FishId = append(pack.FishId, id)
					pack.FishIdNum = append(pack.FishIdNum, v)
				}
				logger.Logger.Trace("CSFishLookLockHandler Pb  ", pack)
				playerEx.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_LOOKLOCKFISH), pack)
			}
		}
		return nil
	}
	return nil
}

// 放炮台
type CSFishReadyPranaPacketFactory struct {
}
type CSFishReadyPranaHandler struct {
}

func (this *CSFishReadyPranaPacketFactory) CreatePacket() interface{} {
	pack := &fishing_proto.CSReadyPrana{}
	return pack
}
func (this *CSFishReadyPranaHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSFishReadyPranaHandler Process recv ", data)
	if msg, ok := data.(*fishing_proto.CSReadyPrana); ok {
		player := base.PlayerMgrSington.GetPlayer(sid)
		if player == nil {
			logger.Logger.Warn("CSFishReadyPranaHandler p == nil")
			return nil
		}
		if player.GetScene() != nil {
			if playerEx, ok := player.GetExtraData().(*FishingPlayerData); ok {

				pack := &fishing_proto.SCReadyPrana{
					SnId: proto.Int32(playerEx.SnId),
					X:    proto.Int32(msg.GetX()),
					Y:    proto.Int32(msg.GetY()),
				}
				if playerEx.PranaPercent == 100 {
					pack.OpRetCode = fishing_proto.OpResultCode_OPRC_Sucess
				} else {
					pack.OpRetCode = fishing_proto.OpResultCode_OPRC_Error // 操作失败
				}

				logger.Logger.Trace("CSFishReadyPranaHandler Pb  ", pack)
				playerEx.GetScene().Broadcast(int(fishing_proto.FIPacketID_FISHING_SC_REALYPRANA), pack, 0)
			}
		}
		return nil
	}
	return nil
}

// 发射能量炮
type CSFishFirePranaPacketFactory struct {
}
type CSFishFirePranaHandler struct {
}

func (this *CSFishFirePranaPacketFactory) CreatePacket() interface{} {
	pack := &fishing_proto.CSFirePrana{}
	return pack
}
func (this *CSFishFirePranaHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSFishFirePranaHandler Process recv ", data)
	if msg, ok := data.(*fishing_proto.CSFirePrana); ok {
		player := base.PlayerMgrSington.GetPlayer(sid)
		if player == nil {
			logger.Logger.Warn("CSFishFirePranaHandler p == nil")
			return nil
		}
		if player.GetScene() != nil {
			if sx, ok := player.GetScene().GetExtraData().(*HPFishingSceneData); ok && sx != nil {
				if playerEx, ok := player.GetExtraData().(*FishingPlayerData); ok {

					pack := &fishing_proto.SCReadyPrana{
						SnId: proto.Int32(playerEx.SnId),
						X:    proto.Int32(msg.GetX()),
						Y:    proto.Int32(msg.GetY()),
					}
					if playerEx.PranaPercent == 100 {
						pack.OpRetCode = fishing_proto.OpResultCode_OPRC_Sucess
						var fishs []int
						for _, v := range msg.FishIds {
							fish := sx.fish_list[v]
							if fish != nil {
								fish.OnHit(0)
								fishs = append(fishs, int(v))
							}
						}
						// TODO 死鱼操作
						sx.firePranaDel(playerEx, fishs, int32(playerEx.Prana))
						playerEx.Prana = 0.0
						playerEx.PranaPercent = 0
					} else {
						pack.OpRetCode = fishing_proto.OpResultCode_OPRC_Error // 操作失败
					}

					logger.Logger.Trace("CSFishFirePranaHandler Pb  ", pack)
					playerEx.GetScene().Broadcast(int(fishing_proto.FIPacketID_FISHING_SC_FIREPRANA), pack, 0)
				}
			}
		}
		return nil
	}
	return nil
}

func init() {
	//装载完毕
	common.RegisterHandler(int(fishing_proto.FIPacketID_FISHING_CS_LOADCOMPLETE), &CSLoadCompleteHandler{})
	netlib.RegisterFactory(int(fishing_proto.FIPacketID_FISHING_CS_LOADCOMPLETE), &CSLoadCompletePacketFactory{})
	//捕鱼
	common.RegisterHandler(int(fishing_proto.FIPacketID_FISHING_CS_OP), &CSFishingOpHandler{})
	netlib.RegisterFactory(int(fishing_proto.FIPacketID_FISHING_CS_OP), &CSFishingOpPacketFactory{})
	//事件
	common.RegisterHandler(int(fishing_proto.FIPacketID_FISHING_CS_FISHVIEW), &CSFishViewHandler{})
	netlib.RegisterFactory(int(fishing_proto.FIPacketID_FISHING_CS_FISHVIEW), &CSFishViewPacketFactory{})
	//瞄准
	common.RegisterHandler(int(fishing_proto.FIPacketID_FISHING_CS_FISHTARGET), &CSFishTargetHandler{})
	netlib.RegisterFactory(int(fishing_proto.FIPacketID_FISHING_CS_FISHTARGET), &CSFishTargetPacketFactory{})
	//查看打死鱼
	common.RegisterHandler(int(fishing_proto.FIPacketID_FISHING_CS_LOOKLOCKFISH), &CSFishLookLockHandler{})
	netlib.RegisterFactory(int(fishing_proto.FIPacketID_FISHING_CS_LOOKLOCKFISH), &CSFishLookLockPacketFactory{})
	//放置炮台
	common.RegisterHandler(int(fishing_proto.FIPacketID_FISHING_CS_REALYPRANA), &CSFishReadyPranaHandler{})
	netlib.RegisterFactory(int(fishing_proto.FIPacketID_FISHING_CS_REALYPRANA), &CSFishReadyPranaPacketFactory{})
	//发射能量炮
	common.RegisterHandler(int(fishing_proto.FIPacketID_FISHING_CS_FIREPRANA), &CSFishFirePranaHandler{})
	netlib.RegisterFactory(int(fishing_proto.FIPacketID_FISHING_CS_FIREPRANA), &CSFishFirePranaPacketFactory{})
	//查询中奖
	//common.RegisterHandler(int(fishing_proto.FIPacketID_FISHING_CS_JACKPOTLIST), &CSFishJackpotHandler{})
	//netlib.RegisterFactory(int(fishing_proto.FIPacketID_FISHING_CS_JACKPOTLIST), &CSFishJackpotPacketFactory{})
}
