package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/CakeForKit/rsoi-lab2/lb-core/db"
	"gorm.io/gorm"
)

type Repository[T any] interface {
	GetById(ctx context.Context, ids []uint) ([]T, error)
	GetAll(ctx context.Context) ([]T, error)

	Create(ctx context.Context, entities []T) ([]T, error)
	Update(ctx context.Context, entities []T) ([]T, error)
	DeleteById(ctx context.Context, ids []uint) error
}

type postgresRepository[T any] struct {
	dataSource *gorm.DB
}

func NewPostgresRepository[T any]() (Repository[T], error) {
	dataSource := db.GetPostgresDataSource()
	var entity T
	if err := dataSource.AutoMigrate(&entity); err != nil && !strings.Contains(err.Error(), "already exists") {
		return nil, err
	}
	return &postgresRepository[T]{dataSource: dataSource}, nil
}

func (pgRepo *postgresRepository[T]) GetById(ctx context.Context, ids []uint) ([]T, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("ids can't be empty")
	}

	var result []T
	tx := pgRepo.dataSource.WithContext(ctx)
	if err := tx.Where("id in ?", ids).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (pgRepo *postgresRepository[T]) GetAll(ctx context.Context) ([]T, error) {
	var result []T
	tx := pgRepo.dataSource.WithContext(ctx)
	if err := tx.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (pgRepo *postgresRepository[T]) Create(ctx context.Context, entities []T) ([]T, error) {
	tx := pgRepo.dataSource.WithContext(ctx)
	result := tx.Create(&entities)
	if err := result.Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (pgRepo *postgresRepository[T]) Update(ctx context.Context, entities []T) ([]T, error) {
	tx := pgRepo.dataSource.WithContext(ctx)
	result := tx.Save(&entities)
	if err := result.Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (pgRepo *postgresRepository[T]) DeleteById(ctx context.Context, ids []uint) error {
	var entity []T
	tx := pgRepo.dataSource.WithContext(ctx)
	result := tx.Where("id in ?", ids).Delete(&entity)
	return result.Error
}
