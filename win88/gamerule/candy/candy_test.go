package candy

import (
	"reflect"
	"testing"
)

func TestCalcLine(t *testing.T) {
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
			"test1",
			args{
				[]int{1, 1, 1, 2, 2, 2, 3, 3, 3},
				[]int64{1, 2, 3},
			},
			[]LineData{
				{1, 1, 3, 0, []int32{1, 2, 3}},
				{2, 2, 3, 800, []int32{4, 5, 6}},
				{3, 3, 3, 400, []int32{7, 8, 9}},
			},
		},
		{
			"test bigwild",
			args{
				[]int{1, 2, 3, 1, 5, 6, 3, 3, 3},
				[]int64{1, 2},
			},
			[]LineData{
				{1, 0, 3, 0, []int32{1, 2, 3}},
				{2, 0, 3, 12, []int32{4, 5, 6}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLines := CalcLine(tt.args.data, tt.args.betLines); !reflect.DeepEqual(gotLines, tt.wantLines) {
				t.Errorf("CalcLine() = %v, want %v", gotLines, tt.wantLines)
			}
		})
	}
}
