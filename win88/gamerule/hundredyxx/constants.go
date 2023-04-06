package hundredyxx

import (
	"time"
)

const DICE_NUM int = 3 //骰子数量
const MAX_RATE = 1

//骰子
const (
	PointKind_Yu   int = iota //鱼（红）
	PointKind_Ji              //鸡（红）
	PointKind_Xia             //虾（绿）
	PointKind_Xie             //蟹（绿）
	PointKind_HuLu            //葫芦（蓝）
	PointKind_Lu              //鹿（蓝）
	PointKind_Max
)

//下注区域
const (
	BetField_Single_1 int = iota
	BetField_Single_2
	BetField_Single_3
	BetField_Single_4
	BetField_Single_5
	BetField_Single_6
	BetField_Double_7
	BetField_Double_8
	BetField_Double_9
	BetField_Double_10
	BetField_Double_11
	BetField_Double_12
	BetField_Double_13
	BetField_Double_14
	BetField_Double_15
	BetField_Double_16
	BetField_Double_17
	BetField_Double_18
	BetField_Double_19
	BetField_Double_20
	BetField_Double_21
	BetField_MAX
)

////////////////////////////////////////////////////////////////////////////////
//百人鱼虾蟹
////////////////////////////////////////////////////////////////////////////////
const (
	HundredYXXSendCardTimeout = time.Second * 3  //摇骰子
	HundredYXXStakeTimeout    = time.Second * 17 //押注
	HundredYXXOpenCardTimeout = time.Second * 5  //开奖
	HundredYXXBilledTimeout   = time.Second * 7  //结算
	HundredYXXSendBetTime     = time.Second * 1  //发送下注数据时间间隔
)

//场景状态
const (
	HundredYXXSceneStateSendCard int = iota //摇骰子
	HundredYXXSceneStateStake               //下注
	HundredYXXSceneStateOpenCard            //开奖
	HundredYXXSceneStateBilled              //结算
	HundredYXXSceneStateMax
)

//玩家操作
const (
	HundredYXXPlayerOpBet           int = iota //下注
	HundredYXXPlayerOpUpBanker                 //上庄
	HundredYXXPlayerOpDwonBanker               //下庄
	HundredYXXPlayerOpUpList                   //上庄列表
	HundredYXXTrend                            //走势
	HundredYXXPlayerList                       //玩家列表
	HundredYXXPlayerOpNowDwonBanker            //在庄的下庄
)

var BetFieldWinPoint = [][]int{{PointKind_Yu}, {PointKind_Xia}, {PointKind_HuLu},
	{PointKind_Lu}, {PointKind_Xie}, {PointKind_Ji},
	{PointKind_Yu, PointKind_Xie}, {PointKind_Lu, PointKind_Xia}, {PointKind_Yu, PointKind_HuLu},
	{PointKind_Lu, PointKind_Ji}, {PointKind_Xie, PointKind_HuLu}, {PointKind_Ji, PointKind_Xia},
	{PointKind_Yu, PointKind_Ji}, {PointKind_Lu, PointKind_HuLu}, {PointKind_Yu, PointKind_Xia},
	{PointKind_Xia, PointKind_HuLu}, {PointKind_Yu, PointKind_Lu}, {PointKind_Xie, PointKind_Xia},
	{PointKind_HuLu, PointKind_Ji}, {PointKind_Xie, PointKind_Lu}, {PointKind_Ji, PointKind_Xie},
}

//var BetFieldMultiple = []int{1, 1, 1, 1, 1, 1, 5, 5, 5, 5, 5, 5}
var WinKindStr = []string{"鱼", "虾", "葫芦", "鹿", "蟹", "鸡",
	"鱼和蟹", "鹿和虾", "鱼和葫芦",
	"鹿和鸡", "蟹和葫芦", "鸡和虾",
	"鱼和鸡", "鹿和葫芦", "鱼和虾",
	"虾和葫芦", "鱼和鹿", "蟹和虾",
	"葫芦和鸡", "蟹和鹿", "鸡和蟹",
}
