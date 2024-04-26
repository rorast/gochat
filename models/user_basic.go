package models

import (
	"fmt"
	"gochat/tools"
	"gorm.io/gorm"
	"log"
	"time"
)

type UserBasic struct {
	gorm.Model
	Name          string
	PassWord      string
	Phone         string `valid:"matches(^0[3-9]{1}\\d{8}$)"` // 大陸門號為matches(^1[3-9]{1}\\d{9}$)
	Email         string `valid:"email"`
	Avatar        string //頭像
	Identity      string
	ClientIp      string
	ClientPort    string
	Salt          string
	LoginTime     time.Time
	HeartbeatTime time.Time
	LoginOutTime  time.Time `gorm:"column:login_out_time" json:"login_out_time"`
	IsLogout      bool
	DeviceInfo    string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func GetUserList() []*UserBasic {
	//data := make([]*UserBasic, 10)
	//DB.Find(&data)
	//for _, v := range data {
	//	fmt.Println(v)
	//}
	// bilibili 教程的寫法，多個 Error
	data := make([]*UserBasic, 0)
	err := DB.Find(&data).Error
	if err != nil {
		log.Println("gorm Init Error : ", err)
	}
	for _, v := range data {
		fmt.Println("User ====> %v \n", v)
	}

	return data
}

func FindUserByName(name string) UserBasic {
	user := UserBasic{}
	DB.Where("name = ?", name).First(&user)
	return user
}

func CreateUser(user UserBasic) *gorm.DB {
	return DB.Create(&user)
}

// 刪除用戶資料
func DeleteUser(user UserBasic) *gorm.DB {
	return DB.Delete(&user)
}

// 更新用戶資料
func UpdateUser(user UserBasic) *gorm.DB {
	return DB.Model(&user).Updates(UserBasic{Name: user.Name, PassWord: user.PassWord, Phone: user.Phone, Email: user.Email, Avatar: user.Avatar, Salt: user.Salt})
}

func FindUserByNameAndPwd(name string, password string) UserBasic {
	user := UserBasic{}
	DB.Where("name = ? and pass_word=?", name, password).First(&user)

	// token 加密
	str := fmt.Sprintf("%d", time.Now().Unix())
	temp := tools.MD5Encode(str)
	DB.Model(&user).Where("id = ?", user.ID).Update("identity", temp)
	return user
}
