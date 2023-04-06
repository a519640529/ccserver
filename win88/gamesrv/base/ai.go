package base

type AI interface {
	//挂载玩家
	SetOwner(*Player)
	//获取挂载玩家
	GetOwner() *Player
	//获取属性
	GetAttribute(key interface{}) (interface{}, bool)
	//设置属性
	SetAttribute(key, val interface{})
	//开启事件
	OnStart()
	//停止事件
	OnStop()
	//心跳事件
	OnTick(s *Scene)
	//自己进入事件
	OnSelfEnter(s *Scene, p *Player)
	//自己离开事件
	OnSelfLeave(s *Scene, p *Player, reason int)
	//其他玩家进入事件
	OnPlayerEnter(s *Scene, p *Player)
	//其他玩家离开事件
	OnPlayerLeave(s *Scene, p *Player, reason int)
	//其他玩家掉线事件
	OnPlayerDropLine(s *Scene, p *Player)
	//其他玩家重连事件
	OnPlayerRehold(s *Scene, p *Player)
	//其他玩家返回房间事件
	OnPlayerReturn(s *Scene, p *Player)
	//其他玩家操作事件
	OnPlayerOp(s *Scene, p *Player, opcode int, params []int64) bool
	//其他玩家操作事件
	OnPlayerOperate(s *Scene, p *Player, params interface{}) bool
	//其他玩家事件
	OnPlayerEvent(s *Scene, p *Player, evtcode int, params []int64)
	//其他观众进入事件
	OnAudienceEnter(s *Scene, p *Player)
	//其他观众离开事件
	OnAudienceLeave(s *Scene, p *Player, reason int)
	//其他观众坐下事件
	OnAudienceSit(s *Scene, p *Player)
	//其他观众掉线事件
	OnAudienceDropLine(s *Scene, p *Player)
	//房间状态变化事件
	OnChangeSceneState(s *Scene, oldstate, newstate int)
}

type BaseAI struct {
	owner     *Player
	attribute map[interface{}]interface{}
}

//挂载玩家
func (b *BaseAI) SetOwner(p *Player) {
	b.owner = p
}

//获取挂载玩家
func (b *BaseAI) GetOwner() *Player {
	return b.owner
}

//获取属性
func (b *BaseAI) GetAttribute(key interface{}) (interface{}, bool) {
	if b.attribute != nil {
		v, ok := b.attribute[key]
		return v, ok
	}
	return nil, false
}

//设置属性
func (b *BaseAI) SetAttribute(key, val interface{}) {
	if b.attribute != nil {
		b.attribute[key] = val
	}
}

//开启事件
func (b *BaseAI) OnStart() {

}

//关闭事件
func (b *BaseAI) OnStop() {

}

//心跳事件
func (b *BaseAI) OnTick(s *Scene) {

}

//自己进入事件
func (b *BaseAI) OnSelfEnter(s *Scene, p *Player) {
	if !p.IsLocal {
		return
	}
	takeCoin, leaveCoin, gameTimes := s.RandTakeCoin(p)
	p.Coin = takeCoin
	p.SetTakeCoin(takeCoin)
	p.ExpectGameTime = int32(gameTimes)
	p.ExpectLeaveCoin = leaveCoin
	//当局游戏结束后剩余金额 起始设置
	p.SetCurrentCoin(takeCoin)
	p.LastSyncCoin = p.Coin
}

//自己离开事件
func (b *BaseAI) OnSelfLeave(s *Scene, p *Player, reason int) {

}

//其他玩家进入事件
func (b *BaseAI) OnPlayerEnter(s *Scene, p *Player) {

}

//其他玩家离开事件
func (b *BaseAI) OnPlayerLeave(s *Scene, p *Player, reason int) {

}

//其他玩家掉线
func (b *BaseAI) OnPlayerDropLine(s *Scene, p *Player) {

}

//其他玩家重连
func (b *BaseAI) OnPlayerRehold(s *Scene, p *Player) {

}

//其他玩家 返回房间
func (b *BaseAI) OnPlayerReturn(s *Scene, p *Player) {

}

//其他玩家操作事件
func (b *BaseAI) OnPlayerOp(s *Scene, p *Player, opcode int, params []int64) bool {
	return true
}

//其他玩家操作事件
func (b *BaseAI) OnPlayerOperate(s *Scene, p *Player, params interface{}) bool {
	return true
}

//其他玩家事件
func (b *BaseAI) OnPlayerEvent(s *Scene, p *Player, evtcode int, params []int64) {

}

//观众进入事件
func (b *BaseAI) OnAudienceEnter(s *Scene, p *Player) {

}

//观众离开事件
func (b *BaseAI) OnAudienceLeave(s *Scene, p *Player, reason int) {

}

//观众坐下事件
func (b *BaseAI) OnAudienceSit(s *Scene, p *Player) {

}

//观众掉线事件
func (b *BaseAI) OnAudienceDropLine(s *Scene, p *Player) {

}

//房间状态变化事件
func (b *BaseAI) OnChangeSceneState(s *Scene, oldstate, newstate int) {

}
