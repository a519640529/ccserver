package base

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	server_proto "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/timer"
	"io"
	"math/rand"
	"strconv"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	login_proto "games.yol.com/win88/protocol/login"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"io/ioutil"
	"os"
)

var ClientMgrSington = &ClientMgr{
	sessionPool:    make(map[string]*netlib.Session),
	Running:        true,
	CycleTimeEvent: [24][]HourEvent{},
}

type HourEvent struct {
	newAcc  string //wait to open
	oldAcc  string //wait to close
	session *netlib.Session
}
type InvalidAcc struct {
	Acc     string
	Session *netlib.Session
}
type ClientMgr struct {
	sessionPool    map[string]*netlib.Session
	Running        bool
	CycleTimeEvent [24][]HourEvent
}
type Client struct {
	SeccionId int
	Gameuser  string
	Seccion   *netlib.Session
}

func (this *ClientMgr) RegisteSession(acc string, s *netlib.Session) {
	this.sessionPool[acc] = s
	StartSessionLoginTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if !s.IsConned() || !this.Running {
			StopSessionLoginTimer(s)
			return false
		}
		this.startLogin(acc, s)
		return true
	}), nil, time.Second*time.Duration(5+rand.Int31n(10)), -1)
}

func (this *ClientMgr) UnRegisteSession(acc string) {
	delete(this.sessionPool, acc)
}

func (this *ClientMgr) startLogin(acc string, s *netlib.Session) {
	ts := time.Now().UnixNano()
	csLogin := &login_proto.CSLogin{
		Username:  proto.String(acc),
		TimeStamp: proto.Int64(ts),
		Platform:  proto.String(common.Platform_Rob),
		Channel:   proto.String(common.Channel_Rob),
	}
	params := &model.PlayerParams{
		Ip:           fmt.Sprintf("%v.%v.%v.%v", 1+rand.Int31n(255), 1+rand.Int31n(255), 1+rand.Int31n(255), 1+rand.Int31n(255)),
		City:         RandZone(),
		Platform:     2,
		Logininmodel: "app",
	}
	data, err := json.Marshal(params)
	if err == nil {
		csLogin.Params = proto.String(string(data[:]))
	}
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%v%v", acc, Config.AppId))
	pwd := hex.EncodeToString(h.Sum(nil))
	h.Reset()
	io.WriteString(h, fmt.Sprintf("%v%v%v", pwd, Config.AppId, ts))
	pwd = hex.EncodeToString(h.Sum(nil))
	csLogin.Password = pwd
	csLogin.LoginType = proto.Int32(0)
	csLogin.Sign = proto.String(common.MakeMd5String(csLogin.GetUsername(), csLogin.GetPassword(),
		strconv.Itoa(int(csLogin.GetTimeStamp())), csLogin.GetParams(), Config.AppId))
	proto.SetDefaults(csLogin)
	s.Send(int(login_proto.LoginPacketID_PACKET_CS_LOGIN), csLogin)
	logger.Logger.Tracef("Client [%v] cslogin.", acc)
}
func (this *ClientMgr) AccountValideCheck() {
	invalidCount := 0                                                           //过期账号数量
	updateLimit := len(accPool) * model.GameParamData.InvalidRobotAccRate / 100 //可更新的账号数量
	invalidAccs := []InvalidAcc{}
	if updateLimit > 0 {
		invalideTime := time.Now().AddDate(0, 0, -model.GameParamData.InvalidRobotDay).UnixNano()
		for _, value := range accPool {
			//检查过期账号
			if value.Create < invalideTime {
				//过期账号
				invalidAccs = append(invalidAccs, InvalidAcc{
					Acc:     value.Acc,
					Session: this.sessionPool[value.Acc],
				})
				invalidCount++
				j := rand.Intn(invalidCount)
				i := invalidCount - 1
				if i != j {
					invalidAccs[i], invalidAccs[j] = invalidAccs[j], invalidAccs[i]
				}
			}
		}
		if len(invalidAccs) >= updateLimit {
			invalidAccs = invalidAccs[:updateLimit]
		}
	}
	//本次需要生成的新账号
	cnt := len(invalidAccs)
	for i := 0; i < cnt; i++ {
		timePoint := i % 24
		eventArr := this.CycleTimeEvent[timePoint]
		eventArr = append(eventArr, HourEvent{
			newAcc:  bson.NewObjectId().Hex(),
			oldAcc:  invalidAccs[i].Acc,
			session: invalidAccs[i].Session,
		})
		this.CycleTimeEvent[timePoint] = eventArr
	}
}
func (this *ClientMgr) isWaitCloseSession(s *netlib.Session) bool {
	for _, arr := range this.CycleTimeEvent {
		for _, value := range arr {
			if value.session != nil && value.session.Id == s.Id {
				return true
			}
		}
	}
	return false
}
func (this *ClientMgr) Update() {
	fileModify := false
	eventArr := this.CycleTimeEvent[time.Now().Hour()]
	for _, event := range eventArr {
		accountChan[event.newAcc] = true //使用新的账号
		cfg := NewSessionConfiguration()
		netlib.Connect(cfg) //创建新的连接
		if session, ok := this.sessionPool[event.oldAcc]; ok && session != nil {
			//删除旧有账号数据
			pack := &server_proto.RWAccountInvalid{
				Acc: proto.String(event.oldAcc),
			}
			session.Send(int(login_proto.LoginPacketID_PACKET_CS_ACCOUNTINVALID), pack)
			//删号标记,不要在断线重连了
			session.SetAttribute(SessionAttributeDelAccount, true)
			//关闭连接
			session.Close()
		}
		//更新本地账号数据信息
		for key, value := range accPool {
			if value.Acc == event.oldAcc {
				accPool[key] = AccountData{
					Acc:    event.newAcc,
					Create: time.Now().UnixNano(),
					Time:   time.Now(),
				}
				fileModify = true
				break
			}
		}
	}
	this.CycleTimeEvent[time.Now().Hour()] = []HourEvent{}
	if fileModify == true {
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
