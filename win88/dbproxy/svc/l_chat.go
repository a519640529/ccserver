package svc

import (
	"errors"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"net/rpc"
)

var (
	ChatDBName   = "log"
	ChatCollName = "log_chat"
	ChatColError = errors.New("chat collection open failed")
)

func ChatCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, ChatDBName)
	if s != nil {
		c, first := s.DB().C(ChatCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"bindsnid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type ChatSvc struct {
}

func (svc *ChatSvc) UpsertChat(args *model.ChatByKey, ret *model.ChatRet) error {
	cc := ChatCollection(args.Platform)
	if cc == nil {
		return ChatColError
	}
	_, err := cc.Upsert(bson.M{"bindsnid": args.BindSnId}, args.C)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("UpsertChat is err: ", err)
		return err
	}
	ret.C = args.C
	return nil
}

func (svc *ChatSvc) QueryChatByKey(args *model.ChatByKey, ret *model.ChatRet) error {
	fc := ChatCollection(args.Platform)
	if fc == nil {
		return ChatColError
	}
	err := fc.Find(bson.M{"bindsnid": args.BindSnId}).One(&ret.C)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("QueryChatByKey is err: ", err)
		return err
	}
	return nil
}

func (svc *ChatSvc) DelChat(args *model.ChatByKey, ret *bool) error {
	cc := ChatCollection(args.Platform)
	if cc == nil {
		return ChatColError
	}
	err := cc.Remove(bson.M{"bindsnid": args.BindSnId})
	if err != nil {
		logger.Logger.Warn("DelChat is err: ", err)
		return err
	}
	return nil
}

var _ChatSvc = &ChatSvc{}

func init() {
	rpc.Register(_ChatSvc)
}
