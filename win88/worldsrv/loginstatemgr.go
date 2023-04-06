package main

import (
	"fmt"
	"github.com/idealeak/goserver/core/logger"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

const (
	LoginState_Logining int = iota
	LoginState_Logined
	LoginState_Logouting
	LoginState_Logouted
)

var LoginStateMgrSington = &LoginStateMgr{
	statesByName:   make(map[string]*AccLoginState),
	statesByAccId:  make(map[string]*AccLoginState),
	statesByPlayer: make(map[string]*AccLoginState),
	statesBySid:    make(map[int64]*LoginState),
}

type LoginState struct {
	userName     string
	sid          int64
	state        int
	gameId       int
	gateSess     *netlib.Session
	als          *AccLoginState
	clog         *model.ClientLoginInfo
	startLoginTs int64
}

type AccLoginState struct {
	acc          *model.Account
	lss          map[int64]*LoginState
	lastLoginTs  int64 //最后登录时间
	lastLogoutTs int64 //最后登出时间
}

type LoginStateMgr struct {
	statesByName   map[string]*AccLoginState
	statesByAccId  map[string]*AccLoginState
	statesByPlayer map[string]*AccLoginState
	statesBySid    map[int64]*LoginState
}

func (this *LoginStateMgr) IsLogining(userName, platform string, tagkey int32) bool {
	userName = fmt.Sprintf("%v_%v_%v", userName, platform, tagkey)
	if v, exist := this.statesByName[userName]; exist {
		if len(v.lss) == 0 {
			return false
		}
		for _, ls := range v.lss {
			if ls.state == LoginState_Logining {
				return true
			}
		}
	}
	return false
}

func (this *LoginStateMgr) IsLogined(userName, platform string, tagkey int32) bool {
	userName = fmt.Sprintf("%v_%v_%v", userName, platform, tagkey)
	if v, exist := this.statesByName[userName]; exist {
		if len(v.lss) == 0 {
			return false
		}
		for _, ls := range v.lss {
			if ls.state == LoginState_Logined {
				return true
			}
		}
	}
	return false
}

func (this *LoginStateMgr) IsLoginedOfSid(sid int64) bool {
	if v, exist := this.statesBySid[sid]; exist {
		return v.state == LoginState_Logined
	}
	return false
}

func (this *LoginStateMgr) GetLoginStateByName(userName string) *AccLoginState {
	if v, exist := this.statesByName[userName]; exist {
		return v
	}
	return nil
}

func (this *LoginStateMgr) GetLoginStateByAccId(accId string) *AccLoginState {
	if v, exist := this.statesByAccId[accId]; exist {
		return v
	}
	return nil
}

func (this *LoginStateMgr) GetLoginStateByTelAndPlatform(tel, platform string) *AccLoginState {
	for _, als := range this.statesByPlayer {
		if als != nil && als.acc != nil {
			if als.acc.Tel == tel && als.acc.Platform == platform {
				return als
			}
		}
	}

	return nil
}

func (this *LoginStateMgr) GetLoginStateOfSid(sid int64) *LoginState {
	if v, exist := this.statesBySid[sid]; exist {
		return v
	}
	return nil
}

func (this *LoginStateMgr) SetLoginStateOfSid(sid int64, state int) {
	if v, exist := this.statesBySid[sid]; exist {
		v.state = state
	}
}

func (this *LoginStateMgr) StartLogin(userName, platform string, sid int64, s *netlib.Session, clog *model.ClientLoginInfo, tagkey int32) bool {
	//注意此处的坑，这个地方要增加平台id
	userName = fmt.Sprintf("%v_%v_%v", userName, platform, tagkey)
	ts := time.Now().Unix()
	if als, exist := this.statesByName[userName]; !exist {
		als = &AccLoginState{
			lss:         make(map[int64]*LoginState),
			lastLoginTs: ts,
		}
		ls := &LoginState{
			userName:     userName,
			sid:          sid,
			gateSess:     s,
			state:        LoginState_Logining,
			als:          als,
			startLoginTs: ts,
			clog:         clog,
		}
		this.statesByName[userName] = als
		als.lss[sid] = ls
		this.statesBySid[sid] = ls
		return true
	} else {
		ls := &LoginState{
			userName:     userName,
			sid:          sid,
			gateSess:     s,
			state:        LoginState_Logining,
			als:          als,
			startLoginTs: ts,
			clog:         clog,
		}
		als.lss[sid] = ls
		this.statesBySid[sid] = ls
		if als.acc != nil {
			return false
		}
		return true
	}
}

func (this *LoginStateMgr) Logined(userName, platform string, sid int64, acc *model.Account, tagkey int32) (oldState map[int64]*LoginState) {
	userName = fmt.Sprintf("%v_%v_%v", userName, platform, tagkey)
	if v, exist := this.statesByName[userName]; exist {
		v.acc = acc
		v.lastLoginTs = time.Now().Unix()
		if acc != nil {
			this.statesByAccId[acc.AccountId.Hex()] = v
			if acc.Platform != common.Platform_Rob {
				this.statesByPlayer[acc.AccountId.Hex()] = v
			}
		}
		if len(v.lss) > 1 {
			oldState = make(map[int64]*LoginState)
			for k, v := range v.lss {
				if k != sid {
					oldState[k] = v
				}
			}
		}
	}

	if v, exist := this.statesBySid[sid]; exist {
		v.state = LoginState_Logined
	}

	return
}

func (this *LoginStateMgr) Logout(s *LoginState) {
	if s != nil {
		delete(this.statesBySid, s.sid)
		//缓存下,避免对DB造成冲击
		s.state = LoginState_Logouted
		s.als.lastLogoutTs = time.Now().Unix()
		delete(s.als.lss, s.sid)
	}
}

func (this *LoginStateMgr) LogoutBySid(sid int64) {
	if s, exist := this.statesBySid[sid]; exist {
		this.Logout(s)
	}
}

func (this *LoginStateMgr) LogoutByAccount(accId string) {
	if als, exist := this.statesByAccId[accId]; exist {
		if als != nil {
			for _, s := range als.lss {
				this.Logout(s)
			}
		}
	}
}

func (this *LoginStateMgr) LogoutAllBySession(session *netlib.Session) {
	for sid, s := range this.statesBySid {
		if s.gateSess == session {
			this.Logout(s)
			p := PlayerMgrSington.GetPlayer(sid)
			if p != nil {
				p.DropLine()
			}
		}
	}
}

// 特殊处理，负责从statesByName删除，为了处理账号的删除问题
func (this *LoginStateMgr) DelAccountByAccid(accid string) {
	for name, s := range this.statesByPlayer {
		if s != nil && s.acc != nil && s.acc.AccountId.Hex() == accid {
			this.DeleteAccount(name, s)
		}
	}

}

// 删除平台cache
func (this *LoginStateMgr) DelCacheByPlatform(platform string) {
	for name, s := range this.statesByPlayer {
		if s != nil && s.acc != nil && s.acc.Platform == platform {
			this.DeleteAccount(name, s)
		}
	}
}

func (this *LoginStateMgr) DeleteAccount(name string, s *AccLoginState) {
	if s != nil && s.acc != nil {
		//注意此处的坑，这个地方要增加平台id,为了和上面的匹配相同
		userName := fmt.Sprintf("%v_%v_%v", s.acc.UserName, s.acc.Platform, s.acc.TagKey)
		delete(this.statesByName, userName)
		userName = fmt.Sprintf("%v_%v_%v", s.acc.Tel, s.acc.Platform, s.acc.TagKey)
		delete(this.statesByName, userName)
		acc := s.acc
		if acc != nil {
			delete(this.statesByAccId, acc.AccountId.Hex())
		}
		delete(this.statesByPlayer, name)
	}
}

// //////////////////////////////////////////////////////////////////
// / Module Implement [beg]
// //////////////////////////////////////////////////////////////////
func (this *LoginStateMgr) ModuleName() string {
	return "LoginStateMgr"
}

func (this *LoginStateMgr) Init() {
	if model.GameParamData.PreLoadRobotCount > 0 {
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			tsBeg := time.Now()
			accounts := model.GetRobotAccounts(model.GameParamData.PreLoadRobotCount)
			tsEnd := time.Now()
			logger.Logger.Tracef("GetRobotAccounts take:%v total:%v", tsEnd.Sub(tsBeg), len(accounts))
			return accounts
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if accounts, ok := data.([]model.Account); ok {
				if accounts != nil {
					ts := time.Now().Add(time.Hour).Unix()
					for i := 0; i < len(accounts); i++ {
						userName := fmt.Sprintf("%v_%v_%v", accounts[i].UserName, accounts[i].Platform, accounts[i].TagKey)
						if _, exist := this.statesByName[userName]; !exist {
							als := &AccLoginState{
								acc:          &accounts[i],
								lss:          make(map[int64]*LoginState),
								lastLoginTs:  ts,
								lastLogoutTs: ts,
							}
							this.statesByName[userName] = als
							this.statesByAccId[accounts[i].AccountId.Hex()] = als
						}
					}
				}
			}
		}), "GetAllRobotAccounts").Start()
	}

	PlayerMgrSington.LoadRobots()
}

func (this *LoginStateMgr) Update() {
	curTs := time.Now().Unix()
	for name, s := range this.statesByPlayer {
		if len(s.lss) == 0 && curTs-s.lastLogoutTs > int64(model.GameParamData.LoginStateCacheSec) {
			if s != nil && s.acc != nil {
				this.DeleteAccount(name, s)
			}
		}
	}
}

func (this *LoginStateMgr) Shutdown() {
	for _, s := range this.statesByName {
		if s.acc != nil && s.acc.Platform != common.Platform_Rob {
			model.LogoutAccount(s.acc)
		}
	}
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(LoginStateMgrSington, time.Minute, 0)
}
