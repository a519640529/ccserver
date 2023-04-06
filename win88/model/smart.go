package model

import "games.yol.com/win88/common"

type SmartConfig struct {
	SwitchPlatform                     []string // 智能化运营平台开关 如： ["-1"] 所有平台不启用,["1","2"] 平台1和平台2开启
	SwitchABTestTailFilter             []int32  // 智能化运营,玩家尾号限定 如:[] 表示不限定尾号 , [1,5,7]表示只有尾号是1,5,7的用户进入ABtest
	SwitchABTestSnIdFilter             []int32  // 智能化运营,玩家限定 如:[] 表示不限定号 , [205123,205124]表示只有205123,205124用户进入ABtest
	ABTestTick                         int32    // AB测试开始tick
	SwitchPlatformABTestTick           []string // 智能化运营”ABTestTick“平台开关 如：[] 全平台开启 ["-1"] 所有平台不启用,["1","2"] 平台1和平台2启用
	SwitchPlatformABTestTickSnIdFilter []int32  // AB测试智能化运营,玩家尾号限定 如:[] 表示不限定号 , [205123,205124]表示只有205123,205124用户进入ABtest
}

// Switch 对战场开关状态
func (c *SmartConfig) Switch(platform string, snid int32) bool {
	n := len(c.SwitchPlatform)
	if n > 0 && c.SwitchPlatform[0] == "-1" {
		return false
	}
	if n == 0 {
		return true
	}

	has := false
	for i := 0; i < n; i++ {
		if c.SwitchPlatform[i] == platform {
			has = true
			break
		}
	}
	if !has {
		return false
	}
	//指定账号
	if len(c.SwitchABTestSnIdFilter) != 0 {
		if common.InSliceInt32(c.SwitchABTestSnIdFilter, snid) {
			return true
		}
		return false
	}
	//指定尾号
	if len(c.SwitchABTestTailFilter) != 0 {
		if common.InSliceInt32(c.SwitchABTestTailFilter, snid%10) {
			return true
		}
		return false
	}
	return true
}

func (c *SmartConfig) SwitchTick(platform string, snid int32) bool {
	n := len(c.SwitchPlatformABTestTick)
	if n > 0 && c.SwitchPlatformABTestTick[0] == "-1" {
		return true
	}

	if n == 0 {
		return false
	}

	has := false
	for i := 0; i < n; i++ {
		if c.SwitchPlatformABTestTick[i] == platform {
			has = true
		}
	}

	if !has {
		return false
	}
	//指定账号
	if len(c.SwitchPlatformABTestTickSnIdFilter) != 0 {
		if common.InSliceInt32(c.SwitchPlatformABTestTickSnIdFilter, snid%10) {
			return true
		}
		return false
	}
	return false
}

// SwitchHundred 百人场开关状态
func (c *SmartConfig) SwitchHundred(platform string) bool {
	n := len(c.SwitchPlatform)
	if n > 0 && c.SwitchPlatform[0] == "-1" {
		return false
	}
	if n == 0 {
		return true
	}

	has := false
	for i := 0; i < n; i++ {
		if c.SwitchPlatform[i] == platform {
			has = true
			break
		}
	}
	return has
}
