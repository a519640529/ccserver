package model

import "math/rand"

const (
	ClubId = "clubid"
)

// ID管理器
type IdManager struct {
	CurList []int
}

func (this *IdManager) PopId() int {
	if len(this.CurList) > 0 {
		id := this.CurList[0]
		this.CurList = this.CurList[1:]
		return id
	} else {
		return 0
	}
}
func (this *IdManager) PushId(id int) {
	this.CurList = append(this.CurList, id)
}
func (this *IdManager) Init() {
	spcId := map[int]bool{111111: true, 222222: true, 333333: true, 444444: true,
		555555: true, 666666: true, 777777: true, 888888: true}
	idArr := rand.Perm(999999)
	for _, value := range idArr {
		if _, ok := spcId[value]; ok {
			continue
		}
		if value > 1000 {
			this.CurList = append(this.CurList, value)
		}
	}
}
