package middlewares

import (
	"github.com/gofiber/fiber/v2"

	"github.com/aghiadodeh/go-crud/models"
)

func ExceptionHandler(ctx *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	message = Translate(ctx, message, nil)

	return ctx.Status(code).JSON(models.BaseResponse[any]{
		Success:    false,
		Data:       nil,
		Message:    message,
		StatusCode: code,
	})
}
