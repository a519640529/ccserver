package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/friend"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type CSFriendListPacketFactory struct {
}

type CSFriendListHandler struct {
}

func (this *CSFriendListPacketFactory) CreatePacket() interface{} {
	pack := &friend.CSFriendList{}
	return pack
}

func (this *CSFriendListHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSFriendListHandler Process recv ", data)
	if msg, ok := data.(*friend.CSFriendList); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSFriendListHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		pack := &friend.SCFriendList{
			ListType:  proto.Int32(msg.GetListType()),
			OpRetCode: friend.OpResultCode_OPRC_Sucess,
		}
		switch msg.GetListType() {
		case ListType_Friend:
			dfl := FriendMgrSington.GetBindFriendList(p.GetSnId())
			if dfl != nil {
				for _, bf := range dfl {
					if bf.SnId == p.SnId {
						continue
					}
					fi := &friend.FriendInfo{
						SnId:       proto.Int32(bf.SnId),
						Name:       proto.String(bf.Name),
						Sex:        proto.Int32(bf.Sex),
						Head:       proto.Int32(bf.Head),
						CreateTs:   proto.Int64(bf.CreateTime),
						LogoutTs:   proto.Int64(bf.LogoutTime),
						LastChatTs: proto.Int64(bf.CreateTime),
						IsShield:   proto.Bool(FriendMgrSington.IsShield(p.SnId, bf.SnId)),
					}
					fi.Online = false
					bfp := PlayerMgrSington.GetPlayerBySnId(bf.SnId)
					if bfp != nil {
						fi.Online = bfp.IsOnLine()
					}
					chat := ChatMgrSington.GetChat(p.SnId, bf.SnId)
					if chat != nil {
						fi.LastChatTs = chat.LastChatTs
					}
					pack.FriendArr = append(pack.FriendArr, fi)
				}
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(friend.FriendPacketID_PACKET_SCFriendList), pack)
			logger.Logger.Trace("SCFriendListHandler: 好友列表 pack: ", pack)
		case ListType_Apply:
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				ret, err := model.QueryFriendApplyBySnid(p.Platform, p.SnId)
				if err != nil {
					return nil
				}
				return ret
			}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
				ret := data.(*model.FriendApply)
				if ret != nil && ret.ApplySnids != nil {
					for _, as := range ret.ApplySnids {
						if as.SnId == p.SnId {
							continue
						}
						fi := &friend.FriendInfo{
							SnId:     proto.Int32(as.SnId),
							Name:     proto.String(as.Name),
							Head:     proto.Int32(as.Head),
							CreateTs: proto.Int64(as.CreateTs),
						}
						fi.Online = false
						bfp := PlayerMgrSington.GetPlayerBySnId(as.SnId)
						if bfp != nil {
							fi.Online = bfp.IsOnLine()
						}
						pack.FriendArr = append(pack.FriendArr, fi)
					}
				}
				proto.SetDefaults(pack)
				p.SendToClient(int(friend.FriendPacketID_PACKET_SCFriendList), pack)
				logger.Logger.Trace("SCFriendListHandler: 申请列表 pack: ", pack)
			})).StartByFixExecutor("QueryFriendApplyBySnid")
		case ListType_Recommend:
			//canGet, timeOut := FriendMgrSington.CanGetRecommendFriendList(p.SnId)
			//if !canGet { //获取推荐好友列表CD
			//	proto.SetDefaults(pack)
			//	pack.Param = proto.Int32(int32(timeOut))
			//	p.SendToClient(int(friend.FriendPacketID_PACKET_SCFriendList), pack)
			//	logger.Logger.Trace("SCFriendListHandler: 推荐列表cd中 pack: ", pack)
			//	return nil
			//}
			friends := PlayerMgrSington.RecommendFriendRule(p.Platform, p.SnId)
			for _, f := range friends {
				if f.Snid == p.SnId {
					continue
				}
				fi := &friend.FriendInfo{
					SnId: proto.Int32(f.Snid),
					Name: proto.String(f.Name),
					Head: proto.Int32(f.Head),
				}
				pack.FriendArr = append(pack.FriendArr, fi)
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(friend.FriendPacketID_PACKET_SCFriendList), pack)
			logger.Logger.Trace("SCFriendListHandler: 推荐列表 pack: ", pack)
		}
	}
	return nil
}

type CSFriendOpPacketFactory struct {
}

type CSFriendOpHandler struct {
}

func (this *CSFriendOpPacketFactory) CreatePacket() interface{} {
	pack := &friend.CSFriendOp{}
	return pack
}

func (this *CSFriendOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSFriendOpHandler Process recv ", data)
	if msg, ok := data.(*friend.CSFriendOp); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSFriendOpHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		destSnid := msg.GetSnId()
		if destSnid == 99999999 {
			return nil
		}
		opCode := msg.GetOpCode()
		destP := PlayerMgrSington.GetPlayerBySnId(destSnid)

		var retCode friend.OpResultCode
		send := func(p *Player, bf *BindFriend) bool {
			pack := &friend.SCFriendOp{
				OpCode:    proto.Int32(opCode),
				SnId:      proto.Int32(destSnid),
				OpRetCode: retCode,
			}
			if retCode == friend.OpResultCode_OPRC_Sucess && bf != nil {
				pack.Friend = &friend.FriendInfo{
					SnId:     proto.Int32(bf.SnId),
					Name:     proto.String(bf.Name),
					Sex:      proto.Int32(bf.Sex),
					Head:     proto.Int32(bf.Head),
					CreateTs: proto.Int64(bf.CreateTime),
					LogoutTs: proto.Int64(bf.LogoutTime),
				}
				pack.Friend.Online = false
				bfp := PlayerMgrSington.GetPlayerBySnId(bf.SnId)
				if bfp != nil {
					pack.Friend.Online = bfp.IsOnLine()
				}
			}
			proto.SetDefaults(pack)
			sendTo := p.SendToClient(int(friend.FriendPacketID_PACKET_SCFriendOp), pack)
			logger.Logger.Trace("SCFriendOpHandler->Snid: ", p.SnId, " isok: ", sendTo, " pack: ", pack)
			return sendTo
		}

		if destP != nil && destP.IsRob {
			return nil
		}

		if destSnid == p.SnId {
			retCode = friend.OpResultCode_OPRC_Friend_NotOpMyself
			send(p, nil)
			return nil
		}

		bf := &BindFriend{ //发起者
			Platform: p.Platform,
			SnId:     p.SnId,
			Name:     p.Name,
			Head:     p.Head,
			Sex:      p.Sex,
		}
		switch opCode {
		case OpType_Apply:
			logger.Logger.Info("申请加好友", p.SnId, " -> ", destSnid)
		case OpType_Agree:
			logger.Logger.Info("同意加好友", p.SnId, " -> ", destSnid)
		case OpType_Refuse:
			logger.Logger.Info("拒绝加好友", p.SnId, " -> ", destSnid)
		case OpType_Delete:
			logger.Logger.Info("删除好友", p.SnId, " -> ", destSnid)
		}
		switch opCode {
		case OpType_Apply:
			if FriendMgrSington.IsFriend(p.SnId, destSnid) {
				retCode = friend.OpResultCode_OPRC_Friend_AlreadyAdd
				send(p, nil)
				return nil
			}
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				ret, err := model.QueryFriendApplyBySnid(p.Platform, destSnid)
				if err != nil {
					return nil
				}
				return ret
			}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
				ret := data.(*model.FriendApply)
				if ret != nil {
					if ret.ApplySnids != nil {
						if len(ret.ApplySnids) > FriendApplyMaxNum {
							retCode = friend.OpResultCode_OPRC_Friend_DestApplyFriendMax //对方好友申请已达上限
						}
						for _, as := range ret.ApplySnids {
							if as.SnId == p.SnId {
								retCode = friend.OpResultCode_OPRC_Friend_AlreadyApply //已经申请过好友
								break
							}
						}
					}
				} else {
					ret = model.NewFriendApply(destSnid)
				}
				if retCode != friend.OpResultCode_OPRC_Sucess {
					send(p, nil)
					return
				} else {
					as := &model.FriendApplySnid{
						SnId:     p.SnId,
						Name:     p.Name,
						Head:     p.Head,
						CreateTs: time.Now().Unix(),
					}
					ret.ApplySnids = append(ret.ApplySnids, as)
					task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
						return model.UpsertFriendApply(p.Platform, destSnid, ret)
					}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
						if destP != nil {
							send(destP, bf)
						}
					})).StartByFixExecutor("UpsertFriendApply")
				}
			})).StartByFixExecutor("QueryFriendApplyBySnid")
		case OpType_Agree:
			me := FriendMgrSington.GetFriendBySnid(p.SnId)
			if me == nil {
				retCode = friend.OpResultCode_OPRC_Error
				send(p, nil)
				return nil
			}
			if FriendMgrSington.IsFriend(p.SnId, destSnid) { //已经是好友了
				retCode = friend.OpResultCode_OPRC_Friend_AlreadyAdd
				send(p, nil)
				// 申请信息更新
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					ret, err := model.QueryFriendApplyBySnid(p.Platform, p.SnId)
					if err != nil {
						return nil
					}
					return ret
				}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
					ret := data.(*model.FriendApply)
					if ret != nil {
						if ret.ApplySnids != nil {
							for i, as := range ret.ApplySnids {
								if as.SnId == destSnid {
									ret.ApplySnids = append(ret.ApplySnids[:i], ret.ApplySnids[i+1:]...)
									task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
										return model.UpsertFriendApply(p.Platform, p.SnId, ret)
									}), nil).StartByFixExecutor("UpsertFriendApply")
									break
								}
							}
						}
					}
				})).StartByFixExecutor("QueryFriendApplyBySnid")
				return nil
			}
			if me.BindFriend != nil {
				if len(me.BindFriend) > FriendMaxNum {
					retCode = friend.OpResultCode_OPRC_Friend_FriendMax
					send(p, nil)
					return nil
				}
			}
			f := FriendMgrSington.GetFriendBySnid(destSnid)
			if f != nil {
				if FriendMgrSington.IsFriend(destSnid, p.SnId) { //已经是好友了
					retCode = friend.OpResultCode_OPRC_Friend_AlreadyAdd
					send(p, nil)
					return nil
				}
				//同意者加入到被同意者好友里
				err := FriendMgrSington.AddBindFriend(f.SnId, bf)
				if err != friend.OpResultCode_OPRC_Sucess {
					logger.Logger.Warn("AddBindFriend error: ", err)
					retCode = err
					send(p, nil)
				} else {
					if destP != nil {
						send(destP, bf)
					}
					// 被同意者加入到同意者好友里
					destBf := &BindFriend{
						Platform: f.Platform,
						SnId:     f.SnId,
						Name:     f.Name,
						Head:     f.Head,
						Sex:      f.Sex,
					}
					err := FriendMgrSington.AddBindFriend(p.SnId, destBf)
					if err == friend.OpResultCode_OPRC_Sucess {
						send(p, destBf)
					}
					// 申请信息更新
					task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
						ret, err := model.QueryFriendApplyBySnid(p.Platform, p.SnId)
						if err != nil {
							return nil
						}
						return ret
					}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
						ret := data.(*model.FriendApply)
						if ret != nil {
							if ret.ApplySnids != nil {
								for i, as := range ret.ApplySnids {
									if as.SnId == destSnid {
										ret.ApplySnids = append(ret.ApplySnids[:i], ret.ApplySnids[i+1:]...)
										task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
											return model.UpsertFriendApply(p.Platform, p.SnId, ret)
										}), nil).StartByFixExecutor("UpsertFriendApply")
										break
									}
								}
							}
						}
					})).StartByFixExecutor("QueryFriendApplyBySnid")
				}
			} else {
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					ret, err := model.QueryFriendBySnid(p.Platform, destSnid)
					if err != nil {
						return nil
					}
					return ret
				}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
					ret := data.(*model.Friend)
					if ret != nil {
						// 同意者加入到被同意者好友里
						if len(ret.BindFriend) > FriendMaxNum {
							retCode = friend.OpResultCode_OPRC_Friend_DestFriendMax
							send(p, nil)
							return
						}
						if ret.BindFriend != nil {
							for _, bindFriend := range ret.BindFriend {
								if bindFriend.SnId == p.SnId {
									retCode = friend.OpResultCode_OPRC_Friend_AlreadyAdd
									send(p, nil)
									return
								}
							}
						}
						as := &model.BindFriend{
							SnId:       p.SnId,
							CreateTime: time.Now().Unix(),
						}
						ret.BindFriend = append(ret.BindFriend, as)
						task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
							return model.UpsertFriend(ret)
						}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
							if destP != nil {
								send(destP, bf)
							}
						})).StartByFixExecutor("UpsertFriend")

						// 被同意者加入到同意者好友里
						destBf := &BindFriend{
							Platform:   ret.Platform,
							SnId:       ret.SnId,
							Name:       ret.Name,
							Head:       ret.Head,
							Sex:        ret.Sex,
							CreateTime: time.Now().Unix(),
							LogoutTime: ret.LogoutTime,
						}
						err := FriendMgrSington.AddBindFriend(p.SnId, destBf)
						if err == friend.OpResultCode_OPRC_Sucess {
							send(p, destBf)
						}
						// 申请信息更新
						task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
							ret, err := model.QueryFriendApplyBySnid(p.Platform, p.SnId)
							if err != nil {
								return nil
							}
							return ret
						}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
							ret := data.(*model.FriendApply)
							if ret != nil {
								if ret.ApplySnids != nil {
									for i, as := range ret.ApplySnids {
										if as.SnId == destSnid {
											ret.ApplySnids = append(ret.ApplySnids[:i], ret.ApplySnids[i+1:]...)
											task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
												return model.UpsertFriendApply(p.Platform, p.SnId, ret)
											}), nil).StartByFixExecutor("UpsertFriendApply")
											break
										}
									}
								}
							}
						})).StartByFixExecutor("QueryFriendApplyBySnid")

					} else {
						retCode = friend.OpResultCode_OPRC_Error
						send(p, nil)
						return
					}
				})).StartByFixExecutor("QueryFriendBySnid")
			}
		case OpType_Refuse:
			if FriendMgrSington.IsFriend(p.SnId, destSnid) {
				retCode = friend.OpResultCode_OPRC_Friend_AlreadyAdd
				send(p, nil)
			}
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				ret, err := model.QueryFriendApplyBySnid(p.Platform, p.SnId)
				if err != nil {
					return nil
				}
				return ret
			}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
				ret := data.(*model.FriendApply)
				if ret != nil {
					if ret.ApplySnids != nil {
						for i, as := range ret.ApplySnids {
							if as.SnId == destSnid {
								ret.ApplySnids = append(ret.ApplySnids[:i], ret.ApplySnids[i+1:]...)
								task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
									return model.UpsertFriendApply(p.Platform, p.SnId, ret)
								}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
									send(p, nil)
								})).StartByFixExecutor("UpsertFriendApply")
								break
							}
						}
					}
				}
			})).StartByFixExecutor("QueryFriendApplyBySnid")

		case OpType_Delete:
			f := FriendMgrSington.GetFriendBySnid(destSnid)
			if f != nil {
				//发起者删除被删除者
				isok1 := FriendMgrSington.DelBindFriend(p.SnId, f.SnId)
				if !isok1 {
					logger.Logger.Warn("DelBindFriend error: ", p.SnId, " del friend:", f.SnId)
					retCode = friend.OpResultCode_OPRC_Error
				}
				send(p, nil)
				//被删除者删除发起者
				isok2 := FriendMgrSington.DelBindFriend(f.SnId, p.SnId)
				if !isok2 {
					logger.Logger.Warn("DelBindFriend error: ", f.SnId, " del friend:", p.SnId)
					//删除失败不用通知
				} else {
					if destP != nil {
						send(destP, bf)
					}
				}
				ChatMgrSington.DelChat(p.Platform, p.SnId, f.SnId)
			} else {
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					ret, err := model.QueryFriendBySnid(p.Platform, destSnid)
					if err != nil {
						return nil
					}
					return ret
				}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
					ret := data.(*model.Friend)
					if ret != nil {
						//发起者删除被删除者
						isok1 := FriendMgrSington.DelBindFriend(p.SnId, ret.SnId)
						if !isok1 {
							logger.Logger.Warn("DelBindFriend error: ", p.SnId, " del friend:", ret.SnId)
							retCode = friend.OpResultCode_OPRC_Error
						}
						send(p, nil)
						//被删除者删除发起者
						if ret.BindFriend != nil {
							for i, bindFriend := range ret.BindFriend {
								if bindFriend.SnId == p.SnId {
									ret.BindFriend = append(ret.BindFriend[:i], ret.BindFriend[i+1:]...)
									task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
										return model.UpsertFriend(ret)
									}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
										if destP != nil {
											send(destP, bf)
										}
									})).StartByFixExecutor("UpsertFriendApply")
									break
								}
							}
						}
					}
					ChatMgrSington.DelChat(p.Platform, p.SnId, ret.SnId)
				})).StartByFixExecutor("DelBindFriend")
			}
		}
	}
	return nil
}

type CSQueryPlayerGameLogPacketFactory struct {
}
type CSQueryPlayerGameLogHandler struct {
}

func (this *CSQueryPlayerGameLogPacketFactory) CreatePacket() interface{} {
	pack := &friend.CSQueryPlayerGameLog{}
	return pack
}

func (this *CSQueryPlayerGameLogHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSQueryPlayerGameLogHandler Process recv ", data)
	if msg, ok := data.(*friend.CSQueryPlayerGameLog); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSQueryPlayerGameLogHandler p == nil")
			return nil
		}
		snid := msg.GetSnid()
		gameId := msg.GetGameId()
		size := msg.GetSize()

		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			ret := model.GetFriendRecordLogBySnid(p.Platform, snid, gameId, int(size))
			return ret
		}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			pack := &friend.SCQueryPlayerGameLog{
				Snid:   snid,
				GameId: gameId,
				Size:   size,
			}
			if data != nil {
				ret := data.([]*model.FriendRecord)
				if ret != nil {
					for _, gpl := range ret {
						gl := &friend.PlayerGameLog{
							GameId:    proto.Int32(gpl.GameId),
							BaseScore: proto.Int32(gpl.BaseScore),
							IsWin:     proto.Int32(gpl.IsWin),
							Ts:        proto.Int64(gpl.Ts),
						}
						pack.GameLogs = append(pack.GameLogs, gl)
					}

				}
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(friend.FriendPacketID_PACKET_SCQueryPlayerGameLog), pack)
			logger.Logger.Trace("SCQueryPlayerGameLogHandler: ", pack)
		})).StartByFixExecutor("GetPlayerListLogBySnid")
	}
	return nil
}

type CSInviteFriendPacketFactory struct {
}

type CSInviteFriendHandler struct {
}

func (this *CSInviteFriendPacketFactory) CreatePacket() interface{} {
	pack := &friend.CSInviteFriend{}
	return pack
}

func (this *CSInviteFriendHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSInviteFriendHandler Process recv ", data)
	if msg, ok := data.(*friend.CSInviteFriend); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSInviteFriendHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			logger.Logger.Warn("CSInviteFriendHandler platform == nil")
			return nil
		}
		friendSnid := msg.GetToSnId()
		var opRetCode = friend.OpResultCode_OPRC_Sucess
		send := func(player *Player) {
			pack := &friend.SCInviteFriend{
				SrcSnId:   proto.Int32(p.SnId),
				SrcName:   proto.String(p.Name),
				SrcHead:   proto.Int32(p.Head),
				OpRetCode: opRetCode,
				GameId:    proto.Int(p.scene.gameId),
				RoomId:    proto.Int(p.scene.sceneId),
				Pos:       proto.Int32(msg.GetPos()),
			}
			proto.SetDefaults(pack)
			player.SendToClient(int(friend.FriendPacketID_PACKET_SCInviteFriend), pack)
			logger.Logger.Trace("SCInviteFriendHandler: ", pack)
		}
		//不能邀请自己
		if p.SnId == friendSnid {
			logger.Logger.Warn("CSInviteFriendHandler invite self")
			opRetCode = friend.OpResultCode_OPRC_Friend_NotOpMyself
			send(p)
			return nil
		}
		//不是好友
		if !FriendMgrSington.IsFriend(p.SnId, friendSnid) {
			logger.Logger.Warn("CSInviteFriendHandler not friend")
			opRetCode = friend.OpResultCode_OPRC_InviteFriend_NotFriend
			send(p)
			return nil
		}
		fp := PlayerMgrSington.GetPlayerBySnId(friendSnid)
		//不在线
		if fp == nil || !fp.IsOnLine() {
			logger.Logger.Warn("CSInviteFriendHandler not online")
			opRetCode = friend.OpResultCode_OPRC_InviteFriend_NoOnline
			send(p)
			return nil
		}
		//CD
		if !FriendMgrSington.CanInvite(p.SnId, friendSnid) {
			logger.Logger.Warn("CSInviteFriendHandler in cd time")
			opRetCode = friend.OpResultCode_OPRC_Error
			send(p)
			return nil
		}
		//scene
		if p.scene == nil {
			logger.Logger.Warn("CSInviteFriendHandler scene is nil")
			opRetCode = friend.OpResultCode_OPRC_InviteFriend_SceneNotExist
			send(p)
			return nil
		}
		//私有房间
		if p.scene.sceneMode != common.SceneMode_Private {
			logger.Logger.Warn("CSInviteFriendHandler scene is common.SceneMode_Private")
			opRetCode = friend.OpResultCode_OPRC_InviteFriend_RoomLimit
			send(p)
			return nil
		}
		//好友已在游戏中
		if fp.scene != nil {
			logger.Logger.Warn("CSInviteFriendHandler scene is common.SceneMode_Private")
			opRetCode = friend.OpResultCode_OPRC_InviteFriend_Gaming
			send(p)
			return nil
		}
		pos := int(msg.GetPos())
		if pos < 0 || pos >= p.scene.playerNum {
			logger.Logger.Trace("CSInviteFriendHandler pos is fail")
			opRetCode = friend.OpResultCode_OPRC_InviteFriend_PosIsError //座位不存在
			send(p)
			return nil
		}
		if opRetCode == friend.OpResultCode_OPRC_Sucess { //成功都通知下
			send(p)
			send(fp)
		}
	}
	return nil
}

type CSInviteFriendOpPacketFactory struct {
}

type CSInviteFriendOpHandler struct {
}

func (this *CSInviteFriendOpPacketFactory) CreatePacket() interface{} {
	pack := &friend.CSInviteFriendOp{}
	return pack
}

func (this *CSInviteFriendOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSInviteFriendOpHandler Process recv ", data)
	if msg, ok := data.(*friend.CSInviteFriendOp); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSInviteFriendOpHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			logger.Logger.Warn("CSInviteFriendOpHandler platform == nil")
			return nil
		}
		srcSnid := msg.GetSnId()
		opcode := msg.GetOpCode()
		var opRetCode = friend.OpResultCode_OPRC_Sucess
		send := func(player *Player) {
			pack := &friend.SCInviteFriendOp{
				SnId:      proto.Int32(p.SnId),
				Name:      proto.String(p.Name),
				OpCode:    proto.Int32(opcode),
				OpRetCode: opRetCode,
				Pos:       proto.Int32(msg.GetPos()),
			}
			proto.SetDefaults(pack)
			isok := player.SendToClient(int(friend.FriendPacketID_PACKET_SCInviteFriendOp), pack)
			logger.Logger.Trace("SCInviteFriendOpHandler isok: ", isok, " pack: ", pack)
		}
		fp := PlayerMgrSington.GetPlayerBySnId(srcSnid)
		//不能操作自己
		if p.SnId == srcSnid {
			logger.Logger.Warn("CSInviteFriendHandler invite self")
			opRetCode = friend.OpResultCode_OPRC_Friend_NotOpMyself
			send(p)
			return nil
		}
		//不在线
		if fp == nil || !fp.IsOnLine() {
			logger.Logger.Warn("CSInviteFriendHandler not online")
			opRetCode = friend.OpResultCode_OPRC_InviteFriend_NoOnline
			send(p)
			return nil
		}
		//不是好友
		if !FriendMgrSington.IsFriend(p.SnId, srcSnid) {
			logger.Logger.Warn("CSInviteFriendHandler not friend")
			opRetCode = friend.OpResultCode_OPRC_InviteFriend_NotFriend
			send(p)
			return nil
		}

		switch int(opcode) {
		case Invite_Agree:
			logger.Logger.Trace("同意邀请")
			if p.scene != nil {
				logger.Logger.Warn("CSInviteFriendHandler scene is not nil")
				opRetCode = friend.OpResultCode_OPRC_InviteFriend_HadInRoom //已在房间中
				send(p)
				return nil
			}
			scene := fp.scene
			if scene == nil {
				logger.Logger.Warn("CSInviteFriendHandler scene is nil")
				opRetCode = friend.OpResultCode_OPRC_InviteFriend_SceneNotExist //场景不存在
				send(p)
				return nil
			}
			//私有房间
			if scene.sceneMode != common.SceneMode_Private {
				logger.Logger.Warn("CSInviteFriendHandler scene is common.SceneMode_Private")
				opRetCode = friend.OpResultCode_OPRC_InviteFriend_RoomLimit //只能进入私有房间
				send(p)
				return nil
			}
			//进入房间
			if scene.limitPlatform != nil {
				if scene.limitPlatform.Isolated && p.Platform != scene.limitPlatform.IdStr {
					logger.Logger.Warn("CSInviteFriendHandler scene room not find")
					opRetCode = friend.OpResultCode_OPRC_InviteFriend_RoomNotExist //房间不存在
					send(p)
					return nil
				}
			}
			if scene.deleting {
				logger.Logger.Warn("CSInviteFriendHandler scene is deleting")
				opRetCode = friend.OpResultCode_OPRC_InviteFriend_SceneDeleting //场景正在删除
				send(p)
				return nil
			}

			if scene.closed {
				logger.Logger.Warn("CSInviteFriendHandler scene is closed")
				opRetCode = friend.OpResultCode_OPRC_InviteFriend_SceneClosed //场景已关闭
				send(p)
				return nil
			}

			dbGameFree := scene.dbGameFree
			if dbGameFree != nil {
				limitCoin := srvdata.CreateRoomMgrSington.GetLimitCoinByBaseScore(int32(scene.gameId), int32(scene.gameSite), scene.BaseScore)
				if p.Coin < limitCoin {
					logger.Logger.Warn("CSInviteFriendHandler player limitCoin")
					opRetCode = friend.OpResultCode_OPRC_InviteFriend_CoinLimit //金币不足
					send(p)
					return nil
				}
			}

			sp := GetScenePolicy(scene.gameId, scene.gameMode)
			if sp == nil {
				logger.Logger.Warn("CSInviteFriendHandler game not exist")
				opRetCode = friend.OpResultCode_OPRC_InviteFriend_GameNotExist //游戏不存在
				send(p)
				return nil
			}
			if reason := sp.CanEnter(scene, p); reason != 0 {
				logger.Logger.Trace("CSInviteFriendHandler CanEnter reason ", reason)
				opRetCode = friend.OpResultCode_OPRC_InviteFriend_GameNotCanEnter //游戏已开始
				send(p)
				return nil
			}
			if scene.IsFull() {
				logger.Logger.Trace("CSInviteFriendHandler sp is full")
				opRetCode = friend.OpResultCode_OPRC_InviteFriend_RoomFull //房间已满员
				send(p)
				return nil
			}
			pos := int(msg.GetPos())
			if pos < 0 || pos >= scene.playerNum {
				logger.Logger.Trace("CSInviteFriendHandler pos is fail")
				opRetCode = friend.OpResultCode_OPRC_InviteFriend_PosIsError //座位不存在
				send(p)
				return nil
			}
			if !p.EnterScene(scene, true, pos) {
				logger.Logger.Trace("CSInviteFriendHandler EnterScene fail")
				opRetCode = friend.OpResultCode_OPRC_Error //进入房间失败
				send(p)
				return nil
			} else {
				//成功进入房间的消息在gameserver上会发送
				CoinSceneMgrSington.OnPlayerEnter(p, dbGameFree.Id)
				return nil
			}
		case Invite_Refuse:
			logger.Logger.Trace("拒绝邀请")
			send(fp) //通知邀请者
		}
	}
	return nil
}

type CSFuzzyQueryPlayerPacketFactory struct {
}

type CSFuzzyQueryPlayerHandler struct {
}

func (this *CSFuzzyQueryPlayerPacketFactory) CreatePacket() interface{} {
	pack := &friend.CSFuzzyQueryPlayer{}
	return pack
}

func (this *CSFuzzyQueryPlayerHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSFuzzyQueryPlayerHandler Process recv ", data)
	if msg, ok := data.(*friend.CSFuzzyQueryPlayer); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSFuzzyQueryPlayerHandler p == nil")
			return nil
		}
		queryContent := msg.GetQueryContent()
		if utf8.RuneCountInString(queryContent) < 3 {
			return nil
		}
		pack := &friend.SCFuzzyQueryPlayer{
			QueryContent: proto.String(queryContent),
		}
		players := PlayerMgrSington.playerSnMap
		if players != nil {
			for _, player := range players {
				if player != nil && player.IsOnLine() /*&& !player.IsRobot() && !FriendMgrSington.IsFriend(p.SnId, player.SnId)*/ { //在线
					snidStr := strconv.FormatInt(int64(player.SnId), 10)
					if strings.Contains(player.Name, queryContent) || strings.Contains(snidStr, queryContent) {
						pi := &friend.PlayerInfo{
							SnId: proto.Int32(player.SnId),
							Name: proto.String(player.Name),
							Sex:  proto.Int32(player.Sex),
							Head: proto.Int32(player.Head),
						}
						pack.Players = append(pack.Players, pi)
					}
				}
			}
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(friend.FriendPacketID_PACKET_SCFuzzyQueryPlayer), pack)
	}
	return nil
}

func init() {
	//好友列表 申请列表 推荐列表
	common.RegisterHandler(int(friend.FriendPacketID_PACKET_CSFriendList), &CSFriendListHandler{})
	netlib.RegisterFactory(int(friend.FriendPacketID_PACKET_CSFriendList), &CSFriendListPacketFactory{})
	//申请 同意 拒绝 删除
	common.RegisterHandler(int(friend.FriendPacketID_PACKET_CSFriendOp), &CSFriendOpHandler{})
	netlib.RegisterFactory(int(friend.FriendPacketID_PACKET_CSFriendOp), &CSFriendOpPacketFactory{})
	//查看别人战绩
	common.RegisterHandler(int(friend.FriendPacketID_PACKET_CSQueryPlayerGameLog), &CSQueryPlayerGameLogHandler{})
	netlib.RegisterFactory(int(friend.FriendPacketID_PACKET_CSQueryPlayerGameLog), &CSQueryPlayerGameLogPacketFactory{})
	//邀请好友对战
	common.RegisterHandler(int(friend.FriendPacketID_PACKET_CSInviteFriend), &CSInviteFriendHandler{})
	netlib.RegisterFactory(int(friend.FriendPacketID_PACKET_CSInviteFriend), &CSInviteFriendPacketFactory{})
	//同意、拒绝好友邀请
	common.RegisterHandler(int(friend.FriendPacketID_PACKET_CSInviteFriendOp), &CSInviteFriendOpHandler{})
	netlib.RegisterFactory(int(friend.FriendPacketID_PACKET_CSInviteFriendOp), &CSInviteFriendOpPacketFactory{})
	//根据id或者昵称查询玩家:
	common.RegisterHandler(int(friend.FriendPacketID_PACKET_CSFuzzyQueryPlayer), &CSFuzzyQueryPlayerHandler{})
	netlib.RegisterFactory(int(friend.FriendPacketID_PACKET_CSFuzzyQueryPlayer), &CSFuzzyQueryPlayerPacketFactory{})
}
