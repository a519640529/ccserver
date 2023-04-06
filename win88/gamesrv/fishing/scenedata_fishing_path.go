package fishing

import (
	"fmt"
	"math/rand"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core"
)

var fishMgr = &FishManager{
	Policy_Data:     map[int32]*Policy{},
	Policy_Rand:     []IPolicy{},
	Policy_Activity: make(map[int32][]IPolicy),
	Path:            make(map[int32]model.FishPath),
}
var Mode1_Policy, Tide1_Policy = []int32{101, 102, 103}, []int32{151, 152, 153}
var Mode2_Policy, Tide2_Policy = []int32{201, 202, 203}, []int32{251, 252, 253}
var Mode3_Policy, Tide3_Policy = []int32{301, 302, 303}, []int32{351, 352, 353}
var Mode4_Policy, Tide4_Policy = []int32{420}, []int32{451, 452, 453}
var Mode5_Policy, Tide5_Policy = []int32{501, 502}, []int32{811, 812, 813, 814, 815, 816}
var Mode6_Policy = []int32{701}

type PolicyMode int
type Policy struct {
	Data    map[int32][]IPolicy
	MaxTick int32
}
type IPolicy interface {
	GetId() int32
	GetTime() int32
	GetFishId() int32
	GetPaths() []int32
	GetCount() int32
	GetSpeed() int32
	GetEvent() int32
	GetRefreshInterval() int32
	GetTimeToLive() int32
}
type FishManager struct {
	Policy_Data      map[int32]*Policy
	Policy_Rand      []IPolicy
	Policy_Activity  map[int32][]IPolicy
	Policy_RandomArr []int64
	Path             map[int32]model.FishPath
}

func InitScenePolicyMode(gameId int, sceneType int) []int32 {
	policyArr := []int32{}
	var nomorl_Policy []int32
	var tide_Policy []int32
	switch gameId {
	case common.GameId_HFishing:
		nomorl_Policy = Mode5_Policy
		tide_Policy = Tide5_Policy
	case common.GameId_TFishing: // 天天捕鱼不要鱼阵
		nomorl_Policy = Mode6_Policy
		// tide_Policy = Tide5_Policy
	default:
		nomorl_Policy = Mode1_Policy
		tide_Policy = Tide1_Policy
	}
	nomorlCount := len(nomorl_Policy)
	tideCount := len(tide_Policy)
	nomorlRand := rand.Perm(nomorlCount) // 生成对应的随机数切片
	tideRand := rand.Perm(tideCount)     // 生成对应的随机数切片
	if tideCount == 0 {
		for i := 0; i < nomorlCount; i++ {
			policyArr = append(policyArr, nomorl_Policy[nomorlRand[i]])
		}
	} else if nomorlCount >= tideCount {
		for i := 0; i < tideCount; i++ {
			policyArr = append(policyArr, nomorl_Policy[nomorlRand[i]])
			policyArr = append(policyArr, tide_Policy[tideRand[i]])
		}
		diff := nomorlCount - tideCount
		for i := 0; i < diff && i < nomorlCount && i < tideCount; i++ {
			policyArr = append(policyArr, nomorl_Policy[nomorlRand[tideCount+i]])
			policyArr = append(policyArr, tide_Policy[tideRand[i]])
		}
	} else {
		for i := 0; i < nomorlCount; i++ {
			policyArr = append(policyArr, nomorl_Policy[nomorlRand[i]])
			policyArr = append(policyArr, tide_Policy[tideRand[i]])
		}
		diff := tideCount - nomorlCount
		for i := 0; i < diff && i < nomorlCount && i < tideCount; i++ {
			policyArr = append(policyArr, nomorl_Policy[nomorlRand[i]])
			policyArr = append(policyArr, tide_Policy[tideRand[nomorlCount+i]])
		}
	}
	return policyArr
}

// 变换当前模式
func (this *FishingSceneData) ChangeFlushFish() {
	if this.Policy_Mode == Policy_Mode_Normal {
		this.NotifySceneStateFishing(common.SceneState_Fishing)
		this.Policy_Mode = Policy_Mode_Tide
		this.PolicyId = this.PolicyArr[this.PolicyArrIndex]
		this.PolicyArrIndex++
		if this.PolicyArrIndex >= len(this.PolicyArr) {
			this.PolicyArrIndex = 0
		}
		// 从普通 变回 鱼潮
		this.MaxTick = fishMgr.Policy_Data[this.PolicyId].MaxTick
		this.NextTime = time.Now().Add(time.Millisecond * time.Duration(this.MaxTick*100)).UnixNano()
	} else {
		this.NotifySceneStateFishing(common.SceneState_Normal)
		this.Policy_Mode = Policy_Mode_Normal
		this.PolicyId = this.PolicyArr[this.PolicyArrIndex]
		this.PolicyArrIndex++
		if this.PolicyArrIndex >= len(this.PolicyArr) {
			this.PolicyArrIndex = 0
		}
		this.MaxTick = fishMgr.Policy_Data[this.PolicyId].MaxTick
		this.NextTime = time.Now().Add(time.Millisecond * time.Duration(this.MaxTick*100)).UnixNano()
	}
	this.TimePoint = 0
	this.LastID = 0
	this.BossTag = 0
	this.lastLittleBossTime = time.Now().Unix()
	this.lastBossTime = time.Now().Unix()
	fishlogger.Infof("Next policyid[%v],maxtick[%v],policy index[%v].", this.PolicyId, this.MaxTick, this.PolicyArrIndex)
}
func (this *FishingSceneData) FlushFishOver() bool {
	return time.Now().UnixNano() > this.NextTime
}
func (this *FishManager) GetFishByTime(policyId int32, tick int32) []IPolicy {
	if data, ok := this.Policy_Data[policyId]; ok {
		return data.Data[tick]
	} else {
		fishlogger.Errorf("%v policy data not find in policy cache.", policyId)
		return []IPolicy{}
	}
}

// 新机制 目前只是针对 404 特有
func (this *FishManager) GetFishByFishID(policyId int32, fishId int32) []IPolicy {
	if data, ok := this.Policy_Data[policyId]; ok {
		return data.Data[fishId]
	} else {
		fishlogger.Errorf("%v policy data not find in policy cache.", policyId)
		return []IPolicy{}
	}
}

//  获取该 Policy中的 所有数据
func (this *FishManager) GetPolicyData(policyId int32) *Policy {
	if _, ok := this.Policy_Data[policyId]; ok {
		return this.Policy_Data[policyId]
	} else {
		fishlogger.Errorf("%v policy data not find in policy cache.", policyId)
		return &Policy{}
	}
}

func AddPolicyData(id int32, data []IPolicy) {
	Policy := &Policy{
		Data:    make(map[int32][]IPolicy),
		MaxTick: 0,
	}
	for _, value := range data {
		tick := value.GetTime()
		fishsArr := Policy.Data[tick]
		if fishsArr == nil {
			Policy.Data[tick] = []IPolicy{value}
		} else {
			Policy.Data[tick] = append(fishsArr, value)
		}
		if tick > Policy.MaxTick {
			Policy.MaxTick = tick
		}
	}
	fishMgr.Policy_Data[id] = Policy
}

//  针对 “寻龙夺宝新的出鱼机制”
func AddPolicyDataByFishId(id int32, data []IPolicy) {
	Policy := &Policy{
		Data:    make(map[int32][]IPolicy),
		MaxTick: 0,
	}
	for _, value := range data {
		fishId := value.GetFishId()
		fishsArr := Policy.Data[fishId]
		if fishsArr == nil {
			Policy.Data[fishId] = []IPolicy{value}
		} else {
			Policy.Data[fishId] = append(fishsArr, value)
		}
	}
	fishMgr.Policy_Data[id] = Policy
}

func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		fishlogger.Info("初始化鱼灯[S]")
		defer fishlogger.Info("初始化鱼灯[E]")
		/*
		 * 101,102,103-151,152,153
		 */
		data := []IPolicy{}
		for _, value := range srvdata.PBDB_Policy101Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(101, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy102Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(102, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy103Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(103, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy151Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(151, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy152Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(152, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy153Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(153, data)
		/*
		 * 201,202,203-251,252,253
		 */
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy201Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(201, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy202Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(202, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy203Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(203, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy251Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(251, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy252Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(252, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy253Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(253, data)
		/*
		 * 301,302,303-351,352,353
		 */
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy301Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(301, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy302Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(302, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy303Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(303, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy351Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(351, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy352Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(352, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy353Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(353, data)
		/*
		 * 401,402,403-451,452,453
		 */
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy401Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(401, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy402Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(402, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy403Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(403, data)
		// 添加420 策略
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy420Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyDataByFishId(420, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy451Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(451, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy452Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(452, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy453Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(453, data)
		/*
		 * 501,502-811,812,813,814,815,816
		 */
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy501Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(501, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy502Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(502, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy811Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(811, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy812Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(812, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy813Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(813, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy814Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(814, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy815Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(815, data)
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy816Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(816, data)
		/*
		 * 701,702
		 */
		data = []IPolicy{}
		for _, value := range srvdata.PBDB_Policy701Mgr.Datas.Arr {
			data = append(data, value)
		}
		AddPolicyData(701, data)
		//data = []IPolicy{}
		//for _, value := range srvdata.PBDB_Policy702Mgr.Datas.Arr {
		//	data = append(data, value)
		//}
		//AddPolicyData(702, data)

		for _, v := range srvdata.PBDB_FishHP2Mgr.Datas.Arr {
			// 后期改表
			for i := 1; i != 4; i++ {
				key := fmt.Sprintf("%v-%v", v.GetFishid(), i)
				data := &FishRealHp{
					CurrHp: 0,
					RateHp: v.GetGold()[i-1], // 死亡判断 由于只有重启服务器读表就不加容错处理
				}
				if _, exist := FishHPListEx.fishList[key]; !exist {
					FishHPListEx.fishList[key] = &FishQueue{}
				}
				FishHPListEx.fishList[key].PutEnd(data)
			}
		}

		/*for id, v := range FishHPListEx.fishList {
			for _, v1 := range v.queue {
				fishlogger.Error("FishHPList data ", v1.CurrHp, v1.RateHp)
			}
			fishlogger.Error("FishHPList Here ", v.num, id)
		}*/
		fishPathArr := model.GetFishPath()
		for _, value := range fishPathArr {
			fishMgr.Path[value.Id] = value
		}
		return nil
	})
}
