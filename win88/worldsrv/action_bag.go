package main

import (
	"fmt"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/bag"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

// 查看背包
type CSBagInfoPacketFactory struct {
}

type CSBagInfoHandler struct {
}

func (this *CSBagInfoPacketFactory) CreatePacket() interface{} {
	pack := &bag.CSBagInfo{}
	return pack
}

func (this *CSBagInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSBagInfoHandler Process recv ", data)
	if _, ok := data.(*bag.CSBagInfo); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSBagInfoHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		//nowLocation := int(msg.NowLocation - 1)
		playbag := BagMgrSington.GetBagInfo(p.SnId)
		pack := &bag.SCBagInfo{RetCode: bag.OpResultCode_OPRC_Sucess, BagNumMax: BagItemMax}
		for _, v := range playbag.BagItem {
			item := srvdata.PBDB_GameItemMgr.GetData(v.ItemId)
			if item != nil && v.ItemNum > 0 /*&& (nowLocation == -1 || (nowLocation < len(item.ShowLocation) && item.ShowLocation[nowLocation] == 1))*/ {
				pack.Infos = append(pack.Infos, &bag.ItemInfo{
					ItemId:  v.ItemId,
					ItemNum: v.ItemNum,
					//Name:           item.Name,
					//ShowLocation:   item.ShowLocation,
					//Classify:       item.Classify,
					//Type:           item.Type,
					//Effect0:        item.Effect0,
					//Effect:         item.Effect,
					//SaleType:       item.SaleType,
					//SaleGold:       item.SaleGold,
					//Composition:    item.Composition,
					//CompositionMax: item.CompositionMax,
					//Time:           item.Time,
					//Location:       item.Location,
					//Describe:       item.Describe,
					ObtainTime: v.ObtainTime,
				})
			}
		}
		logger.Logger.Trace("SCBagInfo:", pack)
		p.SendToClient(int(bag.SPacketID_PACKET_ALL_BAG_INFO), pack)
	}
	return nil
}

// 使用/获取道具 PS严格来说客户端只存在消耗道具 服务器增加 上线要禁止该接口增加道具
type CSUpBagInfoPacketFactory struct {
}

type CSUpBagInfoHandler struct {
}

func (this *CSUpBagInfoPacketFactory) CreatePacket() interface{} {
	pack := &bag.CSUpBagInfo{}
	return pack
}

func (this *CSUpBagInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSUpBagInfoHandler Process recv ", data)
	if msg, ok := data.(*bag.CSUpBagInfo); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSUpBagInfoHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		pack := &bag.SCUpBagInfo{RetCode: bag.OpResultCode_OPRC_Error}
		item := BagMgrSington.GetBagItemById(p.SnId, msg.ItemId)
		if item == nil || msg.ItemNum <= 0 || item.ItemNum < msg.ItemNum || len(item.Effect) != ItemMax || p.SnId == msg.AcceptSnId {
			logger.Logger.Trace("SCUpBagInfo:", pack)
			p.SendToClient(int(bag.SPacketID_PACKET_ALL_BAG_INFO), pack)
			return nil
		}
		var isCanOp int32
		if msg.NowEffect == 0 {
			//竖版
			isCanOp = item.Effect0[msg.Opt]
		} else if msg.NowEffect == 1 {
			//横版
			isCanOp = item.Effect[msg.Opt]
		}
		if isCanOp == 0 {
			logger.Logger.Trace("道具没有操作权限", msg.ItemId)
			pack.RetCode = bag.OpResultCode_OPRC_Error
		} else {
			switch msg.Opt {
			case ItemCanUse:
				logger.Logger.Trace("道具使用", msg.ItemId)
				pack.RetCode = bag.OpResultCode_OPRC_Error
			case ItemCanGive:
				logger.Logger.Trace("道具赠送", msg.ItemId)
				pack.RetCode = bag.OpResultCode_OPRC_Error
				acceptPlayer := PlayerMgrSington.GetPlayerBySnId(msg.AcceptSnId)
				if acceptPlayer != nil {
					BagMgrSington.AddMailByItem(p.Platform, p.SnId, p.Name, msg.AcceptSnId, msg.ShowId, []int32{msg.ItemId, msg.ItemNum})
					item.ItemNum -= msg.ItemNum
					//if item.ItemId == VCard {
					//	p.SendVCoinDiffData()
					//}
					pack.RetCode = bag.OpResultCode_OPRC_Sucess
					pack.NowItemId = msg.ItemId
					pack.NowItemNum = item.ItemNum
					remark := fmt.Sprintf("赠送给玩家(%v)", msg.AcceptSnId)
					BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemConsume, item.ItemId, item.Name, msg.ItemNum, remark)
					logger.Logger.Trace("道具赠送成功", msg.ItemId)
				} else {
					task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
						return model.GetPlayerBaseInfo(p.Platform, msg.AcceptSnId)
					}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
						aPlayer := data.(*model.PlayerBaseInfo)
						if data != nil && aPlayer != nil {
							BagMgrSington.AddMailByItem(p.Platform, p.SnId, p.Name, msg.AcceptSnId, msg.ShowId, []int32{msg.ItemId, msg.ItemNum})
							item.ItemNum -= msg.ItemNum
							//if item.ItemId == VCard {
							//	p.SendVCoinDiffData()
							//}
							pack.RetCode = bag.OpResultCode_OPRC_Sucess
							pack.NowItemId = msg.ItemId
							pack.NowItemNum = item.ItemNum
							remark := fmt.Sprintf("赠送给玩家(%v)", msg.AcceptSnId)
							BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemConsume, item.ItemId, item.Name, msg.ItemNum, remark)
							logger.Logger.Trace("道具赠送成功", msg.ItemId)
						} else {
							pack.RetCode = bag.OpResultCode_OPRC_NotPlayer
						}
						BagMgrSington.SyncBagData(p, item.ItemId)
						logger.Logger.Trace("SCUpBagInfo:", pack)
						p.SendToClient(int(bag.SPacketID_PACKET_ALL_BAG_USE), pack)
					}), "GetPlayerBaseInfo").Start()
					return nil
				}
			case ItemCanSell:
				logger.Logger.Trace("道具出售", msg.ItemId)
				if msg.ItemNum <= 0 {
					pack.RetCode = bag.OpResultCode_OPRC_Error
				} else {
					isF := BagMgrSington.SaleItem(p, msg.ItemId, msg.ItemNum)
					if isF {
						pack.RetCode = bag.OpResultCode_OPRC_Sucess
						if item.SaleGold > 0 {
							if item.SaleType == 1 {
								remark := "道具出售" + fmt.Sprintf("%v-%v", msg.ItemId, msg.ItemNum)
								p.AddCoin(int64(item.SaleGold*msg.ItemNum), common.GainWay_Item_Sale, "sys", remark)
								pack.Coin = int64(item.SaleGold * msg.ItemNum)
								BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemConsume, item.ItemId, item.Name, msg.ItemNum, "道具出售")
							} else if item.SaleType == 2 {
								remark := "道具出售" + fmt.Sprintf("%v-%v", msg.ItemId, msg.ItemNum)
								p.AddDiamond(int64(item.SaleGold*msg.ItemNum), common.GainWay_Item_Sale, "sys", remark)
								pack.Diamond = int64(item.SaleGold * msg.ItemNum)
								BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemConsume, item.ItemId, item.Name, msg.ItemNum, "道具出售")
							}
						}
						pack.NowItemId = item.ItemId
						pack.NowItemNum = item.ItemNum
					}
					//if item.ItemId == VCard {
					//	p.SendVCoinDiffData()
					//}
				}
			}
		}
		BagMgrSington.SyncBagData(p, item.ItemId)
		logger.Logger.Trace("SCUpBagInfo:", pack)
		p.SendToClient(int(bag.SPacketID_PACKET_ALL_BAG_USE), pack)
	}
	return nil
}

func init() {
	common.RegisterHandler(int(bag.SPacketID_PACKET_ALL_BAG_INFO), &CSBagInfoHandler{})
	netlib.RegisterFactory(int(bag.SPacketID_PACKET_ALL_BAG_INFO), &CSBagInfoPacketFactory{})
	common.RegisterHandler(int(bag.SPacketID_PACKET_ALL_BAG_USE), &CSUpBagInfoHandler{})
	netlib.RegisterFactory(int(bag.SPacketID_PACKET_ALL_BAG_USE), &CSUpBagInfoPacketFactory{})
}
