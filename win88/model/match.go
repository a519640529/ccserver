package model

import "net"

const (
	MatchType_Unkonw int32 = iota
	MatchType_Free
	MatchType_Max
)

const (
	MatchOp_Unknow       int32 = iota
	MatchOp_Continue           //继续
	MatchOp_Break              //中断
	MatchOp_Quit               //退赛
	MatchOp_AddSignupFee       //补充报名费
	MatchOp_Max
)

// 比赛报名消耗类型
const (
	MatchCostType_Free   int32 = iota //免费次数
	MatchCostType_Ticket              //报名券
	MatchCostType_Coin                //金币
	MatchCostType_Max
)

type MatchId int32

func MakeMatchId(matchType, matchRuleId, matchSceneType int32) MatchId {
	return MatchId(matchType*10000 + matchRuleId*100 + matchSceneType)
}

func (id MatchId) MatchType() int32 {
	return int32(id) / 10000
}

func (id MatchId) MatchRuleId() int32 {
	return (int32(id) % 10000) / 100
}

func (id MatchId) MatchSceneType() int32 {
	return int32(id) % 100
}

type Match struct {
	Id MatchId
}

// 比赛成绩
type MatchAchievement struct {
	MatchId       int32 //比赛ID
	Coin          int32 //累计获得金币数量
	Grade         int32 //累计获得积分数量
	BestRank      int32 //最高排名
	FinalTimes    int32 //决赛次数
	GameTimes     int32 //游戏次数
	ChampionTimes int32 //冠军次数
	Ts            int32 //时间戳
	CreateTs      int32 //首次创建记录时间戳
	Datas         []int32
}

type PlayerMatchSignup struct {
	SnId      int32  //玩家id
	CostType  int32  //报名费类型 0:免费 1:入场券 2:金币
	CostValue int32  //消耗的值
	Ts        int64  //报名时间戳 秒
	IsRob     bool   //是否是机器人
	IP        net.IP //IP地址
}

// 比赛报名信息
type MatchSignup struct {
	MatchId    int32                //比赛id
	SignupData []*PlayerMatchSignup //报名费
}
