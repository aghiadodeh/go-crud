package dto

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type BaseFilterDto struct {
	Page       int     `query:"page"`
	PerPage    int     `query:"per_page"`
	Pagination *bool   `query:"pagination"`
	Search     *string `query:"search"`
	SortKey    *string `query:"sort_key"`
	SortDir    *string `query:"sort_dir" validate:"omitempty,oneof=ASC DESC"`
}

type FilterDto interface {
	GetBase() *BaseFilterDto
	ToMap() (map[string]interface{}, error)
}

func (f *BaseFilterDto) GetBase() *BaseFilterDto {
	return f
}

// ToMap converts the filter to a map of filter conditions
func (f *BaseFilterDto) ToMap() (map[string]interface{}, error) {
	m := map[string]interface{}{}
	if f.Search != nil {
		m["search"] = *f.Search
	}
	if f.SortKey != nil {
		m["sort_key"] = *f.SortKey
	}
	if f.SortDir != nil {
		m["sort_dir"] = *f.SortDir
	}
	return m, nil
}

func (f *BaseFilterDto) ToMapNoError() map[string]interface{} {
	m, _ := f.ToMap()
	return m
}

func (f *BaseFilterDto) BindQuery(c *fiber.Ctx) error {
	// Parse primitive values
	f.Page, _ = strconv.Atoi(c.Query("page", "1"))
	f.PerPage, _ = strconv.Atoi(c.Query("per_page", "10"))

	// Optional booleans
	if val := c.Query("pagination"); val != "" {
		boolVal := strings.ToLower(val) == "true"
		f.Pagination = &boolVal
	}

	if search := c.Query("search"); search != "" {
		f.Search = &search
	}
	if sortKey := c.Query("sort_key"); sortKey != "" {
		f.SortKey = &sortKey
	}
	if sortDir := c.Query("sort_dir"); sortDir != "" {
		f.SortDir = &sortDir
	}
	return nil
}
