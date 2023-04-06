package baccarat

import (
	"testing"
)

//测试发牌、洗牌、点数
func TestPoker(t *testing.T) {
	p := NewPoker()
	//t.Log(p.TryNextN(0))
	//t.Log(p.Next())
	for k, e := range p.buf {
		t.Log("第", k+1, "张牌为", e)
	}
	//for k,e := range p.buf {
	//	t.Log("牌值：",e)
	//	t.Log("牌点数：",e%13)
	//	t.Log("百家乐中点数：",GetPointNum(p.buf,k))
	//}

	//局部洗牌
	p.ShuffleNumCard(6)

	for k, e := range p.buf {
		t.Log("第", k+1, "张牌为", e)
	}
	//局部洗牌
	p.ShuffleNumCard(6)

	for k, e := range p.buf {
		t.Log("第", k+1, "张牌为", e)
	}
	//局部洗牌
	p.ShuffleNumCard(6)

	for k, e := range p.buf {
		t.Log("第", k+1, "张牌为", e)
	}
}

//测试快照
func TestSnapshot(t *testing.T) {
	p := NewPoker()
	for k, e := range p.buf {
		t.Log("第", k+1, "张牌为", e)
	}

	//x := p.Snapshot()
	//局部洗牌
	t.Log("=1111111111111111111111111111111111111=")
	p.ShuffleNumCard(20)
	for k, e := range p.buf {
		t.Log("第", k+1, "张牌为", e)
	}

	//t.Log("=222222222222222222222222222222222222222=")
	//for k, e := range x.buf {
	//	t.Log("第", k+1, "张牌为", e)
	//}
}
