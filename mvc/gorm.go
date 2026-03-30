package mvc

import (
	"context"
	"fmt"

	errorc "github.com/xsxdot/gokit/err"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// GormDaoImpl GORM数据访问实现
type GormDaoImpl[T schema.Tabler] struct {
	db    *gorm.DB
	model string
}

// NewGormDao 创建GORM数据访问实例
func NewGormDao[T schema.Tabler](db *gorm.DB) IRepository[T] {
	t := *new(T)
	return &GormDaoImpl[T]{
		db:    db,
		model: t.TableName(),
	}
}

// ================= 1. 写入操作 =================

func (d *GormDaoImpl[T]) Create(ctx context.Context, entity *T) error {
	err := ExtractDB(ctx, d.db).Create(entity).Error
	if err != nil {
		return errorc.New(fmt.Sprintf("[%s] 数据库操作失败", d.model), err).DB()
	}
	return nil
}

func (d *GormDaoImpl[T]) CreateBatch(ctx context.Context, entities []*T) error {
	err := ExtractDB(ctx, d.db).Create(entities).Error
	if err != nil {
		return errorc.New(fmt.Sprintf("[%s] 批量创建记录失败", d.model), err).DB()
	}
	return nil
}

// ================= 2. 主键操作（高频核心）=================

func (d *GormDaoImpl[T]) DeleteById(ctx context.Context, id any) error {
	result := ExtractDB(ctx, d.db).Delete(new(T), id)
	if result.Error != nil {
		return errorc.New(fmt.Sprintf("[%s] 删除记录失败", d.model), result.Error).DB()
	}
	if result.RowsAffected == 0 {
		return errorc.New(fmt.Sprintf("[%s] 要删除的记录不存在", d.model), nil).WithCode(errorc.ErrorCodeNotFound)
	}
	return nil
}

func (d *GormDaoImpl[T]) DeleteByIds(ctx context.Context, ids []any) (int64, error) {
	result := ExtractDB(ctx, d.db).Delete(new(T), ids)
	if result.Error != nil {
		return 0, errorc.New(fmt.Sprintf("[%s] 批量删除记录失败", d.model), result.Error).DB()
	}
	if result.RowsAffected == 0 {
		return 0, errorc.New(fmt.Sprintf("[%s] 要删除的记录不存在", d.model), nil).WithCode(errorc.ErrorCodeNotFound)
	}
	return result.RowsAffected, nil
}

func (d *GormDaoImpl[T]) UpdateById(ctx context.Context, id any, entity *T) (int64, error) {
	result := ExtractDB(ctx, d.db).Model(new(T)).Where("id = ?", id).Updates(entity)
	err := result.Error
	if err != nil {
		return 0, errorc.New(fmt.Sprintf("[%s] 更新记录失败", d.model), err).DB()
	}
	affected := result.RowsAffected
	if affected == 0 {
		return 0, errorc.New(fmt.Sprintf("[%s] 要更新的记录不存在", d.model), nil).WithCode(errorc.ErrorCodeNotFound)
	}
	return affected, err
}

func (d *GormDaoImpl[T]) FindById(ctx context.Context, id any) (*T, error) {
	var entity T
	err := ExtractDB(ctx, d.db).First(&entity, id).Error
	if err != nil {
		return nil, errorc.New(fmt.Sprintf("[%s] 查询记录失败", d.model), err).DB()
	}
	return &entity, nil
}

func (d *GormDaoImpl[T]) FindByIds(ctx context.Context, ids []any) ([]*T, error) {
	var entities []*T
	err := ExtractDB(ctx, d.db).Find(&entities, ids).Error
	if err != nil {
		return nil, errorc.New(fmt.Sprintf("[%s] 批量查询记录失败", d.model), err).DB()
	}
	return entities, nil
}

// ================= 3. 动态 Map 条件操作（安全、精准）=================

func (d *GormDaoImpl[T]) UpdateByMap(ctx context.Context, conditions map[string]any, entity *T) (int64, error) {
	db := ExtractDB(ctx, d.db).Model(new(T))
	if len(conditions) > 0 {
		db = db.Where(conditions)
	}
	result := db.Updates(entity)
	err := result.Error
	if err != nil {
		return 0, errorc.New(fmt.Sprintf("[%s] 更新记录失败", d.model), err).DB()
	}
	affected := result.RowsAffected
	if affected == 0 {
		return 0, errorc.New(fmt.Sprintf("[%s] 要更新的记录不存在", d.model), nil).WithCode(errorc.ErrorCodeNotFound)
	}
	return affected, err
}

func (d *GormDaoImpl[T]) FindOneByMap(ctx context.Context, conditions map[string]any) (*T, error) {
	var entity T
	db := ExtractDB(ctx, d.db).Model(new(T))
	if len(conditions) > 0 {
		db = db.Where(conditions)
	}
	err := db.First(&entity).Error
	if err != nil {
		return nil, errorc.New(fmt.Sprintf("[%s] 查询记录失败", d.model), err).DB()
	}
	return &entity, nil
}

func (d *GormDaoImpl[T]) FindListByMap(ctx context.Context, conditions map[string]any) ([]*T, error) {
	var entities []*T
	db := ExtractDB(ctx, d.db).Model(new(T))
	if len(conditions) > 0 {
		db = db.Where(conditions)
	}
	err := db.Find(&entities).Error
	if err != nil {
		return nil, errorc.New(fmt.Sprintf("[%s] 查询记录失败", d.model), err).DB()
	}
	return entities, nil
}

func (d *GormDaoImpl[T]) FindPageByMap(ctx context.Context, page *Page, conditions map[string]any) ([]*T, int64, error) {
	var entities []*T
	var total int64

	db := ExtractDB(ctx, d.db).Model(new(T))
	if len(conditions) > 0 {
		db = db.Where(conditions)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errorc.New(fmt.Sprintf("[%s] 查询记录失败", d.model), err).DB()
	}

	err = db.Scopes(Paginate(page)).Find(&entities).Error
	if err != nil {
		return nil, 0, errorc.New(fmt.Sprintf("[%s] 查询记录失败", d.model), err).DB()
	}

	return entities, total, nil
}

func (d *GormDaoImpl[T]) CountByMap(ctx context.Context, conditions map[string]any) (int64, error) {
	var count int64
	db := ExtractDB(ctx, d.db).Model(new(T))
	if len(conditions) > 0 {
		db = db.Where(conditions)
	}
	err := db.Count(&count).Error
	if err != nil {
		return 0, errorc.New(fmt.Sprintf("[%s] 查询记录失败", d.model), err).DB()
	}
	return count, err
}
