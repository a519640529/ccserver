package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/chat"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
	"html"
	"time"
)

// 用户发送消息
type CSChatMsgPacketFactory struct {
}
type CSChatMsgHandler struct {
}

func (this *CSChatMsgPacketFactory) CreatePacket() interface{} {
	pack := &chat.CSChatMsg{}
	return pack
}

func (this *CSChatMsgHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSChatMsgHandler Process recv ", data)
	if msg, ok := data.(*chat.CSChatMsg); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSChatMsgHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			logger.Logger.Warn("CSChatMsgHandler platform == nil")
			return nil
		}
		var content = msg.GetContent()
		if len(content) > 150 {
			logger.Logger.Warn("CSChatMsgHandler len(content) > 150")
			return nil
		}
		pack := &chat.SCChatMsg{
			Msg2Snid:  proto.Int32(msg.Msg2Snid),
			Snid:      proto.Int32(p.SnId),
			Name:      proto.String(p.Name),
			Head:      proto.Int32(p.Head),
			Ts:        proto.Int64(time.Now().Unix()),
			OpRetCode: chat.OpResultCode_OPRC_Sucess,
		}
		if msg.GetMsg2Snid() == 0 { //只有大厅消息走关键字过滤
			if !ChatMgrSington.CanSendToPlatform(p.SnId) { //大厅聊天CD
				return nil
			}
			has := srvdata.HasSensitiveWord([]rune(msg.GetContent()))
			if has {
				content = string(srvdata.ReplaceSensitiveWord([]rune(msg.GetContent()))[:])
				logger.Logger.Trace("CSChatMsgHandler find HasSensitiveWord then after ReplaceSensitiveWord content=", content)
			}
			content = html.EscapeString(content)
			pack.Content = content
			proto.SetDefaults(pack)
			logger.Logger.Trace("SCChatMsg -> Platform", pack)
			PlayerMgrSington.BroadcastMessageToPlatformWithHall(p.Platform, p.SnId, int(chat.ChatPacketID_PACKET_SCChatMsg), pack)
		} else { //私聊
			FriendUnreadMgrSington.DelFriendUnread(p.SnId, msg.GetMsg2Snid())
			pack.Content = content
			if !FriendMgrSington.IsFriend(p.SnId, msg.GetMsg2Snid()) {
				pack.OpRetCode = chat.OpResultCode_OPRC_Chat_NotFriend
				logger.Logger.Warn("CSChatMsgHandler not friend")
			} else {
				if FriendMgrSington.IsShield(msg.GetMsg2Snid(), p.SnId) { //对方有没有屏蔽发送者
					pack.OpRetCode = chat.OpResultCode_OPRC_Chat_IsShield
					logger.Logger.Warn("CSChatMsgHandler IsShield")
				}
				if FriendMgrSington.IsShield(p.SnId, msg.GetMsg2Snid()) { //有没有屏蔽对方
					pack.OpRetCode = chat.OpResultCode_OPRC_Chat_Shield
					logger.Logger.Warn("CSChatMsgHandler Shield")
				}
			}
			if pack.OpRetCode == chat.OpResultCode_OPRC_Sucess {
				msg2p := PlayerMgrSington.GetPlayerBySnId(msg.GetMsg2Snid())
				if msg2p != nil { //在线
					proto.SetDefaults(pack)
					msg2p.SendToClient(int(chat.ChatPacketID_PACKET_SCChatMsg), pack)
					logger.Logger.Trace("SCChatMsg SnId", p.SnId, " -> Msg2Snid:", msg.GetMsg2Snid(), pack)
					FriendUnreadMgrSington.AddFriendUnread(msg.GetMsg2Snid(), p.SnId)
				} else { //不在线
					task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
						ret, err := model.QueryFriendUnreadBySnid(p.Platform, msg.GetMsg2Snid())
						if err != nil {
							return nil
						}
						return ret
					}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
						var ret *model.FriendUnread
						had := false
						if data != nil && data.(*model.FriendUnread) != nil {
							ret = data.(*model.FriendUnread)
							if ret.UnreadSnids != nil {
								for _, us := range ret.UnreadSnids {
									if us.SnId == p.SnId {
										us.UnreadNum++
										had = true
										break
									}
								}
							}
						} else {
							ret = model.NewFriendUnread(msg.GetMsg2Snid())
						}
						if !had {
							us := &model.FriendUnreadSnid{
								SnId:      p.SnId,
								UnreadNum: 1,
							}
							ret.UnreadSnids = append(ret.UnreadSnids, us)
						}
						task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
							return model.UpsertFriendUnread(p.Platform, msg.GetMsg2Snid(), ret)
						}), nil).StartByFixExecutor("UpsertFriendUnread")
					})).StartByFixExecutor("QueryFriendUnreadBySnid")
				}
				ChatMgrSington.AddChat(p.SnId, msg.GetMsg2Snid(), content)
				ChatMgrSington.SaveChatData(p.Platform, p.SnId, msg.GetMsg2Snid())
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(chat.ChatPacketID_PACKET_SCChatMsg), pack)
			logger.Logger.Trace("SCChatMsg ", pack)
		}

	}
	return nil
}

// 聊天记录
type CSGetChatLogPacketFactory struct {
}
type CSGetChatLogHandler struct {
}

func (this *CSGetChatLogPacketFactory) CreatePacket() interface{} {
	pack := &chat.CSGetChatLog{}
	return pack
}

func (this *CSGetChatLogHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGetChatLogHandler Process recv ", data)
	if msg, ok := data.(*chat.CSGetChatLog); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSGetChatLogHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			logger.Logger.Warn("CSGetChatLogHandler platform == nil")
			return nil
		}
		if !FriendMgrSington.IsFriend(p.SnId, msg.GetSnid()) {
			logger.Logger.Warn("CSGetChatLogHandler not friend")
			return nil
		}
		friendInfo := FriendMgrSington.GetBindFriendList(p.SnId)
		var bf *BindFriend
		if friendInfo != nil && len(friendInfo) != 0 {
			for _, friend := range friendInfo {
				if friend.SnId == msg.GetSnid() {
					bf = friend
				}
			}
		}
		FriendUnreadMgrSington.DelFriendUnread(p.SnId, msg.GetSnid())
		chatLogs := ChatMgrSington.GetChat(p.SnId, msg.GetSnid())
		if chatLogs != nil {
			pack := &chat.SCGetChatLog{
				Snid: proto.Int32(msg.GetSnid()),
			}
			if chatLogs.ChatContent != nil && len(chatLogs.ChatContent) != 0 {
				for _, content := range chatLogs.ChatContent {
					if content.Content != "" {
						if bf != nil {
							srcSnid := int32(0)
							srcName := ""
							srcHead := int32(0)
							toSnId := int32(0)
							toName := ""
							toHead := int32(0)
							if content.SrcSnId == p.SnId { //我说的
								srcSnid = p.SnId
								srcName = p.Name
								srcHead = p.Head
								toSnId = bf.SnId
								toName = bf.Name
								toHead = bf.Head
							} else { //对方说的
								srcSnid = bf.SnId
								srcName = bf.Name
								srcHead = bf.Head
								toSnId = p.SnId
								toName = p.Name
								toHead = p.Head
							}

							log := &chat.ChatLog{
								SrcSnId: proto.Int32(srcSnid),
								SrcName: proto.String(srcName),
								SrcHead: proto.Int32(srcHead),
								ToSnId:  proto.Int32(toSnId),
								ToName:  proto.String(toName),
								ToHead:  proto.Int32(toHead),
								Content: proto.String(content.Content),
								Ts:      proto.Int64(content.Ts),
							}
							pack.ChatLogs = append(pack.ChatLogs, log)
						}
					}
				}
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(chat.ChatPacketID_PACKET_SCGetChatLog), pack)
			logger.Logger.Trace("SCGetChatLog ", pack)
		} else {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				bindSnid := ChatMgrSington.getBindSnid(p.SnId, msg.GetSnid())
				ret, err := model.QueryChatByBindSnid(p.Platform, bindSnid)
				if err != nil {
					return nil
				}
				return ret
			}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
				pack := &chat.SCGetChatLog{
					Snid: proto.Int32(msg.GetSnid()),
				}
				if data != nil && data.(*model.Chat) != nil {
					ret := data.(*model.Chat)
					ChatMgrSington.SetChat(ret.BindSnid, ret)
					if ret.ChatContent != nil && len(ret.ChatContent) != 0 {
						for _, content := range ret.ChatContent {
							if content.Content != "" {
								if bf != nil {
									srcSnid := int32(0)
									srcName := ""
									srcHead := int32(0)
									toSnId := int32(0)
									toName := ""
									toHead := int32(0)
									if content.SrcSnId == p.SnId { //我说的
										srcSnid = p.SnId
										srcName = p.Name
										srcHead = p.Head
										toSnId = bf.SnId
										toName = bf.Name
										toHead = bf.Head
									} else { //对方说的
										srcSnid = bf.SnId
										srcName = bf.Name
										srcHead = bf.Head
										toSnId = p.SnId
										toName = p.Name
										toHead = p.Head
									}

									log := &chat.ChatLog{
										SrcSnId: proto.Int32(srcSnid),
										SrcName: proto.String(srcName),
										SrcHead: proto.Int32(srcHead),
										ToSnId:  proto.Int32(toSnId),
										ToName:  proto.String(toName),
										ToHead:  proto.Int32(toHead),
										Content: proto.String(content.Content),
										Ts:      proto.Int64(content.Ts),
									}
									pack.ChatLogs = append(pack.ChatLogs, log)
								}
							}
						}
					}
				}
				proto.SetDefaults(pack)
				p.SendToClient(int(chat.ChatPacketID_PACKET_SCGetChatLog), pack)
				logger.Logger.Trace("SCGetChatLog ", pack)
			})).StartByFixExecutor("LoadChatData")
		}
	}
	return nil
}

// 读消息
type CSReadChatMsgPacketFactory struct {
}
type CSReadChatMsgHandler struct {
}

func (this *CSReadChatMsgPacketFactory) CreatePacket() interface{} {
	pack := &chat.CSReadChatMsg{}
	return pack
}

func (this *CSReadChatMsgHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSReadChatMsgHandler Process recv ", data)
	if msg, ok := data.(*chat.CSReadChatMsg); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSReadChatMsgHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			logger.Logger.Warn("CSReadChatMsgHandler platform == nil")
			return nil
		}
		if !FriendMgrSington.IsFriend(p.SnId, msg.GetSnid()) {
			logger.Logger.Warn("CSReadChatMsgHandler not friend")
			return nil
		}
		FriendUnreadMgrSington.DelFriendUnread(p.SnId, msg.GetSnid())
	}
	return nil
}

// 屏蔽
type CSShieldMsgPacketFactory struct {
}
type CSShieldMsgHandler struct {
}

func (this *CSShieldMsgPacketFactory) CreatePacket() interface{} {
	pack := &chat.CSShieldMsg{}
	return pack
}

func (this *CSShieldMsgHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSShieldMsgHandler Process recv ", data)
	if msg, ok := data.(*chat.CSShieldMsg); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSShieldMsgHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			logger.Logger.Warn("CSShieldMsgHandler platform == nil")
			return nil
		}

		pack := &chat.SCShieldMsg{
			Snid:       proto.Int32(p.SnId),
			ShieldSnid: proto.Int32(msg.GetShieldSnid()),
			Shield:     proto.Bool(msg.GetShield()),
			Ts:         proto.Int64(proto.Int64(time.Now().Unix())),
			OpRetCode:  chat.OpResultCode_OPRC_Sucess,
		}
		if msg.GetShield() {
			if FriendMgrSington.IsShield(p.SnId, msg.GetShieldSnid()) {
				logger.Logger.Warn("重复屏蔽")
				pack.OpRetCode = chat.OpResultCode_OPRC_Chat_ReShield
			}
		} else {
			if !FriendMgrSington.IsShield(p.SnId, msg.GetShieldSnid()) {
				logger.Logger.Warn("重复解除")
				pack.OpRetCode = chat.OpResultCode_OPRC_Chat_ReUnShield
			}
		}
		if pack.OpRetCode == chat.OpResultCode_OPRC_Sucess {
			if msg.GetShield() {
				FriendMgrSington.AddShield(p.SnId, msg.GetShieldSnid())
			} else {
				FriendMgrSington.DelShield(p.SnId, msg.GetShieldSnid())
			}
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(chat.ChatPacketID_PACKET_SCShieldMsg), pack)
		logger.Logger.Trace("SCShieldMsg ", pack)
	}
	return nil
}

func init() {
	//聊天消息
	common.RegisterHandler(int(chat.ChatPacketID_PACKET_CSChatMsg), &CSChatMsgHandler{})
	netlib.RegisterFactory(int(chat.ChatPacketID_PACKET_CSChatMsg), &CSChatMsgPacketFactory{})
	//聊天记录
	common.RegisterHandler(int(chat.ChatPacketID_PACKET_CSGetChatLog), &CSGetChatLogHandler{})
	netlib.RegisterFactory(int(chat.ChatPacketID_PACKET_CSGetChatLog), &CSGetChatLogPacketFactory{})
	//读消息
	common.RegisterHandler(int(chat.ChatPacketID_PACKET_CSReadChatMsg), &CSReadChatMsgHandler{})
	netlib.RegisterFactory(int(chat.ChatPacketID_PACKET_CSReadChatMsg), &CSReadChatMsgPacketFactory{})
	//屏蔽玩家
	common.RegisterHandler(int(chat.ChatPacketID_PACKET_CSShieldMsg), &CSShieldMsgHandler{})
	netlib.RegisterFactory(int(chat.ChatPacketID_PACKET_CSShieldMsg), &CSShieldMsgPacketFactory{})
}
