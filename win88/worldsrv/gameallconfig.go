package main

//import (
//	"games.yol.com/win88/model"
//	"games.yol.com/win88/srvdata"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/globalsign/mgo/bson"
//)
//
//func InitGameAllConfigData() error {
//	c := model.GameConfigCollection()
//	if c != nil {
//		var datas []model.GameGlobalState
//		err := c.Find(nil).All(&datas)
//		if err != nil {
//			logger.Logger.Trace("InitGameAllConfigData err:", err)
//			return err
//		}
//		for i := 0; i < len(datas); i++ {
//			model.GameAllConfig[datas[i].LogicId] = &datas[i]
//		}
//		logger.Logger.Trace("InitGameAllConfigData:", model.GameAllConfig)
//
//		//把dbFree中的数据写入数据库
//		arr := srvdata.PBDB_GameFreeMgr.Datas.GetArr()
//		for _, dbGame := range arr {
//			if dbGame.GetGameId() > 0 {
//				name := dbGame.GetName()
//				if name != dbGame.GetTitle() {
//					name = dbGame.GetName() + dbGame.GetTitle()
//				}
//				if data, exist := model.GameAllConfig[dbGame.GetId()]; exist {
//					data.Name = name
//					data.GameId = dbGame.GetGameId()
//					data.GameMode = dbGame.GetGameMode()
//					cu := model.GameConfigCollection()
//					if cu != nil {
//						info, err := cu.Upsert(bson.M{"logicid": dbGame.GetId()}, data)
//						if err != nil {
//							logger.Logger.Trace("InitGameAllConfigData :", info, err)
//							return err
//						}
//					}
//				} else {
//					name := dbGame.GetName()
//					if name != dbGame.GetTitle() {
//						name = dbGame.GetName() + dbGame.GetTitle()
//					}
//					data := &model.GameGlobalState{
//						Id:       bson.NewObjectId(),
//						LogicId:  dbGame.GetId(),
//						Name:     name,
//						GameId:   dbGame.GetGameId(),
//						GameMode: dbGame.GetGameMode(),
//						State:    0,
//					}
//					model.GameAllConfig[dbGame.GetId()] = data
//					ci := model.GameConfigCollection()
//					if ci != nil {
//						info, err := ci.Upsert(bson.M{"logicid": dbGame.GetId()}, data)
//						if err != nil {
//							logger.Logger.Trace("InitGameAllConfigData :", info, err)
//							return err
//						}
//					}
//				}
//			}
//		}
//
//		//dbfree表中删除后操作
//		for k, _ := range model.GameAllConfig {
//			gc := srvdata.PBDB_GameFreeMgr.GetData(k)
//			if gc == nil {
//				cgc := model.GameConfigCollection()
//				if cgc != nil {
//					err := cgc.Remove(bson.M{"logicid": k})
//					if err != nil {
//						logger.Logger.Warn("RemoveGameConfig error:", err)
//						return err
//					} else {
//						delete(model.GameAllConfig, k)
//					}
//				}
//			}
//		}
//	}
//	return nil
//}
