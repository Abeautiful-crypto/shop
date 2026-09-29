package model

import "gorm.io/gorm"

type Good struct {
	gorm.Model
	GoodName string
	Stock    int64
	Price    float64 //价格
	Cost     float64 //成本
}
