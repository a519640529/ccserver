package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/idealeak/goserver/core/admin"
	"github.com/idealeak/goserver/core/logger"
)

const (
	ResponseTag_Ok                       int = iota //0
	ResponseTag_ParamError                          //1
	ResponseTag_NoFindService                       //2
	ResponseTag_NoFindUser                          //3
	ResponseTag_DataMarshalError                    //4
	ResponseTag_TransactYield                       //5
	ResponseTag_Unsupport                           //6
	ResponseTag_NoFindRoom                          //7
	ResponseTag_NoData                              //8
	ResponseTag_NoFindClub                          //9
	ResponseTag_ClubHadCreated                      //10
	ResponseTag_SrcCardCntNotEnough                 //11
	ResponseTag_SrcPlayerNotExist                   //12
	ResponseTag_DestPlayerNotExist                  //13
	ResponseTag_TransferCardFailed                  //14
	ResponseTag_OpFailed                            //15
	ResponseTag_InviterNotExist                     //16
	ResponseTag_HadSetInviter                       //17
	ResponseTag_CreateNewPlayerFailed               //18
	ResponseTag_ClubAdminCountReachLimit            //19
	ResponseTag_OnlyBeClubMember                    //20
	ResponseTag_FetchNiceIdFail                     //21
	ResponseTag_MoneyNotEnough                      //22
	ResponseTag_TransferMoneyFailed                 //23
	ResponseTag_SrcMoneyCntNotEnough                //24
	ResponseTag_CoinNotEnough                       //25
	ResponseTag_CoinInUse                           //26
	ResponseTag_RMBNotEnough                        //27
	ResponseTag_AgentNotExist                       //28
)

type HttpResult struct {
	Tag int
	Msg interface{}
}

func ResponseMsg(req *http.Request, res http.ResponseWriter, tag int, msg interface{}) bool {
	result := &HttpResult{
		Tag: tag,
		Msg: msg,
	}

	data, err := json.Marshal(&result)
	if err != nil {
		logger.Logger.Info(req.RemoteAddr, " Marshal error:", err)
		return false
	}

	fmt.Println(string(data[:]))
	dataLen := len(data)
	res.Header().Set("Content-Length", fmt.Sprintf("%v", dataLen))
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.WriteHeader(http.StatusOK)
	pos := 0
	for pos < dataLen {
		writeLen, err := res.Write(data[pos:])
		if err != nil {
			logger.Logger.Info(req.RemoteAddr, " SendData error:", err)
			return false
		}
		pos += writeLen
	}

	return true
}

func responseMsg2(req *http.Request, res http.ResponseWriter, params map[string]interface{}) bool {
	data, err := json.Marshal(params)
	if err != nil {
		logger.Logger.Info(req.RemoteAddr, " Marshal error:", err)
		return false
	}

	fmt.Println(string(data[:]))
	dataLen := len(data)
	res.Header().Set("Content-Length", fmt.Sprintf("%v", dataLen))
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.WriteHeader(http.StatusOK)
	pos := 0
	for pos < dataLen {
		writeLen, err := res.Write(data[pos:])
		if err != nil {
			logger.Logger.Info(req.RemoteAddr, " SendData error:", err)
			return false
		}
		pos += writeLen
	}

	return true
}

func RequestCheck(req *http.Request, whitelist []string) bool {
	strs := strings.Split(req.RemoteAddr, ":")
	if len(strs) != 2 {
		return false
	}

	if len(admin.Config.WhiteHttpAddr) > 0 {
		for _, value := range admin.Config.WhiteHttpAddr {
			if value == strs[0] {
				return true
			}
		}
	}

	if len(whitelist) > 0 {
		for _, value := range whitelist {
			if value == strs[0] {
				return true
			}
		}
	}
	//都没设置也让过
	return len(whitelist) == 0 && len(admin.Config.WhiteHttpAddr) == 0
}
