package rank

import (
	"games.yol.com/win88/model"
	"games.yol.com/win88/ranksrv/base"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
	"strconv"
	"time"
)

const (
	RANK_LIST_MAXSIZE int = 50 //游戏内榜单最大数量
)

type RankDataNode struct {
	*model.RankData
	dirty bool
}

type RankData struct {
	GameRank map[string]*RankDataNode //gameFreeId做为key 游戏排行数据
}

type RankMgr struct {
	PltRankMap map[string]*RankData //platform做为key
}

var RankMgrSignton = &RankMgr{
	PltRankMap: make(map[string]*RankData),
}

//插入数据
func (rm *RankMgr) UpsertGameRankData(plt string, gamefreeId int32, val int64, snid int32, name string) {
	pltRankData, ok := rm.PltRankMap[plt]
	if !ok {
		pltRankData = &RankData{GameRank: make(map[string]*RankDataNode)}
		rm.PltRankMap[plt] = pltRankData
	}

	key := strconv.Itoa(int(gamefreeId))
	node, ok := pltRankData.GameRank[key]
	if !ok {
		node = &RankDataNode{
			RankData: &model.RankData{
				Id:   bson.NewObjectId(),
				Key:  key,
				Data: make([]*model.RankPlayerData, 0, RANK_LIST_MAXSIZE),
			},
		}
		pltRankData.GameRank[key] = node
	}

	//插入数据
	if node != nil {
		cnt := len(node.Data)
		if cnt >= RANK_LIST_MAXSIZE { //比最小数据还小,直接返回
			minData := node.Data[cnt-1]
			if minData.Val >= val {
				return
			}
		}
		obj := &model.RankPlayerData{
			SnId: snid,
			Name: name,
			Val:  val,
		}
		node.Insert(obj, cnt < RANK_LIST_MAXSIZE)
		node.dirty = true
	}
}

//从数据库中读取数据，初始化排行榜
func (rm *RankMgr) InitGameRankData(plt string) {
	pltRankData, ok := rm.PltRankMap[plt]
	if !ok {
		pltRankData = &RankData{GameRank: make(map[string]*RankDataNode)}
		rm.PltRankMap[plt] = pltRankData
	}

	rankData := model.InitRankData(plt)
	if rankData != nil {
		for _, rankData := range rankData {
			node, ok := pltRankData.GameRank[rankData.Key]
			if !ok {
				node = &RankDataNode{RankData: rankData}
				node.Sort()
				pltRankData.GameRank[rankData.Key] = node
			}
		}
	} else {
		logger.Logger.Warnf("RankMgr.InitGameRankData Can't get platform[%v] rank", plt)
	}
}

func (rm *RankMgr) SaveGameRankData(sync bool) {
	for plt, rd := range rm.PltRankMap {
		for key, node := range rd.GameRank {
			if node.dirty {
				node.dirty = false
				if sync {
					model.SaveRankData(plt, key, node.RankData)
				} else {
					save := func(plt, key string, data *model.RankData) {
						task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
							return model.SaveRankData(plt, key, data)
						}), nil, "SaveRankData").Start()
					}
					save(plt, key, node.Clone())
				}
			}
		}
	}
}

func (rm *RankMgr) ModuleName() string {
	return "RankMgr"
}

func (rm *RankMgr) Init() {
	//初始化排行榜数据
	for k, v := range base.PlatformMgrSington.Platforms {
		if v {
			rm.InitGameRankData(k)
		}
	}
}

func (rm *RankMgr) Update() {
	rm.SaveGameRankData(false)
}

func (rm *RankMgr) Shutdown() {
	rm.SaveGameRankData(true)
	module.UnregisteModule(rm)
}

func init() {
	module.RegisteModule(RankMgrSignton, time.Minute, 0)
}
