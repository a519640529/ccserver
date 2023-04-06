package model

import (
	"errors"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	MSGSTATE_UNREAD int32 = iota
	MSGSTATE_READED
	MSGSTATE_REMOVEED
)

const (
	MSGATTACHSTATE_DEFAULT int32 = iota
	MSGATTACHSTATE_GOT
)

const (
	MSGTYPE_MAIL int32 = iota
	MSGTYPE_GONGGAO
	MSGTYPE_INVITECODE
	MSGTYPE_GIFT
	MSGTYPE_GOLDCOMERANK
	MSGTYPE_YEB                // 余额宝
	MSGTYPE_ClubGet            // 俱乐部会长给会员 会员得到
	MSGTYPE_ClubPump           // 俱乐部会长给会员 会长获得抽水
	MSGTYPE_IOSINSTALLSTABLE   //IOS安装稳定版
	MSGTYPE_REBATE             //奖励流水返利
	MSGTYPE_RANDCOIN           //红包雨
	MSGTYPE_MATCH_SIGNUPFEE    //比赛报名费
	MSGTYPE_MATCH_TICKETREWARD //比赛报名券奖励
	MSGTYPE_MATCH_SHOPEXCHANGE //积分商城兑换
	MSGTYPE_MATCH_SHOPERETURN  //积分商城兑换退还
	MSGTYPE_ITEM               //获取道具
)

const (
	HallAll     int64 = 1 << iota //所有大厅都显示
	HallMain                      //主大厅显示
	HallTienlen                   //len大厅显示
	HallFish                      //fish大厅显示
)

const MSG_MAX_COUNT int = 100 //玩家最多保存100封邮件 ,调整大一点

var (
	MessageDBName   = "user"
	MessageCollName = "user_msg"
)

type Message struct {
	Id          bson.ObjectId `bson:"_id"`
	Pid         string        //原始消息的ID
	MType       int32         //消息类型
	Title       string        //标题
	Content     string        //内容
	Oper        int32         //0.系统 1.玩家
	SrcId       int32         //发送人ID
	SrcName     string        //发送人名字
	SnId        int32         //目标人ID
	Coin        int64         //携带金币数量
	Ticket      int64         //比赛报名券
	Grade       int64         //积分
	Diamond     int64         //钻石
	State       int32         //当前消息的状态
	CreatTs     int64         //创建时间戳
	AttachState int32         //附件状态
	GiftId      string        //
	Params      []int32       //额外参数
	Platform    string        //平台信息
	ShowId      int64         //区分主子游戏大厅
}

func NewMessage(pid string, srcId int32, srcName string, snid, mType int32, title, content string, coin, diamond int64,
	state int32, addTime int64, attachState int32, giftId string, params []int32, platform string, showId int64) *Message {

	msg := &Message{
		Id:          bson.NewObjectId(),
		Pid:         pid,
		MType:       mType,
		Title:       title,
		Content:     content,
		SrcId:       srcId,
		SrcName:     srcName,
		SnId:        snid,
		Coin:        coin,
		State:       state,
		CreatTs:     addTime,
		AttachState: attachState,
		GiftId:      giftId,
		Params:      params,
		Platform:    platform,
		Diamond:     diamond,
		ShowId:      showId,
	}
	if msg.Pid == "" {
		msg.Pid = msg.Id.Hex()
	}
	return msg
}
func NewMessageByPlayer(pid string, oper, srcId int32, srcName string, snid, mType int32, title, content string, coin, diamond int64,
	state int32, addTime int64, attachState int32, giftId string, params []int32, platform string, showId int64) *Message {

	msg := &Message{
		Id:          bson.NewObjectId(),
		Pid:         pid,
		MType:       mType,
		Title:       title,
		Content:     content,
		Oper:        oper,
		SrcId:       srcId,
		SrcName:     srcName,
		SnId:        snid,
		Coin:        coin,
		State:       state,
		CreatTs:     addTime,
		AttachState: attachState,
		GiftId:      giftId,
		Params:      params,
		Platform:    platform,
		Diamond:     diamond,
		ShowId:      showId,
	}
	if msg.Pid == "" {
		msg.Pid = msg.Id.Hex()
	}
	return msg
}

type InsertMsgArg struct {
	Platform string
	Msgs     []Message
}

func InsertMessage(plt string, msgs ...*Message) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	docs := make([]Message, 0, len(msgs))
	for _, msg := range msgs {
		docs = append(docs, *msg)
	}
	if len(docs) == 0 {
		return errors.New("no data")
	}
	args := &InsertMsgArg{Platform: plt, Msgs: docs}
	var ret bool
	err = rpcCli.CallWithTimeout("MessageSvc.InsertMessage", args, &ret, time.Second*30)
	if err != nil {
		return err
	}
	if !ret {
		return errors.New("InsertMessage error")
	}
	return nil
}

type GetMessageArgs struct {
	Plt  string
	SnId int32
}

func GetNotDelMessage(plt string, snid int32) (ret []Message, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetMessageArgs{Plt: plt, SnId: snid}
	err = rpcCli.CallWithTimeout("MessageSvc.GetNotDelMessage", args, &ret, time.Second*30)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func GetMessage(plt string, snid int32) (ret []Message, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetMessageArgs{Plt: plt, SnId: snid}
	err = rpcCli.CallWithTimeout("MessageSvc.GetMessage", args, &ret, time.Second*30)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

//func GetMessageByState(snid, state int32) (ret []Message, err error) {
//	err = MessageCollection().Find(bson.M{"snid": snid, "state": state}).All(&ret)
//	return
//}
//func GetMessageByNotState(snid, state int32) (ret []Message, err error) {
//	err = MessageCollection().Find(bson.M{"snid": snid, "state": bson.M{"$ne": state}}).All(&ret)
//	return
//}

type GetMsgArg struct {
	Platform string
	IdStr    string
}

func GetMessageById(IdStr string, plt string) (msg *Message, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetMsgArg{Platform: plt, IdStr: IdStr}
	err = rpcCli.CallWithTimeout("MessageSvc.GetMessageById", args, &msg, time.Second*30)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func GetSubscribeMessage(plt string) (msg []Message, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	err = rpcCli.CallWithTimeout("MessageSvc.GetSubscribeMessage", plt, &msg, time.Second*30)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

type MsgArgs struct {
	Platform                              string
	StartTs, EndTs                        int64
	ToIndex, DestSnId, MsgType, FromIndex int
}
type RetMsg struct {
	Msg   []Message
	Count int
}

func GetMessageInRangeTsLimitByRange(mas *MsgArgs) (ret *RetMsg, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	err = rpcCli.CallWithTimeout("MessageSvc.GetMessageInRangeTsLimitByRange", mas, &ret, time.Second*30)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type ReadMsgArgs struct {
	Platform string
	Id       bson.ObjectId
}

func ReadMessage(id bson.ObjectId, plt string) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &ReadMsgArgs{
		Platform: plt,
		Id:       id,
	}
	var ret bool
	err = rpcCli.CallWithTimeout("MessageSvc.ReadMessage", args, &ret, time.Second*30)
	return err
}

type DelMsgArgs struct {
	Platform string
	Id       bson.ObjectId
	Del      int32 // 默认0 1为假删
}

type DelAllMsgArgs struct {
	Platform string
	Ids      []bson.ObjectId
	Del      int32 // 默认0 1为假删
}

func DelMessage(args *DelMsgArgs) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	var ret bool
	err = rpcCli.CallWithTimeout("MessageSvc.DelMessage", args, &ret, time.Second*30)
	return err
}

func DelAllMessage(args *DelAllMsgArgs) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	var ret bool
	err = rpcCli.CallWithTimeout("MessageSvc.DelAllMessage", args, &ret, time.Second*30)
	if !ret {
		return err
	}
	return nil
}

type AttachMsgArgs struct {
	Platform string
	Id       bson.ObjectId
	Ts       int64
}

func GetMessageAttach(id bson.ObjectId, plt string) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &AttachMsgArgs{
		Platform: plt,
		Id:       id,
	}
	var ret bool
	err = rpcCli.CallWithTimeout("MessageSvc.GetMessageAttach", args, &ret, time.Second*30)
	return err
}

type AttachMsgsArgs struct {
	Platform string
	Ids      []string
}

func GetMessageAttachs(ids []string, plt string) (*[]string, error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &AttachMsgsArgs{
		Platform: plt,
		Ids:      ids,
	}
	var ret []string
	err := rpcCli.CallWithTimeout("MessageSvc.GetMessageAttachs", args, &ret, time.Second*30)
	return &ret, err
}

func RemoveMessages(platform string, ts int64) (ret *mgo.ChangeInfo, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &AttachMsgArgs{
		Platform: platform,
		Ts:       ts,
	}
	err = rpcCli.CallWithTimeout("MessageSvc.RemoveMessages", args, &ret, time.Second*30)
	return
}
