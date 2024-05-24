package service

import "C"
import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gochat/models"
	"gochat/tools"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// GetUserList
// @Summary 所有用戶
// @Tags 用戶模組
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data := make([]*models.UserBasic, 0)
	data = models.GetUserList()
	c.JSON(200, gin.H{
		"code":    0, // 0成功 -1失敗
		"message": "列出用戶列表!",
		"data":    data,
	})
}

// CreateUser
// @Summary 新增用戶
// @Tags 用戶模組
// @param name query string false "用戶名"
// @param password query string false "密碼"
// @param repassword query string false "確認密碼"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [post]
func CreateUser(c *gin.Context) {
	// 建立用戶實體結構
	user := models.UserBasic{}
	// 單後端測試使用 Postman 以 param 測試
	//user.Name = c.Query("name")
	//password := c.Query("password")
	//repassword := c.Query("repassword")
	// 前後端分離，供 vue 使用
	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("repassword")

	data := models.FindUserByName(user.Name)
	// 檢查帳密是否為空
	if user.Name == "" || password == "" || repassword == "" {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失敗
			"message": "用戶名或密碼不能為空！",
			"data":    user,
		})
		return
	}
	// 檢查是否已有帳號
	if data.Name != "" {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失敗
			"message": "用戶名已被注冊！",
			"data":    user,
		})
		return
	}
	// 檢查2次密碼是否為一致
	if password != repassword {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失敗
			"message": "2 次密碼不一致！",
			"data":    user,
		})
		return
	}
	//user.PassWord = password
	salt := fmt.Sprintf(viper.GetString("key.salt"), rand.Int31())
	user.PassWord = tools.MakePassword(password, salt)
	user.Salt = salt
	user.LoginTime = time.Now()
	user.LoginOutTime = time.Now()
	user.HeartbeatTime = time.Now()
	models.CreateUser(user)
	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失敗
		"message": "新增用户成功！",
		"data":    user,
	})
}

// DeleteUser
// @Summary 刪除用戶
// @Tags 用戶模組
// @param id query string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [post]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(id)
	models.DeleteUser(user)
	c.JSON(200, gin.H{
		"code":    0, // 0成功 -1失敗
		"message": "刪除用戶成功",
		"data":    user,
	})
}

// UpdateUser
// @Summary 修改用戶
// @Tags 用戶模組
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code","message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	// 密碼加密
	password := c.PostForm("password")
	salt := fmt.Sprintf(viper.GetString("key.salt"), rand.Int31()) // 重新產生加密字串
	user.Salt = salt                                               // 重新儲存加密字串
	user.PassWord = tools.MakePassword(password, salt)
	//user.PassWord = c.PostForm("password") 未加密
	user.Phone = c.PostForm("phone")
	user.Avatar = c.PostForm("icon")
	user.Email = c.PostForm("email")
	fmt.Println("update :", user)

	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "修改参数不匹配！",
			"data":    user,
		})
	} else {
		models.UpdateUser(user)
		c.JSON(200, gin.H{
			"code":    0, //  0成功   -1失败
			"message": "修改用户成功！",
			"data":    user,
		})
	}

}

// FindUserByNameAndPwd
// @Summary 經由名及密碼查找用戶
// @Tags 用戶模塊
// @param name query string false "用戶名"
// @param password query string false "密碼"
// @Success 200 {string} json{"code","message"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}
	// 單後端測試使用 Postman 以 param 測試
	//name := c.Query("name")
	//password := c.Query("password")
	// 前後端分離，供 vue 使用
	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	fmt.Println(name, password)
	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失敗
			"message": "該用戶不存在",
			"data":    data,
		})
		return
	}

	flag := tools.ValidPassword(password, user.Salt, user.PassWord)
	if !flag {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失敗
			"message": "密碼不正確",
			"data":    data,
		})
		return
	}
	pwd := tools.MakePassword(password, user.Salt)
	data = models.FindUserByNameAndPwd(name, pwd)

	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失敗
		"message": "登錄成功",
		"data":    data,
	})
}

// 防止跟域站點偽造請求
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 聊天路由執行方法
func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

func SendMsg(c *gin.Context) {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)
	MsgHandler(c, ws)
}

func RedisMsg(c *gin.Context) {
	userIdA, _ := strconv.Atoi(c.PostForm("userIdA"))
	userIdB, _ := strconv.Atoi(c.PostForm("userIdB"))
	start, _ := strconv.Atoi(c.PostForm("start"))
	end, _ := strconv.Atoi(c.PostForm("end"))
	isRev, _ := strconv.ParseBool(c.PostForm("isRev"))
	res := models.RedisMsg(int64(userIdA), int64(userIdB), int64(start), int64(end), isRev)
	tools.RespOKList(c.Writer, "ok", res)
}

func MsgHandler(c *gin.Context, ws *websocket.Conn) {
	for {
		// 訂閱 redis ，特別注意-- 若是redis設置連不上，會導致錯誤回圈產生
		msg, err := models.Subscribe(c, models.PublishKey)
		if err != nil {
			fmt.Println(" MsgHandler 发送失败", err)
		}

		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
		err = ws.WriteMessage(1, []byte(m))
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("|service|user_service|訊息 ===> ", msg)
	}
}

// 搜尋好友的邏輯處理
func SearchFriends(c *gin.Context) {
	id, _ := strconv.Atoi(c.Request.FormValue("userId"))
	fmt.Println("<func SearchFriends> ID =======> ", id)
	users := models.SearchFriend(uint(id))
	// c.JSON(200, gin.H{
	// 	"code":    0, //  0成功   -1失败
	// 	"message": "查询好友列表成功！",
	// 	"data":    users,
	// })
	tools.RespOKList(c.Writer, users, len(users))
}
