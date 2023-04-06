package main

import (
	"fmt"
	"games.yol.com/win88/proto"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"os"
)

func main() {
	msg := &webapi_proto.ASUpdatePlatform{}
	platform := &webapi_proto.Platform{
		PlatformName: "官方平台X",
		Isolated:     false,
		Disabled:     false,
		//ConfigId:               1,
		CustomService:          "",
		BindOption:             0,
		ServiceFlag:            false,
		UpgradeAccountGiveCoin: 0,
		NewAccountGiveCoin:     0,
		PerBankNoLimitAccount:  0,
		ExchangeMin:            0,
		ExchangeLimit:          0,
		ExchangeTax:            0,
		ExchangeForceTax:       0,
		ExchangeFlow:           0,
		ExchangeGiveFlow:       0,
		ExchangeFlag:           0,
		ExchangeVer:            0,
		ExchangeMultiple:       0,
		VipRange:               nil,
		SpreadConfig:           0,
		Leaderboard:            nil,
		ClubConfig:             nil,
		VerifyCodeType:         0,
		ThirdGameMerchant:      nil,
		CustomType:             0,
		NeedSameName:           false,
		ExchangeBankMax:        0,
		ExchangeAlipayMax:      0,
		PerBankNoLimitName:     0,
		IsCanUserBindPromoter:  false,
		UserBindPromoterPrize:  0,
	}

	msg.Platforms = append(msg.Platforms, platform)
	proto.SetDefaults(msg)

	file, err := os.Create("output.bin")
	if err != nil {
		fmt.Println("文件创建失败 ", err.Error())
		return
	}
	defer file.Close()
	b, errors := proto.Marshal(msg)
	if errors != nil {
		fmt.Println("编码失败", errors.Error())
		return
	}
	_, err = file.Write(b)
	if err != nil {
		fmt.Println("编码失败", err.Error())
		return
	}
	fmt.Println("编码成功")
}
