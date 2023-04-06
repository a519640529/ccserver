package base

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"time"

	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

const (
	ReplayServerType int = 8
	ReplayServerId       = 801
)

var _replayIgnorePacketIds = map[int]bool{}

type ReplayRecorder struct {
	rs      *server.ReplaySequene
	id      string
	Logid   string
	has     bool
	start   bool
	needRec bool
}

func RegisteReplayIgnorePacketId(ids ...int) {
	for _, id := range ids {
		_replayIgnorePacketIds[id] = true
	}
}

func NewReplayRecorder(id string) *ReplayRecorder {
	rr := &ReplayRecorder{
		rs: &server.ReplaySequene{},
		id: id,
	}
	return rr
}

func (this *ReplayRecorder) Reset(needRec bool) {
	this.needRec = needRec
	this.rs.Sequenes = nil
}

func (this *ReplayRecorder) Init(packid int, pack interface{}) {
	this.Record(-1, -1, packid, pack)
	this.start = true
}

func (this *ReplayRecorder) Record(pos, excludePos int, packid int, pack interface{}, force ...bool) {
	//纯机器人对战可能不需要记录录像
	if !this.needRec {
		return
	}
	if this.start {
		//过滤掉中间玩家掉线重新上线的消息
		if _, exist := _replayIgnorePacketIds[packid]; exist {
			if len(force) == 0 || force[0] == false {
				return
			}
		}
	}
	var data []byte
	var err error
	if model.GameParamData.ReplayDataUseJson {
		data, err = json.Marshal(pack)
	} else {
		data, err = netlib.Gpb.Marshal(pack)
	}
	if err == nil {
		rr := &server.ReplayRecord{
			TimeStamp:  proto.Int64(time.Now().Unix()),
			Pos:        proto.Int(pos),
			ExcludePos: proto.Int(excludePos),
			PacketId:   proto.Int(packid),
		}
		this.rs.Sequenes = append(this.rs.Sequenes, rr)
		if model.GameParamData.ReplayDataUseJson {
			rr.StrData = proto.String(string(data[:]))
		} else {
			rr.BinData = data
		}
		this.has = true
	}
}

func (this *ReplayRecorder) Fini(s *Scene) {
	if this.has && len(this.rs.Sequenes) > 10 {
		pack := &server.GRReplaySequene{
			Name: proto.String(this.id),
			// todo dev
			//Rec:        this.rs,
			LogId:      proto.String(this.Logid),
			GameId:     proto.Int32(s.DbGameFree.GetGameId()),
			RoomMode:   proto.Int32(s.DbGameFree.GetGameMode()),
			NumOfGames: proto.Int(s.NumOfGames),
			Platform:   proto.String(s.Platform),
			DatasVer:   proto.Int32(s.rrVer),
			GameFreeid: proto.Int32(s.GetGameFreeId()),
			RoomId:     proto.Int(s.SceneId),
		}
		if s.ClubId != 0 {
			pack.ClubId = proto.Int32(s.ClubId)
			pack.ClubRoom = proto.String(s.RoomId)
		}
		for _, player := range s.Players {
			if player != nil {
				pack.Channel = proto.String(player.Channel)
				pack.Promoter = proto.String(player.BeUnderAgentCode)
				break
			}
		}
		LogChannelSington.WriteMQData(&model.RabbitMQData{MQName: "log_playback", Data: pack})
	}
}

func (this *ReplayRecorder) GetId() string {
	return this.id
}

//////////////////////////////////////////////////
//内存落地
type ReplayRecorderPO struct {
	RS    *server.ReplaySequene
	Id    string
	LogId string
	Has   bool
	Start bool
}

//序列化
func (this *ReplayRecorder) Marshal() ([]byte, error) {
	po := ReplayRecorderPO{
		RS:    this.rs,
		Id:    this.id,
		LogId: this.Logid,
		Has:   this.has,
		Start: this.start,
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&po)
	if err != nil {
		logger.Logger.Warnf("(this *ReplayRecorder) Marshal() %v gob.Encode err:%v", this.id, err)
		return nil, err
	}
	return buf.Bytes(), nil
}

//反序列化
func (this *ReplayRecorder) Unmarshal(data []byte) error {
	po := &ReplayRecorderPO{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(po)
	if err != nil {
		logger.Logger.Warnf("(this *ReplayRecorder) Unmarshal gob.Decode err:%v", err)
		return err
	}
	this.rs = po.RS
	this.id = po.Id
	this.Logid = po.LogId
	this.has = po.Has
	this.start = po.Start
	return nil
}

//////////////////////////////////////////////////
