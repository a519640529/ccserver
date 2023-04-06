package main

import (
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
	"strconv"
	"time"
)

var ChatMgrSington = &ChatMgr{
	ChatList: make(map[string]*Chat),
	TsChat:   make(map[int32]int64),
}

type ChatMgr struct {
	ChatList map[string]*Chat
	TsChat   map[int32]int64
}

type Chat struct {
	Id          bson.ObjectId `bson:"_id"`
	BindSnid    string        //snid1 < snid2
	ChatContent []*ChatContent
	LastChatTs  int64
	Dirty       bool
}

type ChatContent struct {
	SrcSnId int32
	Content string
	Ts      int64
}

func (this *ChatMgr) ModuleName() string {
	return "ChatMgr"
}

func (this *ChatMgr) Init() {
}

func (this *ChatMgr) CanSendToPlatform(snid int32) bool {
	if this.TsChat[snid] == 0 {
		this.TsChat[snid] = time.Now().Unix()
		return true
	}
	if time.Now().Unix()-this.TsChat[snid] > 5 {
		this.TsChat[snid] = time.Now().Unix()
		return true
	}
	return false
}

func (this *ChatMgr) getBindSnid(snid, tosnid int32) string {
	bindSnid := strconv.FormatInt(int64(snid), 10) + "_" + strconv.FormatInt(int64(tosnid), 10)
	if snid > tosnid {
		bindSnid = strconv.FormatInt(int64(tosnid), 10) + "_" + strconv.FormatInt(int64(snid), 10)
	}
	return bindSnid
}

// 新增聊天信息
func (this *ChatMgr) AddChat(snid, tosnid int32, content string) {
	bindSnid := this.getBindSnid(snid, tosnid)
	logger.Logger.Trace("(this *ChatMgr) AddChat ", bindSnid)
	cl := this.ChatList[bindSnid]
	if cl == nil {
		cacheCl := this.db2cache(model.NewChat(bindSnid))
		cl = &cacheCl
		this.ChatList[bindSnid] = cl
	}
	cc := &ChatContent{
		SrcSnId: snid,
		Content: content,
		Ts:      time.Now().Unix(),
	}
	cl.ChatContent = append(cl.ChatContent, cc)
	cl.LastChatTs = time.Now().Unix()
	cl.Dirty = true
}

// 删除聊天信息
func (this *ChatMgr) DelChat(platform string, snid, tosnid int32) {
	bindSnid := this.getBindSnid(snid, tosnid)
	logger.Logger.Trace("(this *ChatMgr) DelChat ", bindSnid)
	if this.ChatList[bindSnid] == nil {
		return
	}
	delete(this.ChatList, bindSnid)
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		model.DelChat(platform, bindSnid)
		return nil
	}), nil).StartByFixExecutor("DelChat")
}

// 聊天信息
func (this *ChatMgr) SetChat(bindSnid string, chat *model.Chat) {
	logger.Logger.Trace("(this *ChatMgr) SetChat ", bindSnid)
	if this.ChatList[bindSnid] == nil {
		cacheChat := this.db2cache(chat)
		this.ChatList[bindSnid] = &cacheChat
	}
}

// 获取聊天信息
func (this *ChatMgr) GetChat(snid, tosnid int32) *Chat {
	bindSnid := this.getBindSnid(snid, tosnid)
	logger.Logger.Trace("(this *ChatMgr) GetChat ", bindSnid, this.ChatList[bindSnid])
	return this.ChatList[bindSnid]
}

func (this *ChatMgr) SaveChatData(platform string, snid, tosnid int32) {
	bindSnid := this.getBindSnid(snid, tosnid)
	logger.Logger.Trace("(this *ChatMgr) SaveChatData ", bindSnid)
	chat := this.ChatList[bindSnid]
	if chat != nil && chat.Dirty {
		chat.Dirty = false
		dbchat := this.cache2db(chat)
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.UpsertChat(platform, bindSnid, &dbchat)
		}), nil, "SaveChatData").StartByFixExecutor("SnId:" + strconv.Itoa(int(snid)))
	}
}

// db -> cache
func (this *ChatMgr) db2cache(chat *model.Chat) (c Chat) {
	if chat != nil {
		c.Id = chat.Id
		c.BindSnid = chat.BindSnid
		if chat.ChatContent != nil {
			for _, content := range chat.ChatContent {
				cc := &ChatContent{
					SrcSnId: content.SrcSnId,
					Content: content.Content,
					Ts:      content.Ts,
				}
				c.ChatContent = append(c.ChatContent, cc)
			}
		}
		c.LastChatTs = chat.LastChatTs
	}
	return c
}

// cache -> db
func (this *ChatMgr) cache2db(chat *Chat) (c model.Chat) {
	if chat != nil {
		c.Id = chat.Id
		c.BindSnid = chat.BindSnid
		if chat.ChatContent != nil {
			for _, content := range chat.ChatContent {
				cc := &model.ChatContent{
					SrcSnId: content.SrcSnId,
					Content: content.Content,
					Ts:      content.Ts,
				}
				c.ChatContent = append(c.ChatContent, cc)
			}
		}
		c.LastChatTs = chat.LastChatTs
	}
	return c
}

func (this *ChatMgr) Update() {
}

func (this *ChatMgr) Shutdown() {
	for _, platform := range PlatformMgrSington.Platforms {
		if platform.IdStr == "0" {
			continue
		}
		for bindSnid, chat := range this.ChatList {
			dbchat := this.cache2db(chat)
			model.UpsertChat(platform.IdStr, bindSnid, &dbchat)
		}
	}
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(ChatMgrSington, time.Second, 0)
}
