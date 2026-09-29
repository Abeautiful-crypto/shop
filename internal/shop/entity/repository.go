package entity

import (
	"context"

	shopv1 "github.com/BwCloudWeGo/bw-cli/api/gen/shop/v1"
	dbmodel "github.com/BwCloudWeGo/bw-cli/internal/shop/model"
)

// Repository 定义 shop service 层需要的持久化行为。
type Repository interface {
	Save(ctx context.Context, item *Shop) error
	FindByID(ctx context.Context, id string) (*Shop, error)
	List(ctx context.Context, offset int, limit int) ([]*Shop, int64, error)
	Delete(ctx context.Context, id string) error
	ActivityGoodAdd(ctx context.Context, m *dbmodel.ActivityGood) error
	ActivityGoodFirst(ctx context.Context, id int64, id2 int64) (*dbmodel.ActivityGood, error)
	OrderAdd(ctx context.Context, m *dbmodel.Order) error
	ActivityGoodFind(ctx context.Context) ([]*dbmodel.ActivityGood, error)
	OrderFindById(ctx context.Context, id int64, id2 int64, id3 int64) (count int64, err error)
	ActivityOrderUpdate(ctx context.Context) error
	ActivityByName(ctx context.Context, name string) error
	ActivityAdd(ctx context.Context, m *dbmodel.Activity) error
	ActivityDel(ctx context.Context, id int64) error
	ActivityUpdate(ctx context.Context, req *shopv1.ActivityUpdateRequest) error
	ActivityList(ctx context.Context, req *shopv1.ActivityListRequest) ([]*dbmodel.Activity, error)
}
