package fruits

import (
	"fmt"
	"os"
	"testing"
)

func TestName(t *testing.T) {
	//eleLineAppearRate := []int32{926, 926, 556, 556, 741, 741, 741, 926, 1111, 1296, 1481}

	fileName := fmt.Sprintf("classic888-%v-%d.csv", 0, 0)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}
	file.WriteString("随机倍率\n")
	for i := 0; i < 100000; i++ {
		var wls WinResult
		//wls.CreateLine(eleLineAppearRate)
		wls.Win()
		var rate int64
		for _, v := range wls.WinLine {
			rate += v.Rate
		}
		str := fmt.Sprintf("%v\r\n", rate)
		file.WriteString(str)
	}

}
