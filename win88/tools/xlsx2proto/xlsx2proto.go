package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/idealeak/goserver/core"
	"github.com/tealeg/xlsx"
)

type SheetColumnMetaStruct struct {
	ColName  string
	ColType  int
	CTString string
	IsArray  bool
}

type SheetMetaStruct struct {
	AbsPath   string
	FileName  string
	ProtoName string
	Cols      []*SheetColumnMetaStruct
}

const (
	INTTYPE         string = "(int)"
	INT64TYPE       string = "(int64)"
	STRTYPE                = "(str)"
	ARRINTTYPE             = "(arrint)"
	ARRINT64TYPE           = "(arrint64)"
	ARRSTRTYPE             = "(arrstr)"
	INTTYPE_PROTO          = "int32"
	INT64TYPE_PROTO        = "int64"
	STRTYPE_PROTO          = "string"
)

var templates *template.Template

func main() {

	defer core.ClosePackages()
	core.LoadPackages("config.json")

	smsMap := make(map[string]*SheetMetaStruct)
	for xlsxFileName, xlsxFilePath := range XlsxFiles {
		xlsxFile, err := xlsx.OpenFile(xlsxFilePath)
		if err != nil {
			fmt.Println("excel file open error:", err, " filePath:", xlsxFilePath)
			continue
		}

		for _, sheet := range xlsxFile.Sheets {
			sms := &SheetMetaStruct{
				AbsPath:   xlsxFilePath,
				FileName:  xlsxFileName,
				ProtoName: strings.TrimSuffix(xlsxFileName, ".xlsx"),
				Cols:      make([]*SheetColumnMetaStruct, 0, sheet.MaxCol)}

			for _, row := range sheet.Rows {
				for _, cell := range row.Cells {
					s := cell.String()
					if strings.HasSuffix(s, INTTYPE) {
						sms.Cols = append(sms.Cols, &SheetColumnMetaStruct{strings.TrimSuffix(s, INTTYPE), 1, INTTYPE_PROTO, false})
					} else if strings.HasSuffix(s, STRTYPE) {
						sms.Cols = append(sms.Cols, &SheetColumnMetaStruct{strings.TrimSuffix(s, STRTYPE), 2, STRTYPE_PROTO, false})
					} else if strings.HasSuffix(s, ARRINTTYPE) {
						sms.Cols = append(sms.Cols, &SheetColumnMetaStruct{strings.TrimSuffix(s, ARRINTTYPE), 3, INTTYPE_PROTO, true})
					} else if strings.HasSuffix(s, ARRSTRTYPE) {
						sms.Cols = append(sms.Cols, &SheetColumnMetaStruct{strings.TrimSuffix(s, ARRSTRTYPE), 4, STRTYPE_PROTO, true})
					} else if strings.HasSuffix(s, INT64TYPE) {
						sms.Cols = append(sms.Cols, &SheetColumnMetaStruct{strings.TrimSuffix(s, INT64TYPE), 5, INT64TYPE_PROTO, false})
					} else if strings.HasSuffix(s, ARRINT64TYPE) {
						sms.Cols = append(sms.Cols, &SheetColumnMetaStruct{strings.TrimSuffix(s, ARRINT64TYPE), 6, INT64TYPE_PROTO, true})
					}
				}
				break //only fetch first row
			}

			smsMap[sms.ProtoName] = sms
			break //only fetch first sheet
		}
	}

	geneProto(smsMap)
}

func geneProto(xlsxs map[string]*SheetMetaStruct) {
	if xlsxs == nil {
		return
	}

	geneAgcHelper(xlsxs)       //生成 xlsx转二进制文件
	geneDataStructProto(xlsxs) //生成 xlsx转protobuf结构文件
	//geneDataSingleStructProto(xlsxs)
	//geneChhDataMgrHeader(xlsxs)//生成 c头文件
	//geneCppDataMgrHeader(xlsxs)//生成 c源文件
	//copyNoDataProtoToLua()
	for _, val := range xlsxs {
		genGoDataMgr(val) //生成 go对象文件
		genTsDataMgr(val) //生成 ts对象文件
	}
}

func geneAgcHelper(xlsxs map[string]*SheetMetaStruct) {
	dm := map[string]interface{}{
		"data":  xlsxs,
		"opath": filepath.Join(Config.WorkPath, Config.DataPath),
	}
	outputHelper := bytes.NewBuffer(nil)
	err := templates.ExecuteTemplate(outputHelper, "agc", dm)
	if err != nil {
		fmt.Println("geneProto ExecuteTemplate error:", err)
		return
	}

	helperGoFile := filepath.Join(Config.WorkPath, Config.ConvertToolPath, "agc.go")
	err = ioutil.WriteFile(helperGoFile, outputHelper.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Println("geneProto WriteFile error:", err)
		return
	}
}

func geneDataStructProto(xlsxs map[string]*SheetMetaStruct) {

	outputHelper := bytes.NewBuffer(nil)
	err := templates.ExecuteTemplate(outputHelper, "gpb", xlsxs)
	if err != nil {
		fmt.Println("geneDataStructProto ExecuteTemplate error:", err)
		return
	}

	protoFile := filepath.Join(Config.WorkPath, Config.ProtoPath, Config.ProtoFile)
	err = ioutil.WriteFile(protoFile, outputHelper.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Println("geneDataStructProto error:", err)
	}
}

func geneDataSingleStructProto(xlsxs map[string]*SheetMetaStruct) {
	for k, v := range xlsxs {
		outputHelper := bytes.NewBuffer(nil)
		err := templates.ExecuteTemplate(outputHelper, "gpb_single", v)
		if err != nil {
			fmt.Println("geneDataSingleStructProto ExecuteTemplate error:", err)
			return
		}

		protoFile := filepath.Join(Config.WorkPath, Config.ProtoLuaPath, strings.ToLower(k)+".proto")
		err = ioutil.WriteFile(protoFile, outputHelper.Bytes(), os.ModePerm)
		if err != nil {
			fmt.Println("geneDataSingleStructProto error:", err)
		}
	}
}

func genGoDataMgr(sms *SheetMetaStruct) {
	if sms == nil {
		return
	}

	outputHelper := bytes.NewBuffer(nil)
	err := templates.ExecuteTemplate(outputHelper, "gpb_mgr", sms)
	if err != nil {
		fmt.Println("genGoDataMgr ExecuteTemplate error:", err)
		return
	}
	opath := filepath.Join(Config.WorkPath, Config.GoFilePath, strings.ToLower(fmt.Sprintf("%v.go", sms.ProtoName)))
	err = ioutil.WriteFile(opath, outputHelper.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Println("genGoDataMgr WriteFile error:", err)
		return
	}
}

//生成ts类模板
func genTsDataMgr(sms *SheetMetaStruct) {
	if sms == nil {
		return
	}

	outputHelper := bytes.NewBuffer(nil)
	err := templates.ExecuteTemplate(outputHelper, "gpb_mgr_ts", sms)
	if err != nil {
		fmt.Println("genTsDataMgr ExecuteTemplate error:", err)
		return
	}
	//opath := filepath.Join(Config.WorkPath, Config.TsFilePath, strings.ToLower(fmt.Sprintf("%v.ts", sms.ProtoName)))
	opath := fmt.Sprintf("%v/%v", Config.TsFilePath, strings.ToLower(fmt.Sprintf("%v.ts", sms.ProtoName)))
	err = ioutil.WriteFile(opath, outputHelper.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Println("genTsDataMgr WriteFile error:", err)
		return
	}
}

func geneChhDataMgrHeader(xlsxs map[string]*SheetMetaStruct) {
	dm := map[string]interface{}{
		"data": xlsxs,
	}
	outputHelper := bytes.NewBuffer(nil)
	err := templates.ExecuteTemplate(outputHelper, "gpb_mgr_h", dm)
	if err != nil {
		fmt.Println("geneProto ExecuteTemplate error:", err)
		return
	}

	helperChhFile := filepath.Join(Config.WorkPath, Config.CppFilePath, "DataMgrAgc.h")
	err = ioutil.WriteFile(helperChhFile, outputHelper.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Println("geneProto WriteFile error:", err)
		return
	}
}

func geneCppDataMgrHeader(xlsxs map[string]*SheetMetaStruct) {
	dm := map[string]interface{}{
		"data": xlsxs,
	}
	outputHelper := bytes.NewBuffer(nil)
	err := templates.ExecuteTemplate(outputHelper, "gpb_mgr_c", dm)
	if err != nil {
		fmt.Println("geneProto ExecuteTemplate error:", err)
		return
	}

	helperChhFile := filepath.Join(Config.WorkPath, Config.CppFilePath, "DataMgrAgc.cpp")
	err = ioutil.WriteFile(helperChhFile, outputHelper.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Println("geneProto WriteFile error:", err)
		return
	}
}

func copyNoDataProtoToLua() {
	protoFile := filepath.Join(Config.WorkPath, Config.ProtoPath)
	rootdir, pathopen := os.Open(protoFile)
	if pathopen != nil {
		fmt.Println("Path open error.")
	} else {
		files, _ := rootdir.Readdir(0)
		for _, fi := range files {
			if strings.HasSuffix(fi.Name(), ".proto") && !strings.Contains(fi.Name(), "pbdata.proto") {
				datas, err := ioutil.ReadFile(filepath.Join(Config.WorkPath, Config.ProtoPath, fi.Name()))
				if err == nil {
					strContent := string(datas[:])
					strContent = strings.Replace(strContent, `"protocol/`, `"`, -1)
					dest := filepath.Join(Config.WorkPath, Config.ProtoLuaPath, fi.Name())
					fmt.Println("==========dest", dest)
					ioutil.WriteFile(dest, []byte(strContent), os.ModePerm)
				} else {
					fmt.Println(err)
				}
			}

		}
	}
}
