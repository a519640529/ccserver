package main

import (
	"context"
	"strings"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/etcd"
	"games.yol.com/win88/proto"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"github.com/idealeak/goserver/core/logger"
	"go.etcd.io/etcd/clientv3"
)

var EtcdMgrSington = &EtcdMgr{EtcdClient: &etcd.EtcdClient{}}

type EtcdMgr struct {
	*etcd.EtcdClient
}

// 初始化平台DB配置
func (this *EtcdMgr) InitPlatformDBCfg() {
	initFunc := func() int64 {
		logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_SYS_PLT_DBCFG_PREFIX)
		res, err := this.GetValueWithPrefix(etcd.ETCDKEY_SYS_PLT_DBCFG_PREFIX)
		if err == nil {
			for i := int64(0); i < res.Count; i++ {
				var config webapi_proto.PlatformDbConfig
				err = proto.Unmarshal(res.Kvs[i].Value, &config)
				if err == nil {
					s := strings.TrimPrefix(string(res.Kvs[i].Key), etcd.ETCDKEY_SYS_PLT_DBCFG_PREFIX)
					arr := strings.Split(s, "/")
					if len(arr) >= 1 {
						pltId := arr[0]
						if err == nil {
							//用户库
							mongo.MgoSessionMgrSington.UptCfgWithEtcd(pltId, "user", config.MongoDb)
							//日志库
							mongo.MgoSessionMgrSington.UptCfgWithEtcd(pltId, "log", config.MongoDbLog)
						}
					}
				} else {
					logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_SYS_PLT_DBCFG_PREFIX, err)
				}
			}
			if res.Header != nil {
				return res.Header.Revision
			}
		} else {
			logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_SYS_PLT_DBCFG_PREFIX, err)
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_SYS_PLT_DBCFG_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var config webapi_proto.PlatformDbConfig
					err := proto.Unmarshal(ev.Kv.Value, &config)
					if err == nil {
						s := strings.TrimPrefix(string(ev.Kv.Key), etcd.ETCDKEY_SYS_PLT_DBCFG_PREFIX)
						arr := strings.Split(s, "/")
						if len(arr) >= 1 {
							pltId := arr[0]
							if err == nil {
								//用户库
								mongo.MgoSessionMgrSington.UptCfgWithEtcd(pltId, "user", config.MongoDb)
								//日志库
								mongo.MgoSessionMgrSington.UptCfgWithEtcd(pltId, "log", config.MongoDbLog)
							}
						}
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_SYS_PLT_DBCFG_PREFIX, err)
					}
				}
			}
			return nil
		})
	}
	this.InitAndWatch(initFunc, watchFunc)
}

func (this *EtcdMgr) Init() {
	logger.Logger.Infof("EtcdClient开始连接url:%v;etcduser:%v;etcdpwd:%v", common.CustomConfig.GetStrings("etcdurl"), common.CustomConfig.GetString("etcduser"), common.CustomConfig.GetString("etcdpwd"))
	err := this.Open(common.CustomConfig.GetStrings("etcdurl"), common.CustomConfig.GetString("etcduser"), common.CustomConfig.GetString("etcdpwd"), time.Minute)
	if err != nil {
		logger.Logger.Tracef("EtcdMgr.Open err:%v", err)
	}
}

func (this *EtcdMgr) Shutdown() {
	this.Close()
}

func (this *EtcdMgr) Reset() {
	this.Close()
	this.Init()
	this.ReInitAndWatchAll()
}
