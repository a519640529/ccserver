package main

import (
	"fmt"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/welfare"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/module"
)

var WelfareMgrSington = &WelfareMgr{}

type WelfareMgr struct {
}

func (this *WelfareMgr) ModuleName() string {
	return "WelfareMgr"
}

// 设置 rnums 救济金领取次数
func (this *WelfareMgr) SetWelfData(p *Player, rnums int32) {
	if p.WelfData == nil {
		p.WelfData = new(model.WelfareData)
	}
	if rnums > 0 {
		p.WelfData.ReliefFundTimes += rnums
	}
	p.dirty = true
}

func (this *WelfareMgr) GetReliefFund(p *Player) {
	sdata := srvdata.PBDB_GameSubsidyMgr.GetData(1) // 定义死
	if p.WelfData == nil {
		p.WelfData = new(model.WelfareData)
	}
	pack := &welfare.SCGetReliefFund{
		OpRetCode: welfare.OpResultCode_OPRC_Error,
	}
	if sdata != nil {

		if p.WelfData.ReliefFundTimes >= sdata.Times {
			pack.OpRetCode = welfare.OpResultCode_OPRC_NoTimes

			p.SendToClient(int(welfare.SPacketID_PACKET_SC_WELF_GETRELIEFFUND), pack)

		} else if p.Coin >= int64(sdata.LimitNum) {
			pack.OpRetCode = welfare.OpResultCode_OPRC_CoinTooMore

		} else {
			coin := int64(sdata.Get)
			p.WelfData.ReliefFundTimes += 1
			award := PetMgrSington.GetAwardPetByWelf(p)
			if award > 0 {
				coin = coin * (100 + award) / 100
			}
			p.AddCoin(coin, common.GainWay_ReliefFund, "ReliefFund",
				fmt.Sprintf("领取救济金-%v", coin))
			pack.OpRetCode = welfare.OpResultCode_OPRC_Sucess
			pack.Coin = coin
			pack.Times = p.WelfData.ReliefFundTimes
		}
	}
	p.SendToClient(int(welfare.SPacketID_PACKET_SC_WELF_GETRELIEFFUND), pack)
}

func (this *WelfareMgr) Init() {

}
func (this *WelfareMgr) Update() {

}

func (this *WelfareMgr) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(WelfareMgrSington, time.Second, 0)
}
