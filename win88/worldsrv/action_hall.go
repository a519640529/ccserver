package main

//type CSRebateListPacketFactory struct {
//}
//
//type CSRebateListHandler struct {
//}
//
//func (this *CSRebateListPacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSRebateList{}
//	return pack
//}
//
//func (this *CSRebateListHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSRebateListHandler Process recv ", data)
//
//	p := PlayerMgrSington.GetPlayer(sid)
//	if p == nil {
//		logger.Logger.Errorf("CSRankInfoHandler.Process p is nil")
//		return nil
//	}
//	if p.layered[common.ActId_RebateTask] {
//		logger.Logger.Errorf("CSRankInfoHandler.Process p.layered[common.ActId_RebateTask]")
//		return nil
//	}
//	_, ok := PlatformMgrSington.Platforms[p.Platform]
//	if !ok {
//		logger.Logger.Errorf("CSRankInfoHandler.Process snid %v platform %s is not find in PlatformMgrSington.Platforms", p.SnId, p.Platform)
//		return nil
//	}
//	logger.Logger.Tracef("CSRebateListHandler platform %v snid %v", p.Platform, p.SnId)
//	p.ClearRebate()
//	SendRebateList(p)
//	return nil
//}
//
//func SendRebateList(p *Player) {
//	totalCoin := int64(0)
//	var ThirdRebateList []*gamehall.RebateInfo
//	var RebateList []*gamehall.RebateInfo
//
//	rebateTask := RebateInfoMgrSington.rebateTask[p.Platform]
//	if rebateTask != nil && rebateTask.RebateSwitch {
//		//////////////////////////////第三方///////
//		for k, cfg := range rebateTask.RebateGameThirdCfg {
//			rio := new(gamehall.RebateInfo)
//			if cfg.ThirdShowName == "" {
//				rio.Platform = proto.String(k)
//			} else {
//				rio.Platform = proto.String(cfg.ThirdShowName)
//			}
//
//			rio.Platform = proto.String(k)
//			var vbt int64 = 0
//			if info, ok := p.RebateData[k]; ok {
//				vbt = info.TodayRebateCoin
//				if rebateTask.ReceiveMode == 0 {
//					totalCoin += info.TodayRebateCoin + info.TotalHaveRebateCoin
//				} else {
//					totalCoin += info.YesterdayRebateCoin + info.TotalHaveRebateCoin
//				}
//			}
//			rio.ValidBetTotal = proto.Int64(vbt)
//			ThirdRebateList = append(ThirdRebateList, rio)
//		}
//
//		//排序/////////////////////////
//		if ThirdRebateList != nil {
//			sort.Slice(ThirdRebateList, func(i, j int) bool {
//				if ThirdRebateList[i].GetPlatform() < ThirdRebateList[j].GetPlatform() {
//					return true
//				}
//				return false
//			})
//		}
//
//		//博乐棋牌统计/////////////////////////
//		ri := new(gamehall.RebateInfo)
//		ri.Platform = proto.String("博乐棋牌")
//		var bole int64 = 0
//		for k, _ := range rebateTask.RebateGameCfg {
//			if data, ok := p.RebateData[k]; ok {
//				bole += data.TodayRebateCoin
//				if rebateTask.ReceiveMode == 0 {
//					totalCoin += data.TodayRebateCoin + data.TotalHaveRebateCoin
//				} else {
//					totalCoin += data.YesterdayRebateCoin + data.TotalHaveRebateCoin
//				}
//			}
//		}
//
//		ri.ValidBetTotal = proto.Int64(bole)
//		RebateList = append(RebateList, ri)
//		RebateList = append(RebateList, ThirdRebateList...)
//	}
//
//	pack := &gamehall.SCRebateList{
//		RebateList:      RebateList,
//		RebateTotalCoin: proto.Int64(totalCoin),
//	}
//	proto.SetDefaults(pack)
//	logger.Logger.Trace("SCRebateList: ", pack)
//	p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_REBATE_LIST), pack)
//}
//
//type CSReceiveRebatePacketFactory struct {
//}
//
//type CSReceiveRebateHandler struct {
//}
//
//func (this *CSReceiveRebatePacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSReceiveRebate{}
//	return pack
//}
//
//func (this *CSReceiveRebateHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSReceiveRebateHandler Process recv ", data)
//
//	p := PlayerMgrSington.GetPlayer(sid)
//	if p == nil {
//		logger.Logger.Errorf("CSRankInfoHandler.Process p is nil")
//		return nil
//	}
//	if p.layered[common.ActId_RebateTask] {
//		logger.Logger.Errorf("CSRankInfoHandler.Process p.layered[common.ActId_RebateTask]")
//		return nil
//	}
//
//	_, ok := PlatformMgrSington.Platforms[p.Platform]
//	if !ok {
//		logger.Logger.Errorf("CSRankInfoHandler.Process snid %v platform %s is not find in PlatformMgrSington.Platforms", p.SnId, p.Platform)
//		return nil
//	}
//
//	logger.Logger.Tracef("CSReceiveRebateHandler platform %v snid %v", p.Platform, p.SnId)
//
//	p.ClearRebate()
//
//	isFToday := false
//	var total int64 = 0
//	rebateTask := RebateInfoMgrSington.rebateTask[p.Platform]
//	//有些平台可能没有返利信息，就不要往下执行了
//	if rebateTask == nil {
//		logger.Logger.Warnf("CSRankInfoHandler.Process snid %v rebateTask is nil", p.SnId)
//		return nil
//	}
//	if rebateTask.RebateSwitch && p.RebateData != nil {
//		if rebateTask.ReceiveMode == 0 {
//			isFToday = true
//			for _, v := range p.RebateData {
//				total += v.TodayRebateCoin + v.TotalHaveRebateCoin
//				v.TotalRebateCoin += v.TodayRebateCoin
//				v.TodayRebateCoin = 0
//				v.ValidBetTotal = 0
//				v.TotalHaveRebateCoin = 0
//			}
//		} else {
//			for _, v := range p.RebateData {
//				total += v.YesterdayRebateCoin + v.TotalHaveRebateCoin
//				v.TotalRebateCoin += v.YesterdayRebateCoin
//				v.YesterdayRebateCoin = 0
//				v.TotalHaveRebateCoin = 0
//			}
//		}
//		if total > 0 {
//			p.AddCoin(total, common.GainWay_RebateTask, "", "RebateTask")
//			//增加泥码
//			p.AddDirtyCoin(0, total)
//			p.ReportSystemGiveEvent(int32(total), common.GainWay_RebateTask, false)
//		}
//
//	}
//	pack := &gamehall.SCReceiveRebate{
//		OpRetCode: gamehall.OpResultCode_Hall_OPRC_Sucess_Hall,
//		Coin:      proto.Int64(total),
//	}
//	if total == 0 {
//		return nil
//	}
//	proto.SetDefaults(pack)
//	logger.Logger.Trace("SCReceiveRebate: ", pack)
//	p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_REBATE_RECEIVE), pack)
//	p.RebateRedIsShow(false)
//	/////////////////////////////////////////////////////
//	if isFToday {
//		SendRebateList(p)
//	}
//	return nil
//}
//
//type CSGetIsCanRebatePacketFactory struct {
//}
//
//type CSGetIsCanRebateHandler struct {
//}
//
//func (this *CSGetIsCanRebatePacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSGetIsCanRebate{}
//	return pack
//}
//
//func (this *CSGetIsCanRebateHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSGetIsCanRebateHandler Process recv ", data)
//
//	p := PlayerMgrSington.GetPlayer(sid)
//	if p == nil {
//		logger.Logger.Errorf("CSGetIsCanRebateHandler.Process p is nil")
//		return nil
//	}
//	if p.layered[common.ActId_RebateTask] {
//		logger.Logger.Errorf("CSRankInfoHandler.Process p.layered[common.ActId_RebateTask]")
//		return nil
//	}
//	SendErrorMsg := func(opretcode gamehall.OpResultCode_Hall) {
//		pack := &gamehall.SCGetIsCanRebate{
//			OpRetCode: opretcode,
//			IsCan:     proto.Int64(0),
//		}
//		proto.SetDefaults(pack)
//		logger.Logger.Trace("SCGetIsCanRebateHandler: ", pack)
//		p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_GETISCANREBATE), pack)
//	}
//	_, ok := PlatformMgrSington.Platforms[p.Platform]
//	if !ok {
//		logger.Logger.Errorf("CSGetIsCanRebateHandler.Process snid %v platform %s is not find in PlatformMgrSington.Platforms", p.SnId, p.Platform)
//		SendErrorMsg(gamehall.OpResultCode_Hall_OPRC_OnlineReward_Info_FindPlatform_Fail_Hall)
//		return nil
//	}
//
//	rebateTask := RebateInfoMgrSington.rebateTask[p.Platform]
//	//有些平台可能没有返利信息，就不要往下执行了
//	if rebateTask == nil {
//		logger.Logger.Warnf("CSGetIsCanRebateHandler.Process snid %v rebateTask is nil", p.SnId)
//		SendErrorMsg(gamehall.OpResultCode_Hall_OPRC_Error_Hall)
//		return nil
//	}
//
//	//判断是否开启返利
//	if rebateTask.RebateSwitch == false {
//		SendErrorMsg(gamehall.OpResultCode_Hall_OPRC_Sucess_Hall)
//		return nil
//	}
//
//	if rebateTask.RebateManState == 0 || (rebateTask.RebateManState == 1 && p.IsCanRebate == 1) {
//		pack := &gamehall.SCGetIsCanRebate{
//			OpRetCode: gamehall.OpResultCode_Hall_OPRC_Sucess_Hall,
//			IsCan:     proto.Int64(2),
//		}
//		proto.SetDefaults(pack)
//		logger.Logger.Trace("SCGetIsCanRebateHandler: ", pack)
//		p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_GETISCANREBATE), pack)
//		return nil
//	}
//
//	//这种需要后台拉取
//	var Url string
//	var WX string
//	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//		url, wx, err := webapi.API_GetRebateImgUrl(common.GetAppId(), p.Platform)
//		if err != nil {
//			logger.Logger.Infof("get url err: %v", err.Error())
//			return err
//		} else {
//			Url = url
//			WX = wx
//			return nil
//		}
//	}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//		if data != nil {
//			SendErrorMsg(gamehall.OpResultCode_Hall_OPRC_Error_Hall)
//		} else {
//			pack := &gamehall.SCGetIsCanRebate{
//				OpRetCode: gamehall.OpResultCode_Hall_OPRC_Sucess_Hall,
//				IsCan:     proto.Int64(1),
//				Url:       proto.String(Url),
//				WX:        proto.String(WX),
//			}
//			proto.SetDefaults(pack)
//			logger.Logger.Trace("SCGetIsCanRebateHandler: ", pack)
//			p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_GETISCANREBATE), pack)
//		}
//	}), "CSGetIsCanRebateHandler").Start()
//
//	return nil
//}
//
//type CSNewPlayerInfoPacketFactory struct {
//}
//
//type CSNewPlayerInfoHandler struct {
//}
//
//func (this *CSNewPlayerInfoPacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSNewPlayerInfo{}
//	return pack
//}
//
//func (this *CSNewPlayerInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSNewPlayerInfoHandler Process recv ", data)
//	if _, ok := data.(*gamehall.CSNewPlayerInfo); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Errorf("CSNewPlayerInfoHandler.Process p is nil")
//			return nil
//		}
//
//		_, ok := PlatformMgrSington.Platforms[p.Platform]
//		if !ok {
//			logger.Logger.Errorf("CSNewPlayerInfoHandler.Process snid %v platform %s is not find in PlatformMgrSington.Platforms", p.SnId, p.Platform)
//			return nil
//		}
//		logger.Logger.Tracef("CSNewPlayerInfoHandler platform %v snid %v", p.Platform, p.SnId)
//		p.ClearRebate()
//		var total, TotalCoin int64 //总局数
//		type MaxKeyVal struct {
//			MaxCoin int64
//			MaxKey  string
//		}
//		var maxPartake = new(MaxKeyVal)        //参与最多的游戏局数
//		var maxProfit = new(MaxKeyVal)         //单局最多盈利
//		var maxCreate = new(MaxKeyVal)         //创建房间最多
//		var maxCreateClubRoom = new(MaxKeyVal) //创建包间最多
//		var totalCreateRoom int64
//		for k, gd := range p.GameData {
//			total += gd.GameTimes
//			if gd.GameTimes > maxPartake.MaxCoin {
//				maxPartake.MaxCoin = gd.GameTimes
//				maxPartake.MaxKey = k
//			}
//			if gd.MaxSysOut > maxProfit.MaxCoin {
//				maxProfit.MaxCoin = gd.MaxSysOut
//				maxProfit.MaxKey = k
//			}
//			if gd.CreateRoomTimes > maxCreate.MaxCoin {
//				maxCreate.MaxCoin = gd.CreateRoomTimes
//				maxCreate.MaxKey = k
//			}
//			if gd.CreateClubRoomTimes > maxCreateClubRoom.MaxCoin {
//				maxCreateClubRoom.MaxCoin = gd.CreateClubRoomTimes
//				maxCreateClubRoom.MaxKey = k
//			}
//			totalCreateRoom += gd.CreateRoomTimes
//		}
//		var canTotal int64
//		rebateTask := RebateInfoMgrSington.rebateTask[p.Platform]
//		if rebateTask != nil && rebateTask.RebateSwitch {
//			//////////////////////////////第三方///////
//			for k, _ := range rebateTask.RebateGameThirdCfg {
//				if info, ok := p.RebateData[k]; ok {
//					TotalCoin += info.ValidBetTotal
//					if rebateTask.ReceiveMode == 0 {
//						canTotal += info.TodayRebateCoin + info.TotalHaveRebateCoin
//					} else {
//						canTotal += info.YesterdayRebateCoin + info.TotalHaveRebateCoin
//					}
//				}
//			}
//			//博乐棋牌统计/////////////////////////
//			for k, _ := range rebateTask.RebateGameCfg {
//				if info, ok := p.RebateData[k]; ok {
//					TotalCoin += info.ValidBetTotal
//					if rebateTask.ReceiveMode == 0 {
//						canTotal += info.TodayRebateCoin + info.TotalHaveRebateCoin
//					} else {
//						canTotal += info.YesterdayRebateCoin + info.TotalHaveRebateCoin
//					}
//				}
//			}
//		}
//		pack := &gamehall.SCNewPlayerInfo{
//			GameTotalNum:     proto.Int32(int32(total)),
//			CreateRoomNum:    proto.Int32(int32(totalCreateRoom)),
//			CreateClubNum:    proto.Int32(p.CreateClubNum),
//			TotalCoin:        proto.Int64(TotalCoin),
//			LastGetCoinTime:  proto.Int64(p.LastRebateTime),
//			Coin:             proto.Int64(canTotal),
//			TeamNum:          proto.Int32(0),
//			AchievementTotal: proto.Int32(0),
//			RewardTotal:      proto.Int32(0),
//		}
//		if len(maxPartake.MaxKey) > 0 {
//			maxPartake.MaxKey = strconv.Itoa(int(GameFreeMgrEx.GetGameFreeIdByGameDif(maxPartake.MaxKey)))
//			pack.GameMostPartake = proto.String(maxPartake.MaxKey)
//		}
//		if len(maxProfit.MaxKey) > 0 {
//			maxProfit.MaxKey = strconv.Itoa(int(GameFreeMgrEx.GetGameFreeIdByGameDif(maxProfit.MaxKey)))
//			pack.GameMostProfit = proto.String(maxProfit.MaxKey)
//			pack.GameMostProfitNum = proto.Int32(int32(maxProfit.MaxCoin))
//		}
//		if len(maxCreate.MaxKey) > 0 {
//			maxCreate.MaxKey = strconv.Itoa(int(GameFreeMgrEx.GetGameFreeIdByGameDif(maxCreate.MaxKey)))
//			pack.CreateRoomMost = proto.String(maxCreate.MaxKey)
//		}
//		if len(maxCreateClubRoom.MaxKey) > 0 {
//			maxCreateClubRoom.MaxKey = strconv.Itoa(int(GameFreeMgrEx.GetGameFreeIdByGameDif(maxCreateClubRoom.MaxKey)))
//			pack.CreateClubRoomMost = proto.String(maxCreateClubRoom.MaxKey)
//		}
//
//		//第三方大类控制
//		cfgid := PlatformMgrSington.GetPlatformConfigId(p.Platform, "")
//		gps := make(map[int32]*PlatConDataDetail)
//		PlatformMgrSington.GetPlatformConfig(cfgid, 0, true, gps)
//		classType := make(map[int32]bool)
//		//棋牌 电子 捕鱼 固定存在
//		classType[int32(gamehall.HallOperaCode_HallChessGame)] = true
//		classType[int32(gamehall.HallOperaCode_HallElectronicGame)] = true
//		classType[int32(gamehall.HallOperaCode_HallFishingGame)] = true
//		for _, v := range gps {
//			if v.State == 1 {
//				//视讯 电子 体育
//				showTypeId := v.DBGameFree.GetGameClass()
//				switch gamehall.HallOperaCode(showTypeId) {
//				case gamehall.HallOperaCode_HallLiveVideo, gamehall.HallOperaCode_HallLotteryGame,
//					gamehall.HallOperaCode_HallSportsGame:
//					classType[showTypeId] = true
//				}
//			}
//		}
//		classType[int32(gamehall.HallOperaCode_HallThirdPlt)] = true
//		//俱乐部大类控制
//		pms := PlatformMgrSington.GetPlatform(p.Platform)
//		if pms != nil {
//			if pms.ClubConfig != nil && pms.ClubConfig.IsOpenClub {
//				classType[int32(gamehall.HallOperaCode_HallPrivateRoom)] = true
//				classType[int32(gamehall.HallOperaCode_HallClubRoom)] = true
//			}
//		}
//		for k, ok := range classType {
//			if ok {
//				pack.ClassType = append(pack.ClassType, k)
//			}
//		}
//		sort.Slice(pack.ClassType, func(i, j int) bool {
//			if pack.ClassType[i] > pack.ClassType[j] {
//				return false
//			}
//			return true
//		})
//		//返利模式
//		if rs, ok := RebateInfoMgrSington.rebateTask[p.Platform]; ok {
//			pack.CodeType = proto.Int32(int32(rs.ReceiveMode))
//		}
//		type Spread struct {
//			Player_Group  int32
//			Count_History int32
//			Total_Amount  int32
//		}
//		type ApiResult struct {
//			Tag int
//			Msg Spread
//		}
//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//			buff, err := webapi.API_GetSpreadPlayer(common.GetAppId(), p.SnId, p.Platform)
//			if err != nil {
//				return nil
//			}
//			return buff
//		}), task.CompleteNotifyWrapper(func(data interface{}, tt *task.Task) {
//			if apiData, ok := data.([]byte); ok {
//				logger.Logger.Trace("API_GetSpreadPlayer: ", string(apiData))
//				ar := ApiResult{}
//				err := json.Unmarshal(apiData, &ar)
//				if err == nil && ar.Tag == 0 {
//					pack.TeamNum = proto.Int32(ar.Msg.Player_Group)
//					pack.AchievementTotal = proto.Int32(ar.Msg.Count_History)
//					pack.RewardTotal = proto.Int32(ar.Msg.Total_Amount)
//				}
//			}
//			proto.SetDefaults(pack)
//			logger.Logger.Trace("SCNewPlayerInfo: ", pack)
//			p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_NEWPLAYERINFO), pack)
//			return
//		}), "GetRebateLog").StartByFixExecutor("RebateLog_w")
//	}
//	return nil
//}
//
//type CSCodeTypeRecordPacketFactory struct {
//}
//
//type CSCodeTypeRecordHandler struct {
//}
//
//func (this *CSCodeTypeRecordPacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSCodeTypeRecord{}
//	return pack
//}
//
//func (this *CSCodeTypeRecordHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSCodeTypeRecordHandler Process recv ", data)
//	if msg, ok := data.(*gamehall.CSCodeTypeRecord); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Errorf("CSRankInfoHandler.Process p is nil")
//			return nil
//		}
//
//		_, ok := PlatformMgrSington.Platforms[p.Platform]
//		if !ok {
//			logger.Logger.Errorf("CSRankInfoHandler.Process snid %v platform %s is not find in PlatformMgrSington.Platforms", p.SnId, p.Platform)
//			return nil
//		}
//
//		logger.Logger.Tracef("CSCodeTypeRecordHandler platform %v snid %v", p.Platform, p.SnId)
//
//		//p.ClearRebate()
//
//		pack := &gamehall.SCCodeTypeRecord{
//			ShowType: msg.GetShowTypeId(),
//		}
//		rebateTask := RebateInfoMgrSington.rebateTask[p.Platform]
//		ShowTypeId := int(msg.GetShowTypeId())
//		for k, cfg := range rebateTask.RebateGameCfg {
//			//var gameName string
//			var iFContinue bool
//			gfm := GameFreeMgrEx.GetDBGameFreeMgrByGameDif(k)
//			if gfm != nil && gfm.GetGameType() == 0 && int(gfm.GetGameClass()) == ShowTypeId {
//				//gameName = gfm.GetPlatformName() + "-" + gfm.GetName()
//				iFContinue = true
//			}
//			if !iFContinue {
//				continue
//			}
//			var CanTotal, showCoin int64
//			if info, ok := p.RebateData[k]; ok {
//				showCoin += info.ValidBetTotal
//				if rebateTask.ReceiveMode == 0 {
//					CanTotal += info.TodayRebateCoin
//				} else {
//					CanTotal += info.YesterdayRebateCoin
//				}
//			}
//			id := strconv.Itoa(int(GameFreeMgrEx.GetGameFreeIdByGameDif(k)))
//			ctr := &gamehall.CodeTypeRecord{
//				GameBetCoin: proto.Int64(showCoin),
//				Rate:        proto.Int32(cfg.RebateRate[0]),
//				Coin:        proto.Int32(int32(CanTotal)),
//				MaxCoin:     proto.Int32(int32(cfg.MaxRebateCoin)),
//				MinCoin:     proto.Int32(cfg.BaseCoin[0]),
//				GameName:    proto.String(id),
//			}
//			//配置根据当前打码量变化
//			for i := 0; i < 3; i++ {
//				a := math.Floor(float64(showCoin) * float64(cfg.RebateRate[i]) / 10000)
//				if int64(a) == CanTotal && a != 0 {
//					ctr.Rate = proto.Int32(cfg.RebateRate[i])
//				}
//			}
//			pack.CodeTypeRecord = append(pack.CodeTypeRecord, ctr)
//		}
//		for k, cfg := range rebateTask.RebateGameThirdCfg {
//			var iFContinue bool
//			if ShowTypeId == int(gamehall.HallOperaCode_HallLiveVideo) && cfg.ThirdId == "28" {
//				iFContinue = true
//			}
//			if ShowTypeId == int(gamehall.HallOperaCode_HallThirdPlt) && cfg.ThirdId != "28" {
//				iFContinue = true
//			}
//			if !iFContinue {
//				continue
//			}
//			var canTotal, rebateCoin int64
//			if info, ok := p.RebateData[k]; ok {
//				rebateCoin += info.ValidBetTotal
//				if rebateTask.ReceiveMode == 0 {
//					canTotal += info.TodayRebateCoin + info.TotalHaveRebateCoin
//				} else {
//					canTotal += info.YesterdayRebateCoin + info.TotalHaveRebateCoin
//				}
//			}
//			ctr := &gamehall.CodeTypeRecord{
//				GameName:    proto.String(cfg.ThirdShowName + "-所有游戏"),
//				GameBetCoin: proto.Int64(rebateCoin),
//				Rate:        proto.Int32(cfg.RebateRate[0]),
//				Coin:        proto.Int32(int32(canTotal)),
//				MaxCoin:     proto.Int32(int32(cfg.MaxRebateCoin)),
//				MinCoin:     proto.Int32(cfg.BaseCoin[0]),
//			}
//			//配置根据当前打码量变化
//			for i := 0; i < 3; i++ {
//				a := math.Floor(float64(rebateCoin) * float64(cfg.RebateRate[i]) / 10000)
//				if int64(a) == canTotal && a != 0 {
//					ctr.Rate = proto.Int32(cfg.RebateRate[i])
//				}
//			}
//			pack.CodeTypeRecord = append(pack.CodeTypeRecord, ctr)
//		}
//		//根据名字排序
//		var na []string
//		for _, v := range pack.CodeTypeRecord {
//			na = append(na, v.GetGameName())
//		}
//		sort.Strings(na)
//		var ctr []*gamehall.CodeTypeRecord
//		ctr = pack.CodeTypeRecord
//		pack.CodeTypeRecord = nil
//		for _, v := range na {
//			for _, n := range ctr {
//				if v == n.GetGameName() {
//					pack.CodeTypeRecord = append(pack.CodeTypeRecord, n)
//					break
//				}
//			}
//		}
//		proto.SetDefaults(pack)
//		logger.Logger.Info("CodeTypeRecord: ", pack)
//		p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_CODETYPERECORD), pack)
//	}
//
//	return nil
//}
//
//type CSBetCoinRecordPacketFactory struct {
//}
//type CSBetCoinRecordHandler struct {
//}
//
//func (this *CSBetCoinRecordPacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSBetCoinRecord{}
//	return pack
//}
//
//func (this *CSBetCoinRecordHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSBetCoinRecordHandler Process recv ", data)
//	if msg, ok := data.(*gamehall.CSBetCoinRecord); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Errorf("CSBetCoinRecordHandler.Process p is nil")
//			return nil
//		}
//		var roomType int32 = common.RoomType_Public
//		switch msg.GetShowTypeId() {
//		case gamehall.HallOperaCode_HallPrivateRoom:
//			roomType = common.RoomType_Private
//		case gamehall.HallOperaCode_HallClubRoom:
//			roomType = common.RoomType_Club
//		}
//		startTime, endTime := int64(0), int64(0)
//		switch msg.GetTimeIndex() {
//		case 1:
//			t := time.Now()
//			startTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
//			endTime = startTime + 86400 - 1
//		case 2:
//			t := time.Now()
//			startTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix() - 86400
//			endTime = startTime + 86400 - 1
//		case 3:
//			t := time.Now()
//			endTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
//			startTime = endTime - (30-1)*86400
//			endTime += 86400 - 1
//		}
//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//			return model.GetPlayerListByHall(p.SnId, p.Platform, int(msg.GetPageNo()), 50,
//				startTime, endTime, roomType, int32(msg.GetShowTypeId()))
//		}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//			gpl := data.(model.GamePlayerListType)
//			pack := &gamehall.SCBetCoinRecord{
//				PageNo:   proto.Int32(int32(gpl.PageNo)),
//				PageNum:  proto.Int32(int32(gpl.PageSum)),
//				PageSize: proto.Int32(int32(gpl.PageSize)),
//			}
//			for _, v := range gpl.Data {
//				rec := &gamehall.BetCoinRecord{
//					Ts:           proto.Int64(int64(v.Ts)),
//					GameName:     proto.String(strconv.Itoa(int(v.GameFreeid))),
//					RecordId:     proto.String(v.GameDetailedLogId),
//					BetCoin:      proto.Int64(v.BetAmount),
//					ReceivedCoin: proto.Int64(v.WinAmountNoAnyTax),
//				}
//				if rec.GetRecordId() == "" {
//					rec.RecordId = proto.String(v.ThirdOrderId)
//				}
//				pack.BetCoinRecord = append(pack.BetCoinRecord, rec)
//			}
//			proto.SetDefaults(pack)
//			logger.Logger.Info("SCBetCoinRecord: ", pack)
//			p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_BETCOINRECORD), pack)
//		}), "CSBetCoinRecordHandler").Start()
//	}
//	return nil
//}
//
//type CSCoinDetailedPacketFactory struct {
//}
//type CSCoinDetailedHandler struct {
//}
//
//func (this *CSCoinDetailedPacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSCoinDetailed{}
//	return pack
//}
//
//func (this *CSCoinDetailedHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSCoinDetailedHandler Process recv ", data)
//	if msg, ok := data.(*gamehall.CSCoinDetailed); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Errorf("CSCoinDetailedHandler.Process p is nil")
//			return nil
//		}
//		if msg.GetCoinType() == 0 && msg.GetTimeIndex() == 0 {
//			tt := []int32{
//				common.GainWaySort_Pay,
//				common.GainWaySort_Exchange,
//				common.GainWaySort_Mail,
//				common.GainWaySort_SafeBox,
//				common.GainWaySort_Manual,
//				common.GainWaySort_Act,
//				common.GainWaySort_Thrid,
//				common.GainWaySort_VIP,
//				common.GainWaySort_PrivateScene,
//			}
//			//返利
//			if data, ok := RebateInfoMgrSington.rebateTask[p.Platform]; ok {
//				if data.RebateSwitch {
//					tt = append(tt, common.GainWaySort_Rebate)
//				}
//			}
//			//俱乐部
//			//if !clubManager.IsClubNotOpen(p.Platform) {
//			//	tt = append(tt, common.GainWaySort_Club)
//			//}
//			////余额宝
//			//if cfg, ok := ActYebMgrSington.Configs[p.Platform]; ok {
//			//	if cfg.StartAct == 1 {
//			//		tt = append(tt, common.GainWaySort_YebDeposit)
//			//	}
//			//}
//			sort.Slice(tt, func(i, j int) bool {
//				if tt[i] < tt[j] {
//					return true
//				}
//				return false
//			})
//			ct := &gamehall.SCCoinTotal{
//				RechargeCoin:    proto.Int64(p.CoinPayTotal),
//				ExchangeCoin:    proto.Int64(p.CoinExchangeTotal),
//				ClubAddCoin:     proto.Int64(p.ClubInCoin - p.ClubOutCoin),
//				RebateCoin:      proto.Int64(p.TotalRebateCoin),
//				Activity:        proto.Int64(int64(p.ActivityCoin)),
//				TransactionType: tt,
//			}
//			proto.SetDefaults(ct)
//			logger.Logger.Info("SCCoinTotal: ", ct)
//			p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_COINTOTAL), ct)
//		}
//		const (
//			TimeIndex_Today    int64 = iota + 1 //今天
//			TimeIndex_Yestoday                  //昨天
//			TimeIndex_Month                     //一个月内
//		)
//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//			startTime, endTime := int64(0), int64(0)
//			switch msg.GetTimeIndex() {
//			case TimeIndex_Today:
//				t := time.Now()
//				startTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
//				endTime = startTime + 86400 - 1
//			case TimeIndex_Yestoday:
//				t := time.Now()
//				startTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix() - 86400
//				endTime = startTime + 86400 - 1
//			case TimeIndex_Month:
//				t := time.Now()
//				endTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
//				startTime = endTime - (30-1)*86400
//				endTime += 86400 - 1
//			}
//
//			coin_type := common.GetGainWaySort(int32(msg.GetCoinType()))
//			var logTypeParam int32
//			if int32(msg.GetCoinType()) == common.GainWaySort_All || int32(msg.GetCoinType()) == common.GainWaySort_Manual {
//				logTypeParam = common.GainWaySort_Api
//			}
//			result, _ := model.GetCoinLogBySnidAndTypeAndInRangeTsByPage(p.SnId, coin_type, logTypeParam, msg.GetPageNo(), 50, startTime, endTime)
//			return result
//		}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//			pack := &gamehall.SCCoinDetailedTotal{}
//			result := data.(*model.CoinLogLog)
//			if result != nil {
//				pack = &gamehall.SCCoinDetailedTotal{
//					PageNo:   proto.Int32(result.PageNo),
//					PageNum:  proto.Int32(result.PageNum),
//					PageSize: proto.Int32(result.PageSize),
//				}
//				for _, v := range result.Logs {
//					item := &gamehall.CoinDetailed{
//						Ts:       proto.Int64(v.Ts),
//						CoinType: proto.Int64(int64(v.LogType)),
//						Coin:     proto.Int64(v.RestCount),
//					}
//					if v.Count > 0 {
//						item.Income = proto.Int64(v.Count)
//					} else {
//						item.Disburse = proto.Int64(-v.Count)
//					}
//					pack.CoinDetailed = append(pack.CoinDetailed, item)
//				}
//				proto.SetDefaults(pack)
//				p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_COINDETAILEDTOTAL), pack)
//				return
//			}
//			proto.SetDefaults(pack)
//			p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_COINDETAILEDTOTAL), pack)
//		}), "CSCoinDetailedHandler").Start()
//	}
//	return nil
//}
//
//type CSReportFormPacketFactory struct {
//}
//type CSReportFormHandler struct {
//}
//
//func (this *CSReportFormPacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSReportForm{}
//	return pack
//}
//
//func (this *CSReportFormHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSReportFormHandler Process recv ", data)
//	if msg, ok := data.(*gamehall.CSReportForm); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Errorf("CSReportFormHandler.Process p is nil")
//			return nil
//		}
//		//对应数据表GameClass
//		showTypeId := int(msg.GetShowTypeId())
//		profitCoin, betCoin, flowCoin := int64(0), int64(0), int64(0)
//		if p.TotalGameData != nil {
//			if p.TotalGameData[showTypeId] != nil {
//				n := len(p.TotalGameData[showTypeId])
//				switch msg.GetTimeIndex() {
//				case 0:
//					if n >= 1 {
//						tgd := p.TotalGameData[showTypeId][n-1]
//						profitCoin += tgd.ProfitCoin
//						betCoin += tgd.BetCoin
//						flowCoin += tgd.FlowCoin
//					}
//				case 1:
//					if n >= 2 {
//						tgd := p.TotalGameData[showTypeId][n-2]
//						profitCoin += tgd.ProfitCoin
//						betCoin += tgd.BetCoin
//						flowCoin += tgd.FlowCoin
//					}
//				case 2:
//					for _, v := range p.TotalGameData[showTypeId] {
//						profitCoin += v.ProfitCoin
//						betCoin += v.BetCoin
//						flowCoin += v.FlowCoin
//					}
//				default:
//					return nil
//				}
//			}
//		}
//		if profitCoin < 0 {
//			profitCoin = 0
//		}
//		pack := &gamehall.SCReportForm{
//			ShowType:   proto.Int32(int32(msg.GetShowTypeId())),
//			ProfitCoin: proto.Int64(profitCoin),
//			BetCoin:    proto.Int64(betCoin),
//			FlowCoin:   proto.Int64(flowCoin),
//		}
//		proto.SetDefaults(pack)
//		logger.Logger.Info("SCReportForm: ", pack)
//		p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_REPORTFORM), pack)
//	}
//	return nil
//}
//
//type CSHistoryRecordPacketFactory struct {
//}
//type CSHistoryRecordHandler struct {
//}
//
//func (this *CSHistoryRecordPacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSHistoryRecord{}
//	return pack
//}
//
//func (this *CSHistoryRecordHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSHistoryRecordHandler Process recv ", data)
//	if msg, ok := data.(*gamehall.CSHistoryRecord); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Errorf("CSHistoryRecordHandler.Process p is nil")
//			return nil
//		}
//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//			return model.GetRebateLog(int(msg.GetPageNo()), 50, p.SnId)
//		}), task.CompleteNotifyWrapper(func(data interface{}, tt *task.Task) {
//			rlog := data.(*model.RebateLog)
//			if rlog != nil {
//				pack := &gamehall.SCHistoryRecord{
//					PageNo:   proto.Int32(int32(rlog.PageNo)),
//					PageNum:  proto.Int32(int32(rlog.PageSum)),
//					PageSize: proto.Int32(int32(rlog.PageSize)),
//				}
//				for _, v := range rlog.Rebates {
//					pack.HistoryRecord = append(pack.HistoryRecord, &gamehall.HistoryRecord{
//						Coin:        proto.Int32(int32(v.RebateCoin)),
//						ReceiveType: proto.Int32(v.ReceiveType),
//						Ts:          proto.Int64(v.Ts),
//						CodeCoin:    proto.Int32(int32(v.CodeCoin)),
//					})
//				}
//				proto.SetDefaults(pack)
//				logger.Logger.Info("SCHistoryRecord: ", pack)
//				p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_HISTORYRECORD), pack)
//			}
//			return
//		}), "GetRebateLog").StartByFixExecutor("RebateLog_w")
//	}
//	return nil
//}
//
//type CSReceiveCodeCoinPacketFactory struct {
//}
//type CSReceiveCodeCoinHandler struct {
//}
//
//func (this *CSReceiveCodeCoinPacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSReceiveCodeCoin{}
//	return pack
//}
//
//func (this *CSReceiveCodeCoinHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSReceiveCodeCoinHandler Process recv ", data)
//	if _, ok := data.(*gamehall.CSReceiveCodeCoin); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Errorf("CSReceiveCodeCoinHandler.Process p is nil")
//			return nil
//		}
//		p.ClearRebate()
//		var total int64 = 0
//		var validBetTotal int64
//		rebateTask := RebateInfoMgrSington.rebateTask[p.Platform]
//		if rebateTask.RebateSwitch && p.RebateData != nil {
//			if rebateTask.ReceiveMode == 0 {
//				for _, v := range p.RebateData {
//					validBetTotal += v.ValidBetTotal
//					total += v.TodayRebateCoin + v.TotalHaveRebateCoin
//					v.TotalRebateCoin += v.TodayRebateCoin
//					v.TodayRebateCoin = 0
//					v.ValidBetTotal = 0
//					v.TotalHaveRebateCoin = 0
//				}
//			} else {
//				for _, v := range p.RebateData {
//					validBetTotal += v.YesterdayValidBetTotal + v.TotalHaveValidBetTotal
//					total += v.YesterdayRebateCoin + v.TotalHaveRebateCoin
//					v.TotalRebateCoin += v.YesterdayRebateCoin + v.TotalHaveRebateCoin
//					v.YesterdayRebateCoin = 0
//					v.YesterdayValidBetTotal = 0
//					v.TotalHaveValidBetTotal = 0
//					v.TotalHaveRebateCoin = 0
//				}
//			}
//			if total > 0 {
//				p.AddCoin(total, common.GainWay_RebateTask, "", "RebateTask")
//				//增加泥码
//				p.AddDirtyCoin(0, total)
//				p.ReportSystemGiveEvent(int32(total), common.GainWay_RebateTask, false)
//				p.LastRebateTime = time.Now().Unix()
//				pack := &gamehall.SCReceiveCodeCoin{
//					OpRetCode: gamehall.OpResultCode_Hall_OPRC_Sucess_Hall,
//					Coin:      proto.Int64(total),
//				}
//				proto.SetDefaults(pack)
//				logger.Logger.Info("SCReceiveCodeCoin: ", pack)
//				p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_RECEIVECODECOIN), pack)
//
//				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//					return model.InsertRebateLog(&model.Rebate{
//						SnId:        p.SnId,
//						RebateCoin:  total,
//						ReceiveType: 0,
//						CodeCoin:    validBetTotal,
//					})
//				}), nil, "InsertRebateLog").StartByFixExecutor("ReceiveCodeCoin")
//
//				rti := &gamehall.SCRebateTotalInfo{
//					TotalCoin:       proto.Int64(0),
//					LastGetCoinTime: proto.Int64(p.LastRebateTime),
//					Coin:            proto.Int64(0),
//				}
//				if rs, ok := RebateInfoMgrSington.rebateTask[p.Platform]; ok {
//					rti.CodeType = proto.Int32(int32(rs.ReceiveMode))
//				}
//				proto.SetDefaults(rti)
//				logger.Logger.Info("SCRebateTotalInfo: ", rti)
//				p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_REBATETOTALINFO), rti)
//				return nil
//			}
//			//pack := &gamehall.SCReceiveCodeCoin{
//			//	OpRetCode: gamehall.OpResultCode_Hall_OPRC_Error_Hall,
//			//}
//			//proto.SetDefaults(pack)
//			//logger.Logger.Info("SCReceiveCodeCoin: ", pack)
//			//p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_RECEIVECODECOIN), pack)
//		}
//	}
//	return nil
//}

// type CSFishBetCoinRecordPacketFactory struct {
// }
// type CSFishBetCoinRecordHandler struct {
// }
//
//	func (this *CSFishBetCoinRecordPacketFactory) CreatePacket() interface{} {
//		pack := &gamehall.CSFishBetCoinRecord{}
//		return pack
//	}
//
//	func (this *CSFishBetCoinRecordHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//		logger.Logger.Trace("CSFishBetCoinRecordHandler Process recv ", data)
//		if msg, ok := data.(*gamehall.CSFishBetCoinRecord); ok {
//			p := PlayerMgrSington.GetPlayer(sid)
//			if p == nil {
//				logger.Logger.Errorf("CSFishBetCoinRecordHandler.Process p is nil")
//				return nil
//			}
//			startTime, endTime := int64(0), int64(0)
//			switch msg.GetTimeIndex() {
//			case 1:
//				t := time.Now()
//				startTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
//				endTime = startTime + 86400 - 1
//			case 2:
//				t := time.Now()
//				startTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix() - 86400
//				endTime = startTime + 86400 - 1
//			case 3:
//				t := time.Now()
//				endTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
//				startTime = endTime - (30-1)*86400
//				endTime += 86400 - 1
//			}
//			isLocal := true
//			needsLocal := []string{} //本地需要查询的gamedif
//			needsThird := []int32{}  //三方需要查询的dbgamefreeid
//			needsThirdName := []string{}
//			gameDifStr := ""
//			for _, mv := range msg.GetGameType() {
//				for _, v := range srvdata.PBDB_GameFreeMgr.Datas.Arr {
//					if v.GetGameId() == mv.GetGameId() && v.GetGameMode() == mv.GetGameMode() {
//						if v.GetGameType() == 0 {
//							needsLocal = append(needsLocal, v.GetGameDif())
//						} else if v.GetGameType() == 1 {
//							needsThird = append(needsThird, v.GetId())
//							name := ThirdPltGameMappingConfig.SystemGamefreeidMappingThirdGameName(v.GetId())
//							if len(name) != 0 {
//								needsThirdName = append(needsThirdName, name)
//							}
//							gameDifStr = v.GetGameDif()
//							isLocal = false
//						}
//					}
//				}
//			}
//			if isLocal {
//				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//					return model.GetPlayerListByHall(p.SnId, p.Platform, int(msg.GetPageNo()), 50,
//						startTime, endTime, 0, needsLocal)
//				}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//					gpl := data.(model.GamePlayerListType)
//					pack := &gamehall.SCBetCoinRecord{
//						PageNo:   proto.Int32(int32(gpl.PageNo)),
//						PageNum:  proto.Int32(int32(gpl.PageSum)),
//						PageSize: proto.Int32(int32(gpl.PageSize)),
//					}
//					for _, v := range gpl.Data {
//						gfm := srvdata.PBDB_GameFreeMgr.GetData(v.GameFreeid)
//						gn := GameFreeMgrEx.GetGameName(gfm.GetGameDif())
//						//表现层（参考三方）
//						if v.WinAmountNoAnyTax < 0 {
//							v.BetAmount -= v.WinAmountNoAnyTax
//						}
//						pack.BetCoinRecord = append(pack.BetCoinRecord, &gamehall.BetCoinRecord{
//							Ts:           proto.Int64(int64(v.Ts)),
//							GameName:     proto.String(gn),
//							RecordId:     proto.String(v.LogId.Hex()),
//							BetCoin:      proto.Int64(v.BetAmount),
//							ReceivedCoin: proto.Int64(v.WinAmountNoAnyTax),
//						})
//					}
//					proto.SetDefaults(pack)
//					logger.Logger.Info("SCBetCoinRecord: ", pack)
//					p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_BETCOINRECORD), pack)
//				}), "CSBetCoinRecordHandler").Start()
//				return nil
//			}
//			gameDifInt, _ := strconv.Atoi(gameDifStr)
//			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//				_, buff := webapi.API_GetThirdHotGameDetail(common.GetAppId(), p.Platform, p.SnId, msg.GetPageNo(), 50, int32(msg.GetTimeIndex()), int32(gameDifInt), needsThirdName)
//				type ApiResult struct {
//					Tag int32                      `json:"Tag"`
//					Msg model.ApiThirdDetailResult `json:"Msg"`
//				}
//				result := ApiResult{}
//				err := json.Unmarshal(buff, &result)
//				if err != nil {
//					logger.Logger.Errorf("API_GetThirdHotGameDetail json.Unmarshal err:", err)
//					return nil
//				}
//				return result.Msg
//			}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//				if data != nil {
//					result := data.(model.ApiThirdDetailResult)
//					pack := &gamehall.SCBetCoinRecord{
//						PageNo:   proto.Int32(result.PageNo),
//						PageNum:  proto.Int32(result.PageNum),
//						PageSize: proto.Int32(result.PageSize),
//					}
//					for _, v := range result.Data {
//						item := &gamehall.BetCoinRecord{
//							Ts:           proto.Int64(v.Ts),
//							RecordId:     proto.String(v.RecordId),
//							BetCoin:      proto.Int64(int64(v.BetCoin)),
//							ReceivedCoin: proto.Int64(int64(v.ReceivedCoin)),
//						}
//						gamefreeid := ThirdPltGameMappingConfig.FindSystemGamefreeidByThirdGameInfo(v.ThirdPltName, v.ThirdGameId, v.ThirdGameName)
//						if gamefreeid == 0 {
//							//如果没有找到名字，就用后台发过来的
//							item.GameName = proto.String(v.ThirdGameName)
//						} else {
//							item.GameName = proto.String(srvdata.PBDB_GameFreeMgr.GetData(int32(gamefreeid)).GetName())
//						}
//						pack.BetCoinRecord = append(pack.BetCoinRecord, item)
//					}
//					proto.SetDefaults(pack)
//					logger.Logger.Info("CSFishBetCoinRecordHandler: ", pack)
//					p.SendToClient(int(gamehall.HallPacketID_PACKET_SC_BETCOINRECORD), pack)
//				}
//
//			}), "API_GetThirdHotGameDetail").Start()
//
//		}
//		return nil
//	}
func init() {
	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_REBATE_LIST), &CSRebateListHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_REBATE_LIST), &CSRebateListPacketFactory{})
	//
	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_REBATE_RECEIVE), &CSReceiveRebateHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_REBATE_RECEIVE), &CSReceiveRebatePacketFactory{})
	//
	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_GETISCANREBATE), &CSGetIsCanRebateHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_GETISCANREBATE), &CSGetIsCanRebatePacketFactory{})
	//
	////NewPlayerInfo
	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_NEWPLAYERINFO), &CSNewPlayerInfoHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_NEWPLAYERINFO), &CSNewPlayerInfoPacketFactory{})
	//
	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_CODETYPERECORD), &CSCodeTypeRecordHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_CODETYPERECORD), &CSCodeTypeRecordPacketFactory{})
	//
	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_BETCOINRECORD), &CSBetCoinRecordHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_BETCOINRECORD), &CSBetCoinRecordPacketFactory{})
	//
	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_COINDETAILED), &CSCoinDetailedHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_COINDETAILED), &CSCoinDetailedPacketFactory{})
	//
	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_REPORTFORM), &CSReportFormHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_REPORTFORM), &CSReportFormPacketFactory{})
	//
	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_HISTORYRECORD), &CSHistoryRecordHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_HISTORYRECORD), &CSHistoryRecordPacketFactory{})
	//
	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_RECEIVECODECOIN), &CSReceiveCodeCoinHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_RECEIVECODECOIN), &CSReceiveCodeCoinPacketFactory{})

	//common.RegisterHandler(int(gamehall.HallPacketID_PACKET_CS_FISHBETCOINRECORD), &CSFishBetCoinRecordHandler{})
	//netlib.RegisterFactory(int(gamehall.HallPacketID_PACKET_CS_FISHBETCOINRECORD), &CSFishBetCoinRecordPacketFactory{})
}
