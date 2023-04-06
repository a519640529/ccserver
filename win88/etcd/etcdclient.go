package etcd

import (
	"context"
	"time"

	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"go.etcd.io/etcd/clientv3"
)

type (
	EtcdInitFunc    func() int64
	EtcdWatchFunc   func(context.Context, int64)
	EtcdKeyFuncPair struct {
		initFunc  EtcdInitFunc
		watchFunc EtcdWatchFunc
	}
)

type EtcdClient struct {
	cli    *clientv3.Client
	funcs  []EtcdKeyFuncPair
	closed bool
}

func (this *EtcdClient) IsClosed() bool {
	return this.closed
}

func (this *EtcdClient) Ctx() context.Context {
	if this.cli != nil {
		return this.cli.Ctx()
	}
	return context.TODO()
}

func (this *EtcdClient) Open(etcdUrl []string, userName, passWord string, dialTimeout time.Duration) error {
	var err error
	this.cli, err = clientv3.New(clientv3.Config{
		Endpoints:            etcdUrl,
		Username:             userName,
		Password:             passWord,
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    5 * time.Second,
		DialKeepAliveTimeout: 30 * time.Second,
	})

	if err != nil {
		logger.Logger.Warnf("EtcdClient.open(%v) err:%v", etcdUrl, err)
		return err
	}

	this.closed = false
	return err
}

func (this *EtcdClient) Close() error {
	logger.Logger.Warn("EtcdClient.close")
	this.closed = true
	if this.cli != nil {
		return this.cli.Close()
	}
	return nil
}

//添加键值对
func (this *EtcdClient) PutValue(key, value string) (*clientv3.PutResponse, error) {
	resp, err := this.cli.Put(context.TODO(), key, value)
	if err != nil {
		logger.Logger.Warnf("EtcdClient.PutValue(%v,%v) error:%v", key, value, err)
	}
	return resp, err
}

//查询
func (this *EtcdClient) GetValue(key string) (*clientv3.GetResponse, error) {
	resp, err := this.cli.Get(context.TODO(), key)
	if err != nil {
		logger.Logger.Warnf("EtcdClient.GetValue(%v) error:%v", key, err)
	}
	return resp, err
}

// 返回删除了几条数据
func (this *EtcdClient) DelValue(key string) (*clientv3.DeleteResponse, error) {
	res, err := this.cli.Delete(context.TODO(), key)
	if err != nil {
		logger.Logger.Warnf("EtcdClient.DelValue(%v) error:%v", key, err)
	}
	return res, err
}

//按照前缀删除
func (this *EtcdClient) DelValueWithPrefix(prefix string) (*clientv3.DeleteResponse, error) {
	res, err := this.cli.Delete(context.TODO(), prefix, clientv3.WithPrefix())
	if err != nil {
		logger.Logger.Warnf("EtcdClient.DelValueWithPrefix(%v) error:%v", prefix, err)
	}
	return res, err
}

//获取前缀
func (this *EtcdClient) GetValueWithPrefix(prefix string) (*clientv3.GetResponse, error) {
	resp, err := this.cli.Get(context.TODO(), prefix, clientv3.WithPrefix())
	if err != nil {
		logger.Logger.Warnf("EtcdClient.GetValueWIthPrefix(%v) error:%v", prefix, err)
	}
	return resp, err
}

func (this *EtcdClient) WatchWithPrefix(prefix string, revision int64) clientv3.WatchChan {
	if this.cli != nil {
		opts := []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithCreatedNotify()}
		if revision > 0 {
			opts = append(opts, clientv3.WithRev(revision))
		}
		return this.cli.Watch(clientv3.WithRequireLeader(context.Background()), prefix, opts...)
	}
	return nil
}

func (this *EtcdClient) Compact() {
	if this.closed {
		return
	}

	resp, err := this.GetValue("@@@GET_LASTEST_REVISION@@@")
	if err == nil {
		ctx, _ := context.WithCancel(this.cli.Ctx())
		start := time.Now()
		compactResponse, err := this.cli.Compact(ctx, resp.Header.Revision, clientv3.WithCompactPhysical())
		if err == nil {
			logger.Logger.Infof("EtcdClient.Compact From %v CompactResponse %v take %v", resp.Header.Revision, compactResponse.Header, time.Now().Sub(start))
		} else {
			logger.Logger.Errorf("EtcdClient.Compact From %v CompactResponse:%v take:%v err:%v", resp.Header.Revision, compactResponse, time.Now().Sub(start), err)
		}
		endpoints := this.cli.Endpoints()
		for _, endpoint := range endpoints {
			ctx1, _ := context.WithCancel(this.cli.Ctx())
			start := time.Now()
			defragmentResponse, err := this.cli.Defragment(ctx1, endpoint)
			if err == nil {
				logger.Logger.Infof("EtcdClient.Defragment %v,%v take %v", endpoint, defragmentResponse.Header, time.Now().Sub(start))
			} else {
				logger.Logger.Errorf("EtcdClient.Defragment DefragmentResponse:%v take:%v err:%v", defragmentResponse, time.Now().Sub(start), err)
			}
		}
	}
}

func (this *EtcdClient) ReInitAndWatchAll() {
	if this.closed {
		return
	}

	oldFuncs := this.funcs
	this.funcs = nil
	for i := 0; i < len(oldFuncs); i++ {
		this.InitAndWatch(oldFuncs[i].initFunc, oldFuncs[i].watchFunc)
	}
}

func (this *EtcdClient) InitAndWatch(initFunc EtcdInitFunc, watchFunc EtcdWatchFunc) {
	funcPair := EtcdKeyFuncPair{
		initFunc:  initFunc,
		watchFunc: watchFunc,
	}
	this.funcs = append(this.funcs, funcPair)
	lastRevision := initFunc()
	ctx, _ := context.WithCancel(this.cli.Ctx())
	watchFunc(ctx, lastRevision+1)
}

func (this *EtcdClient) GoWatch(ctx context.Context, revision int64, prefix string, f func(res clientv3.WatchResponse) error) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Logger.Errorf("etcd watch WithPrefix(%v) panic:%v", prefix, err)
			}
			logger.Logger.Warnf("etcd watch WithPrefix(%v) quit!!!", prefix)
		}()
		var times int64
		for !this.closed {
			times++
			logger.Logger.Warnf("etcd watch WithPrefix(%v) base revision %v start[%v]!!!", prefix, revision, times)
			rch := this.WatchWithPrefix(prefix, revision)
			for {
				skip := false
				select {
				case _, ok := <-ctx.Done():
					if !ok {
						return
					}
				case wresp, ok := <-rch:
					if !ok {
						logger.Logger.Warnf("etcd watch WithPrefix(%v) be closed", prefix)
						skip = true
						break
					}
					if wresp.Header.Revision > revision {
						revision = wresp.Header.Revision
					}
					if wresp.Canceled {
						logger.Logger.Warnf("etcd watch WithPrefix(%v) be closed, reason:%v", prefix, wresp.Err())
						skip = true
						break
					}
					if err := wresp.Err(); err != nil {
						logger.Logger.Warnf("etcd watch WithPrefix(%v) err:%v", prefix, wresp.Err())
						continue
					}
					if !model.GameParamData.UseEtcd {
						continue
					}
					if len(wresp.Events) == 0 {
						continue
					}

					logger.Logger.Tracef("@goWatch %v changed, header:%#v", prefix, wresp.Header)
					obj := core.CoreObject()
					if obj != nil {
						func(res clientv3.WatchResponse) {
							obj.SendCommand(basic.CommandWrapper(func(*basic.Object) error {
								return f(res)
							}), true)
						}(wresp)
					}
				}

				if skip {
					break
				}
			}
			time.Sleep(time.Second)
		}
	}()
}
