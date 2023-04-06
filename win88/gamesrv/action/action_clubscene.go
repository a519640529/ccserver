package action

//type CSClubEnterRoomPacketFactory struct {
//}
//type CSClubEnterRoomHandler struct {
//}
//
//func (this *CSClubEnterRoomPacketFactory) CreatePacket() interface{} {
//	pack := &protocol.CSClubEnterRoom{}
//	return pack
//}
//
//func (this *CSClubEnterRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	if msg, ok := data.(*protocol.CSClubEnterRoom); ok {
//		logger.Logger.Trace("CSClubEnterRoomHandler ", msg)
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSCoinSceneOpHandler p == nil", data)
//			return nil
//		}
//		scene := p.scene
//		if scene == nil {
//			logger.Logger.Warn("CSCoinSceneOpHandler p.scene == nil")
//			return nil
//		}
//		if !scene.IsCoinScene() {
//			return nil
//		}
//		switch msg.GetOpType() { //离开
//		case common.CoinSceneOp_Leave:
//			if !scene.HasPlayer(p) {
//				return nil
//			}
//			if scene.CanChangeCoinScene(p) {
//				scene.PlayerLeave(p, common.PlayerLeaveReason_Normal, true)
//				return nil
//			}
//		case common.CoinSceneOp_DownRiceLeave:
//			if !scene.HasAudience(p) {
//				return nil
//			}
//			if scene.CanChangeCoinScene(p) {
//				scene.AudienceLeave(p, common.PlayerLeaveReason_Normal)
//				return nil
//			}
//		}
//		return nil
//	}
//	return nil
//}
//
//func init() {
//	common.RegisterHandler(int(protocol.ClubPacketID_PACKET_CS_CLUB_ENTERROOM), &CSClubEnterRoomHandler{})
//	netlib.RegisterFactory(int(protocol.ClubPacketID_PACKET_CS_CLUB_ENTERROOM), &CSClubEnterRoomPacketFactory{})
//}
