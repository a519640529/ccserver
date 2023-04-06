package model

// 赢三张
type WinThreeType struct {
	RoomId                int32            //房间ID
	RoomRounds            int32            //建房局数
	RoomType              int32            //房间类型
	NowRound              int32            //当前局数
	PlayerCount           int              //玩家数量
	BaseScore             int32            //底分
	PlayerData            []WinThreePerson //玩家信息
	Chip                  []PlayerChip     //出牌详情
	ClubRate              int32            //俱乐部抽水比例
	IsSmartOperation      bool             // 是否启用智能化运营
	GamePreUseSmartState  int              //-1:正常
	GameLastUseSmartState int              //-1:正常
	RobotUseSmartState    int64            // Robot使用智能化运营状况
}
type PlayerChip struct {
	UserId       int32 //玩家ID
	BetTotal     int64 //玩家总投注
	Chip         int64 //玩家得分
	StartCoin    int64 //玩家开始前金币
	BetAfterCoin int64 //玩家投注后金币，也就是客户端应该显示的金币
	IsCheck      bool  //是否看牌
	Round        int32 //当前轮次
	Op           int32 //操作
}

type WinThreePerson struct {
	UserId     int32   //玩家ID
	UserIcon   int32   //玩家头像
	ChangeCoin int64   //玩家得分
	Cardinfo   []int32 //牌值
	KindOfCard int32   //牌型
	IsWin      int32   //输赢
	IsRob      bool    //是否是机器人
	Tax        int64   //税，不一定有值，只是作为一个临时变量使用
	ClubPump   int64   //俱乐部额外抽水
	Flag       int     //标识
	Pos        int32   //位置
	StartCoin  int64   //开始金币
	BetCoin    int64   //押注额度
	RoundCheck int32   //轮次，看牌的
	IsFirst    bool    //是否第一次
	IsLeave    bool
	Platform   string `json:"-"`
	Channel    string `json:"-"`
	Promoter   string `json:"-"`
	PackageTag string `json:"-"`
	InviterId  int32  `json:"-"`
	WBLevel    int32  //黑白名单等级
	SingleFlag int32  //单控标记
}

// 经典牛牛 抢庄牛牛
type BullFightType struct {
	RoomId           int32             //房间ID
	RoomRounds       int32             //建房局数
	RoomType         int32             //房间类型
	NowRound         int32             //当前局数
	BankId           int32             //庄家ID
	PlayerCount      int               //玩家数量
	BaseScore        int32             //底分
	PlayerData       []BullFightPerson //玩家信息
	ClubRate         int32             //俱乐部抽水比例
	IsSmartOperation bool              // 是否启用智能化运营
}
type BullFightPerson struct {
	UserId       int32   //玩家ID
	UserIcon     int32   //玩家头像
	ChangeCoin   int64   //玩家得分
	Cardinfo     []int32 //牌值
	CardBakinfo  []int32 //牌值
	IsWin        int32   //输赢
	Tax          int64   //税，不一定有值，只是作为一个临时变量使用
	TaxLottery   int64   //彩金池，增加值
	ClubPump     int64   //俱乐部额外抽水
	IsRob        bool    //是否是机器人
	Flag         int     //标识
	IsFirst      bool
	StartCoin    int64  //开始金币
	Platform     string `json:"-"`
	Channel      string `json:"-"`
	Promoter     string `json:"-"`
	PackageTag   string `json:"-"`
	InviterId    int32  `json:"-"`
	Rate         int64  //斗牛倍率
	WBLevel      int32  //黑白名单等级
	RobBankRate  int64  //抢庄倍率
	SingleAdjust int32  // 单控输赢 1赢 2输
}

// 斗地主、跑得快
type DoudizhuType struct {
	RoomId           int              //房间Id
	BasicScore       int              //基本分
	Spring           int              //春天  1代表春天
	BombScore        int              //炸弹分
	BaseScore        int32            //底分
	PlayerCount      int32            //玩家数量
	BaseCards        []int            //底牌
	BankerId         int32            //斗地主地主Id
	PlayerData       []DoudizhuPerson //玩家信息
	ClubRate         int32            //俱乐部抽水比例
	IsSmartOperation bool             // 是否启用智能化运营
}
type DoudizhuPerson struct {
	UserId       int32     //玩家ID
	UserIcon     int32     //玩家头像
	ChangeCoin   int64     //玩家得分
	OutCards     [][]int64 //出的牌
	SurplusCards []int32   //剩下的牌
	IsWin        int64     //输赢
	IsRob        bool      //是否是机器人
	IsFirst      bool
	Tax          int64  //税，不一定有值，只是作为一个临时变量使用
	ClubPump     int64  //俱乐部额外抽水
	StartCoin    int64  //开始的金币数量
	Platform     string `json:"-"`
	Channel      string `json:"-"`
	Promoter     string `json:"-"`
	PackageTag   string `json:"-"`
	InviterId    int32  `json:"-"`
	WBLevel      int32  //黑白名单等级
}

// 推饼
type TuibingType struct {
	RoomId           int32           //房间ID
	RoomRounds       int32           //建房局数
	RoomType         int32           //房间类型
	NowRound         int32           //当前局数
	BankId           int32           //庄家ID
	PlayerCount      int             //玩家数量
	BaseScore        int32           //底分
	PlayerData       []TuibingPerson //玩家信息
	ClubRate         int32           //俱乐部抽水比例
	IsSmartOperation bool            // 是否启用智能化运营
}
type TuibingPerson struct {
	UserId     int32   //玩家ID
	UserIcon   int32   //玩家头像
	ChangeCoin int64   //玩家得分
	Cardinfo   []int32 //牌值
	IsWin      int32   //输赢
	IsRob      bool    //是否是机器人
	IsFirst    bool
	Tax        int64  //税，不一定有值，只是作为一个临时变量使用
	ClubPump   int64  //俱乐部额外抽水
	Flag       int    //标识
	StartCoin  int64  //开始金币
	Platform   string `json:"-"`
	Channel    string `json:"-"`
	Promoter   string `json:"-"`
	PackageTag string `json:"-"`
	InviterId  int32  `json:"-"`
	Rate       int64  //倍率
	WBLevel    int32  //黑白名单等级
}

// 十三水
// 十三水牌值
type ThirteenWaterPoker struct {
	Head      [3]int
	Mid       [5]int
	End       [5]int
	PokerType int
}
type ThirteenWaterType struct {
	RoomId      int32                 //房间ID
	RoomRounds  int32                 //建房局数
	RoomType    int32                 //房间类型
	NowRound    int32                 //当前局数
	PlayerCount int                   //玩家数量
	BaseScore   int32                 //底分
	PlayerData  []ThirteenWaterPerson //玩家信息
	ClubRate    int32                 //俱乐部抽水比例
}
type ThirteenWaterPerson struct {
	UserId     int32              //玩家ID
	UserIcon   int32              //玩家头像
	ChangeCoin int64              //玩家得分
	Cardinfo   []int              //牌值
	CardsO     ThirteenWaterPoker //十三水确定的牌型
	IsWin      int32              //输赢
	Tax        int64              //税，不一定有值，只是作为一个临时变量使用
	ClubPump   int64              //俱乐部额外抽水
	IsRob      bool               //是否是机器人
	IsFirst    bool
	Flag       int    //标识
	StartCoin  int64  //开始金币
	Platform   string `json:"-"`
	Channel    string `json:"-"`
	Promoter   string `json:"-"`
	PackageTag string `json:"-"`
	InviterId  int32  `json:"-"`
	Rate       int64  //十三水分数
	WBLevel    int32  //黑白名单等级
}

// 二人麻将
type MahjongType struct {
	RoomId      int32           //房间ID
	RoomRounds  int32           //建房局数
	RoomType    int32           //房间类型
	NowRound    int32           //当前局数
	BankId      int32           //庄家ID
	PlayerCount int32           //玩家数量
	BaseScore   int32           //底分
	HuType      []int64         //本局胡牌牌型
	PlayerData  []MahjongPerson //玩家信息
}
type MahjongPerson struct {
	UserId        int32     //玩家ID
	UserIcon      int32     //玩家头像
	ChangeCoin    int64     //玩家得分
	SurplusCards  []int64   //剩下的牌
	CardsChow     [][]int64 //吃的牌
	CardsPong     []int64   //碰的牌
	CardsMingKong []int64   //明杠的牌
	CardsAnKong   []int64   //暗杠的牌
	IsWin         int32     //输赢
	StartCoin     int64     //开始金币
	Tax           int64     //税，不一定有值，只是作为一个临时变量使用
	ClubPump      int64     //俱乐部额外抽水
	IsRob         bool      //是否是机器人
	IsFirst       bool
	Platform      string `json:"-"`
	Channel       string `json:"-"`
	Promoter      string `json:"-"`
	PackageTag    string `json:"-"`
	InviterId     int32  `json:"-"`
	WBLevel       int32  //黑白名单等级
}

// 百人牛牛 龙虎斗 百家乐 红黑大战
type HundredType struct {
	RegionId         int32           //边池ID 庄家 天 地 玄 黄
	IsWin            int             //边池输赢
	Rate             int             //倍数
	CardsInfo        []int32         //扑克牌值
	PlayerData       []HundredPerson //玩家属性
	CardsKind        int32           //牌类型
	CardPoint        int32           //点数
	IsSmartOperation bool            // 是否启用智能化运营
}
type HundredPerson struct {
	UserId             int32 //用户Id
	UserBetTotal       int64 //用户下注
	BeforeCoin         int64 //下注前金额
	AfterCoin          int64 //下注后金额
	ChangeCoin         int64 //金额变化
	IsRob              bool  //是否是机器人
	IsFirst            bool
	WBLevel            int32   //黑白名单等级
	Result             int32   //单控结果
	UserBetTotalDetail []int64 //用户下注明细
}

// 碰撞
type CrashType struct {
	RegionId         int32         //边池ID 庄家 天 地 玄 黄
	IsWin            int           //边池输赢
	Rate             int           //倍数
	CardsInfo        []int32       //扑克牌值
	PlayerData       []CrashPerson //玩家属性
	CardsKind        int32         //牌类型
	CardPoint        int32         //点数
	IsSmartOperation bool          // 是否启用智能化运营
	Hash             string        //Hash
	Period           int           //当前多少期
	Wheel            int           //第几轮
}

// 碰撞
type CrashPerson struct {
	UserId             int32 //用户Id
	UserBetTotal       int64 //用户下注
	UserMultiple       int32 //下注倍数
	BeforeCoin         int64 //下注前金额
	AfterCoin          int64 //下注后金额
	ChangeCoin         int64 //金额变化
	IsRob              bool  //是否是机器人
	IsFirst            bool
	WBLevel            int32   //黑白名单等级
	Result             int32   //单控结果
	UserBetTotalDetail []int64 //用户下注明细
	Tax                int64   //税收
}

// 十点半
type TenHalfType struct {
	RoomId           int32           //房间ID
	RoomRounds       int32           //建房局数
	RoomType         int32           //房间类型
	NowRound         int32           //当前局数
	BankId           int32           //庄家ID
	PlayerCount      int             //玩家数量
	BaseScore        int32           //底分
	PlayerData       []TenHalfPerson //玩家信息
	ClubRate         int32           //俱乐部抽水比例
	IsSmartOperation bool            // 是否启用智能化运营
}

type TenHalfPerson struct {
	UserId     int32   //玩家ID
	UserIcon   int32   //玩家头像
	ChangeCoin int64   //玩家得分
	Cardinfo   []int32 //牌值
	IsWin      int32   //输赢
	Tax        int64   //税，不一定有值，只是作为一个临时变量使用
	ClubPump   int64   //俱乐部额外抽水
	IsRob      bool    //是否是机器人
	Flag       int     //标识
	IsFirst    bool
	StartCoin  int64  //开始金币
	Platform   string `json:"-"`
	Channel    string `json:"-"`
	Promoter   string `json:"-"`
	PackageTag string `json:"-"`
	InviterId  int32  `json:"-"`
	WBLevel    int32  //黑白名单等级
}

type GanDengYanType struct {
	RoomId      int                // 房间Id
	RoomRounds  int32              // 建房局数
	NowRound    int32              // 当前局数
	BombScore   int                // 炸弹倍率
	BaseScore   int32              // 底分
	PlayerCount int32              // 玩家数量
	BankerId    int32              // 庄家id
	PlayerData  []GanDengYanPlayer // 玩家信息
	ClubRate    int32              // 俱乐部抽水比例
}

type GanDengYanHandCards struct {
	Cards []int64
	Index int32
}

type GanDengYanPlayer struct {
	UserId       int32                  // 玩家ID
	UserIcon     int32                  // 玩家头像
	ChangeCoin   int64                  // 玩家得分
	OutCards     []*GanDengYanHandCards // 出的牌
	SurplusCards []int32                // 剩下的牌
	IsWin        int64                  // 输赢
	IsRob        bool                   // 是否是机器人
	IsFirst      bool
	Tax          int64  // 税，不一定有值，只是作为一个临时变量使用
	ClubPump     int64  // 俱乐部额外抽水
	StartCoin    int64  // 开始的金币数量
	Platform     string `json:"-"`
	Channel      string `json:"-"`
	Promoter     string `json:"-"`
	PackageTag   string `json:"-"`
	InviterId    int32  `json:"-"`
	WBLevel      int32  //黑白名单等级
	NumMagic     int32  // 癞子或2倍数
	NumCards     int32  // 剩余牌数
	NumHand      int32  // 手牌倍数
	Num          int32  // 总倍数
}

// 红包扫雷
type RedPackGameType struct {
	BankerId   int32            //庄家id
	BankerBet  int32            //庄家下注金额
	BankerWin  int32            //庄家赢取
	BankerRest int32            //庄家退回钱数
	Rate       int32            //倍数
	BombNum    int32            //雷号
	HitBombCnt int32            //中雷的人数
	PlayerData []*RedPackPerson //玩家属性
}
type RedPackPerson struct {
	UserId     int32 //用户Id
	IsFirst    bool
	BeforeCoin int64 //抢包前金额
	AfterCoin  int64 //抢包后金额
	ChangeCoin int64 //金额变化
	GrabCoin   int32 //抢到红包的数量
	IsHit      bool  //是否中雷
	IsRob      bool  //是否是机器人
}

// 鱼场
type BulletLevelTimes struct {
	Level int32 //等级
	Times int32 //次数
}
type FishCoinNum struct {
	ID     int32
	Num    int32 //打死数量
	Power  int32 //子弹价值
	Coin   int32 //金币
	HitNum int32 //击中次数
}

type FishPlayerData struct {
	UserId   int32 //玩家ID
	UserIcon int32 //玩家头像
	TotalIn  int64 //玩家该阶段总投入
	TotalOut int64 //玩家该阶段总产出
	CurrCoin int64 //记录时玩家当前金币量
}

type FishDetiel struct {
	BulletInfo *[]BulletLevelTimes //子弹统计
	HitInfo    *[]FishCoinNum      //
	PlayData   *FishPlayerData     //统计
}

// 拉霸
// 绝地求生记录详情
type IslandSurvivalGameType struct {
	//all
	RoomId            int32 //房间Id
	BasicScore        int32 //基本分
	PlayerSnid        int32 //玩家id
	BeforeCoin        int64 //下注前金额
	AfterCoin         int64 //下注后金额
	BetCoin           int64 //下注金额
	WinCoin           int64 //下注赢取金额
	IsFirst           bool
	Smallgamewinscore int64 //吃鸡游戏赢取的分数
	ChangeCoin        int64 //本局游戏金额总变化
	FreeTimes         int32 //免费转动次数
	AllWinNum         int32 //中奖的线数
	LeftEnemy         int32 //剩余敌人，需保存
	Killed            int32 //总击杀敌人，需保存
	HitPoolIdx        int   //下注索引
	Cards             []int //15张牌
}

const (
	WaterMarginSmallGame_Unop int32 = iota
	WaterMarginSmallGame_AddOp
	WaterMarginSmallGame_SubOp
)

// 水浒传小游戏数据
type WaterMarginSmallGameInfo struct {
	AddOrSub int32 //加减操作 0:表示未操作 1:加 2:减
	Score    int64 //本局小游戏参与的分数
	Multiple int32 //倍数 0:表示本局输了 >1:表示猜对小游戏
	Dice1    int32 //骰子1的点数
	Dice2    int32 //骰子2的点数
}

type WaterMarginType struct {
	RoomId         int32                       //房间Id
	BasicScore     int32                       //基本分
	PlayerSnid     int32                       //玩家id
	BeforeCoin     int64                       //下注前金额
	AfterCoin      int64                       //下注后金额
	IsFirst        bool                        //是否一次游戏
	ChangeCoin     int64                       //金额变化
	Score          int32                       //总押注数
	AllWinNum      int32                       //中奖的线数
	FreeTimes      int32                       //免费转动次数
	WinScore       int32                       //中奖的倍率
	AllLine        int32                       //线路数
	JackpotNowCoin int64                       //爆奖金额
	Cards          []int                       //15张牌
	HitPoolIdx     int                         //压分的索引
	JackpotHitFlag int                         //如果jackpotnowcoin>0;该值标识中了哪些奖池，二进制字段 1:小奖(第0位) 2:中奖(第1位) 4:大奖(第2位)
	TrigFree       bool                        //是否触发免费转动
	SMGame         []*WaterMarginSmallGameInfo //小游戏数据
}

type FootBallHeroesType struct {
	RoomId            int32 //房间Id
	BasicScore        int32 //基本分
	PlayerSnid        int32 //玩家id
	BeforeCoin        int64 //下注前金额
	AfterCoin         int64 //下注后金额
	ChangeCoin        int64 //金额变化
	IsFirst           bool
	Score             int32   //总押注数
	FreeTimes         int32   //免费转动次数
	WinScore          int32   //中奖的倍率
	Cards             []int   //15张牌
	CoinReward        []int64 //礼物奖励
	Smallgamescore    int64   //小游戏分数
	Smallgamewinscore int64   //小游戏赢取的分数
	SmallgameList     []int32 //小游戏记录
	HitPoolIdx        int     //当前命中的奖池
}
type FruitMachineType struct {
	RoomId         int32 //房间Id
	BasicScore     int32 //基本分
	PlayerSnid     int32 //玩家id
	BeforeCoin     int64 //下注前金额
	AfterCoin      int64 //下注后金额
	IsFirst        bool
	ChangeCoin     int64 //金额变化
	Score          int32 //总押注数
	AllWinNum      int32 //中奖的线数
	FreeTimes      int32 //免费转动次数
	WinScore       int32 //中奖的倍率
	AllLine        int32 //线路数
	JackpotNowCoin int64 //爆奖金额
	Cards          []int //15张牌
	Smallgamescore int64 //小游戏分数
	HitPoolIdx     int   //当前命中的奖池
}
type GoddessType struct {
	RoomId            int32 //房间Id
	BasicScore        int32 //基本分
	PlayerSnid        int32 //玩家id
	BeforeCoin        int64 //下注前金额
	AfterCoin         int64 //下注后金额
	ChangeCoin        int64 //金额变化
	Score             int32 //总押注数
	IsFirst           bool
	Smallgamescore    int64    //小游戏分数
	Smallgamewinscore int64    //小游戏赢取的分数
	SmallgameCard     int32    //小游戏卡牌
	CardsGoddess      [3]int32 //3张牌
	BetMultiple       int32    //押倍倍数
	SmallgameList     []int32  //小游戏记录
	HitPoolIdx        int      //当前命中的奖池
}
type RollLineType struct {
	RoomId     int32 //房间Id
	BasicScore int32 //基本分
	PlayerSnid int32 //玩家id
	BeforeCoin int64 //下注前金额
	AfterCoin  int64 //下注后金额
	ChangeCoin int64 //金额变化
	Score      int32 //总押注数
	IsFirst    bool
	AllWinNum  int32   //中奖的线数
	FreeTimes  int32   //免费转动次数
	WinScore   int32   //中奖的倍率
	AllLine    int32   //线路数
	RoseCount  int32   //玫瑰数量
	Cards      []int   //15张牌
	CoinReward []int64 //礼物奖励
	HitPoolIdx int     //当前命中的奖池
}
type IceAgeType struct {
	RoomId          int32 //房间Id
	BasicScore      int32 //基本分
	PlayerSnid      int32 //玩家id
	BeforeCoin      int64 //下注前金额
	AfterCoin       int64 //下注后金额
	ChangeCoin      int64 //金额变化
	Score           int32 //总押注数
	Tax             int64 //暗税
	IsFirst         bool
	AllWinNum       int32     //中奖的线数
	FreeTimes       int32     //免费转动次数
	WinScore        int32     //中奖的倍率
	AllLine         int32     //线路数
	Cards           [][]int32 // 消除前后的牌（消除前15张，消除后15张...）
	BetLines        []int64   //下注的线
	UserName        string    // 昵称
	TotalPriceValue int64     // 总赢分
	IsFree          bool      // 是否免费
	TotalBonusValue int64     // 总bonus数
	WBLevel         int32     //黑白名单等级
	WinLines        [][]int   // 赢分的线
	WinJackpot      int64     // 赢奖池分数
	WinBonus        int64     // 赢小游戏分数
}
type TamQuocType struct {
	RoomId          int32 //房间Id
	BasicScore      int32 //基本分
	PlayerSnid      int32 //玩家id
	BeforeCoin      int64 //下注前金额
	AfterCoin       int64 //下注后金额
	ChangeCoin      int64 //金额变化
	Score           int32 //总押注数
	Tax             int64 //暗税
	IsFirst         bool
	AllWinNum       int32   //中奖的线数
	FreeTimes       int32   //免费转动次数
	WinScore        int32   //中奖的倍率
	AllLine         int32   //线路数
	Cards           []int32 //15张牌
	BetLines        []int64 //下注的线
	WBLevel         int32   //黑白名单等级
	UserName        string  // 昵称
	TotalPriceValue int64   // 总赢分
	IsFree          bool    // 是否免费
	TotalBonusValue int64   // 总bonus数
	WinLines        []int   // 赢分的线
	WinJackpot      int64   // 赢奖池分数
	WinBonus        int64   // 赢小游戏分数
}
type CaiShenType struct {
	RoomId          int32 //房间Id
	BasicScore      int32 //基本分
	PlayerSnid      int32 //玩家id
	BeforeCoin      int64 //下注前金额
	AfterCoin       int64 //下注后金额
	ChangeCoin      int64 //金额变化
	Score           int32 //总押注数
	Tax             int64 //暗税
	IsFirst         bool
	AllWinNum       int32   //中奖的线数
	FreeTimes       int32   //免费转动次数
	WinScore        int32   //中奖的倍率
	AllLine         int32   //线路数
	Cards           []int32 //15张牌
	WBLevel         int32   //黑白名单等级
	BetLines        []int64 //下注的线
	UserName        string  // 昵称
	TotalPriceValue int64   // 总赢分
	IsFree          bool    // 是否免费
	TotalBonusValue int64   // 总bonus数
	WinLines        []int   // 赢分的线
	WinJackpot      int64   // 赢奖池分数
	WinBonus        int64   // 赢小游戏分数
}
type AvengersType struct {
	RoomId          int32 //房间Id
	BasicScore      int32 //基本分
	PlayerSnid      int32 //玩家id
	BeforeCoin      int64 //下注前金额
	AfterCoin       int64 //下注后金额
	ChangeCoin      int64 //金额变化
	Score           int32 //总押注数
	Tax             int64 //暗税
	IsFirst         bool
	AllWinNum       int32   //中奖的线数
	FreeTimes       int32   //免费转动次数
	WinScore        int32   //中奖的倍率
	AllLine         int32   //线路数
	Cards           []int32 //15张牌
	WBLevel         int32   //黑白名单等级
	BetLines        []int64 //下注的线
	UserName        string  // 昵称
	TotalPriceValue int64   // 总赢分
	IsFree          bool    // 是否免费
	TotalBonusValue int64   // 总bonus数
	WinLines        []int   // 赢分的线
	WinJackpot      int64   // 赢奖池分数
	WinBonus        int64   // 赢小游戏分数
}
type EasterIslandType struct {
	RoomId          int32 //房间Id
	BasicScore      int32 //基本分
	PlayerSnid      int32 //玩家id
	BeforeCoin      int64 //下注前金额
	AfterCoin       int64 //下注后金额
	ChangeCoin      int64 //金额变化
	Score           int32 //总押注数
	Tax             int64 //暗税
	IsFirst         bool
	AllWinNum       int32   //中奖的线数
	FreeTimes       int32   //免费转动次数
	WinScore        int32   //中奖的倍率
	AllLine         int32   //线路数
	Cards           []int32 //15张牌
	WBLevel         int32   //黑白名单等级
	BetLines        []int64 //下注的线
	UserName        string  // 昵称
	TotalPriceValue int64   // 总赢分
	IsFree          bool    // 是否免费
	TotalBonusValue int64   // 总bonus数
	WinLines        []int   // 赢分的线
	WinJackpot      int64   // 赢奖池分数
	WinBonus        int64   // 赢小游戏分数
}
type RollTeamType struct {
	RoomId        int32 //房间Id
	BasicScore    int32 //基本分
	PlayerSnid    int32 //玩家id
	BeforeCoin    int64 //下注前金额
	AfterCoin     int64 //下注后金额
	ChangeCoin    int64 //金额变化
	Score         int32 //总押注数
	AllWinNum     int32 //中奖的线数
	IsFirst       bool
	FreeTimes     int32 //免费转动次数
	WinScore      int32 //中奖的倍率
	AllLine       int32 //线路数
	PokerBaseCoin int32 //扑克游戏获得的金币
	PokerRate     int32 //游戏的翻倍值
	GameCount     int32 //游戏次数
	HitPoolIdx    int   //当前命中的奖池
	Cards         []int //15张牌
}
type RollPerson struct {
	UserId       int32 //用户Id
	UserBetTotal int64 //用户下注
	BeforeCoin   int64 //下注前金额
	AfterCoin    int64 //下注后金额
	ChangeCoin   int64 //金额变化
	IsFirst      bool
	IsRob        bool  //是否是机器人
	WBLevel      int32 //黑白名单等级
}
type RollHundredType struct {
	RegionId   int32 //边池ID  -1.庄家 0.大众 1.雷克萨斯 2.宝马 3.奔驰 4.保时捷 5.玛莎拉蒂 6.兰博基尼 7.法拉利
	IsWin      int   //边池输赢 1.赢 0.平 -1.输
	Rate       int   //倍数
	SType      int32 //特殊牌型 临时使用
	RollPerson []RollPerson
}

// 梭哈
type FiveCardType struct {
	RoomId      int32            //房间ID
	RoomRounds  int32            //建房局数
	RoomType    int32            //房间类型
	PlayerCount int              //玩家数量
	BaseScore   int32            //底分
	PlayerData  []FiveCardPerson //玩家信息
	ClubRate    int32            //俱乐部抽水比例
}

// 梭哈
type FiveCardPerson struct {
	UserId     int32   //玩家ID
	UserIcon   int32   //玩家头像
	ChangeCoin int64   //玩家得分
	Cardinfo   []int32 //牌值
	IsWin      int32   //输赢
	Tax        int64   //税，不一定有值，只是作为一个临时变量使用
	ClubPump   int64   //俱乐部额外抽水
	IsRob      bool    //是否是机器人
	IsFirst    bool    //是否第一次
	IsLeave    bool    //中途离开
	Platform   string  `json:"-"`
	Channel    string  `json:"-"`
	Promoter   string  `json:"-"`
	PackageTag string  `json:"-"`
	InviterId  int32   `json:"-"`
	BetTotal   int64   //用户当局总下注
	IsAllIn    bool    //是否全下
	WBLevel    int32   //黑白名单等级
}

// 骰子
type RollPointPerson struct {
	UserId       int32 //用户Id
	UserBetTotal int64 //用户下注
	BeforeCoin   int64 //下注前金额
	AfterCoin    int64 //下注后金额
	ChangeCoin   int64 //金额变化
	IsFirst      bool
	IsRob        bool    //是否是机器人
	WBLevel      int32   //黑白名单等级
	BetCoin      []int32 //押注金币
}
type RollPointType struct {
	RoomId  int32
	Point   []int32
	Score   []int32
	BetCoin int64
	WinCoin int64
	Person  []RollPointPerson
}

// 轮盘
type RouletteType struct {
	BankerInfo     RouletteBanker   //庄家信息
	Person         []RoulettePerson //下注玩家列表
	RouletteRegion []RouletteRegion //下区域
}
type RouletteBanker struct {
	Point        int   //当局开的点数
	TotalBetCoin int64 //总下注金额
	TotalWinCoin int64 //总输赢金额
}
type RouletteRegion struct {
	Id      int              //0-156 下注位置编号
	IsWin   int              //是否中奖
	BetCoin int64            //当前区域总下注金额
	WinCoin int64            //当前区域总输赢金额
	Player  []RoulettePlayer //当前区域下注玩家列表
}
type RoulettePlayer struct {
	UserId  int32 //用户Id
	BetCoin int64 //当局下注额
}
type RoulettePerson struct {
	UserId       int32         //用户Id
	BeforeCoin   int64         //下注前金额
	AfterCoin    int64         //下注后金额
	UserBetTotal int64         //用户总下注
	UserWinCoin  int64         //用户输赢
	IsRob        bool          //是否是机器人
	WBLevel      int32         //黑白名单等级
	BetCoin      map[int]int64 //下注区域对应金额
}

// 九线拉王
type NineLineKingType struct {
	RoomId         int32 //房间Id
	BasicScore     int32 //基本分
	PlayerSnid     int32 //玩家id
	BeforeCoin     int64 //下注前金额
	AfterCoin      int64 //下注后金额
	IsFirst        bool
	ChangeCoin     int64 //金额变化
	Score          int32 //总押注数
	AllWinNum      int32 //中奖的线数
	FreeTimes      int32 //免费转动次数
	WinScore       int32 //中奖的倍率
	AllLine        int32 //线路数
	JackpotNowCoin int64 //爆奖金额
	Cards          []int //15张牌
	HitPoolIdx     int   //当前命中的奖池
	CommPool       int64 //公共奖池
	PersonPool     int64 //私人奖池
}

// 飞禽走兽
type RollAnimalsType struct {
	BetTotal int64             //总下注
	WinCoin  int64             //用户输赢
	WinFlag  []int64           //中奖元素
	RollLog  []RollHundredType //每个区域下注信息
}

// 血战
type BloodMahjongType struct {
	RoomId      int32                //房间ID
	RoomRounds  int32                //建房局数
	RoomType    int32                //房间类型
	NowRound    int32                //当前局数
	BankId      int32                //庄家ID
	PlayerCount int32                //玩家数量
	BaseScore   int32                //底分
	PlayerData  []BloodMahjongPerson //玩家信息
}

// 碰杠牌
type BloodMahjongCardsLog struct {
	Card int64 //牌
	Pos  []int //0.东 1.南 2.西 3.北
	Flag int   //1.碰 2.明杠 3.暗杠 4.补杠
}

// 分数类型
type BloodMahjongScoreTiles struct {
	LogType    int     //0.胡 1.刮风 2.下雨 3.退税 4.查花猪 5.查大叫 6.被抢杠 补杠退钱 7.呼叫转移
	OtherPos   []int   //源自哪个位置的玩家
	Coin       int64   //理论进账
	ActualCoin int64   //实际进账
	Rate       int32   //倍率
	Params     []int64 //【Honor】//胡牌类型
}
type BloodMahjongPerson struct {
	UserId     int32                    //玩家ID
	UserIcon   int32                    //玩家头像
	ChangeCoin int64                    //玩家得分
	Pos        int32                    //玩家位置 0.东 1.南 2.西 3.北
	IsLeave    bool                     //是否离场
	Bankruptcy bool                     //是否破产
	HuNumber   int32                    //第几胡 1 2 3
	LackColor  int64                    //定缺花色
	Hands      []int64                  //手牌
	HuCards    []int64                  //胡牌
	CardsLog   []BloodMahjongCardsLog   //碰杠牌
	ScoreTiles []BloodMahjongScoreTiles //分数
	IsWin      int32                    //输赢
	Tax        int64                    //税，不一定有值，只是作为一个临时变量使用
	StartCoin  int64                    //开始金币
	ClubPump   int64                    //俱乐部额外抽水
	IsRob      bool                     //是否是机器人
	Platform   string                   `json:"-"`
	Channel    string                   `json:"-"`
	Promoter   string                   `json:"-"`
	PackageTag string                   `json:"-"`
	InviterId  int32                    `json:"-"`
	WBLevel    int32                    //黑白名单等级
}

type PCProp struct {
	Id       uint32
	TypeName string // 金币类型
	AreaType int32  // 所在区域 0平台上 1有效区 2无效区 3小车内
	CoinVal  int64  // 金币面值
	X        float32
	Y        float32
	Z        float32
	RotX     float32
	RotY     float32
	RotZ     float32
}

// 推币机
type PushingCoinRecord struct {
	RoomId      int       // 房间id
	RoomType    int       // 房间类型
	GameMode    int       // 游戏模式
	BaseScore   int64     // 底分
	ShakeTimes  int32     // 震动次数
	WallUpTimes int32     // 升墙次数
	EventTimes  []int32   // 事件次数
	Props       []*PCProp // 所有金币
}

type HuntingRecord struct {
	RoomId     int   // 房间ID
	BaseScore  int   // 底分
	SnId       int32 // 玩家ID
	StartCoin  int64 // 下注前金额
	Coin       int64 // 下注后金额
	ChangeCoin int64 // 金币变化
	CoinPool   int64 // 爆奖金额
	Point      int64 // 单线点数
	LineNum    int64 // 线数
	BetCoin    int64 // 下注金额
	WinCoin    int64 // 产出金额
	Level      int   // 当前关卡
	Gain       int64 // 翻牌奖励
}

// 对战三公
type SanGongPVPType struct {
	RoomId           int32              //房间ID
	RoomRounds       int32              //建房局数
	RoomType         int32              //房间类型
	NowRound         int32              //当前局数
	BankId           int32              //庄家ID
	PlayerCount      int                //玩家数量
	BaseScore        int32              //底分
	PlayerData       []SanGongPVPPerson //玩家信息
	ClubRate         int32              //俱乐部抽水比例
	IsSmartOperation bool               //是否启用智能化运营
}
type SanGongPVPPerson struct {
	UserId     int32   //玩家ID
	UserIcon   int32   //玩家头像
	ChangeCoin int64   //玩家得分
	Cardinfo   []int32 //牌值
	IsWin      int32   //输赢
	Tax        int64   //税，不一定有值，只是作为一个临时变量使用
	ClubPump   int64   //俱乐部额外抽水
	IsRob      bool    //是否是机器人
	Flag       int     //标识
	IsFirst    bool
	StartCoin  int64  //开始金币
	Platform   string `json:"-"`
	Channel    string `json:"-"`
	Promoter   string `json:"-"`
	PackageTag string `json:"-"`
	InviterId  int32  `json:"-"`
	WBLevel    int32  //黑白名单等级
}

// 德州牛仔
type DZNZCardInfo struct {
	CardId    int32   //手牌ID 牛仔 公共牌 公牛
	CardsInfo []int32 //扑克牌值
	CardsKind int32   //牌类型
}
type DZNZZoneInfo struct {
	RegionId         int32           //13个下注区域
	IsWin            int             //边池输赢
	Rate             float32         //倍数
	PlayerData       []HundredPerson //玩家属性
	IsSmartOperation bool            //是否启用智能化运营
}
type DZNZHundredInfo struct {
	CardData []DZNZCardInfo //发牌信息
	ZoneInfo []DZNZZoneInfo //每个下注区域信息
}

// 财运之神
type FortuneZhiShenType struct {
	//基本信息
	RoomId          int   //房间Id
	BasicScore      int32 //单注
	PlayerSnId      int32 //玩家id
	BeforeCoin      int64 //下注前金额
	AfterCoin       int64 //下注后金额
	ChangeCoin      int64 //金额变化
	TotalBetCoin    int64 //总押注
	TotalLine       int32 //总线数(固定)
	TotalWinCoin    int64 //总派彩
	NowGameState    int   //当前游戏模式(0,1,2,3)普通/免费/停留旋转/停留旋转2
	NowNRound       int   //第几轮
	IsOffline       int   //0,1  正常(不显示)/掉线(显示)
	FirstFreeTimes  int   //免费游戏剩余次数
	SecondFreeTimes int   //停留旋转游戏剩余次数
	//中奖统计
	HitPrizePool    []int64 //命中奖池(小奖|中奖|大奖|巨奖)
	WinLineNum      int     //中奖线个数
	WinLineRate     int64   //中奖线总倍率
	WinLineCoin     int64   //中奖线派彩
	GemstoneNum     int     //宝石数量
	GemstoneWinCoin int64   //宝石派彩
	//详情
	Cards            []int32 //元素顺序 横向
	GemstoneRateCoin []int64 //宝石金额 横向
	//中奖线详情
	WinLine []FortuneZhiShenWinLine
	WBLevel int32 //黑白名单等级
	TaxCoin int64 //税收
}
type FortuneZhiShenWinLine struct {
	Id          int   //线号
	EleValue    int32 //元素值
	Num         int   //数量
	Rate        int64 //倍率
	WinCoin     int64 //单线派彩
	WinFreeGame int   //(0,1,2)旋转并停留*3/免费游戏*6/免费游戏*3
}

// 金鼓齐鸣记录详情
type GoldDrumWinLineInfo struct {
	EleValue int   //元素值
	Rate     int64 //倍率
	WinCoin  int64 //单线派彩
	GameType int   //(0,1,2)无/免费游戏/聚宝盆游戏
}
type GoldDrumGameType struct {
	//all
	RoomId     int32 //房间Id
	BasicScore int32 //单注分
	PlayerSnid int32 //玩家id
	BeforeCoin int64 //下注前金额
	AfterCoin  int64 //下注后金额
	BetCoin    int64 //下注金额
	WinCoin    int64 //下注赢取金额总金额
	//Smallgamewinscore int64 //小游戏赢取的分数
	ChangeCoin int64 //本局游戏金额总变化
	FreeTimes  int32 //免费转动次数
	AllWinNum  int32 //中奖的线数
	HitPoolIdx int   //下注索引
	Cards      []int //15张牌

	NowGameState int                   //当前游戏模式(0,1)普通/免费
	HitPrizePool []int64               //命中奖池(多喜小奖|多寿中奖|多禄大奖|多福巨奖)
	WinLineRate  int64                 //中奖线总倍率
	WinLineCoin  int64                 //中奖线派彩
	WinLineInfo  []GoldDrumWinLineInfo //中奖线详情

	NowFreeGameTime int32   //当前免费游戏第几次
	CornucopiaCards []int32 //聚宝盆游戏数据	-1 未开启 0小将 1中奖 2大奖 3巨奖
	IsOffline       bool    //玩家是否掉线	true 掉线
	WBLevel         int32   //黑白名单等级
	TaxCoin         int64   //税收
}

// 金福报喜记录详情
type CopperInfo struct {
	Pos  int32 //铜钱元素索引，从0开始
	Coin int64 //铜钱奖励金币
}
type GoldBlessWinLineInfo struct {
	EleValue int   //元素值
	Rate     int64 //倍率
	WinCoin  int64 //单线派彩
	GameType int   //(0,1,2,3)无/免费游戏/聚宝盆游戏/招福纳财游戏
}
type GoldBlessGameType struct {
	//all
	RoomId     int32 //房间Id
	BasicScore int32 //单注分
	PlayerSnid int32 //玩家id
	BeforeCoin int64 //下注前金额
	AfterCoin  int64 //下注后金额
	BetCoin    int64 //下注金额
	WinCoin    int64 //本局总赢取金额
	//Smallgamewinscore int64 //小游戏赢取的分数
	ChangeCoin int64 //本局游戏金额总变化
	FreeTimes  int32 //免费转动次数
	AllWinNum  int32 //中奖的线数
	HitPoolIdx int   //下注索引
	Cards      []int //15张牌

	NowGameState int                    //当前游戏模式(0,1,2)普通/免费/招福纳财
	HitPrizePool []int64                //命中奖池(多喜小奖|多寿中奖|多禄大奖|多福巨奖)
	WinLineRate  int64                  //中奖线总倍率
	WinLineCoin  int64                  //中奖线派彩
	WinLineInfo  []GoldBlessWinLineInfo //中奖线详情
	CopperNum    int32                  //本局铜钱数量
	CopperCoin   int64                  //本局铜钱金额
	CoppersInfo  []CopperInfo           //铜钱结构

	NowFreeGameTime int32   //当前免费游戏第几次
	CornucopiaCards []int32 //聚宝盆游戏数据	-1 未开启 0小将 1中奖 2大奖 3巨奖
	IsOffline       bool    //玩家是否掉线	true 掉线
	WBLevel         int32   //黑白名单等级
	TaxCoin         int64   //税收
}

// 发发发
type Classic888Type struct {
	//基本信息
	RoomId          int   //房间Id
	BasicScore      int32 //单注
	PlayerSnId      int32 //玩家id
	BeforeCoin      int64 //下注前金额
	AfterCoin       int64 //下注后金额
	ChangeCoin      int64 //金额变化
	TotalBetCoin    int64 //总押注
	TotalLine       int32 //总线数(固定)
	TotalWinCoin    int64 //总派彩
	NowGameState    int   //当前游戏模式(0,1,2)普通/免费/停留旋转
	NowNRound       int   //第几轮
	IsOffline       int   //0,1  正常(不显示)/掉线(显示)
	FirstFreeTimes  int   //免费游戏剩余次数
	SecondFreeTimes int   //停留旋转游戏剩余次数
	//中奖统计
	HitPrizePool   []int64 //命中奖池(小奖|中奖|大奖|巨奖)
	WinLineNum     int     //中奖线个数
	WinLineRate    int64   //中奖线总倍率
	WinLineCoin    int64   //中奖线派彩
	LanternNum     int     //灯笼数量
	LanternWinCoin int64   //灯笼派彩
	//详情
	Cards           []int32 //元素顺序 横向
	LanternRateCoin []int64 //灯笼金额 横向
	//中奖线详情
	WinLine []Classic888WinLine
	WBLevel int32 //黑白名单等级
	TaxCoin int64 //税收
}
type Classic888WinLine struct {
	Id          int   //线号
	EleValue    int32 //元素值
	Num         int   //数量
	Rate        int64 //倍率
	WinCoin     int64 //单线派彩
	WinFreeGame int   //(0,1,2,3)旋转并停留*3/免费游戏*6/免费游戏*9/免费游戏*15
}

// 经典777
type Classic777Type struct {
	//基本信息
	RoomId         int   //房间Id
	BasicScore     int32 //单注
	PlayerSnId     int32 //玩家id
	BeforeCoin     int64 //下注前金额
	AfterCoin      int64 //下注后金额
	ChangeCoin     int64 //金额变化
	TotalBetCoin   int64 //总押注
	TotalLine      int32 //总线数(固定)
	TotalWinCoin   int64 //总派彩
	NowGameState   int   //当前游戏模式(0,1,2)普通/免费/小玛丽
	NowNRound      int   //第几轮
	IsOffline      int   //0,1  正常(不显示)/掉线(显示)
	FirstFreeTimes int   //免费游戏剩余次数
	MaryFreeTimes  int   //停留旋转游戏剩余次数
	//中奖统计
	HitPrizePool   int64 //命中奖池金额
	WinLineNum     int   //中奖线个数
	WinLineRate    int64 //中奖线总倍率
	WinLineCoin    int64 //中奖线派彩
	JackPotNum     int   //777数量
	JackPotWinCoin int64 //777派彩
	//详情
	Cards []int32 //普通游戏/免费游戏
	//玛丽游戏
	MaryOutSide  int32   //外圈
	MaryMidCards []int32 //内圈
	//中奖线详情
	WinLine []Classic777WinLine
	WBLevel int32 //黑白名单等级
	TaxCoin int64 //税收
}
type Classic777WinLine struct {
	Id          int   //线号
	EleValue    int32 //元素值
	Num         int   //数量
	Rate        int64 //倍率
	WinCoin     int64 //单线派彩
	WinFreeGame int   //(0,1,2,3,4,5)小玛丽1/小玛丽2/小玛丽3/免费5/免费8/免费10
}

// 无尽宝藏记录详情
type EndlessTreasureWinLineInfo struct {
	EleValue int   //元素值
	Rate     int64 //倍率
	WinCoin  int64 //单线派彩
	GameType int   //(0,1,2,3)无/免费游戏/聚宝盆游戏/节节高游戏
}
type EndlessTreasureGameType struct {
	//all
	RoomId        int32   //房间Id
	TotalBetScore int32   //总押注
	BasicScore    float32 //单注分
	PlayerSnid    int32   //玩家id
	BeforeCoin    int64   //下注前金额
	AfterCoin     int64   //下注后金额
	BetCoin       int64   //下注金额
	WinCoin       int64   //本局总赢取金额
	ChangeCoin    int64   //本局游戏金额总变化
	FreeTimes     int32   //免费转动次数
	AllWinNum     int32   //中奖的线数
	Cards         []int   //15张牌

	NowGameState int                          //当前游戏模式(0,1,2)普通/免费/节节高
	HitPrizePool []int64                      //命中奖池(小奖|中奖|大奖|巨奖)
	WinLineRate  int64                        //中奖线总倍率
	WinLineCoin  int64                        //中奖线派彩
	WinLineInfo  []EndlessTreasureWinLineInfo //中奖线详情
	CopperNum    int32                        //本局铜钱数量
	CopperCoin   int64                        //本局铜钱金额
	CoppersInfo  []CopperInfo                 //铜钱结构

	NowFreeGameTime int32 //当前免费游戏第几次
	IsOffline       bool  //玩家是否掉线	true 掉线
	AddFreeTimes    int32 //本局新增免费转动次数
	WBLevel         int32 //黑白名单等级
	TaxCoin         int64 //税收
}

// 拉霸类游戏 基础牌局记录
type SlotBaseResultType struct {
	RoomId       int32   //房间Id
	BasicBet     int32   //基本分（单注金额）
	PlayerSnid   int32   //玩家id
	BeforeCoin   int64   //下注前金额
	AfterCoin    int64   //下注后金额
	ChangeCoin   int64   //金额变化
	IsFirst      bool    //是否第一次玩游戏
	IsFree       bool    //是否免费
	TotalBet     int32   //总押注金额
	WinRate      int32   //中奖的倍率
	FreeTimes    int32   //免费转动次数
	AllWinNum    int32   //中奖的总线数
	Tax          int64   //暗税
	WBLevel      int32   //黑白名单等级
	SingleFlag   int32   //0不控1单控赢2单控输
	WinLineScore int64   //中奖线赢取分数
	WinJackpot   int64   //奖池赢取分数
	WinSmallGame int64   //小游戏赢取分数
	WinTotal     int64   //本次游戏总赢取(本次游戏赢取总金额=中奖线赢取+奖池赢取+小游戏赢取)
	Cards        []int32 //15张牌
}

type GameResultLog struct {
	BaseResult *SlotBaseResultType
	AllLine    int32   //线路数
	UserName   string  //昵称
	WinLines   []int   //赢分的线
	BetLines   []int64 //下注的线
}

// 幸运骰子
type LuckyDiceType struct {
	RoomId     int32   //房间Id
	RoundId    int32   //局数编号
	BaseScore  int32   //底分
	PlayerSnid int32   //玩家id
	UserName   string  //昵称
	BeforeCoin int64   //本局前金额
	AfterCoin  int64   //本局后金额
	ChangeCoin int64   //金额变化
	Bet        int64   //总押注数
	Refund     int64   //返还押注数
	Award      int64   //获奖金额
	BetSide    int32   //压大压小 0大 1小
	Dices      []int32 //3个骰子值
	Tax        int64   //赢家税收
	WBLevel    int32   //黑白名单等级
}

type CandyType struct {
	RoomId          int32 //房间Id
	SpinID          int32 //局数编号
	BasicScore      int32 //基本分
	PlayerSnid      int32 //玩家id
	BeforeCoin      int64 //下注前金额
	AfterCoin       int64 //下注后金额
	ChangeCoin      int64 //金额变化
	Score           int32 //总押注数
	Tax             int64 //暗税
	IsFirst         bool
	AllWinNum       int32   //中奖的线数
	WinScore        int32   //中奖的倍率
	AllLine         int32   //线路数
	Cards           []int32 //9张牌
	BetLines        []int64 //下注的线
	WBLevel         int32   //黑白名单等级
	UserName        string  // 昵称
	TotalPriceValue int64   // 总赢分
	WinLines        []int   // 赢分的线
	WinJackpot      int64   // 赢奖池分数
}
type MiniPokerType struct {
	RoomId          int32 //房间Id
	SpinID          int32 //局数编号
	BasicScore      int32 //基本分
	PlayerSnid      int32 //玩家id
	BeforeCoin      int64 //下注前金额
	AfterCoin       int64 //下注后金额
	ChangeCoin      int64 //金额变化
	Score           int32 //总押注数
	Tax             int64 //暗税
	IsFirst         bool
	WinScore        int32   //中奖的倍率
	Cards           []int32 //5张牌
	WBLevel         int32   //黑白名单等级
	UserName        string  // 昵称
	TotalPriceValue int64   // 总赢分
	WinJackpot      int64   // 赢奖池分数
}
type CaoThapType struct {
	RoomId          int32 //房间Id
	BasicScore      int32 //基本分
	PlayerSnid      int32 //玩家id
	BeforeCoin      int64 //下注前金额
	AfterCoin       int64 //下注后金额
	ChangeCoin      int64 //金额变化
	Score           int32 //总押注数
	Tax             int64 //暗税
	IsFirst         bool
	Cards           []int32          //翻的牌
	WBLevel         int32            //黑白名单等级
	UserName        string           // 昵称
	TotalPriceValue int64            // 总赢分
	WinJackpot      int64            // 赢奖池分数
	BetInfo         []CaoThapBetInfo // 每次下注信息
}
type CaoThapBetInfo struct {
	TurnID     int32 // 操作ID
	TurnTime   int64 // 操作时间
	BetValue   int64 // 下注金额
	Card       int32 // 牌值
	PrizeValue int64 // 赢分
}

// 21点
type BlackJackType struct {
	RoomId          int32              //房间ID
	RoomType        int32              //房间类型
	NumOfGames      int                //当前局数
	PlayerCount     int                //玩家数量
	PlayerData      []*BlackJackPlayer //玩家信息
	BankerCards     []int32            //庄家牌
	BankerCardType  int32              //牌型 1：黑杰克 2：五小龙 3：其它点数 4：爆牌
	BankerCardPoint []int32            //点数
	BetCoin         int64              //总下注
	GainCoinTax     int64              //总输赢分(税前)
}

type BlackJackCardInfo struct {
	Cards         []int32 //闲家牌
	CardType      int32   //牌型 1：黑杰克 2：五小龙 3：其它点数 4：爆牌
	CardPoint     []int32 //点数
	BetCoin       int64   //下注
	GainCoinNoTax int64   //总输赢分(税后)
	IsWin         int32   //输赢 1赢 0平 -1输
}

type BlackJackPlayer struct {
	UserId        int32               //玩家ID
	UserIcon      int32               //玩家头像
	Platform      string              `json:"-"`
	Channel       string              `json:"-"`
	Promoter      string              `json:"-"`
	PackageTag    string              `json:"-"`
	InviterId     int32               `json:"-"`
	WBLevel       int32               //黑白名单等级
	IsRob         bool                //是否是机器人
	Flag          int                 //标识
	IsFirst       bool                //是否第一次
	Hands         []BlackJackCardInfo //牌值
	IsWin         int32               //输赢
	GainCoinNoTax int64               //总输赢分(税后)
	Tax           int64               //税，不一定有值，只是作为一个临时变量使用
	BaoCoin       int64               //保险金
	BaoChange     int64               //保险金输赢分
	BetCoin       int64               //下注额
	BetChange     int64               //下注输赢分
	Seat          int                 //座位号
}

type DezhouPots struct {
	BetTotal int64             //边池下注
	Player   []DezhouPotPlayer //边池的玩家
}
type DezhouPotPlayer struct {
	Snid  int32 //玩家ID
	IsWin int32 //边池输赢
}

// 德州牌局记录
type DeZhouUserOp struct {
	Snid        int32   // 操作人
	Op          int32   // 操作类型 见 dezhoupoker.proto
	Stage       int     // 所处牌局阶段 见 constants.go
	Chip        int64   // 操作筹码, (不下注为0)
	ChipOnTable int64   // 操作后桌子上筹码
	Round       int32   // 轮数
	Sec         float64 // 操作时距离本局开始时的秒数
	TargetId    int32   // 操作对象ID(没其他玩家为对象为0)
}

// 德州
type DezhouType struct {
	RoomId      int32           //房间ID
	RoomType    int32           //房间类型
	NumOfGames  int32           //当前局数
	BankId      int32           //庄家ID
	PlayerCount int             //玩家数量
	BaseScore   int32           //底分
	BaseCards   []int32         //公牌 只限于德州用
	PlayerData  []DezhouPerson  //玩家信息
	Pots        []DezhouPots    //边池情况
	Actions     string          //牌局记录
	UserOps     []*DeZhouUserOp //牌局记录new
}
type DezhouPerson struct {
	UserId        int32   //玩家ID
	UserIcon      int32   //玩家头像
	Platform      string  `json:"-"`
	Channel       string  `json:"-"`
	Promoter      string  `json:"-"`
	PackageTag    string  `json:"-"`
	InviterId     int32   `json:"-"`
	WBLevel       int32   //黑白名单等级
	IsRob         bool    //是否是机器人
	IsFirst       bool    //是否第一次
	IsLeave       bool    //中途离开
	InitCard      []int32 //初始牌值
	Cardinfo      []int32 //牌值
	IsWin         int32   //输赢
	GainCoinNoTax int64   //总输赢分(税后)
	Tax           int64   //税，不一定有值，只是作为一个临时变量使用
	BetTotal      int64   //用户当局总下注
	IsAllIn       bool    //是否全下
	RoundFold     int32   //第几轮弃牌
	CardInfoEnd   []int32 //结算时的牌型
	Seat          int     //座位号
}

// tienlen
type TienLenType struct {
	GameId      int             //游戏id
	RoomId      int32           //房间ID
	RoomType    int32           //房间类型
	NumOfGames  int32           //当前局数
	BankId      int32           //房主ID
	PlayerCount int             //玩家数量
	BaseScore   int32           //底分
	TaxRate     int32           //税率（万分比）
	PlayerData  []TienLenPerson //玩家信息
	RoomMode    int
}
type TienLenPerson struct {
	UserId        int32           //玩家ID
	UserIcon      int32           //玩家头像
	Platform      string          `json:"-"`
	Channel       string          `json:"-"`
	Promoter      string          `json:"-"`
	PackageTag    string          `json:"-"`
	InviterId     int32           `json:"-"`
	WBLevel       int32           //黑白名单等级
	IsRob         bool            //是否是机器人
	IsFirst       bool            //是否第一次
	IsLeave       bool            //中途离开
	IsWin         int32           //输赢
	Seat          int             //座位号
	GainCoin      int64           //手牌输赢分(税后)
	BombCoin      int64           //炸弹输赢分(税后)
	BillCoin      int64           //最终得分(税后)
	GainTaxCoin   int64           //手牌税收
	BombTaxCoin   int64           //炸弹税收
	BillTaxCoin   int64           //最终税收
	DelOrderCards map[int][]int32 //已出牌
	CardInfoEnd   []int32         //结算时的牌型
	IsTianHu      bool            //是否天胡
}
