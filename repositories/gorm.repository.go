package repositories

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/aghiadodeh/go-crud/configs"
	"github.com/aghiadodeh/go-crud/dto"
	"github.com/aghiadodeh/go-crud/middlewares"
	"github.com/aghiadodeh/go-crud/models"
)

type GormRepository[T any] struct {
	DB        *gorm.DB
	Config    *configs.GormConfig
	TableName string
}

func NewGormRepository[T any](db *gorm.DB, config *configs.GormConfig, tableName string) *GormRepository[T] {
	return &GormRepository[T]{DB: db, Config: config, TableName: tableName}
}

func (r *GormRepository[T]) Create(ctx context.Context, createDto any, args ...any) (string, error) {
	if err := r.DB.WithContext(ctx).Table(r.TableName).Create(createDto).Error; err != nil {
		return "", err
	}

	// Assuming ID is a string field
	id := ""
	if err := r.DB.WithContext(ctx).Table(r.TableName).Model(new(T)).Select("id").Last(&id).Error; err != nil {
		return "", err
	}

	return id, nil
}

func (r *GormRepository[T]) BulkCreate(ctx context.Context, createDto []any, args ...any) ([]string, error) {
	var entities []T
	if err := r.DB.WithContext(ctx).Model(&entities).Create(createDto).Error; err != nil {
		return nil, err
	}

	var ids []string
	if err := r.DB.WithContext(ctx).Model(&entities).Select("id").Find(&ids).Error; err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *GormRepository[T]) UpdateByPK(ctx context.Context, id string, updateDto any, args ...any) error {
	return r.DB.WithContext(ctx).Table(r.TableName).Where("id = ?", id).Updates(updateDto).Error
}

func (r *GormRepository[T]) Update(ctx context.Context, conditions any, updateDto any, args ...any) error {
	return r.DB.WithContext(ctx).Table(r.TableName).Where(conditions).Updates(updateDto).Error
}

func (r *GormRepository[T]) FindAll(ctx context.Context, conditions any, filter dto.FilterDto, config *configs.GormConfig, args ...any) ([]T, error) {
	var models []T
	query := r.buildBaseQuery(ctx, conditions, filter, config)
	err := query.Find(&models).Error
	return models, err
}

func (r *GormRepository[T]) FindAllWithPaging(ctx context.Context, conditions any, filter dto.FilterDto, config *configs.GormConfig, args ...any) (*models.ListResponse[T], error) {
	var entities []T
	var total int64

	query := r.buildBaseQuery(ctx, conditions, filter, config)
	countQuery := r.BuildQueryConditions(ctx, conditions)

	if err := countQuery.Table(r.TableName).Count(&total).Error; err != nil {
		return nil, err
	}

	filterDto := filter.GetBase()
	if filterDto.Pagination == nil || *filterDto.Pagination {
		query = query.Scopes(Paginate(filterDto.Page, filterDto.PerPage))
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}

	return &models.ListResponse[T]{
		Total: total,
		Data:  entities,
	}, nil
}

func (r *GormRepository[T]) FindOne(ctx context.Context, conditions any, config *configs.GormConfig, args ...any) (*T, error) {
	var model T
	query := r.BuildQueryConfig(ctx, conditions, config)
	err := query.First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &model, err
}

func (r *GormRepository[T]) FindOneByPK(ctx context.Context, id string, config *configs.GormConfig, args ...any) (*T, error) {
	var model T
	condition := make(map[string]any)
	condition["id = ?"] = id
	query := r.BuildQueryConfig(ctx, condition, config)
	err := query.First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &model, err
}

func (r *GormRepository[T]) Delete(ctx context.Context, conditions any, args ...any) error {
	return r.DB.WithContext(ctx).Table(r.TableName).Where(conditions).Delete(new(T)).Error
}

func (r *GormRepository[T]) DeleteOneByPK(ctx context.Context, id string, args ...any) error {
	return r.DB.WithContext(ctx).Table(r.TableName).Where("id = ?", id).Delete(new(T)).Error
}

func (r *GormRepository[T]) Count(ctx context.Context, conditions any, args ...any) (int64, error) {
	var count int64
	query := r.DB.WithContext(ctx).Table(r.TableName)
	if conditions != nil {
		query = query.Where(conditions)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *GormRepository[T]) QueryBuilder(ctx context.Context, filter dto.FilterDto, gormConfig *configs.GormConfig, args ...any) (any, error) {
	var queryStrings []string
	var queryValues []any

	var config configs.GormConfig
	if gormConfig == nil {
		config = *r.Config
	} else {
		config = *gormConfig
	}

	// Handle search
	filterDto := filter.GetBase()
	if filterDto.Search != nil && len(config.Searchable) > 0 {
		var searchParts []string
		for _, field := range config.Searchable {
			searchParts = append(searchParts, fmt.Sprintf("%s LIKE ?", field))
			queryValues = append(queryValues, fmt.Sprintf("%%%s%%", *filterDto.Search))
		}
		queryStrings = append(queryStrings, "("+strings.Join(searchParts, " OR ")+")")
	}

	// Handle filters
	result, err := filter.ToMap()
	if err != nil {
		return nil, err
	}

	for key, value := range result {
		if prop, ok := config.Filterable[key]; ok {
			column := prop.ColumnName
			if column == "" {
				column = key
			}

			switch prop.FilterType {
			case configs.GormFilterTypeEqual:
				queryStrings = append(queryStrings, fmt.Sprintf("%s = ?", column))
				queryValues = append(queryValues, value)
			case configs.GormFilterTypeIn:
				queryStrings = append(queryStrings, fmt.Sprintf("%s IN (?)", column))
				queryValues = append(queryValues, value)
			case configs.GormFilterTypeLT:
				queryStrings = append(queryStrings, fmt.Sprintf("%s < ?", column))
				queryValues = append(queryValues, value)
			case configs.GormFilterTypeGT:
				queryStrings = append(queryStrings, fmt.Sprintf("%s > ?", column))
				queryValues = append(queryValues, value)
			case configs.GormFilterTypeLTE:
				queryStrings = append(queryStrings, fmt.Sprintf("%s <= ?", column))
				queryValues = append(queryValues, value)
			case configs.GormFilterTypeGTE:
				queryStrings = append(queryStrings, fmt.Sprintf("%s >= ?", column))
				queryValues = append(queryValues, value)
			case configs.GormFilterTypeRegex:
				queryStrings = append(queryStrings, fmt.Sprintf("%s LIKE ?", column))
				queryValues = append(queryValues, fmt.Sprintf("%%%s%%", value))
			}
		}
	}

	finalQuery := strings.Join(queryStrings, " AND ")
	return map[string]any{
		"query": finalQuery,
		"args":  queryValues,
	}, nil
}

func (r *GormRepository[T]) BuildQueryConditions(ctx context.Context, conditions any) *gorm.DB {
	query := r.DB.WithContext(ctx).Table(r.TableName)
	if conditionsMap, ok := conditions.(map[string]any); ok {
		if q, ok := conditionsMap["query"].(string); ok && q != "" {
			args := conditionsMap["args"].([]interface{})
			query = query.Where(q, args...)
		}
	}
	return query
}

func (r *GormRepository[T]) BuildQueryConfig(ctx context.Context, conditions any, gormConfig *configs.GormConfig) *gorm.DB {
	query := r.BuildQueryConditions(ctx, conditions)

	var config configs.GormConfig
	if gormConfig == nil {
		config = *r.Config
	} else {
		config = *gormConfig
	}

	lang := middlewares.GetLangFromContext(ctx)
	// Handle dynamic SELECTs
	if config.SelectHandler != nil {
		selects := config.SelectHandler(lang)
		var selectClauses []string
		for _, f := range selects {
			selectClauses = append(selectClauses, fmt.Sprintf("%s AS %s", f.Column, f.Alias))
		}
		query = query.Select(strings.Join(selectClauses, ", "))
	}

	// Handle dynamic Preloads
	for _, preload := range config.Preloads {
		if preload.SelectHandler != nil {
			selects := preload.SelectHandler(lang)
			var preloadClauses []string
			for _, f := range selects {
				preloadClauses = append(preloadClauses, fmt.Sprintf("%s AS %s", f.Column, f.Alias))
			}
			query = query.Preload(preload.Relation, func(db *gorm.DB) *gorm.DB {
				return db.Select(strings.Join(preloadClauses, ", "))
			})
		} else {
			query = query.Preload(preload.Relation)
		}
	}

	// Get all raw including soft-deleted
	if config.UnScoped {
		query = query.Unscoped()
	}

	return query
}

func (r *GormRepository[T]) buildBaseQuery(ctx context.Context, conditions any, filter dto.FilterDto, gormConfig *configs.GormConfig) *gorm.DB {
	query := r.BuildQueryConfig(ctx, conditions, gormConfig)
	var config configs.GormConfig
	if gormConfig == nil {
		config = *r.Config
	} else {
		config = *gormConfig
	}

	// Apply sorting
	filterDto := filter.GetBase()

	sortKey := "created_at"
	if filterDto.SortKey != nil {
		sortKey = *filterDto.SortKey
	} else if config.DefaultSort != "" {
		sortKey = config.DefaultSort
	}

	sortDir := "desc"
	if filterDto.SortDir != nil {
		sortDir = strings.ToLower(*filterDto.SortDir)
	}

	query = query.Order(fmt.Sprintf("%s %s", sortKey, sortDir))

	return query
}

func Paginate(page, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if size <= 0 {
			size = 10
		}
		offset := (page - 1) * size
		return db.Offset(offset).Limit(size)
	}
}

func Sort(sort, dir string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if sort != "" {
			if dir == "" {
				dir = "desc"
			}
			db.Order(fmt.Sprintf("%s %s", sort, dir))
		}
		return db
	}
}
