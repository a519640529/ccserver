package webapi

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type PlayerStatement struct {
	Platform     string `json:"platform"`       //平台id
	PackageId    string `json:"packageTag"`     //包标识
	SnId         int32  `json:"snid"`           //玩家id
	IsBind       int32  `json:"is_bind"`        //是否是正式账号 1:正式用户 0:非正式
	TotalWin     int32  `json:"total_win"`      //折算后赢钱数
	TotalLose    int32  `json:"total_lose"`     //折算后输钱数
	TotaSrclWin  int32  `json:"total_src_win"`  //赢钱数
	TotaSrclLose int32  `json:"total_src_lose"` //输钱数
}

type PlayerStatementSrc struct {
	Platform     string  `json:"platform"`       //平台id
	PackageId    string  `json:"packageTag"`     //包标识
	SnId         int32   `json:"snid"`           //玩家id
	IsBind       int32   `json:"is_bind"`        //是否是正式账号 1:正式用户 0:非正式
	TotalWin     float64 `json:"total_win"`      //折算后赢钱数
	TotalLose    float64 `json:"total_lose"`     //折算后输钱数
	TotaSrclWin  float64 `json:"total_src_win"`  //赢钱数
	TotaSrclLose float64 `json:"total_src_lose"` //输钱数
}

// 推广信息
type PromoterData struct {
	Platform     int32  `json:"platform"`      //平台id
	PromoterId   int32  `json:"promoter_id"`   //推广员id
	ChannelId    int32  `json:"channel_id"`    //渠道id
	Spreader     int32  `json:"spreader"`      //推广人id
	PromoterTree int32  `json:"promoter_tree"` //无级推广树id
	Tag          string `json:"tag"`           //包标识
}

type ImgVerifyMsg struct {
	ImgBase string `json:"img_base"`
	Length  int32  `json:"length"`
	Code    string `json:"code"`
}

// 获取包对应的平台和上级关系
func API_SendSms(appId string, snid int32, phone string, content string, platform string) error {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["phone"] = phone
	//params["content"] = content
	//params["platform"] = platform
	//buff, err := getRequest(appId, "/send_sms", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg string
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//if err != nil {
	//	return err
	//}
	//if result.Tag != 0 {
	//	return errors.New(result.Msg)
	//} else {
	//	return nil
	//}
	return nil
}

// 获取包对应的平台和上级关系
func API_PackageList(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/package_list", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 游戏返利配置列表
func API_GetGameRebateConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/getGameRebateConfig", params, "http", time.Duration(time.Second*120))
	return nil, nil
}

////获取游戏分组列表
//func API_GetGameGroupData(appId string) ([]byte, error) {
//	//params := make(map[string]string)
//	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
//	//return getRequest(appId, "/game_config_group_list", params, "http", DEFAULT_TIMEOUT)
//	return nil, nil
//}

// 获取公告列表
func API_GetBulletData(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/notice_list", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 获取招商列表
func API_GetCustomerData(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/agent_customer_list", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

////平台详细配置
//func API_GetPlatformConfigData(appId string) ([]byte, error) {
//	//params := make(map[string]string)
//	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
//	//return getRequest(appId, "/game_config_list", params, "http", time.Duration(time.Second*120))
//	return nil, nil
//}

// 获得代理配置
func API_GetPromoterConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/game_promoter_list", params, "http", time.Duration(time.Second*120))
	return nil, nil
}

// 黑名单列表
func API_GetBlackData(appId string, page int) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//params["page"] = strconv.Itoa(page)
	//return getRequest(appId, "/black_list", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 团队信息
func API_GetSpreadPlayer(appId string, SnId int32, platform string) ([]byte, error) {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(SnId))
	//params["platform"] = platform
	//return getRequest(appId, "/spread_player", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 获取公共黑名单信息
func API_GetCommonBlackData(appId string, page int) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//params["page"] = strconv.Itoa(page)
	//return getRequest(appId, "/black_list_common", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 支付方式
func API_GetPayList(appId string, platform string, guestUser, newUser, userVip int, logicLevels []int32, os, snid int32) ([]byte, error) {
	//params := make(map[string]string)
	//params["platform"] = platform
	//params["guest"] = strconv.Itoa(guestUser)
	//params["new"] = strconv.Itoa(newUser)
	//params["vip"] = strconv.Itoa(userVip)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//params["os"] = strconv.Itoa(int(os))
	//params["snid"] = strconv.Itoa(int(snid))
	//
	//var logiclevel string
	//for _, v := range logicLevels {
	//	logiclevel = logiclevel + fmt.Sprintf("%v,", v)
	//}
	//if len(logiclevel) > 0 {
	//	logiclevel = logiclevel[:len(logiclevel)-1]
	//	params["logiclevel"] = logiclevel
	//}
	//return getRequest(appId, "/pay_platform_list", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// CToAPITransfer
func API_Transfer(appId string, info string) ([]byte, error) {
	//params := make(map[string]string)
	//params["info"] = info
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/c2api_transfer", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

/*
  int32 Snid = 1;//用户id
  int32 Platform = 2;//用户id
  int32 Type = 3;//商品类型
  int32 GoodsId = 4;//商品ID
  int32 VCard = 5;//消耗V卡
  string GoodsName = 6;//兑换物品名称
  string UserName = 7;//兑换人姓名
  string Mobile = 8;//兑换人手机号
  string Comment = 9; //备注信息
*/

// 兑换订单
func API_CreateExchange(appId string, body proto.Message) ([]byte, error) {

	return postRequest(appId, "/create_exchange_order", nil, body, "http", DEFAULT_TIMEOUT)
}

// 获取兑换列表
func API_ExchangeList(appId string, body proto.Message) ([]byte, error) {

	return postRequest(appId, "/get_exchange_shop", nil, body, "http", DEFAULT_TIMEOUT)
}

// 获取兑换记录
func API_ExchangeRecord(appId string, body proto.Message) ([]byte, error) {

	return postRequest(appId, "/get_exchange_order", nil, body, "http", DEFAULT_TIMEOUT)
}

// 积分商城兑换订单
func API_GradeShopExchangeList(appId, platform string, snid, page, count int32) ([]byte, error) {
	//params := make(map[string]string)
	//params["Platform"] = platform
	//params["SnId"] = strconv.Itoa(int(snid))
	//params["Page"] = strconv.Itoa(int(page))
	//if page > 0 {
	//	params["Count"] = strconv.Itoa(int(count))
	//}
	//params["Ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/list_exchange_shop_order", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 积分商城重置商品库存
func API_GradeShopInitShop(appId string, InitShopIds string) ([]byte, error) {
	//params := make(map[string]string)
	//params["InitShopIds"] = InitShopIds
	//params["Ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/zero_to_reset_exshop", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 发送短信
func API_SendSMSCode(appId string, tel string, code string) error {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//params["phone"] = tel
	//params["code"] = code
	//buff, err := getRequest(appId, "/send_sms", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg string
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//if err != nil {
	//	return err
	//}
	//if result.Tag != 0 {
	//	return errors.New(result.Msg)
	//} else {
	//	return nil
	//}
	return nil
}

// 支付订单
func API_CreateOrder(snid int32, platform, channel, promoter, packageTag, os string, coin, safe, count int64, ip, code, appId, extCode string,
	orderid int32, guestUser, newUser, userVip int, promoterTree int32, regTime time.Time, telephonePromoter int32, deviceId string) (string, error) {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["platform"] = platform
	//params["packageTag"] = packageTag
	//params["p_id"] = strconv.Itoa(int(orderid))
	//params["os"] = DeviceOs(os)
	//params["before_coin"] = strconv.Itoa(int(coin))
	//params["before_safe"] = strconv.Itoa(int(safe))
	//params["count"] = strconv.Itoa(int(count))
	//params["ip"] = ip
	//params["p_uid"] = code
	//params["channel"] = channel
	//params["promoter"] = promoter
	//params["promoter_tree"] = strconv.Itoa(int(promoterTree))
	//params["ext"] = extCode
	//params["guest"] = strconv.Itoa(guestUser)
	//params["new"] = strconv.Itoa(newUser)
	//params["vip"] = strconv.Itoa(userVip)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//params["register_time"] = strconv.Itoa(int(regTime.Unix()))
	//params["telephone_promoter"] = strconv.Itoa(int(telephonePromoter))
	//params["deviceid"] = deviceId
	//buff, err := getRequest(appId, "/create_order", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return "", err
	//}
	//type UrlData struct {
	//	Url string
	//	Err string
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg UrlData
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//if err != nil {
	//	return "", err
	//}
	//if result.Tag != 0 {
	//	return "", errors.New(result.Msg.Err)
	//} else {
	//	return result.Msg.Url, nil
	//}
	return "", nil
}

// 兑换订单
func API_CreateExchangeOrder(snid int32, platform string, before_coin, before_safe, exchange_count, tag, bank_id int64,
	account, username, ip string, win_times, lose_times, win_total, lose_total int64, os, appId, channel, agent string,
	promoterTree int32, packageTag string, giveGold int64, forceTax int64, needFlow int64, curFlow int64, payts int64,
	payEndTs int64, newUser int, telephonePromoter int32, deviceId string) (error, int64) {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["platform"] = platform
	//params["before_coin"] = strconv.Itoa(int(before_coin))
	//params["before_safe"] = strconv.Itoa(int(before_safe))
	//params["exchange_count"] = strconv.Itoa(int(exchange_count))
	//params["tag"] = strconv.Itoa(int(tag))
	//params["bank_id"] = strconv.Itoa(int(bank_id))
	//params["account"] = account
	//params["username"] = username
	//params["ip"] = ip
	//params["win_times"] = strconv.Itoa(int(win_times))
	//params["lose_times"] = strconv.Itoa(int(lose_times))
	//params["win_total"] = strconv.Itoa(int(win_total))
	//params["lose_total"] = strconv.Itoa(int(lose_total))
	//params["os"] = DeviceOs(os)
	//params["channel"] = channel
	//params["promoter"] = agent
	//params["promoter_tree"] = strconv.Itoa(int(promoterTree))
	//params["packageTag"] = packageTag
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//params["force_tax"] = strconv.Itoa(int(forceTax))
	//params["give_gold"] = strconv.Itoa(int(giveGold))
	//params["need_flow"] = strconv.Itoa(int(needFlow))
	//params["cur_flow"] = strconv.Itoa(int(curFlow))
	//params["new"] = strconv.Itoa(newUser)
	//params["pay_ts"] = strconv.Itoa(int(payts))
	//params["pay_endts"] = strconv.Itoa(int(payEndTs))
	//params["telephone_promoter"] = strconv.Itoa(int(telephonePromoter))
	//params["deviceid"] = deviceId
	//buff, err := getRequest(appId, "/create_exchange_order", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return err, 0
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg interface{}
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//if err != nil {
	//	return err, 0
	//}
	//if result.Tag != 0 {
	//	errMsg := fmt.Sprintf("Create exchange order result failed._%v", result.Msg)
	//	return errors.New(errMsg), 0
	//} else {
	//	var id int64
	//	switch result.Msg.(type) {
	//	case string:
	//		c, cr := strconv.Atoi(result.Msg.(string))
	//		if cr == nil {
	//			id = int64(c)
	//		}
	//	case float64:
	//		id = int64(result.Msg.(float64))
	//	}
	//	return nil, id
	//}
	return nil, 0
}

// 商品兑换订单
func API_CreateGradeShopExchangeOrder(LogId, appId, platform string, snid, ShopId int32, Ip, ReceiveName, ReceiveTel, ReceiveAddr string,
	LastGrade int64) error {
	//params := make(map[string]string)
	//params["Platform"] = platform                       //平台号
	//params["SnId"] = strconv.Itoa(int(snid))            //玩家Id
	//params["LogId"] = LogId                             //订单id
	//params["ShopId"] = strconv.Itoa(int(ShopId))        //商品Id
	//params["Ip"] = Ip                                   //IP地址
	//params["ReceiveName"] = ReceiveName                 //收货人名字
	//params["ReceiveTel"] = ReceiveTel                   //收货人电话
	//params["ReceiveAddr"] = ReceiveAddr                 //收货人地址
	//params["LastGrade"] = strconv.Itoa(int(LastGrade))  //玩家兑换完剩余积分
	//params["Ts"] = strconv.Itoa(int(time.Now().Unix())) //创建时间
	//buff, err := getRequest(appId, "/create_exchange_shop_order", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg string
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//if err != nil {
	//	return err
	//}
	//if result.Tag != 0 {
	//	errMsg := fmt.Sprintf("Create GradeShop exchange order result failed._%v", result.Msg)
	//	return errors.New(errMsg)
	//} else {
	//	return nil
	//}
	return nil
}

// 税收分成
func API_TaxDivide(snid int32, platform, channel, promoter, packageTag string, tax int64, appId string, gameid, gamemode int, gamefreeid, promoterTree int32) (int32, error) {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["platform"] = platform
	//params["tax"] = fmt.Sprintf("%v", tax)
	//params["channel"] = channel
	//params["promoter"] = promoter
	//params["promoter_tree"] = strconv.Itoa(int(promoterTree))
	//params["packageTag"] = packageTag
	//params["gameId"] = strconv.Itoa(gameid)
	//params["modeId"] = strconv.Itoa(gamemode)
	//params["gamefreeId"] = strconv.Itoa(int(gamefreeid))
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//buff, err := getRequest(appId, "/tax_divide", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return -1, err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg interface{}
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//return result.Tag, err
	return 0, nil
}

// 流水推送
func API_SpreadAccount(appId string, gamefreeId int32, data []*PlayerStatement) (int32, error) {
	//d, err := json.Marshal(data)
	//if err != nil {
	//	return -1, err
	//}
	//params := make(map[string]string)
	//params["gamefreeId"] = strconv.Itoa(int(gamefreeId))
	//params["data"] = string(d[:])
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//buff, err := getRequest(appId, "/spread_account_push", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return -1, err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg interface{}
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//return result.Tag, err
	return 0, nil
}

// 系统赠送
func API_SystemGive(snid int32, platform, channel, promoter string, ammount, tag int32, appId string, packageTag string) (int32, error) {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["platform"] = platform
	//params["channel"] = channel
	//params["promoter"] = promoter
	//params["package_tag"] = packageTag
	//params["amount"] = fmt.Sprintf("%v", ammount)
	//params["tg"] = strconv.Itoa(int(tag))
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//buff, err := getRequest(appId, "/system_give", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return -1, err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg interface{}
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//return result.Tag, err
	return 0, nil
}

// 首次登录通知
func API_PlayerEvent(event int, platform, packageTag string, snid int32, channel string, promoter string, promoterTree int32, isCreate int, isNew int, isBind int, appId string) (int32, error) {
	//params := make(map[string]string)
	//params["event"] = strconv.Itoa(event)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["platform"] = platform
	//params["channel"] = channel
	//params["promoter"] = promoter
	//params["promoter_tree"] = strconv.Itoa(int(promoterTree))
	//params["packageTag"] = packageTag
	//params["isCreate"] = strconv.Itoa(isCreate)
	//params["isNew"] = strconv.Itoa(isNew)
	//params["isBind"] = strconv.Itoa(isBind)
	//params["create_time"] = strconv.Itoa(int(time.Now().Unix()))
	//buff, err := getRequest(appId, "/player_event", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return -1, err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg interface{}
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//return result.Tag, err
	return 0, nil
}

// 推送全民推广关系链
func API_PushSpreadLink(snid int32, platform, packageTag string, inviterId int, isBind, isForce int, appId string) (int32, error) {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["platform"] = platform
	//params["packageTag"] = packageTag
	//params["inviter"] = strconv.Itoa(inviterId)
	//params["is_bind"] = strconv.Itoa(isBind)
	//params["is_force"] = strconv.Itoa(isForce)
	//buff, err := getRequest(appId, "/push_spread_link", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return -1, err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg interface{}
	//}
	////fmt.Println("push_spread_link Response:", string(buff[:]))
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//return result.Tag, err
	return 0, nil
}

// 推送全民推广关系链
func API_PushInviterIp(snid, inviterId, promoterTree int32, promoter string, bundleId, ip string, os string, appId string) (int32, *PromoterData, error) {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["bundle_id"] = bundleId
	//params["ip"] = ip
	//params["os"] = DeviceOs(os)
	//params["inviter"] = strconv.Itoa(int(inviterId))
	//params["promoter_tree"] = strconv.Itoa(int(promoterTree))
	//params["promoter"] = promoter
	//
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//buff, err := getRequest(appId, "/push_inviter_ip", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return -1, nil, err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg *PromoterData
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//if err != nil {
	//	type ErrResult struct {
	//		Tag int32
	//		Msg string
	//	}
	//	errRes := ErrResult{}
	//	errTag := json.Unmarshal(buff, &errRes)
	//	if errTag == nil {
	//		logger.Logger.Warnf("API_PushInviterIp response tag:%v msg:%v err:%v snid:%v packagetag:%v ip:%v ", errRes.Tag, errRes.Msg, snid, bundleId, ip)
	//	}
	//}
	//return result.Tag, result.Msg, err
	return 0, nil, nil
}

// 推送全民推广关系链
func API_PushInvitePromoter(snid int32, promoter string, appId string) (int32, error) {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["promoter"] = promoter
	//
	//buff, err := getRequest(appId, "/bind_pay_exchange", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return -1, err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg interface{}
	//}
	////fmt.Println("push_spread_link Response:", string(buff[:]))
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//return result.Tag, err
	return 0, nil
}

// 无限代信息校验
func API_ValidPromoterTree(snid int32, packageTag string, promoterTree int32, appId string) (int32, error) {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["package_tag"] = packageTag
	//params["promoter_tree"] = strconv.Itoa(int(promoterTree))
	//buff, err := getRequest(appId, "/valid_promoter_tree", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return -1, err
	//}
	//type ApiResult struct {
	//	Tag int32
	//	Msg interface{}
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//return result.Tag, err
	return 0, nil
}

// 玩家透传API
func API_PlayerPass(snid int32, platform, channel, promoter, apiName, param, appId string, logicLvls []int32) (string, error) {
	//params := make(map[string]string)
	//if param != "" {
	//	cparam := make(map[string]interface{})
	//	err := json.Unmarshal([]byte(param), &cparam)
	//	if err == nil {
	//		for k, v := range cparam {
	//			switch v.(type) {
	//			case float64:
	//				params[k] = fmt.Sprintf("%v", int64(v.(float64)))
	//			default:
	//				params[k] = fmt.Sprintf("%v", v)
	//			}
	//		}
	//	}
	//}
	//params["snid"] = strconv.Itoa(int(snid))
	//params["platform"] = platform
	//params["channel"] = channel
	//params["promoter"] = promoter
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//buff, err := getRequest(appId, apiName, params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return "", err
	//}
	//return string(buff[:]), err
	return "", nil
}

// 系统透传API
func API_SystemPass(apiName, param, appId string) (string, error) {
	//params := make(map[string]string)
	//if param != "" {
	//	cparam := make(map[string]interface{})
	//	err := json.Unmarshal([]byte(param), &cparam)
	//	if err == nil {
	//		for k, v := range cparam {
	//			switch v.(type) {
	//			case float64:
	//				params[k] = fmt.Sprintf("%v", int64(v.(float64)))
	//			default:
	//				params[k] = fmt.Sprintf("%v", v)
	//			}
	//		}
	//	}
	//}
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//buff, err := getRequest(appId, apiName, params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return "", err
	//}
	//return string(buff[:]), err
	return "", nil
}

// 定时任务
func API_ScheduleDo(appId, action string, dura time.Duration) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//buff, err := getRequest(appId, action, params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return buff, err
	//}
	//return buff, nil
	return nil, nil
}

// 平台详细配置
func API_GetPlatformSignConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/active_sign_config", params, "http", time.Duration(time.Second*120))
	return nil, nil
}

// 活跃任务
func API_GetTaskConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/active_task_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 财神任务
func API_GetGoldTaskConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/active_goldtask_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 财神降临
func API_GetGoldComeConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/active_goldcome_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// vip活动配置
func API_GetActVipConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/active_vip_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 首充奖励活动配置
func API_GetActFPayConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/active_fpay_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 活动相关统一配置
func API_GetActConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/active_act_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 支付活动配置
func API_GetPayActConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/active_payact_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 获取在线奖励活动配置
func API_GetOnlineRewardConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	////params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	////params["platform"] = platform
	//return getRequest(appId, "/online_reward_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 获取幸运转盘活动配置
func API_GetLuckyTurntableConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	////params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	////params["platform"] = platform
	//return getRequest(appId, "/luckly_turntable_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 获取微信分享彩金配置
func API_GetWeiXinShareConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//return getRequest(appId, "/weixin_share_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 获取余额宝配置
func API_GetYebConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	////params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	////params["platform"] = platform
	//return getRequest(appId, "/yeb_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 获取周卡月卡配置
func API_GetCardConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//return getRequest(appId, "/card_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 获取阶梯充值配置
func API_GetStepRechargeConfig(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//return getRequest(appId, "/step_recharge_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 获取自动黑白名单配置
func API_GetAutoBWConfig(appId string, page int) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//params["page"] = strconv.Itoa(page)
	//return getRequest(appId, "/autobw_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 获取包对应的平台和上级关系
func API_GetImgVerify(appId string, phone string) (*ImgVerifyMsg, error) {
	//params := make(map[string]string)
	//params["phone"] = phone
	//buff, err := getRequest(appId, "/get_img_verify", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return nil, err
	//}
	//
	//type ApiResult struct {
	//	Tag int32        `json:"Tag"`
	//	Msg ImgVerifyMsg `json:"Msg"`
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//if err != nil {
	//	return nil, err
	//}
	//if result.Tag != 0 {
	//	return nil, errors.New("Get Image Verify Failed.")
	//} else {
	//	return &result.Msg, nil
	//}
	return nil, nil
}

type RebateImgUrlMsg struct {
	Wx    string `json:"wx"`
	Image string `json:"image"`
}

// 获取包对应的平台和上级关系
func API_GetRebateImgUrl(appId string, platform string) (string, string, error) {
	//params := make(map[string]string)
	//params["platform"] = platform
	//buff, err := getRequest(appId, "/get_weixin_by_range", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return "", "", err
	//}
	//
	//type ApiResult struct {
	//	Tag int32           `json:"Tag"`
	//	Msg RebateImgUrlMsg `json:"Msg"`
	//}
	//result := ApiResult{}
	//err = json.Unmarshal(buff, &result)
	//if err != nil {
	//	return "", "", err
	//}
	//if result.Tag != 0 {
	//	return "", "", errors.New("Get Image Url Failed.")
	//} else {
	//	return result.Msg.Image, result.Msg.Wx, nil
	//}
	return "", "", nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////
//请求参数如下：
//int32 snid 玩家id
//====================================================================
//int64 showTypeId = 2	棋牌游戏
//		showTypeId = 3	捕鱼游戏（废弃,不在使用）
//		showTypeId = 4	电子游艺
//		showTypeId = 5	真人视讯
//		showTypeId = 6	彩票游戏
//		showTypeId = 380720001	WWG 大师捕鱼		在三方中的游戏id=3
//		showTypeId = 391590001	FG	捕鱼嘉年华3D	在三方中的游戏id=fish_3D
//		showTypeId = 391600001	FG	捕鸟达人		在三方中的游戏id=fish_bn
//		showTypeId = 391610001	FG	欢乐捕鱼		在三方中的游戏id=fish_hl
//		showTypeId = 391620001	FG	美人捕鱼		在三方中的游戏id=fish_mm
//		showTypeId = 391630001	FG	天天捕鱼		在三方中的游戏id=fish_tt
//		showTypeId = 391640001	FG	雷霆战警		在三方中的游戏id=fish_zj
//		showTypeId = 391650001	FG	魔法王者		在三方中的游戏id=fish_mfwz
//======================================================================
//int64 timeIndex  		0.全部 1.今天 2.昨天 3.一个月内
//int64 thirdId 		那个第三方 0=全部 WWG平台=38  FG平台=39 体育赛事=41 VR彩票=43 真人视讯=28
//int32 pageNo 			当前页
//int32 pageCount 		共几页

// API返回的每条格式如下：
// int64 Ts		 		//注单时间戳
// string ThirdPltName 	//三方平台名字
// string ThirdGameId 	//三方游戏id
// string ThirdGameName 	//三方游戏名字
// string SysGamefreeid  //我们系统的游戏id,这个不返回占位
// string RecordId		//注单号
// int64 BetCoin			//投注金额
// int64 ReceivedCoin 	//已派奖
func API_GetThirdDetail(appId, pltform string, snid, pageNo, pageCount, showTypeId, timeIndex, thirdId int32) (error, []byte) {
	return nil, nil
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["pageNo"] = strconv.Itoa(int(pageNo))
	//params["pageCount"] = strconv.Itoa(int(pageCount))
	//params["showTypeId"] = strconv.Itoa(int(showTypeId))
	//params["timeIndex"] = strconv.Itoa(int(timeIndex))
	//params["thirdId"] = strconv.Itoa(int(thirdId))
	//params["platform"] = pltform
	//buff, err := getRequest(appId, "/third_detail", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return err, buff
	//}
	//return nil, buff

	//下面是以post方式请求，备用
	//var client = &http.Client{}
	//var signupDataBuff []byte
	//var err error
	//ts := time.Now().Unix()
	//if params != nil {
	//	signupDataBuff, err = json.Marshal(params)
	//	if err != nil {
	//		return err, nil
	//	}
	//}
	//
	////fmt.Println(string(signupDataBuff))
	//sign := MakeMd5String(fmt.Sprintf("%v;%v;%v", ts, string(signupDataBuff), appId))
	//url := fmt.Sprintf("%v?ts=%v&sign=%v", Config.GameApiURL+"/third_detail", ts, sign)
	//logger.Trace("API_GetThirdDetail request url:", url)
	//request, err := http.NewRequest("POST", url, bytes.NewReader(signupDataBuff))
	//if err != nil {
	//	return err, nil
	//}
	//client.Timeout = DEFAULT_TIMEOUT
	//resp, err := client.Do(request)
	//if err != nil {
	//	logger.Errorf("Snid=%v API_GetThirdDetail api :%v", snid, err)
	//	return err, nil
	//}
	//defer resp.Body.Close()
	//
	//if resp.StatusCode == 200 {
	//	buff, err := ioutil.ReadAll(resp.Body)
	//	if err != nil {
	//		return err, nil
	//	}
	//	return nil, buff
	//}
	//io.Copy(ioutil.Discard, resp.Body)
	//return fmt.Errorf("API_GetThirdDetail HttpStatusCode:%d", resp.StatusCode), nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////////////////////////
//请求参数如下：
//int32 snid 玩家id
//====================================================================
//string hotGameNameSet = "狂欢派对","复仇者联盟","百人牛牛"
//在mysql 中使用in来查询
//======================================================================
//int64 timeIndex  		0.全部 1.今天 2.昨天 3.一个月内
//int64 thirdId 		那个第三方 WWG平台=38  FG平台=39 体育赛事=41 VR彩票=43 真人视讯=28
//int32 pageNo 			当前页
//int32 pageCount 		共几页

// API返回的每条格式如下：
// int64 Ts		 		//注单时间戳
// string ThirdPltName 	//三方平台名字
// string ThirdGameId 	//三方游戏id
// string ThirdGameName 	//三方游戏名字
// string SysGamefreeid  //我们系统的游戏id,这个不返回占位
// string RecordId		//注单号
// int64 BetCoin			//投注金额
// int64 ReceivedCoin 	//已派奖
func API_GetThirdHotGameDetail(appId, pltform string, snid, pageNo, pageCount, timeIndex, thirdId int32, hotGameNameSet []string) (error, []byte) {
	//params := make(map[string]string)
	//params["snid"] = strconv.Itoa(int(snid))
	//params["pageNo"] = strconv.Itoa(int(pageNo))
	//params["pageCount"] = strconv.Itoa(int(pageCount))
	//params["hotGameNameSet"] = strings.Join(hotGameNameSet, ",")
	//params["timeIndex"] = strconv.Itoa(int(timeIndex))
	//params["thirdId"] = strconv.Itoa(int(thirdId))
	//params["platform"] = pltform
	//buff, err := getRequest(appId, "/hot_third_game", params, "http", DEFAULT_TIMEOUT)
	//if err != nil {
	//	return err, buff
	//}
	//return nil, buff
	return nil, nil
}

// ////////////////////////////////////////////////////////////////////////////////////////////////
func API_GetRandCoinData(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/activity_read_envelope_config", params, "http", time.Duration(time.Second*120))
	return nil, nil
}

// ////////////////////////////////////////////////////////////////////////////////////////////////
func API_GetChildData(appId string, platform string, snid int32, ts1, ts2 int64) ([]byte, error) {
	//params := make(map[string]string)
	//params["platform"] = platform
	//params["snid"] = strconv.Itoa(int(snid))
	//params["start_time"] = strconv.Itoa(int(ts1))
	//params["end_time"] = strconv.Itoa(int(ts2))
	//return getRequest(appId, "/spread_child_list", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

func API_PushPlayerSingleAdjustCount(appId string, id int32, count int32) ([]byte, error) {
	//params := make(map[string]string)
	//params["id"] = strconv.Itoa(int(id))
	//params["times"] = strconv.Itoa(int(count))
	//return getRequest(appId, "/game_ctrl_alone_times_push", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}
func API_PlayerSingleAdjustData(appId string, page int32, pagecount int32) ([]byte, error) {
	//params := make(map[string]string)
	//params["page"] = strconv.Itoa(int(page))
	//params["limit"] = strconv.Itoa(int(pagecount))
	//return getRequest(appId, "/game_ctrl_alone_config", params, "http", DEFAULT_TIMEOUT)
	return nil, nil
}

// 平台杀率配置
func API_GetPlatformProfitControlConfigData(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/profitcontrol_config_list", params, "http", time.Duration(time.Second*120))
	return nil, nil
}

// 比赛配置
func API_GetMatchConfigData(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/match_list", params, "http", time.Duration(time.Second*120))
	return nil, nil
}

// 比赛报名券活动配置
func API_GetActTicketConfigData(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/activity_ticket_config", params, "http", time.Duration(time.Second*120))
	return nil, nil
}

// 比赛积分商城配置
func API_GetGradeShopConfigData(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/query_exchange_shop", params, "http", time.Duration(time.Second*120))
	return nil, nil
}

// 用户分层配置
func API_GetLogicLevelConfigData(appId string) ([]byte, error) {
	//params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	//return getRequest(appId, "/logic_level_config", params, "http", time.Duration(time.Second*120))
	return nil, nil
}
