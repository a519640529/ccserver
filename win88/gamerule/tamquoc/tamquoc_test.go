package tamquoc

import (
	"testing"
)

func TestIsLine(t *testing.T) {
	type TestData struct {
		data []int
		line int
	}
	testData := []TestData{
		{data: []int{0, 0, 0, 1, 1}, line: 3},
		{data: []int{1, 0, 0, 0, 1}, line: 3},
		{data: []int{1, 1, 0, 0, 0}, line: 3},

		{data: []int{0, 1, 0, 0, 1}, line: 3},
		{data: []int{0, 1, 1, 0, 0}, line: 3},

		{data: []int{0, 0, 1, 0, 1}, line: 3},
		{data: []int{0, 0, 1, 1, 0}, line: 3},

		{data: []int{0, 0, 0, 0, 1}, line: 4},
		{data: []int{1, 0, 0, 0, 0}, line: 4},
		{data: []int{0, 1, 0, 0, 0}, line: 4},
		{data: []int{0, 0, 1, 0, 0}, line: 4},
		{data: []int{0, 0, 0, 1, 0}, line: 4},

		{data: []int{0, 0, 0, 0, 0}, line: 5},
	}
	for _, value := range testData {
		if _, count := isLine(value.data); count != value.line {
			t.Error(isLine(value.data))
			t.Error("Error line data:", value)
			t.Fatal("TestIsLine")
		}
	}
	errorData := []TestData{
		{data: []int{1, 2, 0, 0, 1}, line: -1},
		{data: []int{1, 2, 2, 0, 1}, line: -1},
		{data: []int{1, 2, 2, 4, 1}, line: -1},
		{data: []int{1, 2, 2, 1, 3}, line: -1},
	}
	for _, value := range errorData {
		if _, count := isLine(value.data); count != value.line {
			t.Error(isLine(value.data))
			t.Error("Error data:", value)
			t.Fatal("TestIsLine")
		}
	}
}

func TestCalcLine(t *testing.T) {
	type TestData struct {
		data []int
		line int64
	}
	testData := []TestData{
		{data: []int{5, 6, 1, 7, 8, 4, 4, 4, 4, 4, 1, 0, 8, 1, 7}, line: 1},
		{data: []int{5, 6, 1, 7, 8, 4, 4, 0, 0, 4, 1, 0, 8, 1, 7}, line: 1},
	}
	for _, value := range testData {
		lines := CalcLine(value.data, []int64{value.line})
		if int64(lines[0].Index) != value.line {
			t.Log("lines:", lines)
			t.Log("Error line data:", value.data)
			t.Fatal("TestIsLine")
		}
	}

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
			name: "A",
			args: args{
				totalBet:   100,
				startBonus: 1,
			},
			want: BonusGameResult{},
		},
		{
			name: "B",
			args: args{
				totalBet:   250,
				startBonus: 5,
			},
			want: BonusGameResult{},
		},
		{
			name: "C",
			args: args{
				totalBet:   250,
				startBonus: 8000,
			},
			want: BonusGameResult{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateBonusGame(tt.args.totalBet, tt.args.startBonus)
			t.Logf("GenerateBonusGame() = %v", got)
		})
	}
}

func TestCalcLine1(t *testing.T) {
	type args struct {
		data     []int
		betLines []int64
	}
	tests := []struct {
		name      string
		args      args
		wantLines []LineData
	}{
		{
			name: "A",
			args: args{
				data:     []int{6, 7, 2, 1, 4, 7, 3, 5, 5, 3, 3, 5, 7, 4, 5},
				betLines: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			wantLines: nil,
		},
		{
			name: "B",
			args: args{
				data:     []int{4, 3, 2, 6, 5, 5, 4, 6, 6, 3, 2, 6, 7, 5, 5},
				betLines: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			wantLines: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLines := CalcLine(tt.args.data, tt.args.betLines)
			PrintHuman(tt.args.data)
			t.Logf("CalcLine() = %v", gotLines)
		})
	}
}
