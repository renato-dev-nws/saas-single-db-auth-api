package i18n

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// Supported languages
const (
	LangPtBR = "pt-BR"
	LangPt   = "pt"
	LangEn   = "en"
	LangEs   = "es"
)

// DefaultLang is the fallback language
const DefaultLang = LangPtBR

// ValidLanguages for validation
var ValidLanguages = []string{LangPtBR, LangPt, LangEn, LangEs}

// IsValidLanguage checks if a language code is supported
func IsValidLanguage(lang string) bool {
	for _, l := range ValidLanguages {
		if l == lang {
			return true
		}
	}
	return false
}

// T returns the translated message for the given key using the language from Gin context.
// Falls back to DefaultLang if language not set or key not found.
func T(c *gin.Context, key string) string {
	lang, _ := c.Get("language")
	l, _ := lang.(string)
	if l == "" {
		l = DefaultLang
	}
	return Translate(l, key)
}

// Tf returns a formatted translated message (like fmt.Sprintf).
func Tf(c *gin.Context, key string, args ...interface{}) string {
	lang, _ := c.Get("language")
	l, _ := lang.(string)
	if l == "" {
		l = DefaultLang
	}
	return Translatef(l, key, args...)
}

// Translate returns the translated message for the given key and language.
func Translate(lang, key string) string {
	if msgs, ok := messages[lang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	// Fallback: pt-BR → pt, or default
	if lang == LangPtBR {
		if msgs, ok := messages[LangPt]; ok {
			if msg, ok := msgs[key]; ok {
				return msg
			}
		}
	}
	if msgs, ok := messages[DefaultLang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	return key
}

// Translatef returns a formatted translated message.
func Translatef(lang, key string, args ...interface{}) string {
	tpl := Translate(lang, key)
	if len(args) > 0 {
		return fmt.Sprintf(tpl, args...)
	}
	return tpl
}

// TranslateValidation returns translated validation messages for a Gin context.
func TranslateValidation(c *gin.Context, field, tag, param string) string {
	lang, _ := c.Get("language")
	l, _ := lang.(string)
	if l == "" {
		l = DefaultLang
	}
	return BuildValidationMessage(l, field, tag, param)
}

// BuildValidationMessage creates a localized validation message.
func BuildValidationMessage(lang, field, tag, param string) string {
	label := FieldLabel(lang, field)
	vKey := "validation." + tag
	tpl := Translate(lang, vKey)
	if tpl == vKey {
		// Fallback
		return fmt.Sprintf("%s is invalid", label)
	}

	switch tag {
	case "required":
		return fmt.Sprintf(tpl, label)
	case "email":
		return fmt.Sprintf(tpl, label)
	case "min":
		return fmt.Sprintf(tpl, label, param)
	case "max":
		return fmt.Sprintf(tpl, label, param)
	case "url":
		return fmt.Sprintf(tpl, label)
	case "oneof":
		return fmt.Sprintf(tpl, label, param)
	case "uuid":
		return fmt.Sprintf(tpl, label)
	case "gte":
		return fmt.Sprintf(tpl, label, param)
	case "lte":
		return fmt.Sprintf(tpl, label, param)
	case "len":
		return fmt.Sprintf(tpl, label, param)
	default:
		return fmt.Sprintf(tpl, label)
	}
}

// FieldLabel returns the localized label for a field name.
func FieldLabel(lang, field string) string {
	if labels, ok := fieldLabels[lang]; ok {
		if label, ok := labels[field]; ok {
			return label
		}
	}
	if labels, ok := fieldLabels[DefaultLang]; ok {
		if label, ok := labels[field]; ok {
			return label
		}
	}
	return strings.Title(strings.ReplaceAll(field, "_", " "))
}
