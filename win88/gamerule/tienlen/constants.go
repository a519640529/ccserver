package tienlen

import "time"

////////////////////////////////////////////////////////////////////////////////
//tienlen
////////////////////////////////////////////////////////////////////////////////

const (
	TestOpen          bool  = false //测试开关
	MaxNumOfPlayer    int   = 4     //最多人数
	HandCardNum       int32 = 13    //手牌数量
	InvalideCard      int32 = -1    //默认牌
	InvalidePos       int32 = -1
	RobotGameTimesMin int32 = 5 //机器人参与游戏次数下限
	RobotGameTimesMax int32 = 5 //机器人参与游戏次数上限
	DelayCanOp        int   = 3 //根据上家牌型额外延迟下家出牌时间
)

const (
	TienLenWaitPlayerTimeout = time.Second * 1
	TienLenWaitStartTimeout  = time.Second * 10 //人数够开启游戏, 延迟X秒开始游戏
	TienLenHandCardTimeout   = time.Second * 3  //发牌
	TienLenPlayerOpTimeout   = time.Second * 15 //出牌（玩家操作阶段）
	TienLenBilledTimeout     = time.Second * 7  //结算
)

// 场景状态
const (
	TienLenSceneStateWaitPlayer int = iota //0 人数不够开启游戏，等待玩家上线
	TienLenSceneStateWaitStart             //1 人数够开启游戏, 延迟X秒开始游戏
	TienLenSceneStateHandCard              //2 发牌
	TienLenSceneStatePlayerOp              //3 出牌（玩家操作阶段）
	TienLenSceneStateBilled                //4 结算
	TienLenSceneStateMax
)

// 玩家操作
const (
	TienLenPlayerOpNull  int32 = iota //0,初始值
	TienLenPlayerOpPlay               //1,出牌
	TienLenPlayerOpPass               //2,过牌
	TienLenPlayerOpStart              //3,房主开始游戏
)
