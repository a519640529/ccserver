package fishing

//房间类型
const (
	RoomMode_HuanLe   = 0 //欢乐捕鱼
	RoomMode_TianTian = 0 //天天捕鱼
	RoomMode_Max
)

const (
	Event_Group_Max     = 999   //1~999标识同组鱼,例如:一网打尽之类的
	Event_Ring          = 10000 //冰冻事件
	Event_Booms         = 10001 //全屏炸弹
	Event_Boom          = 10002 //局部炸弹
	Event_Same          = 10003 //托盘鱼事件
	Event_Lightning     = 10004 //连锁闪电
	Event_Rand          = 10005 //随机事件
	Event_FreePower     = 10006 //免费炮台事件
	Event_NewBoom       = 10007 //新的局部炸弹事件,返奖不包括自己
	Event_Bit           = 10008 //钻头鱼事件
	Event_TreasureChest = 10009 //连续开宝箱事件
)
const ( //特殊鱼
	Fish_CaiShen = 11601 //财神鱼
)

const (
	ROOM_LV_CHU  = 1
	ROOM_LV_ZHO  = 2
	ROOM_LV_GAO  = 3
	ROOM_LV_RICH = 4
)
const MaxPlayer = 4
