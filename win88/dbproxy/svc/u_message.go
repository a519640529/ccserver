package svc

import (
	"errors"
	"net/rpc"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

func MessageCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.MessageDBName)
	if s != nil {
		c, first := s.DB().C(model.MessageCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"state"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type MessageSvc struct {
}

var _MessageSvc = &MessageSvc{}

func init() {
	rpc.Register(_MessageSvc)
}

func (svc *MessageSvc) GetMessage(args *model.GetMessageArgs, ret *[]model.Message) error {
	logger.Logger.Info("MessageSvc GetMessage")
	c := MessageCollection(args.Plt)
	if c != nil {
		var sql []bson.M
		sql = append(sql, bson.M{"state": 0})
		sql = append(sql, bson.M{"state": 1})
		err := c.Find(bson.M{"snid": args.SnId, "$or": sql}).All(ret)
		return err
	}
	return nil
}

func (svc *MessageSvc) GetNotDelMessage(args *model.GetMessageArgs, ret *[]model.Message) error {
	logger.Logger.Info("MessageSvc GetMessage")
	c := MessageCollection(args.Plt)
	if c != nil {
		err := c.Find(bson.M{"snid": args.SnId, "state": bson.M{"$ne": model.MSGSTATE_REMOVEED}}).All(ret)
		return err
	}
	return nil
}

func (svc *MessageSvc) GetSubscribeMessage(args string, ret *[]model.Message) error {
	logger.Logger.Info("MessageSvc GetSubscribeMessage")
	c := MessageCollection(args)
	if c != nil {
		err := c.Find(bson.M{"snid": 0}).All(ret)
		return err
	}
	return nil
}
func (svc *MessageSvc) GetMessageInRangeTsLimitByRange(args *model.MsgArgs, ret *model.RetMsg) error {
	logger.Logger.Info("MessageSvc GetMessageInRangeTsLimitByRange")
	c := MessageCollection(args.Platform)
	if c == nil {
		return errors.New("c == nil")
	}
	limitDataNum := args.ToIndex - args.FromIndex
	if limitDataNum < 0 {
		limitDataNum = 0
	}
	var selecter bson.M
	if args.StartTs == 0 && args.EndTs == 0 {
		if len(args.Platform) == 0 {
			selecter = bson.M{"mtype": args.MsgType}
		} else {
			selecter = bson.M{"mtype": args.MsgType, "platform": args.Platform}
		}
	} else {
		if len(args.Platform) == 0 {
			selecter = bson.M{"mtype": args.MsgType, "creatts": bson.M{"$gte": args.StartTs, "$lte": args.EndTs}}
		} else {
			selecter = bson.M{"mtype": args.MsgType, "platform": args.Platform, "creatts": bson.M{"$gte": args.StartTs, "$lte": args.EndTs}}
		}
	}
	if args.DestSnId != -1 {
		selecter["snid"] = args.DestSnId
	}
	//下面的Count和Skip均是Mgo中比较耗时的操作，根据线上日志来看两个语句执行时间已经超过10s(具体mgo中数据量保密)
	//那重点就根据业务优化下面两句即可
	count, _ := c.Find(selecter).Count()
	err := c.Find(selecter).Skip(args.FromIndex).Limit(limitDataNum).All(&ret.Msg)
	ret.Count = count
	return err
}

func (svc *MessageSvc) ReadMessage(args *model.ReadMsgArgs, ret *bool) error {
	logger.Logger.Info("MessageSvc ReadMessage")
	c := MessageCollection(args.Platform)
	if c != nil {
		err := c.Update(bson.M{"_id": args.Id}, bson.D{{"$set", bson.D{{"state", model.MSGSTATE_READED}}}})
		if err != nil {
			return err
		}
		*ret = true
	}
	return nil
}
func (svc *MessageSvc) DelMessage(args *model.DelMsgArgs, ret *bool) error {
	logger.Logger.Info("MessageSvc DelMessage")
	c := MessageCollection(args.Platform)
	if c != nil {
		var err error
		if args.Del == 0 {
			err = c.Remove(bson.M{"_id": args.Id}) // 目前需求
		} else {
			err = c.Update(bson.M{"_id": args.Id}, bson.D{{"$set", bson.D{{"state", model.MSGSTATE_REMOVEED}}}})
		}
		if err != nil {
			return err
		}
		*ret = true
	}
	return nil
}
func (svc *MessageSvc) DelAllMessage(args *model.DelAllMsgArgs, ret *bool) error {
	logger.Logger.Info("MessageSvc DelMessage")
	c := MessageCollection(args.Platform)
	if c != nil {
		//for _, id := range args.Ids {
		//	var err error
		//	if args.Del == 0 {
		//		err = c.Remove(bson.M{"_id": id}) // 目前需求
		//	} else {
		//		err = c.Update(bson.M{"_id": id}, bson.D{{"$set", bson.D{{"state", model.MSGSTATE_REMOVEED}}}})
		//	}
		//	if err != nil {
		//		return err
		//	}
		//}
		//*ret = true
		_, err := c.RemoveAll(bson.M{"oper": 1, "state": model.MSGSTATE_REMOVEED})
		err = c.Remove(bson.M{"oper": 1, "state": model.MSGSTATE_REMOVEED})
		if err != nil && err != mgo.ErrNotFound {
			logger.Logger.Error("DelAllMessage:", err)
			return err
		}
		*ret = true
	}
	return nil
}
func (svc *MessageSvc) GetMessageAttach(args *model.AttachMsgArgs, ret *bool) error {
	logger.Logger.Info("MessageSvc GetMessageAttach")
	c := MessageCollection(args.Platform)
	if c != nil {
		err := c.Update(bson.M{"_id": args.Id}, bson.D{{"$set", bson.D{{"attachstate", model.MSGATTACHSTATE_GOT}}}})
		if err != nil {
			return err
		}
		*ret = true
	}
	return nil
}
func (svc *MessageSvc) GetMessageById(args *model.GetMsgArg, ret *model.Message) error {
	logger.Logger.Info("MessageSvc GetMessageById")
	c := MessageCollection(args.Platform)
	if c != nil {
		Id := bson.ObjectIdHex(args.IdStr)
		err := c.Find(bson.M{"_id": Id}).One(ret)
		return err
	}
	return nil
}

func (svc *MessageSvc) GetMessageAttachs(args *model.AttachMsgsArgs, ret *[]string) error {
	logger.Logger.Info("MessageSvc GetMessageAttachs")
	c := MessageCollection(args.Platform)
	if c != nil {
		for _, Idstr := range args.Ids {
			Id := bson.ObjectIdHex(Idstr)
			var msg model.Message
			if err := c.Find(bson.M{"_id": Id}).One(&msg); err == nil {
				if msg.AttachState == model.MSGATTACHSTATE_GOT {
					continue
				}
				err := c.Update(bson.M{"_id": Id}, bson.D{{"$set", bson.D{{"attachstate", model.MSGATTACHSTATE_GOT}}}})
				if err != nil {
					logger.Logger.Infof("MessageSvc GetMessageAttachs attachstate err %v", err)
					continue
				}

				*ret = append(*ret, msg.Id.Hex())
			}

		}
		return nil
	}
	return errors.New("redis is nil")
}

func (svc *MessageSvc) InsertMessage(args *model.InsertMsgArg, ret *bool) error {
	//logger.Logger.Trace("MessageSvc InsertMessage")
	c := MessageCollection(args.Platform)
	var err error
	if c != nil {
		if len(args.Msgs) == 1 {
			err = c.Insert(args.Msgs[0])
		} else if len(args.Msgs) > 1 {
			docs := make([]interface{}, 0, len(args.Msgs))
			for _, msg := range args.Msgs {
				docs = append(docs, msg)
			}
			err = c.Insert(docs...)
		}
		if err != nil {
			return err
		}
		*ret = true
	}
	return nil
}
func (svc *MessageSvc) RemoveMessages(args *model.AttachMsgArgs, ret *mgo.ChangeInfo) (err error) {
	logger.Logger.Info("MessageSvc GetMessageAttach")
	c := MessageCollection(args.Platform)
	if c != nil {
		ret, err = c.RemoveAll(bson.M{"creatts": bson.M{"$lt": args.Ts}})
	}
	return
}
