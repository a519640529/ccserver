package main

import (
	"fmt"
	"math"
	"strconv"
	"time"

	hall_proto "games.yol.com/win88/protocol/gamehall"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/shop"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"games.yol.com/win88/srvdata"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
)

// 消费类型
const (
	Shop_Consume_Coin = iota + 1
	Shop_Consume_Diamond
	Shop_Consume_Dollar
)

// 商品类型
const (
	Shop_Type_Coin = iota + 1
	Shop_Type_Diamond
	Shop_Type_Item
)

// 兑换商品状态
const (
	Shop_Status_Keep    = iota //0为待审核
	Shop_Status_Pass           // 1为审核通过
	Shop_Status_Send           // 2为已发货
	Shop_Status_NotSend        // 3为审核不通过
	Shop_Status_Revoke         // 4为撤单
)
const (
	VCard int32 = 30001
)

/*
1.兑换成功：兑换成功，请等待审核
2.今日兑换已达上限，请明日再来
3.该物品已被兑换完
*/
const (
	Err_ExShopNEnough string = "ExShopNotEnough" // 该物品已被兑换完
	Err_ExShopLimit   string = "ExShopIsLimit"   // 今日兑换已达上限，请明日再来
	Err_ExShopData    string = "ExShopDataErr"   // 收件人地址有误
)

const (
	Type_ShopCoin    int32 = iota + 1 // 1，金币
	Type_ShopDiamond                  // 2，钻石
	Type_ShopItem1                    // 3，道具类型1：用金币或者钻石购买
	Type_ShopItem2                    // 4.道具类型2：走充值购买
	Type_ShopOther                    // 5，其他
)

var ShopMgrSington = &ShopMgr{
	Shop: make(map[int32]*ShopInfo),
}

type ShopMgr struct {
	Shop    map[int32]*ShopInfo // etcd商品
	ExShops []*ExchangeShopInfo //兑换商品 使用切片因为商品并不多 遍历没压力 目前没有区分平台
}

type ShopInfo struct {
	Id                int32   //商品ID
	ItemId            int32   //道具ID
	Page              int32   //页面 1，金币页面 2，钻石页面 3，道具页面
	Order             int32   //排序  页面内商品的位置排序
	Type              int32   // 类型 1，金币 2，钻石 3，道具类型1：用金币或者钻石购买 4.道具类型2：走充值购买 5，其他
	Location          []int32 // 显示位置 第1位，竖版大厅 第2位，Tienlen1级选场 第3位，捕鱼1级选场
	Picture           string  // 图片id
	Name              string  // 名称
	Ad                int32   //是否观看广告 1，是 2，不是
	AdTime            int32   // 观看几次广告
	RepeatTimes       int32   // 领取次数
	CoolingTime       []int32 // 观看冷却时间
	Label             []int32 // 标签
	Added             int32   // 加送百分比
	Amount            int32   // 货币金额
	Consume           int32   // 购买消耗类型 1，金币 2，钻石 3，美金 4，柬埔寨币
	ConsumptionAmount int32   // 消耗数量 加送百分比（比如加送10%，就配置110）
	//缓存数据
	AdLookedNum   int32 //已经观看的次数
	AdReceiveNum  int32 //已经领取的次数
	RemainingTime int32
	LastLookTime  int32
	RoleAdded     int32
	PetAdded      int32
}

type ExchangeShopInfo struct {
	Id      int32  //商品ID
	Picture string // 图片
	Type    int32  // 类型 1，话费2，实物
	Name    string // 名称
	NeedNum int32  // 消耗V卡
	Rule    string //规则说明
}

func (this *ShopMgr) ModuleName() string {
	return "ShopMgr"
}

func (this *ShopMgr) Init() {

	if !model.GameParamData.UseEtcd {
		// 后台说现在没有不走ETCD情况~
	} else {

		EtcdMgrSington.InitExchangeShop()

		EtcdMgrSington.InitItemShop()
	}
}
func (this *ShopMgr) InitItemShop(cfgs *webapi_proto.ItemShopList) {

	for _, cfg := range cfgs.List {
		this.Shop[cfg.Id] = &ShopInfo{
			Id:                cfg.Id,
			ItemId:            cfg.ItemId,
			Page:              cfg.Page,
			Order:             cfg.Order,
			Type:              cfg.Type,
			Location:          cfg.Location,
			Picture:           cfg.Picture,
			Name:              cfg.Name,
			Ad:                cfg.Ad,
			AdTime:            cfg.AdTime,
			RepeatTimes:       cfg.RepeatTimes,
			CoolingTime:       cfg.CoolingTime,
			Label:             cfg.Label,
			Added:             cfg.Added,
			Amount:            cfg.Amount,
			Consume:           cfg.Consume,
			ConsumptionAmount: cfg.ConsumptionAmount,
		}

	}

}
func (this *ShopMgr) UpItemShop(cfgs *webapi_proto.ItemShopList) {

	for _, cfg := range cfgs.List {

		this.Shop[cfg.Id] = &ShopInfo{
			Id:                cfg.Id,
			ItemId:            cfg.ItemId,
			Page:              cfg.Page,
			Order:             cfg.Order,
			Type:              cfg.Type,
			Location:          cfg.Location,
			Picture:           cfg.Picture,
			Name:              cfg.Name,
			Ad:                cfg.Ad,
			AdTime:            cfg.AdTime,
			RepeatTimes:       cfg.RepeatTimes,
			CoolingTime:       cfg.CoolingTime,
			Label:             cfg.Label,
			Added:             cfg.Added,
			Amount:            cfg.Amount,
			Consume:           cfg.Consume,
			ConsumptionAmount: cfg.ConsumptionAmount,
		}

	}

}

func (this *ShopMgr) UpExShop(cfgs *webapi_proto.ExchangeShopList) {

	this.ExShops = this.ExShops[:0]
	for _, cfg := range cfgs.List {
		/*for _, v := range this.ExShops{
			if v.Id == cfg.Id{

				break
			}
		}*/
		this.ExShops = append(this.ExShops, &ExchangeShopInfo{
			Id:      cfg.Id,
			Picture: cfg.Picture,
			Type:    cfg.Type,
			Name:    cfg.Name,
			NeedNum: cfg.Price,
			Rule:    cfg.Content,
		})
	}
	// logger.Logger.Infof("", this.ExShops)
}

func convertShopInfo(shopId int32) *ShopInfo {
	if cfg := srvdata.PBDB_Shop1Mgr.GetData(shopId); cfg != nil {
		return &ShopInfo{
			Id:                cfg.Id,
			ItemId:            cfg.ItemId,
			Page:              cfg.Page,
			Order:             cfg.Order,
			Type:              cfg.Type,
			Location:          cfg.Location,
			Picture:           cfg.Picture,
			Name:              cfg.Name,
			Ad:                cfg.Ad,
			AdTime:            cfg.AdTime,
			RepeatTimes:       cfg.RepeatTimes,
			CoolingTime:       cfg.CoolingTime,
			Label:             cfg.Label,
			Added:             cfg.Added,
			Amount:            cfg.Amount,
			Consume:           cfg.Consume,
			ConsumptionAmount: cfg.ConsumptionAmount,
		}
	}
	return nil
}

func (this *ShopMgr) getShopInfos() []*ShopInfo {
	// 加载配置表
	shops := make([]*ShopInfo, 0)
	for _, cfg := range srvdata.PBDB_Shop1Mgr.Datas.Arr {
		if cfg.Type == Type_ShopCoin {
			continue
		}
		shops = append(shops, &ShopInfo{
			Id:                cfg.Id,
			ItemId:            cfg.ItemId,
			Page:              cfg.Page,
			Order:             cfg.Order,
			Type:              cfg.Type,
			Location:          cfg.Location,
			Picture:           cfg.Picture,
			Name:              cfg.Name,
			Ad:                cfg.Ad,
			AdTime:            cfg.AdTime,
			RepeatTimes:       cfg.RepeatTimes,
			CoolingTime:       cfg.CoolingTime,
			Label:             cfg.Label,
			Added:             cfg.Added,
			Amount:            cfg.Amount,
			Consume:           cfg.Consume,
			ConsumptionAmount: cfg.ConsumptionAmount,
		})

	}
	for _, cfg := range this.Shop { // 深拷贝
		shops = append(shops, &ShopInfo{
			Id:                cfg.Id,
			ItemId:            cfg.ItemId,
			Page:              cfg.Page,
			Order:             cfg.Order,
			Type:              cfg.Type,
			Location:          cfg.Location,
			Picture:           cfg.Picture,
			Name:              cfg.Name,
			Ad:                cfg.Ad,
			AdTime:            cfg.AdTime,
			RepeatTimes:       cfg.RepeatTimes,
			CoolingTime:       cfg.CoolingTime,
			Label:             cfg.Label,
			Added:             cfg.Added,
			Amount:            cfg.Amount,
			Consume:           cfg.Consume,
			ConsumptionAmount: cfg.ConsumptionAmount,
		})

	}
	return shops
}

func (this *ShopMgr) GetShopInfo(shopId int32, p *Player) *ShopInfo {
	if shopId == 0 {
		return nil
	}
	shopInfo := this.Shop[shopId] // srvdata.PBDB_Shop1Mgr.GetData(shopId)
	if shopInfo == nil {          // 配置表
		shopInfo = convertShopInfo(shopId)
	}
	if shopInfo != nil {
		var AdLookedNum, AdReceiveNum, remainingTime int32
		var lastLookTime int64
		if shopInfo.Ad > 0 {
			shopTotal := p.ShopTotal[shopInfo.Id]
			if shopTotal != nil {
				AdLookedNum = shopTotal.AdLookedNum
				AdReceiveNum = shopTotal.AdReceiveNum
			}
			lastLookTime = p.ShopLastLookTime[shopInfo.Id]
			if AdLookedNum < int32(len(shopInfo.CoolingTime)) {
				remainingTime = shopInfo.CoolingTime[AdLookedNum]
			}
			if lastLookTime > 0 {
				dif := int32(time.Now().Unix() - lastLookTime)
				if dif >= remainingTime {
					remainingTime = 0
				} else {
					remainingTime = remainingTime - dif
				}
			}
		}
		award := PetMgrSington.GetShopAward(shopInfo, p)

		/*if shopInfo.Type == Type_ShopCoin {
			shop.ItemId = shopInfo.ItemId
			shop.Order = shopInfo.Order
			shop.Page = shopInfo.Page
			shop.Type = shopInfo.Type
			shop.Location = shopInfo.Location
			shop.Picture = shopInfo.Picture
			shop.Name = shopInfo.Name
			shop.Label = shopInfo.Label
			shop.Added = shopInfo.Added
			shop.Amount = shopInfo.Amount
			shop.Consume = shopInfo.Consume
			shop.ConsumptionAmount = shopInfo.ConsumptionAmount
		}*/
		return &ShopInfo{
			Id:                shopInfo.Id,
			Ad:                shopInfo.Ad,
			AdTime:            shopInfo.AdTime,
			RepeatTimes:       shopInfo.RepeatTimes,
			CoolingTime:       shopInfo.CoolingTime,
			AdLookedNum:       AdLookedNum,
			AdReceiveNum:      AdReceiveNum,
			RemainingTime:     remainingTime,
			LastLookTime:      int32(lastLookTime),
			RoleAdded:         award,
			ItemId:            shopInfo.ItemId,
			Order:             shopInfo.Order,
			Page:              shopInfo.Page,
			Type:              shopInfo.Type,
			Location:          shopInfo.Location,
			Picture:           shopInfo.Picture,
			Name:              shopInfo.Name,
			Label:             shopInfo.Label,
			Added:             shopInfo.Added,
			Amount:            shopInfo.Amount,
			Consume:           shopInfo.Consume,
			ConsumptionAmount: shopInfo.ConsumptionAmount,
		}
	}
	return nil
}
func (this *ShopMgr) GetShopInfos(p *Player, nowLocation int) []*shop.ShopInfo {
	if p.ShopTotal == nil {
		p.ShopTotal = make(map[int32]*model.ShopTotal)
	}
	if p.ShopLastLookTime == nil {
		p.ShopLastLookTime = make(map[int32]int64)
	}
	shops := this.getShopInfos() //this.Shop // srvdata.PBDB_Shop1Mgr.Datas.Arr
	var newShops = make([]*shop.ShopInfo, 0)

	for _, shopInfo := range shops {
		if nowLocation == -1 || (nowLocation < len(shopInfo.Location) && shopInfo.Location[nowLocation] == 1) {
			var AdLookedNum, AdReceiveNum, remainingTime int32
			var lastLookTime int64
			if shopInfo.Ad > 0 {
				shopTotal := p.ShopTotal[shopInfo.Id]
				if shopTotal != nil {
					AdLookedNum = shopTotal.AdLookedNum
					AdReceiveNum = shopTotal.AdReceiveNum
				}
				lastLookTime = p.ShopLastLookTime[shopInfo.Id]
				if AdLookedNum < int32(len(shopInfo.CoolingTime)) {
					remainingTime = shopInfo.CoolingTime[AdLookedNum]
				}
				dif := int32(time.Now().Unix() - lastLookTime)
				if dif >= remainingTime {
					remainingTime = 0
				} else {
					remainingTime = remainingTime - dif
				}
			}
			award := PetMgrSington.GetShopAward(shopInfo, p)
			shop := &shop.ShopInfo{
				Id:           shopInfo.Id,
				AdLookedNum:  AdLookedNum,
				AdReceiveNum: AdReceiveNum,
				LastLookTime: int32(lastLookTime),
				RoleAdded:    award,
			}
			if shopInfo.Type == Type_ShopCoin {
				shop.ItemId = shopInfo.ItemId
				shop.Order = shopInfo.Order
				shop.Page = shopInfo.Page
				shop.Type = shopInfo.Type
				shop.Location = shopInfo.Location
				shop.Picture = shopInfo.Picture
				shop.Name = shopInfo.Name
				shop.Ad = shopInfo.Ad
				shop.AdTime = shopInfo.AdTime
				shop.RepeatTimes = shopInfo.RepeatTimes
				shop.CoolingTime = shopInfo.CoolingTime
				shop.Label = shopInfo.Label
				shop.Added = shopInfo.Added
				shop.Amount = shopInfo.Amount
				shop.Consume = shopInfo.Consume
				shop.ConsumptionAmount = shopInfo.ConsumptionAmount

			}
			newShops = append(newShops, shop)
		}
	}

	return newShops
}

// 领取VC
func (this *ShopMgr) ReceiveVCPayShop(shopId int32, p *Player) bool {
	shopInfo := this.Shop[shopId] //srvdata.PBDB_Shop1Mgr.GetData(shopId)
	if shopInfo == nil {          // 配置表
		shopInfo = convertShopInfo(shopId)
	}
	if shopInfo == nil {
		logger.Logger.Errorf("this shop == nil  shopId[%v] ", shopId)
		return false
	}
	//需要观看广告的
	if shopInfo.Ad > 0 {
		if _, ok1 := p.ShopTotal[shopId]; !ok1 {
			p.ShopTotal[shopId] = &model.ShopTotal{}
		}
		shopTotal := p.ShopTotal[shopId]
		if shopTotal.AdReceiveNum < shopInfo.RepeatTimes {
			//产生订单
			if this.CreateVCOrder(shopInfo, p.SnId, p.Platform) {
				shopTotal.AdReceiveNum++
				this.AddOrder(shopInfo, p)
				return true
			}
		}
	} else {
		//不需要观看广告的
		switch shopInfo.Consume {
		case Shop_Consume_Coin:
			//金币
			if int64(shopInfo.ConsumptionAmount) > p.Coin {
				return false
			}
		case Shop_Consume_Diamond:
			//钻石
			if int64(shopInfo.ConsumptionAmount) > p.Diamond {
				return false
			}
		case Shop_Consume_Dollar:

		default:
			return false
		}
		//产生订单
		if this.CreateVCOrder(shopInfo, p.SnId, p.Platform) {
			this.AddOrder(shopInfo, p)
			return true
		}
	}
	return false
}
func (this *ShopMgr) AddOrder(shopInfo *ShopInfo, p *Player) {
	shopName := shopInfo.Name + "|" + strconv.Itoa(int(shopInfo.Id))
	if shopInfo.Ad <= 0 {
		logger.Logger.Tracef("AddOrder Consume[%v],shopName[%v]", shopInfo.Consume, shopName, shopInfo.ConsumptionAmount)
		switch shopInfo.Consume {
		case Shop_Consume_Coin:
			p.AddCoin(int64(-shopInfo.ConsumptionAmount), common.GainWay_Shop_Buy, "sys", shopName)
		case Shop_Consume_Diamond:
			p.AddDiamond(int64(-shopInfo.ConsumptionAmount), common.GainWay_Shop_Buy, "sys", shopName)
		}
	}
	//默认加成
	var addTotal = int64(shopInfo.Amount)
	if shopInfo.Added > 0 {
		addTotal = int64(math.Floor((1 + float64(shopInfo.Added)/100) * float64(shopInfo.Amount)))
	}
	logger.Logger.Trace("addTotal  ", addTotal, shopInfo.Amount, shopInfo.Added)
	switch shopInfo.Type {
	case Shop_Type_Coin:
		var addCoin int64
		award := PetMgrSington.GetShopAward(shopInfo, p)
		if award > 0 {
			addCoin = int64(math.Floor(float64(award) / 100 * float64(shopInfo.Amount)))
		}
		//增加金币
		p.AddCoin(addTotal+addCoin, common.GainWay_Shop_Buy, "sys", shopName)
	case Shop_Type_Diamond:
		//增加钻石
		p.AddDiamond(addTotal, common.GainWay_Shop_Buy, "sys", shopName)
	case Shop_Type_Item:
		//增加道具
		item := &Item{ItemId: shopInfo.ItemId, ItemNum: shopInfo.Amount, ObtainTime: time.Now().Unix()}
		BagMgrSington.AddJybBagInfo(p, []*Item{item})
		data := srvdata.PBDB_GameItemMgr.GetData(item.ItemId)
		if data != nil {
			BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemObtain, shopInfo.ItemId, data.Name, shopInfo.Amount, "商城购买")
		}
		PetMgrSington.CheckShowRed(p)
	}
}

// 产生VC订单
func (this *ShopMgr) CreateVCOrder(shopInfo *ShopInfo, snid int32, platform string) (ret bool) {
	ret = true
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		dbShop := model.NewDbShop(platform, shopInfo.Id, shopInfo.Name, snid, 1, 1, "", "", "", "sys")
		return model.InsertDbShopLog(platform, dbShop)
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		if data != nil {
			logger.Logger.Errorf("err:", data.(error))
			ret = false
		}
	}), "CreateVCOrder").Start()
	return
}

func (this *ShopMgr) GetExchangeData(id int32) *ExchangeShopInfo {

	for _, data := range this.ExShops {
		if data.Id == id {
			return data
		}
	}
	return nil
}

// 生成兑换订单
func (this *ShopMgr) Exchange(p *Player, id int32, username, mobile, comment string) (ret bool) {
	ret = true
	cdata := this.GetExchangeData(id)

	pack := &shop.SCShopExchange{
		RetCode: shop.OpResultCode_OPRC_VCoinNotEnough,
	}
	if cdata == nil {
		pack.RetCode = shop.OpResultCode_OPRC_ExchangeSoldOut
		p.SendToClient(int(shop.SPacketID_PACKET_SC_SHOP_EXCHANGE), pack)
		return false
	}

	// TODO 服务器处理 减劵 成功后调后台生成订单
	// 判断p.VCoin是否足够 不足返回错误 足够扣掉 另外需从后台操作回执成功生成扣除V卡的订单 回执失败
	if isF := BagMgrSington.SaleItem(p, VCard, cdata.NeedNum); !isF { // 扣掉V卡
		p.SendToClient(int(shop.SPacketID_PACKET_SC_SHOP_EXCHANGE), pack)
		return false
	}

	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		pack := &webapi_proto.ASCreateExchangeOrder{
			Snid:      p.SnId,
			Platform:  p.Platform,
			Type:      cdata.Type,
			GoodsId:   cdata.Id,
			VCard:     cdata.NeedNum,
			GoodsName: cdata.Name,
			UserName:  username,
			Mobile:    mobile,
			Comment:   comment,
		}
		buff, err := webapi.API_CreateExchange(common.GetAppId(), pack)
		if err != nil {
			logger.Logger.Error("API_CreateExchange error:", err)
		}

		as := &webapi_proto.SACreateExchangeOrder{}
		if err := proto.Unmarshal(buff, as); err != nil {
			logger.Logger.Errorf("API_CreateExchange err: %v %v", err, as.Tag)
		}
		return as

	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		pack := &shop.SCShopExchange{
			RetCode: shop.OpResultCode_OPRC_Error,
		}

		as := data.(*webapi_proto.SACreateExchangeOrder) // 必不为空

		if as.Tag == webapi_proto.TagCode_SUCCESS {
			pack.RetCode = shop.OpResultCode_OPRC_Sucess
			//p.SendVCoinDiffData() // 强推
			name := "V卡"
			if item := BagMgrSington.GetBagItemById(p.SnId, VCard); item != nil && item.Name != "" {
				name = item.Name
			} else {
				logger.Logger.Errorf("ExchangeList name err snid %v item %v ", p.SnId, VCard)
			}
			BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemConsume, VCard, name, cdata.NeedNum, fmt.Sprintf("兑换订单 %v-%v", cdata.Id, cdata.Name))
			BagMgrSington.SyncBagData(p, VCard)
		} else {
			if as.GetReturnCPO() != nil {
				switch as.ReturnCPO.Err {
				case Err_ExShopNEnough:
					pack.RetCode = shop.OpResultCode_OPRC_ExchangeNotEnough
				case Err_ExShopLimit:
					pack.RetCode = shop.OpResultCode_OPRC_ExchangeLimit
				case Err_ExShopData:
					pack.RetCode = shop.OpResultCode_OPRC_ExchangeDataRtt
				default:
					pack.RetCode = shop.OpResultCode_OPRC_ExchangeSoldOut
				}
			}
			logger.Logger.Trace("API_CreateExchange: ", as.Tag, as.GetReturnCPO())
			items := []*Item{&Item{
				ItemId:  VCard,         // 物品id
				ItemNum: cdata.NeedNum, // 数量
			}}
			BagMgrSington.AddJybBagInfo(p, items) // 后台订单创建失败 返回V卡
		}

		p.SendToClient(int(shop.SPacketID_PACKET_SC_SHOP_EXCHANGE), pack)
	}), "CreateExchange").Start()

	return
}

// 兑换列表
func (this *ShopMgr) ExchangeList(p *Player) (ret bool) {
	ret = true

	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		pack := &webapi_proto.ASGetExchangeShop{
			Snid:     p.SnId,
			Platform: p.Platform,
		}
		buff, err := webapi.API_ExchangeList(common.GetAppId(), pack)

		// spack := &shop.SCShopExchangeRecord{}
		as := &webapi_proto.SAGetExchangeShop{}
		logger.Logger.Trace("SCShopOrder:", pack)
		if err != nil {
			logger.Logger.Error("API_ExchangeList error:", err)
			return nil
		} else if err := proto.Unmarshal(buff, as); err != nil { // 有订单

			logger.Logger.Errorf("ExchangeRecord: %v", err)
			return nil

		}

		return as
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {

		pack := &shop.SCShopExchangeList{
			RetCode: shop.OpResultCode_OPRC_Sucess,
		}

		if data != nil {
			if as := data.(*webapi_proto.SAGetExchangeShop); as != nil {

				for _, v := range as.List {
					pack.Infos = append(pack.Infos, &shop.ShopExchangeInfo{
						Type:         v.Type,
						Picture:      v.Picture,
						Name:         v.Name,
						NeedNum:      v.Price,
						Rule:         v.Content,
						GoodsId:      v.Id,
						ShopLimit:    v.ShopLimit,
						DayMaxLimit:  v.DayMaxLimit,
						DayPlayLimit: v.DayPlayLimit,
					})
				}
			}
		} else {
			pack.RetCode = shop.OpResultCode_OPRC_Error
		}
		p.SendToClient(int(shop.SPacketID_PACKET_SC_SHOP_EXCHANGELIST), pack)
	}), "ExchangeList").Start()

	return
}

// 获取兑换记录
func (this *ShopMgr) GetExchangeRecord(p *Player, pageNo int32) {
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		pack := &webapi_proto.ASGetExchangeOrder{
			Snid:     p.SnId,
			Platform: p.Platform,
			Page:     pageNo,
		}
		buff, err := webapi.API_ExchangeRecord(common.GetAppId(), pack)

		// spack := &shop.SCShopExchangeRecord{}
		as := &webapi_proto.SAGetExchangeOrder{}
		logger.Logger.Trace("SCShopOrder:", pack)
		if err != nil {
			logger.Logger.Error("API_ExchangeRecord error:", err)
			return nil
		} else if err := proto.Unmarshal(buff, as); err != nil { // 有订单

			logger.Logger.Errorf("ExchangeRecord: %v", err)
			return nil

		}

		return as
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		pack := &shop.SCShopExchangeRecord{}
		if data != nil {
			if as := data.(*webapi_proto.SAGetExchangeOrder); as != nil {

				pack.PageNo = as.CurPage
				pack.PageSize = as.PageLimit
				pack.PageSum = as.PageTotal
				for _, v := range as.OrderList {

					record := &shop.ShopExchangeRecord{
						CreateTs: v.CreateTime,
						OrderId:  fmt.Sprintf("%v", v.Id),
						State:    v.Status,
						Name:     v.Name,
						Remark:   v.Remark,
					}
					// remark
					pack.Infos = append(pack.Infos, record)
				}

			}
		}
		p.SendToClient(int(shop.SPacketID_PACKET_SC_SHOP_EXCHANGERECORD), pack)
	}), "ExchangeRecord").Start()

}

func (this *ShopMgr) ShopCheckShowRed(p *Player) {
	if p == nil {
		return
	}
	for i := 0; i < 3; i++ {
		shops := this.getShopInfos() //this.Shop // srvdata.PBDB_Shop1Mgr.Datas.Arr
		var isShow bool
		for _, shopInfo := range shops {
			var AdLookedNum, remainingTime int32
			var lastLookTime int64
			if shopInfo.Ad > 0 && (i == -1 || (i < len(shopInfo.Location) && shopInfo.Location[i] == 1)) {
				shopTotal := p.ShopTotal[shopInfo.Id]
				if shopTotal != nil {
					AdLookedNum = shopTotal.AdLookedNum
				}
				lastLookTime = p.ShopLastLookTime[shopInfo.Id]
				if AdLookedNum < int32(len(shopInfo.CoolingTime)) {
					remainingTime = shopInfo.CoolingTime[AdLookedNum]
				}
				dif := int32(time.Now().Unix() - lastLookTime)
				if dif >= remainingTime {
					remainingTime = 0
					if i+1 <= len(shopInfo.Location) {
						isShow = true
						p.SendShowRed(hall_proto.ShowRedCode_Shop, int32(i+1), 1)
					}
					break
				}
			}
		}
		if !isShow {
			p.SendShowRed(hall_proto.ShowRedCode_Shop, int32(i+1), 0)
		}
	}
}

func (this *ShopMgr) Update() {

}

func (this *ShopMgr) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(ShopMgrSington, time.Second, 0)
}
