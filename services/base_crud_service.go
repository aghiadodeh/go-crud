package services

import (
	"context"

	"github.com/aghiadodeh/go-crud/dto"
	"github.com/aghiadodeh/go-crud/models"
	"github.com/aghiadodeh/go-crud/repositories"
)

type BaseCrudService[T any, C any, R repositories.BaseRepository[T, C]] struct {
	Repository R
}

func NewBaseCrudService[T any, C any, R repositories.BaseRepository[T, C]](repository R) *BaseCrudService[T, C, R] {
	return &BaseCrudService[T, C, R]{Repository: repository}
}

func (s *BaseCrudService[T, C, R]) Create(ctx context.Context, createDto any, config *C, args ...any) (*T, error) {
	id, err := s.Repository.Create(ctx, createDto, args...)
	if err != nil {
		return nil, err
	}
	item, err := s.Repository.FindOneByPK(ctx, id, config, args...)
	return item, err
}

func (s *BaseCrudService[T, C, R]) Update(ctx context.Context, id any, updateDto any, config *C, args ...any) (*T, error) {
	if err := s.Repository.UpdateByPK(ctx, id, updateDto, args...); err != nil {
		return nil, err
	}
	return s.Repository.FindOneByPK(ctx, id, config, args...)
}

func (s *BaseCrudService[T, C, R]) UpdateColumnsByPK(ctx context.Context, id any, columns map[string]any, args ...any) error {
	return s.Repository.UpdateColumnsByPK(ctx, id, columns, args...)
}

func (s *BaseCrudService[T, C, R]) FindAll(ctx context.Context, conditions any, filter dto.FilterDto, config *C, args ...any) ([]T, error) {
	return s.Repository.FindAll(ctx, conditions, filter, config, args...)
}

func (s *BaseCrudService[T, C, R]) FindAllWithPaging(ctx context.Context, conditions any, filter dto.FilterDto, config *C, args ...any) (*models.ListResponse[T], error) {
	return s.Repository.FindAllWithPaging(ctx, conditions, filter, config, args...)
}

func (s *BaseCrudService[T, C, R]) FindOne(ctx context.Context, conditions any, config *C, args ...any) (*T, error) {
	return s.Repository.FindOne(ctx, conditions, config, args...)
}

func (s *BaseCrudService[T, C, R]) FindOneByPK(ctx context.Context, id any, config *C, args ...any) (*T, error) {
	return s.Repository.FindOneByPK(ctx, id, config, args...)
}

func (s *BaseCrudService[T, C, R]) FindByIDs(ctx context.Context, ids []any, config *C, args ...any) ([]T, error) {
	return s.Repository.FindByIDs(ctx, ids, config, args...)
}

func (s *BaseCrudService[T, C, R]) Delete(ctx context.Context, conditions any, args ...any) error {
	return s.Repository.Delete(ctx, conditions, args...)
}

func (s *BaseCrudService[T, C, R]) DeleteOneByPK(ctx context.Context, id any, args ...any) error {
	return s.Repository.DeleteOneByPK(ctx, id, args...)
}

func (s *BaseCrudService[T, C, R]) DeleteByIDs(ctx context.Context, ids []any, args ...any) error {
	return s.Repository.DeleteByIDs(ctx, ids, args...)
}

func (s *BaseCrudService[T, C, R]) Count(ctx context.Context, conditions any, args ...any) (int64, error) {
	return s.Repository.Count(ctx, conditions, args...)
}

func (s *BaseCrudService[T, C, R]) Exists(ctx context.Context, conditions any, args ...any) (bool, error) {
	return s.Repository.Exists(ctx, conditions, args...)
}

func (s *BaseCrudService[T, C, R]) ExistsByPK(ctx context.Context, id any, args ...any) (bool, error) {
	return s.Repository.ExistsByPK(ctx, id, args...)
}

func (s *BaseCrudService[T, C, R]) Pluck(ctx context.Context, column string, conditions any, args ...any) ([]any, error) {
	return s.Repository.Pluck(ctx, column, conditions, args...)
}

func (s *BaseCrudService[T, C, R]) QueryBuilder(ctx context.Context, filter dto.FilterDto, config *C, args ...any) (any, error) {
	return s.Repository.QueryBuilder(ctx, filter, config, args...)
}
