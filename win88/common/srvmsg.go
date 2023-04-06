package common

import (
	"fmt"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/player"
)

const (
	SRVMSG_CODE_DEFAULT int32 = iota
)

type SrvMsgSender interface {
	SendToClient(packetid int, rawpack interface{}) bool
}

func SendSrvMsg(sender SrvMsgSender, msgId int32, params ...interface{}) bool {
	pack := CreateSrvMsg(msgId, params)
	return sender.SendToClient(int(player.PlayerPacketID_PACKET_SC_SRVMSG), pack)
}

func CreateSrvMsg(msgId int32, params ...interface{}) *player.SCSrvMsg {
	pack := &player.SCSrvMsg{
		MsgId: msgId,
	}
	for _, p := range params {
		switch d := p.(type) {
		case string:
			pack.Params = append(pack.Params, &player.SrvMsgParam{StrParam: d})
		case int:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		case int8:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		case int16:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		case int32:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: d})
		case int64:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		case uint:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		case uint8:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		case uint16:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		case uint32:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		case uint64:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		case float32:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		case float64:
			pack.Params = append(pack.Params, &player.SrvMsgParam{IntParam: int32(d)})
		default:
			pack.Params = append(pack.Params, &player.SrvMsgParam{StrParam: fmt.Sprintf("%v", p)})
		}
	}
	proto.SetDefaults(pack)
	return pack
}
