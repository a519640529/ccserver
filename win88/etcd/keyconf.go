package etcd

/*

#mongo系统配置

#全局库
/sys/global/mgo
#平台用户库
/sys/plt/mgo/user/[平台id]
#平台日志库
/sys/plt/mgo/log/[平台id]

#平台配置信息
/game/plt/config/[平台id]


#平台公告信息
/game/plt/bulletin/[公告id]


#平台招商代理信息
/game/plt/agent_customer/[代理id]


#平台游戏配置信息
/game/plt/game_config/[平台id]/[场次id]


#平台包信息
/game/plt/package/[包标识]


#游戏组配置信息
/game/group_config/[组id]


#平台黑名单
/game/plt/black_list/[黑名单id]


#平台支付方式
/game/plt/pay_list/[支付id]


#平台签到活动
/game/activity/signin/[平台id]


#平台财神任务活动
/game/activity/goldtask/[平台id]


#平台财神降临活动
/game/activity/goldcome/[平台id]


#平台在线奖励活动
/game/activity/onlinereward/[平台id]


#平台幸运转盘活动
/game/activity/lucklyturntable/[平台id]


#平台余额宝活动
/game/activity/yeb/[平台id]


#平台返利
/game/plt/game_rebate_config


#平台杀率配置
/game/plt/profitcontrol/


#比赛配置
/game/match/

#积分商城配置
/game/match/gradeshop

*/

const (
	//系统配置
	ETCDKEY_SYS_ROOT_PREFIX      = "/sys/"
	ETCDKEY_SYS_PLT_DBCFG_PREFIX = "/sys/plt/dbcfg/"

	//业务配置
	ETCDKEY_ROOT_PREFIX                = "/game/"
	ETCDKEY_PLATFORM_PREFIX            = "/game/plt/config/"
	ETCDKEY_BULLETIN_PREFIX            = "/game/plt/bulletin/"
	ETCDKEY_AGENTCUSTOMER_PREFIX       = "/game/plt/agent_customer/"
	ETCDKEY_GAME_CONFIG_GLOBAL         = "/game/plt/game_config_global"
	ETCDKEY_GAMECONFIG_PREFIX          = "/game/plt/game_config/"
	ETCDKEY_PACKAGE_PREFIX             = "/game/plt/package/"
	ETCDKEY_GROUPCONFIG_PREFIX         = "/game/group_config/"
	ETCDKEY_BLACKLIST_PREFIX           = "/game/plt/black_list/"
	ETCDKEY_ACT_SIGNIN_PREFIX          = "/game/activity/signin/"
	ETCDKEY_ACT_TASK_PREFIX            = "/game/activity/task/"
	ETCDKEY_ACT_GOLDTASK_PREFIX        = "/game/activity/goldtask/"
	ETCDKEY_ACT_GOLDCOME_PREFIX        = "/game/activity/goldcome/"
	ETCDKEY_ACT_ONLINEREWARD_PREFIX    = "/game/activity/onlinereward/"
	ETCDKEY_ACT_LUCKLYTURNTABLE_PREFIX = "/game/activity/lucklyturntable/"
	ETCDKEY_ACT_YEB_PREFIX             = "/game/activity/yeb/"
	ETCDKEY_CONFIG_REBATE              = "/game/plt/game_rebate_config/"
	ETCDKEY_PROMOTER_PREFIX            = "/game/plt/promoter/"
	ETCDKEY_ACT_VIP_PREFIX             = "/game/plt/actvip/"
	ETCDKEY_ACT_WEIXIN_SHARE_PREFIX    = "/game/plt/actshare/"
	ETCDKEY_ACT_GIVE_PREFIX            = "/game/plt/actgive/"
	ETCDKEY_ACT_PAY_PREFIX             = "/game/plt/payact/"
	ETCDKEY_ACT_RANDCOIN_PREFIX        = "/game/plt/randcoin/"
	ETCDKEY_ACT_FPAY_PREFIX            = "/game/plt/fpay/"
	ETCDKEY_PLATFORM_PROFITCONTROL     = "/game/plt/profitcontrol/"
	ETCDKEY_MATCH_PROFIX               = "/game/match/"
	ETCDKEY_ACT_TICKET_PROFIX          = "/game/activity/ticket/"
	ETCDKEY_ACT_TICKET_RUNNING         = "/game/activity/ticket/running"
	ETCDKEY_MATCH_GRADESHOP            = "/game/match/gradeshop/"
	ETCDKEY_CONFIG_LOGICLEVEL          = "/game/logiclevel/"
	ETCDKEY_SHOP_EXCHANGE              = "/game/exchange_shop"
	ETCDKEY_GAME_NOTICE                = "/game/common_notice"
	ETCDKEY_SHOP_ITEM                  = "/game/item_shop"
	ETCDKEY_GAME_MATCH                 = "/game/game_match"
)
