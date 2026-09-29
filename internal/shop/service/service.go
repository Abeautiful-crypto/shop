package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	shopv1 "github.com/BwCloudWeGo/bw-cli/api/gen/shop/v1"
	"github.com/BwCloudWeGo/bw-cli/internal/shop/dto"
	"github.com/BwCloudWeGo/bw-cli/internal/shop/entity"
	"github.com/BwCloudWeGo/bw-cli/internal/shop/model"
	"github.com/BwCloudWeGo/bw-cli/pkg/alipayx"
	"github.com/BwCloudWeGo/bw-cli/pkg/config"
	"github.com/BwCloudWeGo/bw-cli/pkg/redisx"
	"github.com/BwCloudWeGo/bw-cli/pkg/utils"
	"go.uber.org/zap"
)

// Service 编排 shop 用例。
type Service struct {
	repo   entity.Repository
	log    *zap.Logger
	locker *redisx.Locker
}

// NewService 创建 shop 用例服务。
func NewService(repo entity.Repository, log *zap.Logger, locker *redisx.Locker) *Service {
	if log == nil {
		log = zap.NewNop()
	}
	return &Service{repo: repo, log: log, locker: locker}
}

// Create 创建 shop 记录。
func (s *Service) Create(ctx context.Context, cmd dto.CreateCommand) (*dto.ShopDTO, error) {
	item, err := entity.NewShop(cmd.Name, cmd.Description)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return nil, err
	}
	s.log.Info("shop created", zap.String("aggregate_id", item.ID), zap.String("use_case", "CreateShop"))
	return dto.FromShop(item), nil
}

// Get 根据 ID 返回一条 shop 记录。
func (s *Service) Get(ctx context.Context, id string) (*dto.ShopDTO, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.FromShop(item), nil
}

// List 返回分页的 shop 记录。
func (s *Service) List(ctx context.Context, cmd dto.ListCommand) (*dto.ListShopDTO, error) {
	offset, limit := normalizePagination(cmd.Page, cmd.PageSize)
	items, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	output := &dto.ListShopDTO{Items: make([]*dto.ShopDTO, 0, len(items)), Total: total}
	for _, item := range items {
		output.Items = append(output.Items, dto.FromShop(item))
	}
	return output, nil
}

// Update 根据 ID 修改一条 shop 记录。
func (s *Service) Update(ctx context.Context, cmd dto.UpdateCommand) (*dto.ShopDTO, error) {
	item, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if err := item.Update(cmd.Name, cmd.Description); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return nil, err
	}
	s.log.Info("shop updated", zap.String("aggregate_id", cmd.ID), zap.String("use_case", "UpdateShop"))
	return dto.FromShop(item), nil
}

// Delete 根据 ID 删除一条 shop 记录。
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.log.Info("shop deleted", zap.String("aggregate_id", id), zap.String("use_case", "DeleteShop"))
	return nil
}

func normalizePagination(page int32, pageSize int32) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return int((page - 1) * pageSize), int(pageSize)
}

func (s *Service) ActivityGoodRedis(ctx context.Context) {
	goods, err := s.repo.ActivityGoodFind(ctx)
	if err != nil {
		return
	}
	res := redisx.NewClient(config.GlobalConfig.Redis)
	for _, good := range goods {
		key := fmt.Sprintf("activity_good_%d", good.ID)
		//marshal, err := json.Marshal(&good)
		//if err != nil {
		//	return
		//}
		err = res.HSet(ctx, key, map[string]interface{}{
			"id":             good.ID,
			"activity_id":    good.ActivityId,
			"good_name":      good.GoodName,
			"activity_price": good.ActivityPrice,
			"activity_stock": good.ActivityStock,
			"limit":          good.Limit,
			"status":         good.Status,
		}).Err()
		if err != nil {
			fmt.Println("活动商品表数据redis同步失败")
			return
		}
	}
	fmt.Println("活动商品表数据redis同步成功")
	return
}

func (s *Service) ActivityGoodAdd(ctx context.Context, req *shopv1.ActivityGoodAddRequest) error {
	activityGood := model.ActivityGood{
		ActivityId:    req.ActivityId,
		GoodName:      req.GoodName,
		ActivityPrice: float64(req.ActivityPrice),
		ActivityStock: req.ActivityStock,
		Limit:         req.Limit,
		Status:        1,
	}

	err := s.repo.ActivityGoodAdd(ctx, &activityGood)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ActivityOrderUpdate(ctx context.Context) {
	err := s.repo.ActivityOrderUpdate(ctx)
	if err != nil {
		fmt.Println("订单状态修改失败")
		return
	}
	fmt.Println("订单状态修改成功")
	return
}

func (s *Service) ActivityOrder(ctx context.Context, req *shopv1.ActivityOrderRequest) (*shopv1.ActivityOrderResponse, error) {

	//限流
	allowed, err := s.locker.Allow(ctx, fmt.Sprintf("rate_limit:activity_good:%d", req.ActivityGoodId), 1000, 100)
	if err != nil || !allowed {
		return nil, errors.New("当前抢购人数过多，请稍后再试")
	}

	//requestId := time.Now().Unix()
	//res := redisx.NewClient(config.GlobalConfig.Redis)
	//
	//key := fmt.Sprintf("order_%d", requestId)
	////幂等性
	//err = res.Get(ctx, key).Err()
	//if err == nil {
	//	fmt.Println("用户重复点击了")
	//	return nil, errors.New("用户重复点击了")
	//}
	//
	//err = res.Set(ctx, key, "ture", time.Minute).Err()
	//if err != nil {
	//	fmt.Println("requestId存储失败")
	//	return nil, err
	//}

	activityGood, err := s.repo.ActivityGoodFirst(ctx, req.ActivityId, req.ActivityGoodId)
	if err != nil {
		return nil, errors.New("没有这个活动商品")
	}

	if activityGood.Status == 2 {
		return nil, errors.New("活动商品下架了")
	}

	count, err := s.repo.OrderFindById(ctx, req.UserId, req.ActivityId, req.ActivityGoodId)
	if err != nil {
		return nil, err
	}

	if count > activityGood.Limit {
		fmt.Println("用户购买数量超过活动商品限制数量")
		return nil, errors.New("用户购买数量超过活动商品限制数量")
	}

	if activityGood.ActivityStock < req.BuyCount {
		return nil, errors.New("商品库存不足")
	}

	//orderNo := strconv.Itoa(int(time.Now().UnixMilli())) + strconv.Itoa(rand.Intn(10000))
	orderNo := "O" + strconv.Itoa(int(utils.SnowId())) + strconv.Itoa(rand.Intn(10000))

	totalPrice := activityGood.ActivityPrice * float64(req.BuyCount)
	endTime := time.Now().Add(15 * time.Minute)

	order := model.Order{
		OrderNo:        orderNo,
		UserId:         req.UserId,
		OrderType:      2,
		GoodId:         0,
		ActivityId:     req.ActivityId,
		ActivityGoodId: req.ActivityGoodId,
		BuyCount:       req.BuyCount,
		TotalPrice:     totalPrice,
		PayTime:        nil,
		EntTime:        &endTime,
		Status:         1,
	}

	err = s.repo.OrderAdd(ctx, &order)
	if err != nil {
		return nil, err
	}
	totalAmount := strconv.FormatFloat(totalPrice, 'f', -1, 64)
	client, _ := alipayx.NewClient(config.GlobalConfig.Alipay)
	url, err := client.PagePayURL(alipayx.PayRequest{
		OutTradeNo:  orderNo,
		Subject:     "支付宝支付",
		TotalAmount: totalAmount,
	})
	if err != nil {
		return nil, err
	}

	return &shopv1.ActivityOrderResponse{
		TotalPrice: float32(totalPrice),
		Url:        url,
		OrderNo:    orderNo,
	}, nil
}

func (s *Service) CartAdd(ctx context.Context, req *shopv1.CartAddRequest) error {
	f, err := s.repo.ActivityGoodFirst(ctx, req.ActivityId, req.ActivityGoodId)
	if err != nil {
		return err
	}

	if f.Status != 1 {
		return errors.New("活动商品未上架")
	}

	if f.Limit < req.BuyCount {
		return fmt.Errorf("每人限购%d件", f.Limit)
	}

	//totalPrice := f.ActivityPrice * float64(req.BuyCount)
	//cart := model.Cart{
	//	UserId:         req.UserId,
	//	ActivityId:     req.ActivityId,
	//	ActivityGoodId: req.ActivityGoodId,
	//	BuyCount:       req.BuyCount,
	//	TotalPrice:     totalPrice,
	//}

	field := fmt.Sprintf("%d:%d", req.ActivityId, req.ActivityGoodId)
	//field := map[string]interface{}{
	//	"user_id": req.UserId,
	//	"activity_id": req.ActivityId,
	//	"good_id": req.ActivityGoodId,
	//	"buy_count":  req.BuyCount,
	//	"total_price": totalPrice,
	//}

	res := redisx.NewClient(config.GlobalConfig.Redis)
	key := fmt.Sprintf("UserCart_%d", req.UserId)
	err = res.HIncrBy(ctx, key, field, req.BuyCount).Err()
	if err != nil {
		return err
	}

	//key2 := fmt.Sprintf("UserCart2_%d", req.ActivityId)
	//marshal, err := json.Marshal(&cart)
	//if err != nil {
	//	return err
	//}
	//err = res.RPush(ctx, key2, marshal).Err()
	//if err != nil {
	//	return err
	//}
	return nil
}

func (s *Service) CartDel(ctx context.Context, req *shopv1.CartDelRequest) error {

	res := redisx.NewClient(config.GlobalConfig.Redis)
	key := fmt.Sprintf("UserCart_%d", req.UserId)

	err := res.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CartAcc(ctx context.Context, req *shopv1.CartAccRequest) error {
	field := fmt.Sprintf("%d:%d", req.ActivityId, req.ActivityGoodId)

	res := redisx.NewClient(config.GlobalConfig.Redis)
	key := fmt.Sprintf("UserCart_%d", req.UserId)

	code, err := res.HGet(ctx, key, field).Result()
	if err != nil {
		fmt.Println("redis获取失败")
		return err
	}
	f, err := s.repo.ActivityGoodFirst(ctx, req.ActivityId, req.ActivityGoodId)
	if err != nil {
		fmt.Println("未找到这个商品")
		return err
	}
	count, _ := strconv.Atoi(code)
	cot := int64(count) + req.BuyCount
	if cot > f.Limit {
		fmt.Println("每人限购")
		return fmt.Errorf("每人限购%d件", f.Limit)
	}

	err = res.HIncrBy(ctx, key, field, req.BuyCount).Err()
	if err != nil {
		fmt.Println("累加失败")
		return err
	}

	return nil

}

func (s *Service) CartList(ctx context.Context, req *shopv1.CartListRequest) (*shopv1.CartListResponse, error) {
	res := redisx.NewClient(config.GlobalConfig.Redis)
	key := fmt.Sprintf("UserCart_%d", req.UserId)

	result, err := res.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var list []*shopv1.CartList
	for f, v := range result {
		data := shopv1.CartList{
			Key:   f,
			Value: v,
		}
		list = append(list, &data)
	}
	return &shopv1.CartListResponse{List: list}, nil
}

func (s *Service) CartOrder(ctx context.Context, req *shopv1.CartOrderRequest) (*shopv1.CartOrderResponse, error) {
	res := redisx.NewClient(config.GlobalConfig.Redis)
	//限流
	allowed, err := s.locker.Allow(ctx, fmt.Sprintf("rate_limit:activity_good:%d", req.ActivityGoodId), 100, 10)
	if err != nil || !allowed {
		return nil, errors.New("当前抢购人数过多，请稍后再试")
	}

	requestId := time.Now().Unix()

	key := fmt.Sprintf("order_%d", requestId)

	//幂等性
	err = res.Get(ctx, key).Err()
	if err == nil {
		fmt.Println("用户重复点击了")
		return nil, errors.New("用户重复点击了")
	}

	err = res.Set(ctx, key, "ture", time.Minute).Err()
	if err != nil {
		fmt.Println("requestId存储失败")
		return nil, err
	}

	field := fmt.Sprintf("%d:%d", req.ActivityId, req.ActivityGoodId)

	key2 := fmt.Sprintf("UserCart_%d", req.UserId)

	buy, err := res.HGet(ctx, key2, field).Result()
	if err != nil {
		return nil, err
	}

	f, err := s.repo.ActivityGoodFirst(ctx, req.ActivityId, req.ActivityGoodId)
	if err != nil {
		return nil, err
	}

	buyCount, _ := strconv.Atoi(buy)

	totalPrice := f.ActivityPrice * float64(buyCount)

	orderNo := strconv.Itoa(int(time.Now().Unix())) + strconv.Itoa(rand.Intn(10000))

	endTime := time.Now().Add(15 * time.Minute)

	order := model.Order{
		OrderNo:        orderNo,
		UserId:         req.UserId,
		OrderType:      2,
		GoodId:         0,
		ActivityId:     req.ActivityId,
		ActivityGoodId: req.ActivityGoodId,
		BuyCount:       int64(buyCount),
		TotalPrice:     totalPrice,
		PayTime:        nil,
		EntTime:        &endTime,
		Status:         1,
	}

	err = s.repo.OrderAdd(ctx, &order)
	if err != nil {
		return nil, err
	}

	client, err := alipayx.NewClient(config.GlobalConfig.Alipay)
	if err != nil {
		return nil, err
	}

	totalAmount := strconv.FormatFloat(totalPrice, 'f', -1, 64)

	url, err := client.PagePayURL(alipayx.PayRequest{
		OutTradeNo:  orderNo,
		Subject:     "支付宝支付",
		TotalAmount: totalAmount,
	})
	if err != nil {
		return nil, err
	}

	return &shopv1.CartOrderResponse{
		TotalPrice: float32(totalPrice),
		Url:        url,
		OrderNo:    orderNo,
	}, nil
}

func (s *Service) ActivityAdd(ctx context.Context, req *shopv1.ActivityAddRequest) error {

	//err := s.repo.ActivityByName(ctx, req.Name)
	//if err == nil {
	//	return errors.New("活动名重复")
	//}

	startTime, _ := time.ParseInLocation("2006-01-02 15:04", req.StartTime, time.Local)
	endTime, _ := time.ParseInLocation("2006-01-02 15:04", req.EndTime, time.Local)
	if startTime.Unix() > endTime.Unix() {
		return errors.New("活动时间有问题")
	}

	activity := model.Activity{
		ActivityName: req.Name,
		StartTime:    &startTime,
		EndTime:      &endTime,
		Status:       1,
	}

	err := s.repo.ActivityAdd(ctx, &activity)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ActivityDel(ctx context.Context, req *shopv1.ActivityDelRequest) error {
	err := s.repo.ActivityDel(ctx, req.Id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ActivityUpdate(ctx context.Context, req *shopv1.ActivityUpdateRequest) error {
	err := s.repo.ActivityUpdate(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) ActivityList(ctx context.Context, req *shopv1.ActivityListRequest) (*shopv1.ActivityListResponse, error) {
	activitys, err := s.repo.ActivityList(ctx, req)
	if err != nil {
		return nil, err
	}

	data := []*shopv1.ActivityList{}

	for _, a := range activitys {
		list := &shopv1.ActivityList{
			Id:           int64(a.ID),
			ActivityName: a.ActivityName,
			StartTime:    a.StartTime.String(),
			EndTime:      a.EndTime.String(),
			Status:       a.Status,
		}
		data = append(data, list)
	}

	return &shopv1.ActivityListResponse{List: data}, nil
}
