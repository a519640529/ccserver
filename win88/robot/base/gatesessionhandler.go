package base

import (
	player_proto "games.yol.com/win88/protocol/player"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

const (
	GateSessionHandlerName              = "handler-gate-session"
	SessionAttributeClientAccountId int = iota
	SessionAttributeUser
	SessionAttributeScene
	SessionAttributeSceneId
	SessionAttributeStrategy
	SessionAttributeTimer
	SessionAttributeGameTimer
	SessionAttributePingTimer
	SessionAttributeLoginTimer
	SessionAttributeDelAccount
	SessionAttributeCoinSceneQueue
	SessionAttributeWaitingMatch
	SessionAttributeMatchDoing
)

type GateSessionHandler struct {
	netlib.BasicSessionHandler
}

func (sfcl GateSessionHandler) GetName() string {
	return GateSessionHandlerName
}

func (this *GateSessionHandler) GetInterestOps() uint {
	return 1<<netlib.InterestOps_Opened | 1<<netlib.InterestOps_Closed
}

func (this *GateSessionHandler) OnSessionOpened(s *netlib.Session) {
	for accId, _ := range accountChan {
		logger.Logger.Infof("GateSessionHandler OnSessionOpened: accid:%v sid:%v accid:%v", clientArray[accId], s.Id, accId)
		s.SetAttribute(SessionAttributeClientAccountId, accId)
		ClientMgrSington.RegisteSession(accId, s)
		logger.Logger.Info("CSLogin send complete.")
		delete(accountChan, accId)
		return
	}
	logger.Logger.Info("Accid read failed from accountChan.")
	s.Close()
	return
}

func (this *GateSessionHandler) OnSessionClosed(s *netlib.Session) {
	accIdParam := s.GetAttribute(SessionAttributeClientAccountId)
	isDelAcc := s.GetAttribute(SessionAttributeDelAccount)
	reconnect := false
	if accId, ok := accIdParam.(string); ok {
		logger.Logger.Warnf("GateSessionHandler OnSessionClosed accid:%v sid:%v accid:%v", clientArray[accId], s.Id, accId)
		if isDelAcc == nil {
			accountChan[accId] = true
		}

		ClientMgrSington.UnRegisteSession(accId)
		logger.Logger.Warn("AccountChan len:", len(accountChan))
		if isDelAcc == nil {
			reconnect = true
		}
	}
	scPlayerInfoParam := s.GetAttribute(SessionAttributeUser)
	if scPlayerInfo, ok := scPlayerInfoParam.(*player_proto.SCPlayerData); ok {
		PlayerMgrSington.DelPlayer(scPlayerInfo.GetData().GetSnId())
	}
	StopSessionPingTimer(s)
	if reconnect {
		cfg := Config.Connects
		cfg.Id = s.GetSessionConfig().Id
		logger.Logger.Info("ReStart sno ", cfg.Id, " Client Connect.")
		cfg.Init()
		WaitReconnectCfg = append(WaitReconnectCfg, &cfg)
	}
}

func init() {
	netlib.RegisteSessionHandlerCreator(GateSessionHandlerName, func() netlib.SessionHandler {
		return &GateSessionHandler{}
	})
}
