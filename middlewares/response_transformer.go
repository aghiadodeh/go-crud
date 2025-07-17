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

	// Skip Transform
	if skip, ok := ctx.Locals("skipResponseTransform").(bool); ok && skip {
		return nil
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
	if len(originalBody) > 0 {
		if err := json.Unmarshal(originalBody, &data); err != nil {
			// fallback: treat body as raw string or binary if it's not valid JSON
			data = string(originalBody)
		}
	}

	// Check if response is already a BaseResponse
	var maybeMap map[string]interface{}
	if err := json.Unmarshal(originalBody, &maybeMap); err == nil {
		_, hasSuccess := maybeMap["success"]
		_, hasData := maybeMap["data"]
		_, hasMessage := maybeMap["message"]
		if hasSuccess && hasData && hasMessage {
			// Already in base response format
			return nil
		}
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
