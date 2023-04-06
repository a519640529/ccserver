// config
package main

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/idealeak/goserver/core"
)

var Config = Configuration{}
var templates *template.Template

type Configuration struct {
	WorkPath     string
	XlsxPath     string
	TsPath       string
	LanguageType []string
}

func (this *Configuration) Name() string {
	return "language"
}

func (this *Configuration) Init() (err error) {
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
