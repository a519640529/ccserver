package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/netlib"
	srvproto "github.com/idealeak/goserver/srvlib/protocol"
)

var gameStateMgr = &GameStateManager{
	gameList: make(map[int32]map[int32]*Player),
	gameIds:  make(map[int32][]int32),
}

type GameStateManager struct {
	gameList map[int32]map[int32]*Player //gameid-snid-player 推送消息的用户列表
	gameIds  map[int32][]int32
}

var ids = []int32{
	int32(common.GameId_RollCoin),
	int32(common.GameId_RollColor),
	int32(common.GameId_RedVsBlack),
	int32(common.GameId_DragonVsTiger),
	int32(common.GameId_Baccarat),
	int32(common.GameId_Roulette),
	int32(common.GameId_RollPoint),
	int32(common.GameId_RollAnimals),
	//int32(common.GameId_BlackJack),
	int32(common.GameId_HundredDZNZ),
	int32(common.GameId_HundredYXX),
	int32(common.GameId_Crash),
}

func (gsm *GameStateManager) Init() {
	var idsMap = make(map[int32]bool)
	for _, v := range ids {
		idsMap[v] = true
	}
	dbGameFree := srvdata.PBDB_GameFreeMgr.Datas.Arr
	for _, gfs := range dbGameFree {
		if _, ok := idsMap[gfs.GameId]; ok {
			gsm.gameIds[gfs.GameId] = append(gsm.gameIds[gfs.GameId], gfs.Id)
		}
	}
}
func (gsm *GameStateManager) PlayerRegiste(player *Player, gameid int32, b bool) {
	playerList := gsm.gameList[gameid]
	if playerList == nil {
		playerList = make(map[int32]*Player)
		gsm.gameList[gameid] = playerList
	}
	playerList[player.SnId] = player
}
func (gsm *GameStateManager) PlayerClear(player *Player) {
	for _, value := range gsm.gameList {
		if value == nil {
			continue
		}
		delete(value, player.SnId)
	}
}
func (gsm *GameStateManager) BrodcastGameState(gameId int32, platform string, packid int, pack interface{}) {
	mgs := make(map[*netlib.Session][]*srvproto.MCSessionUnion)
	playerList := gsm.gameList[gameId]
	for _, p := range playerList {
		if p != nil && p.gateSess != nil && p.IsOnLine() && p.Platform == platform {
			mgs[p.gateSess] = append(mgs[p.gateSess], &srvproto.MCSessionUnion{
				Mccs: &srvproto.MCClientSession{
					SId: proto.Int64(p.sid),
				},
			})
		}
	}
	for gateSess, v := range mgs {
		if gateSess != nil && len(v) != 0 {
			pack, err := MulticastMaker.CreateMulticastPacket(packid, pack, v...)
			if err == nil {
				proto.SetDefaults(pack)
				gateSess.Send(int(srvproto.SrvlibPacketID_PACKET_SS_MULTICAST), pack)
			}
		}
	}
}
func init() {
	//使用并行加载
	RegisteParallelLoadFunc("选场游戏场次配置", func() error {
		gameStateMgr.Init()
		return nil
	})
	//gameStateMgr.gameIds[int32(common.GameId_RollCoin)] = []int32{110030001, 110030002, 110030003, 110030004}
	//gameStateMgr.gameIds[int32(common.GameId_RollColor)] = []int32{150010001, 150010002, 150010003, 150010004}
	//gameStateMgr.gameIds[int32(common.GameId_RedVsBlack)] = []int32{140010001, 140010002, 140010003, 140010004}
	//gameStateMgr.gameIds[int32(common.GameId_DragonVsTiger)] = []int32{120010001, 120010002, 120010003, 120010004}
	//gameStateMgr.gameIds[int32(common.GameId_Baccarat)] = []int32{350010001, 350010002, 350010003, 350010004}
	//gameStateMgr.gameIds[int32(common.GameId_Roulette)] = []int32{540000001, 540000002, 540000003, 540000004}
	//gameStateMgr.gameIds[int32(common.GameId_RollPoint)] = []int32{530000001, 530000002, 530000003, 530000004}
	//gameStateMgr.gameIds[int32(common.GameId_RollAnimals)] = []int32{560000001, 560000002, 560000003, 560000004}
	//gameStateMgr.gameIds[int32(common.GameId_BlackJack)] = []int32{450000001, 450000002, 450000003, 450000004, 450000005}
	//gameStateMgr.gameIds[int32(common.GameId_HundredDZNZ)] = []int32{660000001, 660000002, 660000003, 660000004}
	//gameStateMgr.gameIds[int32(common.GameId_HundredYXX)] = []int32{670000001, 670000002, 670000003, 670000004}
	//
	//// 冰河世纪, 百战成神, 财神, 复仇者联盟, 复活岛
	//gameStateMgr.gameIds[int32(common.GameId_CaiShen)] = []int32{790000001, 790000002, 790000003, 790000004}
	//gameStateMgr.gameIds[int32(common.GameId_Avengers)] = []int32{800000001, 800000002, 800000003, 800000004}
	//gameStateMgr.gameIds[int32(common.GameId_EasterIsland)] = []int32{810000001, 810000002, 810000003, 810000004}
	//gameStateMgr.gameIds[int32(common.GameId_IceAge)] = []int32{820000001, 820000002, 820000003}
	//gameStateMgr.gameIds[int32(common.GameId_TamQuoc)] = []int32{830000001, 830000002, 830000003}
}
