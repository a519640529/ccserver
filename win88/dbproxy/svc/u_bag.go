package svc

import (
	"errors"
	"net/rpc"
	"time"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	BagDBName       = "user"
	BagCollName     = "user_bag"
	ErrBagDBNotOpen = model.NewDBError(BagDBName, BagCollName, model.NOT_OPEN)
)

func BagCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, BagDBName)
	if s != nil {
		c, first := s.DB().C(BagCollName)
		if first {
			// c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type BagSvc struct {
}

func (svc *BagSvc) GetBagItem(args *model.GetBagInfoArgs, ret *model.BagInfo) error {
	cbag := BagCollection(args.Plt)
	if cbag == nil {
		return errors.New("cbag is nil")
	}
	err := cbag.Find(bson.M{"snid": args.SnId, "platform": &args.Plt}).One(ret)
	if err != nil && err != mgo.ErrNotFound {
		//if err == mgo.ErrNotFound {
		//	ret = model.NewBagInfo(args.SnId, args.Plt)
		//	err = cbag.Insert(ret) // 考虑创建失败~
		//	logger.Logger.Error("svc.GetBagItem is Insert", err)
		//	return nil
		//}
		logger.Logger.Error("svc.GetBagItem is error: ", err)
		return nil
	}
	return nil
}

func (svc *BagSvc) UpgradeBag(args *model.BagInfo, ret *bool) error {
	cbag := BagCollection(args.Platform)
	if cbag == nil {
		return ErrBagDBNotOpen
	}
	*ret = true
	bag := &model.BagInfo{}
	err := cbag.Find(bson.M{"snid": args.SnId}).One(bag)
	if err != nil && err != mgo.ErrNotFound {
		*ret = false
		logger.Logger.Error("UpgradeBag err:", err)
		return err
	}
	args.BagId = bag.BagId
	if args.BagId == "" {
		args.BagId = bson.NewObjectId()
	}
	_, err = cbag.Upsert(bson.M{"_id": args.BagId}, args)
	if err != nil {
		*ret = false
		logger.Logger.Info("UpgradeBag error ", err)
	}
	return err
}

func (svc *BagSvc) AddBagItem(args *model.BagInfo, ret *bool) error {
	cbag := BagCollection(args.Platform)
	if cbag == nil {
		return ErrBagDBNotOpen
	}
	*ret = true
	bag := &model.BagInfo{}
	err := cbag.Find(bson.M{"snid": args.SnId}).One(bag)
	if err != nil && err != mgo.ErrNotFound {
		*ret = false
		logger.Logger.Error("AddBagItem err:", err)
		return err
	}

	if bag.BagId == "" {
		bag.BagId = bson.NewObjectId()
	}
	for id, v := range args.BagItem {
		if item, exist := bag.BagItem[id]; !exist {
			bag.BagItem[id] = &model.Item{
				ItemId:     v.ItemId,
				ItemNum:    v.ItemNum,
				ObtainTime: time.Now().Unix(),
			}
		} else {
			item.ItemNum += v.ItemNum
		}
	}
	_, err = cbag.Upsert(bson.M{"_id": bag.BagId}, bag)
	if err != nil {
		*ret = false
		logger.Logger.Info("AddBagItem error ", err)
	}
	return err
}

func (svc *BagSvc) UpdateBag(args *model.BagInfo, ret *bool) error {
	cbag := BagCollection(args.Platform)
	if cbag == nil {
		*ret = false
		return ErrBagDBNotOpen
	}

	*ret = true
	err := cbag.UpdateId(args.BagId, args)
	if err != nil {
		*ret = false
		logger.Logger.Info("UpdateBag error:", err)
	}

	return err
}

var _BagSvc = &BagSvc{}

func init() {
	rpc.Register(_BagSvc)
}
