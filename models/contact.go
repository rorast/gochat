package models

import "gorm.io/gorm"

// 人員關係
type Contact struct {
	gorm.Model
	OwnerId  uint // 誰的關係信息
	TargetId uint // 對應的誰 /群 ID
	Type     int  // 對應的類型  1好友  2群  3xx
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}

// 搜尋好友
func SearchFriend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	DB.Where("owner_id = ? and type=1", userId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0)
	DB.Where("id in ?", objIds).Find(&users)
	return users
}
