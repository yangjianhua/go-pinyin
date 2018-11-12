package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-pinyin"
)

var a pinyin.Args

func initPinyinArgs(arg int) { // arg should be pinyin.Tone, pinyin.Tone1, pinyin.Tone2, pinyin.Tone3, see go-pinyin doc
	a = pinyin.NewArgs()
	a.Style = arg
}

func getPinyin(c *gin.Context) {
	han := c.DefaultQuery("han", "")
	p := pinyin.Pinyin(han, a)

	c.JSON(200, gin.H{"code": 0, "data": p})
}

func getPinyinOne(c *gin.Context) {
	han := c.DefaultQuery("han", "")
	p := pinyin.Pinyin(han, a)
	s := ""

	if len(p) > 0 {
		s = p[0][0]
	}

	c.JSON(200, gin.H{"code": 0, "data": s})
}

func allowCors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	// init pinyin output format
	initPinyinArgs(pinyin.Tone)

	fmt.Print("\n\nDEFAULT PORT: 8080, USING '-port portnum' TO START ANOTHER PORT.\n\n")

	port := flag.Int("port", 8080, "Port Number, default 8080")
	flag.Parse()
	sPort := ":" + strconv.Itoa(*port)

	// using gin as a web output
	r := gin.Default()
	r.Use(allowCors())
	r.GET("/pinyin", getPinyin) // Call like GET http://localhost:8080/pinyin?han=我来了
	r.GET("/pinyin1", getPinyinOne)
	r.Run(sPort)
}
