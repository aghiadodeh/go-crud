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

This package come with three middlwares:

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
1. Create
1. BulkCreate
1. UpdateByPK
1. Update
1. FindAll
1. FindAllWithPaging
1. FindOne
1. FindOneByPK
1. Delete
1. DeleteOneByPK
1. Count

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

**I18nMiddleware** will get the client `"language"` form Http request headers and inject the language in the request context:

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

	// ("en") the default language if user didn't pass "language" key in the request headers
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

// Custom method example
func (s *roleService) FindByName(ctx context.Context, name string) (*Role, error) {
	return s.FindOne(ctx, map[string]any{"name_en LIKE ?": name}, nil)
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
func (c *RoleController) MapCreateDtoToEntity(createDto RoleCreateDto) Role {
	return Role{
		NameEn: createDto.NameEN,
		NameAr: createDto.NameAR,
	}
}

func (c *RoleController) MapUpdateDtoToEntity(updateDto RoleUpdateDto) Role {
	return Role{
		NameEn: updateDto.NameEN,
		NameAr: updateDto.NameAR,
	}
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
	GormFilterTypeLT    GormFilterType = "lt"
	GormFilterTypeGT    GormFilterType = "gt"
	GormFilterTypeLTE   GormFilterType = "lte"
	GormFilterTypeGTE   GormFilterType = "gte"
	GormFilterTypeRegex GormFilterType = "regex"
)
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

#### 5- override methods:
Usually, You need filtering on some data that not sent if filterDto, like return only posts which belong to user.

Here is Example for overriding `FindAll` method in `UserController`

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

	// inject your custom filters
	if conditionsMap, ok := conditions.(map[string]any); ok {
		queryStr, _ := conditionsMap["query"].(string)
		args, _ := conditionsMap["args"].([]any)
		if args == nil {
			args = []any{}
		}

		// Add your custom filters
		queryStr, args = appendCondition(queryStr, args, "role_id = ?", 1)
		queryStr, args = appendCondition(queryStr, args, "username LIKE ?", "go_user")

		// Save back
		conditionsMap["query"] = queryStr
		conditionsMap["args"] = args
		conditions = conditionsMap
	}

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

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

func appendCondition(query string, args []any, condition string, value any) (string, []any) {
	if query != "" {
		query += " AND " + condition
	} else {
		query = condition
	}
	args = append(args, value)
	return query, args
}
```
<hr />

### **GormCrudController** is extends **BaseCrudController** which offers these functions:

Function, Function Parameters, Parsing Data from, Response

| Function  | Function Parameters  | Parsing Data from  | Response  |
|:----------|:----------|:----------|:----------|
| Create    | `*fiber.Ctx`   | **Body**    | T    |
| Update    | `*fiber.Ctx`    | `id` from **Params**,<br /> `updateDto` from **Body**  | T   |
| FindAll    | `*fiber.Ctx`    | Query    | `T[]` / `ListResponse[T]`    |
| FindOne    | `*fiber.Ctx`    | `id` from **Params**    | `T`    |
| Delete    | `*fiber.Ctx`    | `id` from **Params**    | `null`    |

<hr />

