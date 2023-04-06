package main

import (
	"fmt"
	"strconv"
	"time"

	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/bag"
	player_proto "games.yol.com/win88/protocol/player"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
)

// 背包
const (
	BagItemMax int32 = 200
)

// 道具功能  Function
const (
	ItemCanUse  = iota //可以使用
	ItemCanGive        //可以赠送
	ItemCanSell        //可以出售
	ItemMax
)

var BagMgrSington = &BagMgr{
	PlayerBag: make(map[int32]*BagInfo),
}

type BagMgr struct {
	PlayerBag map[int32]*BagInfo // 考虑sync.Map
}
type BagInfo struct {
	SnId     int32           //玩家账号直接在这里生成
	Platform string          //平台
	BagItem  map[int32]*Item //背包数据  key为itemId
	dirty    bool
}

type Item struct {
	ItemId  int32 // 物品ID
	ItemNum int32 // 物品数量
	////数据表数据
	Name string // 名称
	//ShowLocation   []int32 // 显示位置
	//Classify       []int32 // 分页类型 1，道具类 	2，资源类	3，兑换类
	//Type           int32   // 道具种类 1，宠物碎片 2，角色碎片
	Effect0  []int32 // 竖版道具功能 1，使用 2，赠送 3，出售
	Effect   []int32 // 横版道具功能 1，使用 2，赠送 3，出售
	SaleType int32   // 出售类型
	SaleGold int32   // 出售金额
	//Composition    int32   // 能否叠加 1，能 2，不能
	//CompositionMax int32   // 叠加上限
	//Time           int32   // 道具时效 0为永久
	//Location       string  // 跳转页面
	//Describe       string  // 道具描述
	//数据库数据
	ObtainTime int64 //获取的时间
}

func (this *BagMgr) ModuleName() string {
	return "BagMgr"
}

func (this *BagMgr) Init() {

}
func (this *BagMgr) GetBagInfo(sid int32) *BagInfo {
	if v, exist := this.PlayerBag[sid]; exist {
		return v
	}
	return nil
}

func (this *BagMgr) InitBagInfo(p *Player, plt string) {
	if p == nil {
		return
	}
	if this.PlayerBag[p.SnId] != nil {
		PetMgrSington.CheckShowRed(p)
		return
	}
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		return model.GetBagInfo(p.SnId, plt)
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		if data != nil {
			bagInfo := data.(*model.BagInfo)
			if bagInfo != nil {
				//数据表 数据库数据相结合
				newBagInfo := &BagInfo{
					SnId:     p.SnId,
					Platform: plt,
					BagItem:  make(map[int32]*Item),
				}
				for k, bi := range bagInfo.BagItem {
					item := srvdata.PBDB_GameItemMgr.GetData(bi.ItemId)
					if item != nil && bi.ItemNum > 0 {
						newBagInfo.BagItem[k] = &Item{
							ItemId:     bi.ItemId,
							ItemNum:    bi.ItemNum,
							ObtainTime: bi.ObtainTime,
						}
					} else {
						logger.Logger.Error("InitBagInfo err: item is nil. ItemId:", bi.ItemId)
					}
				}
				this.PlayerBag[p.SnId] = newBagInfo
			}
		}
		PetMgrSington.CheckShowRed(p)
	}), "InitBagInfo").Start()
}

func BagSubverify(args *BagInfo) {

	var del []int32
	for i, v := range args.BagItem {
		if v.ItemNum <= 0 {
			del = append(del, i)
		}
	}
	for _, v := range del {
		delete(args.BagItem, v)
	}

}

// 获取个人的指定道具信息
func (this *BagMgr) GetBagItemById(snid, propId int32) *Item {
	if bagItme, ok := this.PlayerBag[snid]; ok {
		if bagItme != nil {
			item := bagItme.BagItem[propId]
			if item != nil {
				itemX := srvdata.PBDB_GameItemMgr.GetData(item.ItemId)
				if itemX != nil {
					item.Name = itemX.Name
					//item.ShowLocation = itemX.ShowLocation
					//item.Classify = itemX.Classify
					//item.Type = itemX.Type
					item.Effect0 = itemX.Effect
					item.Effect = itemX.Effect
					item.SaleType = itemX.SaleType
					item.SaleGold = itemX.SaleGold
					//item.Composition = itemX.Composition
					//item.CompositionMax = itemX.CompositionMax
					//item.Time = itemX.Time
					//item.Location = itemX.Location
					//item.Describe = itemX.Describe
				}
				return item
			}
		}
	}
	return nil
}
func (this *BagMgr) AddJybBagInfo(p *Player, additems []*Item) (*BagInfo, bag.OpResultCode) {
	var itemids []int32
	var newBagInfo *BagInfo
	sid := p.SnId
	plt := p.Platform
	if _, exist := this.PlayerBag[sid]; !exist {
		//newBagInfo = model.NewBagInfo(sid, plt)
		newBagInfo = &BagInfo{
			SnId:     sid,
			Platform: plt,
			BagItem:  make(map[int32]*Item),
		}
	} else {
		newBagInfo = this.PlayerBag[sid]
	}
	//vdirty := false
	//if !BagsAddverify(newBagInfo, additems) { //预判断 超过
	//	return newBagInfo, bag.OpResultCode_OPRC_BagFull
	//}
	for _, additem := range additems {
		itemids = append(itemids, additem.ItemId)
		if itm, exist := newBagInfo.BagItem[additem.ItemId]; exist {
			itm.ItemNum += additem.ItemNum
		} else {
			item := srvdata.PBDB_GameItemMgr.GetData(additem.ItemId)
			if item != nil {
				newBagInfo.BagItem[additem.ItemId] = &Item{
					ItemId:     item.Id,         // 物品id
					ItemNum:    additem.ItemNum, // 数量
					ObtainTime: time.Now().Unix(),
					//Name:           item.Name,           // 名称
					//Classify:       item.Classify,       // 分页类型 1，道具类 	2，资源类	3，兑换类
					//Type:           item.Type,           // 道具种类 1，宠物碎片 2，角色碎片
					//Effect:         item.Effect,         //  浅拷贝 道具功能 1，使用 2，赠送 3，出售
					//SaleType:       item.SaleType,       // 出售类型
					//SaleGold:       item.SaleGold,       // 出售金额
					//Composition:    item.Composition,    // 能否叠加 1，能 2，不能
					//CompositionMax: item.CompositionMax, // 叠加上限
					//Time:           item.Time,           // 道具时效 0为永久
					//Location:       item.Location,       // 跳转页面
					//Describe:       item.Describe,       // 道具描述
				}
			} else {
				return newBagInfo, bag.OpResultCode_OPRC_IdErr
			}
		}
		//if additem.ItemId == VCard {
		//	vdirty = true
		//}
	}
	newBagInfo.dirty = true
	p.dirty = true
	this.PlayerBag[sid] = newBagInfo
	//if vdirty {
	//	p.SendVCoinDiffData() // 强推
	//}
	//if err := model.UpBagItem(newBagInfo); err != nil {
	//	return newBagInfo, bag.OpResultCode_OPRC_DbErr
	//}

	this.SyncBagData(p, itemids...)

	return newBagInfo, bag.OpResultCode_OPRC_Sucess
}

// 出售道具
func (this *BagMgr) SaleItem(p *Player, itemId int32, num int32) bool {
	if bagInfo, ok := this.PlayerBag[p.SnId]; ok {
		if item, ok1 := bagInfo.BagItem[itemId]; ok1 {
			if item.ItemNum >= num {
				//可以出售
				item.ItemNum -= num
				bagInfo.dirty = true
				p.dirty = true
				return true
			}
		}
	}
	return false
}

func (this *BagMgr) SaveBagData(snid int32, plt string) {
	bagInfo := this.PlayerBag[snid]
	logger.Logger.Trace("SaveBagData:", bagInfo)
	if bagInfo != nil && bagInfo.dirty {
		bagInfo.dirty = false
		type BagInfoMap struct {
			SnId     int32   //玩家账号直接在这里生成
			Platform string  //平台
			BagItem  []*Item //背包数据  key为itemId
		}
		var biMap = BagInfoMap{
			SnId:     bagInfo.SnId,
			Platform: bagInfo.Platform,
		}
		for _, v := range bagInfo.BagItem {
			biMap.BagItem = append(biMap.BagItem, &Item{ItemId: v.ItemId, ItemNum: v.ItemNum, ObtainTime: v.ObtainTime})
		}
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			newBagInfo := &model.BagInfo{
				SnId:     biMap.SnId,
				Platform: biMap.Platform,
				BagItem:  make(map[int32]*model.Item),
			}
			for _, v := range biMap.BagItem {
				newBagInfo.BagItem[v.ItemId] = &model.Item{ItemId: v.ItemId, ItemNum: v.ItemNum, ObtainTime: v.ObtainTime}
			}
			return model.UpBagItem(newBagInfo)
		}), nil, "SaveBagData").StartByFixExecutor("SnId:" + strconv.Itoa(int(snid)))
	}
}
func (this *BagMgr) RecordItemLog(platform string, snid, logType, itemId int32, itemName string, count int32, remark string) {
	log := model.NewItemLogEx(platform, snid, logType, itemId, itemName, count, remark)
	if log != nil {
		//logger.Logger.Trace("RecordItemLog 开始记录 道具操作")
		LogChannelSington.WriteLog(log)
	}
}

// 赠送道具到邮件
// srcId 发送人 srcName发送人名字
// items[0]:道具id items[1]:道具数量 items[2]:道具id items[3]:道具数量
func (this *BagMgr) AddMailByItem(platform string, srcId int32, srcName string, snid int32, showId int64, items []int32) {
	logger.Logger.Trace("AddMailByItem:", srcId, srcName, items)
	var newMsg *model.Message
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		content := fmt.Sprintf("玩家%v给您赠送了礼物，请注意查收", srcName)
		newMsg = model.NewMessageByPlayer("", 1, srcId, srcName, snid, model.MSGTYPE_ITEM, "玩家赠送", content,
			0, 0, model.MSGSTATE_UNREAD, time.Now().Unix(), 0, "", items, platform, showId)
		return model.InsertMessage(platform, newMsg)
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		if data == nil {
			p := PlayerMgrSington.GetPlayerBySnId(snid)
			if p != nil {
				p.AddMessage(newMsg)
			}
		}
	}), "AddMailByItem").Start()
}
func (this *BagMgr) VerifyUpJybInfo(p *Player, args *model.VerifyUpJybInfoArgs) {

	type VerifyInfo struct {
		jyb *model.JybInfo
		err error
	}
	pack := &player_proto.SCPlayerSetting{
		OpRetCode: player_proto.OpResultCode_OPRC_Sucess,
	}
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {

		jyb, err := model.VerifyUpJybInfo(args)
		info := &VerifyInfo{
			jyb: jyb,
			err: err,
		}
		return info
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		if data != nil && data.(*VerifyInfo) != nil {
			jyb := data.(*VerifyInfo).jyb
			err := data.(*VerifyInfo).err
			if jyb != nil && jyb.Award != nil { // 领取到礼包
				pack.GainItem = &player_proto.JybInfoAward{}
				if jyb.Award.Item != nil {
					if len(jyb.Award.Item) > 0 {
						items := make([]*Item, 0)
						for _, v := range jyb.Award.Item {
							items = append(items, &Item{
								ItemId:     v.ItemId,  // 物品id
								ItemNum:    v.ItemNum, // 数量
								ObtainTime: time.Now().Unix(),
							})
						}
						if _, code := this.AddJybBagInfo(p, items); code != bag.OpResultCode_OPRC_Sucess { //TODO 添加失败 要回退礼包
							logger.Logger.Errorf("CSPlayerSettingHandler AddJybBagInfo err", code)
							pack.OpRetCode = player_proto.OpResultCode_OPRC_Error
							proto.SetDefaults(pack)
							p.SendToClient(int(player_proto.PlayerPacketID_PACKET_ALL_SETTING), pack)
						} else {
							for _, v := range items {
								itemData := srvdata.PBDB_GameItemMgr.GetData(v.ItemId)
								if itemData != nil {
									BagMgrSington.RecordItemLog(p.Platform, p.SnId, ItemObtain, v.ItemId, itemData.Name, v.ItemNum, "礼包领取")
								}
							}
							PetMgrSington.CheckShowRed(p)
						}
						p.dirty = true
					}
				}
				for _, v := range jyb.Award.Item {
					//if _, code := BagMgrSington.UpBagInfo(p.SnId, p.Platform, v.ItemId, v.ItemNum); code == bag.OpResultCode_OPRC_Sucess { // 需修改
					pack.GainItem.ItemId = append(pack.GainItem.ItemId, &player_proto.ItemInfo{
						ItemId:  v.ItemId,
						ItemNum: v.ItemNum,
					})
				}
				p.Coin += jyb.Award.Coin
				p.Diamond += jyb.Award.Diamond
				p.dirty = true
				pack.GainItem.Coin = jyb.Award.Coin
				pack.GainItem.Diamond = jyb.Award.Diamond
			} else {
				switch err.Error() {
				case model.ErrJybISReceive.Error():
					pack.OpRetCode = player_proto.OpResultCode_OPRC_Jyb_Receive
				case model.ErrJYBPlCode.Error():
					pack.OpRetCode = player_proto.OpResultCode_OPRC_Jyb_CodeExist
				case model.ErrJybTsTimeErr.Error():
					pack.OpRetCode = player_proto.OpResultCode_OPRC_Jyb_TimeErr
				default:
					pack.OpRetCode = player_proto.OpResultCode_OPRC_Jyb_CodeErr
				}
			}
		} else {
			proto.SetDefaults(pack)
			p.SendToClient(int(player_proto.PlayerPacketID_PACKET_ALL_SETTING), pack)
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(player_proto.PlayerPacketID_PACKET_ALL_SETTING), pack)
	}), "VerifyUpJybInfo").Start()
	// 先检查玩家背包是否足够

}
func (this *BagMgr) SyncBagData(p *Player, itemIds ...int32) {
	var itemInfos []*bag.ItemInfo
	for _, itemId := range itemIds {
		itemInfo := BagMgrSington.GetBagItemById(p.SnId, itemId)
		if itemInfo != nil {
			itemInfos = append(itemInfos, &bag.ItemInfo{
				ItemId:  itemInfo.ItemId,
				ItemNum: itemInfo.ItemNum,
				//Name:           itemInfo.Name,
				//Classify:       itemInfo.Classify,
				//Type:           itemInfo.Type,
				//Effect0:        itemInfo.Effect0,
				//Effect:         itemInfo.Effect,
				//SaleType:       itemInfo.SaleType,
				//SaleGold:       itemInfo.SaleGold,
				//Composition:    itemInfo.Composition,
				//CompositionMax: itemInfo.CompositionMax,
				//Time:           itemInfo.Time,
				//Location:       itemInfo.Location,
				//Describe:       itemInfo.Describe,
				ObtainTime: itemInfo.ObtainTime,
			})
			if itemInfo.ItemId == VCard {
				FriendMgrSington.UpdateFriendVCard(p.SnId, int64(itemInfo.ItemNum))
			}
		}
	}
	p.SyncBagData(itemInfos)
}

func (this *BagMgr) Update() {

}

func (this *BagMgr) Shutdown() {
	for _, bagInfo := range this.PlayerBag {
		newBagInfo := &model.BagInfo{
			SnId:     bagInfo.SnId,
			Platform: bagInfo.Platform,
			BagItem:  make(map[int32]*model.Item),
		}
		for k, v := range bagInfo.BagItem {
			newBagInfo.BagItem[k] = &model.Item{ItemId: v.ItemId, ItemNum: v.ItemNum, ObtainTime: v.ObtainTime}
		}
		model.UpBagItem(newBagInfo)
	}
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(BagMgrSington, time.Second, 0)
}
