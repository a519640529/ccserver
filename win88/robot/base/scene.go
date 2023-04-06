package base

import "github.com/idealeak/goserver/core/netlib"

var NilScene Scene

type Scene interface {
	GetAIMode() int32
	SetAIMode(aiMode int32)
	GetRoomId() int32
	GetRoomMode() int32
	GetGameId() int32
	GetPlayerByPos(pos int32) Player
	GetPlayerBySnid(snid int32) Player
	GetMe(s *netlib.Session) Player
	AddPlayer(p Player)
	DelPlayer(snid int32)
	IsFull() bool
	IsMatchScene() bool
	Update(ts int64)
}

type BaseScene struct {
	AIMode          int32
	RobotTypeAIName map[int]string
}

func (this *BaseScene) GetAIMode() int32 {
	return this.AIMode
}

func (this *BaseScene) SetAIMode(aiMode int32) {
	this.AIMode = aiMode
}
