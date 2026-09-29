package model

import (
	"time"

	"gorm.io/gorm"
)

type Activity struct {
	gorm.Model
	ActivityName string
	StartTime    *time.Time
	EndTime      *time.Time
	Status       int64 //1未开始 2进行中 3已结束
}

type ActivityGood struct {
	gorm.Model
	ActivityId    int64
	GoodName      string
	ActivityPrice float64
	ActivityStock int64
	Limit         int64 //限购数量
	Status        int64 //1上架 2下架
}

type Cart struct {
	UserId         int64
	ActivityId     int64
	ActivityGoodId int64
	BuyCount       int64
	TotalPrice     float64
}
