package main

import (
	"games.yol.com/win88/common"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	"games.yol.com/win88/ranksrv/rank"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"strconv"
)

//请求排行榜信息
type CSGetRankInfoPacketFactory struct {
}
type CSGetRankInfoHandler struct {
}

func (this *CSGetRankInfoPacketFactory) CreatePacket() interface{} {
	pack := &hall_proto.CSGetRankInfo{}
	return pack
}

func (this *CSGetRankInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGetRankInfoHandler Process recv ", data)
	if msg, ok := data.(*hall_proto.CSGetRankInfo); ok {
		pack := &hall_proto.SCGetRankInfo{}
		key := msg.GetPlt()
		gameFreeId := int64(msg.GetGameFreeId())

		if rankMap, ok := rank.RankMgrSignton.PltRankMap[key]; ok {
			if node, ok := rankMap.GameRank[strconv.Itoa(int(gameFreeId))]; ok {
				for i := 0; i < len(node.Data); i++ {
					data := node.Data[i]
					if data != nil {
						pack.Info = append(pack.Info, &hall_proto.RankInfo{
							Snid:     data.SnId,
							Name:     data.Name,
							TotalIn:  0,
							TotalOut: data.Val,
						})
					}
				}
			}
		}
		common.SendToGate(sid, int(hall_proto.HallPacketID_PACKET_SC_GETRANKINFO), pack, s)
	}
	return nil
}

func init() {
	//请求排行榜
	common.RegisterHandler(int(hall_proto.HallPacketID_PACKET_CS_GETRANKINFO), &CSGetRankInfoHandler{})
	netlib.RegisterFactory(int(hall_proto.HallPacketID_PACKET_CS_GETRANKINFO), &CSGetRankInfoPacketFactory{})
}
