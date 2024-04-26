package main

import (
	"gochat/models"
	"gochat/router"
)

func main() {
	// 讀入初始化設定內容
	models.InitConfig()
	// 讀入 MySQL 資料庫設定內容
	models.InitMySQL()
	// 讀入 Redis 緩存資料庫設定內容
	models.InitRedis()

	r := router.Router()
	r.Run(":8080")
}
