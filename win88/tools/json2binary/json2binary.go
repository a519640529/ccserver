package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/klauspost/compress/zip"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

type Conf struct {
	SrcPath string `json:"srcPath"`
	DesPath string `json:"destPath"`
	CmdPath string `json:"cmdPath"`
}


func pathExist(path string)(bool,error){
	_,err := os.Stat(path)
	if err == nil{
		return  true,nil
	}
	if os.IsNotExist(err){
		return false,nil
	}
	return  false,err
}


func main(){
	fileBuffer,err := ioutil.ReadFile("./config.json")
	if err != nil{
		fmt.Println("获取srcPath desPath失败")
		return
	}
	var c Conf
	e := json.Unmarshal(fileBuffer,&c)
	if e != nil{
		fmt.Println("从配置文件中获取srcPath desPath失败")
		return
	}
	fmt.Println("源路径为",c.SrcPath)
	fmt.Println("目标路径为",c.DesPath)

	srcstate,_ := pathExist(c.SrcPath)
	if !srcstate{
		fmt.Println("源路径不存在无法进行转换")
		return
	}
	desState,_ := pathExist(c.DesPath)
	if !desState{
		fmt.Println("目标路径，请修改配置文件")
		return
	}

	//直接将目录下的文件全部进行转换
	fileInfo,dirErr := ioutil.ReadDir(c.SrcPath)
	if dirErr != nil{
		fmt.Println("遍历源路径出现错误")
		return
	}
	for _,v := range fileInfo{
		name := v.Name()

		if path.Ext(name) ==".json"{
			//将json文件进行转换
			fileAllName := path.Base(name)
			fileSufix := path.Ext(name)
			filePrefix := fileAllName[0:len(fileAllName)-len(fileSufix)]
			cvt2Bin(c.SrcPath,c.DesPath,filePrefix,c.CmdPath)
		}
	}

}

//将单个文件转换
func cvt2Bin(src string,des string,name string,cmdPath string){
	srcPath := path.Join(src,name + ".json")
	desPath := path.Join(des,name + ".bin")
	fmt.Println("压缩文件开始",srcPath,desPath)

	//如果目标文件已经存在则删除
	if exist,_ :=pathExist(desPath);exist {
		os.Remove(desPath)
	}

	//f,err := os.Open(srcPath)
	//if err != nil{
	//	return
	//}
	//defer f.Close()
	//d,_ := os.Create(desPath)
	//defer d.Close()
	//w := zip.NewWriter(d)
	//defer w.Close()
	//err = compress(f,"",w)
	//if err != nil{
	//	fmt.Println(err)
	//	return
	//}

	//直接使用7z 进行压缩就可以
	//7z a -tzip  D:\VietnamGame\trunk\fishpath\json\custom1.bin   custom1.json
	cmdStr := "%s/7z.exe a -tzip %s %s"
	cmdStr = fmt.Sprintf(cmdStr,cmdPath, desPath, srcPath)
	cmd := exec.Command("cmd.exe", "/c", cmdStr)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	fmt.Println("Result: " + out.String())


	//cmd := exec.Command(cmdStr,"/c")
	//if err := cmd.Run(); err != nil {
	//	fmt.Println("压缩文件是失败", desPath)
	//	return
	//}
	fmt.Println("压缩文件到成功",desPath)
}


func compress(file * os.File,prefix string,zw * zip.Writer) error{
	info,err := file.Stat()
	if err != nil{
		return err
	}
	header,err := zip.FileInfoHeader(info)
	if len(prefix) ==0 {
		header.Name = header.Name
	}else {
		header.Name = prefix + "/" + header.Name
	}
	if err != nil{
		return err
	}
	writer,err := zw.CreateHeader(header)
	if err != nil{
		return  err
	}
	_,err = io.Copy(writer,file)
	file.Close()
	if err != nil{
		return  err
	}
	return  nil
}
