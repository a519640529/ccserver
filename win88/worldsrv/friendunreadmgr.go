package main

import (
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/friend"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
	"time"
)

var FriendUnreadMgrSington = &FriendUnreadMgr{
	FriendUnreadList:  make(map[int32]map[int32]int),
	FriendUnreadDirty: make(map[int32]bool),
}

type FriendUnreadMgr struct {
	FriendUnreadList  map[int32]map[int32]int
	FriendUnreadDirty map[int32]bool
}

func (this *FriendUnreadMgr) AddFriendUnread(snid, chatsnid int32) {
	ful := this.FriendUnreadList[snid]
	if ful == nil {
		ful = make(map[int32]int)
		ful[chatsnid] = 1
		this.FriendUnreadList[snid] = ful
	} else {
		had := false
		for cSnid, unreadNum := range ful {
			if cSnid == chatsnid {
				unreadNum++
				had = true
				break
			}
		}
		if !had {
			ful[chatsnid] = 1
		}
	}
	this.FriendUnreadDirty[snid] = true
}

func (this *FriendUnreadMgr) DelFriendUnread(snid, chatsnid int32) {
	ful := this.FriendUnreadList[snid]
	if ful == nil {
		return
	}
	delete(ful, chatsnid)
	this.FriendUnreadDirty[snid] = true
}

func (this *FriendUnreadMgr) LoadFriendUnreadData(platform string, snid int32) {
	if this.FriendUnreadList[snid] != nil {
		return
	}
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		ret, err := model.QueryFriendUnreadBySnid(platform, snid)
		if err != nil {
			return nil
		}
		return ret
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		if data != nil {
			ret := data.(*model.FriendUnread)
			if ret != nil {
				if ret.UnreadSnids != nil && len(ret.UnreadSnids) > 0 {
					this.FriendUnreadList[ret.Snid] = make(map[int32]int)
					for _, us := range ret.UnreadSnids {
						this.FriendUnreadList[ret.Snid][us.SnId] = int(us.UnreadNum)
					}
					this.FriendUnreadDirty[ret.Snid] = false
					this.CheckSendFriendUnreadData(ret.Snid)
				}
			}
		}
	})).StartByFixExecutor("QueryFriendUnreadBySnid")
}

func (this *FriendUnreadMgr) CheckSendFriendUnreadData(snid int32) {
	ful := this.FriendUnreadList[snid]
	if ful == nil || len(ful) == 0 {
		return
	}
	pack := &friend.SCFriendUnreadData{}
	for cSnid, unreadNum := range ful {
		isFriend := FriendMgrSington.IsFriend(snid, cSnid)
		if isFriend {
			fu := &friend.FriendUnread{
				Snid:      proto.Int32(cSnid),
				UnreadNum: proto.Int(unreadNum),
			}
			pack.FriendUnreads = append(pack.FriendUnreads, fu)
		} else {
			delete(ful, cSnid)
			this.FriendUnreadDirty[snid] = true
		}
	}
	if len(pack.FriendUnreads) > 0 {
		proto.SetDefaults(pack)
		p := PlayerMgrSington.GetPlayerBySnId(snid)
		if p != nil {
			p.SendToClient(int(friend.FriendPacketID_PACKET_SCFriendUnreadData), pack)
			logger.Logger.Trace("SCFriendUnreadData: 未读消息列表 pack: ", pack)
		}
	}
}

func (this *FriendUnreadMgr) SaveFriendUnreadData(platform string, snid int32) {
	if this.FriendUnreadDirty[snid] {
		ful := this.FriendUnreadList[snid]
		fu := model.NewFriendUnread(snid)
		for cSnid, unreadNum := range ful {
			if unreadNum > 0 && FriendMgrSington.IsFriend(snid, cSnid) {
				us := &model.FriendUnreadSnid{
					SnId:      cSnid,
					UnreadNum: int32(unreadNum),
				}
				fu.UnreadSnids = append(fu.UnreadSnids, us)
			}
		}
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			ret, _ := model.QueryFriendUnreadBySnid(platform, snid)
			if ret != nil {
				fu.Id = ret.Id
			}
			model.UpsertFriendUnread(platform, snid, fu)
			return nil
		}), nil).StartByFixExecutor("UpsertFriendUnread")
	}
}

func (this *FriendUnreadMgr) ModuleName() string {
	return "FriendUnreadMgr"
}

func (this *FriendUnreadMgr) Init() {
}

func (this *FriendUnreadMgr) Update() {
}

func (this *FriendUnreadMgr) Shutdown() {
	for snid, dirty := range this.FriendUnreadDirty {
		if dirty {
			p := PlayerMgrSington.GetPlayerBySnId(snid)
			if p != nil {
				this.SaveFriendUnreadData(p.Platform, p.SnId)
			}
		}
	}
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(FriendUnreadMgrSington, time.Hour, 0)
}
