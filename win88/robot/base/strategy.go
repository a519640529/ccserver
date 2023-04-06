package base

const (
	Strategy_CreateRoom int = iota
	Strategy_EnterRoom
	Strategy_LeaveRoom
	Strategy_DestroyRoom
	Strategy_Watch
	Strategy_CreateAndDestroyRoom
	Strategy_WaitInvite
)

var StrategyMgrSington = &StrategyMgr{}

type StrategyMgr struct {
	strategies []int
}

func (sm *StrategyMgr) Init() {
	//sm.strategies = []int{Strategy_CreateRoom, Strategy_EnterRoom, Strategy_EnterRoom, Strategy_EnterRoom}
	//sm.strategies = []int{Strategy_EnterRoom, Strategy_EnterRoom, Strategy_EnterRoom, Strategy_EnterRoom}
}

func (sm *StrategyMgr) PopStrategy() int {
	cnt := len(sm.strategies)
	if cnt != 0 {
		strategy := sm.strategies[0]
		sm.strategies = sm.strategies[1:]
		return strategy
	}
	return Strategy_WaitInvite
}

func (sm *StrategyMgr) TryPopStrategy(strategy int) {
	for i, s := range sm.strategies {
		if s == strategy {
			arr := sm.strategies[:i]
			sm.strategies = append(arr, sm.strategies[i+1:]...)
		}
	}
}
func init() {
	StrategyMgrSington.Init()
}
