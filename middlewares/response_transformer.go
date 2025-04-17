package middlewares

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"github.com/aghiadodeh/go-crud/models"
)

func ResponseTransformer(ctx *fiber.Ctx) error {
	// Call next middleware/handler
	err := ctx.Next()
	if err != nil {
		return err
	}

	// Get the status code
	statusCode := ctx.Response().StatusCode()
	success := statusCode >= 200 && statusCode <= 299

	// Default message
	message := "operation_done_successfully"
	if !success {
		return nil
	}

	// Translate the message
	message = Translate(ctx, message, nil)

	// Get the original response body
	originalBody := ctx.Response().Body()
	var data any
	// Try to get the original response data if it exists
	if len(originalBody) > 0 {
		var bodyData any
		if err := json.Unmarshal(originalBody, &bodyData); err == nil {
			data = bodyData
		}
	}

	// Try to unmarshal the body to check if it's already a BaseResponse
	var baseResponse models.BaseResponse[any]
	if err := json.Unmarshal(originalBody, &baseResponse); err != nil {
		// If unmarshaling succeeds, it's likely already a BaseResponse
		// So we don't need to transform it again
		return nil
	}

	// If we got here, the response wasn't a BaseResponse, so we'll transform it
	response := models.BaseResponse[any]{
		Success:    success,
		Data:       data,
		Message:    message,
		StatusCode: statusCode,
	}

	return ctx.JSON(response)
}
