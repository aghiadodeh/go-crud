package controllers

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/aghiadodeh/go-crud/dto"
	"github.com/aghiadodeh/go-crud/services"
)

type BaseCrudController[T any, C any, CreateDto any, UpdateDto any, FilterDto dto.FilterDto] struct {
	Service services.IBaseCrudService[T, C]
	Filter  func(ctx *fiber.Ctx) (FilterDto, error)
	Mapper  CreateDtoMapper[CreateDto, UpdateDto, T]
}

func NewBaseCrudController[T any, C any, CreateDto any, UpdateDto any, FilterDto dto.FilterDto](service services.IBaseCrudService[T, C], filter func(ctx *fiber.Ctx) (FilterDto, error)) *BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto] {
	return &BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto]{Service: service, Filter: filter}
}

func (c *BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto]) Create(ctx *fiber.Ctx) error {
	var createDto CreateDto

	// 1. Try parsing JSON
	if err := ctx.BodyParser(&createDto); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// 2. Validate parsed data
	var validate = validator.New()
	if err := validate.Struct(createDto); err != nil {
		// Collect error messages
		var messages []string
		for _, err := range err.(validator.ValidationErrors) {
			messages = append(messages, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		return fiber.NewError(fiber.StatusBadRequest, strings.Join(messages, ", "))
	}

	// 3. Map Dto to Entity
	if c.Mapper == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "No Mapper")
	}
	entity := c.Mapper.MapCreateDtoToEntity(createDto)

	// 4. Continue to business logic
	item, err := c.Service.Create(ctx.UserContext(), entity, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(item)
}

func (c *BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto]) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var updateDto UpdateDto

	// 1. Try parsing JSON
	if err := ctx.BodyParser(&updateDto); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// 2. Validate parsed data
	var validate = validator.New()
	if err := validate.Struct(updateDto); err != nil {
		// Collect error messages
		var messages []string
		for _, err := range err.(validator.ValidationErrors) {
			messages = append(messages, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		return fiber.NewError(fiber.StatusBadRequest, strings.Join(messages, ", "))
	}

	// 3. Map Dto to Entity
	if c.Mapper == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "No Mapper")
	}
	entity := c.Mapper.MapUpdateDtoToEntity(updateDto)

	// 4. Continue to business logic
	item, err := c.Service.Update(ctx.UserContext(), id, entity, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if item == nil {
		return fiber.NewError(fiber.ErrNotFound.Code, "item_not_found")
	}

	return ctx.JSON(item)
}

func (c *BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto]) FindAll(ctx *fiber.Ctx) error {
	filter, err := c.Filter(ctx)
	if err != nil {
		return fiber.NewError(fiber.ErrBadRequest.Code, err.Error())
	}

	if c.Service == nil {
		return fiber.NewError(fiber.ErrBadRequest.Code, "BaseService is null, check controller injection")
	}

	filterDto := filter.GetBase()
	conditions, err := c.Service.QueryBuilder(ctx.UserContext(), filter, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if filterDto.Pagination == nil || *filterDto.Pagination {
		response, err := c.Service.FindAllWithPaging(ctx.UserContext(), conditions, filter, nil)
		if err != nil {
			return err
		}
		return ctx.JSON(response)
	}

	items, err := c.Service.FindAll(ctx.UserContext(), conditions, filter, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(items)
}

func (c *BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto]) FindOne(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	item, err := c.Service.FindOneByPK(ctx.UserContext(), id, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if item == nil {
		return fiber.ErrNotFound
	}
	return ctx.JSON(item)
}

func (c *BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto]) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.Service.DeleteOneByPK(ctx.UserContext(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(nil)
}

type CreateDtoMapper[CreateDto any, UpdateDto any, T any] interface {
	MapCreateDtoToEntity(createDto CreateDto) T
	MapUpdateDtoToEntity(updateDto UpdateDto) T
}
