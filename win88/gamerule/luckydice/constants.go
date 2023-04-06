package luckydice

// BetSide 下注大小
const (
	Big int32 = iota
	Small
)

// 机器人下注倍率 须乘上底注
var BetValues = [][]int64{
	{10, 20, 30, 40, 50, 60, 70, 80, 0, 90, 100, 150, 200, 250, 350, 500},                  //poor 2000
	{10, 20, 50, 100, 200, 300, 400, 0, 500, 1000, 5550, 6000, 6660, 9990},                 //normal 2000
	{10, 20, 50, 100, 150, 300, 500, 0, 1000, 1500, 2000, 3000, 5000, 10000, 15000, 20000}, //rich 999
}
