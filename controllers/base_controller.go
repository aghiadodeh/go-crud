package controllers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/aghiadodeh/go-crud/dto"
	"github.com/aghiadodeh/go-crud/services"
)

type BaseCrudController[T any, C any, CreateDto any, UpdateDto any, FilterDto dto.FilterDto] struct {
	Service services.IBaseCrudService[T, C]
	Filter  func(ctx *fiber.Ctx) (FilterDto, error)
}

func NewBaseCrudController[T any, C any, CreateDto any, UpdateDto any, FilterDto dto.FilterDto](service services.IBaseCrudService[T, C], filter func(ctx *fiber.Ctx) (FilterDto, error)) *BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto] {
	return &BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto]{Service: service, Filter: filter}
}

func (c *BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto]) Create(ctx *fiber.Ctx) error {
	var createDto CreateDto
	if err := ctx.BodyParser(&createDto); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	item, err := c.Service.Create(ctx.UserContext(), c.MapCreateDtoToEntity(createDto), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(item)
}

func (c *BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto]) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var updateDto UpdateDto
	if err := ctx.BodyParser(&updateDto); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	item, err := c.Service.Update(ctx.UserContext(), id, updateDto, nil)
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

func (c *BaseCrudController[T, C, CreateDto, UpdateDto, FilterDto]) MapCreateDtoToEntity(createDto CreateDto) T {
	var entity T
	return entity
}
