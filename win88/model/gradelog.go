package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

var (
	GradeLogDBName   = "log"
	GradeLogCollName = "log_grade"
)

const GradeLogMaxLimitPerQuery = 100

type GradeLog struct {
	LogId      bson.ObjectId `bson:"_id"`
	SnId       int32
	Platform   string //平台名称
	Channel    string //渠道名称
	Promoter   string //推广员
	PackageTag string //推广包标识
	InviterId  int32  //邀请人id
	Count      int64
	RestCount  int64
	Oper       string
	Remark     string
	Time       time.Time
	Ver        int32
	LogType    int32
	Status     int32 //状态 0 默认  1超时
}

func NewGradeLog() *GradeLog {
	log := &GradeLog{LogId: bson.NewObjectId()}
	return log
}

func NewGradeLogEx(snid int32, count, restCount int64, ver, logType int32, oper, remark string, platform, channel, promoter, packageId string, inviterId int32) *GradeLog {
	cl := NewGradeLog()
	cl.SnId = snid
	cl.Count = count
	cl.RestCount = restCount
	cl.Oper = oper
	cl.Remark = remark
	cl.Platform = platform
	cl.Channel = channel
	cl.Promoter = promoter
	cl.PackageTag = packageId
	cl.InviterId = inviterId
	cl.Time = time.Now()
	cl.Ver = ver
	cl.LogType = logType
	return cl
}
func InsertGradeLog(snid int32, count, restCount int64, ver, logType int32, oper, remark string, platform, channel, promoter, packageId string, inviterId int32) error {

	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	cl := NewGradeLog()
	cl.SnId = snid
	cl.Count = count
	cl.RestCount = restCount
	cl.Oper = oper
	cl.Remark = remark
	cl.Platform = platform
	cl.Channel = channel
	cl.Promoter = promoter
	cl.PackageTag = packageId
	cl.InviterId = inviterId
	cl.Time = time.Now()
	cl.Ver = ver
	cl.LogType = logType

	var ret bool
	return rpcCli.CallWithTimeout("GradeLogSvc.InsertGradeLog", cl, &ret, time.Second*30)
}

func InsertGradeLogs(logs ...*GradeLog) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	return rpcCli.CallWithTimeout("GradeLogSvc.InsertGradeLogs", logs, &ret, time.Second*30)
}

type UpdateGradeLogStatusArgs struct {
	Plt    string
	Id     bson.ObjectId
	Status int32
}

func UpdateGradeLogStatus(plt string, id bson.ObjectId, status int32) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &UpdateGradeLogStatusArgs{
		Plt:    plt,
		Id:     id,
		Status: status,
	}
	var ret bool
	return rpcCli.CallWithTimeout("GradeLogSvc.UpdateGradeLogStatus", args, &ret, time.Second*30)
}

type GradeLogLog struct {
	Logs     []*GradeLog
	PageNo   int32
	PageSize int32
	PageNum  int32
}

type GetGradeLogByPageAndSnIdArgs struct {
	Plt      string
	SnId     int32
	PageNo   int32
	PageSize int32
}

func GetGradeLogByPageAndSnId(plt string, snid, pageNo, pageSize int32) (ret *GradeLogLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetGradeLogByPageAndSnIdArgs{
		Plt:      plt,
		SnId:     snid,
		PageNo:   pageNo,
		PageSize: pageSize,
	}
	err = rpcCli.CallWithTimeout("GradeLogSvc.GetGradeLogByPageAndSnId", args, &ret, time.Second*30)
	return
}

type RemoveGradeLogArgs struct {
	Plt string
	Id  bson.ObjectId
}

func RemoveGradeLog(plt string, id bson.ObjectId) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &RemoveGradeLogArgs{
		Plt: plt,
		Id:  id,
	}
	var ret bool
	return rpcCli.CallWithTimeout("GradeLogSvc.RemoveGradeLog", args, &ret, time.Second*30)
}

type GetGradeLogBySnidAndLessTsArgs struct {
	Plt  string
	SnId int32
	Ts   time.Time
}

func GetGradeLogBySnidAndLessTs(plt string, id int32, ts time.Time) (ret []GradeLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetGradeLogBySnidAndLessTsArgs{
		Plt:  plt,
		SnId: id,
		Ts:   ts,
	}
	err = rpcCli.CallWithTimeout("GradeLogSvc.GetGradeLogBySnidAndLessTs", args, &ret, time.Second*30)
	return
}

// ///////////////////////////////////////////////////////////////////////////////
// ///////////////////////////////////////////////////////////////////////////////
// 积分兑换对账数据
var (
	BillGradeLogDBName   = "user"
	BillGradeLogCollName = "user_billgradelog"
)

type BillGradeLog struct {
	LogId      string    //账单ID
	Platform   string    //平台号
	GradeLogId string    //账变对应记录id
	SnId       int32     //用户ID
	ShopId     int32     //商品id
	ShopType   int32     //商品类型 0.虚拟 1.实物
	ShopName   string    //商品名称
	Grade      int64     //账单积分
	Coin       int64     //账单兑换金额
	TimeStamp  int64     //生效时间戳
	CreateTime time.Time //账单时间
	Ver        int32     //账单版本号
	Oper       string    //操作人
	Status     int32     //状态 0.创建  1.超时 3.取消订单 4.异常 9.已发货
}

func NewBillGradeLog(platform string, SnId, ShopId, ShopType int32, ShopName string, Grade, Coin int64, Oper, GradeLogId string) *BillGradeLog {
	tNow := time.Now()
	log := &BillGradeLog{
		LogId:      bson.NewObjectId().Hex(),
		Platform:   platform,
		SnId:       SnId,
		ShopId:     ShopId,
		ShopType:   ShopType,
		ShopName:   ShopName,
		Grade:      Grade,
		Coin:       Coin,
		Ver:        VER_PLAYER_MAX - 1,
		TimeStamp:  tNow.UnixNano(),
		CreateTime: tNow,
		Oper:       Oper,
		GradeLogId: GradeLogId,
		Status:     0,
	}
	return log
}

func InsertBillGradeLogs(logs ...*BillGradeLog) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	return rpcCli.CallWithTimeout("BillGradeLogSvc.InsertBillGradeLogs", logs, &ret, time.Second*30)
}

func InsertBillGradeLog(log *BillGradeLog) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	return rpcCli.CallWithTimeout("BillGradeLogSvc.InsertBillGradeLog", log, &ret, time.Second*30)
}

type RemoveBillGradeLogArgs struct {
	Plt   string
	LogId string
}

func RemoveBillGradeLog(plt string, logid string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &RemoveBillGradeLogArgs{
		Plt:   plt,
		LogId: logid,
	}
	var ret bool
	return rpcCli.CallWithTimeout("BillGradeLogSvc.InsertBillGradeLog", args, &ret, time.Second*30)
}

type UpdateBillGradeLogStatusArgs struct {
	Plt    string
	LogId  string
	Status int32
}

func UpdateBillGradeLogStatus(plt string, logId string, status int32) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &UpdateBillGradeLogStatusArgs{
		Plt:    plt,
		LogId:  logId,
		Status: status,
	}
	var ret bool
	return rpcCli.CallWithTimeout("BillGradeLogSvc.UpdateBillGradeLogStatus", args, &ret, time.Second*30)
}

type GetAllBillGradeLogArgs struct {
	Plt  string
	SnId int32
	Ts   int64
}

func GetAllBillGradeLog(plt string, snid int32, ts int64) (ret []BillGradeLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetAllBillGradeLogArgs{
		Plt:  plt,
		SnId: snid,
		Ts:   ts,
	}
	err = rpcCli.CallWithTimeout("BillGradeLogSvc.GetAllBillGradeLog", args, &ret, time.Second*30)
	return
}

type GetBillGradeLogByLogIdArgs struct {
	Plt   string
	LogId string
}

func GetBillGradeLogByLogId(plt string, logId string) (ret *BillGradeLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetBillGradeLogByLogIdArgs{
		Plt:   plt,
		LogId: logId,
	}
	err = rpcCli.CallWithTimeout("BillGradeLogSvc.GetBillGradeLogByLogId", args, &ret, time.Second*30)
	return
}

type GetBillGradeLogByBillNoArgs struct {
	Plt    string
	SnId   int32
	BillNo int64
}

func GetBillGradeLogByBillNo(plt string, snid int32, billNo int64) (ret *BillGradeLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetBillGradeLogByBillNoArgs{
		Plt:    plt,
		SnId:   snid,
		BillNo: billNo,
	}
	err = rpcCli.CallWithTimeout("BillGradeLogSvc.GetBillGradeLogByLogId", args, &ret, time.Second*30)
	return
}
