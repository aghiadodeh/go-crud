package controllers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/aghiadodeh/go-crud/configs"
	"github.com/aghiadodeh/go-crud/dto"
	"github.com/aghiadodeh/go-crud/services"
)

type GormCrudController[T any, CreateDto any, UpdateDto any, FilterDto dto.FilterDto] struct {
	BaseCrudController[T, configs.GormConfig, CreateDto, UpdateDto, FilterDto]
}

func NewGormBaseController[T any, CreateDto any, UpdateDto any, FilterDto dto.FilterDto](service services.IBaseCrudService[T, configs.GormConfig], filter func(ctx *fiber.Ctx) (FilterDto, error)) *GormCrudController[T, CreateDto, UpdateDto, FilterDto] {
	return &GormCrudController[T, CreateDto, UpdateDto, FilterDto]{
		BaseCrudController: *NewBaseCrudController[T, configs.GormConfig, CreateDto, UpdateDto](service, filter),
	}
}
