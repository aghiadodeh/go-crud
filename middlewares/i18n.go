package middlewares

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type ctxKey string

const LangContextKey ctxKey = "lang"

const (
	ContextLanguage = "lang"
)

var (
	bundle *i18n.Bundle
)

func InitLocalization(defaultLanguage language.Tag, assets []string) error {
	// Initialize bundle
	bundle = i18n.NewBundle(defaultLanguage)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load translations
	for _, path := range assets {
		// Determine file type from extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".json" {
			continue // skip unsupported file types
		}

		// Load the message file
		if _, err := bundle.LoadMessageFile(path); err != nil {
			return err
		}
	}

	return nil
}

// I18nMiddleware sets up the i18n localizer for each request
func I18nMiddleware(defaultLanguage string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get language from header, query, cookie, or default
		lang := c.Get("Accept-Language", defaultLanguage)
		if l := c.Query("lang"); l != "" {
			lang = l
		}
		if l := c.Cookies("lang"); l != "" {
			lang = l
		}

		// You can parse lang here more robustly if needed
		ctx := context.WithValue(c.UserContext(), LangContextKey, lang)
		c.SetUserContext(ctx)

		// Store localizer in context
		localizer := i18n.NewLocalizer(bundle, lang)
		c.Locals(ContextLanguage, localizer)

		return c.Next()
	}
}

func GetLangFromContext(ctx context.Context) string {
	if l, ok := ctx.Value(LangContextKey).(string); ok {
		return l
	}
	return "en"
}

// GetLocalizer retrieves the localizer from context
func GetLocalizer(c *fiber.Ctx) *i18n.Localizer {
	if l, ok := c.Locals(ContextLanguage).(*i18n.Localizer); ok {
		return l
	}
	return i18n.NewLocalizer(bundle)
}

// Translate translates a message using the context's localizer
func Translate(c *fiber.Ctx, messageID string, templateData map[string]interface{}) string {
	localizer := GetLocalizer(c)

	// Try to find message in i18n namespace first
	if translated, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    "i18n." + messageID,
		TemplateData: templateData,
	}); err == nil {
		return translated
	}

	// Fallback to regular message
	translated, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		return messageID
	}
	return translated
}
