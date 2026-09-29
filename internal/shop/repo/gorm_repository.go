package repo

import (
	"context"
	"errors"
	"time"

	shopv1 "github.com/BwCloudWeGo/bw-cli/api/gen/shop/v1"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/BwCloudWeGo/bw-cli/internal/shop/entity"
	dbmodel "github.com/BwCloudWeGo/bw-cli/internal/shop/model"
)

// GormRepository 使用 Gorm 持久化 shop 聚合。
type GormRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewGormRepository 创建 shop 仓储，并支持可选结构化日志。
func NewGormRepository(db *gorm.DB, loggers ...*zap.Logger) *GormRepository {
	log := zap.NewNop()
	if len(loggers) > 0 && loggers[0] != nil {
		log = loggers[0]
	}
	return &GormRepository{db: db, log: log}
}

// AutoMigrate 创建或更新 shops 表结构。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&dbmodel.User{}, &dbmodel.Activity{}, &dbmodel.ActivityGood{},
		&dbmodel.Cart{}, &dbmodel.Order{})
}

// Save 新增或更新 shop 聚合。
func (r *GormRepository) Save(ctx context.Context, item *entity.Shop) error {
	start := time.Now()
	tx := r.db.WithContext(ctx).Save(toRecord(item))
	r.logOperation("Save", tx.RowsAffected, start, tx.Error)
	return tx.Error
}

// FindByID 根据 ID 加载 shop 聚合。
func (r *GormRepository) FindByID(ctx context.Context, id string) (*entity.Shop, error) {
	start := time.Now()
	var record dbmodel.ShopModel
	tx := r.db.WithContext(ctx).Where("id = ?", id).First(&record)
	err := tx.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = entity.ErrShopNotFound
	}
	if err != nil {
		r.logOperation("FindByID", tx.RowsAffected, start, err)
		return nil, err
	}
	r.logOperation("FindByID", tx.RowsAffected, start, nil)
	return toDomain(&record), nil
}

// List 加载分页 shop 聚合。
func (r *GormRepository) List(ctx context.Context, offset int, limit int) ([]*entity.Shop, int64, error) {
	start := time.Now()
	var total int64
	countTx := r.db.WithContext(ctx).Model(&dbmodel.ShopModel{}).Count(&total)
	if countTx.Error != nil {
		r.logOperation("Count", countTx.RowsAffected, start, countTx.Error)
		return nil, 0, countTx.Error
	}
	var records []dbmodel.ShopModel
	tx := r.db.WithContext(ctx).
		Order("created_at desc").
		Offset(offset).
		Limit(limit).
		Find(&records)
	if tx.Error != nil {
		r.logOperation("List", tx.RowsAffected, start, tx.Error)
		return nil, 0, tx.Error
	}
	items := make([]*entity.Shop, 0, len(records))
	for i := range records {
		items = append(items, toDomain(&records[i]))
	}
	r.logOperation("List", tx.RowsAffected, start, nil)
	return items, total, nil
}

// Delete 根据 ID 删除 shop 聚合。
func (r *GormRepository) Delete(ctx context.Context, id string) error {
	start := time.Now()
	tx := r.db.WithContext(ctx).Where("id = ?", id).Delete(&dbmodel.ShopModel{})
	err := tx.Error
	if err == nil && tx.RowsAffected == 0 {
		err = entity.ErrShopNotFound
	}
	r.logOperation("Delete", tx.RowsAffected, start, err)
	return err
}

func (r *GormRepository) logOperation(operation string, rows int64, start time.Time, err error) {
	fields := []zap.Field{
		zap.String("repository", "shop"),
		zap.String("operation", operation),
		zap.Int64("rows_affected", rows),
		zap.Float64("latency_ms", float64(time.Since(start).Microseconds())/1000),
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
		r.log.Warn("repository operation completed with error", fields...)
		return
	}
	r.log.Info("repository operation completed", fields...)
}

func toRecord(item *entity.Shop) *dbmodel.ShopModel {
	return &dbmodel.ShopModel{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func toDomain(record *dbmodel.ShopModel) *entity.Shop {
	return &entity.Shop{
		ID:          record.ID,
		Name:        record.Name,
		Description: record.Description,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

func (r *GormRepository) ActivityGoodAdd(ctx context.Context, m *dbmodel.ActivityGood) error {
	return r.db.Create(&m).Error
}

func (r *GormRepository) ActivityGoodFirst(ctx context.Context, id int64, id2 int64) (*dbmodel.ActivityGood, error) {
	var activityGood dbmodel.ActivityGood
	err := r.db.WithContext(ctx).Where("id=? AND activity_id=?", id2, id).First(&activityGood).Error
	if err != nil {
		return nil, err
	}

	return &activityGood, nil
}

func (r *GormRepository) OrderAdd(ctx context.Context, m *dbmodel.Order) error {
	err := r.db.WithContext(ctx).Create(&m).Error
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(&dbmodel.ActivityGood{}).
		Where("id=? AND activity_id=?", m.ActivityGoodId, m.ActivityId).
		Update("activity_stock", gorm.Expr("activity_stock - ?", m.BuyCount)).
		Error
}

func (r *GormRepository) ActivityGoodFind(ctx context.Context) ([]*dbmodel.ActivityGood, error) {
	var activityGoods []*dbmodel.ActivityGood
	err := r.db.WithContext(ctx).Find(&activityGoods).Error
	if err != nil {
		return nil, err
	}

	return activityGoods, nil
}

func (r *GormRepository) OrderFindById(ctx context.Context, id int64, id2 int64, id3 int64) (int64, error) {
	var orders []*dbmodel.Order
	err := r.db.WithContext(ctx).Where("user_id=? AND activity_id=? AND activity_good_id=?", id, id2, id3).Find(&orders).Error
	if err != nil {
		return 0, err
	}

	var count int64
	for _, order := range orders {
		count += order.BuyCount
	}

	return count, nil
}

func (r *GormRepository) ActivityOrderUpdate(ctx context.Context) error {
	orders := []*dbmodel.Order{}

	err := r.db.WithContext(ctx).Find(&orders).Error
	if err != nil {
		return err
	}
	newTime := time.Now().Unix()
	for _, order := range orders {
		endTime := order.EntTime.Unix()
		if newTime > endTime {
			err = r.db.WithContext(ctx).Model(&dbmodel.Order{}).
				Where("order_no=?", order.OrderNo).
				Update("status", 3).Error
			if err != nil {
				return err
			}

			err = r.db.WithContext(ctx).Model(&dbmodel.ActivityGood{}).
				Where("id=? AND activity_id=?", order.ActivityGoodId, order.ActivityId).
				Update("activity_stock", gorm.Expr("activity_stock + ?", order.BuyCount)).Error
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func (r *GormRepository) ActivityByName(ctx context.Context, name string) error {
	return r.db.WithContext(ctx).
		Where("activity_name=?", name).First(&dbmodel.Activity{}).Error
}

func (r *GormRepository) ActivityAdd(ctx context.Context, m *dbmodel.Activity) error {
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormRepository) ActivityDel(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id=?", id).Delete(&dbmodel.Activity{}).Error
}

func (r *GormRepository) ActivityUpdate(ctx context.Context, req *shopv1.ActivityUpdateRequest) error {
	var activity dbmodel.Activity
	err := r.db.WithContext(ctx).Where("id=?", req.Id).First(&activity).Error
	if err != nil {
		return err
	}

	if req.Name != "" {
		activity.ActivityName = req.Name
	}
	if req.EndTime != "" {
		endTime, _ := time.ParseInLocation("2006-01-02 15:04", req.EndTime, time.Local)
		activity.EndTime = &endTime
	}

	if req.StartTime != "" {
		startTime, _ := time.ParseInLocation("2006-01-02 15:04", req.EndTime, time.Local)
		activity.EndTime = &startTime
	}

	if req.Status != 0 {
		activity.Status = req.Status
	}

	return r.db.WithContext(ctx).Model(&dbmodel.Activity{}).
		Where("id=?", req.Id).
		Updates(activity).Error

}

func (r *GormRepository) ActivityList(ctx context.Context, req *shopv1.ActivityListRequest) ([]*dbmodel.Activity, error) {
	var activitys []*dbmodel.Activity
	query := r.db.WithContext(ctx)
	if req.Name != "" {
		query = query.Where("activity_name LIKE ?", "%"+req.Name+"%")
	}
	if req.Status != 0 {
		query = query.Where("status = ?", req.Status)
	}
	err := query.Find(&activitys).Error
	if err != nil {
		return nil, err
	}

	return activitys, nil
}

var _ entity.Repository = (*GormRepository)(nil)
