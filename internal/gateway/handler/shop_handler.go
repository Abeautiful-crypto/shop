package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	shopv1 "github.com/BwCloudWeGo/bw-cli/api/gen/shop/v1"
	"github.com/BwCloudWeGo/bw-cli/internal/gateway/request"
	apperrors "github.com/BwCloudWeGo/bw-cli/pkg/errors"
	"github.com/BwCloudWeGo/bw-cli/pkg/httpx"
)

// ShopHandler 将 shop HTTP 接口适配到生成的 gRPC client。
type ShopHandler struct {
	client shopv1.ShopServiceClient
	log    *zap.Logger
}

// NewShopHandler 将 shop gRPC client 注入 HTTP handler 方法。
func NewShopHandler(client shopv1.ShopServiceClient, log *zap.Logger) *ShopHandler {
	if log == nil {
		log = zap.NewNop()
	}
	return &ShopHandler{
		client: client,
		log:    log,
	}
}

// Create 将 POST /api/v1/shops 代理到 CreateShop。
func (h *ShopHandler) Create(c *gin.Context) {
	var req request.CreateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, apperrors.InvalidArgument("invalid_request", err.Error()))
		return
	}
	resp, err := h.client.CreateShop(outgoingContext(c), &shopv1.CreateShopRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		httpx.Error(c, apperrors.FromGRPC(err))
		return
	}
	h.log.Info("gateway shop create proxied", zap.String("request_id", httpx.RequestID(c)), zap.String("aggregate_id", resp.GetId()))
	httpx.Created(c, resp)
}

// Get 将 GET /api/v1/shops/:id 代理到 GetShop。
func (h *ShopHandler) Get(c *gin.Context) {
	resp, err := h.client.GetShop(outgoingContext(c), &shopv1.GetShopRequest{Id: c.Param("id")})
	if err != nil {
		httpx.Error(c, apperrors.FromGRPC(err))
		return
	}
	httpx.OK(c, resp)
}

// List 将 GET /api/v1/shops 代理到 ListShops。
func (h *ShopHandler) List(c *gin.Context) {
	var req request.ListShopRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httpx.Error(c, apperrors.InvalidArgument("invalid_request", err.Error()))
		return
	}
	resp, err := h.client.ListShops(outgoingContext(c), &shopv1.ListShopsRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		httpx.Error(c, apperrors.FromGRPC(err))
		return
	}
	httpx.OK(c, resp)
}

// Update 将 PUT /api/v1/shops/:id 代理到 UpdateShop。
func (h *ShopHandler) Update(c *gin.Context) {
	var req request.UpdateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, apperrors.InvalidArgument("invalid_request", err.Error()))
		return
	}
	resp, err := h.client.UpdateShop(outgoingContext(c), &shopv1.UpdateShopRequest{
		Id:          c.Param("id"),
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		httpx.Error(c, apperrors.FromGRPC(err))
		return
	}
	h.log.Info("gateway shop update proxied", zap.String("request_id", httpx.RequestID(c)), zap.String("aggregate_id", resp.GetId()))
	httpx.OK(c, resp)
}

// Delete 将 DELETE /api/v1/shops/:id 代理到 DeleteShop。
func (h *ShopHandler) Delete(c *gin.Context) {
	resp, err := h.client.DeleteShop(outgoingContext(c), &shopv1.DeleteShopRequest{Id: c.Param("id")})
	if err != nil {
		httpx.Error(c, apperrors.FromGRPC(err))
		return
	}
	h.log.Info("gateway shop delete proxied", zap.String("request_id", httpx.RequestID(c)), zap.String("aggregate_id", c.Param("id")))
	httpx.OK(c, resp)
}

func (h *ShopHandler) ActivityOrder(c *gin.Context) {
	var req request.ActivityOrder
	if err := c.ShouldBind(&req); err != nil {
		httpx.Error(c, apperrors.InvalidArgument("invalid_request", err.Error()))
		return
	}

	order, err := h.client.ActivityOrder(c, &shopv1.ActivityOrderRequest{
		UserId:         req.UserId,
		ActivityId:     req.ActivityId,
		ActivityGoodId: req.ActivityGoodId,
		BuyCount:       req.BuyCount,
	})
	if err != nil {
		httpx.Error(c, apperrors.FromGRPC(err))
		return
	}
	h.log.Info("gateway shop delete proxied", zap.String("request_id", httpx.RequestID(c)), zap.String("aggregate_id", c.Param("id")))
	httpx.OK(c, order)

}
