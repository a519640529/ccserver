package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"runtime"
)

var GinEngine *gin.Engine

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	err := LoadConfig()
	if err != nil {
		fmt.Printf("Fail to LoadConfig: %v", err)
		os.Exit(1)
	}

	// Force log's color
	gin.ForceConsoleColor()
	gin.SetMode(AppCfg.GinMode)
	GinEngine = gin.Default()
	GinEngine.Use(gin.Recovery())
	// Logging to a file.
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	MgoSession, err = newDBSession()
	if err != nil {
		fmt.Println("mgo session create err:", err)
		os.Exit(1)
	}

	//启动mongo定时ping
	StartMgoPing()
	//启动日志发布
	StartPublishLog()
	//注册前端api
	ResisteFrontEndAPI()
	//注册后端api
	RegisteBackEndAPI()

	GinEngine.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	GinEngine.Run(fmt.Sprintf(":%d", AppCfg.SC.Port))
}
