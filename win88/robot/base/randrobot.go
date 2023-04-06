package base

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"

	"games.yol.com/win88/srvdata"
)

const HEAD_URL = "http://cdnres.doudoubei.com/doubeires/head/"

type Zone struct {
	Area   string
	Weight int
}

var Zones []Zone
var ZonesWeight int

type HeadPool struct {
	Boy  []string
	Girl []string
}

var HEADIMG_POOL = &HeadPool{}

func init() {
	buf, err := ioutil.ReadFile("../data/icon_rob.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(buf, HEADIMG_POOL)
	if err != nil {
		panic(err)
	}

	buf, err = ioutil.ReadFile("../data/zone_rob.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(buf, &Zones)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(Zones); i++ {
		ZonesWeight += Zones[i].Weight
		Zones[i].Weight = ZonesWeight
	}
}

func CalcuHashCode(str string) int64 {
	data := []byte(str)
	cnt := len(data)
	code := int64(0)
	for i := 0; i < cnt; i++ {
		code = 31*code + int64(data[i])
	}
	return code
}

func RandRobotInfo(accId string) (name, icon, sex, ip string) {
	ts := time.Now().Unix()
	code := CalcuHashCode(fmt.Sprintf("%v-%v", ts, accId))
	rand.Seed(code)
	nSex := 1 + rand.Int31n(2)
	sex = strconv.Itoa(int(nSex))
	var iconCnt int
	var iconPool []string
	switch nSex {
	case 1:
		namePool := srvdata.PBDB_NameBoyMgr.Datas.GetArr()
		namePoolLen := int32(len(namePool))
		if namePoolLen > 0 {
			name = namePool[rand.Int31n(namePoolLen)].GetName()
		}
		iconPool = HEADIMG_POOL.Boy
		iconCnt = len(iconPool)
	case 2:
		namePool := srvdata.PBDB_NameGirlMgr.Datas.GetArr()
		namePoolLen := int32(len(namePool))
		if namePoolLen > 0 {
			name = namePool[rand.Int31n(namePoolLen)].GetName()
		}
		iconPool = HEADIMG_POOL.Girl
		iconCnt = len(iconPool)
	}

	if iconCnt > 0 {
		icon = Config.HeadUrl + iconPool[rand.Int31n(int32(iconCnt))]
	}
	ip = fmt.Sprintf("%v.%v.%v.%v", 1+rand.Int31n(255), 1+rand.Int31n(255), 1+rand.Int31n(255), 1+rand.Int31n(255))
	return
}

func RandLongitudeAndLatitude() (longitude, latitude int32) {
	//中国范围经纬度
	//113216260,34963671
	//114103930,34592702
	longitude = rand.Int31n(120212509-110241174) + 110241174
	latitude = rand.Int31n(40960003-21558402) + 21558402
	return longitude, latitude
}

func RandZone() string {
	idx := -1
	weight := rand.Intn(ZonesWeight)
	for i := 0; i < len(Zones); i++ {
		if Zones[i].Weight >= weight {
			idx = i
			break
		}
	}
	if idx != -1 {
		return Zones[idx].Area
	}
	return "深圳"
}
