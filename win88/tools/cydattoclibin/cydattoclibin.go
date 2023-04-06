package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type Conf struct {
	SrcPath  string `json:"srcPath"`
	DestPath string `json:"destPath"`
}

//判断路径是否存在
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

func main() {
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

	fmt.Println("源路径为:%s", c.SrcPath)
	fmt.Println("目标路径为:%s", c.DestPath)

	//判定源路径是否存在
	srcstat, _ := pathExist(c.SrcPath)
	if !srcstat {
		fmt.Println("源路径不存在，无法执行转换")
	}
	deststate, _ := pathExist(c.DestPath)
	if !deststate {
		fmt.Println("目标路径不存在，请修改配置文件")
	}

	copyDataToBin(c.SrcPath, c.DestPath)
}

//将 srcPath目录中的文件copy到 目标对应文件中 并且要修改后缀名
func copyDataToBin(srcPath string, targetPath string) {
	rd, err := ioutil.ReadDir(srcPath)
	if err != nil {
		fmt.Println("遍历文件是发生错误")
		return
	}
	for _, fi := range rd {
		if fi.IsDir() {
			//copyDataToBin(srcPath + fi.Name(),targetPath + fi.Name())
			copyDataToBin(path.Join(srcPath, fi.Name()), path.Join(targetPath, fi.Name()))
		} else {

			fullPath := path.Join(srcPath, fi.Name())
			ext := path.Ext(fullPath)
			//判定文件的后缀名为.bat则copy
			if ext == ".dat" {
				createTargetDir(targetPath)

				//判定目标文件是否已经存在。如果存在删除
				//fulleTargetPath := targetPath + fi.Name()
				strArr := strings.Split(fi.Name(), ".")
				baseName := strArr[0]
				tFileName := path.Join(targetPath, baseName) + ".bin"
				targetState, _ := pathExist(tFileName)
				if targetState {
					os.Remove(tFileName)
				}
				copyFile(fullPath, tFileName)
			}
		}
	}
}

func copyFile(srcFile string, targetFile string) {
	src, err := os.Open(srcFile)
	if err != nil {
		fmt.Println("Open file err = %v/n", err)
		return
	}
	defer src.Close()

	dst, err := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println("Open file err = %c/n", err)
		return
	}
	defer dst.Close()
	io.Copy(dst, src)
}

//创建目录
func createTargetDir(destDir string) {
	dirState, _ := pathExist(destDir)
	if !dirState {
		os.Mkdir(destDir, os.ModePerm)
	}
}
