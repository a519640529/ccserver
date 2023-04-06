package fishing

import (
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
)

type Policy struct {
	Data    map[int32][]IPolicyData
	MaxTick int32
}

type IPolicyData interface {
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

func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		logger.Logger.Info("初始化鱼灯[S]")
		defer logger.Logger.Info("初始化鱼灯[E]")
		/*
		 * 101,102,103-151,152,153
		 */
		data := []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy101Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(101, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy102Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(102, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy103Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(103, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy151Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(151, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy152Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(152, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy153Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(153, data)
		/*
		 * 201,202,203-251,252,253
		 */
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy201Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(201, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy202Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(202, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy203Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(203, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy251Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(251, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy252Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(252, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy253Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(253, data)
		/*
		 * 301,302,303-351,352,353
		 */
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy301Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(301, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy302Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(302, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy303Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(303, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy351Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(351, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy352Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(352, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy353Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(353, data)
		/*
		 * 401,402,403-451,452,453
		 */
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy401Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(401, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy402Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(402, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy403Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(403, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy451Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(451, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy452Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(452, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy453Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(453, data)
		/*
		 * 501,502-811,812,813,814,815,816
		 */
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy501Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(501, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy502Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(502, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy811Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(811, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy812Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(812, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy813Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(813, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy814Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(814, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy815Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(815, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy816Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(816, data)
		/*
		 * 701,702
		 */
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy701Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(701, data)
		data = []IPolicyData{}
		for _, value := range srvdata.PBDB_Policy702Mgr.Datas.Arr {
			data = append(data, value)
		}
		FishPolicyMgrSington.AddPolicyData(702, data)

		return nil
	})
}

var FishPolicyMgrSington = &FishPolicyMgr{
	data: make(map[int32]*Policy),
}

type FishPolicyMgr struct {
	data map[int32]*Policy
}

func (this *FishPolicyMgr) GetFishByTime(policyId int32, tick int32) []IPolicyData {
	if data, ok := this.data[policyId]; ok {
		return data.Data[tick]
	}
	return nil
}

func (this *FishPolicyMgr) AddPolicyData(id int32, data []IPolicyData) {
	p := &Policy{
		Data:    make(map[int32][]IPolicyData),
		MaxTick: 0,
	}
	for _, value := range data {
		tick := value.GetTime()
		fishsArr := p.Data[tick]
		if fishsArr == nil {
			p.Data[tick] = []IPolicyData{value}
		} else {
			p.Data[tick] = append(fishsArr, value)
		}
		if tick > p.MaxTick {
			p.MaxTick = tick
		}
	}
	this.data[id] = p
}
