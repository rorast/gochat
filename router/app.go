package router

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gochat/docs"
	"gochat/service"
)

func Router() *gin.Engine {
	r := gin.Default()
	//swagger 加入以下2行
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 給gin設置一個信任的ip
	//r.SetTrustedProxies([]string{"127.0.0.1"})

	// 靜態資源
	r.Static("/asset", "asset/")
	r.StaticFile("/favicon.ico", "asset/images/favicon.ico")
	// r.StaticFS() 靜態目錄下全部
	r.LoadHTMLGlob("views/**/*")

	// 首頁 在 index.go 中的路徑
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/toRegister", service.ToRegister)
	r.GET("/toChat", service.ToChat)
	r.GET("/chat", service.Chat)
	r.POST("/searchFriends", service.SearchFriends)

	//用戶模組 在 user_service.go 中的路徑
	r.POST("/user/getUserList", service.GetUserList)
	r.POST("/user/createUser", service.CreateUser)
	r.POST("/user/deleteUser", service.DeleteUser)
	r.POST("/user/updateUser", service.UpdateUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)

	// 發送消息
	r.GET("/user/sendMsg", service.SendMsg)
	//發送消息
	r.GET("/user/sendUserMsg", service.SendUserMsg)

	r.POST("/user/redisMsg", service.RedisMsg)

	return r
}
