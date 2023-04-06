package main

import (
	"fmt"
	"github.com/globalsign/mgo"
	"time"
)

var autoPingInterval time.Duration = time.Minute
var MgoSession *mgo.Session

func newDBSession() (s *mgo.Session, err error) {
	login := ""
	if AppCfg.MC.UserName != "" {
		login = AppCfg.MC.UserName + ":" + AppCfg.MC.Password + "@"
	}
	host := "localhost"
	if AppCfg.MC.Host != "" {
		host = AppCfg.MC.Host
	}

	// http://goneat.org/pkg/labix.org/v2/mgo/#Session.Mongo
	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	url := fmt.Sprintf("mongodb://%s%s/admin", login, host)
	//fmt.Println(url)
	session, err := mgo.Dial(url)
	if err != nil {
		return
	}

	s = session
	return
}

func Ping() {
	var err error
	if MgoSession != nil {
		err = MgoSession.Ping()
		if err != nil {
			Log.Errorf("mongo.Ping err:%v", err)
			MgoSession.Refresh()
		} else {
			Log.Tracef("mongo.Ping suc")
		}
	}
}

func StartMgoPing() {
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				Log.Warnf("StartMgoPing panic: %s", err)
			}
		}()
		for {
			select {
			case <-time.After(autoPingInterval):
				Ping()
			}
		}
	}()
}

func SetAutoPing(interv time.Duration) {
	autoPingInterval = interv
}
