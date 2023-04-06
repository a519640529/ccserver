package minipoker

import (
	"fmt"
	"os"
	"testing"
)

var betValue = int64(10)

func TestCalcCardsTypeScore(t *testing.T) {
	type args struct {
		betValue  int64
		cardsType int
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{"cardtype-3", args{betValue, 3}, 1500},
		{"cardtype-4", args{betValue, 4}, 500},
		{"cardtype-5", args{betValue, 5}, 200},
		{"cardtype-6", args{betValue, 6}, 130},
		{"cardtype-7", args{betValue, 7}, 80},
		{"cardtype-8", args{betValue, 8}, 50},
		{"cardtype-9", args{betValue, 9}, 25},
		{"cardtype-10", args{betValue, 10}, 0},
		{"cardtype-11", args{betValue, 11}, 0},
		{"cardtype-12", args{betValue, 12}, 0},
		{"cardtype-13", args{betValue, 13}, 10000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalcCardsTypeScore(tt.args.betValue, tt.args.cardsType); got != tt.want {
				t.Errorf("CalcCardsTypeScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalcCardsType(t *testing.T) {
	type args struct {
		cards []int32
	}
	tests := []struct {
		name string
		args args
		want int
	}{}
	fileName := "test.csv"
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}
	file.WriteString("cards\t cardsType\t wantCardsType\r\n")
	for i := 0; i < len(cardType); i++ {
		for j := 0; j < len(cardType[i]); j++ {
			tests = append(tests, struct {
				name string
				args args
				want int
			}{
				fmt.Sprintf("cardtype-%d-%d", i+3, j),
				args{cardType[i][j]},
				i + 3,
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalcCardsType(tt.args.cards); got != tt.want {
				if (tt.want == 10 || tt.want == 11) && got != 10 && got != 11 {
					t.Errorf("CalcCardsType() = %v, want %v", got, tt.want)
					// str := fmt.Sprintf("{%v}\n", tt.args.cards)
					str := fmt.Sprintf("%v\t %v\t %v\r\n", tt.args.cards, got, tt.want)
					file.WriteString(str)
				}
			}
		})
	}
}

func TestGetCardsName(t *testing.T) {
	type args struct {
		cards []int32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"cardtype-3-2", args{[]int32{6, 19, 28, 34, 46}}, "test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCardsName(tt.args.cards); got != tt.want {
				t.Errorf("GetCardsName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCardsPair(t *testing.T) {
	type args struct {
		cards    []int32
		cardsMap map[int32]int32
	}
	tests := []struct {
		name string
		args args
	}{}
	var cardsMap = make(map[int32]int32, 0)
	for i := 0; i < len(pair); i++ {
		tests = append(tests, struct {
			name string
			args args
		}{
			fmt.Sprintf("GetCardsPair-%d", i+1),
			args{pair[i], cardsMap},
		})
	}
	for _, tt := range tests {
		// t.Run(tt.name, func(t *testing.T) {
		GetCardsPair(tt.args.cards, cardsMap)
		// })
	}
	for k, v := range cardsMap {
		fmt.Printf("cardvalue %v num %v \n", k, v)
	}
}
