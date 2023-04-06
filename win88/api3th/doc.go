package api3th

/*
1.三方接口接入方法参考推饼(智能化运营)和十点半(智能ai)
新增智能ai配置 config.json
"costum": {
	"UseSDBRobotApi3th":true,
	"SDBApi3thTimeout": 10,
	"SDBApi3thAddr":"https://api.vonabcs.com",
	"SDBApiKey": "284672e9733249c69a4d1558b9080f1e",
}
新增智能化运营配置 config.json
"costum": {
	"UseTBSmartApi3th": true,
    "TBApi3thTimeout": 10,
    "TBApi3thAddr":"https://api.vonabcs.com",
    "TBApiKey": "a8c38779-41e0-4a78-b996-8964e752a4d3",
}

2.智能化运营配置在mode/gameparam.go 中的GameParam结构体中新增配置；参考百人炸金花和对战场21点
新增配置 data/gameparam.json
"SmartHZJH": {
	"SwitchPlatform": ["1"],
	"SwitchABTestTailFilter": [],
	"SwitchABTestSnIdFilter": [],
	"ABTestTick":0 ,
	"SwitchPlatformABTestTick": []     ,
	"SwitchPlatformABTestTickSnIdFilter": []
}

3.在具体游戏中添加代码实现智能化运营或智能ai
*/
