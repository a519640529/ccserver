package main

import (
	"time"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
)

type SaveTaskHandler interface {
	Time2Save()
}

var SaverSliceNumber = 600

var DbSaver_Inst = &DbSaver{
	Tick:  int32(SaverSliceNumber),
	index: 0,
	init:  false,
	list:  make([]*SaverArray, SaverSliceNumber),
	queue: make([]*BalanceQueue, 10),
	pool:  make(map[SaveTaskHandler]*SaverArray),
}

type DbSaver struct {
	Tick  int32
	index int32
	step  int32
	list  []*SaverArray
	queue []*BalanceQueue
	init  bool
	pool  map[SaveTaskHandler]*SaverArray
}

func (this *DbSaver) pushBalanceSaverArray(sth SaveTaskHandler) {
	if sth == nil {
		return
	}
	if _, exist := this.pool[sth]; exist {
		return
	}
	for pos, bq := range this.queue {
		size := len(bq.queue)
		if size > 0 {
			arr := bq.queue[size-1]
			if pos+1 >= len(this.queue) {
				this.queue = append(this.queue, &BalanceQueue{})
			}
			this.queue[pos+1].queue = append(this.queue[pos+1].queue, arr)
			this.queue[pos].queue = bq.queue[:size-1]
			arr.bqPos = len(this.queue[pos+1].queue) - 1
			arr.Array = append(arr.Array, sth)
			this.pool[sth] = arr
			return
		}
	}
	return
}

func (this *DbSaver) RegisterDbSaverTask(i interface{}) {
	if st, ok := i.(SaveTaskHandler); ok {
		this.pushBalanceSaverArray(st)
	}
}

func (this *DbSaver) UnregisteDbSaveTask(i interface{}) {
	if sth, ok := i.(SaveTaskHandler); ok {
		if arr, exist := this.pool[sth]; exist {
			delete(this.pool, sth)
			count := len(arr.Array)
			for i := 0; i < count; i++ {
				if arr.Array[i] == sth {
					arr.Array[i] = arr.Array[count-1]
					arr.Array = arr.Array[:count-1]

					bqPos := arr.bqPos
					queCount := len(this.queue[count].queue)
					this.queue[count].queue[bqPos] = this.queue[count].queue[queCount-1]
					this.queue[count].queue[bqPos].bqPos = bqPos
					this.queue[count].queue = this.queue[count].queue[:queCount-1]
					this.queue[count-1].queue = append(this.queue[count-1].queue, arr)
					arr.bqPos = len(this.queue[count-1].queue) - 1
					return
				}
			}
		} else {
			logger.Logger.Info("Player not in dbsaver")
		}
	}
}

type SaverArray struct {
	Array []SaveTaskHandler
	bqPos int
}
type BalanceQueue struct {
	queue []*SaverArray
}

// //////////////////////////////////////////////////////////////////
// / Module Implement [beg]
// //////////////////////////////////////////////////////////////////
func (this *DbSaver) ModuleName() string {
	return "dbsaver"
}

func (this *DbSaver) Init() {
	if this.init == false {
		for i := 0; i < len(this.queue); i++ {
			this.queue[i] = &BalanceQueue{}
		}
		//初始化平衡数组，所有平衡队列容量为0
		for i := 0; i < int(this.Tick); i++ {
			this.list[i] = &SaverArray{bqPos: i}
			this.queue[0].queue = append(this.queue[0].queue, this.list[i])
		}
		this.init = true
	}
}

func (this *DbSaver) Update() {
	if this.index == this.Tick {
		this.index = 0
	}
	sa := this.list[this.index]
	for _, sth := range sa.Array {
		sth.Time2Save()
	}
	this.index = this.index + 1
}

func (this *DbSaver) Shutdown() {
	module.UnregisteModule(this)
}

////////////////////////////////////////////////////////////////////
/// Module Implement [end]
////////////////////////////////////////////////////////////////////

func init() {
	module.RegisteModule(DbSaver_Inst, time.Second, 0)
}
