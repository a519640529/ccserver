package srvdata

import "encoding/json"

type ClientVer struct {
	MinApkVer    int32
	LatestApkVer int32
	MinResVer    int32
	LatestResVer int32
}

var ClientVersSingeton = &ClientVers{
	packVers: make(map[string]map[string]ClientVer),
	gameVers: make(map[string]map[string]ClientVer),
}

type ClientVers struct {
	packVers map[string]map[string]ClientVer
	gameVers map[string]map[string]ClientVer
}

func updateClientVers() error {
	packVers := make(map[string]map[string]ClientVer)
	gameVers := make(map[string]map[string]ClientVer)
	for _, clientVer := range PBDB_ClientVerMgr.Datas.Arr {
		packVerStr := clientVer.GetPackVers()
		ver := make(map[string]ClientVer)
		err := json.Unmarshal([]byte(packVerStr), &ver)
		if err == nil {
			packVers[clientVer.GetPackageFlag()] = ver
		}
		gameVerStr := clientVer.GetGameVers()
		ver = make(map[string]ClientVer)
		err = json.Unmarshal([]byte(gameVerStr), &ver)
		if err == nil {
			gameVers[clientVer.GetPackageFlag()] = ver
		}
	}
	ClientVersSingeton.packVers = packVers
	ClientVersSingeton.gameVers = gameVers
	return nil
}

func GetPackVers(platformTag string) map[string]ClientVer {
	if vers, exist := ClientVersSingeton.packVers[platformTag]; exist {
		return vers
	}
	return nil
}

func GetGameVers(platformTag string) map[string]ClientVer {
	if vers, exist := ClientVersSingeton.gameVers[platformTag]; exist {
		return vers
	}
	return nil
}
