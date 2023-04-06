package main

//转移到schedulesrv中专门处理
//import (
//	"time"

//	"games.yol.com/win88/model"

//	"github.com/idealeak/goserver/core"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/idealeak/goserver/core/schedule"
//)

//func init() {
//	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
//		//每日凌晨4点执行清理任务
//		//0 0 4 * * *
//		task := schedule.NewTask("定期清理cardlog", "0 0 4 * * *", func() error {
//			logger.Logger.Info("@@@执行定期清理任务@@@")
//			tNow := time.Now()
//			m := model.GameParamData.LogHoldDuration
//			if m <= 0 {
//				m = 1
//			}
//			tStart := tNow.AddDate(0, int(-m), 0)
//			changeInfo, err := model.RemoveCoinLog(tStart.Unix())
//			if err != nil {
//				logger.Logger.Warnf("定期清理coinlog失败: %v", err)
//			} else {
//				logger.Logger.Warnf("定期清理coinlog成功: updated:%v removed:%v", changeInfo.Updated, changeInfo.Removed)
//			}
//			//1个月前的回放记录
//			changeInfo, err = model.RemoveGameRecs(tStart)
//			if err != nil {
//				logger.Logger.Warnf("定期清理gamerec失败: %v", err)
//			} else {
//				logger.Logger.Warnf("定期清理gamerec成功: updated:%v removed:%v", changeInfo.Updated, changeInfo.Removed)
//			}
//			//APIlog
//			changeInfo, err = model.RemoveAPILog(tStart.Unix())
//			if err != nil {
//				logger.Logger.Warnf("定期清理APIlog失败: %v", err)
//			} else {
//				logger.Logger.Warnf("定期清理APIlog成功: updated:%v removed:%v", changeInfo.Updated, changeInfo.Removed)
//			}

//			//10天前的数据
//			tStart = tNow.AddDate(0, 0, -10)
//			changeInfo, err = model.RemoveAgentGameRecs(tStart.Unix())
//			if err != nil {
//				logger.Logger.Warnf("定期清理user_agentgamerec失败: %v", err)
//			} else {
//				logger.Logger.Warnf("定期清理user_agentgamerec成功: updated:%v removed:%v", changeInfo.Updated, changeInfo.Removed)
//			}

//			//3天前的数据
//			tStart = tNow.AddDate(0, 0, -3)
//			changeInfo, err = model.RemoveSceneCoinLog(tStart.Unix())
//			if err != nil {
//				logger.Logger.Warnf("定期清理scenecoinlog失败: %v", err)
//			} else {
//				logger.Logger.Warnf("定期清理scenecoinlog成功: updated:%v removed:%v", changeInfo.Updated, changeInfo.Removed)
//			}

//			//7天前的数据
//			tStart = tNow.AddDate(0, 0, -7)
//			changeInfo, err = model.RemoveGameCoinLog(tStart.Unix())
//			if err != nil {
//				logger.Logger.Warnf("定期清理gamecoinlog失败: %v", err)
//			} else {
//				logger.Logger.Warnf("定期清理gamecoinlog成功: updated:%v removed:%v", changeInfo.Updated, changeInfo.Removed)
//			}
//			//游戏记录
//			changeInfo, err = model.RemoveGameLog(tStart.Unix())
//			if err != nil {
//				logger.Logger.Warnf("定期清理游戏记录失败: %v", err)
//			} else {
//				logger.Logger.Warnf("定期清理游戏记录成功: updated:%v removed:%v", changeInfo.Updated, changeInfo.Removed)
//			}
//			return nil
//		})
//		if task != nil {
//			logger.Logger.Info("@@@执行定期清理任务@@@加入调度")
//			schedule.AddTask(task.Taskname, task)
//		}

//		//每日凌晨执行汇总任务
//		//0 0 0 * * *
//		taskDay := schedule.NewTask("定期汇总玩家金币总额", "0 0 0 * * *", func() error {
//			logger.Logger.Info("@@@执行定期汇总任务@@@")
//			return model.AggregatePlayerCoin()
//		})
//		if taskDay != nil {
//			logger.Logger.Info("@@@执行定期汇总任务@@@加入调度")
//			schedule.AddTask(taskDay.Taskname, taskDay)
//		}
//		return nil
//	})
//}
