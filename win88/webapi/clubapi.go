package webapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/idealeak/goserver/core/logger"
)

// 俱乐部创建审核，阻塞协程
// 如果调用失败只会有日志记录，后台也没有这个俱乐部的信息。
// 出现该问题后，后台可以指定（也就是手动输入）相关信息调用审核结果接口即可。
// 下面的俱乐部公告审核和该接口一样
// 注意注意：如果给后台发送成功的话，err==nil那么请把这个俱乐部的v.CreateCheckPosted标记为true
func API_ClubCreateWaitCheck(appId string, ClubID, ClubOwner int32, PltID, ClubName, ClubNotice string) error {
	params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	params["platform"] = PltID
	params["snid"] = strconv.Itoa(int(ClubOwner))
	params["club_name"] = ClubName
	params["club_id"] = strconv.Itoa(int(ClubID))
	params["club_notice"] = ClubNotice
	buff, err := getRequest(appId, "/push_club_create_review", params, nil, "http", DEFAULT_TIMEOUT)
	//fmt.Println(string(buff))
	//fmt.Println(err)
	if err != nil {
		//后台没有俱乐部信息的话，游服人员可以在这里查日志！！！
		logger.Logger.Errorf("ClubID=v% ClubOwner=v% at API_ClubCreateWaitCheck() req failed, err:v%", ClubID, ClubOwner, err)
		return err
	}
	fmt.Println(string(buff))
	type ApiResult struct {
		Tag int32
		Msg string
	}
	result := ApiResult{}
	err = json.Unmarshal(buff, &result)
	if err != nil {
		return err
	}
	if result.Tag != 0 {
		errMsg := fmt.Sprintf("Create ClubCreateWaitCheck result failed._%v", result.Msg)
		return errors.New(errMsg)
	} else {
		return nil
	}
}

// 俱乐部公告审核，阻塞协程
// 注意注意：如果给后台发送成功的话，err==nil那么请把这个俱乐部的v.NoticeCheckPosted标记为true
// OpSnid为操作者的snid,注意：操作者不一定是俱乐部创建者
func API_ClubNoticeWaitCheck(appId string, ClubID, OpSnId int32, PltID, ClubName, ClubNotice string) error {
	params := make(map[string]string)
	//params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	params["platform"] = PltID
	params["snid"] = strconv.Itoa(int(OpSnId))
	params["club_name"] = ClubName
	params["club_id"] = strconv.Itoa(int(ClubID))
	params["club_notice"] = ClubNotice
	buff, err := getRequest(appId, "/push_club_notice_review", params, nil, "http", DEFAULT_TIMEOUT)
	//fmt.Println(string(buff))
	if err != nil {
		//后台没有俱乐部信息的话，游服人员可以在这里查日志！！！
		logger.Logger.Errorf("ClubID=v% ClubOwner=v% at API_ClubNoticeWaitCheck() req failed, err:v%", ClubID, OpSnId, err)
		return err
	}
	type ApiResult struct {
		Tag int32
		Msg string
	}
	result := ApiResult{}
	err = json.Unmarshal(buff, &result)
	if err != nil {
		return err
	}
	if result.Tag != 0 {
		errMsg := fmt.Sprintf("Create ClubNoticeWaitCheck result failed._%v", result.Msg)
		return errors.New(errMsg)
	} else {
		return nil
	}
}

// 请求后台俱乐部的流水返给客户端用，阻塞协程
func ReqClubTurnover(appId string, clubID int32, DateTs int64) ([]byte, error) {
	params := make(map[string]string)
	params["club_id"] = strconv.Itoa(int(clubID))
	params["date_ts"] = strconv.Itoa(int(DateTs))
	buff, err := getRequest(appId, "/club_room_statistics", params, nil, "http", DEFAULT_TIMEOUT)
	//fmt.Println(string(buff))
	if err != nil {
		logger.Logger.Errorf("ReqClubTurnover failed err:", err)
		return nil, err
	}
	return buff, nil
}

// 请求后台俱乐部的抽水，阻塞协程
func ReqClubPump(appId string, pltID string, DateTs int64) ([]byte, error) {
	params := make(map[string]string)
	params["platform"] = pltID
	params["date_ts"] = strconv.Itoa(int(DateTs))
	buff, err := getRequest(appId, "/platform_club_statistics", params, nil, "http", DEFAULT_TIMEOUT)
	//fmt.Println(string(buff))
	if err != nil {
		logger.Logger.Errorf("ReqClubPump failed err:", err)
		return nil, err
	}
	return buff, nil
}

// 请求后台俱乐部包间详细的抽水，阻塞协程
func ReqClubRoomPumpDetail(appId string, clubID int32, clubRoomID string, PageSize, PageNum int32, DateTs int64) ([]byte, error) {
	params := make(map[string]string)
	params["club_id"] = strconv.Itoa(int(clubID))
	params["room_id"] = clubRoomID
	params["date_ts"] = strconv.Itoa(int(DateTs))
	params["PageSize"] = strconv.Itoa(int(PageSize))
	params["PageNum"] = strconv.Itoa(int(PageNum))
	buff, err := getRequest(appId, "/clubroom_statistics", params, nil, "http", DEFAULT_TIMEOUT)
	//fmt.Println(string(buff))
	if err != nil {
		logger.Logger.Errorf("ReqClubRoomPumpDetail failed err:", err)
		return nil, err
	}
	return buff, nil
}
