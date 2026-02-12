package repositories

import (
	"fmt"
	"strings"
)

// Condition represents a composable query condition with a fluent builder API.
//
// Use the package-level constructor functions (Eq, Gt, In, etc.) to start building conditions,
// then chain with And/Or to compose complex queries.
//
// Example usage:
//
//	// Simple equality
//	cond := repositories.Eq("status", "active")
//
//	// Multiple conditions (AND)
//	cond := repositories.Eq("status", "active").
//	    And(repositories.Gte("age", 18)).
//	    And(repositories.In("role", []string{"admin", "editor"}))
//
//	// OR conditions
//	cond := repositories.Eq("role", "admin").
//	    Or(repositories.Eq("role", "moderator"))
//
//	// Grouped sub-conditions: status = 'active' AND (role = 'admin' OR role = 'moderator')
//	cond := repositories.Eq("status", "active").
//	    And(
//	        repositories.Eq("role", "admin").Or(repositories.Eq("role", "moderator")),
//	    )
//
//	// NULL checks, BETWEEN, LIKE
//	cond := repositories.IsNull("deleted_at")
//	cond := repositories.Between("age", 18, 65)
//	cond := repositories.Contains("name", "john")
//
//	// Pass directly to any repository method that accepts conditions:
//	repo.FindOne(ctx, repositories.Eq("email", email), config)
//	repo.FindAll(ctx, repositories.Eq("active", true), filter, config)
//	repo.Count(ctx, repositories.Gt("age", 18))
//	repo.Delete(ctx, repositories.In("id", expiredIDs))
type Condition struct {
	parts []conditionPart
}

type conditionPart struct {
	connector string     // "" for the first part, "AND" or "OR" for subsequent parts
	fragment  string     // SQL fragment like "status = ?"
	args      []any      // bind values for this fragment
	group     *Condition // nested group (if set, fragment/args are ignored)
}

// --- Constructor functions (start a new condition) ---

// Eq creates a condition: column = value
func Eq(column string, value any) *Condition {
	return newLeaf(fmt.Sprintf("%s = ?", column), value)
}

// NotEq creates a condition: column != value
func NotEq(column string, value any) *Condition {
	return newLeaf(fmt.Sprintf("%s != ?", column), value)
}

// Gt creates a condition: column > value
func Gt(column string, value any) *Condition {
	return newLeaf(fmt.Sprintf("%s > ?", column), value)
}

// Gte creates a condition: column >= value
func Gte(column string, value any) *Condition {
	return newLeaf(fmt.Sprintf("%s >= ?", column), value)
}

// Lt creates a condition: column < value
func Lt(column string, value any) *Condition {
	return newLeaf(fmt.Sprintf("%s < ?", column), value)
}

// Lte creates a condition: column <= value
func Lte(column string, value any) *Condition {
	return newLeaf(fmt.Sprintf("%s <= ?", column), value)
}

// In creates a condition: column IN (values)
func In(column string, values any) *Condition {
	return newLeaf(fmt.Sprintf("%s IN (?)", column), values)
}

// NotIn creates a condition: column NOT IN (values)
func NotIn(column string, values any) *Condition {
	return newLeaf(fmt.Sprintf("%s NOT IN (?)", column), values)
}

// Like creates a condition: column LIKE pattern
//
// You provide the full pattern including wildcards:
//
//	Like("name", "%john%")   // contains "john"
//	Like("name", "john%")    // starts with "john"
func Like(column string, pattern string) *Condition {
	return newLeaf(fmt.Sprintf("%s LIKE ?", column), pattern)
}

// ILike creates a case-insensitive LIKE: LOWER(column) LIKE pattern
//
// The pattern is automatically lowercased:
//
//	ILike("name", "%John%")  // matches "john", "JOHN", "John", etc.
func ILike(column string, pattern string) *Condition {
	return newLeaf(fmt.Sprintf("LOWER(%s) LIKE ?", column), strings.ToLower(pattern))
}

// Contains creates a case-insensitive substring search.
//
// Equivalent to: LOWER(column) LIKE '%value%'
//
//	Contains("name", "john")  // matches "John Doe", "JOHNNY", etc.
func Contains(column string, value string) *Condition {
	return newLeaf(
		fmt.Sprintf("LOWER(%s) LIKE ?", column),
		fmt.Sprintf("%%%s%%", strings.ToLower(value)),
	)
}

// StartsWith creates a case-insensitive prefix search.
//
// Equivalent to: LOWER(column) LIKE 'value%'
func StartsWith(column string, value string) *Condition {
	return newLeaf(
		fmt.Sprintf("LOWER(%s) LIKE ?", column),
		fmt.Sprintf("%s%%", strings.ToLower(value)),
	)
}

// EndsWith creates a case-insensitive suffix search.
//
// Equivalent to: LOWER(column) LIKE '%value'
func EndsWith(column string, value string) *Condition {
	return newLeaf(
		fmt.Sprintf("LOWER(%s) LIKE ?", column),
		fmt.Sprintf("%%%s", strings.ToLower(value)),
	)
}

// IsNull creates a condition: column IS NULL
func IsNull(column string) *Condition {
	return &Condition{
		parts: []conditionPart{
			{fragment: fmt.Sprintf("%s IS NULL", column)},
		},
	}
}

// IsNotNull creates a condition: column IS NOT NULL
func IsNotNull(column string) *Condition {
	return &Condition{
		parts: []conditionPart{
			{fragment: fmt.Sprintf("%s IS NOT NULL", column)},
		},
	}
}

// Between creates a condition: column BETWEEN low AND high
func Between(column string, low, high any) *Condition {
	return &Condition{
		parts: []conditionPart{
			{fragment: fmt.Sprintf("%s BETWEEN ? AND ?", column), args: []any{low, high}},
		},
	}
}

// NotBetween creates a condition: column NOT BETWEEN low AND high
func NotBetween(column string, low, high any) *Condition {
	return &Condition{
		parts: []conditionPart{
			{fragment: fmt.Sprintf("%s NOT BETWEEN ? AND ?", column), args: []any{low, high}},
		},
	}
}

// Raw creates a condition from a raw SQL fragment with optional bind values.
// Use this as an escape hatch for complex expressions not covered by the builder.
//
//	Raw("age > ? AND age < ?", 18, 65)
//	Raw("json_extract(data, '$.role') = ?", "admin")
func Raw(fragment string, args ...any) *Condition {
	return &Condition{
		parts: []conditionPart{
			{fragment: fragment, args: args},
		},
	}
}

// --- Chaining methods ---

// And appends another condition with AND logic.
//
//	Eq("status", "active").And(Gte("age", 18))
//	// => status = ? AND age >= ?
func (c *Condition) And(other *Condition) *Condition {
	if other == nil {
		return c
	}
	c.parts = append(c.parts, conditionPart{
		connector: "AND",
		group:     other,
	})
	return c
}

// Or appends another condition with OR logic.
//
//	Eq("role", "admin").Or(Eq("role", "moderator"))
//	// => role = ? OR role = ?
func (c *Condition) Or(other *Condition) *Condition {
	if other == nil {
		return c
	}
	c.parts = append(c.parts, conditionPart{
		connector: "OR",
		group:     other,
	})
	return c
}

// --- Build / Output ---

// Build compiles the condition tree into the map[string]any format
// used internally by GormRepository:
//
//	map[string]any{"query": "status = ? AND age >= ?", "args": []any{"active", 18}}
//
// You typically don't need to call Build() directly -- pass the *Condition
// straight to repository methods. This method is available for manual use
// or interoperability with code that expects the raw map format.
func (c *Condition) Build() map[string]any {
	if c == nil || len(c.parts) == 0 {
		return map[string]any{
			"query": "",
			"args":  []any{},
		}
	}

	query, args := c.compile()
	return map[string]any{
		"query": query,
		"args":  args,
	}
}

// compile recursively builds the SQL string and args from the condition tree.
func (c *Condition) compile() (string, []any) {
	if c == nil || len(c.parts) == 0 {
		return "", nil
	}

	var segments []string
	var allArgs []any

	for i, part := range c.parts {
		var fragment string
		var args []any

		if part.group != nil {
			fragment, args = part.group.compile()
			if fragment == "" {
				continue
			}
			// Wrap nested groups in parentheses when they contain mixed logic
			if len(part.group.parts) > 1 {
				fragment = "(" + fragment + ")"
			}
		} else {
			fragment = part.fragment
			args = part.args
		}

		if i > 0 && part.connector != "" {
			segments = append(segments, part.connector)
		}
		segments = append(segments, fragment)
		allArgs = append(allArgs, args...)
	}

	return strings.Join(segments, " "), allArgs
}

// --- Internal helpers ---

func newLeaf(fragment string, args ...any) *Condition {
	return &Condition{
		parts: []conditionPart{
			{fragment: fragment, args: args},
		},
	}
}
