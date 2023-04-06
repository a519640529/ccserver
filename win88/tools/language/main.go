package main

import (
	"bytes"
	"fmt"
	"github.com/idealeak/goserver/core"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Language struct {
	Excel string
	Data  map[string][]string
}

func main() {
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	file := filepath.Join(Config.WorkPath, Config.XlsxPath, "Language.xlsx")
	xlsxFile, err := xlsx.OpenFile(file)
	if err != nil {
		fmt.Println("excel file open error:", err, "filename:", file)
		return
	}
	sheet1 := xlsxFile.Sheets[0]
	var languages = make(map[string]*Language)
	var language = new(Language)
	var excelName string
	var total = len(sheet1.Rows[0].Cells)
	for _, row := range sheet1.Rows[2:] {
		key := strings.Replace(row.Cells[1].Value, "_", "", -1)
		if excelName != key {
			if excelName != "" {
				if _, ok := languages[excelName]; !ok {
					languages[excelName] = language
				} else {
					for s, i := range language.Data {
						languages[excelName].Data[s] = i
					}
				}
			}
			//excelName = row.Cells[1].Value
			excelName = key
			language = &Language{Data: map[string][]string{}}
			language.Excel = excelName
		}
		var val []string
		for i := 3; i < total; i++ {
			var val2 = "\"\""
			if i < len(row.Cells) {
				val2 = strconv.Quote(row.Cells[i].Value)
			}
			val = append(val, val2)
		}

		language.Data[row.Cells[2].Value] = val
	}

	if _, ok := languages[language.Excel]; !ok {
		languages[language.Excel] = language
	} else {
		for s, i := range language.Data {
			languages[language.Excel].Data[s] = i
		}
	}

	fmt.Println(len(languages))
	//for _, v := range languages {
	//	fmt.Println(v.Excel)
	//	for m, n := range v.Data {
	//		fmt.Println(m, "===", n)
	//	}
	//}
	var needLang []Language
	for _, v := range languages {
		needLang = append(needLang, *v)
	}
	for k := range Config.LanguageType {
		if k < total-3 {
			geneLanguageHelper(needLang, k)
		}
	}
}

func geneLanguageHelper(languages []Language, num int) {
	filename := Config.LanguageType[num]
	dm := map[string]interface{}{
		"Languages": languages,
		"Idx":       num,
	}
	outputHelper := bytes.NewBuffer(nil)
	err := templates.ExecuteTemplate(outputHelper, "languagets", dm)
	if err != nil {
		fmt.Println("geneLanguageHelper ExecuteTemplate error:", err)
		return
	}
	helperGoFile := filepath.Join(Config.WorkPath, Config.TsPath, filename)
	err = ioutil.WriteFile(helperGoFile, outputHelper.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Println("geneLanguageHelper WriteFile error:", err)
		return
	}
	//for k, v := range languages {
	//	for m, n := range v.Data {
	//		if len(n) > 0 {
	//			languages[k].Data[m] = n[1:]
	//		}
	//	}
	//}
}
