package base

const (
	PlayerState_Online           int32 = 1 << iota //在线标记 1
	PlayerState_Ready                              //准备标记 2
	PlayerState_SceneOwner                         //房主标记 3
	PlayerState_Choke                              //呛标记 被复用于金花，是否被动弃牌 4
	PlayerState_Ting                               //听牌标记 5  金花复用，标记最后押注时，是否看牌
	PlayerState_NoisyBanker                        //闹庄标记 6  金花复用，标记allin时，是否看牌
	PlayerState_WaitOp                             //等待操作标记 7
	PlayerState_Auto                               //托管状态 8
	PlayerState_Check                              //已看牌状态 9
	PlayerState_Fold                               //弃牌状态 10
	PlayerState_Lose                               //输状态 11
	PlayerState_Win                                //赢状态 12
	PlayerState_WaitNext                           //等待下一局游戏 13
	PlayerState_GameBreak                          //不能继续游戏 14
	PlayerState_Leave                              //暂离状态 15
	PlayerState_Audience                           //观众标记 16
	PlayerState_AllIn                              //allin标记 17
	PlayerState_FinalAllIn                         //最后一圈，最后一个人allin标记 18
	PlayerState_Show                               //亮牌标记 19
	PlayerState_EnterSceneFailed                   //进场失败 20
	PlayerState_PKLost                             //发起Pk,失败 21
	PlayerState_IsChangeCard                       //牛牛标识是否换牌 22
	PlayerState_IsPayChangeCard                    //牛牛标识是否充值换牌 23
	PlayerState_Bankruptcy                         //玩家破产 24
	PlayerState_MatchQuit                          //退赛标记 25
	PlayerState_AllFollow                          //跟到底状态 26
	PlayerState_SAdjust                            //单控状态 27
	PlayerState_Max
)
