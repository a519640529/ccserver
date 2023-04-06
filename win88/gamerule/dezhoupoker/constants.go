package dezhoupoker

import "time"

////////////////////////////////////////////////////////////////////////////////
//德州扑克
////////////////////////////////////////////////////////////////////////////////
//房间类型
const (
	RoomMode_Normal   int = iota //德州
	RoomMode_FiveXTwo            //德州五选二
	RoomMode_Max
)
const (
	MaxNumOfPlayer    int32 = 9 //最多人数
	MaxNumOfPlayerTwo int32 = 8 //最多人数(五选二)
	HandCardNum       int32 = 2 //手牌数量
	CommunityCardNum  int32 = 5 //公牌
	TotalCardNum      int32 = HandCardNum + CommunityCardNum
	FlopCardNum       int32 = 3               //翻牌
	TurnCardPos       int32 = FlopCardNum + 1 //转牌位置
	RiverCardPos      int32 = TurnCardPos + 1 //河牌位置

	ANTE_SCORE int32 = 0 //前注
)
const (
	BEHIND_CARD   int32 = 55
	INVALIDE_POS  int32 = -1
	INVALIDE_CARD int32 = -1
)

//dz创建房间的参数信息
const (
	DZSceneParam_MaxCoin int = iota //

	DZThreeSceneParam_Max
)

const (
	RobotGameTimesMin int32 = 5  //机器人参与游戏次数下限
	RobotGameTimesMax int32 = 10 //机器人参与游戏次数上限
)

const (
	CardType_HandCard int32 = iota //公牌
	CardType_FlopCard
	CardType_TrunCard
	CardType_RiverCard
)

const (
	RoleType_Player int32 = iota //普通玩家
	RoleType_Banker
	RoleType_SmallBlind
	RoleType_BigBlind
)

//玩家位置
const (
	Pos_UTG  int32 = iota //Under the gun 枪口
	Pos_UTG1              //Under the gun +1 枪口+1
	Pos_MP1               //Middle position 1 中位1
	Pos_MP2               //Middle position 2 中位2
	Pos_HJ                //Hijack 劫位
	Pos_CO                //Cut off 关位
	Pos_BTN               //Button 庄家
	Pos_SB                //Small blind 小盲位
	Pos_BB                //Big blind 大盲位
)

var PosDesc = []string{"枪口", "枪口+1", "中位1", "中位2", "劫位", "关位", "庄位", "小盲位", "大盲位"}

const (
	DezhouPokerOffsetTimeout = 1 //结算等待时间

	DezhouPokerWaitPlayerTimeout = time.Second * 10 //等待真人时间
	DezhouPokerWaitStartTimeout  = time.Second * 3  //延迟开始时间
	//DezhouPokerSelectBankerAndBlindsTimeout    = time.Second * 2  //选庄时间
	DezhouPokerAntTimeout = time.Second * 1 //前注时间
	//DezhouPokerBlindsTimeout    = time.Second * 0  //大小盲下注时间
	DezhouPokerHandCardTimeout   = time.Second * 1  //发手牌等待时间
	DezhouPokerSelectCardTimeout = time.Second * 10 //选手牌等待时间

	DezhouPokerHandCardBetTimeout = time.Second * 15 * DezhouPokerOffsetTimeout //发手牌等待时间
	DezhouPokerFlopTimeout        = time.Second * 1                             //发3张翻牌等待时间
	DezhouPokerFlopBetTimeout     = time.Second * 15 * DezhouPokerOffsetTimeout //发3张翻牌等待时间
	DezhouPokerTurnTimeout        = time.Second * 1                             //发1张转牌等待时间
	DezhouPokerTurnBetTimeout     = time.Second * 15 * DezhouPokerOffsetTimeout //发1张转牌等待时间
	DezhouPokerRiverTimeout       = time.Second * 1                             //发1张转牌等待时间
	DezhouPokerRiverBetTimeout    = time.Second * 15 * DezhouPokerOffsetTimeout //发1张转牌等待时间
	//DezhouPokerBilledTimeout    = time.Second * 10//结算等待时间
	DezhouPokerBilledTimeoutNormal          = time.Second * 2 //结算等待时间
	DezhouPokerBilledTimeoutMiddle          = time.Second * 1 //结算等待时间
	DezhouPokerBilledTimeoutAllIn           = time.Second * 1 //补牌
	DezhouPokerBilledTimeoutPerPlayerBilled = time.Second * 2 //等待结束
	DezhouPokerBilledTimeoutWaitCheckCard   = time.Second * 3 //结算结束,预留一点时间让玩家能看清牌型
)

//场景状态
const (
	DezhouPokerSceneStateWaitPlayer            int = iota //0 人数不够开启游戏，等待玩家上线
	DezhouPokerSceneStateWaitStart                        //1 人数够开启游戏, 延迟X秒开始游戏
	DezhouPokerSceneStateSelectBankerAndBlinds            //2 选庄家和大小盲

	DezhouPokerSceneStateAnte   //3 下前注0，预留
	DezhouPokerSceneStateBlinds //4 下大小盲

	DezhouPokerSceneStateHandCard    //5 发手牌
	DezhouPokerSceneStateSelectCard  //6 选牌状态
	DezhouPokerSceneStateHandCardBet //7 手牌下注1

	DezhouPokerSceneStateFlop    //8 发3张翻牌
	DezhouPokerSceneStateFlopBet //9 翻牌下注2

	DezhouPokerSceneStateTurn    //10 转牌
	DezhouPokerSceneStateTurnBet //11 转牌下注3

	DezhouPokerSceneStateRiver    //12 河牌
	DezhouPokerSceneStateRiverBet //13 河牌下注4

	DezhouPokerSceneStateBilled       //14 结算方式选择
	DezhouPokerSceneStateBilledNormal //15 正常结算
	DezhouPokerSceneStateBilledMiddle //16 半路结算 还剩一个未弃牌的玩家
	DezhouPokerSceneStateBilledAllIn  //17 全部allin, 或者只有一个未Allin

	DezhouPokerSceneStateGameEnd //18 结算状态

	DezhouPokerSceneStateMax
)

//玩家操作。当轮第一个下注的是加注。大小盲 不显示任何动作
const (
	DezhouPokerPlayerOpNull       int32 = iota //0,初始值
	DezhouPokerPlayerOpCallAntes               //1,下底注
	DezhouPokerPlayerOpCall                    //2,跟进
	DezhouPokerPlayerOpFold                    //3,弃牌
	DezhouPokerPlayerOpCheck                   //4,让牌
	DezhouPokerPlayerOpRaise                   //5,加注
	DezhouPokerPlayerOpAllIn                   //6,全压
	DezhouPokerPlayerOpSmallBlind              //7,小盲
	DezhouPokerPlayerOpBigBlind                //8,大盲

	DezhouPokerPlayerOpAutoBuyIn    //9,自动买入 买入数量
	DezhouPokerPlayerOpAutoBuyInCfg //10,请求自动买入配置
	DezhouPokerPlayerOpSelectCard   //11,选牌

	//DezhouPokerPlayerOpSitDown		//坐下
	//DezhouPokerPlayerOpStandUp		//站起

)
