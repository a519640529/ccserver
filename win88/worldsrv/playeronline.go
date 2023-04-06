package main

import (
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/module"
	"time"
)

var PlayerOnlineSington = &PlayerOnlineEvent{
	Online: make(map[string]int),
}

type PlayerOnlineEvent struct {
	Online map[string]int
	Check  bool
}

func (p *PlayerOnlineEvent) ModuleName() string {
	return "PlayerOnlineEvent"
}

func (p *PlayerOnlineEvent) Init() {
}

// 每五秒钟统计一次在线数据
// 没有登录，登出，掉线情况直接不统计
func (p *PlayerOnlineEvent) Update() {
	if !p.Check {
		return
	}
	p.Check = false
	m := map[string]int{}
	for _, player := range PlayerMgrSington.playerMap {
		if player != nil && !player.IsRob && player.IsOnLine() {
			m[player.Platform] = m[player.Platform] + 1
		}
	}
	if len(m) == len(p.Online) {
		for k, v := range m {
			if p.Online[k] != v {
				goto here
			}
		}
		return
	}

here:
	p.Online = m
	LogChannelSington.WriteMQData(model.GenerateOnline(p.Online))
}

func (p *PlayerOnlineEvent) Shutdown() {
	module.UnregisteModule(p)
}

func init() {
	module.RegisteModule(PlayerOnlineSington, 5*time.Second, 0)
}
