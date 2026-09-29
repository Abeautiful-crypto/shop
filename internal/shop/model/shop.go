package model

import "time"

// ShopModel 是 shops 表的 Gorm 持久化模型。
type ShopModel struct {
	ID          string `gorm:"primaryKey;size:64"`
	Name        string `gorm:"size:128;not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (ShopModel) TableName() string {
	return "shops"
}

const shopMongoCollectionName = "shops"

// ShopDocument 是 shops 集合的 MongoDB 文档模型。
type ShopDocument struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

func (ShopDocument) MongoCollectionName() string {
	return shopMongoCollectionName
}
