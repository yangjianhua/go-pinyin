# go-pinyin - Mac Word下的自动拼音实现方法

This Tools is using for those who want to make a Chinese Pinyin Book, so the following description is only in Chinese.

## 用到的工具

Mac Word - 可运行VBA
Mac脚本编辑器
本代码的执行程序

## 问题所在

Word提供了为文字添加拼音的功能，这个功能在Word的**格式-中文版式-拼音指南**可以找到，或者在工具栏区的**拼音指南**按钮也同样可以做到。Word会根据分词情况进行拼音标注，比如在“我的”和“的确”中，多音字“的”会被正确对待标注。但存在一个问题，就是**每次最多只能处理30个字符**，如果有更多的字符的话，也只处理前30个。这导致的问题就是，如果一篇上千字的文章，需要反复操作多次，才能完成一篇文章的拼音标注工作。

Mac下的Pages也有拼音标注的功能，但功能使用起来比Word还要差，可用性不强，于是不采用Pages进行排版。

## 解决思路

首先是视图使用VBA的拼音处理功能，即录制一段宏对文字进行拼音化处理。但感觉VBA下没有特别好的文本转拼音的脚本，于是舍弃了单纯用VBA自动设置拼音的方法，而采用一个Golang的执行程序来进行拼音转换操作。

确定方式方法之后，确定需要以下的步骤完成对文本的拼音标注。

Word VBA            - 选中文本，并且将文本发送到接口，由接口返回拼音
Mac Script          - 由于在Mac Word中无法直接创建XMLHTTP对象，因此无法直接调用网络接口，需要通过Mac Script的 JSON Helper完成(如未安装JSON Helper需要在AppStore安装)
Golang Executable   - 提供可直接运行的程序，通过gin提供API接口

## Mac Script

```
on getPinyin(han)
	set curlURL to "http://localhost:8001/pinyin1?han=" & han
	set pinyin to ""

	tell application "JSON Helper"
		set json to fetch JSON from curlURL
		set pinyin to |data| of json
	end tell

	return pinyin
end getPinyin

getPinyin("好")
```

注意：由于Word的权限限制，该Mac Script需要放置在指定的目录中才可以执行，这里是：
/Users/{MacUserName}/Library/Application Scripts/com.microsoft.Word

## VBA Script

``` VB
Sub AutoPinyin()
    With Selection
        Dim nStart As Long
        Dim nEnd As Long
        nStart = .Start
        nEnd = .End

        Dim ret As String

        While nStart < nEnd
            .Start = nStart
            .End = nStart + 1
            ret = AppleScriptTask("getPinyin.scpt", "getPinyin", .Range)
            If ret <> "" Then
                .Range.PhoneticGuide Text:=ret, Alignment:=wdPhoneticGuideAlignmentCenter, Raise:=18, FontSize:=9, FontName:= _
                "微软雅黑"

                nStart = nStart + 1 + Len(ret) + 51
                nEnd = nEnd + Len(ret) + 51
            Else
                nStart = nStart + 1
            End If

            ActiveWindow.ScrollIntoView .Range, True

        Wend
    End With
End Sub

Sub DetermineLength()
    With Selection
        MsgBox .Start
        MsgBox .End
    End With
End Sub

```
这里需要注意的是：当给汉字添加完拼音之后，整个文档的篇幅已经自动增长了，被加拼音的汉字实际上是一个域，可以通过Word的**查看域代码**进行查看，这个长度并不是简单的加几个字符，而是比较多域代码。所以在这段代码里，字符长度额外增加51（这个51未深究，可能不同的字体、字号会有不同，这个可以在增加前后通过**DeterminLength**自行试验一下。

## Golang File

``` go
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
```

Golang的代码其实很少，多亏有[go-pinyin](github.com/mozillazg/go-pinyin)和[gin](github.com/gin-gonic/gin)才可以使代码如此精简，gin使用了allowCors方法解决了跨域访问的问题。

## 缺点

比较明显的缺点是对多音字的处理不如Word原来的**拼音指南**，拼音指南会对多音字进行分词和处理，但通过这个脚本完全不考虑多音字了，因此就需要在处理完成之后自行校对和手工修改。

另一个问题是Word VBA的执行效率比较低，运行起来比较慢，容易造成类假死的情况。建议不要一下子完成一本书，可以一次性200~400字左右，标注完当前选中的一段之后，再标注下一段。

总的来说效率提升非常明显，原来大约30页左右的一本书，如果使用拼音指南的话，由于需要反复用鼠标和键盘操作，比较辛苦；大概拖拖拉拉的需要一个星期左右；现在大约2~3个小时就可以自动加完拼音和手工校对的工作了，总的来说还是解决了本人的大问题。

