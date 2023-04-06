package rollpoint

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/rollpoint"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

func init() {
	//SCRollPointRoomInfo
	netlib.RegisterHandler(int(rollpoint.RPPACKETID_ROLLPOINT_SC_ROOMINFO), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, data interface{}) error {
		//logger.Logger.Info("(this *SCRollPointRoomInfo) Process", data)
		if msg, ok := data.(*rollpoint.SCRollPointRoomInfo); ok {
			scene := base.SceneMgrSington.GetScene(msg.GetSceneId())
			if scene == nil {
				scene = &RollPointScene{
					SCRollPointRoomInfo: msg,
					Players:             make(map[int32]base.Player),
				}
				base.SceneMgrSington.AddScene(scene)
			}
			scene.AddPlayer(&RollPointPlayerData{
				RollPointPlayer: msg.GetPlayer(),
			})
			s.SetAttribute(base.SessionAttributeScene, scene)
			if RollPointScene, ok := scene.(*RollPointScene); ok {
				params := RollPointScene.GetParamsEx()
				if len(params) != 0 {
					RollPointScene.dbGameFree = base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
				}
				RollPointScene.BankerId = 0
			}
		}
		return nil
	}))
	netlib.RegisterFactory(int(rollpoint.RPPACKETID_ROLLPOINT_SC_ROOMINFO), netlib.PacketFactoryWrapper(func() interface{} {
		return &rollpoint.SCRollPointRoomInfo{}
	}))
	//SCRollPointRoomState
	netlib.RegisterHandler(int(rollpoint.RPPACKETID_ROLLPOINT_SC_ROOMSTATE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, data interface{}) error {
		//logger.Logger.Trace("(this *SCRollPointRoomState) Process:", data)
		if msg, ok := data.(*rollpoint.SCRollPointRoomState); ok {
			if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*RollPointScene); ok {
				scene.State = msg.State
				p := scene.GetMe(s)
				if me, ok := p.(*RollPointPlayerData); ok {
					switch msg.GetState() {
					case RollPointSceneStateWait: //过场
					case RollPointSceneStateStart: //过场
						scene.Clear()
					case RollPointSceneStateBet: //押注
						scene.RollPoint(s, me)
					case RollPointSceneStateBilled: //结算
						scene.Clear()
						me.Clear()
					}
				} else {
					logger.Logger.Info("Get player data failed.")
				}
			} else {
				logger.Logger.Info("Get roll coin scene data failed.")
			}
		}
		return nil
	}))
	netlib.RegisterFactory(int(rollpoint.RPPACKETID_ROLLPOINT_SC_ROOMSTATE), netlib.PacketFactoryWrapper(func() interface{} {
		return &rollpoint.SCRollPointRoomState{}
	}))
	//SCRollPointBill
	netlib.RegisterHandler(int(rollpoint.RPPACKETID_ROLLPOINT_SC_BILL), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, data interface{}) error {
		//logger.Logger.Info("(this *SCRollPointBill) Process", data)
		if msg, ok := data.(*rollpoint.SCRollPointBill); ok {
			if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*RollPointScene); ok {
				p := scene.GetMe(s)
				if me, ok := p.(*RollPointPlayerData); ok {
					me.Coin = proto.Int64(me.GetCoin() + int64(msg.GetCoin()))
				}
			}
		}
		return nil
	}))
	netlib.RegisterFactory(int(rollpoint.RPPACKETID_ROLLPOINT_SC_BILL), netlib.PacketFactoryWrapper(func() interface{} {
		return &rollpoint.SCRollPointBill{}
	}))

	//SCRollPointCoinLog
	netlib.RegisterHandler(int(rollpoint.RPPACKETID_ROLLPOINT_SC_COINLOG), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, data interface{}) error {
		return nil
	}))
	netlib.RegisterFactory(int(rollpoint.RPPACKETID_ROLLPOINT_SC_COINLOG), netlib.PacketFactoryWrapper(func() interface{} {
		return &rollpoint.SCRollPointCoinLog{}
	}))

	//SCPUSHCOIN
	netlib.RegisterHandler(int(rollpoint.RPPACKETID_ROLLPOINT_SC_PUSHCOIN), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, data interface{}) error {
		return nil
	}))
	netlib.RegisterFactory(int(rollpoint.RPPACKETID_ROLLPOINT_SC_PUSHCOIN), netlib.PacketFactoryWrapper(func() interface{} {
		return &rollpoint.SCRollPointPushCoin{}
	}))
	//SCPLAYERNUM
	netlib.RegisterHandler(int(rollpoint.RPPACKETID_ROLLPOINT_SC_PLAYERNUM), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, data interface{}) error {
		return nil
	}))
	netlib.RegisterFactory(int(rollpoint.RPPACKETID_ROLLPOINT_SC_PLAYERNUM), netlib.PacketFactoryWrapper(func() interface{} {
		return &rollpoint.SCRollPointPlayerNum{}
	}))
}
