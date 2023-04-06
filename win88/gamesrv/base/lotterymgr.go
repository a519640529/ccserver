package base

import (
	"games.yol.com/win88/model"
)

var LotteryMgrSington = &LotteryMgr{
	cfgParser: make(map[int32]model.LotteryConfigParser),
}

//彩金池管理器
type LotteryMgr struct {
	cfgParser map[int32]model.LotteryConfigParser
}

func (this *LotteryMgr) RegisteLotteryConfigParser(gameid, mode int32, parser model.LotteryConfigParser) {
	key := gameid*10000 + mode
	this.cfgParser[key] = parser
}

func (this *LotteryMgr) ParseLotteryConfig(gameid, mode int32, cfg string) (model.LotteryConfiger, error) {
	key := gameid*10000 + mode
	if parser, exist := this.cfgParser[key]; exist {
		return parser(cfg)
	}
	return nil, nil
}
