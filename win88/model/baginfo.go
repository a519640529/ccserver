package model

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

type BagInfo struct {
	BagId    bson.ObjectId   `bson:"_id"`
	SnId     int32           //玩家账号直接在这里生成
	Platform string          //平台
	BagItem  map[int32]*Item //背包数据  key为itemId
}

type Item struct {
	ItemId     int32 // 物品ID
	ItemNum    int32 // 物品数量
	ObtainTime int64 //获取的时间
}
type GetBagInfoArgs struct {
	Plt  string
	SnId int32
}

func NewBagInfo(sid int32, plt string) *BagInfo {
	return &BagInfo{BagId: bson.NewObjectId(), SnId: sid, Platform: plt, BagItem: make(map[int32]*Item)}
}

func GetBagInfo(sid int32, plt string) *BagInfo {
	if rpcCli == nil {
		return nil
	}
	ret := &BagInfo{}
	args := &GetBagInfoArgs{
		SnId: sid,
		Plt:  plt,
	}
	err := rpcCli.CallWithTimeout("BagSvc.GetBagItem", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("GetBagInfo err:%v SnId:%v ", err, args.SnId)
		return nil
	}
	return ret
}

func UpBagItem(args *BagInfo) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	ret := false
	err := rpcCli.CallWithTimeout("BagSvc.UpgradeBag", args, &ret, time.Second*30)
	if err != nil {
		return fmt.Errorf("UpgradeBag err:%v SnId:%v BagId:%v", err, args.SnId, args.BagId)
	}

	return nil
}

func SaveDBBagItem(args *BagInfo) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	ret := false
	err := rpcCli.CallWithTimeout("BagSvc.AddBagItem", args, &ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

func SaveToDelBackupBagItem(args *BagInfo) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	ret := false
	err := rpcCli.CallWithTimeout("BagSvc.UpdateBag", args, &ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}
