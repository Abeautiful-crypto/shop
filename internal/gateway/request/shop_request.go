package request

// CreateShopRequest 是 POST /api/v1/shops 使用的 JSON 载荷。
type CreateShopRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateShopRequest 是 PUT /api/v1/shops/:id 使用的 JSON 载荷。
type UpdateShopRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// ListShopRequest 是 GET /api/v1/shops 使用的查询参数载荷。
type ListShopRequest struct {
	Page     int32 `form:"page"`
	PageSize int32 `form:"page_size"`
}

type ActivityOrder struct {
	UserId         int64 `form:"userId" json:"userId" xml:"userId"  binding:"required"`
	ActivityId     int64 `form:"activityId" json:"activityId" xml:"activityId"  binding:"required"`
	ActivityGoodId int64 `form:"activityGoodId" json:"activityGoodId" xml:"activityGoodId"  binding:"required"`
	BuyCount       int64 `form:"buyCount" json:"buyCount" xml:"buyCount"  binding:"required"`
}
