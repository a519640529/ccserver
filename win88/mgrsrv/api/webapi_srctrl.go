package api

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	msg_proto "games.yol.com/win88/protocol/message"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/protocol/webapi"
	srvlibprotocol "github.com/idealeak/goserver.v3/srvlib/protocol"
	"github.com/idealeak/goserver/core/admin"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
	"github.com/idealeak/goserver/core/utils"
	"github.com/idealeak/goserver/srvlib"
	"github.com/idealeak/goserver/srvlib/protocol"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

//-------------------------------切换状态-------------------------------------------------------

func ServerStateSwitch(rw http.ResponseWriter, data []byte) {
	pack := &webapi.SAServerStateSwitch{}
	var gs webapi.ASServerStateSwitch
	err := proto.Unmarshal(data, &gs)
	if err != nil {
		pack.Tag = webapi.TagCode_FAILED
		pack.Msg = "proto.Unmarshal is err:" + err.Error()
		by, _ := proto.Marshal(pack)
		webApiResponse(rw, by)
		return
	}

	if gs.SrvId == 0 || gs.SrvType == 0 {
		pack.Tag = webapi.TagCode_FAILED
		pack.Msg = "SrvCtrlStateSwitch failed SrvId or SrvType is zero."
		by, _ := proto.Marshal(pack)
		webApiResponse(rw, by)
		return
	}
	s := srvlib.ServerSessionMgrSington.GetSession(common.GetSelfAreaId(), int(gs.SrvType), int(gs.SrvId))
	logger.Logger.Trace("srvCtrl_StateSwitch enter")
	if s != nil {
		ctrlPacket := &server.ServerCtrl{
			CtrlCode: proto.Int32(common.SrvCtrlStateSwitchCode),
		}
		proto.SetDefaults(ctrlPacket)
		s.Send(int(server.SSPacketID_PACKET_MS_SRVCTRL), ctrlPacket, true)
		pack.Tag = webapi.TagCode_SUCCESS
		pack.Msg = "Server is changed."
		rep, _ := proto.Marshal(pack)
		webApiResponse(rw, rep)
		return
	}
	pack.Tag = webapi.TagCode_FAILED
	pack.Msg = "[stateswitch] no find " + strconv.Itoa(int(gs.SrvId)) + "-" + strconv.Itoa(int(gs.SrvType)) + " server"
	rep, _ := proto.Marshal(pack)
	webApiResponse(rw, rep)
}
func SrvCtrlClose(rw http.ResponseWriter, data []byte) {
	logger.Logger.Trace("srvCtrl_Close enter")
	msg := &webapi.ASSrvCtrlClose{}
	pack := &webapi.SASrvCtrlClose{
		Tag: webapi.TagCode_SUCCESS,
	}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		pack.Msg = "Unmarshal CtrlClose" + err.Error()
		rep, _ := proto.Marshal(pack)
		webApiResponse(rw, rep)
		return
	}
	ctrlPacket := &server.ServerCtrl{
		CtrlCode: proto.Int32(common.SrvCtrlCloseCode),
	}
	proto.SetDefaults(ctrlPacket)
	var closeType []int
	srvType := int(msg.SrvType)
	if msg.SrvType == 0 {
		closeType = append(closeType, srvlib.GameServerType, srvlib.GateServerType, srvlib.WorldServerType)
	} else if srvType == srvlib.GameServerType || srvType == srvlib.GateServerType || srvType == srvlib.WorldServerType {
		closeType = append(closeType, int(msg.SrvType))
	}
	for _, serverType := range closeType {
		srvlib.ServerSessionMgrSington.Broadcast(int(server.SSPacketID_PACKET_MS_SRVCTRL), ctrlPacket, common.GetSelfAreaId(), serverType)
	}
	pack.Msg = "success"
	rep, _ := proto.Marshal(pack)
	webApiResponse(rw, rep)
}
func srvCtrlExecShell(shell string) bool {
	cmd := exec.Command(shell)
	if cmd == nil {
		logger.Logger.Info("srvCtrlExecShell exec.Command(", shell, ") failed")
		return false
	}
	logger.Logger.Info("srvCtrlExecShell exec.Command(", shell, ")")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Logger.Info("srvCtrlExecShell.cmd.StdoutPipe() error:", err)
		return false
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Logger.Info("srvCtrlExecShell.cmd.StderrPipe() error:", err)
		return false
	}

	err = cmd.Start()
	if err != nil {
		logger.Logger.Info("srvCtrlExecShell.cmd.Start()", err)
		return false
	}

	ber, err := ioutil.ReadAll(stderr)
	if err != nil {
		logger.Logger.Info("srvCtrlExecShell.ioutil.ReadAll(stderr)", err)
		return false
	}
	logger.Logger.Info("stderr=", string(ber[:]))

	bsr, err := ioutil.ReadAll(stdout)
	if err != nil {
		logger.Logger.Info("srvCtrlExecShell.ioutil.ReadAll(stdout)", err)
		return false
	}
	logger.Logger.Info("stdout=", string(bsr[:]))

	err = cmd.Wait()
	if err != nil {
		logger.Logger.Info("srvCtrlExecShell.cmd.Wait()", err)
		return false
	}
	return true
}

func SrvCtrlStartScript(rw http.ResponseWriter, data []byte) {
	logger.Logger.Trace("srvCtrl_Start enter")
	msg := &webapi.ASSrvCtrlStartScript{}
	pack := &webapi.SASrvCtrlStartScript{
		Tag: webapi.TagCode_SUCCESS,
	}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		pack.Tag = webapi.TagCode_FAILED
		pack.Msg = "Unmarshal ASSrvCtrlStartScript is err:" + err.Error()
		r, _ := proto.Marshal(pack)
		webApiResponse(rw, r)
		return
	}
	if srvCtrlExecShell(Config.StartScript) {
		pack.Msg = "Start success"
	} else {
		pack.Tag = webapi.TagCode_FAILED
		pack.Msg = "Start failed"
	}
	rep, _ := proto.Marshal(pack)
	webApiResponse(rw, rep)
}
func SrvApi(rw http.ResponseWriter, req *http.Request) {
	defer utils.DumpStackIfPanic("api.SrvCtrlApi")
	logger.Logger.Info("srvCtrl_StateSwitch receive:", req.URL.Path, req.URL.RawQuery)

	if common.RequestCheck(req, model.GameParamData.WhiteHttpAddr) == false {
		logger.Logger.Info("RemoteAddr [%v] require api.", req.RemoteAddr)
		return
	}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		webApiResponse(rw, nil)
		return
	}
	switch req.URL.Path {
	case "/api/Ctrl/ServerStateSwitch":
		ServerStateSwitch(rw, data)
	case "/api/Ctrl/SrvCtrlClose":
		SrvCtrlClose(rw, data)
	case "/api/Ctrl/SrvCtrlStartScript":
		SrvCtrlStartScript(rw, data)
	case "/api/Ctrl/SrvCtrlNotice":
		SrvCtrlNotice(rw, data)
	}
}

var noticeMap = make(map[timer.TimerHandle]struct{})

func SrvCtrlNotice(rw http.ResponseWriter, data []byte) {
	logger.Logger.Trace("SrvCtrlNotice")
	pack := &webapi.SASrvCtrlNotice{}
	msg := &webapi.ASSrvCtrlNotice{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		pack.Tag = webapi.TagCode_FAILED
		pack.Msg = "Unmarshal ASSrvCtrlNotice is err:" + err.Error()
		r, _ := proto.Marshal(pack)
		webApiResponse(rw, r)
		return
	}
	if msg.OpNotice == 1 {
		logger.Logger.Trace("srvCtrl_StopNotice enter")
		for h, _ := range noticeMap {
			timer.StopTimer(h)
		}
		noticeMap = make(map[timer.TimerHandle]struct{})
		pack.Tag = webapi.TagCode_SUCCESS
		pack.Msg = "stop all notice success"
		r, _ := proto.Marshal(pack)
		webApiResponse(rw, r)
		return
	}
	logger.Logger.Trace("srvCtrl_Notice enter")
	if msg.Notice == "" {
		pack.Tag = webapi.TagCode_FAILED
		pack.Msg = "Notice is nil"
		r, _ := proto.Marshal(pack)
		webApiResponse(rw, r)
		return
	}
	noticePacket := &server.ServerNotice{
		Text: proto.String(msg.GetNotice()),
	}
	proto.SetDefaults(noticePacket)
	sc := &protocol.BCSessionUnion{
		Bccs: &protocol.BCClientSession{},
	}
	broadcast, err := BroadcastMaker.CreateBroadcastPacket(sc, int(msg_proto.MSGPacketID_PACKET_SC_NOTICE), noticePacket)
	if err != nil || broadcast == nil {
		pack.Tag = webapi.TagCode_FAILED
		pack.Msg = "send notice failed(inner error)"
		r, _ := proto.Marshal(pack)
		webApiResponse(rw, r)
		return
	}
	funcNotice := func() {
		srvlib.ServerSessionMgrSington.Broadcast(int(srvlibprotocol.SrvlibPacketID_PACKET_SS_BROADCAST), broadcast, common.GetSelfAreaId(), srvlib.GateServerType)
	}
	funcNotice()
	h, b := timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		funcNotice()
		return true
	}), nil, time.Second*time.Duration(msg.GetInterval()), int(msg.GetTimes()))
	if b {
		noticeMap[h] = struct{}{}
	}
	pack.Tag = webapi.TagCode_SUCCESS
	pack.Msg = "send notice success"
	r, _ := proto.Marshal(pack)
	webApiResponse(rw, r)
}

//--------------------------------------------------------------------------------------
func init() {
	//切换状态
	admin.MyAdminApp.Route("/api/Ctrl/ServerStateSwitch", SrvApi)
	//获取服务器列表
	admin.MyAdminApp.Route("/api/Ctrl/ListServerStates", WorldSrvApi)
	//关服
	admin.MyAdminApp.Route("/api/Ctrl/SrvCtrlClose", SrvApi)
	//执行脚本
	admin.MyAdminApp.Route("/api/Ctrl/SrvCtrlStartScript", SrvApi)
	//发送 停止 公告
	admin.MyAdminApp.Route("/api/Ctrl/SrvCtrlNotice", SrvApi)
	//重置Etcd
	admin.MyAdminApp.Route("/api/Ctrl/ResetEtcdData", WorldSrvApi)
}
