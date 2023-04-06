package fishing

import "sync"

type Fish struct {
	id        int32 //实例id
	tempId    int32 //模板id
	coin      int32 //价值
	birthTick int32 //出生时间
	dieTick   int32 //死亡时间
}

var fishPool = sync.Pool{
	New: func() interface{} {
		return &Fish{}
	},
}

func NewFish(id, tempId, coin, birthTick, dieTick int32) *Fish {
	f := fishPool.Get().(*Fish)
	f.id = id
	f.tempId = tempId
	f.coin = coin
	f.birthTick = birthTick
	f.dieTick = dieTick
	return f
}

func DestoryFish(f *Fish) {
	fishPool.Put(f)
}
