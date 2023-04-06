package base

//
//import (
//	match_proto "games.yol.com/win88/protocol/match"
//	player_proto "games.yol.com/win88/protocol/player"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/idealeak/goserver/core/netlib"
//)
//
//func init() {
//
//	netlib.RegisterFactory(int(match_proto.MatchPacketID_PACKET_SC_MATCH_SIGNUP), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCMatchSignup{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MatchPacketID_PACKET_SC_MATCH_SIGNUP), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCMahjongMatchSignup:", pack)
//		if scMatchSignup, ok := pack.(*match_proto.SCMatchSignup); ok {
//			if scMatchSignup.GetOpCode() != match_proto.OpResultCode_OPRC_Sucess {
//				matchid := scMatchSignup.GetMatchId()
//				hmatchid := s.GetAttribute(SessionAttributeWaitingMatch)
//				if id, ok := hmatchid.(int32); ok && id == matchid {
//					s.RemoveAttribute(SessionAttributeWaitingMatch)
//				}
//				if strategy, ok := s.GetAttribute(SessionAttributeStrategy).(int); ok {
//					ChooiesStrategy(s, strategy)
//				}
//			} else {
//				matchid := scMatchSignup.GetMatchId()
//				if scMatchSignup.GetNeedAwait() {
//					s.SetAttribute(SessionAttributeWaitingMatch, matchid)
//				}
//			}
//		} else {
//			logger.Logger.Info("receive SCMatchSignup error.")
//		}
//		return nil
//	}))
//
//	netlib.RegisterFactory(int(match_proto.MatchPacketID_PACKET_SC_MATCH_CANCELSIGNUP), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCMatchCancelSignup{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MatchPacketID_PACKET_SC_MATCH_CANCELSIGNUP), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCMatchCancelSignup:", pack)
//		if msg, ok := pack.(*match_proto.SCMatchCancelSignup); ok {
//			if msg.GetOpCode() == match_proto.OpResultCode_OPRC_Sucess {
//				matchid := msg.GetMatchId()
//				hmatchid := s.GetAttribute(SessionAttributeWaitingMatch)
//				if id, ok := hmatchid.(int32); ok && id == matchid {
//					s.RemoveAttribute(SessionAttributeWaitingMatch)
//				}
//				if strategy, ok := s.GetAttribute(SessionAttributeStrategy).(int); ok {
//					ChooiesStrategy(s, strategy)
//				}
//			}
//		}
//		return nil
//	}))
//
//	netlib.RegisterFactory(int(match_proto.MatchPacketID_PACKET_SC_MATCH_ACHIEVEMENT), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCMatchAchievement{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MatchPacketID_PACKET_SC_MATCH_ACHIEVEMENT), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCMatchAchievement:", pack)
//		return nil
//	}))
//
//	netlib.RegisterFactory(int(match_proto.MatchPacketID_PACKET_SC_MATCH_PHASE), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCMatchPhase{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MatchPacketID_PACKET_SC_MATCH_PHASE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCMatchPhase:", pack)
//		if msg, ok := pack.(*match_proto.SCMatchPhase); ok {
//			if msg.GetIsOut() {
//				s.RemoveAttribute(SessionAttributeMatchDoing)
//				s.RemoveAttribute(SessionAttributeScene)
//				s.RemoveAttribute(SessionAttributeSceneId)
//				matchid := msg.GetMatchId()
//				hmatchid := s.GetAttribute(SessionAttributeWaitingMatch)
//				if id, ok := hmatchid.(int32); ok && id == matchid {
//					s.RemoveAttribute(SessionAttributeWaitingMatch)
//				}
//				if strategy, ok := s.GetAttribute(SessionAttributeStrategy).(int); ok {
//					ChooiesStrategy(s, strategy)
//				}
//			}
//		}
//		return nil
//	}))
//
//	netlib.RegisterFactory(int(match_proto.MatchPacketID_PACKET_SC_MATCH_REFRESHRANK), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCMatchRefreshRank{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MatchPacketID_PACKET_SC_MATCH_REFRESHRANK), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCMatchRefreshRank:", pack)
//		return nil
//	}))
//
//	netlib.RegisterFactory(int(match_proto.MatchPacketID_PACKET_SC_MATCH_BASESCORECHANGE), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCMatchBaseScoreChange{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MatchPacketID_PACKET_SC_MATCH_BASESCORECHANGE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCMatchBaseScoreChange:", pack)
//		return nil
//	}))
//
//	netlib.RegisterFactory(int(match_proto.MatchPacketID_PACKET_SC_MATCH_DATACHANGE), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCMatchDataChange{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MatchPacketID_PACKET_SC_MATCH_DATACHANGE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCMatchDataChange:", pack)
//		return nil
//	}))
//
//	netlib.RegisterFactory(int(match_proto.MatchPacketID_PACKET_SC_MATCH_START), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCMatchStart{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MatchPacketID_PACKET_SC_MATCH_START), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCMatchStart:", pack)
//		return nil
//	}))
//
//	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERCOINCHANGE), netlib.PacketFactoryWrapper(func() interface{} {
//		return &player_proto.SCPlayerCoinChange{}
//	}))
//	netlib.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERCOINCHANGE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		return nil
//	}))

//	netlib.RegisterFactory(int(match_proto.MmoPacketID_PACKET_SC_TASKLIST), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCTaskList{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MmoPacketID_PACKET_SC_TASKLIST), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCTaskList:", pack)
//		return nil
//	}))

//	netlib.RegisterFactory(int(match_proto.MmoPacketID_PACKET_SC_TASKCHG), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCTaskChg{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MmoPacketID_PACKET_SC_TASKCHG), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCTaskChg:", pack)
//		return nil
//	}))

//	netlib.RegisterFactory(int(match_proto.MmoPacketID_PACKET_SC_TACKCOMPLETE), netlib.PacketFactoryWrapper(func() interface{} {
//		return &match_proto.SCTaskComplete{}
//	}))
//	netlib.RegisterHandler(int(match_proto.MmoPacketID_PACKET_SC_TACKCOMPLETE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
//		logger.Logger.Trace("receive SCTaskComplete:", pack)
//		return nil
//	}))
//}
