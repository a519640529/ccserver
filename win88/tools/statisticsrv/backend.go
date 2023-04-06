package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strconv"
)

func apiTopNFileDownFile(c *gin.Context) {
	strPageNo := c.DefaultQuery("pageno", "1")
	strPageSize := c.DefaultQuery("pagesize", "50")
	pageNo, _ := strconv.Atoi(strPageNo)
	pageSize, _ := strconv.Atoi(strPageSize)
	datas, cnt := SelectTopNFileDownFailed(pageSize, pageNo)
	d, _ := json.MarshalIndent(datas, "", "\t")
	c.JSON(200, gin.H{
		"status": "ok",
		"pageno": pageNo,
		"count":  cnt,
		"data":   string(d),
	})
}

func apiFileDownFailDetail(c *gin.Context) {
	fileName := c.Query("filename")
	strPageNo := c.DefaultQuery("pageno", "1")
	strPageSize := c.DefaultQuery("pagesize", "50")
	pageNo, _ := strconv.Atoi(strPageNo)
	pageSize, _ := strconv.Atoi(strPageSize)
	datas, cnt, _ := SelectFileDownFailedDetail(fileName, pageSize, pageNo)
	d, _ := json.MarshalIndent(datas, "", "\t")
	c.JSON(200, gin.H{
		"status": "ok",
		"pageno": pageNo,
		"count":  cnt,
		"data":   string(d),
	})
}

func RegisteBackEndAPI() {
	//http://127.0.0.1:8080/api/topn_filedown_fail?pageno=1&pagesize=10
	GinEngine.Any("/api/topn_filedown_fail", apiTopNFileDownFile)

	//http://127.0.0.1:8080/api/filedown_fail_detail?filename=a.mp3&pageno=1&pagesize=10
	GinEngine.Any("/api/filedown_fail_detail", apiFileDownFailDetail)
}
