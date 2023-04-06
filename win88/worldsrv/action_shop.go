package main

import (
	"time"

	"games.yol.com/win88/proto"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/shop"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSShopInfoPacketFactory struct {
}

type CSShopInfoHandler struct {
}

func (this *CSShopInfoPacketFactory) CreatePacket() interface{} {
	pack := &shop.CSShopInfo{}
	return pack
}

func (this *CSShopInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSShopInfoHandler Process recv ", data)
	if msg, ok := data.(*shop.CSShopInfo); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSShopInfoHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		shopInfos := ShopMgrSington.GetShopInfos(p, int(msg.NowLocation-1))
		pack := &shop.SCShopInfo{
			Infos: shopInfos,
		}
		logger.Logger.Trace("SCShopInfo:", pack, len(shopInfos))
		p.SendToClient(int(shop.SPacketID_PACKET_SC_SHOP_INFO), pack)
	}

	return nil
}

type CSAdLookedPacketFactory struct {
}

type CSAdLookedHandler struct {
}

func (this *CSAdLookedPacketFactory) CreatePacket() interface{} {
	pack := &shop.CSAdLooked{}
	return pack
}

func (this *CSAdLookedHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSAdLookedHandler Process recv ", data)
	if msg, ok := data.(*shop.CSAdLooked); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSAdLookedHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		shopInfo := ShopMgrSington.GetShopInfo(msg.ShopId, p)
		if shopInfo.Ad == 0 {
			logger.Logger.Warnf("CSAdLookedHandler shopId(%v) Ad(%v)", msg.ShopId, shopInfo.Ad == 0)
			return nil
		}
		if shopInfo.RemainingTime > 0 {
			logger.Logger.Warn("冷却中ing RemainingTime", shopInfo.RemainingTime)
			return nil
		}
		if shopTotal, ok1 := p.ShopTotal[msg.ShopId]; ok1 {
			if shopInfo.RepeatTimes <= shopTotal.AdReceiveNum {
				logger.Logger.Warn("次数已满 已经领取的次数AdReceiveNum", shopInfo.AdReceiveNum)
				return nil
			}
			shopTotal.AdLookedNum++
		} else {
			p.ShopTotal[msg.ShopId] = &model.ShopTotal{AdLookedNum: 1}
		}
		b := ShopMgrSington.ReceiveVCPayShop(msg.ShopId, p)
		pack := &shop.SCAdLooked{
			RetCode: shop.OpResultCode_OPRC_Sucess,
		}
		if !b {
			pack.RetCode = shop.OpResultCode_OPRC_Error
		} else {
			p.ShopLastLookTime[msg.ShopId] = time.Now().Unix()
		}
		si := ShopMgrSington.GetShopInfo(msg.ShopId, p)
		if si != nil {
			pack.ShopInfo = &shop.ShopInfo{
				Id:           si.Id,
				AdLookedNum:  si.AdLookedNum,
				AdReceiveNum: si.AdReceiveNum,
				LastLookTime: si.LastLookTime,
				RoleAdded:    si.RoleAdded,
				PetAdded:     si.PetAdded,
			}
			if shopInfo.Type == Type_ShopCoin {
				pack.ShopInfo.ItemId = si.ItemId
				pack.ShopInfo.Order = si.Order
				pack.ShopInfo.Page = si.Page
				pack.ShopInfo.Type = si.Type
				pack.ShopInfo.Location = si.Location
				pack.ShopInfo.Picture = si.Picture
				pack.ShopInfo.Name = si.Name
				pack.ShopInfo.Ad = si.Ad
				pack.ShopInfo.AdTime = si.AdTime
				pack.ShopInfo.RepeatTimes = si.RepeatTimes
				pack.ShopInfo.CoolingTime = si.CoolingTime
				pack.ShopInfo.Label = si.Label
				pack.ShopInfo.Added = si.Added
				pack.ShopInfo.Amount = si.Amount
				pack.ShopInfo.Consume = si.Consume
				pack.ShopInfo.ConsumptionAmount = si.ConsumptionAmount

			}
		}
		logger.Logger.Trace("ShopTotal:", p.ShopTotal[msg.ShopId])
		logger.Logger.Trace("SCAdLooked:", pack)
		p.SendToClient(int(shop.SPacketID_PACKET_SC_SHOP_ADLOOKED), pack)

		ShopMgrSington.ShopCheckShowRed(p)
	}
	return nil
}

type CSVCPayShopPacketFactory struct {
}

type CSVCPayShopHandler struct {
}

func (this *CSVCPayShopPacketFactory) CreatePacket() interface{} {
	pack := &shop.CSVCPayShop{}
	return pack
}

func (this *CSVCPayShopHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSVCPayShopHandler Process recv ", data)
	if msg, ok := data.(*shop.CSVCPayShop); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSVCPayShopHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		shopInfo := ShopMgrSington.GetShopInfo(msg.ShopId, p)

		var lastLookTime int64
		if shopInfo.Ad > 0 {
			lastLookTime = p.ShopLastLookTime[msg.ShopId]
			adLookedNum := p.ShopTotal[msg.ShopId].AdLookedNum
			coolingTime := int64(shopInfo.CoolingTime[adLookedNum])
			if coolingTime != 0 && time.Now().Unix()-lastLookTime < coolingTime {
				logger.Logger.Error("时间差:", time.Now().Unix()-lastLookTime, int64(shopInfo.CoolingTime[adLookedNum]))
				return nil
			}
		}
		//领取VC
		b := ShopMgrSington.ReceiveVCPayShop(msg.ShopId, p)
		logger.Logger.Trace("ShopTotal:", p.ShopTotal[msg.ShopId])
		pack := &shop.SCVCPayShop{
			RetCode: shop.OpResultCode_OPRC_Sucess,
		}
		si := ShopMgrSington.GetShopInfo(msg.ShopId, p)
		if si != nil {
			pack.ShopInfo = &shop.ShopInfo{
				Id:           si.Id,
				AdLookedNum:  si.AdLookedNum,
				AdReceiveNum: si.AdReceiveNum,
				LastLookTime: si.LastLookTime,
				RoleAdded:    si.RoleAdded,
				PetAdded:     si.PetAdded,
			}
			if shopInfo.Type == Type_ShopCoin {
				pack.ShopInfo.ItemId = si.ItemId
				pack.ShopInfo.Order = si.Order
				pack.ShopInfo.Page = si.Page
				pack.ShopInfo.Type = si.Type
				pack.ShopInfo.Location = si.Location
				pack.ShopInfo.Picture = si.Picture
				pack.ShopInfo.Name = si.Name
				pack.ShopInfo.Ad = si.Ad
				pack.ShopInfo.AdTime = si.AdTime
				pack.ShopInfo.RepeatTimes = si.RepeatTimes
				pack.ShopInfo.CoolingTime = si.CoolingTime
				pack.ShopInfo.Label = si.Label
				pack.ShopInfo.Added = si.Added
				pack.ShopInfo.Amount = si.Amount
				pack.ShopInfo.Consume = si.Consume
				pack.ShopInfo.ConsumptionAmount = si.ConsumptionAmount

			}
		}
		if !b {
			pack.RetCode = shop.OpResultCode_OPRC_Error
		}
		logger.Logger.Trace("SCVCPayShop:", pack)
		proto.SetDefaults(pack)
		p.SendToClient(int(shop.SPacketID_PACKET_SC_SHOP_VCPAYSHOP), pack)
	}

	return nil
}

type CSShopExchangeRecordPacketFactory struct {
}

type CSShopExchangeRecordHandler struct {
}

func (this *CSShopExchangeRecordPacketFactory) CreatePacket() interface{} {
	pack := &shop.CSShopExchangeRecord{}
	return pack
}

func (this *CSShopExchangeRecordHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSShopExchangeRecordHandler Process recv ", data)
	if msg, ok := data.(*shop.CSShopExchangeRecord); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSShopExchangeRecordHandler p == nil")
			return nil
		}
		platform := p.GetPlatform()
		if platform == nil {
			return nil
		}
		ShopMgrSington.GetExchangeRecord(p, msg.PageNo)

	}
	return nil
}

// 兑换v zzz
type CSShopExchangePacketFactory struct {
}

type CSShopExchangeHandler struct {
}

func (this *CSShopExchangePacketFactory) CreatePacket() interface{} {
	pack := &shop.CSShopExchange{}
	return pack
}

func (this *CSShopExchangeHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSShopExchangeHandler Process recv ", data)
	if msg, ok := data.(*shop.CSShopExchange); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSShopExchangeHandler p == nil")
			return nil
		}

		ShopMgrSington.Exchange(p, msg.GoodsId, msg.UserName, msg.Mobile, msg.Comment)
	}
	return nil
}

// 兑换列表
type CSShopExchangeListPacketFactory struct {
}

type CSShopExchangeListHandler struct {
}

func (this *CSShopExchangeListPacketFactory) CreatePacket() interface{} {
	pack := &shop.CSShopExchangeList{}
	return pack
}

func (this *CSShopExchangeListHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSShopExchangeListHandler Process recv ", data)
	if _, ok := data.(*shop.CSShopExchangeList); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSShopExchangeListHandler p == nil")
			return nil
		}

		ShopMgrSington.ExchangeList(p)
	}
	return nil
}
func init() {
	common.RegisterHandler(int(shop.SPacketID_PACKET_CS_SHOP_INFO), &CSShopInfoHandler{})
	netlib.RegisterFactory(int(shop.SPacketID_PACKET_CS_SHOP_INFO), &CSShopInfoPacketFactory{})
	common.RegisterHandler(int(shop.SPacketID_PACKET_CS_SHOP_ADLOOKED), &CSAdLookedHandler{})
	netlib.RegisterFactory(int(shop.SPacketID_PACKET_CS_SHOP_ADLOOKED), &CSAdLookedPacketFactory{})
	common.RegisterHandler(int(shop.SPacketID_PACKET_CS_SHOP_VCPAYSHOP), &CSVCPayShopHandler{})
	netlib.RegisterFactory(int(shop.SPacketID_PACKET_CS_SHOP_VCPAYSHOP), &CSVCPayShopPacketFactory{})
	common.RegisterHandler(int(shop.SPacketID_PACKET_CS_SHOP_EXCHANGERECORD), &CSShopExchangeRecordHandler{})
	netlib.RegisterFactory(int(shop.SPacketID_PACKET_CS_SHOP_EXCHANGERECORD), &CSShopExchangeRecordPacketFactory{})
	common.RegisterHandler(int(shop.SPacketID_PACKET_CS_SHOP_EXCHANGE), &CSShopExchangeHandler{})
	netlib.RegisterFactory(int(shop.SPacketID_PACKET_CS_SHOP_EXCHANGE), &CSShopExchangePacketFactory{})
	common.RegisterHandler(int(shop.SPacketID_PACKET_CS_SHOP_EXCHANGELIST), &CSShopExchangeListHandler{})
	netlib.RegisterFactory(int(shop.SPacketID_PACKET_CS_SHOP_EXCHANGELIST), &CSShopExchangeListPacketFactory{})
}
