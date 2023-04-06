package svc

import (
	"errors"
	"math"
	"net/rpc"
	"time"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

func GamePlayerListLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.GamePlayerListLogDBName)
	if s != nil {
		c_gameplayerlistlog, first := s.DB().C(model.GamePlayerListLogCollName)
		if first {
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"channel"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"promoter"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"packagetag"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"gameid"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"clubid"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"sceneid"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"gamefreeid"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"gamedetailedlogid"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"ts"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
			c_gameplayerlistlog.EnsureIndex(mgo.Index{Key: []string{"name"}, Background: true, Sparse: true})
		}
		return c_gameplayerlistlog
	}
	return nil
}

type GamePlayerListSvc struct {
}

func (svc *GamePlayerListSvc) InsertGamePlayerListLog(log *model.GamePlayerListLog, ret *bool) (err error) {
	clog := GamePlayerListLogsCollection(log.Platform)
	if clog == nil {
		return
	}
	err = clog.Insert(log)
	if err != nil {
		logger.Logger.Warn("GamePlayerListSvc.InsertGamePlayerListLog error:", err)
		return
	}
	*ret = true
	return
}

func (svc *GamePlayerListSvc) GetPlayerCount(args *model.GamePlayerListArg, ret *model.GamePlayerListRet) error {
	gdlc := GamePlayerListLogsCollection(args.Platform)
	if gdlc == nil {
		logger.Logger.Error("svc.GetPlayerCount gdlc == nil")
		return nil
	}
	var sql []bson.M
	sql = append(sql, bson.M{"snid": args.SnId, "ts": bson.M{"$gte": args.StartTime, "$lte": args.EndTime}})
	sql = append(sql, bson.M{"clubid": args.ClubId})
	sql = append(sql, bson.M{"platform": args.Platform})
	gameTotal, err := gdlc.Find(bson.M{"$and": sql}).Count()
	if err != nil {
		logger.Logger.Error("svc.GetPlayerCount is error", err)
		return err
	}
	if gameTotal == 0 {
		return errors.New("gameTotal==0")
	}
	prt := new(model.GameTotalRecord)
	prt.GameTotal = int32(gameTotal)
	var tc []model.TotalCoin
	err = gdlc.Pipe([]bson.M{
		{"$match": bson.M{
			"snid":   args.SnId,
			"clubid": args.ClubId,
			"ts":     bson.M{"$gte": args.StartTime, "$lte": args.EndTime},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"snid":   "$snid",
				"clubid": "$clubid",
			},
			"totalin":      bson.M{"$sum": "$totalin"},
			"totalout":     bson.M{"$sum": "$totalout"},
			"taxcoin":      bson.M{"$sum": "$taxcoin"},
			"clubpumpcoin": bson.M{"$sum": "$clubpumpcoin"},
		}}}).AllowDiskUse().All(&tc)
	if err != nil {
		logger.Logger.Error("svc.GetPlayerCount AllowDiskUse is error", err)
		return err
	}
	if len(tc) > 0 {
		d := tc[0]
		prt.GameTotalCoin = int32(d.TotalIn + d.TotalOut)
		prt.GameWinTotal = int32(d.TotalOut - d.TotalIn - d.TaxCoin - d.ClubPumpCoin)
	}
	ret.Gtr = prt
	return nil
}
func (svc *GamePlayerListSvc) GetPlayerListLog(args *model.GamePlayerListArg, ret *model.GamePlayerListRet) error {
	gdlc := GamePlayerListLogsCollection(args.Platform)
	if gdlc == nil {
		return nil
	}
	var sql []bson.M
	if args.SnId != 0 {
		sql = append(sql, bson.M{"snid": args.SnId})
	}
	if args.Platform != "" {
		sql = append(sql, bson.M{"platform": args.Platform})
	}
	if args.ClubId != 0 {
		sql = append(sql, bson.M{"clubid": args.ClubId})
	}
	if args.StartTime != 0 {
		sql = append(sql, bson.M{"ts": bson.M{"$gte": args.StartTime, "$lte": args.EndTime}})
	}
	total, err := gdlc.Find(bson.M{"$and": sql}).Count()
	if err != nil {
		logger.Logger.Warn("select log_gamedetailed error: ", err)
		return err
	}
	gdt := model.GamePlayerListType{}
	if total == 0 {
		gdt.PageNo = args.PageNo
		gdt.PageSize = args.PageSize
		return nil
	}
	gdt.PageSum = int(math.Ceil(float64(total) / float64(args.PageSize)))
	if args.PageNo > gdt.PageSum {
		args.PageNo = gdt.PageSum
	}
	if args.PageNo <= 0 {
		args.PageNo = 1
	}
	limitNum := (args.PageNo - 1) * args.PageSize

	var data []*model.NeedGameRecord

	err = gdlc.Find(bson.M{"$and": sql}).Sort("-ts").Limit(args.PageSize).Skip(limitNum).All(&data)
	if err != nil {
		logger.Logger.Warn("select log_gameplayerlistlog error: ", err)
		return err
	}
	gdt.PageNo = args.PageNo
	gdt.PageSize = args.PageSize
	gdt.Data = data
	ret.Gplt = gdt
	return nil
}
func (svc *GamePlayerListSvc) GetPlayerListByHall(args *model.GamePlayerListArg, ret *model.GamePlayerListRet) error {
	gdlc := GamePlayerListLogsCollection(args.Platform)
	if gdlc == nil {
		return nil
	}
	var sql []bson.M
	if args.SnId != 0 {
		sql = append(sql, bson.M{"snid": args.SnId})
	}
	if args.Platform != "" {
		sql = append(sql, bson.M{"platform": args.Platform})
	}

	sql = append(sql, bson.M{"roomtype": args.RoomType})

	if args.StartTime != 0 {
		sql = append(sql, bson.M{"ts": bson.M{"$gte": args.StartTime, "$lte": args.EndTime}})
	}
	if args.GameClass <= 6 {
		sql = append(sql, bson.M{"gameclass": args.GameClass})
	}
	total, err := gdlc.Find(bson.M{"$and": sql}).Count()
	if err != nil {
		logger.Logger.Warn("select log_gameplayerlistlog error: ", err)
		return err
	}
	gdt := model.GamePlayerListType{}
	if total == 0 {
		gdt.PageNo = args.PageNo
		gdt.PageSize = args.PageSize
		return nil
	}
	gdt.PageSum = int(math.Ceil(float64(total) / float64(args.PageSize)))
	if args.PageNo > gdt.PageSum {
		args.PageNo = gdt.PageSum
	}
	if args.PageNo <= 0 {
		args.PageNo = 1
	}
	limitNum := (args.PageNo - 1) * args.PageSize

	var data []*model.NeedGameRecord

	err = gdlc.Find(bson.M{"$and": sql}).Sort("-ts").Limit(args.PageSize).Skip(limitNum).All(&data)
	if err != nil {
		logger.Logger.Error("svc.GetPlayerListByHall error: ", err)
		return err
	}
	gdt.PageNo = args.PageNo
	gdt.PageSize = args.PageSize
	gdt.Data = data
	ret.Gplt = gdt
	return nil
}

func (svc *GamePlayerListSvc) GetPlayerListByHallEx(args *model.GamePlayerListArg, ret *model.GamePlayerListRet) error {
	logger.Logger.Tracef("GamePlayerListSvc.GetPlayerListByHallEx=====> args:%v", args)
	gdlc := GamePlayerListLogsCollection(args.Platform)
	if gdlc == nil {
		return nil
	}
	var sql []bson.M
	if args.SnId != 0 {
		sql = append(sql, bson.M{"snid": args.SnId})
	}
	if args.Platform != "" {
		sql = append(sql, bson.M{"platform": args.Platform})
	}

	sql = append(sql, bson.M{"roomtype": args.RoomType})
	sql = append(sql, bson.M{"gameid": args.GameId})

	if args.StartTime != 0 {
		sql = append(sql, bson.M{"ts": bson.M{"$gte": args.StartTime, "$lte": args.EndTime}})
	}
	if args.GameClass <= 6 {
		sql = append(sql, bson.M{"gameclass": args.GameClass})
	}
	total, err := gdlc.Find(bson.M{"$and": sql}).Count()
	if err != nil {
		logger.Logger.Warn("svc.GetPlayerListByHallEx Count error: ", err)
		return err
	}
	gdt := model.GamePlayerListType{}
	if total == 0 {
		gdt.PageNo = args.PageNo
		gdt.PageSize = args.PageSize
		return nil
	}
	gdt.PageSum = int(math.Ceil(float64(total) / float64(args.PageSize)))
	if args.PageNo > gdt.PageSum {
		args.PageNo = gdt.PageSum
	}
	if args.PageNo <= 0 {
		args.PageNo = 1
	}
	limitNum := (args.PageNo - 1) * args.PageSize

	var data []*model.NeedGameRecord

	err = gdlc.Find(bson.M{"$and": sql}).Sort("-ts").Limit(args.PageSize).Skip(limitNum).All(&data)
	if err != nil {
		logger.Logger.Warn("svc.GetPlayerListByHallEx error: ", err)
		return err
	}
	gdt.PageNo = args.PageNo
	gdt.PageSize = args.PageSize
	gdt.Data = data
	ret.Gplt = gdt
	logger.Logger.Tracef("GamePlayerListSvc.GetPlayerListByHallEx=====> ret:%v", ret)
	return nil
}

func (svc *GamePlayerListSvc) GetPlayerListByHallExAPI(args *model.GamePlayerListAPIArg, ret *model.GamePlayerListRet) error {
	logger.Logger.Tracef("GamePlayerListSvc.GetPlayerListByHallExAPI=====> args:%v", args)
	gdlc := GamePlayerListLogsCollection(args.Platform)
	if gdlc == nil {
		return nil
	}
	var sql []bson.M
	if args.SnId != 0 {
		sql = append(sql, bson.M{"snid": args.SnId})
	}
	if args.Platform != "" {
		sql = append(sql, bson.M{"platform": args.Platform})
	}

	if args.StartTime != 0 {
		sql = append(sql, bson.M{"ts": bson.M{"$gte": args.StartTime, "$lte": args.EndTime}}) // >= StartTime <= EndTime
	}

	total, err := gdlc.Find(bson.M{"$and": sql}).Count()
	if err != nil {
		logger.Logger.Warn("svc.GetPlayerListByHallExAPI Count error: ", err)
		return err
	}
	gdt := model.GamePlayerListType{}
	if total == 0 {
		gdt.PageNo = args.PageNo
		gdt.PageSize = args.PageSize
		return nil
	}
	gdt.PageSum = int(math.Ceil(float64(total) / float64(args.PageSize)))
	if args.PageNo > gdt.PageSum {
		args.PageNo = gdt.PageSum
	}
	if args.PageNo <= 0 {
		args.PageNo = 1
	}
	limitNum := (args.PageNo - 1) * args.PageSize

	var data []*model.NeedGameRecord

	err = gdlc.Find(bson.M{"$and": sql}).Sort("-ts").Limit(args.PageSize).Skip(limitNum).All(&data)
	if err != nil {
		logger.Logger.Warn("svc.GetPlayerListByHallEx error: ", err)
		return err
	}
	for _, v := range data {
		//v.Username = "aaa"
		//continue
		a, err := _AccountSvc.getAccountBySnId(args.Platform, v.SnId)
		if err != nil {
			logger.Logger.Warnf("model.getAccountBySnId(%v) failed:%v", args.SnId, err)
			return err
		}
		v.Username = a.UserName
	}
	gdt.PageNo = args.PageNo
	gdt.PageSize = args.PageSize
	gdt.Data = data
	ret.Gplt = gdt
	logger.Logger.Tracef("GamePlayerListSvc.GetPlayerListByHallEx=====> ret:%v", ret)
	return nil
}

func (svc *GamePlayerListSvc) GetPlayerExistListByTs(args *model.GamePlayerExistListArg, ret *[]int64) error {
	logger.Logger.Tracef("GamePlayerListSvc.GetPlayerListByHallExAPI=====> args:%v", args)
	gdlc := GamePlayerListLogsCollection(args.Platform)
	if gdlc == nil {
		return nil
	}
	t := time.Now()
	for i := 0; i != args.DayNum; i++ {
		var sql []bson.M
		if args.SnId != 0 {
			sql = append(sql, bson.M{"snid": args.SnId})
		}
		if args.Platform != "" {
			sql = append(sql, bson.M{"platform": args.Platform})
		}
		startTime := time.Date(t.Year(), t.Month(), t.Day()-i, 0, 0, 0, 0, t.Location()).Unix() // 今日凌晨
		endTime := time.Date(t.Year(), t.Month(), t.Day()-i+1, 0, 0, 0, 0, t.Location()).Unix() // 明日凌晨

		sql = append(sql, bson.M{"ts": bson.M{"$gte": startTime, "$lt": endTime}}) // >= StartTime < EndTime

		total, err := gdlc.Find(bson.M{"$and": sql}).Count()
		if err != nil {
			logger.Logger.Warn("svc.GetPlayerListByHallExAPI Count error: ", err)
			return err
		}
		if total == 0 {
			continue
		}
		// 当日不为空
		*ret = append(*ret, startTime)
	}

	logger.Logger.Tracef("GamePlayerListSvc.GetPlayerListByHallEx=====> ret:%v", ret)
	return nil
}

func init() {
	rpc.Register(new(GamePlayerListSvc))
}
