package base

//主要用来存放配置的平台标记，有了平台标记，就可以通过标记去相应的数据库中读取数据
var PlatformMgrSington = &PlatformMgr{
	Platforms: make(map[string]bool), //全局游戏开关
}

type PlatformMgr struct {
	//结构性数据
	Platforms map[string]bool //key 是平台标记
}

func (this *PlatformMgr) DelPlatform(plt string) {
	delete(PlatformMgrSington.Platforms, plt)
}
func (this *PlatformMgr) UpsertPlatform(plt string) {
	PlatformMgrSington.Platforms[plt] = true
}
