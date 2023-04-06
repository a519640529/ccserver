package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"sort"
	"time"
)

type Chat struct {
	Id          bson.ObjectId `bson:"_id"`
	BindSnid    string        //snid1 < snid2
	ChatContent []*ChatContent
	LastChatTs  int64
}

type ChatContent struct {
	SrcSnId int32
	Content string
	Ts      int64
}

const ChatMaxNum = 1000 //聊天记录上限

type ChatRet struct {
	C *Chat
}

type ChatByKey struct {
	Platform string
	BindSnId string
	C        *Chat
}

func NewChat(bindsnid string) *Chat {
	f := &Chat{Id: bson.NewObjectId()}
	f.BindSnid = bindsnid
	f.LastChatTs = time.Now().Unix()
	f.ChatContent = []*ChatContent{}
	return f
}

func UpsertChat(platform, bindsnid string, chat *Chat) *Chat {
	if rpcCli == nil {
		logger.Logger.Error("model.UpsertChat rpcCli == nil")
		return nil
	}
	if chat != nil && chat.ChatContent != nil && len(chat.ChatContent) > ChatMaxNum {
		sort.Slice(chat.ChatContent, func(i, j int) bool {
			if chat.ChatContent[i].Ts < chat.ChatContent[j].Ts {
				return false
			}
			return true
		})
		chat.ChatContent = append(chat.ChatContent[:ChatMaxNum])
	}
	args := &ChatByKey{
		Platform: platform,
		BindSnId: bindsnid,
		C:        chat,
	}
	ret := &ChatRet{}
	err := rpcCli.CallWithTimeout("ChatSvc.UpsertChat", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("UpsertChat error:", err)
		return nil
	}
	return ret.C
}

func QueryChatByBindSnid(platform, bindsnid string) (chat *Chat, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryChatByBindSnid rpcCli == nil")
		return
	}
	args := &ChatByKey{
		Platform: platform,
		BindSnId: bindsnid,
	}
	var ret *ChatRet
	err = rpcCli.CallWithTimeout("ChatSvc.QueryChatByKey", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("QueryChatByBindSnid error:", err)
	}
	if ret != nil {
		chat = ret.C
	}
	return
}

func DelChat(platform, bindsnid string) {
	if rpcCli == nil {
		logger.Logger.Error("model.DelChat rpcCli == nil")
		return
	}
	args := &ChatByKey{
		Platform: platform,
		BindSnId: bindsnid,
	}
	err := rpcCli.CallWithTimeout("ChatSvc.DelChat", args, nil, time.Second*30)
	if err != nil {
		logger.Logger.Error("DelChat error:", err)
	}
	return
}
