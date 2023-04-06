package common

import (
	"testing"
	"time"
)

type TimeTestCaseData struct {
	t1           time.Time
	t2           time.Time
	expectResult bool
}

func TestInSameDay(t *testing.T) {

	testCases := []*TimeTestCaseData{
		&TimeTestCaseData{
			t1:           time.Date(2016, time.May, 17, 15, 12, 15, 0, time.Local),
			t2:           time.Date(2016, time.May, 16, 15, 12, 15, 0, time.Local),
			expectResult: false,
		},
		&TimeTestCaseData{
			t1:           time.Date(2016, time.May, 16, 23, 59, 59, 0, time.Local),
			t2:           time.Date(2016, time.May, 16, 15, 12, 15, 0, time.Local),
			expectResult: true,
		},
		&TimeTestCaseData{
			t1:           time.Date(2017, time.May, 16, 23, 59, 59, 0, time.Local),
			t2:           time.Date(2016, time.May, 16, 15, 12, 15, 0, time.Local),
			expectResult: false,
		},
	}

	for i := 0; i < len(testCases); i++ {
		tc := testCases[i]
		if InSameDay(tc.t1, tc.t2) != tc.expectResult {
			t.Fatal("IsSameDay(", tc.t1, tc.t2, ") expect result is ", tc.expectResult)
		}
	}
}

func TestIsContinueDay(t *testing.T) {

	testCases := []*TimeTestCaseData{
		&TimeTestCaseData{
			t1:           time.Date(2016, time.May, 17, 15, 12, 15, 0, time.Local),
			t2:           time.Date(2016, time.May, 16, 15, 12, 15, 0, time.Local),
			expectResult: true,
		},
		&TimeTestCaseData{
			t1:           time.Date(2016, time.May, 16, 23, 59, 59, 0, time.Local),
			t2:           time.Date(2016, time.May, 16, 15, 12, 15, 0, time.Local),
			expectResult: false,
		},
		&TimeTestCaseData{
			t1:           time.Date(2017, time.May, 17, 23, 59, 59, 0, time.Local),
			t2:           time.Date(2016, time.May, 16, 15, 12, 15, 0, time.Local),
			expectResult: false,
		},
	}

	for i := 0; i < len(testCases); i++ {
		tc := testCases[i]
		if IsContinueDay(tc.t1, tc.t2) != tc.expectResult {
			t.Fatal("IsContinueDay(", tc.t1, tc.t2, ") expect result is ", tc.expectResult)
		}
	}
}

func TestInSameMonth(t *testing.T) {

	testCases := []*TimeTestCaseData{
		&TimeTestCaseData{
			t1:           time.Date(2016, time.May, 17, 15, 12, 15, 0, time.Local),
			t2:           time.Date(2016, time.May, 16, 15, 12, 15, 0, time.Local),
			expectResult: true,
		},
		&TimeTestCaseData{
			t1:           time.Date(2016, time.June, 1, 23, 59, 59, 0, time.Local),
			t2:           time.Date(2016, time.May, 31, 15, 12, 15, 0, time.Local),
			expectResult: false,
		},
		&TimeTestCaseData{
			t1:           time.Date(2017, time.May, 17, 23, 59, 59, 0, time.Local),
			t2:           time.Date(2016, time.May, 16, 15, 12, 15, 0, time.Local),
			expectResult: false,
		},
	}

	for i := 0; i < len(testCases); i++ {
		tc := testCases[i]
		if InSameMonth(tc.t1, tc.t2) != tc.expectResult {
			t.Fatal("InSameMonth(", tc.t1, tc.t2, ") expect result is ", tc.expectResult)
		}
	}
}

func TestInSameWeek(t *testing.T) {

	testCases := []*TimeTestCaseData{
		&TimeTestCaseData{
			t1:           time.Date(2016, time.May, 17, 15, 12, 15, 0, time.Local),
			t2:           time.Date(2016, time.May, 16, 15, 12, 15, 0, time.Local),
			expectResult: true,
		},
		&TimeTestCaseData{
			t1:           time.Date(2016, time.May, 16, 23, 59, 59, 0, time.Local),
			t2:           time.Date(2016, time.May, 15, 15, 12, 15, 0, time.Local),
			expectResult: false,
		},
		&TimeTestCaseData{
			t1:           time.Date(2016, time.January, 1, 23, 59, 59, 0, time.Local),
			t2:           time.Date(2015, time.December, 31, 15, 12, 15, 0, time.Local),
			expectResult: true,
		},
		&TimeTestCaseData{
			t1:           time.Date(2016, time.January, 3, 23, 59, 59, 0, time.Local),
			t2:           time.Date(2016, time.January, 4, 15, 12, 15, 0, time.Local),
			expectResult: false,
		},
	}

	for i := 0; i < len(testCases); i++ {
		tc := testCases[i]
		if InSameWeek(tc.t1, tc.t2) != tc.expectResult {
			t.Fatal("InSameWeek(", tc.t1, tc.t2, ") expect result is ", tc.expectResult)
		}
	}
}
