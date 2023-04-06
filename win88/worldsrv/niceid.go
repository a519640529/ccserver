package main

import (
	"math/rand"

	"games.yol.com/win88/proto"
	player_proto "games.yol.com/win88/protocol/player"
	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
)

var niceIdMgr = &NiceIdManager{
	SnIds:   []int32{},
	UsedIds: make(map[int32]int32),
}

type NiceIdManager struct {
	SnIds   []int32
	UsedIds map[int32]int32
}

func (this *NiceIdManager) init() {
	//this.SnIds = model.GetInvalidSnid()
	this.SnIds = make([]int32, 0, len(srvdata.PBDB_PlayerInfoMgr.Datas.Arr))
	for _, value := range srvdata.PBDB_PlayerInfoMgr.Datas.Arr {
		this.SnIds = append(this.SnIds, value.GetId())
	}
	snidLen := len(this.SnIds)
	logger.Logger.Info("NiceIdManager snid lens:", snidLen)
	for i := 0; i < snidLen; i++ {
		index := rand.Intn(snidLen)
		this.SnIds[i], this.SnIds[index] = this.SnIds[index], this.SnIds[i]
	}
	for _, value := range niceIdMgr.SnIds {
		this.UsedIds[value] = 0
	}
}
func (this *NiceIdManager) PopNiceId(user int32) int32 {
	if len(this.SnIds) <= 0 {
		return 0
	}
	selId := this.SnIds[len(this.SnIds)-1]
	this.SnIds = this.SnIds[:len(this.SnIds)-1]
	this.UsedIds[selId] = user
	logger.Logger.Infof("NiceIdManager pop niceid %v to %v", selId, user)
	return selId
}
func (this *NiceIdManager) PushNiceId(snid int32) {
	if _, ok := this.UsedIds[snid]; ok {
		this.SnIds = append(this.SnIds, snid)
		snidLen := len(this.SnIds)
		index := rand.Intn(snidLen)
		this.SnIds[snidLen-1], this.SnIds[index] = this.SnIds[index], this.SnIds[snidLen-1]
		this.UsedIds[snid] = 0
		logger.Logger.Infof("NiceIdManager push niceid %v to cache", snid)
	}
}
func (this *NiceIdManager) NiceIdCheck(playerid int32) {
	logger.Logger.Infof("%v be used in NiceIdManager.", playerid)
	if userid, ok := this.UsedIds[playerid]; ok {
		delete(this.UsedIds, playerid)
		if userid != 0 {
			user := PlayerMgrSington.GetPlayerBySnId(userid)
			if user != nil {
				user.NiceId = this.PopNiceId(userid)
				if user.scene != nil {
					pack := &server_proto.WGNiceIdRebind{
						User:  proto.Int32(userid),
						NewId: proto.Int32(user.NiceId),
					}
					user.SendToGame(int(server_proto.SSPacketID_PACKET_GW_NICEIDREBIND), pack)
					packNr := &player_proto.SCNiceIdRebind{
						SnidId: proto.Int32(userid),
						NiceId: proto.Int32(user.NiceId),
					}
					user.scene.Broadcast(int(player_proto.PlayerPacketID_PACKET_SC_NICEIDREBIND), packNr, 0)
				}
			}
		} else {
			niceIndex := -1
			for key, value := range this.SnIds {
				if value == playerid {
					niceIndex = key
					break
				}
			}
			if niceIndex != -1 {
				curCount := len(this.SnIds)
				this.SnIds[niceIndex], this.SnIds[curCount-1] = this.SnIds[curCount-1], this.SnIds[niceIndex]
				this.SnIds = this.SnIds[:curCount-1]
			}
		}
	}
}
func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		niceIdMgr.init()
		return nil
	})
}
