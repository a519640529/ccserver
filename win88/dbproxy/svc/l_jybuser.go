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
	JybUserDBName       = "log"
	JybUserCollName     = "log_jybuser"
	ErrJybUserDBNotOpen = model.NewDBError(JybUserDBName, JybUserCollName, model.NOT_OPEN)
)

func JybuserCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, JybUserDBName)
	if s != nil {
		c, first := s.DB().C(JybUserCollName)
		if first {
			// c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type JybUserSvc struct {
}

// 验证+
func (svc *JybUserSvc) VerifyJybUser(args *model.GetJybInfoArgs, ret *bool) error {
	cjybuse := JybuserCollection(args.Plt)
	if cjybuse == nil {
		return ErrJybUserDBNotOpen
	}
	cjyb := JybCollection(args.Plt)
	if cjyb == nil {
		return ErrJybDBNotOpen
	}
	jyb := &model.JybInfo{}
	err := cjyb.Find(bson.M{"_id": bson.ObjectIdHex(args.Id)}).One(jyb)
	if err != nil {
		logger.Logger.Errorf("VerifyJybUser error %v %v", args, err)
		return err
	}
	if _, exist := jyb.Code[args.UseCode]; !exist { // 该兑换码已经被领取
		return model.ErrJybISReceive
	}

	jybuser := &model.JybUserInfo{}
	err = cjybuse.Find(bson.M{"snid": args.SnId, "platform": &args.Plt}).One(jybuser)
	if err != nil {
		if err == mgo.ErrNotFound {
			info := model.NewJybUserInfo(args.SnId, args.Plt)
			err = cjybuse.Insert(info) // 考虑创建失败~
			if err != nil {
				logger.Logger.Error("svc.VerifyJybUser is Insert err ", err)

				return err
			}
			*ret = true // 说明没有领取过礼包 直接返回
			return nil
		}
		logger.Logger.Errorf("VerifyJybUser error ", args, err)
		return err
	}
	if _, exist := jybuser.JybInfos[args.Id]; exist { // 该类型礼包玩家已经领取过
		return model.ErrJYBPlCode
	}
	// 可以领取
	*ret = true
	return nil
}

// 验证+更新
func (svc *JybUserSvc) VerifyUpJybUser(args *model.VerifyUpJybInfoArgs, ret *model.JybInfo) error {
	cjybuse := JybuserCollection(args.Plt)
	if cjybuse == nil {
		return ErrJybUserDBNotOpen
	}
	cjyb := JybCollection(args.Plt)
	if cjyb == nil {
		return ErrJybDBNotOpen
	}
	// jyb := &model.JybInfo{}
	if args.CodeType != 1 {
		err := cjyb.Find(bson.M{"codestart": args.CodeStart}).One(ret)
		if err != nil {
			logger.Logger.Errorf("VerifyJybUser error  JybInfo %v %v", args.CodeStart, err)
			return model.ErrJybIsNotExist
		}
		if ret.StartTime > 0 && ret.EndTime > 0 {
			ts := time.Now().Unix()
			if ts < ret.StartTime || ts > ret.EndTime { // 未开始
				return model.ErrJybTsTimeErr
			}
		}
		if _, exist := ret.Code[args.UseCode]; !exist { // 该兑换码已经被领取
			return model.ErrJybISReceive
		}
	} else {
		jk := &model.JybKey{}
		if jk = getJyb(cjyb, args.Plt); jk != nil {

		} else {
			logger.Logger.Errorf("DelJybjybkey is nil ")
			return errors.New("jybkey is nil")
		}

		if jk.GeCode == nil { // 不存在通用礼包
			logger.Logger.Error("VerifyUpJybUser NewJybKey GeCode is nil")
			return model.ErrJybIsNotExist
		} else if id, exist := jk.GeCode[args.UseCode]; !exist {
			logger.Logger.Error("VerifyUpJybUser NewJybKey Find err")
			return model.ErrJybIsNotExist
		} else {

			cid := bson.ObjectIdHex(id)
			err := cjyb.Find(bson.M{"_id": cid}).One(ret)
			if err != nil {
				logger.Logger.Errorf("VerifyJybUser error id %v %v %v", id, args, err)
				return err
			}
		}
		if ret.StartTime > 0 && ret.EndTime > 0 {
			ts := time.Now().Unix()
			if ts < ret.StartTime || ts > ret.EndTime { // 未开始
				return model.ErrJybTsTimeErr
			}
		}
	}

	return upJybUser(cjybuse, cjyb, args.SnId, args.CodeType, args.Plt, args.UseCode, ret)
}

func upJybUser(cjybuse, cjyb *mongo.Collection, snId, codeType int32, plt, useCode string, ret *model.JybInfo) error {
	jybuser := &model.JybUserInfo{}
	err := cjybuse.Find(bson.M{"snid": snId, "platform": plt}).One(jybuser)
	if err != nil {
		if err == mgo.ErrNotFound {
			jybuser = model.NewJybUserInfo(snId, plt)
			err = cjybuse.Insert(jybuser) // 考虑创建失败~
			if err != nil {
				logger.Logger.Error("svc.VerifyJybUser is Insert err ", err)

				return err
			}
		} else {
			logger.Logger.Errorf("VerifyJybUser error ", err)
			return err
		}
	}
	jybuseerid := ret.JybId.Hex()
	if jybuser.JybInfos == nil {
		jybuser.JybInfos = make(map[string]int32)
	} else if _, exist := jybuser.JybInfos[jybuseerid]; exist { // 该类型礼包玩家已经领取过
		return model.ErrJYBPlCode
	}

	jybuser.JybInfos[jybuseerid] = 1

	err = cjybuse.Update(bson.M{"_id": jybuser.JybUserId}, bson.D{{"$set", bson.D{{"jybinfos", jybuser.JybInfos}}}})
	if err != nil {

		logger.Logger.Info("UpgradeJyb error ", err)
		return err
	}
	if codeType != 1 {
		delete(ret.Code, useCode) // 局部map不存在并发安全问题 重复删除问题无法解决
	}
	ret.Receive++
	err = cjyb.Update(bson.M{"_id": ret.JybId}, bson.D{{"$set", bson.D{{"code", ret.Code}, {"receive", ret.Receive}}}})
	if err != nil {

		logger.Logger.Errorf("UpgradeJyb error ", err)

	}
	return err
}

// 更新记录玩家领取过的礼包 先只是记录id 更新之前必须先验证
func (svc *JybUserSvc) UpgradeJybUser(args *model.GetJybInfoArgs, ret *model.JybInfo) error {
	cjybuse := JybuserCollection(args.Plt)
	if cjybuse == nil {
		return ErrJybUserDBNotOpen
	}
	cjyb := JybCollection(args.Plt)
	if cjyb == nil {
		return ErrJybDBNotOpen
	}
	cid := bson.ObjectIdHex(args.Id)
	err := cjyb.Find(bson.M{"_id": cid}).One(ret)
	if err != nil {
		logger.Logger.Errorf("VerifyJybUser error id %v %v %v", args.Id, args, err)
		return err
	}
	return upJybUser(cjybuse, cjyb, args.SnId, args.CodeType, args.Plt, args.UseCode, ret)
}

// 回退
func (svc *JybUserSvc) BackOffJybUser(args *model.GetJybInfoArgs, ret *bool) error {
	cjybuse := JybuserCollection(args.Plt)
	if cjybuse == nil {
		return ErrJybUserDBNotOpen
	}
	cjyb := JybCollection(args.Plt)
	if cjyb == nil {
		return ErrJybDBNotOpen
	}

	jyb := &model.JybInfo{}
	err := cjyb.Find(bson.M{"_id": bson.ObjectIdHex(args.Id)}).One(jyb)
	if err != nil {
		logger.Logger.Errorf("BackOffJybUser error ", err)
		return err
	}
	jyb.Code[args.UseCode] = 1 // 将该礼包存入
	jyb.Receive--
	err = cjyb.Update(bson.M{"_id": jyb.JybId}, bson.D{{"$set", bson.D{{"code", jyb.Code}, {"receive", jyb.Receive}}}})

	jybuser := &model.JybUserInfo{}
	err = cjybuse.Find(bson.M{"snid": args.SnId, "platform": &args.Plt}).One(jybuser)
	if err != nil {
		if err == mgo.ErrNotFound { //
			info := model.NewJybUserInfo(args.SnId, args.Plt)
			err = cjybuse.Insert(info) // 考虑创建失败~
			if err != nil {
				logger.Logger.Error("svc.BackOffJybUser is Insert err ", err)

				return err
			}
			*ret = true // 说明没有领取过礼包 直接返回
			return nil
		}
		logger.Logger.Errorf("BackOffJybUser error ", err)
		return err
	}

	delete(jybuser.JybInfos, args.Id) // 删除该玩家领取记录
	err = cjybuse.Update(bson.M{"_id": jybuser.JybUserId}, bson.D{{"$set", bson.D{{"jybinfos", jybuser.JybInfos}}}})

	return err
}

var _JybUserSvc = &JybUserSvc{}

func init() {
	rpc.Register(_JybUserSvc)
}
