package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

var (
	DbShopDBName   = "log"
	DbShopCollName = "log_dbshop"
)

type DbShop struct {
	LogId    bson.ObjectId `bson:"_id"`
	Platform string        //平台
	ShopId   int32         //商品id
	ShopName string        //商品名称
	SnId     int32         //兑换玩家
	Count    int32         //兑换数量
	State    int32         //状态 0.等待审核 1.已通过 2.未通过
	UserName string        //姓名
	UserTel  string        //手机号
	Remark   string        //备注信息
	Operator string        //操作人
	CreateTs time.Time     //订单生成时间
	OpTs     time.Time     //订单最后操作时间
}
type DbShopLogArgs struct {
	Platform string
	Log      *DbShop
}

func NewDbShop(platform string, shopId int32, shopName string, snId, count, state int32, userName, userTel, remark, operator string) *DbShop {
	return &DbShop{
		LogId:    bson.NewObjectId(),
		Platform: platform,
		ShopId:   shopId,
		ShopName: shopName,
		SnId:     snId,
		Count:    count,
		State:    state,
		UserName: userName,
		UserTel:  userTel,
		Remark:   remark,
		Operator: operator,
		CreateTs: time.Now(),
		OpTs:     time.Now(),
	}
}
func InsertDbShopLog(platform string, log *DbShop) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.InsertDbShopLog rpcCli == nil")
		return
	}
	var ret bool
	args := &DbShopLogArgs{
		Platform: platform,
		Log:      log,
	}
	err = rpcCli.CallWithTimeout("DbShopLogSvc.InsertDbShopLog", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("InsertDbShopLog error:", err)
	}
	return
}
