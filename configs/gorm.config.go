package configs

type GormPropertyType string

const (
	GormPropertyTypeDate   GormPropertyType = "date"
	GormPropertyTypeNormal GormPropertyType = "normal"
)

type GormFilterType string

const (
	GormFilterTypeEqual GormFilterType = "equal"
	GormFilterTypeIn    GormFilterType = "in"
	GormFilterTypeNotIn GormFilterType = "not_in"
	GormFilterTypeLT    GormFilterType = "lt"
	GormFilterTypeGT    GormFilterType = "gt"
	GormFilterTypeLTE   GormFilterType = "lte"
	GormFilterTypeGTE   GormFilterType = "gte"
	GormFilterTypeRegex GormFilterType = "regex"
)

type GormRegexConcat struct {
	Keys      []string
	Separator *string
}

type GormFilterProperty struct {
	ColumnName string
	FilterType GormFilterType
}

type GormSearchProperty struct {
	Key string
}

type GormConfig struct {
	Model         interface{}
	Filterable    map[string]GormFilterProperty
	Searchable    []string
	DefaultSort   string
	SelectHandler func(lang string) []GormSelectField
	Preloads      []GormPreloadConfig
	Joins         string
	UnScoped      bool
	Group         string

	// List-specific overrides (used by FindAll/FindAllWithPaging only).
	// When set, these take precedence over SelectHandler/Preloads for list queries.
	ListSelectHandler func(lang string) []GormSelectField
	ListPreloads      []GormPreloadConfig
}

type GormSelectField struct {
	Column string
	Alias  string
}

type GormPreloadConfig struct {
	Relation      string
	UnScoped      bool
	SelectHandler func(lang string) []GormSelectField
}

type GormQueryField struct {
	Operation string
	Column    string
	Value     any
}

// Implement RepositoryConfig interface
func (c *GormConfig) IsRepositoryConfig() {}
