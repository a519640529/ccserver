package base

import (
	"context"
	"games.yol.com/win88/common"
	"games.yol.com/win88/etcd"
	"games.yol.com/win88/proto"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"github.com/idealeak/goserver/core/logger"
	"go.etcd.io/etcd/clientv3"
	"strconv"
	"time"
)

var EtcdMgrSington = &EtcdMgr{EtcdClient: &etcd.EtcdClient{}}

type EtcdMgr struct {
	*etcd.EtcdClient
}

//初始化平台数据
func (this *EtcdMgr) InitPlatform() {
	initFunc := func() int64 {
		logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_PLATFORM_PREFIX)
		res, err := this.GetValueWithPrefix(etcd.ETCDKEY_PLATFORM_PREFIX)
		if err == nil {
			for i := int64(0); i < res.Count; i++ {
				var value webapi_proto.Platform
				err = proto.Unmarshal(res.Kvs[i].Value, &value)
				if err == nil {
					PlatformMgrSington.UpsertPlatform(strconv.Itoa(int(value.Id)))
				} else {
					logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_PLATFORM_PREFIX, err)
				}
			}
			if res.Header != nil {
				return res.Header.Revision
			}
		} else {
			logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_PLATFORM_PREFIX, err)
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_PLATFORM_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
					var value webapi_proto.Platform
					err := proto.Unmarshal(ev.Kv.Value, &value)
					if err == nil {
						PlatformMgrSington.DelPlatform(strconv.Itoa(int(value.Id)))
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_PLATFORM_PREFIX, err)
					}
				case clientv3.EventTypePut:
					var value webapi_proto.Platform
					err := proto.Unmarshal(ev.Kv.Value, &value)
					if err == nil {
						PlatformMgrSington.UpsertPlatform(strconv.Itoa(int(value.Id)))
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_PLATFORM_PREFIX, err)
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