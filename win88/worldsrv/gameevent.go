package main

const (
	GAMEEVENT_NIL            int32 = iota
	GAMEEVENT_LOGIN                //登录事件
	GAMEEVENT_LOGOUT               //退出事件
	GAMEEVENT_COMPLETE_SHARE       //分享事件
	GAMEEVENT_COMPLETE_GAME        //完成牌局事件
	GAMEEVENT_MAX
)

var GameEventHandlerMgrSingleton = &GameEventHandlerMgr{
	ehs: make(map[int32][]GameEventHandler, GAMEEVENT_MAX),
}

type GameEvent struct {
	eventType int32
	eventSrc  interface{}
	params    [4]int32
}

type GameEventHandler interface {
	Handler(*Player, *GameEvent)
}

type GameEventHandlerWapper func(*Player, *GameEvent)

func (gehw GameEventHandlerWapper) Handler(p *Player, ge *GameEvent) {
	gehw(p, ge)
}

type GameEventHandlerMgr struct {
	ehs map[int32][]GameEventHandler
}

func (this *GameEventHandlerMgr) RegisteGameEventHandler(eventType int32, geh GameEventHandler) {
	if pool, exist := this.ehs[eventType]; exist {
		pool = append(pool, geh)
		this.ehs[eventType] = pool
	} else {
		this.ehs[eventType] = []GameEventHandler{geh}
	}
}

func (this *GameEventHandlerMgr) UnregisteGameEventHandler(eventType int32, geh GameEventHandler) {
	if pool, exist := this.ehs[eventType]; exist {
		cnt := len(pool)
		for i := 0; i < cnt; i++ {
			if pool[i] == geh {
				pool[i] = pool[cnt-1]
				pool = pool[:cnt-1]
				this.ehs[eventType] = pool
				break
			}
		}
	}
}

func (this *GameEventHandlerMgr) ProcessGameEvent(p *Player, ge *GameEvent) {
	if gehs, exist := this.ehs[ge.eventType]; exist {
		for _, h := range gehs {
			h.Handler(p, ge)
		}
	}
}

func (this *Player) FireGameEvent() {
	GameEventHandlerMgrSingleton.ProcessGameEvent(this, &this.ge)
}
