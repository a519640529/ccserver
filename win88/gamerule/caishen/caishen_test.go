package caishen

import (
	"testing"
)

func TestGenBonusGame(t *testing.T) {
	GenerateBonusGame(10, 3)
}
func TestGenerateBonusGame(t *testing.T) {
	type args struct {
		totalBet   int
		startBonus int
	}
	tests := []struct {
		name string
		args args
		want BonusGameResult
	}{
		{
			name: "bet:100 bonus*3",
			args: args{
				totalBet:   100,
				startBonus: 1,
			},
			want: BonusGameResult{
				BonusData:       nil,
				DataMultiplier:  0,
				Mutiplier:       0,
				TotalPrizeValue: 0,
			},
		},
		{
			name: "bet:100 bonus*5",
			args: args{
				totalBet:   100,
				startBonus: 3,
			},
			want: BonusGameResult{
				BonusData:       nil,
				DataMultiplier:  0,
				Mutiplier:       0,
				TotalPrizeValue: 0,
			},
		},
		{
			name: "bet:1000 bonus*2",
			args: args{
				totalBet:   100,
				startBonus: 3,
			},
			want: BonusGameResult{
				BonusData:       nil,
				DataMultiplier:  0,
				Mutiplier:       0,
				TotalPrizeValue: 0,
			},
		},
		{
			name: "bet:1000 bonus*0",
			args: args{
				totalBet:   1000,
				startBonus: 0,
			},
			want: BonusGameResult{
				BonusData:       nil,
				DataMultiplier:  0,
				Mutiplier:       0,
				TotalPrizeValue: 0,
			},
		},
		{
			name: "bet:0 bonus*3",
			args: args{
				totalBet:   0,
				startBonus: 3,
			},
			want: BonusGameResult{
				BonusData:       nil,
				DataMultiplier:  0,
				Mutiplier:       0,
				TotalPrizeValue: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := GenerateBonusGame(tt.args.totalBet, tt.args.startBonus)
			t.Logf("BonusData:%v", r.BonusData)
			t.Logf("TotalPrizeValue:%v", r.TotalPrizeValue)
			t.Logf("Mutiplier:%v", r.Mutiplier)
			t.Logf("DataMultiplier:%v", r.DataMultiplier)
		})
	}
}

//func TestGenerateSlotsData_v2(t *testing.T) {
//	type args struct {
//		s Symbol
//	}
//	tests := []struct {
//		name  string
//		args  args
//		want  []int
//		want1 int
//	}{
//		{
//			name: "01",
//			args: args{
//				s: SYMBOL1,
//			},
//			want:  nil,
//			want1: 0,
//		},
//		{
//			name: "02",
//			args: args{
//				s: SYMBOL2,
//			},
//			want:  nil,
//			want1: 0,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			v, times := GenerateSlotsData_v2(tt.args.s)
//			t.Logf("GenerateSlotsData_v2() data = %v, times %v", v, times)
//			lines := CalcLine(v, AllBetLines)
//			t.Logf("lines:%v", lines)
//			PrintHuman(v)
//			CaclScore(v, AllBetLines)
//		})
//	}
//}
