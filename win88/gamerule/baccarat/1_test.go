package baccarat

import (
	"fmt"
	"testing"
)

func TestExample(t *testing.T) {
	//p := NewPoker()
	//cards := p.GetTIE()
	//PrintNewCards(cards, 1, p)
	//cards = p.GetBankerWin()
	//PrintNewCards(cards, 2, p)
	//cards = p.GetXianWin()
	//PrintNewCards(cards, 3, p)
	//cards = p.GetBankerAndBankerPair()
	//PrintNewCards(cards, 4, p)
	//cards = p.GetBankerAndXianPair()
	//PrintNewCards(cards, 5, p)
	//cards = p.GetXianAndXianPair()
	//PrintNewCards(cards, 6, p)
	//cards = p.GetXianAndBankerPair()
	//PrintNewCards(cards, 7, p)
	//cards = p.GetBankerAndBankerXianPair()
	//PrintNewCards(cards, 8, p)
	//cards = p.GetXianAndBankerXianPair()
	//PrintNewCards(cards, 9, p)
	//cards = p.GetTieAndBankerPair()
	//PrintNewCards(cards, 10, p)
	//cards = p.GetTieAndXianPair()
	//PrintNewCards(cards, 11, p)
	//cards = p.GetTieAndBankerXianPair()
	//PrintNewCards(cards, 12, p)

	//cards := []int32{7, 7, -1, 5, 11, -1}
	//PrintNewCards(cards, 1, p)
}
func PrintNewCards(cs []int32, n int, p *Poker) {
	if len(cs) == 0 {
		fmt.Println(" n ")
		return
	}
	PrintTheCards(cs, n)
	r := false
	ncs := make([]int32, 0, 0)
	r, ncs = p.SingleRepairCard(cs)
	fmt.Println(r)
	if r {
		cs = ncs
		PrintTheCards(cs, n)
	}
	PrintTheCards2(cs, n)
	fmt.Println("==========================================")
	cs = []int32{-1, -1, -1, -1, -1, -1}
}
func PrintTheCards(cs []int32, n int) {
	if len(cs) == 0 {
		fmt.Println("  ")
		return
	}
	fmt.Print(n, ":    ")
	for i := 0; i < 6; i++ {
		if i == 3 {
			fmt.Print("  ")
		}
		if cs[i] != -1 {
			fmt.Print(cs[i]%13+1, " ")
		} else {
			fmt.Print(-1, " ")
		}
	}
	fmt.Println("  ")
}
func PrintTheCards2(cs []int32, n int) {
	if len(cs) == 0 {
		fmt.Println("  ")
		return
	}
	fmt.Print(n, ":    ")
	for i := 0; i < 6; i++ {
		if i == 3 {
			fmt.Print("  ")
		}
		if cs[i] != -1 {
			fmt.Print(cs[i]%52, " ")
		} else {
			fmt.Print(-1, " ")
		}
	}
	fmt.Println("  ")
}
