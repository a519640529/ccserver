package main

import (
	"errors"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"time"
)

var c_filedownloadfaildetail *mgo.Collection

var ChanFileDownFailed chan *FileDownLoadFailDetailData

type FileDownLoadFailDetailData struct {
	Id         bson.ObjectId `bson:"_id"`
	FileName   string
	Url        string
	Ip         string
	Package    string
	Md5        string
	StatusCode string
	ErrorCode  string
	ErrorMsg   string
	Size       int64
	Time       time.Time
}

func FileDetailCollection() *mgo.Collection {
	if c_filedownloadfaildetail == nil {
		c_filedownloadfaildetail = MgoSession.DB(AppCfg.MC.Database).C("log_filedownfail")
		if c_filedownloadfaildetail != nil {
			c_filedownloadfaildetail.EnsureIndex(mgo.Index{Key: []string{"filename"}, Background: true, Sparse: true})
			c_filedownloadfaildetail.EnsureIndex(mgo.Index{Key: []string{"url"}, Background: true, Sparse: true})
			c_filedownloadfaildetail.EnsureIndex(mgo.Index{Key: []string{"package"}, Background: true, Sparse: true})
		}
	}
	return c_filedownloadfaildetail
}

func InsertFileDownloadDetail(logs ...*FileDownLoadFailDetailData) (err error) {
	clog := FileDetailCollection()
	if clog == nil {
		return
	}
	switch len(logs) {
	case 0:
		return errors.New("no data")
	case 1:
		err = clog.Insert(logs[0])
	default:
		docs := make([]interface{}, 0, len(logs))
		for _, log := range logs {
			docs = append(docs, log)
		}
		err = clog.Insert(docs...)
	}
	if err != nil {
		Log.Warn("InsertCoinLogs error:", err)
		return
	}
	return
}

func SelectFileDownFailedDetail(filename string, pageSize, pageNo int) (datas []*FileDownLoadFailDetailData, cnt int, err error) {
	clog := FileDetailCollection()
	if clog == nil {
		return nil, 0, nil
	}
	start := (pageNo - 1) * pageSize
	q := clog.Find(bson.M{"filename": filename}).Sort("time:-1")
	if q != nil {
		cnt, err = q.Count()
		if err != nil {
			return
		}
		err = q.Skip(start).Limit(pageSize).All(&datas)
	}
	return
}

func SelectTopNFileDownFailed(pageSize, pageNo int) (datas []*TopNFileMsg, cnt int) {
	clog := FileDetailCollection()
	if clog == nil {
		return nil, 0
	}

	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$filename", "total": bson.M{"$sum": 1}, "size": bson.M{"$avg": "$size"}}},
		{"$sort": bson.M{"total": -1}},
	}
	pipe := clog.Pipe(pipeline)
	if pipe != nil {
		pipe.AllowDiskUse()
		err := pipe.All(&datas)
		if err != nil {
			Log.Warnf("SelectTopNFileDownFailed err:%v", err)
			return nil, 0
		}

		start := (pageNo - 1) * pageSize

		cnt = len(datas)
		if start < 0 {
			start = 0
		}
		if start > cnt {
			start = cnt
		}
		end := start + pageSize
		if end > cnt {
			end = cnt
		}
		datas = datas[start:end]
	}
	//if start != 0 {
	//	pipeline = append(pipeline, bson.M{"$skip": start})
	//}
	//if pageSize != 0 {
	//	pipeline = append(pipeline, bson.M{"$limit": pageSize})
	//}

	//pipe := clog.Pipe(pipeline)
	//if pipe != nil {
	//	pipe.AllowDiskUse()
	//	err := pipe.All(&datas)
	//	if err != nil {
	//		Log.Warnf("SelectTopNFileDownFailed err:%v", err)
	//		return nil
	//	}
	//}
	return
}

func AsyncBatchInsertFileDownloadDetail(logs []*FileDownLoadFailDetailData) {
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				Log.Warnf("InsertFileDownloadDetail panic: %s", err)
			}
		}()
		err := InsertFileDownloadDetail(logs...)
		if err != nil {
			Log.Warn("InsertCoinLogs error:", err)
		}
	}()
}

func WriteFileDownLog(log *FileDownLoadFailDetailData) {
	select {
	case ChanFileDownFailed <- log:
	default:
		Log.Warn("WriteFileDownLog channel full, then drop")
	}
}

func StartPublishLog() {
	ChanFileDownFailed = make(chan *FileDownLoadFailDetailData, AppCfg.MC.MaxDone*10)
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				Log.Warnf("StartPublishLog panic: %s", err)
			}
		}()
		var logs []*FileDownLoadFailDetailData
		for {
			select {
			case log, ok := <-ChanFileDownFailed:
				if !ok {
					return
				}
				logs = append(logs, log)
				if len(logs) >= AppCfg.MC.MaxDone {
					wlogs := logs
					logs = nil
					AsyncBatchInsertFileDownloadDetail(wlogs)
				}
			case <-time.After(time.Millisecond * time.Duration(AppCfg.MC.MaxInterval)):
				if len(logs) != 0 {
					wlogs := logs
					logs = nil
					AsyncBatchInsertFileDownloadDetail(wlogs)
				}
			}
		}
	}()
}
