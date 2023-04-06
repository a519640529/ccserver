package common

const (
	ClientState_WaitBindBundle int = iota
	ClientState_WaitLogin
	ClientState_Logined
	ClientState_WaiteGetPlayerInfo
	ClientState_CreatePlayer
	ClientState_GetPlayerInfo
	ClientState_EnterGame
	ClientState_EnterMap
	ClientState_EnterFight
	ClientState_WaitLogout
	ClientState_Logouted
	ClientState_Dropline
	ClientState_Droplined
	ClientState_WaitRehold
	ClientState_Reholded
)
