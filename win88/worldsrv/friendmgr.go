package main

import (
	"errors"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/friend"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
	"strconv"
	"time"
)

const (
	FriendState_define   int = iota //确定好友关系
	FriendState_undefine            //待确定好友关系
)

const (
	ListType_Friend    int32 = iota // 0.好友列表
	ListType_Apply                  // 1.申请列表
	ListType_Recommend              // 2.推荐列表
)

const (
	OpType_Apply  int32 = iota // 0.申请
	OpType_Agree               // 1.同意
	OpType_Refuse              // 2.拒绝
	OpType_Delete              // 3.删除
)

const (
	Invite_Agree  int = iota // 0.同意
	Invite_Refuse            // 1.拒绝
)

const FriendMaxNum = 200
const ShieldMaxNum = 10000
const FriendApplyMaxNum = 1000 //好友申请上限

var FriendMgrSington = &FriendMgr{
	FriendList:      make(map[int32]*Friend),
	TsRecommendList: make(map[int32]int64),
	TsInviteCd:      make(map[string]int64),
}

type Friend struct {
	Id         bson.ObjectId `bson:"_id"`
	Platform   string
	SnId       int32
	BindFriend []*BindFriend
	Dirty      bool
	Name       string
	Head       int32
	Sex        int32
	Coin       int64
	Diamond    int64
	VCard      int64
	Roles      map[int32]int32 //人物
	Pets       map[int32]int32 //宠物
	Shield     []int32
	LogoutTime int64 //登出时间
}

type BindFriend struct {
	SnId       int32
	CreateTime int64 //建立时间
	//缓存数据
	Platform   string
	Name       string
	Head       int32 //头像
	Sex        int32 //性别
	LogoutTime int64 //登出时间
}

type FriendMgr struct {
	FriendList      map[int32]*Friend
	TsRecommendList map[int32]int64
	TsInviteCd      map[string]int64
}

func (this *FriendMgr) ModuleName() string {
	return "FriendMgr"
}

func (this *FriendMgr) Init() {
}

func (this *FriendMgr) CanGetRecommendFriendList(snid int32) (bool, int64) {
	if this.TsRecommendList[snid] == 0 {
		this.TsRecommendList[snid] = time.Now().Unix()
		return true, 0
	}
	if time.Now().Unix()-this.TsRecommendList[snid] > 20 {
		this.TsRecommendList[snid] = time.Now().Unix()
		return true, 0
	}
	return false, time.Now().Unix() - this.TsRecommendList[snid]
}

func (this *FriendMgr) CanInvite(snid, fsnid int32) bool {
	strSnid := strconv.FormatInt(int64(snid), 10) + "_" + strconv.FormatInt(int64(fsnid), 10)
	if this.TsInviteCd[strSnid] == 0 {
		this.TsInviteCd[strSnid] = time.Now().Unix()
		return true
	}
	if time.Now().Unix()-this.TsInviteCd[strSnid] > 30 {
		this.TsInviteCd[strSnid] = time.Now().Unix()
		return true
	}
	return false
}

func (this *FriendMgr) ExchangeModelFriend2Cache(friend *model.Friend) *Friend {
	if friend == nil {
		return nil
	}
	f := &Friend{
		Id:         friend.Id,
		Platform:   friend.Platform,
		SnId:       friend.SnId,
		Name:       friend.Name,
		Head:       friend.Head,
		Sex:        friend.Sex,
		Coin:       friend.Coin,
		Diamond:    friend.Diamond,
		VCard:      friend.VCard,
		Roles:      friend.Roles,
		Pets:       friend.Pets,
		Shield:     friend.Shield,
		LogoutTime: friend.LogoutTime,
	}
	if friend.BindFriend != nil {
		for _, bindFriend := range friend.BindFriend {
			bf := &BindFriend{
				SnId:       bindFriend.SnId,
				CreateTime: bindFriend.CreateTime,
			}
			f.BindFriend = append(f.BindFriend, bf)
		}
	}
	return f
}

// 新增一条信息
func (this *FriendMgr) AddFriend(snid int32) {
	logger.Logger.Trace("(this *FriendMgr) AddFriend ", snid)
	if this.FriendList[snid] == nil {
		p := PlayerMgrSington.GetPlayerBySnId(snid)
		if p != nil {
			vCard := int64(0)
			item := BagMgrSington.GetBagItemById(p.SnId, VCard)
			if item != nil {
				vCard = int64(item.ItemNum)
			}
			roles := map[int32]int32{}
			if p.Roles != nil && p.Roles.ModUnlock != nil {
				roles = p.Roles.ModUnlock
			}
			pets := map[int32]int32{}
			if p.Pets != nil && p.Pets.ModUnlock != nil {
				pets = p.Pets.ModUnlock
			}
			f := model.NewFriend(p.Platform, p.SnId, p.Name, p.Head, p.Sex, p.Coin, p.Diamond, vCard, roles, pets)
			nf := this.ExchangeModelFriend2Cache(f)
			nf.Dirty = true
			this.FriendList[snid] = nf
		}
	}
}

// 同意好友申请
func (this *FriendMgr) AddBindFriend(snid int32, bf *BindFriend) friend.OpResultCode {
	logger.Logger.Trace("(this *FriendMgr) AddBindFriend ", snid, " BindFriend: ", bf)
	if bf == nil {
		return friend.OpResultCode_OPRC_Friend_NoPlayer
	}
	if this.IsFriend(snid, bf.SnId) {
		return friend.OpResultCode_OPRC_Friend_AlreadyAdd
	}

	dfl_dest := this.GetBindFriendList(snid)
	if dfl_dest != nil {
		if len(dfl_dest) >= FriendMaxNum {
			return friend.OpResultCode_OPRC_Friend_DestFriendMax
		}
	}
	dfl := this.GetBindFriendList(bf.SnId)
	if dfl != nil {
		if len(dfl) >= FriendMaxNum {
			return friend.OpResultCode_OPRC_Friend_FriendMax
		}
	}

	bf.CreateTime = time.Now().Unix()
	this.FriendList[snid].BindFriend = append(this.FriendList[snid].BindFriend, bf)
	this.FriendList[snid].Dirty = true
	return friend.OpResultCode_OPRC_Sucess
}

// 删除一条缓存信息
func (this *FriendMgr) DelFriendCache(snid int32) {
	if this.FriendList[snid] == nil {
		return
	}
	delete(this.FriendList, snid)
}

// 删除好友
func (this *FriendMgr) DelBindFriend(snid, destSnid int32) bool {
	logger.Logger.Trace("(this *FriendMgr) DelBindFriend ", snid, " destSnid: ", destSnid)
	fl := this.FriendList[snid]
	if fl == nil {
		return false
	}
	for i, f := range fl.BindFriend {
		if f.SnId == destSnid {
			fl.BindFriend = append(fl.BindFriend[:i], fl.BindFriend[i+1:]...)
			fl.Dirty = true
			return true
		}
	}
	return false
}

// 更新金币
func (this *FriendMgr) UpdateFriendCoin(snid int32, coin int64) {
	logger.Logger.Trace("(this *FriendMgr) UpdateFriendCoin ", snid)
	fl := this.FriendList[snid]
	if fl == nil {
		return
	}
	if fl.Coin == coin {
		return
	}
	fl.Coin = coin
	fl.Dirty = true
}

// 更新钻石
func (this *FriendMgr) UpdateFriendDiamond(snid int32, diamond int64) {
	logger.Logger.Trace("(this *FriendMgr) UpdateFriendDiamond ", snid)
	fl := this.FriendList[snid]
	if fl == nil {
		return
	}
	if fl.Diamond == diamond {
		return
	}
	fl.Diamond = diamond
	fl.Dirty = true
}

// 更新V卡
func (this *FriendMgr) UpdateFriendVCard(snid int32, vCard int64) {
	logger.Logger.Trace("(this *FriendMgr) UpdateFriendVCard ", snid)
	fl := this.FriendList[snid]
	if fl == nil {
		return
	}
	if fl.VCard == vCard {
		return
	}
	fl.VCard = vCard
	fl.Dirty = true
}

// 更新角色
func (this *FriendMgr) UpdateFriendRoles(snid int32, modUnlock map[int32]int32) {
	logger.Logger.Trace("(this *FriendMgr) UpdateFriendRoles ", snid)
	fl := this.FriendList[snid]
	if fl == nil {
		return
	}
	fl.Roles = modUnlock
	fl.Dirty = true
}

// 更新宠物
func (this *FriendMgr) UpdateFriendPets(snid int32, modUnlock map[int32]int32) {
	logger.Logger.Trace("(this *FriendMgr) UpdateFriendPets ", snid)
	fl := this.FriendList[snid]
	if fl == nil {
		return
	}
	fl.Pets = modUnlock
	fl.Dirty = true
}

func (this *FriendMgr) GetFriendBySnid(snid int32) *Friend {
	return this.FriendList[snid]
}

func (this *FriendMgr) GetBindFriendList(snid int32) []*BindFriend {
	allFriend := []*BindFriend{}
	friend := this.FriendList[snid]
	if friend != nil && friend.BindFriend != nil {
		for _, bindFriend := range friend.BindFriend {
			allFriend = append(allFriend, bindFriend)
		}
	}
	return allFriend
}

func (this *FriendMgr) IsFriend(snid, destSnid int32) bool {
	logger.Logger.Trace("(this *FriendMgr) IsFriend ", snid, " -> ", destSnid)
	dfl := this.GetBindFriendList(snid)
	if dfl != nil && len(dfl) != 0 {
		for _, df := range dfl {
			if df.SnId == destSnid {
				return true
			}
		}
	}
	return false
}

func (this *FriendMgr) LoadFriendData(platform string, snid int32) {
	if this.FriendList[snid] != nil {
		return
	}
	var friendData *Friend
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		friendDB, err := model.QueryFriendBySnid(platform, snid)
		if err != nil {
			logger.Logger.Error("QueryFriendBySnid:", err, snid)
			return errors.New("QueryFriendBySnid,err:" + err.Error())
		}
		if friendDB == nil {
			return nil
		}
		friendData = this.ExchangeModelFriend2Cache(friendDB)
		offSnids := []int32{} //离线玩家id
		if friendData.BindFriend != nil {
			for _, bf := range friendData.BindFriend {
				p := PlayerMgrSington.GetPlayerBySnId(bf.SnId)
				if p != nil {
					bf.Platform = p.Platform
					bf.Name = p.Name
					bf.Head = p.Head
					bf.Sex = p.Sex
					if !p.IsOnLine() {
						bf.LogoutTime = p.LastLogoutTime.Unix()
					}
				} else {
					offSnids = append(offSnids, bf.SnId)
				}
			}
		}
		if len(offSnids) > 0 {
			offFriends, err := model.QueryFriendsBySnids(platform, offSnids)
			if err != nil {
				logger.Logger.Error("QueryFriendsBySnids is err:", err)
				return errors.New("QueryFriendsBySnids,err:" + err.Error())
			}
			for _, offFriend := range offFriends {
				for _, i3 := range friendData.BindFriend {
					if i3.SnId == offFriend.SnId {
						i3.Platform = offFriend.Platform
						i3.Name = offFriend.Name
						i3.Head = offFriend.Head
						i3.Sex = offFriend.Sex
						i3.LogoutTime = offFriend.LogoutTime
						break
					}
				}
			}
		}
		return nil
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		if data != nil {
			logger.Logger.Error("login LoadFriendData err:", data)
			return
		}
		if friendData == nil {
			this.AddFriend(snid)
			logger.Logger.Info("(this *FriendMgr) LoadFriendData cache: ", this.FriendList)
			return
		}
		this.FriendList[snid] = friendData
		logger.Logger.Info("(this *FriendMgr) LoadFriendData db: ", this.FriendList)
		p := PlayerMgrSington.GetPlayerBySnId(snid)
		if p != nil {
			if p.Roles != nil {
				this.UpdateFriendRoles(snid, p.Roles.ModUnlock)
			}
			if p.Pets != nil {
				this.UpdateFriendPets(snid, p.Pets.ModUnlock)
			}
		}
		FriendUnreadMgrSington.LoadFriendUnreadData(platform, snid)
	})).StartByFixExecutor("LoadFriendData")
}

func (this *FriendMgr) CheckSendFriendApplyData(p *Player) {
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		ret, err := model.QueryFriendApplyBySnid(p.Platform, p.SnId)
		if err != nil {
			return nil
		}
		return ret
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		ret := data.(*model.FriendApply)
		if ret != nil {
			if ret.ApplySnids != nil {
				pack := &friend.SCFriendApplyData{}
				for _, as := range ret.ApplySnids {
					fa := &friend.FriendApply{
						Snid:     proto.Int32(as.SnId),
						Name:     proto.String(as.Name),
						CreateTs: proto.Int64(as.CreateTs),
					}
					pack.FriendApplys = append(pack.FriendApplys, fa)
				}
				if len(pack.FriendApplys) > 0 {
					proto.SetDefaults(pack)
					p.SendToClient(int(friend.FriendPacketID_PACKET_SCFriendApplyData), pack)
					logger.Logger.Trace("SCFriendApplyData: 好友申请列表 pack: ", pack)
				}
			}
		}
	})).StartByFixExecutor("QueryFriendApplyBySnid")
}

func (this *FriendMgr) SaveFriendData(snid int32, logout bool) {
	logger.Logger.Trace("(this *FriendMgr) SaveFriendData ", snid)
	friendList := this.FriendList[snid]
	if friendList != nil && friendList.Dirty {
		friendList.Dirty = false
		fl := &model.Friend{
			Id:         friendList.Id,
			Platform:   friendList.Platform,
			SnId:       friendList.SnId,
			Name:       friendList.Name,
			Head:       friendList.Head,
			Sex:        friendList.Sex,
			Coin:       friendList.Coin,
			Diamond:    friendList.Diamond,
			VCard:      friendList.VCard,
			Roles:      friendList.Roles,
			Pets:       friendList.Pets,
			Shield:     friendList.Shield,
			LogoutTime: friendList.LogoutTime,
		}
		if friendList.BindFriend != nil {
			for _, bf := range friendList.BindFriend {
				f := &model.BindFriend{
					SnId:       bf.SnId,
					CreateTime: bf.CreateTime,
				}
				fl.BindFriend = append(fl.BindFriend, f)
			}
		}
		if friendList.Roles != nil {
			for id, level := range friendList.Roles {
				fl.Roles[id] = level
			}
		}
		if friendList.Pets != nil {
			for id, level := range friendList.Pets {
				fl.Pets[id] = level
			}
		}
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.UpsertFriend(fl)
		}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			if logout {
				this.DelFriendCache(snid)
			}
		})).StartByFixExecutor("SnId:" + strconv.Itoa(int(snid)))
	}
}

func (this *FriendMgr) AddShield(snid, ssnid int32) {
	fl := this.FriendList[snid]
	if fl == nil {
		return
	}
	if fl.Shield == nil {
		fl.Shield = []int32{}
	}
	if len(fl.Shield) > ShieldMaxNum {
		return
	}
	fl.Shield = append(fl.Shield, ssnid)
	fl.Dirty = true
}

func (this *FriendMgr) DelShield(snid, ssnid int32) {
	fl := this.FriendList[snid]
	if fl == nil {
		return
	}
	if fl.Shield == nil {
		return
	}
	if len(fl.Shield) == 0 {
		return
	}
	fl.Shield = common.DelSliceInt32(fl.Shield, ssnid)
	fl.Dirty = true
}

func (this *FriendMgr) IsShield(snid, ssnid int32) bool {
	fl := this.FriendList[snid]
	if fl == nil {
		return false
	}
	if fl.Shield == nil {
		return false
	}
	if len(fl.Shield) == 0 {
		return false
	}
	return common.InSliceInt32(fl.Shield, ssnid)
}

func (this *FriendMgr) UpdateLogoutTime(snid int32) {
	fl := this.FriendList[snid]
	if fl == nil {
		return
	}
	fl.LogoutTime = time.Now().Unix()
	fl.Dirty = true

	if fl.BindFriend != nil {
		for _, bf := range fl.BindFriend {
			if data, ok := this.FriendList[bf.SnId]; ok {
				if data.BindFriend != nil {
					for _, bindFriend := range data.BindFriend {
						if bindFriend.SnId == snid {
							bindFriend.LogoutTime = time.Now().Unix()
							break
						}
					}
				}
			}
		}
	}
}

func (this *FriendMgr) SaveAllFriendData() {
	for _, fl := range this.FriendList {
		this.SaveFriendData(fl.SnId, false)
	}
}

func (this *FriendMgr) Update() {
	this.SaveAllFriendData()
}

func (this *FriendMgr) Shutdown() {
	this.SaveAllFriendData()
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(FriendMgrSington, time.Hour, 0)
}
