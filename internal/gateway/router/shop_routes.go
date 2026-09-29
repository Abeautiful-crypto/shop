package router

import (
	"github.com/gin-gonic/gin"

	"github.com/BwCloudWeGo/bw-cli/internal/gateway/handler"
)

// registerShopRoutes 在独立业务文件中注册 /api/v1/shops 接口。
func registerShopRoutes(v1 *gin.RouterGroup, shopHandler *handler.ShopHandler) {
	routes := v1.Group("/shops")
	routes.POST("", shopHandler.Create)
	routes.GET("", shopHandler.List)
	routes.GET("/:id", shopHandler.Get)
	routes.PUT("/:id", shopHandler.Update)
	routes.DELETE("/:id", shopHandler.Delete)
	routes.POST("/activity/order", shopHandler.ActivityOrder)
}
