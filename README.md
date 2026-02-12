## GoLang CRUD
Manage repetitive [Fiber](https://github.com/gofiber/fiber) CRUDs Operations with a few lines of code with.

### Install:
```bash
go get github.com/aghiadodeh/go-crud
```
<hr />

### Manage Go CRUD:
This Package offer:
- Response Transformers.
- Localization.
- Base CRUD (Create, Read, Update and Delete) functionality.

### Initialize App with [Fiber](https://github.com/gofiber/fiber):
```go
package main

import (
	"log"

	"users/core"
)

func main() {
	// Initialize application
	application, err := core.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Start the application
	if err := application.Start(":8000"); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
```
```go
package core

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"users/core/database"
)

type App struct {
	fiber *fiber.App
}

func NewApp() (*App, error) {
	// Initialize database with GORM
	db, err := database.NewDBConnection()
	if err != nil {
		return nil, err
	}

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Users API",
	})

	// for testing
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	return &App{
		fiber: app,
		DB:    db,
	}, nil
}

func (a *App) Start(address string) error {
	return a.fiber.Listen(address)
}
```

#### BaseResponse:
We use a BaseResponse for transform all APIs response:
```go
type BaseResponse[T any] struct {
	Success    bool   `json:"success"`
	Data       T      `json:"data"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}
```
<hr />

## Middlewares:

This package comes with three middlewares:

### 1- Error Handler:

to transform throw errors to **BaseResponse**:

Let's throw an error with fiber:
```go
app.Get("/", func(c *fiber.Ctx) error {
	return fiber.NewError(404, "item_not_found")
})
```
The response will be a string `"item_not_found"` without using **ErrorHandler**.
##### Error Handler Middleware
```go
package core

import (
	// ...
	"github.com/aghiadodeh/go-crud/middlewares" // <-- add here
)

type App struct {
	fiber *fiber.App
}

func NewApp() (*App, error) {
	// ...

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName:           "Users API",
		ErrorHandler:      middlewares.ExceptionHandler, // <-- add here
	})

	// ...
}
```
After Add **middlewares.ExceptionHandler** The response will be like this:
```json
{
  "success": false,
  "data": null,
  "message": "item_not_found",
  "statusCode": 404
}
```

<hr />

### 2- Localization:

When You need support multi-languages client-side, you need to translate the messages which returned by server.

Previously we returned error with message `"item_not_found"`, let's handle this error with multi-languages

- Add Locales json files
```
/core
  app.go
/locales
  /ar
    i18n.ar.json
  /en
    i18n.en.json
main.go
```

- Add Locale Config
```go
import (
	// ...
	"github.com/aghiadodeh/go-crud/middlewares"
	"golang.org/x/text/language" // <-- add here
)

func NewApp() (*App, error) {
	// Initialize database
	db, err := database.NewDBConnection()
	if err != nil {
		return nil, err
	}

	app := fiber.New(fiber.Config{
		AppName:           "Users API",
		ErrorHandler:      middlewares.ExceptionHandler,
	})

	// full path of locales json files
	i18nPath := []string{
		"locales/ar/i18n.ar.json",
		"locales/en/i18n.en.json",
	}
	
	// Initialize Locale middleware with default language
	if err := middlewares.InitLocalization(language.English, i18nPath); err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}

	return &App{
		fiber: app,
		DB:    db,
	}, nil
}

```

After Add **middlewares.InitLocalization** The response will be like this:
```json
{
  "success": false,
  "data": null,
  "message": "Item not Found",
  "statusCode": 404
}
```
<hr />

### 3- Response Transformer
Let's return a simple json response with fiber:
```go
app.Get("/", func(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "Ok"})
})
```

The Response will be:
```json
{
  "status": "Ok"
}
```

To Transform the response to **BaseResponse**
```go
import (
	// ...
	"github.com/aghiadodeh/go-crud/middlewares"
)

func NewApp() (*App, error) {
	// ...

	app := fiber.New(fiber.Config{
		AppName:           "Users API",
		// ...
	})

	app.Use(middlewares.ResponseTransformer) // <-- add here

	// ...
}

```
After Add **middlewares.ResponseTransformer** The response will be like this:
```json
{
  "success": true,
  "data": {
    "status": "Ok"
  },
  "message": "Operation done successfully",
  "statusCode": 200
}
```


If You Want to Skip Transforming in Some API:
```go
func (c *AdminOrderController) ExportOrderStatistics(ctx *fiber.Ctx) error {
	ctx.Locals("skipResponseTransform", true)

	file := /* Your Excel File */

	ctx.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.xlsx"`, "Orders"))
	ctx.Set("File-Transfer-Encoding", "binary")

	// Write to buffer first
	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		return err
	}
	defer file.Close()

	// Send the buffer as response
	return ctx.SendStream(bytes.NewReader(buf.Bytes()))
}
```

<hr />

## Manage CRUDs:
This package offers Powerful & Simple functionality to make your repetitive operations easier and faster.

### 1- Declare Entity:
```go
package roles

type Role struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	NameEn string `gorm:"size:255;not null" json:"name_en,omitempty"`
	NameAr string `gorm:"size:255;not null" json:"name_ar,omitempty"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at,omitempty"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"-"`
}
```
<hr />

### 2- Declare Repository:
**BaseRepository[T]** is a generic Repository which contains the basic CRUD operations:

| # | Method | Description |
|---|--------|-------------|
| 1 | `Create` | Create a single entity |
| 2 | `BulkCreate` | Create multiple entities at once |
| 3 | `UpdateByPK` | Update entity by primary key (struct-based, skips zero values) |
| 4 | `Update` | Update entities matching conditions |
| 5 | `UpdateColumnsByPK` | Update specific columns by primary key (map-based, includes zero values) |
| 6 | `FindAll` | Find all entities matching conditions |
| 7 | `FindAllWithPaging` | Find all entities with pagination |
| 8 | `FindOne` | Find a single entity by conditions |
| 9 | `FindOneByPK` | Find a single entity by primary key |
| 10 | `FindByIDs` | Find multiple entities by a list of IDs |
| 11 | `Delete` | Delete entities matching conditions |
| 12 | `DeleteOneByPK` | Delete a single entity by primary key |
| 13 | `DeleteByIDs` | Delete multiple entities by a list of IDs |
| 14 | `Count` | Count entities matching conditions |
| 15 | `Exists` | Check if any entity matches conditions (returns bool) |
| 16 | `ExistsByPK` | Check if an entity exists by primary key (returns bool) |
| 17 | `Pluck` | Extract a single column's values from matching entities |
| 18 | `QueryBuilder` | Build query conditions from a FilterDto |

**GormRepository** also provides these GORM-specific methods (not on the interface):

| Method | Description |
|--------|-------------|
| `CreateOrUpdate` | Upsert -- insert or update on conflict |
| `FindOrCreate` | Find by conditions or create if not found |
| `WithTransaction` | Execute operations inside a database transaction |
| `Restore` | Restore a soft-deleted record by primary key |
| `RestoreByConditions` | Restore soft-deleted records matching conditions |

So, you need to create a repository that extends **BaseRepository**:

```go
package roles

import (
	"github.com/aghiadodeh/go-crud/configs"
	"github.com/aghiadodeh/go-crud/repositories"
	"gorm.io/gorm"
)

type RoleRepository interface {
	repositories.BaseRepository[
		Role, // Your Entity
		configs.GormConfig, // Repository Config (Always will be "GormConfig")
	]
}

type roleRepository struct {
	*repositories.GormRepository[Role]
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	config := configs.GormConfig{
		Model: &Role{}, // Your Entity
		DefaultSort: "created_at", // sort entities by "created_at"
	}

	return &roleRepository{
		GormRepository: repositories.NewGormRepository[Role](
			db, // gorm instance
			&config, // repository configuration
			"roles", // your entity tableName
		),
	}
}
```
#### Entity Projection:
By Default **BaseRepository** Select **`(*)`** for each `SELECT` query, but you can select specific fields as you wish by `SelectHandler` property:
```go
func NewRoleRepository(db *gorm.DB) RoleRepository {
	config := configs.GormConfig{
		Model: &Role{}, // Your Entity
		DefaultSort: "created_at", // sort entities by "created_at"
		SelectHandler: func(lang string) []configs.GormSelectField {
			// select fields as you wish
			return []configs.GormSelectField{
				{Column: "id", Alias: "id"},
				{Column: "created_at", Alias: "created_at"},
				{Column: "updated_at", Alias: "updated_at"},
			}
		},
	}

	return &roleRepository{
		GormRepository: repositories.NewGormRepository[Role](
			db, // gorm instance
			&config, // repository configuration
			"roles", // your entity tableName
		),
	}
}
```

#### Support Multi-Languages with Entity Projection:
**SelectHandler** helps you to **expose** the role name depending on client language.

##### 1- Add Context i18n middleware

**I18nMiddleware** will get the client `"accept-language"` form Http request headers and inject the language in the request context:

```go
import (
	// ...
	"github.com/aghiadodeh/go-crud/middlewares"
)

func NewApp() (*App, error) {
	// ...

	app := fiber.New(fiber.Config{
		AppName:           "Users API",
		// ...
	})

	// ("en") the default language if user didn't pass "accept-language" key in the request headers
	app.Use(middlewares.I18nMiddleware("en")) // <-- add here
	// ...
}

```
##### 2- Add Alias column to your entity
We need to add a new property to the entity:
```go
type Role struct {
	// ...
	Name   string `gorm:"->;column:name;-:migration" json:"name,omitempty"` // ignore this field by gorm on write/migration
}

```


##### 3- Expose Entity column depending on user language
```go
func NewRoleRepository(db *gorm.DB) RoleRepository {
	config := configs.GormConfig{
		// ... 
		SelectHandler: func(lang string) []configs.GormSelectField {
			// select fields as you wish
			return []configs.GormSelectField{
				// ...
				// expose Role name_en/name_ar as "name" depending on client language
				{Column: fmt.Sprintf("name_%s", lang), Alias: "name"},
			}
		},
	}

	// ...
}
```
<hr />

### 3- Declare Service:
For use the declared repository, You can declare a service that extends **GormCrudService**:
```go
package roles

import (
	"github.com/aghiadodeh/go-crud/configs"
	"github.com/aghiadodeh/go-crud/services"
)

type RoleService interface {
	services.IBaseCrudService[Role, configs.GormConfig]
	FindByName(ctx context.Context, name string) (*Role, error) // custom method
}

type roleService struct {
	*services.GormCrudService[Role]
}

func NewRoleService(repository RoleRepository) RoleService {
	return &roleService{
		GormCrudService: services.NewGormCrudService(repository),
	}
}

// Custom method example using the Condition Builder
func (s *roleService) FindByName(ctx context.Context, name string) (*Role, error) {
	return s.FindOne(ctx, repositories.Contains("name_en", name), nil)
}
```
<hr />

### 4- Declare Controller:
For use the declared service, You can declare a service that extends **GormCrudController**:
#### 1- Declare Your Create/Update Dto (Required for exposing Create/Update APIs):
```go
type RoleCreateDto struct {
	NameEN string `json:"name_en" form:"name_en" validate:"required" gorm:"column:name_en"`
	NameAR string `json:"name_ar" form:"name_ar" validate:"required" gorm:"column:name_ar"`
}

type RoleUpdateDto struct {
	NameEN *string `json:"name_en,omitempty"`
	NameAR *string `json:"name_ar,omitempty"`
}
```
Gorm requires Entity struct for create a model, so we need mapping our createDto:
```go
// override controller mapper 
func (c *RoleController) MapCreateDtoToEntity(createDto RoleCreateDto) (Role, error) {
	return Role{
		NameEn: createDto.NameEN,
		NameAr: createDto.NameAR,
	}, nil
}

func (c *RoleController) MapUpdateDtoToEntity(updateDto RoleUpdateDto) (Role, error) {
	return Role{
		NameEn: updateDto.NameEN,
		NameAr: updateDto.NameAR,
	}, nil
}


func NewRoleController(service RoleService) *RoleController {
	// ....
	controller := &RoleController{
		GormCrudController: *baseController,
		srv:                service,
	}

	controller.Mapper = controller // <-- assign mapper

	return controller
}
```
<hr />

#### 2- Declare Your Filter Dto (Required for exposing FindAll/FindOne APIs):
**FilterDto** is struct contains the base query properties for pagination and filtration:
```go
type BaseFilterDto struct {
	Page       int     `query:"page"`
	PerPage    int     `query:"per_page"`
	Pagination *bool   `query:"pagination"`
	Search     *string `query:"search"`
	SortKey    *string `query:"sort_key"`
	SortDir    *string `query:"sort_dir" validate:"omitempty,oneof=ASC DESC"`
}
```

So you need to create a new FilterDto that extends **BaseFilterDto**:
```go
import (
	"github.com/aghiadodeh/go-crud/dto"
)

type RoleFilterDto struct {
	dto.BaseFilterDto // include BaseFilterDto

	// filter fields must be included in the filterDto
	NameEN *string `query:"name_en"` 
	NameAR *string `query:"name_ar"`
}

// ToFilterMap converts the filter to a map for repository
func (f *RoleFilterDto) ToMap() (map[string]interface{}, error) {
	filters := f.BaseFilterDto.ToMapNoError()

	if f.NameEN != nil {
		filters["name_en"] = *f.NameEN
	}

	if f.NameAR != nil {
		filters["name_ar"] = *f.NameAR
	}

	return filters, nil
}
```
<hr />

#### 3- Add FilterDto Properties to Repository:
You need tell your repository for filterable columns which declared in filterDto:
```go
func NewRoleRepository(db *gorm.DB) RoleRepository {
	config := configs.GormConfig{
		// ...

		// all properties must be declared in the filterDto
		Filterable: map[string]configs.GormFilterProperty{
			"name_en": {FilterType: configs.GormFilterTypeRegex},
			"name_ar": {FilterType: configs.GormFilterTypeRegex},
		},

		// searchable columns, matching with `search` queryParam value
		Searchable:  []string{"name_en", "name_ar"}, 
		// ...
	}

	return &roleRepository{
		GormRepository: repositories.NewGormRepository[Role](
			db,
			&config,
			"roles",
		),
	}
}
```

You can check filtering types with **GormFilterType**:
```go
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
```

<hr />

## Condition Builder:

Instead of writing raw SQL query strings, use the **fluent Condition Builder** to construct type-safe, readable query conditions.

### Before vs After:
```go
// Before (raw strings, error-prone):
conditions := map[string]any{
    "query": "status = ? AND age >= ?",
    "args":  []any{"active", 18},
}

// After (fluent builder):
conditions := repositories.Eq("status", "active").
    And(repositories.Gte("age", 18))
```

Pass `*Condition` directly to any repository/service method -- no `.Build()` needed:
```go
repo.FindOne(ctx, repositories.Eq("email", email), config)
repo.FindAll(ctx, repositories.Eq("active", true), filter, config)
repo.Count(ctx, repositories.Gt("age", 18))
repo.Delete(ctx, repositories.In("id", expiredIDs))
repo.Exists(ctx, repositories.Eq("username", name))
```

### Available Constructors:

| Function | SQL Generated |
|----------|---------------|
| `Eq(col, val)` | `col = ?` |
| `NotEq(col, val)` | `col != ?` |
| `Gt(col, val)` | `col > ?` |
| `Gte(col, val)` | `col >= ?` |
| `Lt(col, val)` | `col < ?` |
| `Lte(col, val)` | `col <= ?` |
| `In(col, vals)` | `col IN (?)` |
| `NotIn(col, vals)` | `col NOT IN (?)` |
| `Like(col, pattern)` | `col LIKE ?` |
| `ILike(col, pattern)` | `LOWER(col) LIKE ?` (auto-lowercased) |
| `Contains(col, val)` | `LOWER(col) LIKE '%val%'` |
| `StartsWith(col, val)` | `LOWER(col) LIKE 'val%'` |
| `EndsWith(col, val)` | `LOWER(col) LIKE '%val'` |
| `IsNull(col)` | `col IS NULL` |
| `IsNotNull(col)` | `col IS NOT NULL` |
| `Between(col, lo, hi)` | `col BETWEEN ? AND ?` |
| `NotBetween(col, lo, hi)` | `col NOT BETWEEN ? AND ?` |
| `Raw(sql, args...)` | any custom SQL fragment |

### Chaining with And / Or:

```go
// AND: status = 'active' AND age >= 18 AND role IN ('admin','editor')
cond := repositories.Eq("status", "active").
    And(repositories.Gte("age", 18)).
    And(repositories.In("role", []string{"admin", "editor"}))

// OR: role = 'admin' OR role = 'moderator'
cond := repositories.Eq("role", "admin").
    Or(repositories.Eq("role", "moderator"))
```

### Grouped Sub-Conditions:

Nested conditions are automatically wrapped in parentheses:

```go
// status = 'active' AND (role = 'admin' OR role = 'moderator')
cond := repositories.Eq("status", "active").
    And(
        repositories.Eq("role", "admin").Or(repositories.Eq("role", "moderator")),
    )
```

### More Examples:

```go
// NULL checks
cond := repositories.IsNull("deleted_at")

// BETWEEN
cond := repositories.Between("created_at", startDate, endDate)

// Case-insensitive search
cond := repositories.Contains("name", "john") // LOWER(name) LIKE '%john%'

// Escape hatch for complex SQL
cond := repositories.Raw("json_extract(data, '$.role') = ?", "admin")
```

<hr />

## GORM-Specific Methods:

### Upsert (CreateOrUpdate):
Insert a record, or update specific columns if a conflict is detected:
```go
id, err := repo.CreateOrUpdate(ctx, role, 
    []string{"email"},           // conflict columns
    []string{"name", "status"},  // columns to update on conflict (empty = update all)
)
```

### FindOrCreate:
Atomically find a record by conditions, or create it if it doesn't exist:
```go
entity, created, err := repo.FindOrCreate(ctx, 
    repositories.Eq("email", "user@example.com"), // conditions
    newUser,  // entity to create if not found
    config,
)
if created {
    fmt.Println("New user created")
}
```

### Transactions (WithTransaction):
Execute multiple operations inside a single database transaction:
```go
err := repo.WithTransaction(ctx, func(tx *gorm.DB) error {
    if err := tx.Create(&order).Error; err != nil {
        return err // triggers rollback
    }
    if err := tx.Create(&orderItems).Error; err != nil {
        return err // triggers rollback
    }
    return nil // triggers commit
})
```

### Soft Delete Restore:
Restore a soft-deleted record:
```go
// Restore by primary key
err := repo.Restore(ctx, id)

// Restore by conditions
err := repo.RestoreByConditions(ctx, repositories.Eq("email", email))
```

<hr />

#### 4- Declare Your Controller:
```go
package roles

import (
	// ...
	"github.com/aghiadodeh/go-crud/controllers"
	"github.com/gofiber/fiber/v2"
)

type RoleController struct {
	controllers.GormCrudController[Role, RoleCreateDto, RoleUpdateDto, *RoleFilterDto]
	srv RoleService
}

func NewRoleController(service RoleService) *RoleController {
	baseController := controllers.NewGormBaseController[Role, RoleCreateDto, RoleUpdateDto](
		service,
		func(ctx *fiber.Ctx) (*RoleFilterDto, error) {
			// parsing `RoleFilterDto` to queryParams
			var filterDto RoleFilterDto
			if err := ctx.QueryParser(&filterDto); err != nil {
				return nil, fiber.NewError(fiber.ErrBadRequest.Code, err.Error())
			}
			return &filterDto, nil
		},
	)
	controller := &RoleController{
		GormCrudController: *baseController,
		srv:                service,
	}
	controller.Mapper = controller // <-- assign mapper (if you expose create/update in the controller)
	return controller
}
```
<hr />

#### 5- Override Methods:
Usually, you need filtering on some data that is not sent in filterDto, like returning only posts which belong to a user.

Here is an example of overriding `FindAll` method in `UserController` using the **Condition Builder**:

```go
func (c *UserController) FindAll(ctx *fiber.Ctx) error {
	// parsing filterDto
	filter, err := c.Filter(ctx)
	if err != nil {
		return fiber.NewError(fiber.ErrBadRequest.Code, err.Error())
	}

	// BuildQuery depending on filterDto & repository config
	filterDto := filter.GetBase()
	conditions, err := c.Service.QueryBuilder(ctx.UserContext(), filter, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Inject custom filters using the Condition Builder
	builtConditions := conditions.(map[string]any)
	extraConditions := repositories.Eq("role_id", 1).
		And(repositories.Contains("username", "go_user"))

	// Merge the QueryBuilder output with your custom conditions
	extraBuilt := extraConditions.Build()
	if query, ok := builtConditions["query"].(string); ok && query != "" {
		builtConditions["query"] = query + " AND " + extraBuilt["query"].(string)
		builtConditions["args"] = append(builtConditions["args"].([]any), extraBuilt["args"].([]any)...)
	} else {
		builtConditions = extraBuilt
	}
	conditions = builtConditions

	if filterDto.Pagination == nil || *filterDto.Pagination {
		// get data with pagination
		response, err := c.Service.FindAllWithPaging(ctx.UserContext(), conditions, filter, nil)
		if err != nil {
			return err
		}
		return ctx.JSON(response)
	}

	// get data without pagination
	items, err := c.Service.FindAll(ctx.UserContext(), conditions, filter, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(items)
}
```

For simpler cases where you don't use `QueryBuilder`, you can pass conditions directly:

```go
// Find all active users with a specific role
items, err := service.FindAll(ctx,
	repositories.Eq("active", true).And(repositories.Eq("role_id", 1)),
	filter, nil,
)

// Check if username is taken
taken, err := service.Exists(ctx, repositories.Eq("username", "john"))

// Get all emails for users in a department
emails, err := service.Pluck(ctx, "email", repositories.Eq("department_id", deptID))

// Batch delete expired sessions
err := service.DeleteByIDs(ctx, expiredSessionIDs)

// Update specific columns (includes zero values, unlike UpdateByPK)
err := service.UpdateColumnsByPK(ctx, userID, map[string]any{
	"login_count": 0,
	"last_login":  time.Now(),
})
```
<hr />

### **GormCrudController** extends **BaseCrudController** which offers these functions:

| Function  | Function Parameters  | Parsing Data from  | Response  |
|:----------|:----------|:----------|:----------|
| Create    | `*fiber.Ctx`   | **Body**    | `T`    |
| Update    | `*fiber.Ctx`    | `id` from **Params**,<br /> `updateDto` from **Body**  | `T`   |
| FindAll    | `*fiber.Ctx`    | **Query**    | `T[]` / `ListResponse[T]`    |
| FindOne    | `*fiber.Ctx`    | `id` from **Params**    | `T`    |
| Delete    | `*fiber.Ctx`    | `id` from **Params**    | `null`    |

### **IBaseCrudService** provides these methods:

| Method | Description |
|--------|-------------|
| `Create` | Create entity and return it with full config (preloads, selects) |
| `Update` | Update entity by PK and return the updated entity |
| `UpdateColumnsByPK` | Update specific columns by primary key |
| `FindAll` | Find all entities matching conditions |
| `FindAllWithPaging` | Find all entities with pagination response |
| `FindOne` | Find a single entity by conditions |
| `FindOneByPK` | Find a single entity by primary key |
| `FindByIDs` | Find multiple entities by a list of IDs |
| `Delete` | Delete entities matching conditions |
| `DeleteOneByPK` | Delete a single entity by primary key |
| `DeleteByIDs` | Delete multiple entities by a list of IDs |
| `Count` | Count entities matching conditions |
| `Exists` | Check existence by conditions (returns `bool`) |
| `ExistsByPK` | Check existence by primary key (returns `bool`) |
| `Pluck` | Extract a single column from matching entities |
| `QueryBuilder` | Build query conditions from a FilterDto |

<hr />

