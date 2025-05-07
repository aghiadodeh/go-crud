package services

import (
	"context"

	"github.com/aghiadodeh/go-crud/dto"
	"github.com/aghiadodeh/go-crud/models"
)

type IBaseCrudService[T any, C any] interface {
	Create(ctx context.Context, createDto any, config *C, args ...any) (*T, error)
	Update(ctx context.Context, id any, updateDto any, config *C, args ...any) (*T, error)
	FindAll(ctx context.Context, conditions any, filter dto.FilterDto, config *C, args ...any) ([]T, error)
	FindAllWithPaging(ctx context.Context, conditions any, filter dto.FilterDto, config *C, args ...any) (*models.ListResponse[T], error)
	FindOne(ctx context.Context, conditions any, config *C, args ...any) (*T, error)
	FindOneByPK(ctx context.Context, id any, config *C, args ...any) (*T, error)
	Delete(ctx context.Context, conditions any, args ...any) error
	DeleteOneByPK(ctx context.Context, id any, args ...any) error
	Count(ctx context.Context, conditions any, args ...any) (int64, error)
	QueryBuilder(ctx context.Context, filter dto.FilterDto, config *C, args ...any) (any, error)
}
