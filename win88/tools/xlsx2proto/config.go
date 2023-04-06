// config
package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/idealeak/goserver/core"
)

var XlsxFiles = make(map[string]string)

var Config = Configuration{}

type Configuration struct {
	WorkPath        string
	XlsxPath        string
	ProtoPath       string
	ProtoLuaPath    string
	GoFilePath      string
	TsFilePath      string
	ConvertToolPath string
	CppFilePath     string
	DataPath        string
	ProtoFile       string
}

func (this *Configuration) Name() string {
	return "proto"
}

func (this *Configuration) Init() (err error) {

	protoAbsPath := filepath.Join(this.WorkPath, this.ProtoPath)
	err = os.MkdirAll(protoAbsPath, os.ModePerm)
	if err != nil {
		return
	}

	xlsxAbsPath := filepath.Join(this.WorkPath, this.XlsxPath)
	var fis []os.FileInfo
	fis, err = ioutil.ReadDir(xlsxAbsPath)
	if err != nil {
		return
	}

	for _, v := range fis {
		if !v.IsDir() {
			if !strings.HasSuffix(v.Name(), ".xlsx") {
				continue
			}

			pfAbs := filepath.Join(xlsxAbsPath, v.Name())
			XlsxFiles[v.Name()] = pfAbs
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return
	}

	pattern := filepath.Join(wd, "templ", "*.templ")
	funcMap := template.FuncMap{
		"inc": func(n int) int {
			n++
			return n
		},
	}
	templates, err = template.Must(template.New("mytempl").Funcs(funcMap).ParseGlob(pattern)).Parse("")

	return
}

func (this *Configuration) Close() (err error) {
	return
}

func init() {
	core.RegistePackage(&Config)
}
