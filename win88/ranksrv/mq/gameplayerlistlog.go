package mq

import (
	"encoding/json"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/mq"
	"games.yol.com/win88/ranksrv/rank"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/broker"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
)

var _GameIdFilter []int

//消费玩家游戏记录，统计玩家的总投入、总产出等数据，对玩家进行排序
func init() {
	mq.RegisteSubscriber(model.GamePlayerListLogCollName, func(e broker.Event) (err error) {
		msg := e.Message()
		if msg != nil {
			defer func() {
				if err != nil {
					mq.BackUp(e, err)
				}

				e.Ack()

				recover()
			}()

			var log model.GamePlayerListLog
			err = json.Unmarshal(msg.Body, &log)
			if err != nil {
				return
			}

			//过滤掉<=0的记录
			if log.TotalOut <= 0 {
				return
			}

			//过滤掉不需要排行榜的游戏
			if len(_GameIdFilter) > 0 && !common.InSliceInt(_GameIdFilter, int(log.GameId)) {
				return
			}

			//通知主线程执行后续操作
			core.CoreObject().SendCommand(basic.CommandWrapper(func(o *basic.Object) error {
				//Todo 排行榜管理器对数据进行管理
				rank.RankMgrSignton.UpsertGameRankData(log.Platform, log.GameFreeid, log.TotalOut, log.SnId, log.Name)
				return nil
			}), true)

			return
		}
		return nil
	}, broker.Queue(model.GamePlayerListLogCollName+"_rank"), broker.DisableAutoAck(), rabbitmq.DurableQueue())

	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		_GameIdFilter = common.CustomConfig.GetInts("GameIdFilter")
		return nil
	})
}
