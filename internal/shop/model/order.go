package model

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderNo        string
	UserId         int64
	OrderType      int64 //1普通加购 2秒杀活动加购
	GoodId         int64 //商品id
	ActivityId     int64 //活动id
	ActivityGoodId int64 //活动商品id
	BuyCount       int64
	TotalPrice     float64
	PayTime        *time.Time
	EntTime        *time.Time
	Status         int64 //1待支付 2已支付 3已取消 4已退款
}
