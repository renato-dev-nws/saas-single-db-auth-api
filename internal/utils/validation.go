package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/saas-single-db-api/internal/i18n"
)

// fieldLabels maps JSON field names to human-readable labels
var fieldLabels = map[string]string{
	"name":             "Name",
	"email":            "Email",
	"password":         "Password",
	"plan_id":          "Plan",
	"billing_cycle":    "Billing cycle",
	"subdomain":        "Subdomain",
	"company_name":     "Company name",
	"is_company":       "Is company",
	"promo_code":       "Promo code",
	"title":            "Title",
	"slug":             "Slug",
	"role_id":          "Role",
	"permission_id":    "Permission",
	"category":         "Category",
	"status":           "Status",
	"current_password": "Current password",
	"new_password":     "New password",
	"token":            "Token",
	"tenant_id":        "Tenant",
	"full_name":        "Full name",
	"description":      "Description",
	"price":            "Price",
	"sku":              "SKU",
	"stock":            "Stock",
	"duration":         "Duration",
}

// FormatValidationErrors converts Go validator errors into a user-friendly map
// keyed by JSON field name with human-readable messages.
// If a *gin.Context is provided, messages are localized via i18n.
func FormatValidationErrors(err error, c ...*gin.Context) map[string]string {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{"_error": err.Error()}
	}

	useI18n := len(c) > 0 && c[0] != nil
	var lang string
	if useI18n {
		lang = c[0].GetString("language")
		if lang == "" {
			lang = i18n.DefaultLang
		}
	}

	result := make(map[string]string, len(errs))
	for _, fe := range errs {
		field := jsonFieldName(fe)
		var label string
		if useI18n {
			label = i18n.FieldLabel(lang, field)
		} else {
			label = fieldLabel(field)
		}
		if useI18n {
			result[field] = buildMessageI18n(c[0], label, fe.Tag(), fe.Param())
		} else {
			result[field] = buildMessage(label, fe.Tag(), fe.Param())
		}
	}
	return result
}

// jsonFieldName extracts the JSON field name from a validator.FieldError.
func jsonFieldName(fe validator.FieldError) string {
	// fe.Field() returns the struct field name; we need the json tag.
	// The namespace format is "StructName.FieldName"
	name := fe.Field()
	// Convert PascalCase to snake_case as a fallback
	return toSnakeCase(name)
}

// fieldLabel returns a human-readable label for a JSON field name.
func fieldLabel(jsonField string) string {
	if label, ok := fieldLabels[jsonField]; ok {
		return label
	}
	// Capitalize and replace underscores
	return strings.Title(strings.ReplaceAll(jsonField, "_", " "))
}

// buildMessage creates a human-readable validation message.
func buildMessage(label, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", label)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", label)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", label, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", label, param)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", label)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", label, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", label)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", label, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", label, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", label, param)
	default:
		return fmt.Sprintf("%s is invalid", label)
	}
}

// buildMessageI18n creates a localized validation message using i18n.
func buildMessageI18n(c *gin.Context, label, tag, param string) string {
	key := "validation." + tag
	switch tag {
	case "required", "email", "url", "uuid":
		return i18n.Tf(c, key, label)
	case "min", "max", "gte", "lte", "len":
		return i18n.Tf(c, key, label, param)
	case "oneof":
		return i18n.Tf(c, key, label, param)
	default:
		return i18n.Tf(c, "validation.default", label)
	}
}

// toSnakeCase converts PascalCase/camelCase to snake_case.
// Handles consecutive uppercase letters like "ID" → "id", "PlanID" → "plan_id".
func toSnakeCase(s string) string {
	var result strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				// Only add underscore if previous char is lowercase
				// OR if next char is lowercase (handles "PlanID" → "plan_id", not "plan_i_d")
				prev := runes[i-1]
				if prev >= 'a' && prev <= 'z' {
					result.WriteByte('_')
				} else if i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z' {
					result.WriteByte('_')
				}
			}
			result.WriteRune(r + 32) // toLower
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
