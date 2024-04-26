package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gochat/models"
	"html/template"
	"strconv"
)

// GetIndex
// @Tags 首頁
// @Success 200 {string} welcome
// @Router /index [get]
func GetIndex(c *gin.Context) {
	ind, err := template.ParseFiles("views/index.html", "views/chat/head.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "index")
	//c.JSON(200, gin.H{
	//	"message": "welcome !! ",
	//})
}

// 注冊頁路由
func ToRegister(c *gin.Context) {
	ind, err := template.ParseFiles("views/user/register.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "register")
}

// views 目錄下的 chat 目錄中的所有 html 檔案
func ToChat(c *gin.Context) {
	ind, err := template.ParseFiles("views/chat/index.html",
		"views/chat/head.html",
		"views/chat/foot.html",
		"views/chat/tabmenu.html",
		"views/chat/concat.html",
		"views/chat/group.html",
		"views/chat/profile.html",
		"views/chat/createcom.html",
		"views/chat/userinfo.html",
		"views/chat/main.html")
	if err != nil {
		panic(err)
	}
	userId, _ := strconv.Atoi(c.Query("userId"))
	fmt.Println(" index userId ===> ", userId)
	token := c.Query("token")  // 取得登入後的驗証 token
	user := models.UserBasic{} // 建立用戶資料物件
	user.ID = uint(userId)
	user.Identity = token
	ind.Execute(c.Writer, user)
}

// 聊天路由執行方法
func Chat(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}
