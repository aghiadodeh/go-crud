package repositories

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

func (r *GormRepository[T]) Create(ctx context.Context, createDto any, args ...any) (any, error) {
	entity, ok := createDto.(T)
	if !ok {
		return "", fmt.Errorf("invalid type passed to Create: expected %T", entity)
	}

	err := r.DB.WithContext(ctx).Model(new(T)).Create(&entity).Error
	if err != nil {
		return "", err
	}

	// Assuming ID is a string field
	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return "", fmt.Errorf("ID field not found on entity")
	}
	switch idField.Kind() {
	case reflect.String:
		return idField.String(), nil
	case reflect.Int, reflect.Int64:
		return idField.Int(), nil
	case reflect.Uint, reflect.Uint64:
		return idField.Uint(), nil
	default:
		return "", fmt.Errorf("unsupported ID type: %s", idField.Kind())
	}
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

func (r *GormRepository[T]) UpdateByPK(ctx context.Context, id any, updateDto any, args ...any) error {
	return r.DB.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(updateDto).Error
}

func (r *GormRepository[T]) Update(ctx context.Context, conditions any, updateDto any, args ...any) error {
	query := r.BuildQueryConfig(ctx, conditions, nil)
	return query.Updates(updateDto).Error
}

func (r *GormRepository[T]) FindAll(ctx context.Context, conditions any, filter dto.FilterDto, config *configs.GormConfig, args ...any) ([]T, error) {
	var models []T
	listConfig := r.resolveListConfig(config)
	query := r.BuildBaseQuery(ctx, conditions, filter, listConfig)
	err := query.Find(&models).Error
	return models, err
}

func (r *GormRepository[T]) FindAllWithPaging(ctx context.Context, conditions any, filter dto.FilterDto, config *configs.GormConfig, args ...any) (*models.ListResponse[T], error) {
	var entities []T
	var total int64

	listConfig := r.resolveListConfig(config)
	query := r.BuildBaseQuery(ctx, conditions, filter, listConfig)
	countQuery := r.BuildQueryConditions(ctx, conditions, listConfig)

	if listConfig.Group != "" {
		query = query.Group(listConfig.Group)
		countQuery = countQuery.Group(listConfig.Group)
	}

	if err := countQuery.Model(new(T)).Count(&total).Error; err != nil {
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

func (r *GormRepository[T]) FindOneByPK(ctx context.Context, id any, config *configs.GormConfig, args ...any) (*T, error) {
	var model T
	query := r.BuildQueryConfig(ctx, Eq("id", id), config)
	err := query.First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &model, err
}

func (r *GormRepository[T]) FindByIDs(ctx context.Context, ids []any, config *configs.GormConfig, args ...any) ([]T, error) {
	var entities []T
	query := r.BuildQueryConfig(ctx, In("id", ids), config)
	err := query.Find(&entities).Error
	return entities, err
}

func (r *GormRepository[T]) Delete(ctx context.Context, conditions any, args ...any) error {
	return r.DB.WithContext(ctx).Model(new(T)).Where(conditions).Delete(new(T)).Error
}

func (r *GormRepository[T]) DeleteOneByPK(ctx context.Context, id any, args ...any) error {
	return r.DB.WithContext(ctx).Model(new(T)).Where("id = ?", id).Delete(new(T)).Error
}

func (r *GormRepository[T]) DeleteByIDs(ctx context.Context, ids []any, args ...any) error {
	return r.DB.WithContext(ctx).Model(new(T)).Where("id IN (?)", ids).Delete(new(T)).Error
}

func (r *GormRepository[T]) Count(ctx context.Context, conditions any, args ...any) (int64, error) {
	var count int64
	query := r.BuildQueryConditions(ctx, conditions, r.Config)
	err := query.Count(&count).Error
	return count, err
}

func (r *GormRepository[T]) Exists(ctx context.Context, conditions any, args ...any) (bool, error) {
	count, err := r.Count(ctx, conditions, args...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormRepository[T]) ExistsByPK(ctx context.Context, id any, args ...any) (bool, error) {
	return r.Exists(ctx, Eq("id", id), args...)
}

func (r *GormRepository[T]) Pluck(ctx context.Context, column string, conditions any, args ...any) ([]any, error) {
	var results []any
	query := r.BuildQueryConditions(ctx, conditions, r.Config)
	err := query.Model(new(T)).Pluck(column, &results).Error
	return results, err
}

func (r *GormRepository[T]) UpdateColumnsByPK(ctx context.Context, id any, columns map[string]any, args ...any) error {
	result := r.DB.WithContext(ctx).Model(new(T)).Where("id = ?", id).UpdateColumns(columns)
	return result.Error
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
			searchParts = append(searchParts, fmt.Sprintf("lower(%s) LIKE ?", field))
			queryValues = append(queryValues, fmt.Sprintf("%%%s%%", strings.ToLower(*filterDto.Search)))
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
			case configs.GormFilterTypeNotIn:
				queryStrings = append(queryStrings, fmt.Sprintf("%s NOT IN (?)", column))
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
				queryStrings = append(queryStrings, fmt.Sprintf("lower(%s) LIKE ?", column))
				queryValues = append(queryValues, fmt.Sprintf("%%%s%%", strings.ToLower(value.(string))))
			}
		}
	}

	finalQuery := strings.Join(queryStrings, " AND ")
	return map[string]any{
		"query": finalQuery,
		"args":  queryValues,
	}, nil
}

// resolveListConfig returns a config with ListSelectHandler/ListPreloads applied
// as overrides for list queries (FindAll, FindAllWithPaging).
func (r *GormRepository[T]) resolveListConfig(config *configs.GormConfig) *configs.GormConfig {
	var cfg configs.GormConfig
	if config == nil {
		cfg = *r.Config
	} else {
		cfg = *config
	}

	if cfg.ListSelectHandler != nil {
		cfg.SelectHandler = cfg.ListSelectHandler
	}
	if cfg.ListPreloads != nil {
		cfg.Preloads = cfg.ListPreloads
	}

	return &cfg
}

func (r *GormRepository[T]) BuildQueryConditions(ctx context.Context, conditions any, gormConfig *configs.GormConfig) *gorm.DB {
	query := r.DB.WithContext(ctx).Model(new(T))

	var config configs.GormConfig
	if gormConfig == nil {
		config = *r.Config
	} else {
		config = *gormConfig
	}

	if config.Joins != "" {
		query = query.Joins(config.Joins)
	}

	// Accept *Condition directly so callers don't need to call .Build()
	if cond, ok := conditions.(*Condition); ok {
		conditions = cond.Build()
	}

	if conditionsMap, ok := conditions.(map[string]any); ok {
		if q, ok := conditionsMap["query"].(string); ok && q != "" {
			args := conditionsMap["args"].([]interface{})
			query = query.Where(q, args...)
		}
	}
	return query
}

func (r *GormRepository[T]) BuildQueryConfig(ctx context.Context, conditions any, gormConfig *configs.GormConfig) *gorm.DB {
	var config configs.GormConfig
	if gormConfig == nil {
		config = *r.Config
	} else {
		config = *gormConfig
	}

	query := r.BuildQueryConditions(ctx, conditions, &config)

	lang := middlewares.GetLangFromContext(ctx)
	// Handle dynamic SELECTs
	if config.SelectHandler != nil {
		selects := config.SelectHandler(lang)
		var selectClauses []string
		for _, f := range selects {
			alias := f.Alias
			if alias == "" {
				alias = f.Column
			}
			selectClauses = append(selectClauses, fmt.Sprintf("%s AS %s", f.Column, alias))
		}
		query = query.Select(strings.Join(selectClauses, ", "))
	}

	// Handle dynamic Preloads
	for _, preload := range config.Preloads {
		if preload.SelectHandler != nil {
			selects := preload.SelectHandler(lang)
			var preloadClauses []string
			for _, f := range selects {
				alias := f.Alias
				if alias == "" {
					alias = f.Column
				}
				preloadClauses = append(preloadClauses, fmt.Sprintf("%s AS %s", f.Column, alias))
			}
			query = query.Preload(preload.Relation, func(db *gorm.DB) *gorm.DB {
				if preload.UnScoped {
					db = db.Unscoped()
				}
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

func (r *GormRepository[T]) BuildBaseQuery(ctx context.Context, conditions any, filter dto.FilterDto, gormConfig *configs.GormConfig) *gorm.DB {
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

// CreateOrUpdate performs an upsert operation. It creates the entity if it doesn't exist,
// or updates the specified columns if a conflict is found on the given conflictColumns.
// If updateColumns is empty, all columns are updated on conflict.
func (r *GormRepository[T]) CreateOrUpdate(ctx context.Context, entity any, conflictColumns []string, updateColumns []string, args ...any) (any, error) {
	typedEntity, ok := entity.(T)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to CreateOrUpdate: expected %T", new(T))
	}

	var onConflict clause.OnConflict
	if len(conflictColumns) > 0 {
		columns := make([]clause.Column, len(conflictColumns))
		for i, col := range conflictColumns {
			columns[i] = clause.Column{Name: col}
		}
		onConflict.Columns = columns
	}

	if len(updateColumns) > 0 {
		onConflict.DoUpdates = clause.AssignmentColumns(updateColumns)
	} else {
		onConflict.UpdateAll = true
	}

	err := r.DB.WithContext(ctx).Model(new(T)).Clauses(onConflict).Create(&typedEntity).Error
	if err != nil {
		return nil, err
	}

	val := reflect.ValueOf(typedEntity)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return nil, fmt.Errorf("ID field not found on entity")
	}
	switch idField.Kind() {
	case reflect.String:
		return idField.String(), nil
	case reflect.Int, reflect.Int64:
		return idField.Int(), nil
	case reflect.Uint, reflect.Uint64:
		return idField.Uint(), nil
	default:
		return nil, fmt.Errorf("unsupported ID type: %s", idField.Kind())
	}
}

// FindOrCreate finds the first record matching conditions, or creates a new one with createDto.
// Returns the entity and a boolean indicating whether it was created (true) or found (false).
func (r *GormRepository[T]) FindOrCreate(ctx context.Context, conditions any, createDto any, config *configs.GormConfig, args ...any) (*T, bool, error) {
	entity, ok := createDto.(T)
	if !ok {
		return nil, false, fmt.Errorf("invalid type passed to FindOrCreate: expected %T", new(T))
	}

	query := r.BuildQueryConfig(ctx, conditions, config)
	result := query.FirstOrCreate(&entity)
	if result.Error != nil {
		return nil, false, result.Error
	}

	created := result.RowsAffected > 0
	return &entity, created, nil
}

// WithTransaction executes the given function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// If the function returns nil, the transaction is committed.
func (r *GormRepository[T]) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.DB.WithContext(ctx).Transaction(fn)
}

// Restore restores a soft-deleted record by its primary key.
// This only works with models that use GORM's soft delete (DeletedAt field).
func (r *GormRepository[T]) Restore(ctx context.Context, id any, args ...any) error {
	return r.DB.WithContext(ctx).Model(new(T)).Unscoped().Where("id = ?", id).UpdateColumn("deleted_at", nil).Error
}

// RestoreByConditions restores soft-deleted records matching the given conditions.
func (r *GormRepository[T]) RestoreByConditions(ctx context.Context, conditions any, args ...any) error {
	// Accept *Condition directly
	if cond, ok := conditions.(*Condition); ok {
		conditions = cond.Build()
	}

	query := r.DB.WithContext(ctx).Model(new(T)).Unscoped()
	if conditionsMap, ok := conditions.(map[string]any); ok {
		if q, ok := conditionsMap["query"].(string); ok && q != "" {
			queryArgs := conditionsMap["args"].([]interface{})
			query = query.Where(q, queryArgs...)
		}
	}
	return query.UpdateColumn("deleted_at", nil).Error
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

func GormConditionBuilder(conditions []configs.GormQueryField) map[string]any {
	queryStrings := []string{}
	queryValues := []any{}

	for _, condition := range conditions {
		operation := condition.Operation
		if operation == "" {
			operation = "="
		}
		queryStrings = append(queryStrings, fmt.Sprintf("%s %s ?", condition.Column, operation))
		queryValues = append(queryValues, condition.Value)
	}

	finalQuery := strings.Join(queryStrings, " AND ")
	return map[string]any{
		"query": finalQuery,
		"args":  queryValues,
	}
}
