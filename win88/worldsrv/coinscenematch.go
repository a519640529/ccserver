package main

type FuncCoinSceneMatch func(csp *CoinScenePool, p *Player, scenes map[int]*Scene, sameIpLimit bool, exclude []int32) *Scene

var coinSceneMatchFuncPool = make(map[int]FuncCoinSceneMatch)

func RegisteCoinSceneMatchFunc(id int, f FuncCoinSceneMatch) {
	coinSceneMatchFuncPool[id] = f
}

func GetCoinSceneMatchFunc(id int) FuncCoinSceneMatch {
	if f, exist := coinSceneMatchFuncPool[id]; exist {
		return f
	}
	return nil
}

//
////通用的根据游戏id决定匹配规则
//func CoinSceneMatch_ByNormalGameId(csp *CoinScenePool, p *Player, scenes map[int]*Scene, sameIpLimit bool, exclude []int32) *Scene {
//	if csp == nil || p == nil {
//		return nil
//	}
//	if p.IsFoolPlayer == nil ||
//		(p.IsFoolPlayer != nil && p.IsFoolPlayer[csp.dbGameFree.GetGameDif()]) {
//		//如果该玩家是新手玩家	炸金花新手
//		var foolScene *Scene
//		for _, s := range scenes {
//			if s != nil {
//				if sceneLimitMgr.LimitAvgPlayer(s, len(csp.players)) {
//					continue
//				}
//				if sp, ok := s.sp.(*ScenePolicyData); ok {
//					if !s.starting || sp.EnterAfterStart {
//						cnt := s.GetWhitePlayerCnt()
//						fcnt := s.GetFoolPlayerCnt()
//						lcnt := s.GetLostPlayerCnt()
//						//没有白名单的玩家  没有新手玩家  没有输钱的玩家
//						if cnt == 0 && fcnt == 0 && lcnt == 0 {
//							foolScene = s
//						}
//					}
//				}
//			}
//		}
//		return foolScene
//	}
//	return nil
//}
//
//func init() {
//	RegisteCoinSceneMatchFunc(common.GameId_WinThree, CoinSceneMatch_ByNormalGameId)
//}
