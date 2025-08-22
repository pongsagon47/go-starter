package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

var (
	validate  *validator.Validate
	sanitizer *XSSSanitizer
)

// XSSSanitizer handles XSS sanitization with different policies
type XSSSanitizer struct {
	strictPolicy *bluemonday.Policy // No HTML allowed
	ugcPolicy    *bluemonday.Policy // User Generated Content - basic HTML
	richPolicy   *bluemonday.Policy // Rich text editor - more HTML tags
}

func init() {
	validate = validator.New()
	sanitizer = NewXSSSanitizer()

	// Register custom tag name func for better field names in errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// NewXSSSanitizer creates a new XSS sanitizer with different policies
func NewXSSSanitizer() *XSSSanitizer {
	return &XSSSanitizer{
		strictPolicy: bluemonday.StrictPolicy(),
		ugcPolicy:    bluemonday.UGCPolicy(),
		richPolicy:   createRichTextPolicy(),
	}
}

// createRichTextPolicy creates a policy suitable for rich text editors
func createRichTextPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()

	// Text formatting
	p.AllowElements("b", "i", "u", "s", "strong", "em", "mark", "small", "del", "ins", "sub", "sup")

	// Headings
	p.AllowElements("h1", "h2", "h3", "h4", "h5", "h6")

	// Paragraphs and line breaks
	p.AllowElements("p", "br", "hr")

	// Lists
	p.AllowElements("ul", "ol", "li")

	// Links with safe attributes
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("title").OnElements("a")
	p.RequireNoReferrerOnLinks(true)

	// Images with safe attributes
	p.AllowAttrs("src", "alt", "title", "width", "height").OnElements("img")

	// Tables
	p.AllowElements("table", "thead", "tbody", "tfoot", "tr", "th", "td", "caption")
	p.AllowAttrs("colspan", "rowspan").OnElements("th", "td")

	// Blockquotes and code
	p.AllowElements("blockquote", "code", "pre")

	// Divs and spans with limited attributes
	p.AllowElements("div", "span")

	// Style attributes for basic formatting (limited)
	p.AllowAttrs("style").OnElements("p", "div", "span", "h1", "h2", "h3", "h4", "h5", "h6")
	p.AllowStyles("color", "background-color", "font-size", "font-weight", "text-align", "text-decoration").OnElements("p", "div", "span", "h1", "h2", "h3", "h4", "h5", "h6")

	// Class attributes for styling
	p.AllowAttrs("class").OnElements("p", "div", "span", "h1", "h2", "h3", "h4", "h5", "h6", "table", "tr", "td", "th")

	return p
}

// sanitizeByMode sanitizes input based on the specified mode
func (s *XSSSanitizer) sanitizeByMode(input string, mode string) string {
	input = strings.TrimSpace(input)

	switch mode {
	case "strict":
		return s.strictPolicy.Sanitize(input)
	case "ugc":
		return s.ugcPolicy.Sanitize(input)
	case "rich":
		return s.richPolicy.Sanitize(input)
	case "none":
		return input
	default:
		// Default to UGC mode
		return s.ugcPolicy.Sanitize(input)
	}
}

// SanitizeStruct sanitizes all string fields in a struct based on sanitize tags
func (s *XSSSanitizer) SanitizeStruct(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct, got %T", v)
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Only process string fields
		if field.Kind() != reflect.String {
			continue
		}

		originalValue := field.String()
		if originalValue == "" {
			continue
		}

		// Get sanitize mode from struct tag
		sanitizeTag := fieldType.Tag.Get("sanitize")

		// Default to "strict" if no tag specified (Security First)
		if sanitizeTag == "" {
			sanitizeTag = "strict"
		}

		// Sanitize the field
		sanitizedValue := s.sanitizeByMode(originalValue, sanitizeTag)
		field.SetString(sanitizedValue)
	}

	return nil
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) map[string]string {
	// First sanitize based on struct tags
	if err := sanitizer.SanitizeStruct(s); err != nil {
		return map[string]string{
			"sanitization_error": err.Error(),
		}
	}

	// Then validate
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)

	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		tag := err.Tag()

		switch tag {
		case "required":
			errors[field] = fmt.Sprintf("%s is required", field)
		case "email":
			errors[field] = fmt.Sprintf("%s must be a valid email", field)
		case "min":
			errors[field] = fmt.Sprintf("%s must be at least %s characters", field, err.Param())
		case "max":
			errors[field] = fmt.Sprintf("%s must be at most %s characters", field, err.Param())
		case "gte":
			errors[field] = fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
		case "lte":
			errors[field] = fmt.Sprintf("%s must be less than or equal to %s", field, err.Param())
		case "datetime":
			errors[field] = fmt.Sprintf("%s must be in format %s", field, err.Param())
		default:
			errors[field] = fmt.Sprintf("%s is invalid", field)
		}
	}

	return errors
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	return validate
}

// GetSanitizer returns the sanitizer instance
func GetSanitizer() *XSSSanitizer {
	return sanitizer
}
