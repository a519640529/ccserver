//ai全局管理
package base

import (
	"github.com/idealeak/goserver/core/logger"
	b3 "github.com/magicsea/behavior3go"
	b3config "github.com/magicsea/behavior3go/config"
	b3core "github.com/magicsea/behavior3go/core"
	b3loader "github.com/magicsea/behavior3go/loader"
	"strconv"
)

var aiPath = "../data/botai/"

func InitBev() {
	mapTrees = make(map[string]*b3core.BehaviorTree)
}

func GetBevTree(name string) *b3core.BehaviorTree {
	return CreateBevTree(name)
}

func InitTree(name string) {
	CreateBevTree(name)
}

//创建一个行为树
var mapTrees map[string]*b3core.BehaviorTree

func CreateBevTree(name string) *b3core.BehaviorTree {
	if mapTrees == nil {
		InitBev()
	}

	b, ok := mapTrees[name]
	if ok {
		return b
	}
	logger.Logger.Tracef("create tree:%v", name)
	fullName := aiPath + name
	config, ok := b3config.LoadTreeCfg(fullName)
	if !ok {
		logger.Logger.Errorf("LoadTreeCfg fail:" + fullName)
		return nil
	}
	extMaps := createExtStructMaps()
	tree := b3loader.CreateBevTreeFromConfig(config, extMaps)

	mapTrees[name] = tree
	return tree
}

func RefleshBevTree(name string) *b3core.BehaviorTree {
	if mapTrees == nil {
		InitBev()
	}

	logger.Logger.Tracef("create tree:%v", name)
	fullName := aiPath + name
	config, ok := b3config.LoadTreeCfg(fullName)
	if !ok {
		logger.Logger.Errorf("LoadTreeCfg fail:" + fullName)
		return nil
	}
	extMaps := createExtStructMaps()
	tree := b3loader.CreateBevTreeFromConfig(config, extMaps)

	mapTrees[name] = tree
	return tree
}

//自定义的节点
func createExtStructMaps() *b3.RegisterStructMaps {
	st := b3.NewRegisterStructMaps()
	//actions
	st.Register("RandIntAction", &RandIntAction{})
	st.Register("RandWait", &RandWait{})
	st.Register("SubTree", &SubTreeNode{})
	st.Register("LogAction", &LogAction{})
	//st.Register("RVBLastWinArea", &RVBLastWinArea{})
	st.Register("LeaveGame", &LeaveGame{})
	st.Register("GetOutLimitCoin", &GetOutLimitCoin{})
	st.Register("SetIntAction", &SetIntAction{})
	st.Register("SetIntMulti", &SetIntMulti{})
	//st.Register("RVBBetPct", &RVBBetPct{})
	//st.Register("RVBBetCoin", &RVBBetCoin{})
	//st.Register("RVBCheckBetCoin", &RVBCheckBetCoin{})
	st.Register("GetPlayerCoin", &GetPlayerCoin{})
	st.Register("SetIntDiv", &SetIntDiv{})
	st.Register("GetPlayerTakeCoin", &GetPlayerTakeCoin{})
	//st.Register("DVTLastWinArea", &DVTLastWinArea{})
	//st.Register("DVTBetCoin", &DVTBetCoin{})
	//st.Register("DVTCheckBetCoin", &DVTCheckBetCoin{})

	//conditions
	st.Register("CheckBool", &CheckBool{})
	st.Register("CheckInt", &CheckInt{})
	st.Register("CheckPlayerCoin", &CheckPlayerCoin{})
	//st.Register("RVBSceneState", &RVBSceneState{})
	//st.Register("RVBHistoryIsSame", &RVBHistoryIsSame{})
	st.Register("CheckPlayerGameNum", &CheckPlayerGameNum{})
	st.Register("CheckPlayerLastWinOrLost", &CheckPlayerLastWinOrLost{})
	//st.Register("DVTSceneState", &DVTSceneState{})
	//st.Register("DVTHistoryIsSame", &DVTHistoryIsSame{})

	//composite
	st.Register("Random", &RandomComposite{})
	st.Register("Parallel", &ParallelComposite{})
	st.Register("RandomWeightComposite", &RandomWeightComposite{})

	return st
}

func getBoardIntByPreStr(blackData *b3core.Blackboard, key, treeScope, nodeScope string) int {
	ret := 0
	if len(key) <= 0 {
		return 0
	}
	if key[0] == '@' {
		ret, _ = strconv.Atoi(key[1:])
	} else {
		ret = blackData.GetInt(key, treeScope, nodeScope)
	}

	return ret
}
