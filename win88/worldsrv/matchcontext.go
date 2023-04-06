package main

import "sort"

type MatchContext struct {
	tm     *TmMatch        //比赛
	p      *Player         //玩家数据
	scene  *Scene          //比赛房间
	round  int32           //第几轮
	seq    int             //报名序号
	grade  int32           //比赛积分
	rank   int32           //当前第几名
	gaming bool            //是否比赛中
	record map[int32]int32 //战绩
}

type MatchContextSlice []*MatchContext

func (p MatchContextSlice) Len() int { return len(p) }
func (p MatchContextSlice) Less(i, j int) bool {
	if p[i].grade > p[j].grade {
		return true
	} else if p[i].grade == p[j].grade {
		return p[i].seq < p[j].seq
	} else {
		return false
	}
}

func (p MatchContextSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p MatchContextSlice) Sort(isFinals bool) {
	sort.Sort(p)
	for i, mc := range p {
		mc.rank = int32(i + 1)
	}
	if isFinals {
		lastRank := int32(0)
		lastGrade := int32(0)
		for i := 0; i < len(p); i++ {
			mc := p[i]
			if i > 0 && mc.grade == lastGrade {
				mc.rank = lastRank
			}
			lastRank = mc.rank
			lastGrade = mc.grade
		}
	}
}

func NewMatchContext(p *Player, tm *TmMatch, grade int32, seq int) *MatchContext {
	return &MatchContext{
		tm:    tm,
		p:     p,
		grade: grade,
		seq:   seq,
		rank:  int32(seq),
	}
}
