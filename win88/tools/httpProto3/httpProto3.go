package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/qpapi"
	"games.yol.com/win88/protocol/webapi"
	"github.com/idealeak/goserver/core/admin"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

//const TestIp = "http://192.168.10.240:9899"

const TestIp = "http://127.0.0.1:9899"

var ApiKey = []string{
	"/api/Game/AddCoinByIdAndPT",
	"/api/Game/AddCoinById",
	"/api/Cache/ListRoom",
	"/api/Cache/DestroyRoom",
	"/api/Ctrl/ListServerStates",
	"/api/Ctrl/ServerStateSwitch",
	"/api/Ctrl/SrvCtrlClose",
	"/api/Game/CreateShortMessage",
	"/api/Game/DeleteShortMessage",
	"/api/Game/SinglePlayerAdjust",
	"/api/Message/QueryHorseRaceLampList",
	"/api/Game/QueryGamePoolByGameId",
	"/api/Report/OnlineReportTotal",
	"/api/Message/EditHorseRaceLamp",
	"/api/Message/CreateHorseRaceLamp",
	"/api/Message/GetHorseRaceLampById",
	"/api/Report/QueryOnlineReportList",
	"/api/Member/QPAPIRegisterOrLogin",
	"/api/Member/QPGetMemberGoldById",
	"/api/Member/QPGetGameHistory",
	"/api/Game/QPAPIAddSubCoinById",
}

func (clm *APIHttpData) Init() {
	clm.RouteMap["/api/Game/AddCoinByIdAndPT"] = APIData{
		M: &webapi.ASAddCoinByIdAndPT{
			Platform: "1",
			ID:       24953700,
			Gold:     10000,
			Oper:     "test",
			Desc:     "",
			BillNo:   int64(rand.Int31n(100000)),
			LogType:  1,
		},
		Pack: &webapi.SAAddCoinByIdAndPT{},
	}

	clm.RouteMap["/api/Message/GetHorseRaceLampById"] = APIData{
		M: &webapi.ASGetHorseRaceLampById{
			Platform: "1",
			NoticeId: "614d2d43e1382378b76be9a5",
		},
		Pack: &webapi.SAGetHorseRaceLampById{},
	}
	clm.RouteMap["/api/Message/CreateHorseRaceLamp"] = APIData{
		M: &webapi.ASCreateHorseRaceLamp{
			Platform:  "1",
			Title:     "",
			Content:   "dsfsdfff",
			Footer:    "",
			Count:     1,
			State:     1,
			StartTime: 1,
			Priority:  1,
			MsgType:   1,
			StandSec:  1,
			Target:    []int32{},
		},
		Pack: &webapi.SACreateHorseRaceLamp{},
	}
	clm.RouteMap["/api/Message/EditHorseRaceLamp"] = APIData{
		M: &webapi.ASEditHorseRaceLamp{
			HorseRaceLamp: &webapi.HorseRaceLamp{
				Id:        "614d2f500d96096240c27c70",
				Platform:  "1",
				Title:     "",
				Content:   "dsfsdfff1111",
				Footer:    "",
				Count:     1,
				State:     1,
				StartTime: 1,
				Priority:  1,
				MsgType:   1,
				StandSec:  1,
				Target:    []int32{},
			},
		},
		Pack: &webapi.SAEditHorseRaceLamp{},
	}
	clm.RouteMap["/api/Game/QueryGamePoolByGameId"] = APIData{
		M: &webapi.ASQueryGamePoolByGameId{
			GameId:   306,
			GameMode: 0,
			Platform: "1",
			GroupId:  0,
		},
		Pack: &webapi.SAQueryGamePoolByGameId{},
	}
	clm.RouteMap["/api/Game/SinglePlayerAdjust"] = APIData{
		M: &webapi.ASSinglePlayerAdjust{
			Opration: 3,
			PlayerSingleAdjust: &webapi.PlayerSingleAdjust{
				Platform:   "1",
				SnId:       35153500,
				GameFreeId: 6040001,
				GameId:     604,
				TotalTime:  10,
			},
		},
		Pack: &webapi.SASinglePlayerAdjust{},
	}
	clm.RouteMap["/api/Ctrl/SrvCtrlClose"] = APIData{
		M: &webapi.ASSrvCtrlClose{
			SrvType: 0,
		},
		Pack: &webapi.SASrvCtrlClose{},
	}
	clm.RouteMap["/api/Ctrl/ServerStateSwitch"] = APIData{
		M: &webapi.ASServerStateSwitch{
			SrvId:   777,
			SrvType: 7,
		},
		Pack: &webapi.SAServerStateSwitch{},
	}
	clm.RouteMap["/api/Ctrl/ListServerStates"] = APIData{
		M:    &webapi.ASListServerStates{},
		Pack: &webapi.SAListServerStates{},
	}
	clm.RouteMap["/api/Game/AddCoinById"] = APIData{
		M: &webapi.ASAddCoinById{
			Platform: "",
			ID:       13931000,
			Gold:     1000000,
			Oper:     "test",
			Desc:     "",
			BillNo:   int64(rand.Int31n(10000)),
			LogType:  0,
		},
		Pack: &webapi.SAAddCoinById{},
	}
	clm.RouteMap["/api/Cache/ListRoom"] = APIData{
		M: &webapi.ASListRoom{
			PageNo:   1,
			PageSize: 50,
			RoomType: -1,
		},
		Pack: &webapi.SAListRoom{},
	}
	clm.RouteMap["/api/Cache/DestroyRoom"] = APIData{
		M: &webapi.ASDestroyRoom{
			Platform:    "",
			SceneIds:    []int32{},
			DestroyType: 0,
		},
		Pack: &webapi.SADestroyRoom{},
	}
	clm.RouteMap["/api/Game/CreateShortMessage"] = APIData{
		M: &webapi.ASCreateShortMessage{
			Platform:      "1",
			SrcSnid:       35153500,
			DestSnid:      35153500,
			NoticeTitle:   "221",
			NoticeContent: "qwqewqe",
		},
		Pack: &webapi.SACreateShortMessage{},
	}
	clm.RouteMap["/api/Game/DeleteShortMessage"] = APIData{
		M: &webapi.ASDeleteShortMessage{
			Platform: "1",
			Id:       "6143f8b30d9609caec249b1a",
		},
		Pack: &webapi.SADeleteShortMessage{},
	}
	clm.RouteMap["/api/Message/QueryHorseRaceLampList"] = APIData{
		M: &webapi.ASQueryHorseRaceLampList{
			Platform: "1",
			PageNo:   1,
			PageSize: 20,
			MsgType:  0,
		},
		Pack: &webapi.SAQueryHorseRaceLampList{},
	}
	clm.RouteMap["/api/Report/OnlineReportTotal"] = APIData{
		M: &webapi.ASOnlineReportTotal{
			Platform: "1",
		},
		Pack: &webapi.SAOnlineReportTotal{},
	}
	clm.RouteMap["/api/Report/QueryOnlineReportList"] = APIData{
		M: &webapi.ASQueryOnlineReportList{
			Platform: "1",
			PageNo:   1,
			PageSize: 20,
		},
		Pack: &webapi.SAQueryOnlineReportList{},
	}

	ml := &qpapi.ASLogin{
		MerchantTag: "1",
		UserName:    "abcd",
		Ts:          time.Now().Unix(),
		Sign:        "",
	}

	rawl := fmt.Sprintf("%v%v%v%v", ml.GetMerchantTag(), ml.GetUserName(), "", ml.GetTs())
	hl := md5.New()
	io.WriteString(hl, rawl)
	newsignl := hex.EncodeToString(hl.Sum(nil))
	ml.Sign = newsignl

	//创建用户
	clm.RouteMap["/api/Member/QPAPIRegisterOrLogin"] = APIData{
		M:    ml,
		Pack: &qpapi.SALogin{},
	}
	//clm.RouteMap["/api/Game/CrashVerifier"] = APIData{
	//	M: &qpapi.ASCrachHash{
	//		Hash:"583f8e896a5a333e2eb532c31adeffda430d7121e1d4c44914972b5070f7881c",
	//		Wheel:33,
	//	},
	//	Pack: &qpapi.SACrachHash{},
	//}

	mu := &qpapi.ASMemberGold{
		Username:    "abcd",
		MerchantTag: "1",
		Ts:          time.Now().Unix(),
		Sign:        "",
	}

	rawu := fmt.Sprintf("%v%v%v%v", mu.GetUsername(), mu.GetMerchantTag(), "", mu.GetTs())
	hu := md5.New()
	io.WriteString(hu, rawu)
	newsignu := hex.EncodeToString(hu.Sum(nil))
	mu.Sign = newsignu

	//获取用户金币
	clm.RouteMap["/api/Member/QPGetMemberGoldById"] = APIData{
		M:    mu,
		Pack: &qpapi.SAMemberGold{},
	}

	ma := &qpapi.ASAddCoinById{
		Username:    "abcd",
		Gold:        10000,
		BillNo:      time.Now().Unix(),
		MerchantTag: "1",
		Ts:          time.Now().Unix(),
		Sign:        "",
	}

	rawa := fmt.Sprintf("%v%v%v%v%v%v", ma.GetUsername(), ma.GetGold(), ma.GetBillNo(), ma.GetMerchantTag(), "", ma.GetTs())
	ha := md5.New()
	io.WriteString(ha, rawa)
	newsigna := hex.EncodeToString(ha.Sum(nil))
	ma.Sign = newsigna

	//加减币
	clm.RouteMap["/api/Game/QPAPIAddSubCoinById"] = APIData{
		M:    ma,
		Pack: &qpapi.SAAddCoinById{},
	}

	m := &qpapi.ASPlayerHistory{
		//Username:"111111",
		MerchantTag:      "1",
		GameHistoryModel: 1,
		StartTime:        1654587626,
		EndTime:          1655809512,
		PageNo:           2,
		PageSize:         50,
		Ts:               time.Now().Unix(),
		Sign:             "",
	}

	raw := fmt.Sprintf("%v%v%v%v%v%v%v%v%v", m.GetUsername(), m.GetMerchantTag(), m.GetGameHistoryModel(),
		m.GetStartTime(), m.GetEndTime(), m.GetPageNo(), m.GetPageSize(), "", m.GetTs())
	h := md5.New()
	io.WriteString(h, raw)
	newsign := hex.EncodeToString(h.Sum(nil))
	m.Sign = newsign
	clm.RouteMap["/api/Member/QPGetGameHistory"] = APIData{
		M:    m,
		Pack: &qpapi.SAPlayerHistory{},
	}
}

func init() {
	for _, key := range ApiKey {
		admin.MyAdminApp.Route(key, WorldSrvApi)
	}
	module.RegisteModule(APIHttpSington, time.Second, 1)
}

var APIHttpSington = &APIHttpData{
	RouteMap: make(map[string]APIData),
}

type APIData struct {
	M    proto.Message
	Pack proto.Message
}
type APIHttpData struct {
	RouteMap map[string]APIData
}

func (clm *APIHttpData) ModuleName() string {
	return "APIHttp"
}
func (clm *APIHttpData) Update() {
}

func (clm *APIHttpData) Shutdown() {
	module.UnregisteModule(clm)
}
func WorldSrvApi(rw http.ResponseWriter, req *http.Request) {
	Path := req.URL.Path
	Method := req.Method
	fmt.Println(Path, Method)
	if Method == "POST" {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logger.Logger.Info("Body err.", err)
			WebApiResponseByte(rw, nil)
			return
		}
		fmt.Println(data)
	} else {
		fmt.Println("path......", Path)
		if apiData, ok := APIHttpSington.RouteMap[Path]; ok {
			WebApiResponseByte(rw, SendData(Path, apiData.M, apiData.Pack))
		} else {
			WebApiResponseByte(rw, nil)
		}
	}
	return
}
func WebApiResponseByte(rw http.ResponseWriter, data []byte) bool {
	dataLen := len(data)
	rw.Header().Set("Content-Length", fmt.Sprintf("%v", dataLen))
	rw.WriteHeader(http.StatusOK)
	pos := 0
	for pos < dataLen {
		writeLen, err := rw.Write(data[pos:])
		if err != nil {
			logger.Logger.Info("webApiResponse SendData error:", err, " data=", string(data[:]), " pos=", pos, " writelen=", writeLen, " dataLen=", dataLen)
			return false
		}
		pos += writeLen
	}
	return true
}

func SendData(url string, m, pack proto.Message) []byte {
	a, err := proto.Marshal(m)
	startTime := time.Now().UnixNano()
	args := fmt.Sprintf("%v;%v;%v;%v", common.Config.AppId, url, string(a), startTime)
	h := md5.New()
	io.WriteString(h, args)
	realSign := hex.EncodeToString(h.Sum(nil))
	url = fmt.Sprintf("%v?nano=%v&sign=%v", TestIp+url, startTime, realSign)

	new_str := bytes.NewBuffer(a)
	req, err := http.NewRequest("POST", url, new_str)
	// req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("status", resp.Status)
	//fmt.Println("response:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(body))
	proto.Unmarshal(body, pack)
	fmt.Println("=============", pack)
	type Api struct {
		Info string
	}
	info, _ := json.Marshal(pack)
	api := Api{
		Info: string(info),
	}
	b, _ := json.Marshal(api)
	return b
}
