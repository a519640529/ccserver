package webapi

import "games.yol.com/win88/protocol/webapi"

// 平台配置列表
func API_GetPlatformData(appId string) ([]byte, error) {
	return postRequest(appId, "/game_srv/platform_list", nil, &webapi.SAPlatformInfo{PlatformId: 0}, "http", DEFAULT_TIMEOUT)
}

// 平台游戏配置
func API_GetPlatformConfigData(appId string) ([]byte, error) {
	return postRequest(appId, "/game_srv/game_config_list", nil, nil, "http", DEFAULT_TIMEOUT)
}

// 游戏分组列表
func API_GetGameGroupData(appId string) ([]byte, error) {
	return postRequest(appId, "/game_srv/game_config_group", nil, nil, "http", DEFAULT_TIMEOUT)
}

// 全局游戏开关
func API_GetGlobalGameStatus(appId string) ([]byte, error) {
	return postRequest(appId, "/game_srv/game_config_global", nil, nil, "http", DEFAULT_TIMEOUT)
}
