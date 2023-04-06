package base

import (
	"encoding/json"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/srvdata"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type AccountData struct {
	Acc    string
	Create int64
	Time   time.Time
}

var (
	accountChan = make(map[string]bool)
	clientArray = make(map[string]string)
	accPool     = []AccountData{}
)
var accountFileName = "robotaccount.json"

func init() {
	model.InitGameParam()

	srvdata.SrvDataModifyCB = func(fileName string, fullName string) {
		if strings.Contains(fullName, "botai") {
			RefleshBevTree(fileName)
		}
	}
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		model.StartupRPClient(common.CustomConfig.GetString("MgoRpcCliNet"), common.CustomConfig.GetString("MgoRpcCliAddr"), time.Duration(common.CustomConfig.GetInt("MgoRpcCliReconnInterV"))*time.Second)
		initAccountData()
		return nil
	})
}
func initAccountData() {
	dirty := false
	fileData, err := ioutil.ReadFile(accountFileName)
	if err != nil {
		logger.Logger.Error("Read robot account file error:", err)
		//没有账号文件，创建新的一组账号
		for i := 0; i < Config.Count; i++ {
			accId := bson.NewObjectId().Hex()
			accountChan[accId] = true
			accPool = append(accPool, AccountData{
				Acc:    accId,
				Create: time.Now().UnixNano(),
				Time:   time.Now(),
			})
		}
		dirty = true
	} else {
		err := json.Unmarshal(fileData, &accPool)
		if err != nil {
			logger.Logger.Error("Unmarshal robot account data error:", err)
			//账号文件数据反序列化失败，创建新的账号数据
			for i := 0; i < Config.Count; i++ {
				accId := bson.NewObjectId().Hex()
				accountChan[accId] = true
				accPool = append(accPool, AccountData{
					Acc:    accId,
					Create: time.Now().UnixNano(),
					Time:   time.Now(),
				})
			}
			dirty = true
		} else {
			usedData := []AccountData{}
			//核对文件中的账号数量
			if len(accPool) > Config.Count {
				//数量过大，截取一部分账号
				usedData = accPool[:Config.Count]
			} else if len(accPool) < Config.Count {
				//数量过少，添加一部分账号
				newCount := Config.Count - len(accPool)
				for i := 0; i < newCount; i++ {
					accPool = append(accPool, AccountData{
						Acc:    bson.NewObjectId().Hex(),
						Create: time.Now().UnixNano(),
						Time:   time.Now(),
					})
				}
				usedData = append(usedData, accPool...)
				dirty = true
			} else {
				usedData = accPool
			}
			//使用文件中的账号
			for _, value := range usedData {
				accountChan[value.Acc] = true
			}
		}
	}
	if dirty {
		//持久化本次的账号数据
		buff, err := json.Marshal(accPool)
		if err != nil {
			logger.Logger.Error("Marshal account data error:", err)
		} else {
			err := ioutil.WriteFile(accountFileName, buff, os.ModePerm)
			if err != nil {
				logger.Logger.Error("Write robot account file error:", err)
			}
		}
	}
}
