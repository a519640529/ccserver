package fishing

//2.新增配置信息
const (
	//2.1 历史流水调整系数相关配置
	Wa    = 100000 //历史流水的调整基准，高于此值后历史流水系数不再变化
	Ramax = 1.3    //历史流水调整系数最大值
	Ramin = 0.97   //历史流水调整系数最小值

	//2.2 历史输赢调整系数相关配置
	Hmax_chu  = 20000
	Hmin_chu  = -15000
	Beta_chu  = 0.0001
	Alpha_chu = 0.00015

	Hmax_zho  = 100000
	Hmin_zho  = -100000
	Beta_zho  = 0.0001
	Alpha_zho = 0.0001

	Hmax_gao  = 200000
	Hmin_gao  = -300000
	Beta_gao  = 0.0002
	Alpha_gao = 0.00005

	R1    = 1
	Theta = 0.5

	//2.3 充值行为调整系数相关配置
	P1  = 10000 //充值流水分段1
	Rp1 = 1.01  //在充值后水流在[0，P1]之间的调整系数
	P2  = 30000 //充值流水分段2
	Rp2 = 1.001 //在充值后水流在[P1，P2]之间的调整系数
	Rp3 = 1     //在充值后水流在[P2，∞]之间的调整系数

	//2.4 每日优惠系数相关配置
	D1  = 2000  //日流水分段1
	Rd1 = 1.01  //日水流在[0，D1)之间的调整系数
	D2  = 5000  //日流水分段2
	Rd2 = 1.001 //日水流在[D1，D2)之间的调整系数
	Rd3 = 1     //日水流在[D2，∞)之间的调整系数

	//2.5 捕鱼平台调整系数相关配置
	Tmin      = 0       //平台收益底线
	R2        = 1       //平台收益调整系数基准值
	Rtmin     = 0.5     //平台收益调整系数最小值
	Delta_chu = 0.00001 //初级场下压系数
	Delta_zho = 0.00001 //中级场下压系数
	Delta_gao = 0.00001 //高级场下压系数

	//2.6 捕鱼命中调整
	JunioroolRate1  = 0.8 // 初级水池影响X
	JuniorPoolRate2 = 1
	JuniorPoolRate3 = 1.2
	MiddleroolRate1 = 0.6 // 中级水池影响X
	MiddleroolRate2 = 1
	MiddleroolRate3 = 1.1
	HighpoolRate1   = 0.3 // 高级水池影响X
	HighPoolRate2   = 1
	HighPoolRate3   = 1.05
	//  个人风控区间Y
	YRate1   = 1.4
	YRate2   = 1.3
	YRate3   = 1.1
	YRate4   = 0.8
	YRate5   = 0.5
	YMdRate1 = 1.3
	YMdRate2 = 1.2
	YMdRate3 = 1.1
	YMdRate4 = 0.8
	YMdRate5 = 0.5
	YHgRate1 = 1.2
	YHgRate2 = 1.1
	YHgRate3 = 1.05
	YHgRate4 = 0.8
	YHgRate5 = 0.5
	// 新手保护Z
	Z1 = 1.1
	Z2 = 1.0 // 非新手
	//  盈利比区间H
	HRate1 = 1.2
	HRate2 = 1.1
	HRate3 = 1.0
	HRate4 = 0.8
	HRate5 = 0.5
)

var (
	FishHail   int32 = 10 // 初级场抽成
	MdFishHail int32 = 10 // 中级抽成
	HgFishHail int32 = 10 // 高级抽成
)
