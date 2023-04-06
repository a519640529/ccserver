package main

import (
	"bytes"
	"encoding/gob"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
)

func init() {

	//绑定session
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_GG_PLAYERSESSIONBIND), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.GGPlayerSessionBind{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_GG_PLAYERSESSIONBIND), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GGPlayerSessionBind:", pack)
		if msg, ok := pack.(*server.GGPlayerSessionBind); ok {
			sid := msg.GetSid()
			clientss := srvlib.ClientSessionMgrSington.GetSession(sid)
			if clientss != nil {
				clientss.SetAttribute(common.ClientSessionAttribute_GameServer, s)
			}
		}
		return nil
	}))

	//解绑定session
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_GG_PLAYERSESSIONUNBIND), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.GGPlayerSessionUnBind{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_GG_PLAYERSESSIONUNBIND), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GGPlayerSessionUnBind:", pack)
		if msg, ok := pack.(*server.GGPlayerSessionUnBind); ok {
			sid := msg.GetSid()
			clientss := srvlib.ClientSessionMgrSington.GetSession(sid)
			if clientss != nil {
				clientss.RemoveAttribute(common.ClientSessionAttribute_GameServer)
			}
		}
		return nil
	}))

	//绑定session组标记
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_SG_BINDGROUPTAG), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.SGBindGroupTag{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_SG_BINDGROUPTAG), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGBindGroupTag:", pack)
		if msg, ok := pack.(*server.SGBindGroupTag); ok {
			if len(msg.GetTags()) == 0 {
				return nil
			}

			sid := msg.GetSid()
			clientss := srvlib.ClientSessionMgrSington.GetSession(sid)
			if clientss != nil {
				var myTags []string
				param := clientss.GetAttribute(common.ClientSessionAttribute_GroupTag)
				if param != nil {
					if tags, ok := param.([]string); ok {
						myTags = tags
					}
				}
				switch msg.GetCode() {
				case server.SGBindGroupTag_OpCode_Add: //add tags
					var addTags []string
					for _, t := range msg.GetTags() {
						if !common.InSliceString(myTags, t) {
							addTags = append(addTags, t)
							myTags = append(myTags, t)
						}
					}
					if len(addTags) != 0 {
						clientss.SetAttribute(common.ClientSessionAttribute_GroupTag, myTags)
						CustomGroupMgrSington.AddToGroup(addTags, sid, clientss)
					}
				case server.SGBindGroupTag_OpCode_Del: //del tags
					var delTags []string
					var ok bool
					for _, t := range msg.GetTags() {
						if myTags, ok = common.DelSliceString(myTags, t); ok {
							delTags = append(delTags, t)
						}
					}
					if len(delTags) != 0 {
						clientss.SetAttribute(common.ClientSessionAttribute_GroupTag, myTags)
						CustomGroupMgrSington.DelFromGroup(delTags, sid)
					}
				}
			}
		}
		return nil
	}))

	//自定义组播
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_SS_CUSTOMTAG_MULTICAST), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.SSCustomTagMulticast{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_SS_CUSTOMTAG_MULTICAST), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSCustomTagMulticast:", pack)
		if msg, ok := pack.(*server.SSCustomTagMulticast); ok {
			tags := msg.GetTags()
			CustomGroupMgrSington.Broadcast(tags, msg.RawData)
		}
		return nil
	}))

	//
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WR_PlayerData), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WRPlayerData{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WR_PlayerData), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		if msg, ok := pack.(*server.WRPlayerData); ok {
			pd := &model.PlayerData{}
			buf := bytes.NewBuffer(msg.GetPlayerData())
			dec := gob.NewDecoder(buf)
			err := dec.Decode(pd)
			if err != nil {
				logger.Logger.Warnf("WRPlayerData Unmarshal %v error:%v", pd, err)
				return nil
			}
			ns := srvlib.ClientSessionMgrSington.GetSession(msg.GetSid())
			if ns != nil {
				ns.SetAttribute(common.ClientSessionAttribute_PlayerData, pd)
			}
		}
		return nil
	}))
}
