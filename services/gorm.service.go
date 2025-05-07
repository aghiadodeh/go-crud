package services

import (
	"context"

	"github.com/aghiadodeh/go-crud/configs"
	"github.com/aghiadodeh/go-crud/repositories"
)

type GormCrudService[T any] struct {
	BaseCrudService[T, configs.GormConfig, repositories.BaseRepository[T, configs.GormConfig]]
}

func NewGormCrudService[T any](repository repositories.BaseRepository[T, configs.GormConfig]) *GormCrudService[T] {
	return &GormCrudService[T]{
		BaseCrudService: *NewBaseCrudService(repository),
	}
}

func (s *GormCrudService[T]) Update(ctx context.Context, id any, updateDto any, config *configs.GormConfig, args ...any) (*T, error) {
	// Check if record exists before updating
	count, err := s.Repository.Count(ctx, repositories.GormConditionBuilder([]configs.GormQueryField{{Column: "id", Value: id}}))
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}

	// Delegate to BaseCrudService's Update implementation
	return s.BaseCrudService.Update(ctx, id, updateDto, config, args...)
}
