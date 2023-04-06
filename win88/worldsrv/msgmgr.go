package main

import (
	"games.yol.com/win88/model"
)

var MsgMgrSington = &MsgMgr{
	subscribeMsgs: make(map[string]*model.Message),
}

type MsgMgr struct {
	subscribeMsgs map[string]*model.Message
}

func (mm *MsgMgr) InitMsg() {
	for _, p := range PlatformMgrSington.Platforms {
		msgs, err := model.GetSubscribeMessage(p.IdStr)
		if err == nil {
			mm.init(msgs)
		}
	}
}

func (mm *MsgMgr) init(msgs []model.Message) {
	for i := 0; i < len(msgs); i++ {
		msg := msgs[i]
		if msg.State != model.MSGSTATE_REMOVEED {
			mm.subscribeMsgs[msg.Id.Hex()] = &msg
		}
	}
}

func (mm *MsgMgr) AddMsg(msg *model.Message) {
	if msg != nil {
		mm.subscribeMsgs[msg.Id.Hex()] = msg
	}
}
func (mm *MsgMgr) RemoveMsg(msg *model.Message) {
	if msg != nil {
		delete(mm.subscribeMsgs, msg.Id.Hex())
	}
}
func (mm *MsgMgr) GetSubscribeMsgs(platform string, ts int64) (msgs []*model.Message) {
	for _, msg := range mm.subscribeMsgs {
		if msg.CreatTs > ts {
			if platform == "" || msg.Platform == platform {
				msgs = append(msgs, msg)
			}
		}
	}
	return
}

func init() {
	//core.RegisteHook(core.HOOK_BEFORE_START, func() error {
	//	msgs, err := model.GetSubscribeMessage()
	//	if err == nil {
	//		MsgMgrSington.init(msgs)
	//	}
	//	return nil
	//})

	////使用并行加载
	//RegisteParallelLoadFunc("平台邮件", func() error {
	//	//logger.Logger.Trace("平台邮件")
	//	msgs, err := model.GetSubscribeMessage()
	//	if err == nil {
	//		MsgMgrSington.init(msgs)
	//	}
	//	return err
	//})
}
