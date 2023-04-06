package svc

import (
	"errors"
	"fmt"
	"net/rpc"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	JybDBName       = "log"
	JybCollName     = "log_jyb"
	ErrJybDBNotOpen = model.NewDBError(JybDBName, JybCollName, model.NOT_OPEN)
	// _ids            = make(map[string]bson.ObjectId)
)

func JybCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, JybDBName)
	if s != nil {
		c, first := s.DB().C(JybCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"plakeyid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"codestart"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type JybSvc struct {
}

// 初始化
func (svc *JybSvc) InitJybItem(args *model.InitJybInfoArgs, ret *model.JybInfos) error {
	cjyb := JybCollection(args.Plt)
	if cjyb == nil {
		return errors.New("cjyb is nil")
	}
	err := cjyb.Find(bson.M{"platform": args.Plt}).All(&ret.Jybs)
	if err != nil {

		logger.Logger.Error("svc.InitJybItem is error", err)
		return err
	}
	return nil
}

// 创建礼包
func (svc *JybSvc) CreateJybItem(args *model.CreateJyb, ret *model.JybInfo) error {
	cjyb := JybCollection(args.Platform)
	if cjyb == nil {
		return errors.New("cjyb is nil")
	}

	jk := &model.JybKey{}
	if jk = getJyb(cjyb, args.Platform); jk == nil {
		logger.Logger.Errorf("DelJybjybkey is nil ")
		return errors.New("jybkey is nil")
	}

	if args.CodeType != 1 { // 通用红包不用生成code
		if err := cjyb.Update(bson.M{"_id": jk.KeyId}, bson.D{{"$set", bson.D{{"keyint", jk.Keyint + model.Keystart}}}}); err != nil {
			logger.Logger.Info("CreateJybItem Update error ", err)
			return err
		}
		args.CodeStart = jk.Keyint
		if args.Codelen == 0 {
			args.Codelen = 12
		}
		model.NewJybCode(args.JybInfo, args.Codelen, args.Num)
	} else {
		code := ""
		for id := range args.Code {
			code = id
		}
		if jk.GeCode == nil {
			jk.GeCode = make(map[string]string)
		} else if _, exist := jk.GeCode[code]; exist {
			logger.Logger.Errorf("CreateJybItem error code is exist")
			return model.ErrJYBCode //errors.New("code is exist")
		}
		jk.GeCode[code] = args.JybId.Hex()
		if err := cjyb.Update(bson.M{"_id": jk.KeyId}, bson.D{{"$set", bson.D{{"gecode", jk.GeCode}}}}); err != nil {
			logger.Logger.Info("CreateJybItem Update error ", err)
			return err
		}
	}
	err := cjyb.Insert(args.JybInfo)
	ret = args.JybInfo
	if err != nil {
		logger.Logger.Error("svc.GetJybItem is error: ", err)
		return err
	}
	//jbf := &model.JybInfo{}
	//err = cjyb.Find(bson.M{"_id": args.JybId}).One(jbf)
	//if err != nil {
	//	logger.Logger.Error("svc.GetJybItem is error", err)
	//}
	return nil
}

func (svc *JybSvc) GetJybItem(args *model.GetJybInfoArgs, ret *model.JybInfo) error {
	cjyb := JybCollection(args.Plt)
	if cjyb == nil {
		return errors.New("cjyb is nil")
	}
	id := bson.ObjectIdHex(args.Id)
	err := cjyb.Find(bson.M{"_id": id}).One(ret)
	if err != nil {

		logger.Logger.Error("svc.GetJybItem is error", err)
		return err
	}
	return nil
}

// getJyb .
func getJyb(cjyb *mongo.Collection, plt string) *model.JybKey {
	jk := &model.JybKey{}

	id := fmt.Sprintf("%d_%s", model.Keystart, plt)
	err := cjyb.Find(bson.M{"plakeyid": id}).One(jk)
	if err != nil {
		if err == mgo.ErrNotFound {
			jk = model.NewJybKey(plt)
			err = cjyb.Insert(jk) // 考虑创建失败~
			if err != nil {
				logger.Logger.Error("CreateJybItemis NewJybKey Insert", err)
				return nil
			}
		} else {
			logger.Logger.Error("CreateJybItemis NewJybKey Find 1 ", err)
			return nil
		}
	}

	return jk
}

/*
// getJyb .
func getJyb(cjyb *mongo.Collection, plt string) *model.JybKey {
	jk := &model.JybKey{}
	_id := _ids[plt]
	if _id == "" {
		id := fmt.Sprintf("%d_%s", model.Keystart, plt)
		err := cjyb.Find(bson.M{"plakeyid": id}).One(jk)
		if err != nil {
			if err == mgo.ErrNotFound {
				jk = model.NewJybKey(plt)
				err = cjyb.Insert(jk) // 考虑创建失败~
				if err != nil {
					logger.Logger.Error("CreateJybItemis NewJybKey Insert", err)
					return nil
				}
			} else {
				logger.Logger.Error("CreateJybItemis NewJybKey Find 1 ", err)
				return nil
			}
		}
		_ids[plt] = jk.KeyId
	} else {
		err := cjyb.Find(bson.M{"_id": _id}).One(jk)
		if err == mgo.ErrNotFound { // 重新查找 找不到不走重新创建
			id := fmt.Sprintf("%d_%s", model.Keystart, plt)
			err = cjyb.Find(bson.M{"plakeyid": id}).One(jk)
			if err == nil {
				_ids[plt] = jk.KeyId
			}
		}
		if err != nil {
			logger.Logger.Errorf("CreateJybItemis NewJybKey Find id:%v %v", _id.Hex(), err)
			return nil
		}
	}
	return jk
}*/

func (svc *JybSvc) DelJyb(args *model.GetJybInfoArgs, ret *bool) error {
	cjyb := JybCollection(args.Plt)
	if cjyb == nil {
		return ErrJybDBNotOpen
	}
	jbf := &model.JybInfo{}
	err := cjyb.Find(bson.M{"_id": bson.ObjectIdHex(args.Id)}).One(jbf)
	if err == nil && jbf.CodeType == 1 {
		jk := &model.JybKey{}
		if jk = getJyb(cjyb, args.Plt); jk != nil {

		} else {
			logger.Logger.Errorf("DelJybjybkey is nil ")
			return errors.New("jybkey is nil")
		}

		for code, _ := range jbf.Code {
			delete(jk.GeCode, code)
		}

		if err := cjyb.Update(bson.M{"_id": jk.KeyId}, bson.D{{"$set", bson.D{{"gecode", jk.GeCode}}}}); err != nil {
			logger.Logger.Errorf("DelJyb getJyb Update error ", err)
			return err
		}
	}
	err = cjyb.Remove(bson.M{"_id": bson.ObjectIdHex(args.Id)})
	return err
}

//  TODO
func (svc *JybSvc) DelCodeJyb(args *model.GetJybInfoArgs, ret *bool) error {
	cjyb := JybCollection(args.Plt)
	if cjyb == nil {
		return ErrJybDBNotOpen
	}

	*ret = true
	jyb := &model.JybInfo{}
	err := cjyb.Find(bson.M{"_id": bson.ObjectIdHex(args.Id)}).One(jyb)

	if jyb.CodeType == 1 {
		logger.Logger.Infof("")
		return nil
	}
	if _, exist := jyb.Code[args.UseCode]; !exist { // 已经领取
		return model.ErrJybISReceive
	}
	delete(jyb.Code, args.UseCode) // 局部map不存在并发安全问题 重复删除问题无法解决
	jyb.Receive++
	err = cjyb.Update(bson.M{"_id": jyb.JybId}, bson.D{{"$set", bson.D{{"code", jyb.Code}, {"receive", jyb.Receive}}}})
	if err != nil {
		*ret = false
		logger.Logger.Errorf("UpgradeJyb error ", err)
	}

	return err
}

var _JybSvc = &JybSvc{}

func init() {
	rpc.Register(_JybSvc)
}
