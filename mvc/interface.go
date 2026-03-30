package mvc

import (
	"context"

	"gorm.io/gorm/schema"
)

// IBaseDao 定义通用的数据访问接口
type IRepository[T schema.Tabler] interface {
	// ================= 1. 写入操作 =================
	Create(ctx context.Context, entity *T) error
	CreateBatch(ctx context.Context, entities []*T) error

	// ================= 2. 主键操作（高频核心）=================
	DeleteById(ctx context.Context, id any) error
	DeleteByIds(ctx context.Context, ids []any) (int64, error)
	UpdateById(ctx context.Context, id any, entity *T) (int64, error)
	FindById(ctx context.Context, id any) (*T, error)
	FindByIds(ctx context.Context, ids []any) ([]*T, error)

	// ================= 3. 动态 Map 条件操作（安全、精准）=================
	// 统一使用 map 避免 GORM 零值忽略陷阱
	UpdateByMap(ctx context.Context, conditions map[string]any, entity *T) (int64, error)
	FindOneByMap(ctx context.Context, conditions map[string]any) (*T, error)
	FindListByMap(ctx context.Context, conditions map[string]any) ([]*T, error)
	FindPageByMap(ctx context.Context, page *Page, conditions map[string]any) ([]*T, int64, error)
	CountByMap(ctx context.Context, conditions map[string]any) (int64, error)
}
