package base

import (
	"games.yol.com/win88/common"
)

//TODO 后续优化改为生成当先监控文件的快照，在游戏部署时对比两个快照文件，修改变动文件对应游戏的版本号，由程序自动完成，避免手动修改

var GameDetailedVer = map[int]int32{}
var ReplayRecorderVer = map[int]int32{}

func init() {
	//需要改动游戏详情版本号，在这里修改
	GameDetailedVer[common.GameId_Unknow] = 1
}
