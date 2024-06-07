package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gochat/tools"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

func Upload(c *gin.Context) {
	w := c.Writer
	r := c.Request
	srcFile, head, err := r.FormFile("file")
	if err != nil {
		tools.RespFail(w, err.Error())
	}
	suffix := ".png"
	ofilName := head.Filename
	tem := strings.Split(ofilName, ".")
	if len(tem) > 1 {
		suffix = "." + tem[len(tem)-1]
	}

	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix)
	dstFile, err := os.Create("./asset/upload/" + fileName)
	if err != nil {
		tools.RespFail(w, err.Error())
	}
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		tools.RespFail(w, err.Error())
	}
	url := "./asset/upload/" + fileName
	tools.RespOK(w, url, "發送圖片成功")
}
