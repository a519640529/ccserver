package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Conf struct {
	SrcPath  string   `json:"srcPath"`
	DestPath string   `json:"destPath"`
	CmdPath  string   `json:cmdPath`
	MapCc    []string `json:mapCc`
}

// 判断路径是否存在
func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/*
传入参数 all 则转换 protocol目录下的所有文件
如果传入模块名字 则转换模块名字的 proto文件
*/
func main() {

	//args := os.Args
	//if len(args) < 1 {
	//	fmt.Println("[all | module name]")
	//	return
	//}

	//读取当前目录下的配置文件
	fileBuffer, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println("获取srcpath和destpath发生错误")
		return
	}
	var c Conf

	e := json.Unmarshal(fileBuffer, &c)
	if e != nil {
		fmt.Println("从配置文件中获取srcpath和destpath发生错误")
	}

	//fmt.Println("源路径为:%s", c.SrcPath)
	//fmt.Println("目标路径为:%s", c.DestPath)

	//判定源路径是否存在
	srcstat, _ := pathExist(c.SrcPath)
	if !srcstat {
		fmt.Println("源路径不存在，无法执行转换")
	}
	deststate, _ := pathExist(c.DestPath)
	if !deststate {
		fmt.Println("目标路径不存在，请修改配置文件")
	}
	var mapCc = make(map[string]int)
	if len(c.MapCc) > 0 {
		for _, v := range c.MapCc {
			mapCc[v] = 0
		}
	}
	moduleName := "all"
	if moduleName == "all" {
		//遍历目录 所有目录都执行转换
		fileInfo, dirErr := ioutil.ReadDir(c.SrcPath)
		if dirErr != nil {
			fmt.Println("遍历源目录出现错误")
		}
		for _, v := range fileInfo {
			if _, ok := mapCc[v.Name()]; ok && v.IsDir() {
				doConvert(c.CmdPath, c.DestPath, c.SrcPath, v.Name())
			}
		}
	} else {
		//转换单个模块
		doConvert(c.CmdPath, c.DestPath, c.SrcPath, moduleName)
	}

}

func doConvert(cmdPath string, dPath string, sPath string, name string) {
	destPath := path.Join(dPath, name)
	srcPath := path.Join(sPath, name)
	if isExist, _ := pathExist(destPath); isExist {
		os.RemoveAll(destPath)
	}
	os.Mkdir(destPath, os.ModePerm)

	if isExist, _ := pathExist(srcPath); !isExist {
		fmt.Println("要转换的模块不存在")
		return
	}
	//str := "%s/pbjs --target static-module --no-convert --no-delimited --wrap closure --no-beautify --no-create --force-number  --no-verify  -o %s/%s.js %s/*.proto"
	str := "%s/pbjs --dependency protobufjs/minimal.js --no-service --no-convert  --no-delimited --no-create --no-beautify  --no-verify  --target static-module --wrap commonjs --out %s/%s.js %s/*.proto"
	cmdStr := fmt.Sprintf(str, cmdPath, destPath, name, srcPath)
	//fmt.Println("转换为js")
	//fmt.Println(cmdStr)

	cmd := exec.Command("cmd.exe", "/C", cmdStr)

	if err := cmd.Run(); err != nil {
		fmt.Printf("转换%s proto to js 错误", name)
	}

	//这个地方需要将js中内容以字符串的形式读取出来，然后
	coverJsFile(destPath, name)

	//fmt.Println("转化ts")
	//str = "%s/pbts --no-comments --main -o  %s/%s.d.ts %s/%s.js"
	str = "%s/pbts.cmd --main --out %s/" + name + ".d.ts %s/" + name + ".js"
	cmdStr = fmt.Sprintf(str, cmdPath, destPath, destPath)
	cmd = exec.Command("cmd.exe", "/C", cmdStr)
	//
	if err := cmd.Run(); err != nil {
		fmt.Println("转换 js to ts 错误", name)
	}

	convrTsFile(destPath, name)

	//重新生成js。为了生成ts则必须有注释。在生成ts后重新生成一下js去掉注释
	//str = "%s/pbjs --target static-module --no-convert --no-delimited --no-comments --wrap closure --no-beautify --no-create --force-number  --no-verify  -o %s/%s.js %s/*.proto"
	str = "%s/pbjs --dependency protobufjs/minimal.js --no-service --no-convert --no-comments --no-delimited --no-create --no-beautify  --no-verify  --target static-module --wrap commonjs --out %s/%s.js %s/*.proto"
	cmdStr = fmt.Sprintf(str, cmdPath, destPath, name, srcPath)
	//fmt.Println("重新转换为js")
	//fmt.Println(cmdStr)

	cmd = exec.Command("cmd.exe", "/C", cmdStr)

	if err := cmd.Run(); err != nil {
		fmt.Println("重新去掉注释转换%s proto to js 错误", name)
	}

	//copy *.proto
	str = "copy %s/*.proto %s/"
	cmdStr = fmt.Sprintf(str, srcPath, destPath)
	cmd = exec.Command("powershell.exe", "/C", cmdStr)
	if err := cmd.Run(); err != nil {
		fmt.Println("copy err:", err)
	}

	//这个地方需要将js中内容以字符串的形式读取出来，然后
	coverJsFile(destPath, name)

	fmt.Printf("cc	%-15s to js 完成 \n", name)
}

// 覆盖filePath对应的文本文件
func coverJsFile(filePath string, name string) {
	by, err := ioutil.ReadFile(filePath + "/" + name + ".js")
	if err != nil {
		fmt.Printf("覆盖文件%s出现错误/n", name)
		return
	}

	s := string(by)

	startStr := `(function($protobuf)`

	//targetStr := fmt.Sprintf("(window || global).%s = (function($protobuf)", name)
	targetStr := fmt.Sprintf("module.exports.%s=(function($protobuf)", name)
	s = strings.Replace(s, startStr, targetStr, 1)

	endStr := `(protobuf)`
	endTargetStr := fmt.Sprintf("%s.%s", "(protobuf)", name)
	s = strings.Replace(s, endStr, endTargetStr, 1)
	f, err := os.OpenFile(filePath+"/"+name+".js", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("覆盖写文件出现错误")
		return
	}
	defer f.Close()
	f.WriteString(s)
}

func convrTsFile(filePath string, name string) {
	by, err := ioutil.ReadFile(filePath + "/" + name + ".d.ts")
	if err != nil {
		fmt.Println("覆盖TS文件%s错误", name)
		return
	}
	s := string(by)
	//temple := `declare global {
	//	%s
	//}
	//export {}`
	//temple = fmt.Sprintf(temple, s)

	temple := s

	f, err := os.OpenFile(filePath+"/"+name+".d.ts", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("覆盖TS文件发生错误%s", name)
		return
	}
	defer f.Close()
	f.WriteString(temple)
}
