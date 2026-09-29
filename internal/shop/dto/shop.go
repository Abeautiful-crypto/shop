package dto

import (
	"time"

	"github.com/BwCloudWeGo/bw-cli/internal/shop/entity"
)

// ShopDTO 由用例返回，并由 handler 转换。
type ShopDTO struct {
	ID          string
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
}

// ListShopDTO 包含分页列表出参。
type ListShopDTO struct {
	Items []*ShopDTO
	Total int64
}

// FromShop 将 shop 聚合转换为 service 响应 DTO。
func FromShop(item *entity.Shop) *ShopDTO {
	return &ShopDTO{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   formatTime(item.CreatedAt),
		UpdatedAt:   formatTime(item.UpdatedAt),
	}
}

// formatTime 让零值时间保持为空，并用稳定 API 格式序列化真实时间。
func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339Nano)
}
