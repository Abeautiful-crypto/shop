package handler

import (
	"context"
	stderrors "errors"

	"go.uber.org/zap"

	shopv1 "github.com/BwCloudWeGo/bw-cli/api/gen/shop/v1"
	"github.com/BwCloudWeGo/bw-cli/internal/shop/dto"
	"github.com/BwCloudWeGo/bw-cli/internal/shop/entity"
	"github.com/BwCloudWeGo/bw-cli/internal/shop/service"
	apperrors "github.com/BwCloudWeGo/bw-cli/pkg/errors"
)

// Server 将 shop gRPC 请求适配到 service 用例。
type Server struct {
	shopv1.UnimplementedShopServiceServer
	svc *service.Service
	log *zap.Logger
}

// NewServer 创建 shop gRPC 服务端适配器。
func NewServer(svc *service.Service, log *zap.Logger) *Server {
	if log == nil {
		log = zap.NewNop()
	}
	return &Server{svc: svc, log: log}
}

// CreateShop 处理创建 RPC。
func (s *Server) CreateShop(ctx context.Context, req *shopv1.CreateShopRequest) (*shopv1.ShopResponse, error) {
	item, err := s.svc.Create(ctx, dto.CreateCommand{
		Name:        req.GetName(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, mapShopError(err)
	}
	return toProto(item), nil
}

// GetShop 处理按 ID 查询。
func (s *Server) GetShop(ctx context.Context, req *shopv1.GetShopRequest) (*shopv1.ShopResponse, error) {
	item, err := s.svc.Get(ctx, req.GetId())
	if err != nil {
		return nil, mapShopError(err)
	}
	return toProto(item), nil
}

// ListShops 处理分页列表查询。
func (s *Server) ListShops(ctx context.Context, req *shopv1.ListShopsRequest) (*shopv1.ListShopsResponse, error) {
	list, err := s.svc.List(ctx, dto.ListCommand{
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		return nil, mapShopError(err)
	}
	resp := &shopv1.ListShopsResponse{
		Items: make([]*shopv1.ShopResponse, 0, len(list.Items)),
		Total: list.Total,
	}
	for _, item := range list.Items {
		resp.Items = append(resp.Items, toProto(item))
	}
	return resp, nil
}

// UpdateShop 处理按 ID 更新。
func (s *Server) UpdateShop(ctx context.Context, req *shopv1.UpdateShopRequest) (*shopv1.ShopResponse, error) {
	item, err := s.svc.Update(ctx, dto.UpdateCommand{
		ID:          req.GetId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, mapShopError(err)
	}
	return toProto(item), nil
}

// DeleteShop 处理按 ID 删除。
func (s *Server) DeleteShop(ctx context.Context, req *shopv1.DeleteShopRequest) (*shopv1.DeleteShopResponse, error) {
	if err := s.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, mapShopError(err)
	}
	return &shopv1.DeleteShopResponse{Success: true}, nil
}

func toProto(item *dto.ShopDTO) *shopv1.ShopResponse {
	return &shopv1.ShopResponse{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func mapShopError(err error) error {
	switch {
	case stderrors.Is(err, entity.ErrInvalidShop):
		return apperrors.InvalidArgument("invalid_shop", "invalid shop input")
	case stderrors.Is(err, entity.ErrShopNotFound):
		return apperrors.NotFound("shop_not_found", "shop not found")
	default:
		return apperrors.Wrap(apperrors.KindInternal, "shop_service_error", "shop service error", err)
	}
}

func (s *Server) ActivityGoodAdd(ctx context.Context, req *shopv1.ActivityGoodAddRequest) (*shopv1.ActivityGoodAddResponse, error) {
	err := s.svc.ActivityGoodAdd(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shopv1.ActivityGoodAddResponse{Success: true}, nil
}

func (s *Server) ActivityOrder(ctx context.Context, req *shopv1.ActivityOrderRequest) (*shopv1.ActivityOrderResponse, error) {
	order, err := s.svc.ActivityOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shopv1.ActivityOrderResponse{
		TotalPrice: order.TotalPrice,
		Url:        order.Url,
		OrderNo:    order.OrderNo,
	}, nil
}

func (s *Server) CartAdd(ctx context.Context, req *shopv1.CartAddRequest) (*shopv1.CartAddResponse, error) {
	err := s.svc.CartAdd(ctx, req)
	if err != nil {
		return nil, err
	}
	return &shopv1.CartAddResponse{Success: true}, nil
}

func (s *Server) CartDel(ctx context.Context, req *shopv1.CartDelRequest) (*shopv1.CartDelResponse, error) {
	err := s.svc.CartDel(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shopv1.CartDelResponse{Success: true}, nil
}

func (s *Server) CartAcc(ctx context.Context, req *shopv1.CartAccRequest) (*shopv1.CartAccResponse, error) {
	err := s.svc.CartAcc(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shopv1.CartAccResponse{Success: true}, nil
}

func (s *Server) CartList(ctx context.Context, req *shopv1.CartListRequest) (*shopv1.CartListResponse, error) {
	list, err := s.svc.CartList(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shopv1.CartListResponse{List: list.List}, nil
}

func (s *Server) CartOrder(ctx context.Context, req *shopv1.CartOrderRequest) (*shopv1.CartOrderResponse, error) {
	o, err := s.svc.CartOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shopv1.CartOrderResponse{
		TotalPrice: o.TotalPrice,
		Url:        o.Url,
		OrderNo:    o.OrderNo,
	}, nil
}

func (s *Server) ActivityAdd(ctx context.Context, req *shopv1.ActivityAddRequest) (*shopv1.ActivityAddResponse, error) {
	err := s.svc.ActivityAdd(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shopv1.ActivityAddResponse{Success: true}, nil
}

func (s *Server) ActivityDel(ctx context.Context, req *shopv1.ActivityDelRequest) (*shopv1.ActivityDelResponse, error) {
	err := s.svc.ActivityDel(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shopv1.ActivityDelResponse{Success: true}, nil
}

func (s *Server) ActivityUpdate(ctx context.Context, req *shopv1.ActivityUpdateRequest) (*shopv1.ActivityUpdateResponse, error) {
	err := s.svc.ActivityUpdate(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shopv1.ActivityUpdateResponse{Success: true}, nil
}

func (s *Server) ActivityList(ctx context.Context, req *shopv1.ActivityListRequest) (*shopv1.ActivityListResponse, error) {
	list, err := s.svc.ActivityList(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shopv1.ActivityListResponse{List: list.List}, nil
}
