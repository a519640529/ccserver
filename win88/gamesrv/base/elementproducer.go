package base

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
	"sync/atomic"
	"time"
)

type ElementPool struct {
	key  int32
	pool chan interface{}
}

type ElementProducer struct {
	times int64
	cnt   int32
	max   int32
	pool  []*ElementPool
	nf    ProduceFunc //正常的生产函数
	bf    ProduceFunc //保底的生产函数
	quit  bool
}

type ProduceFunc func() (interface{}, int32)

type ElementProducerMgr struct {
	producer map[int32]*ElementProducer
}

var ElementProducerMgrSington = &ElementProducerMgr{
	producer: make(map[int32]*ElementProducer),
}

func CreateElementProducer(max int32, nf, bf ProduceFunc) *ElementProducer {
	ep := &ElementProducer{
		max:  max,
		nf:   nf,
		bf:   bf,
		pool: make([]*ElementPool, max, max),
	}
	return ep
}

func (ep *ElementProducer) Start() {
	go _start_produce_routine(ep)
}

func (ep *ElementProducer) Stop() {
	ep.quit = true
}

func (ep *ElementProducer) FetchOneElement(key int32) interface{} {
	if key < 0 || key >= int32(len(ep.pool)) {
		return nil
	}
	pool := ep.pool[key]
	if pool != nil {
		select {
		case d, ok := <-pool.pool:
			if ok {
				atomic.AddInt32(&ep.cnt, -1)
				return d
			} else {
				//向下取一个结果
				for i := key - 1; i >= 0; i-- {
					if ep.pool[i] != nil && len(ep.pool[i].pool) != 0 {
						return ep.FetchOneElement(i)
					}
				}
				d, _ := ep.bf()
				return d
			}
		default:
		}
	}
	//向下取一个结果
	for i := key - 1; i >= 0; i-- {
		if ep.pool[i] != nil && len(ep.pool[i].pool) != 0 {
			return ep.FetchOneElement(i)
		}
	}
	d, _ := ep.bf()
	return d
}

func (ep *ElementProducer) Dump() {
	logger.Logger.Tracef("[[[[[[[[[[[[[[[total=%v, calcu=%v]]]]]]]]]]]]]]]", ep.cnt, ep.times)
	for i := 0; i < len(ep.pool); i++ {
		epp := ep.pool[i]
		if epp != nil && len(epp.pool) != 0 {
			logger.Logger.Tracef("[key=%v] [count=%v]", i, len(epp.pool))
		}
	}
}

func _start_produce_routine(ep *ElementProducer) {
	defer utils.DumpStackIfPanic("_start_produce_routine")
	var idle = 0
	for !ep.quit {
		val, key := ep.nf()
		atomic.AddInt64(&ep.times, 1)
		if key >= 0 && key < int32(len(ep.pool)) {
			epp := ep.pool[key]
			if epp == nil {
				epp = &ElementPool{key: key, pool: make(chan interface{}, 128)}
				ep.pool[key] = epp
			}
			select {
			case epp.pool <- val:
				atomic.AddInt32(&ep.cnt, 1)
			case <-time.After(time.Millisecond):
				idle++
			}
		}
		if idle >= 10000 {
			idle = 0
			//ep.Dump()
		}
	}
}

func (this *ElementProducerMgr) GetProducer(id int32) *ElementProducer {
	if p, exist := this.producer[id]; exist {
		return p
	}
	return nil
}

func (this *ElementProducerMgr) SetProducer(id int32, p *ElementProducer) {
	if op, exist := this.producer[id]; exist {
		if op != nil {
			op.Stop()
		}
	}
	this.producer[id] = p
}

func (this *ElementProducerMgr) StopAll() {
	for _, p := range this.producer {
		if p != nil {
			p.Stop()
		}
	}
}
