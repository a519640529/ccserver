package main

import (
	"strconv"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	login_proto "games.yol.com/win88/protocol/login"
	server_proto "games.yol.com/win88/protocol/server"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"games.yol.com/win88/srvdata"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/utils"
	srvlibproto "github.com/idealeak/goserver/srvlib/protocol"
)

const (
	PLATFORM_MIN_SCENE  = 10
	Default_Platform    = "0"
	Default_PlatformInt = 0
)

type PlatformGamePlayerNum struct {
	Nums  map[int32]int //人数信息
	Dirty bool          //变化标记
}

type PlatformObserver interface {
	OnPlatformCreate(p *Platform)
	OnPlatformDestroy(p *Platform)
	OnPlatformChangeIsolated(p *Platform, isolated bool)
	OnPlatformChangeDisabled(p *Platform, disabled bool)
	OnPlatformConfigUpdate(p *Platform, oldCfg, newCfg *webapi_proto.GameFree)
	OnPlatformDestroyByGameFreeId(p *Platform, gameFreeId int32)
}

var PlatformMgrSington = &PlatformMgr{
	Platforms:     make(map[string]*Platform),
	PackageList:   make(map[string]*webapi_proto.AppInfo),
	PromoterList:  make(map[string]PlatformPromoter),
	ScenesByHall:  make(map[int32]map[int]*Scene),
	PlayersInHall: make(map[int32]map[int32]*Player),
	GamePlayerNum: make(map[int32]*PlatformGamePlayerNum), //游戏人数
	GameStatus:    make(map[int32]bool),                   //全局游戏开关
	CommonNotices: make(map[string]*webapi_proto.CommonNoticeList),
}

type PlatformMgr struct {
	BaseClockSinker
	//结构性数据
	Platforms    map[string]*Platform
	PackageList  map[string]*webapi_proto.AppInfo //包对应的平台和上级关系
	PromoterList map[string]PlatformPromoter
	//关联性数据
	ScenesByHall  map[int32]map[int]*Scene         //在公共大厅中房间列表
	PlayersInHall map[int32]map[int32]*Player      //在公共大厅中的玩家
	GamePlayerNum map[int32]*PlatformGamePlayerNum //游戏人数
	dirty         bool
	Observers     []PlatformObserver
	GameStatus    map[int32]bool //全局游戏开关 key:excel表id  value:是否开启 true开启 false关闭
	CommonNotices map[string]*webapi_proto.CommonNoticeList
}

type PlatformPackage struct {
	Tag            string //android包名或者ios标记
	Platform       int32  //所属平台
	Channel        int32  //渠道ID
	Promoter       int32  //推广员ID
	PromoterTree   int32  //无级推广ID
	SpreadTag      int32  //全民包标识 0:普通包 1:全民包
	OpenInstallTag int32  //是否是openinstall包 0:不是 1:是
	Status         int32
	AppStore       int32 //是否是苹果商店包 0:不是 1:是
	ExchangeFlag   int32 //兑换标记 0 关闭包返利 1打开包返利 受平台配置影响
	ExchangeFlow   int32 //兑换比例
	IsForceBind    int32 //是否是强制绑定推广员包 0:不是 1:是
	TagKey         int32 //包标识区分

}
type PlatformPromoter struct {
	Platform string //所属平台
	Promoter string //推广员ID
	Tag      string //android包名或者ios标记
}

func evaluateSceneIncCount(curcount, playernum, perscenemax int) int {
	expectcnt := (playernum/perscenemax + 1) * 2
	if expectcnt < PLATFORM_MIN_SCENE {
		expectcnt = PLATFORM_MIN_SCENE
	}
	return expectcnt - curcount
}

func (pm *PlatformMgr) RegisteObserver(observer PlatformObserver) {
	for _, ob := range pm.Observers {
		if ob == observer {
			return
		}
	}
	pm.Observers = append(pm.Observers, observer)
}

func (pm *PlatformMgr) UnregisteObserver(observer PlatformObserver) {
	for i, ob := range pm.Observers {
		if ob == observer {
			count := len(pm.Observers)
			if i == 0 {
				pm.Observers = pm.Observers[1:]
			} else if i == count-1 {
				pm.Observers = pm.Observers[:count-1]
			} else {
				arr := pm.Observers[:i]
				arr = append(arr, pm.Observers[i+1:]...)
				pm.Observers = arr
			}
		}
	}
}

func (pm *PlatformMgr) CreatePlatform(id int32, isolated bool) *Platform {
	pltId := strconv.Itoa(int(id))
	p := NewPlatform(id, isolated)
	pm.Platforms[pltId] = p
	if p != nil {
		pm.OnPlatformCreate(p)
	}
	return p
}
func (pm *PlatformMgr) CreateDefaultPlatform() {
	//默认平台数据
	defaultPlatform := pm.CreatePlatform(Default_PlatformInt, false)
	defaultPlatform.Disable = false
	//默认平台配置
	pgc := defaultPlatform.PltGameCfg
	if pgc != nil {
		for _, value := range srvdata.PBDB_GameFreeMgr.Datas.Arr {
			if value.GetGameId() > 0 {
				pgc.games[value.GetId()] = &webapi_proto.GameFree{
					Status:     true,
					DbGameFree: value,
				}
			}
		}
		pgc.RecreateCache()
	}
}

func (pm *PlatformMgr) UpsertPlatform(name string, isolated, disable bool, id int32, url string,
	bindOption int32, serviceFlag bool, upgradeAccountGiveCoin, newAccountGiveCoin, perBankNoLimitAccount, exchangeMin,
	exchangeLimit, exchangeTax, exchangeFlow, exchangeFlag, spreadConfig int32, vipRange []int32, otherParam string,
	ccf *ClubConfig, verifyCodeType int32, thirdState map[int32]int32, customType int32, needDeviceInfo bool,
	needSameName bool, exchangeForceTax int32, exchangeGiveFlow int32, exchangeVer int32, exchangeBankMax int32,
	exchangeAlipayMax int32, dgHboCfg int32, PerBankNoLimitName int32, isCanUserBindPromoter bool,
	userBindPromoterPrize int32, spreadWinLose bool, exchangeMultiple int32, registerVerifyCodeSwitch bool, merchantKey string) *Platform {
	pltId := strconv.Itoa(int(id))
	p, ok := pm.Platforms[pltId]
	if !ok {
		p = pm.CreatePlatform(id, isolated)
	}
	oldIsolated := p.Isolated
	oldBindPromoter := p.IsCanUserBindPromoter
	oldDisabled := p.Disable
	//oldBindOption := p.BindOption
	p.Id = id
	p.Name = name
	p.IdStr = pltId
	p.ServiceUrl = url
	p.BindOption = bindOption
	p.ServiceFlag = serviceFlag
	p.CustomType = customType
	p.UpgradeAccountGiveCoin = upgradeAccountGiveCoin
	p.NewAccountGiveCoin = newAccountGiveCoin
	p.PerBankNoLimitAccount = perBankNoLimitAccount
	p.ExchangeMin = exchangeMin
	p.ExchangeLimit = exchangeLimit
	p.SpreadConfig = spreadConfig
	p.ExchangeTax = exchangeTax
	p.ExchangeVer = exchangeVer
	p.ExchangeForceTax = exchangeForceTax
	p.ExchangeFlow = exchangeFlow
	p.ExchangeFlag = exchangeFlag
	p.VipRange = vipRange
	p.OtherParams = otherParam
	p.VerifyCodeType = verifyCodeType
	p.RegisterVerifyCodeSwitch = registerVerifyCodeSwitch
	p.ThirdGameMerchant = thirdState
	//p.NeedDeviceInfo = needDeviceInfo
	p.NeedSameName = needSameName
	p.PerBankNoLimitName = PerBankNoLimitName
	p.ExchangeGiveFlow = exchangeGiveFlow
	p.ExchangeBankMax = exchangeBankMax
	p.ExchangeAlipayMax = exchangeAlipayMax
	p.DgHboConfig = dgHboCfg
	p.IsCanUserBindPromoter = isCanUserBindPromoter
	p.UserBindPromoterPrize = userBindPromoterPrize
	p.SpreadWinLose = spreadWinLose
	p.ExchangeMultiple = exchangeMultiple
	p.MerchantKey = merchantKey
	if p.ExchangeFlag&ExchangeFlag_Flow == 0 { //修正下
		p.ExchangeFlow = 0
	}
	if oldIsolated != isolated {
		pm.ChangeIsolated(name, isolated)
	}
	if oldDisabled != disable {
		pm.ChangeDisabled(name, disable)
	}
	if oldBindPromoter != isCanUserBindPromoter {
		pm.ChangeIsCanBindPromoter(name, isCanUserBindPromoter)
	}
	if ccf != nil {
		p.ClubConfig = ccf
	}
	//PlayerMgrSington.ModifyActSwitchToPlayer(name, oldBindOption != bindOption)
	return p
}

func (pm *PlatformMgr) UpsertPlatformGameConfig(pltId string, data *webapi_proto.GameFree) {
	p := pm.GetPlatform(pltId)
	if p != nil {
		pgc := p.PltGameCfg
		if pgc != nil {
			dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(data.DbGameFree.Id)
			if dbGameFree == nil {
				return
			}
			if data.GetDbGameFree() == nil { //数据容错
				data.DbGameFree = dbGameFree
			} else {
				CopyDBGameFreeField(dbGameFree, data.DbGameFree)
			}

			old, ok := pgc.games[data.DbGameFree.Id]
			pgc.games[data.DbGameFree.Id] = data
			found := false
			if ok && old != nil {
				if c, ok := pgc.cache[data.DbGameFree.Id]; ok {
					for i := 0; i < len(c); i++ {
						if c[i].DbGameFree.Id == data.DbGameFree.Id {
							c[i] = data
							found = true
							break
						}
					}
				}
			}
			if !found {
				pgc.cache[data.DbGameFree.Id] = append(pgc.cache[data.DbGameFree.Id], data)
			}
			if ok && old != nil && !CompareGameFreeConfigChged(old, data) {
				pm.OnPlatformConfigUpdate(p, old, data)
				pm.SyncChgedGameFree(p.IdStr, data)
			}
		}
	}
}

func (pm *PlatformMgr) SyncChgedGameFree(platform string, data *webapi_proto.GameFree) {
	packSgf := &login_proto.SCSyncGameFree{}
	gc := &login_proto.GameConfig{
		GameId:         proto.Int32(data.DbGameFree.GetGameId()),
		GameMode:       proto.Int32(data.DbGameFree.GetGameMode()),
		LogicId:        proto.Int32(data.DbGameFree.GetId()),
		State:          proto.Bool(data.Status),
		LimitCoin:      proto.Int32(data.DbGameFree.GetLimitCoin()),
		MaxCoinLimit:   proto.Int32(data.DbGameFree.GetMaxCoinLimit()),
		BaseScore:      proto.Int32(data.DbGameFree.GetBaseScore()),
		BetScore:       proto.Int32(data.DbGameFree.GetBetLimit()),
		OtherIntParams: data.DbGameFree.GetOtherIntParams(),
		MaxBetCoin:     data.DbGameFree.GetMaxBetCoin(),
		MatchMode:      proto.Int32(data.DbGameFree.GetMatchMode()),
	}
	if data.DbGameFree.GetLottery() != 0 {
		gc.LotteryCfg = data.DbGameFree.LotteryConfig
		//_, gl := LotteryMgrSington.FetchLottery(platform, param.DBGameFree.GetId(), param.DBGameFree.GetGameId())
		//if gl != nil {
		//	gc.LotteryCoin = proto.Int64(gl.Value)
		//}
	}
	packSgf.Data = append(packSgf.Data, gc)
	if len(packSgf.Data) > 0 {
		PlayerMgrSington.BroadcastMessageToPlatform(platform, int(login_proto.LoginPacketID_PACKET_SC_SYNCGAMEFREE), packSgf)
	}
}

func (pm *PlatformMgr) OnPlatformCreate(p *Platform) {
	for _, observer := range pm.Observers {
		observer.OnPlatformCreate(p)
	}
}

func (pm *PlatformMgr) OnPlatformConfigUpdate(p *Platform, oldCfg, newCfg *webapi_proto.GameFree) {
	for _, observer := range pm.Observers {
		observer.OnPlatformConfigUpdate(p, oldCfg, newCfg)
	}
}

func (pm *PlatformMgr) OnPlatformDestroyByGameFreeId(p *Platform, gameFreeId int32) {
	for _, observer := range pm.Observers {
		observer.OnPlatformDestroyByGameFreeId(p, gameFreeId)
	}
}

func (pm *PlatformMgr) GetPlatform(name string) *Platform {
	if p, exist := pm.Platforms[name]; exist {
		return p
	}
	return nil
}
func (pm *PlatformMgr) GetPlatformClubConfig(name string) *ClubConfig {
	if p, exist := pm.Platforms[name]; exist {
		return p.ClubConfig
	}
	return nil
}

func (pm *PlatformMgr) GetOrCreateScenesByHall(hallId int32) map[int]*Scene {
	if ss, exist := pm.ScenesByHall[hallId]; exist {
		return ss
	}
	ss := make(map[int]*Scene)
	pm.ScenesByHall[hallId] = ss
	return ss
}

func (pm *PlatformMgr) GetOrCreatePlayersByHall(hallId int32) map[int32]*Player {
	if pp, exist := pm.PlayersInHall[hallId]; exist {
		return pp
	}
	pp := make(map[int32]*Player)
	pm.PlayersInHall[hallId] = pp
	return pp
}

func (pm *PlatformMgr) PlayerEnterHall(player *Player, hallId int32) {
	if hallId != player.hallId {
		pm.PlayerLeaveHall(player)
	}
	p := pm.GetPlatform(player.Platform)
	if p == nil || !p.Isolated {
		pp := pm.GetOrCreatePlayersByHall(hallId)
		if pp != nil {
			player.hallId = hallId
			pp[player.SnId] = player
			dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(hallId)
			if dbGameFree != nil {
				dbGameRule := srvdata.PBDB_GameRuleMgr.GetData(dbGameFree.GetGameRule())
				if dbGameRule != nil {
					sp := GetScenePolicy(int(dbGameFree.GetGameId()), int(dbGameFree.GetGameMode()))
					if spd, ok := sp.(*ScenePolicyData); ok {
						playernum := spd.getPlayerNum(dbGameRule.GetParams())
						ss := pm.GetOrCreateScenesByHall(hallId)
						if ss != nil {
							inc := evaluateSceneIncCount(len(ss), len(pp), int(playernum))
							if inc > 0 {
								var scenes []*Scene
								for i := 0; i < inc; i++ {
									scene := pm.CreateNewScene(nil, hallId)
									if scene != nil {
										ss[scene.sceneId] = scene
										scenes = append(scenes, scene)
									}
								}
								pm.BroadcastRoomList(hallId, dbGameRule, scenes, true, pp, player.SnId)
							}
							if ss != nil {
								pack := &hall_proto.SCHallRoomList{
									HallId:   proto.Int32(hallId),
									GameId:   proto.Int32(dbGameRule.GetGameId()),
									GameMode: proto.Int32(dbGameRule.GetGameMode()),
									IsAdd:    proto.Bool(false),
									Params:   dbGameRule.GetParams(),
								}
								for _, scene := range ss {
									ri := &hall_proto.RoomInfo{
										RoomId:   proto.Int32(int32(scene.sceneId)),
										Starting: proto.Bool(scene.starting),
									}
									for _, p := range scene.players {
										ri.Players = append(ri.Players, p.CreateRoomPlayerInfoProtocol())
									}
									pack.Rooms = append(pack.Rooms, ri)
								}
								player.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_HALLROOMLIST), pack)
								//发送大厅人数
								pm.SendPlayerNum(dbGameRule.GetGameId(), player)
							}
						}
					}
				}
			}
			return
		}
	} else {
		p.PlayerEnter(player, hallId)
	}
}

func (pm *PlatformMgr) PlayerLeaveHall(player *Player) (hallId int32) {
	hallId = player.hallId
	if hallId != 0 {
		p := pm.GetPlatform(player.Platform)
		if p == nil || !p.Isolated {
			pp := pm.GetOrCreatePlayersByHall(hallId)
			if pp != nil {
				delete(pp, player.SnId)
				player.hallId = 0
			}
		} else {
			p.PlayerLeave(player)
		}
	}

	return
}

func (pm *PlatformMgr) BroadcastRoomList(hallId int32, dbGameRule *server_proto.DB_GameRule, scenes []*Scene, isAdd bool, players map[int32]*Player, exclude int32) {
	pack := &hall_proto.SCHallRoomList{
		HallId:   proto.Int32(hallId),
		GameId:   proto.Int32(dbGameRule.GetGameId()),
		GameMode: proto.Int32(dbGameRule.GetGameMode()),
		IsAdd:    proto.Bool(isAdd),
		Params:   dbGameRule.GetParams(),
	}

	var hallplaynum = make(map[int]int)

	for _, scene := range scenes {
		pack.Rooms = append(pack.Rooms, &hall_proto.RoomInfo{
			RoomId:   proto.Int32(int32(scene.sceneId)),
			Starting: proto.Bool(scene.starting),
		})
		sceneType := int(scene.dbGameFree.GetSceneType())
		hallplaynum[sceneType] = hallplaynum[sceneType] + len(scene.players) + scene.robotNum
	}

	for k, v := range hallplaynum {
		pack.HallData = append(pack.HallData, &hall_proto.HallInfo{
			SceneType: proto.Int32(int32(k)),
			PlayerNum: proto.Int32(int32(v)),
		})
	}
	pm.Broadcast(int(hall_proto.GameHallPacketID_PACKET_SC_HALLROOMLIST), pack, players, exclude)
}

func (pm *PlatformMgr) OnPlayerEnterScene(scene *Scene, player *Player) {
	if scene == nil || player == nil {
		return
	}

	if !scene.IsHallScene() {
		return
	}
	if scene.limitPlatform == nil {
		pp := pm.GetOrCreatePlayersByHall(scene.hallId)
		if pp != nil {
			delete(pp, player.SnId)
			pack := &hall_proto.SCRoomPlayerEnter{
				RoomId: proto.Int32(int32(scene.sceneId)),
				Player: player.CreateRoomPlayerInfoProtocol(),
			}
			pm.Broadcast(int(hall_proto.GameHallPacketID_PACKET_SC_ROOMPLAYERENTER), pack, pp, player.SnId)

			//人数信息
			gameid := int32(scene.gameId)
			if _, exist := pm.GamePlayerNum[gameid]; !exist {
				pm.GamePlayerNum[gameid] = &PlatformGamePlayerNum{
					Nums:  make(map[int32]int),
					Dirty: true,
				}
			}
			if nums, exist := pm.GamePlayerNum[gameid]; exist {
				sceneType := scene.dbGameFree.GetSceneType()
				nums.Nums[sceneType] = nums.Nums[sceneType] + 1
				nums.Dirty = true
				pm.dirty = true
			}
		}
	} else {
		scene.limitPlatform.OnPlayerEnterScene(scene, player)
	}
}

func (pm *PlatformMgr) OnChangeSceneState(scene *Scene, startclose bool) {
	if scene == nil {
		return
	}

	if !scene.IsHallScene() {
		return
	}
	if scene.limitPlatform == nil {
		pp := pm.GetOrCreatePlayersByHall(scene.hallId)
		if pp != nil {
			pack := &hall_proto.SCRoomStateChange{
				RoomId:   proto.Int32(int32(scene.sceneId)),
				Starting: proto.Bool(startclose),
				State:    proto.Int32(scene.state),
			}
			pm.Broadcast(int(hall_proto.GameHallPacketID_PACKET_SC_ROOMSTATECHANG), pack, pp, 0)
		}
		if scene.starting {
			scene.replayCode = SceneMgrSington.AllocReplayCode()
			logger.Logger.Trace("游戏开始------", scene.gameId, scene.sceneId, scene.replayCode, scene.currRound)
		} else {
			logger.Logger.Trace("游戏结束------", scene.gameId, scene.sceneId, scene.replayCode, scene.currRound)
		}
	}
}

func (pm *PlatformMgr) OnPlayerLeaveScene(scene *Scene, player *Player) {
	if scene == nil || player == nil {
		return
	}

	if !scene.IsHallScene() {
		return
	}
	if scene.limitPlatform == nil {
		pp := pm.GetOrCreatePlayersByHall(scene.hallId)
		if pp != nil {
			pack := &hall_proto.SCRoomPlayerLeave{
				RoomId: proto.Int32(int32(scene.sceneId)),
				Pos:    proto.Int32(int32(player.pos)),
			}
			pm.Broadcast(int(hall_proto.GameHallPacketID_PACKET_SC_ROOMPLAYERLEAVE), pack, pp, player.SnId)
			//人数变化信息
			gameid := int32(scene.gameId)
			if nums, exist := pm.GamePlayerNum[gameid]; exist {
				sceneType := scene.dbGameFree.GetSceneType()
				if n, exist := nums.Nums[sceneType]; exist && n > 0 {
					nums.Nums[sceneType] = n - 1
					nums.Dirty = true
					pm.dirty = true
				}
			}
		}
	} else {
		scene.limitPlatform.OnPlayerLeaveScene(scene, player)
	}
}

func (pm *PlatformMgr) CreateNewScene(p *Platform, hallId int32) *Scene {

	conf := p.PltGameCfg.GetGameCfg(hallId)

	dbGameFree := conf.DbGameFree
	if conf.GroupId != 0 {
		pgg := PlatformGameGroupMgrSington.GetGameGroup(conf.GroupId)
		if pgg != nil {
			dbGameFree = pgg.DbGameFree
		}
	}
	if dbGameFree == nil {
		return nil
	}
	dbGameRule := srvdata.PBDB_GameRuleMgr.GetData(dbGameFree.GetGameRule())
	if dbGameRule == nil {
		return nil
	}
	sceneId := SceneMgrSington.GenOneHallSceneId()
	gameId := int(dbGameRule.GetGameId())
	scenetype := int(dbGameFree.GetSceneType())
	gs := GameSessMgrSington.GetMinLoadSess(gameId)
	if gs != nil {
		gameMode := dbGameRule.GetGameMode()
		params := dbGameRule.GetParams()
		var limit *Platform
		if p != nil && p.Isolated {
			limit = p
		}

		scene := SceneMgrSington.CreateScene(0, 0, sceneId, gameId, int(gameMode), scenetype, -1, -1, params, gs, limit, conf.GroupId, dbGameFree, hallId)
		if scene != nil {
			scene.hallId = hallId
			return scene
		}
	}
	return nil
}

func (pm *PlatformMgr) OnDestroyScene(scene *Scene) {
	if scene == nil {
		return
	}

	if !scene.IsHallScene() {
		return
	}
	if scene.limitPlatform == nil {
		ss := pm.GetOrCreateScenesByHall(scene.hallId)
		if ss != nil {
			delete(ss, scene.sceneId)
			pp := pm.GetOrCreatePlayersByHall(scene.hallId)
			if pp != nil {
				pack := &hall_proto.SCDestroyRoom{
					RoomId:    proto.Int32(int32(scene.sceneId)),
					OpRetCode: hall_proto.OpResultCode_Game_OPRC_Sucess_Game,
					IsForce:   proto.Int32(1),
				}
				pm.Broadcast(int(hall_proto.GameHallPacketID_PACKET_SC_DESTROYROOM), pack, pp, 0)
			}
		}
	} else {
		scene.limitPlatform.OnDestroyScene(scene)
	}
}

func (pm *PlatformMgr) Broadcast(packetid int, packet interface{}, players map[int32]*Player, exclude int32) {
	mgs := make(map[*netlib.Session][]*srvlibproto.MCSessionUnion)
	for _, p := range players {
		if p != nil && p.gateSess != nil && p.IsOnLine() && exclude != p.SnId {
			mgs[p.gateSess] = append(mgs[p.gateSess], &srvlibproto.MCSessionUnion{
				Mccs: &srvlibproto.MCClientSession{
					SId: p.sid,
				},
			})
		}
	}
	for gateSess, v := range mgs {
		if gateSess != nil && len(v) != 0 {
			pack, err := MulticastMaker.CreateMulticastPacket(packetid, packet, v...)
			if err == nil {
				gateSess.Send(int(srvlibproto.SrvlibPacketID_PACKET_SS_MULTICAST), pack)
			}
		}
	}
}

func (pm *PlatformMgr) ChangeIsolated(name string, isolated bool) {
	if p, exist := pm.Platforms[name]; exist {
		if p.ChangeIsolated(isolated) {
			pm.OnPlatformChangeIsolated(p, isolated)
		}
	}
}

func (pm *PlatformMgr) OnPlatformChangeIsolated(p *Platform, isolated bool) {
	for _, observer := range pm.Observers {
		observer.OnPlatformChangeIsolated(p, isolated)
	}
}

func (pm *PlatformMgr) ChangeDisabled(name string, disable bool) {
	if p, exist := pm.Platforms[name]; exist {
		if p.ChangeDisabled(disable) {
			pm.OnPlatformChangeDisabled(p, disable)
		}
	}
}
func (pm *PlatformMgr) ChangeIsCanBindPromoter(name string, disable bool) {
	for _, v := range PlayerMgrSington.playerSnMap {
		if v != nil && v.IsOnLine() && v.Platform == name {
			v.SendPlatformCanUsePromoterBind()
		}
	}
}
func (pm *PlatformMgr) OnPlatformChangeDisabled(p *Platform, disable bool) {
	for _, observer := range pm.Observers {
		observer.OnPlatformChangeDisabled(p, disable)
	}
}

func (pm *PlatformMgr) PlayerLogin(p *Player) {
	if p.IsRob {
		return
	}
	platform := pm.GetPlatform(p.Platform)
	if platform != nil {
		platform.PlayerLogin(p)
	}
}

func (pm *PlatformMgr) PlayerLogout(p *Player) {
	platform := pm.GetPlatform(p.Platform)
	if platform != nil && platform.Isolated {
		platform.PlayerLogout(p)
	} else {
		pp := pm.GetOrCreatePlayersByHall(p.hallId)
		if pp != nil {
			delete(pp, p.SnId)
		}
	}
}

func (this *PlatformMgr) SendPlayerNum(gameid int32, pp *Player) {
	if nums, exist := this.GamePlayerNum[gameid]; exist {
		pack := &hall_proto.HallPlayerNum{}
		for k, v := range nums.Nums {
			pack.HallData = append(pack.HallData, &hall_proto.HallInfo{
				SceneType: proto.Int32(k),
				PlayerNum: proto.Int32(int32(v)),
			})
		}
		pp.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_HALLPLAYERNUM), pack)
	}
}

func (this *PlatformMgr) BroadcastPlayerNum() {
	if this.dirty {
		this.dirty = false
		for gameid, nums := range this.GamePlayerNum {
			if nums.Dirty {
				nums.Dirty = false
				pack := &hall_proto.HallPlayerNum{}
				for k, v := range nums.Nums {
					pack.HallData = append(pack.HallData, &hall_proto.HallInfo{
						SceneType: proto.Int32(k),
						PlayerNum: proto.Int32(int32(v)),
					})
				}
				for hallid, players := range this.PlayersInHall {
					dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(hallid)
					if dbGameFree != nil && dbGameFree.GetGameId() == gameid {
						this.Broadcast(int(hall_proto.GameHallPacketID_PACKET_SC_HALLPLAYERNUM), pack, players, 0)
					}
				}
			}
		}
	}
}

func (this *PlatformMgr) GetPackageTag(tag string) *webapi_proto.AppInfo {
	logger.Logger.Tracef("Get %v platform.", tag)
	if pt, ok := this.PackageList[tag]; ok {
		return pt
	}
	return nil
}

func (this *PlatformMgr) GetPlatformByPackageTag(tag string) (int32, int32, int32, int32, int32) {
	logger.Logger.Tracef("Get %v platform.", tag)
	if pt, ok := this.PackageList[tag]; ok {
		return pt.PlatformId, 0, 0, 0, 0
	} else {
		return int32(Default_PlatformInt), 0, 0, 0, 0
	}
}
func (this *PlatformMgr) GetPlatformUpgradeAccountGiveCoinByPackageTag(tag string) int32 {
	platform, _, _, _, _ := this.GetPlatformByPackageTag(tag)
	return this.GetPlatformUpgradeAccountGiveCoinByPlatform(strconv.Itoa(int(platform)))
}
func (this *PlatformMgr) GetPlatformUpgradeAccountGiveCoinByPlatform(platform string) int32 {
	platformData := this.GetPlatform(platform)
	if platformData == nil {
		return 0
	}
	return platformData.UpgradeAccountGiveCoin
}
func (this *PlatformMgr) CheckPackageTag(tag string) bool {
	logger.Logger.Tracef("Check %v platform.", tag)
	if _, ok := this.PackageList[tag]; ok {
		return ok
	} else {
		return false
	}
}

func (this *PlatformMgr) GetPlatformGameConfig(pltId string) map[int32]*webapi_proto.GameFree {
	platfrom := this.GetPlatform(pltId)
	if platfrom == nil {
		return nil
	}

	data := make(map[int32]*webapi_proto.GameFree)
	for id, val := range platfrom.PltGameCfg.games {
		if !val.Status || !this.GameStatus[id] {
			continue
		}

		if val.GroupId != 0 {
			cfg := PlatformGameGroupMgrSington.GetGameGroup(val.GroupId)
			if cfg != nil {
				temp := &webapi_proto.GameFree{
					GroupId:    cfg.GetId(),
					Status:     val.Status,
					DbGameFree: cfg.GetDbGameFree(),
				}
				data[id] = temp
			}
		} else {
			data[id] = val
		}
	}
	return data
}

func (this *PlatformMgr) GetPlatformDgAgentConfig(platform string) (string, string, string) {

	info := this.GetPlatform(platform)
	if info == nil {
		return "", "", ""
	}

	switch info.DgHboConfig {
	case 0:
		//用默认的方法
		return model.GetDgConfigByPlatform(platform)
	case 1:
		//使用dg的配置
		return model.OnlyGetDgConfigByPlatform(platform)
	case 2:
		//使用hbo的配置
		return model.OnlyGetHboConfigByPlatform(platform)
	}

	return "", "", ""
}

func (this *PlatformMgr) CheckGameState(pltId string, gamefreeId int32) bool {

	platfrom := this.GetPlatform(pltId)
	if platfrom == nil {
		return false
	}

	if !this.GameStatus[gamefreeId] {
		return false
	}

	cfg := platfrom.PltGameCfg.GetGameCfg(gamefreeId)
	if cfg == nil {
		return false
	}

	return cfg.Status
}

// 根据三方id，找到第一个gamefreeid用的
func (this *PlatformMgr) GetGameConfigByThird(pltid string, thrird int32) *webapi_proto.GameFree {
	return this.GetGameConfig(pltid, thrird)
}

func (this *PlatformMgr) GetGameConfig(pltid string, gamefreeId int32) *webapi_proto.GameFree {
	platfrom := this.GetPlatform(pltid)
	if platfrom == nil {
		return nil
	}

	if !this.GameStatus[gamefreeId] {
		return nil
	}

	return platfrom.PltGameCfg.GetGameCfg(gamefreeId)
}

func (this *PlatformMgr) GetPlatformByGroup(groupid int32) string {
	//configId := []int32{}
	//for _, cfg := range this.PlatConList {
	//	for _, value := range cfg.Param {
	//		if value.GroupId == groupid {
	//			configId = append(configId, cfg.Id)
	//		}
	//	}
	//}
	//platforms := make(map[string]bool)
	//for _, id := range configId {
	//	for _, value := range this.Platforms {
	//		if value.ConfigId == id {
	//			platforms[value.Name] = true
	//		}
	//	}
	//}
	//selPlatform := ""
	//for key, _ := range platforms {
	//	selPlatform = key
	//	break
	//}
	return ""
}

/*
 * 平台模块
 */
func (this *PlatformMgr) ModuleName() string {
	return "PlatformMgr"
}
func (this *PlatformMgr) Init() {
}

func (this *PlatformMgr) LoadPlatformData() {
	//构建默认的平台数据
	this.CreateDefaultPlatform()
	//获取平台数据 platform_list

	//不使用etcd的情况下走api获取
	if !model.GameParamData.UseEtcd {
		buf, err := webapi.API_GetPlatformData(common.GetAppId())
		if err == nil {
			ar := webapi_proto.ASPlatformInfo{}
			err = proto.Unmarshal(buf, &ar)
			if err == nil && ar.Tag == webapi_proto.TagCode_SUCCESS {
				for _, value := range ar.Platforms {
					platform := this.CreatePlatform(value.Id, value.Isolated)
					platform.Name = value.PlatformName
					//platform.ConfigId = value.ConfigId
					platform.Disable = value.Disabled
					platform.ServiceUrl = value.CustomService
					platform.CustomType = value.CustomType
					platform.BindOption = value.BindOption
					//platform.ServiceFlag = value.ServiceFlag
					platform.UpgradeAccountGiveCoin = value.UpgradeAccountGiveCoin
					platform.NewAccountGiveCoin = value.NewAccountGiveCoin       //新账号奖励金币
					platform.PerBankNoLimitAccount = value.PerBankNoLimitAccount //同一银行卡号绑定用户数量限制
					platform.ExchangeMin = value.ExchangeMin
					platform.ExchangeLimit = value.ExchangeLimit
					platform.ExchangeTax = value.ExchangeTax
					platform.ExchangeFlow = value.ExchangeFlow
					platform.ExchangeVer = value.ExchangeVer
					platform.ExchangeFlag = value.ExchangeFlag
					platform.ExchangeForceTax = value.ExchangeForceTax
					platform.ExchangeGiveFlow = value.ExchangeGiveFlow
					platform.VipRange = value.VipRange
					//platform.OtherParams = value.OtherParams
					platform.SpreadConfig = value.SpreadConfig
					//platform.RankSwitch = value.Leaderboard
					//platform.ClubConfig = value.ClubConfig
					platform.VerifyCodeType = value.VerifyCodeType
					//platform.RegisterVerifyCodeSwitch = value.RegisterVerifyCodeSwitch
					for _, v := range value.ThirdGameMerchant {
						platform.ThirdGameMerchant[v.Id] = v.Merchant
					}
					//platform.NeedDeviceInfo = value.NeedDeviceInfo
					platform.NeedSameName = value.NeedSameName
					platform.PerBankNoLimitName = value.PerBankNoLimitName
					platform.ExchangeAlipayMax = value.ExchangeAlipayMax
					platform.ExchangeBankMax = value.ExchangeBankMax
					//platform.DgHboConfig = value.DgHboConfig
					platform.IsCanUserBindPromoter = value.IsCanUserBindPromoter
					platform.UserBindPromoterPrize = value.UserBindPromoterPrize
					//platform.SpreadWinLose = value.SpreadWinLose
					platform.ExchangeMultiple = value.ExchangeMultiple
					if platform.ExchangeFlag&ExchangeFlag_Flow == 0 { //修正下
						platform.ExchangeFlow = 0
					}
					platform.MerchantKey = value.MerchantKey
				}
				logger.Logger.Trace("Create platform")
			} else {
				logger.Logger.Error("Unmarshal platform data error:", err)
			}
		} else {
			logger.Logger.Error("Get platfrom data error:", err)
		}
	} else {
		EtcdMgrSington.InitPlatform()
	}
}

func (this *PlatformMgr) LoadPlatformConfig() {
	//不使用etcd的情况下走api获取

	//获取平台详细信息 game_config_list
	if !model.GameParamData.UseEtcd {
		logger.Logger.Trace("API_GetPlatformConfigData")
		buf, err := webapi.API_GetPlatformConfigData(common.GetAppId())
		if err == nil {
			pcdr := &webapi_proto.ASGameConfig{}
			err = proto.Unmarshal(buf, pcdr)
			if err == nil && pcdr.Tag == webapi_proto.TagCode_SUCCESS {
				platGameCfgs := pcdr.GetConfigs()

				//遍历所有平台配置
				for _, pgc := range platGameCfgs {
					platormId := pgc.GetPlatformId()
					platform := this.GetPlatform(strconv.Itoa(int(platormId)))
					if platform == nil {
						continue
					}

					for _, config := range pgc.GetDbGameFrees() {
						dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(config.DbGameFree.Id)
						if dbGameFree == nil {
							logger.Logger.Error("Platform config data error logic id:", config)
							continue
						}
						platform.PltGameCfg.games[config.DbGameFree.Id] = config
						if config.GetDbGameFree() == nil { //数据容错
							config.DbGameFree = dbGameFree
						} else {
							CopyDBGameFreeField(dbGameFree, config.DbGameFree)
						}
						logger.Logger.Info("PlatformGameConfig data:", config.DbGameFree.Id)
					}
					platform.PltGameCfg.RecreateCache()
				}
			} else {
				logger.Logger.Error("Unmarshal platform config data error:", err, string(buf))
			}
		} else {
			logger.Logger.Error("Get platfrom config data error:", err)
		}
	} else {
		EtcdMgrSington.InitPlatformGameConfig()
	}
}

type PackageListApi struct {
	Tag            string //android包名或者ios标记
	Platform       int32  //所属平台
	ChannelId      int32  //渠道ID
	PromoterId     int32  //推广员ID
	PromoterTree   int32  //无级推广树
	SpreadTag      int32  //全民包标识 0:普通包 1:全民包
	OpenInstallTag int32  //是否是openinstall包 0:不是 1:是
	Status         int32  //状态
	AppStore       int32  //是否是苹果商店包 0:不是 1:是
	ExchangeFlag   int32  //兑换标记 0 关闭包返利 1打开包返利 受平台配置影响
	ExchangeFlow   int32  //兑换比例
	IsForceBind    int32
	TagKey         int32 `json:"SpecialChannel"` //包标识区分
}

func (this *PlatformMgr) LoadPlatformPackage() {
	EtcdMgrSington.InitPlatformPackage()
}

func (this *PlatformMgr) CommonNotice() {

	if model.GameParamData.UseEtcd {
		EtcdMgrSington.InitCommonNotice()
	} else {

	}
}

func (this *PlatformMgr) GameMatchDate() {

	if model.GameParamData.UseEtcd {
		EtcdMgrSington.InitGameMatchDate()
	} else {

	}
}

func (this *PlatformMgr) GetCommonNotice(plt string) *webapi_proto.CommonNoticeList {
	logger.Logger.Tracef("GetCommonNotice %v platform.", plt)
	if pt, ok := this.CommonNotices[plt]; ok {
		now := time.Now().Unix()
		re := &webapi_proto.CommonNoticeList{
			Platform: pt.Platform,
		}
		for _, v := range pt.List { // 过滤
			if v.GetEndTime() > now && v.GetStartTime() < now {
				re.List = append(re.List, v)
			}
		}
		return re
	}
	return nil
}

func (this *PlatformMgr) Update() {
	//撮合
	/*var hallplaynum = make(map[int]int)
	for _,pm := range this.Platforms{
		logger.Logger.Trace(pm.Name)
		for _, gc := range model.GameAllConfig {
			if _, exist := this.ScenesByHall[gc.LogicId]; exist {
				for _, ss := range this.ScenesByHall[gc.LogicId] {
					hallplaynum[ss.sceneType] = ss.playerNum + ss.robotNum
				}

			}
		}
	}*/
	this.BroadcastPlayerNum()
	for _, pm := range this.Platforms {
		pm.BroadcastPlayerNum()
	}
}
func (this *PlatformMgr) Shutdown() {
	module.UnregisteModule(this)
}

func (this *PlatformMgr) InterestClockEvent() int {
	return 1<<CLOCK_EVENT_HOUR | 1<<CLOCK_EVENT_DAY
}
func (this *PlatformMgr) OnDayTimer() {
	for _, platform := range this.Platforms {
		utils.CatchPanic(func() {
			platform.OnDayTimer()
		})
	}
}

func (this *PlatformMgr) OnHourTimer() {
	for _, platform := range this.Platforms {
		utils.CatchPanic(func() {
			platform.OnHourTimer()
		})
	}
}

// 这个函数不清楚啥作用别瞎改，该一行代码跳坑5小时
func CopyDBGameFreeField(src, dst *server_proto.DB_GameFree) {
	dst.Id = src.Id
	dst.Name = src.Name
	dst.Title = src.Title
	dst.ShowType = src.ShowType
	dst.SubShowType = src.SubShowType
	dst.Flag = src.Flag
	dst.GameRule = src.GameRule
	dst.TestTakeCoin = src.TestTakeCoin
	dst.SceneType = src.SceneType
	dst.GameType = src.GameType
	dst.GameId = src.GameId
	dst.GameMode = src.GameMode
	dst.ShowId = src.ShowId
	dst.ServiceFee = src.ServiceFee
	dst.Turn = src.Turn
	dst.BetDec = src.BetDec
	//dst.CorrectNum = src.CorrectNum
	//dst.CorrectRate = src.CorrectRate
	//dst.Deviation = src.Deviation
	//dst.Ready = src.Ready
	dst.Ai = src.Ai
	dst.Jackpot = src.Jackpot
	//dst.ElementsParams = src.ElementsParams
	//dst.OtherElementsParams = src.OtherElementsParams
	//dst.DownRiceParams = src.DownRiceParams
	//dst.InitValue = src.InitValue
	//dst.LowerLimit = src.LowerLimit
	//dst.UpperLimit = src.UpperLimit
	//dst.UpperOffsetLimit = src.UpperOffsetLimit
	//dst.MaxOutValue = src.MaxOutValue
	//dst.ChangeRate = src.ChangeRate
	//dst.MinOutPlayerNum = src.MinOutPlayerNum
	//dst.UpperLimitOfOdds = src.UpperLimitOfOdds
	if len(dst.RobotNumRng) < 2 {
		dst.RobotNumRng = src.RobotNumRng
	}
	//dst.SameIpLimit = src.SameIpLimit
	//dst.BaseRate = src.BaseRate
	//dst.CtroRate = src.CtroRate
	//dst.HardTimeMin = src.HardTimeMin
	//dst.HardTimeMax = src.HardTimeMax
	//dst.NormalTimeMin = src.NormalTimeMin
	//dst.NormalTimeMax = src.NormalTimeMax
	//dst.EasyTimeMin = src.EasyTimeMin
	//dst.EasyTimeMax = src.EasyTimeMax
	//dst.EasrierTimeMin = src.EasrierTimeMin
	//dst.EasrierTimeMax = src.EasrierTimeMax
	dst.GameType = src.GameType
	dst.GameDif = src.GameDif
	dst.GameClass = src.GameClass
	dst.PlatformName = src.PlatformName
	if len(dst.MaxBetCoin) == 0 {
		dst.MaxBetCoin = src.MaxBetCoin
	}

	//预创建房间数量走后台配置
	//dst.CreateRoomNum = src.CreateRoomNum
	//后台功能做上后，优先使用后台的配置，默认直接用表格种配好的
	//if !model.GameParamData.MatchTrueManUseWeb {
	//	dst.MatchTrueMan = src.MatchTrueMan
	//}

	////默认走本地配置
	//if dst.PlayerWaterRate == nil {
	//	tv := src.GetPlayerWaterRate()
	//	dst.PlayerWaterRate = &tv
	//}
	//
	//if dst.BetWaterRate == nil {
	//	tv := src.GetBetWaterRate()
	//	dst.BetWaterRate = &tv
	//}

}

// 拉取全局游戏状态
func (this *PlatformMgr) LoadGlobalGameStatus() {
	//不使用etcd的情况下走api获取
	if model.GameParamData.UseEtcd {
		EtcdMgrSington.InitGameGlobalStatus()
	} else {
		//获取全局游戏开关
		logger.Logger.Trace("API_GetGlobalGameStatus")
		buf, err := webapi.API_GetGlobalGameStatus(common.GetAppId())
		if err == nil {
			as := webapi_proto.ASGameConfigGlobal{}
			err = proto.Unmarshal(buf, &as)
			if err == nil && as.Tag == webapi_proto.TagCode_SUCCESS {
				if as.GetGameStatus() != nil {
					status := as.GetGameStatus().GetGameStatus()
					for _, v := range status {
						gameId := v.GetGameId()
						status := v.GetStatus()
						PlatformMgrSington.GameStatus[gameId] = status
					}
				}
			} else {
				logger.Logger.Error("Unmarshal GlobalGameStatus data error:", err)
			}
		} else {
			logger.Logger.Error("Get GlobalGameStatus data error:", err)
		}
	}
}

// 后台修改游戏状态（开启、关闭）
func (this *PlatformMgr) ChangeGameStatus(cfgs []*hall_proto.GameConfig1) {
	////有需要关闭的游戏时，关闭当前正在进行的游戏
	//if len(closeGameId) > 0 {
	//	for _, gameFreeId := range closeGameId {
	//		for _, p := range this.Platforms {
	//			this.OnPlatformDestroyByGameFreeId(p, gameFreeId)
	//		}
	//	}
	//}

	//向所有玩家发送状态变化协议
	pack := &hall_proto.SCChangeGameStatus{}

	for _, p := range this.Platforms {
		for _, cfg := range cfgs {
			gameFreeId := cfg.LogicId
			data, ok := p.PltGameCfg.games[gameFreeId]
			if data != nil && ok && data.Status { //自身有这个游戏，且处于开启状态
				if !cfg.Status { //全局关，销毁游戏
					pack.GameCfg = append(pack.GameCfg, cfg)
					this.OnPlatformDestroyByGameFreeId(p, gameFreeId)
				} else {
					lgc := &hall_proto.GameConfig1{
						LogicId:        proto.Int32(data.DbGameFree.Id),
						LimitCoin:      proto.Int32(data.DbGameFree.GetLimitCoin()),
						MaxCoinLimit:   proto.Int32(data.DbGameFree.GetMaxCoinLimit()),
						BaseScore:      proto.Int32(data.DbGameFree.GetBaseScore()),
						BetScore:       proto.Int32(data.DbGameFree.GetBetLimit()),
						OtherIntParams: data.DbGameFree.GetOtherIntParams(),
						MaxBetCoin:     data.DbGameFree.GetMaxBetCoin(),
						MatchMode:      proto.Int32(data.DbGameFree.GetMatchMode()),
						Status:         data.Status,
					}
					if data.DbGameFree.GetLottery() != 0 { //彩金池
						//lgc.LotteryCfg = data.DBGameFree.LotteryConfig
						//_, gl := LotteryMgrSington.FetchLottery(plf, data.DBGameFree.GetId(), data.DBGameFree.GetGameId())
						//if gl != nil {
						//	lgc.LotteryCoin = proto.Int64(gl.Value)
						//}
					}
					pack.GameCfg = append(pack.GameCfg, lgc)
				}
			}
		}

		if len(pack.GameCfg) > 0 {
			PlayerMgrSington.BroadcastMessageToPlatform(p.IdStr, int(hall_proto.GameHallPacketID_PACKET_SC_CHANGEGAMESTATUS), pack)
		}
	}

}

func init() {
	ClockMgrSington.RegisteSinker(PlatformMgrSington)
	module.RegisteModule(PlatformMgrSington, time.Second*2, 0)

	RegisteParallelLoadFunc("拉取平台数据", func() error {
		PlatformMgrSington.LoadPlatformData()
		return nil
	})

	RegisteParallelLoadFunc("拉取平台游戏开关", func() error {
		PlatformMgrSington.LoadGlobalGameStatus()
		return nil
	})

	RegisteParallelLoadFunc("拉取平台游戏详细配置", func() error {
		PlatformMgrSington.LoadPlatformConfig()
		return nil
	})

	RegisteParallelLoadFunc("平台包数据", func() error {
		PlatformMgrSington.LoadPlatformPackage()
		return nil
	})

	//使用并行加载
	RegisteParallelLoadFunc("平台公告", func() error {
		PlatformMgrSington.CommonNotice()
		return nil
	})

	//使用并行加载
	RegisteParallelLoadFunc("平台比赛", func() error {
		PlatformMgrSington.GameMatchDate()
		return nil
	})

	RegisteParallelLoadFunc("平台游戏组数据", func() error {
		PlatformGameGroupMgrSington.LoadGameGroup()
		return nil
	})

	//使用并行加载
	RegisteParallelLoadFunc("平台通知", func() error {
		HorseRaceLampMgrSington.InitHorseRaceLamp()
		return nil
	})

	//使用并行加载
	RegisteParallelLoadFunc("平台邮件", func() error {
		MsgMgrSington.InitMsg()
		//msgs, err := model.GetSubscribeMessage()
		//if err == nil {
		//	MsgMgrSington.init(msgs)
		//}
		return nil
	})
}
