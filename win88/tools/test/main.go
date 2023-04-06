package main

import (
	"fmt"
	"games.yol.com/win88/gamerule/ninelineking"
	"math/rand"
	"time"
)

func main() {
	println(time.Now().UnixNano())
	println(fmt.Sprintf("%v", float64(1002340)))

	rand.Seed(time.Now().UnixNano())
	min := 1000
	max := 0
	total := 0
	times := 100
	fmt.Println("---没干扰的情况下的击中次数")
	for n := 0; n < times; n++ {
		cnt := 0
		for i := 0; i < 1000; i++ {
			cnt++
			if rand.Intn(100) == 1 {
				total += cnt
				if cnt < min {
					min = cnt
				}
				if cnt > max {
					max = cnt
				}
				fmt.Println("第", n+1, "次1%的概率，击中时的次数：", cnt)
				break
			}
		}
	}
	fmt.Println("1%的概率，平均击杀次数", total/times, "最小", min, "最大", max)
	fmt.Println("=========================")
	fmt.Println("+++有干扰的情况下的击中次数")
	total = 0
	min = 1000
	max = 0
	for n := 0; n < times; n++ {
		for m := 0; m < rand.Intn(1000); m++ {
			rand.Float64()
		}
		cnt := 0
		for i := 0; i < 1000; i++ {
			cnt++
			if rand.Intn(100) == 1 {
				total += cnt
				if cnt < min {
					min = cnt
				}
				if cnt > max {
					max = cnt
				}
				fmt.Println("第", n+1, "次1%的概率，击中时的次数：", cnt)
				break
			}
		}
	}
	fmt.Println("1%的概率，平均击杀次数", total/times, "最小", min, "最大", max)

	if false {
		var lhPoker [52]int32
		for i := 0; i < 52; i++ {
			lhPoker[i] = int32(i)
		}
		lhDesc := []string{"0", "-", "1"}
		lhWeight := []int32{47059, 5882, 47059}
		RandSliceIndexByWight := func(s1 []int32) int {
			total := int32(0)
			for _, v := range s1 {
				total += v
			}
			if total <= 0 {
				return 0
			}
			random := rand.Int31n(total)
			total = 0
			for i, v := range s1 {
				total += v
				if random < total {
					return i
				}
			}
			return 0
		}
		_ = lhWeight
		_ = RandSliceIndexByWight
		var lhtotal [3]int
		var lhTrend [100]int
		for i := 0; i < 10000; i++ {
			//fmt.Printf("---开始第[%d]局模拟---\r\n", i+1)
			for j := 0; j < 100; j++ {
				lhRand := rand.Perm(52)
				//lhTrend[j] = RandSliceIndexByWight(lhWeight)
				//fmt.Print(lhDesc[lhTrend[j]], " ")
				lPoker := lhPoker[lhRand[0]]%13 + 1
				hPoker := lhPoker[lhRand[1]]%13 + 1
				if lPoker == hPoker {
					lhTrend[j] = 1
				} else if lPoker > hPoker {
					lhTrend[j] = 0
				} else {
					lhTrend[j] = 2
				}
				lhtotal[lhTrend[j]]++
			}
			//fmt.Print("\r\n")
		}
		fmt.Println(lhtotal)
		sum := 0
		for i := 0; i < 3; i++ {
			sum += lhtotal[i]
		}
		for i := 0; i < 3; i++ {
			fmt.Print(lhDesc[i], ":", float64(lhtotal[i])/float64(sum), "% ")
		}
		fmt.Println()
	}
	//百牛路单模拟测试
	if false {
		var rd [4]*rand.Rand
		for i := 0; i < 4; i++ {
			rd[i] = rand.New(rand.NewSource(rand.Int63()))
		}

		var flagstr = []string{"输", "赢"}
		for i := 0; i < 100; i++ {
			fmt.Println("===第", i+1, "组路单")
			var ld [4][20]int
			for j := 0; j < 20; j++ {
				for k := 0; k < 4; k++ {
					ld[k][j] = rand.Intn(2) //rd[k].Intn(2)
				}
			}
			var total int
			var sum [4]int
			for k := 0; k < 4; k++ {
				var flag int
				var maxlong int
				var long int
				v := ld[k][0]
				for j := 0; j < 20; j++ {
					sum[k] += ld[k][j]
					if v == ld[k][j] {
						long++
						if long > maxlong {
							maxlong = long
							flag = v
						}
					} else {
						v = ld[k][j]
						long = 1
					}
				}
				total += sum[k]
				fmt.Println(ld[k], "胜率:", sum[k]*5, "%,", "长龙:", maxlong, "连", flagstr[flag])
			}

			var alllost int
			var allwin int
			for j := 0; j < 20; j++ {
				flag := ld[0][j]
				allsame := true
				for k := 1; k < 4; k++ {
					if ld[k][j] != flag {
						allsame = false
						break
					}
				}
				if allsame {
					if flag == 0 {
						alllost++
					} else {
						allwin++
					}
				}
			}
			fmt.Println("===20局总胜率", total*100/80, "%,", "通杀次数:", alllost, ",通赔次数", allwin)
		}
	}

	//九线拉王模拟测试
	if false {
		type NineLintItem struct {
			cards        [ninelineking.NINELINEKINF_CARD_NUM]int32
			allscore     int32
			allline      int32
			freeTimes    int32
			jackaptRatio float64
		}
		var freepool [180][]*NineLintItem
		var pool [20000][]*NineLintItem
		start := time.Now()
		for i := 0; i < 1000000; i++ {
			var cards [ninelineking.NINELINEKINF_CARD_NUM]int32
			for i := 0; i < ninelineking.NINELINEKINF_CARD_NUM; i++ {
				cards[i] = rand.Int31n(ninelineking.KindOfCard_KingMax)
			}
			allscore, allline, freeTimes, jackaptRatio, _, _ := ninelineking.CalcuNineLineKingScore(cards[:], 9)
			pool[allscore] = append(pool[allscore], &NineLintItem{
				cards:        cards,
				allscore:     allscore,
				allline:      allline,
				freeTimes:    freeTimes,
				jackaptRatio: jackaptRatio,
			})
			if freeTimes > 0 {
				freepool[freeTimes] = append(freepool[freeTimes], &NineLintItem{
					cards:        cards,
					allscore:     allscore,
					allline:      allline,
					freeTimes:    freeTimes,
					jackaptRatio: jackaptRatio,
				})
			}
		}
		fmt.Println("NineLine take:", time.Now().Sub(start))
		//for k, v := range pool {
		//	if len(v) > 0 {
		//		fmt.Println("allscore:", k, "count:", len(v))
		//	}
		//}

		fmt.Println("freetimes :", len(freepool))
		sumline := int32(0)
		for k, v := range freepool {
			if len(v) > 0 {
				fmt.Println("freetimes:", k)
				for i := 0; i < len(v); i++ {
					fmt.Println("freetime:", v[i].freeTimes, "score:", v[i].allscore, "lines:", v[i].allline)
					sumline += v[i].allline
				}
			}
		}
		fmt.Println("total line:", sumline)
	}
	//	start := time.Now()
	//	var t *time.Timer
	//	t = time.AfterFunc(randomDuration(), func() {
	//		fmt.Println(time.Now().Sub(start))
	//		t.Reset(randomDuration())
	//	})
	//	time.Sleep(5 * time.Second)
	//	for i := 0; i < 100; i++ {
	//		fmt.Println(rand.Int31n(4))
	//	}
	//	m := make(map[int][]int)
	//	fmt.Println(len(m[1]))
	//	var mm []int32
	//	fmt.Println(len(mm))
	//	Comb(100, 5)
	//	for dayIdx := 0; dayIdx <= 6; dayIdx++ {
	//		tNow := time.Now()
	//		tStart := tNow.AddDate(0, 0, int(-dayIdx))
	//		tStart = time.Date(tStart.Year(), tStart.Month(), tStart.Day(), 0, 0, 0, 0, tStart.Location())
	//		tEnd := tStart.AddDate(0, 0, 1)
	//		fmt.Println(tStart, "-", tEnd)
	//	}

	//	type TestCase struct {
	//		hands [5]int32
	//		kind  int32
	//		ok    *bullfight.CardsInfoO
	//	}
	//	p := bullfight.NewPoker()
	//	for i := 0; i < 100; i++ {
	//		p.Reset()
	//		var players [5]TestCase
	//		for j := 0; j < 5; j++ {
	//			players[j].kind = rand.Int31n(bullfight.KindOfCard_Max - 3)
	//			for m := 0; m < 4; m++ {
	//				players[j].hands[m] = int32(p.Next())
	//			}
	//			old := int32(p.Next())
	//			players[j].hands[4] = old
	//			players[j].ok = bullfight.KindOfCardFigureUpSington.FigureUpByCard(players[j].hands[:])
	//			if !(players[j].ok != nil && players[j].ok.GetKind() >= players[j].kind) {
	//				restCnt := p.Count()
	//				for k := 0; k < restCnt; k++ {
	//					players[j].hands[4] = int32(p.TryNextN(k))
	//					ok := bullfight.KindOfCardFigureUpSington.FigureUpByCard(players[j].hands[:])
	//					if ok != nil && ok.GetKind() >= players[j].kind {
	//						players[j].ok = ok
	//						p.ChangeNextN(k, bullfight.Card(old))
	//						break
	//					}
	//					if ok != nil && ok.GetKind() > players[j].ok.GetKind() {
	//						players[j].ok = ok
	//					}
	//				}
	//			}
	//			if players[j].ok != nil && players[j].ok.GetKind() >= players[j].kind {
	//				//fmt.Println("[ok]", i, j, players[j].hands, players[j].kind, players[j].ok)
	//			} else {
	//				fmt.Println("[ng]", i, j, players[j].hands, players[j].kind, players[j].ok)
	//			}
	//		}
	//	}

	//	dbConfig := mongo.DbConfig{
	//		Host:     "127.0.0.1",
	//		Database: "jxjy_game",
	//		User:     "mongouser",
	//		Password: "888",
	//	}
	//	dbConfig := mongo.DbConfig{
	//		Host:     "192.168.1.50",
	//		Database: "superstar_game",
	//		User:     "",
	//		Password: "",
	//	}
	//mongo.Config.Dbs["user"] = dbConfig
	//mongo.Config.Init()
	//type PlayerData struct {
	//	Id               bson.ObjectId `bson:"_id"`
	//	SnId             int32         //数字唯一id
	//	Params           []string      //外部参数
	//	PlatformNickName string        //平台昵称(base64,规避特殊字符)
	//	PlatformHead     string        //平台头像
	//	PlatformSex      int           //平台性别
	//}
	//playerrec := mongo.DatabaseC("user", "user_playerinfo")
	//if playerrec != nil {
	//	var recs []model.PlayerData
	//	err := playerrec.Find(nil).All(&recs)
	//	if err == nil {
	//		fmt.Println(len(recs))
	//		for i := 0; i < len(recs); i++ {
	//			if recs[i].Params != nil && len(recs[i].Params) > 0 && recs[i].PlatformNickName == "" {
	//				if len(recs[i].Params) > 2 {
	//					recs[i].PlatformSex, _ = strconv.Atoi(recs[i].Params[2])
	//				}
	//				recs[i].PlatformNickName = base64.StdEncoding.EncodeToString([]byte(recs[i].Params[0]))
	//				if len(recs[i].Params) > 1 {
	//					recs[i].PlatformHead = recs[i].Params[1]
	//				}
	//				//err := playerrec.Update(bson.M{"snid": recs[i].SnId}, bson.D{{"$set", bson.D{{"platformnickname", name}, {"platformhead", head}, {"platformsex", sex}}}})
	//				err := playerrec.Update(bson.M{"_id": recs[i].Id}, &recs[i])
	//				//, bson.D{{"$set", bson.D{{"logintimes", acc.LoginTimes}, {"lastlogouttime", acc.LastLogoutTime}}}}
	//				if err != nil {
	//					fmt.Println(recs[i].SnId, "update failed", err)
	//				} else {
	//					fmt.Println(recs[i].SnId, "update success")
	//				}
	//			}
	//		}
	//	}
	//}

	sInt := []int32{1, 1, 2, 4, 5, 1, 7, 8, 1}
	for i := 0; i < len(sInt); i++ {
		if sInt[i] == 1 {
			sInt = append(sInt[:i], sInt[i+1:]...)
			i--
		}
	}

	fmt.Println(sInt)
}

func randomDuration() time.Duration {
	return time.Duration(rand.Int63n(1e9))
}

func Comb(m, n int) [][]int {
	comb := make([]int, m, m)
	var all [][]int
	comb[0] = n
	var combination func(m, n int)
	combination = func(m, n int) {
		for i := m; i >= n; i-- {
			comb[n] = i
			if n > 1 {
				combination(i-1, n-1)
			} else {
				var t []int
				for j := comb[0]; j > 0; j-- {
					t = append(t, comb[j])
				}
				all = append(all, t)
			}
		}
	}
	combination(m, comb[0])
	fmt.Println(all)
	return all
}
