package mongo

import (
	"encoding/json"
	"fmt"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"io/ioutil"
	"sync"
	"time"

	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core/logger"
)

const (
	G_P = "__$G_P$__"
)

var MgoSessionMgrSington = &MgoSessionMgr{
	pltCfgs: new(sync.Map),
	pltMgos: make(map[string]*PlatformSession),
}

type MgoCfg struct {
	HostName string
	HostPort int32
	Database string
	Username string
	Password string
	Options  string
	CfgVer   int32
	WithEtcd bool
}

type Config struct {
	Global    map[string]*MgoCfg
	Platforms map[string]map[string]*MgoCfg
}

type Collection struct {
	*mgo.Collection
}

type Database struct {
	sync.RWMutex
	*mgo.Database
	mc map[string]*Collection
}

func (db *Database) C(name string) (*Collection, bool) {
	db.RLock()
	c, ok := db.mc[name]
	if ok {
		db.RUnlock()
		return c, false
	}
	db.RUnlock()

	db.Lock()
	c, ok = db.mc[name]
	if ok {
		db.Unlock()
		return c, false
	}
	cc := &Collection{Collection: db.Database.C(name)}
	db.mc[name] = cc
	db.Unlock()

	return cc, true
}

type Session struct {
	sync.RWMutex
	*mgo.Session
	db  *Database
	cfg *MgoCfg
}

func (s *Session) DB() *Database {
	s.RLock()
	if s.db != nil {
		s.RUnlock()
		return s.db
	}
	s.RUnlock()
	s.Lock()
	s.db = &Database{Database: s.Session.DB(s.cfg.Database), mc: make(map[string]*Collection)}
	s.Unlock()
	return s.db
}

type PlatformSession struct {
	sync.RWMutex
	sesses map[string]*Session
}

func (ps *PlatformSession) GetSession(name string) (*Session, bool) {
	ps.RLock()
	s, ok := ps.sesses[name]
	ps.RUnlock()
	return s, ok
}

func (ps *PlatformSession) SetSession(name string, s *Session) {
	ps.Lock()
	ps.sesses[name] = s
	ps.Unlock()
}

func (ps *PlatformSession) DiscardSession(name string, s *Session) {
	ps.Lock()
	old, ok := ps.sesses[name]
	delete(ps.sesses, name)
	ps.Unlock()
	if ok && old != nil && old == s {
		//1分钟后释放，防止有pending的任务在执行
		time.AfterFunc(time.Minute, func() {
			old.Close()
		})
	}
}

type MgoSessionMgr struct {
	sync.RWMutex
	pltCfgs *sync.Map
	pltMgos map[string]*PlatformSession
}

func newMgoSession(user, password, host string, port int32, options string) (s *mgo.Session, err error) {
	login := ""
	if user != "" {
		login = user + ":" + password + "@"
	}
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 27017
	}
	if options != "" {
		options = "?" + options
	}
	// http://goneat.org/pkg/labix.org/v2/mgo/#Session.Mongo
	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	url := fmt.Sprintf("mongodb://%s%s:%d/admin%s", login, host, port, options)
	//fmt.Println(url)
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	session.SetSafe(&mgo.Safe{})
	return session, nil
}

func (msm *MgoSessionMgr) LoadConfig(cfgFile string) error {
	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		logger.Logger.Errorf("MgoSessionMgr.LoadConfig ReadFile err:%v", err)
		return err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		logger.Logger.Errorf("MgoSessionMgr.LoadConfig Unmarshal err:%v", err)
		return err
	}

	//全局配置
	for key, cfg := range cfg.Global {
		if _, ok := msm.UptCfg(G_P, key, cfg); ok {
			if ps, olds, ok := msm.HasPltMgoSession(G_P, key); ok {
				if olds.cfg.CfgVer < cfg.CfgVer {
					ps.DiscardSession(key, olds)
				}
			}
		}
	}

	//平台配置
	for plt, cfgs := range cfg.Platforms {
		for key, cfg := range cfgs {
			if _, ok := msm.UptCfg(plt, key, cfg); ok {
				if ps, olds, ok := msm.HasPltMgoSession(plt, key); ok {
					if olds.cfg.CfgVer < cfg.CfgVer {
						ps.DiscardSession(key, olds)
					}
				}
			}
		}
	}
	return nil
}

func (msm *MgoSessionMgr) GetPlt(plt string) (*PlatformSession, bool) {
	msm.RLock()
	ps, ok := msm.pltMgos[plt]
	msm.RUnlock()
	return ps, ok
}

func (msm *MgoSessionMgr) HasPltMgoSession(plt, key string) (*PlatformSession, *Session, bool) {
	ps, ok := msm.GetPlt(plt)
	if !ok {
		return nil, nil, false
	}

	s, ok := ps.GetSession(key)
	return ps, s, ok
}

func (msm *MgoSessionMgr) GetPltMgoSession(plt, key string) *Session {
	ps, ok := msm.GetPlt(plt)
	if !ok {
		msm.Lock()
		ps, ok = msm.pltMgos[plt]
		if !ok {
			ps = &PlatformSession{
				sesses: make(map[string]*Session),
			}
			msm.pltMgos[plt] = ps
		}
		msm.Unlock()
	}

	if ps == nil {
		return nil
	}

	if s, ok := ps.GetSession(key); ok {
		return s
	}

	ps.Lock()
	defer ps.Unlock()
	s, ok := ps.sesses[key]
	if ok {
		return s
	}

	if c, ok := msm.GetCfg(plt, key); ok {
		s, err := newMgoSession(c.Username, c.Password, c.HostName, c.HostPort, c.Options)
		if s == nil || err != nil {
			logger.Logger.Error("GetPltMgoSession(%s,%s) err:%v", plt, key, err)
			return nil
		}
		ss := &Session{Session: s, cfg: c}
		ps.sesses[key] = ss
		return ss
	}

	return nil
}

func (msm *MgoSessionMgr) GetGlobal(key string) *Session {
	return msm.GetPltMgoSession(G_P, key)
}

func (msm *MgoSessionMgr) GetCfg(plt, key string) (*MgoCfg, bool) {
	if val, ok := msm.pltCfgs.Load(plt); ok {
		if cfgs, ok := val.(*sync.Map); ok {
			if cfg, ok := cfgs.Load(key); ok {
				if c, ok := cfg.(*MgoCfg); ok {
					return c, true
				}
			}
		}
	}
	return nil, false
}

func (msm *MgoSessionMgr) UptCfg(plt, key string, cfg *MgoCfg) (*MgoCfg, bool) {
	if val, ok := msm.pltCfgs.Load(plt); ok {
		cfgs, _ := val.(*sync.Map)
		if old, ok := cfgs.Load(key); ok {
			cfgs.Store(key, cfg)
			return old.(*MgoCfg), true
		} else {
			cfgs.Store(key, cfg)
			return nil, false
		}
	} else {
		cfgs := new(sync.Map)
		msm.pltCfgs.Store(plt, cfgs)
		cfgs.Store(key, cfg)
		return nil, false
	}
	return nil, false
}

func (msm *MgoSessionMgr) UptCfgWithEtcd(plt, key string, cfg *webapi_proto.MongoDbSetting) {
	//文件配置优先级高于etcd
	if oldCfg, exist := msm.GetCfg(plt, key); exist && !oldCfg.WithEtcd {
		return
	}

	newcfg := &MgoCfg{
		HostName: cfg.HostName,
		HostPort: cfg.HostPort,
		Database: cfg.Database,
		Username: cfg.Username,
		Password: cfg.Password,
		Options:  cfg.Options,
		CfgVer:   cfg.CfgVer,
		WithEtcd: true,
	}
	if _, ok := msm.UptCfg(plt, key, newcfg); ok {
		if ps, olds, ok := msm.HasPltMgoSession(plt, key); ok {
			if olds.cfg.CfgVer < cfg.CfgVer {
				ps.DiscardSession(key, olds)
			}
		}
	}

	return
}

func (msm *MgoSessionMgr) Close() {
	msm.RLock()
	defer msm.RUnlock()
	for _, plt := range msm.pltMgos {
		plt.RLock()
		for _, s := range plt.sesses {
			s.Close()
		}
		plt.RUnlock()
	}
}
