package main

import (
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"strconv"
	"time"
)

func ResisteFrontEndAPI() {
	//http://127.0.0.1:8080/api/report_filedown_fail?platform=7&package=com.qipai.xiaoxin.a001&filename=1.mp4&url=http://cdn.x.com/1.mp4?ver=1001&statuscode=404&errorcode=2&errormsg=xxx
	GinEngine.GET("/api/report_filedown_fail", func(c *gin.Context) {
		pkg := c.DefaultQuery("package", "com.x.qipai")
		filename := c.Query("filename")
		url := c.Query("url")
		md5 := c.Query("md5")
		errormsg := c.Query("errormsg")
		statuscode := c.Query("statuscode")
		errorcode := c.Query("errorcode")
		_size := c.Query("size")
		size, _ := strconv.Atoi(_size)
		tNow := time.Now()
		ip := c.ClientIP()
		fdfdd := &FileDownLoadFailDetailData{
			Id:         bson.NewObjectId(),
			FileName:   filename,
			Url:        url,
			StatusCode: statuscode,
			ErrorCode:  errorcode,
			ErrorMsg:   errormsg,
			Size:       int64(size),
			Ip:         ip,
			Package:    pkg,
			Md5:        md5,
			Time:       tNow,
		}
		WriteFileDownLog(fdfdd)

		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}
