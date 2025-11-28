package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/mail"
	"reflect"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// contextKey type for context keys to avoid collisions.
type contextKey string

const requestBodyKey contextKey = "requestBody"

// RegisterMiddleware attaches common middleware to the router.
func RegisterMiddleware(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
}

// RequireAPIKey is a middleware that validates the X-API-Key header.
func RequireAPIKey(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			providedKey := r.Header.Get("X-API-Key")
			if providedKey == "" || providedKey != apiKey {
				writeJSON(w, http.StatusUnauthorized, apiResponse{Success: false, Error: "invalid or missing API key"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateRequest is a generic middleware that validates request body fields.
// It decodes the JSON body into the provided type T and validates all marked fields.
func ValidateRequest[T any]() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new instance of type T
			var reqValue T

			// Decode the request body
			if err := json.NewDecoder(r.Body).Decode(&reqValue); err != nil {
				writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Error: "invalid request body"})
				return
			}

			// Validate required fields
			if err := validateRequiredFields(&reqValue); err != nil {
				writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Error: err.Error()})
				return
			}

			// Validate email fields
			if err := validateEmails(&reqValue); err != nil {
				writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Error: err.Error()})
				return
			}

			// Validate UUID fields
			if err := validateUUIDs(&reqValue); err != nil {
				writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Error: err.Error()})
				return
			}

			// Validate password fields
			if err := validatePasswords(&reqValue); err != nil {
				writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Error: err.Error()})
				return
			}

			// Store the decoded request in context for the handler to use
			ctx := context.WithValue(r.Context(), requestBodyKey, &reqValue)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateRequiredFields checks if fields marked with validate:"required" are non-empty.
func validateRequiredFields(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	var missingFields []string

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Check if field has validate:"required" tag
		validateTag := field.Tag.Get("validate")
		if !strings.Contains(validateTag, "required") {
			continue
		}

		// Check if string field is empty
		if fieldValue.Kind() == reflect.String && fieldValue.String() == "" {
			jsonTag := field.Tag.Get("json")
			fieldName := strings.Split(jsonTag, ",")[0]
			if fieldName == "" {
				fieldName = field.Name
			}
			missingFields = append(missingFields, fieldName)
		}
	}

	if len(missingFields) > 0 {
		return &ValidationError{Fields: missingFields}
	}

	return nil
}

// validateEmails checks if fields marked with validate:"email" have valid email format.
func validateEmails(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	var invalidFields []string

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Check if field has validate tag with "email"
		validateTag := field.Tag.Get("validate")
		if !strings.Contains(validateTag, "email") {
			continue
		}

		// Only validate string fields
		if fieldValue.Kind() == reflect.String {
			emailStr := fieldValue.String()
			// Skip empty strings - let required validation handle that
			if emailStr == "" {
				continue
			}

			// Validate email format using net/mail
			if _, err := mail.ParseAddress(emailStr); err != nil {
				jsonTag := field.Tag.Get("json")
				fieldName := strings.Split(jsonTag, ",")[0]
				if fieldName == "" {
					fieldName = field.Name
				}
				invalidFields = append(invalidFields, fieldName)
			}
		}
	}

	if len(invalidFields) > 0 {
		return &EmailValidationError{Fields: invalidFields}
	}

	return nil
}

// validateUUIDs checks if fields marked with validate:"uuid" have valid UUID format.
func validateUUIDs(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	var invalidFields []string

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Check if field has validate tag with "uuid"
		validateTag := field.Tag.Get("validate")
		if !strings.Contains(validateTag, "uuid") {
			continue
		}

		// Only validate string fields
		if fieldValue.Kind() == reflect.String {
			uuidStr := fieldValue.String()
			// Skip empty strings - let required validation handle that
			if uuidStr == "" {
				continue
			}

			// Validate UUID format using google/uuid
			if err := uuid.Validate(uuidStr); err != nil {
				jsonTag := field.Tag.Get("json")
				fieldName := strings.Split(jsonTag, ",")[0]
				if fieldName == "" {
					fieldName = field.Name
				}
				invalidFields = append(invalidFields, fieldName)
			}
		}
	}

	if len(invalidFields) > 0 {
		return &UUIDValidationError{Fields: invalidFields}
	}

	return nil
}

// validatePasswords checks if fields marked with validate:"password" meet minimum strength requirements.
func validatePasswords(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	var invalidFields []string

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Check if field has validate tag with "password"
		validateTag := field.Tag.Get("validate")
		if !strings.Contains(validateTag, "password") {
			continue
		}

		// Only validate string fields
		if fieldValue.Kind() == reflect.String {
			passwordStr := fieldValue.String()
			// Skip empty strings - let required validation handle that
			if passwordStr == "" {
				continue
			}

			// Validate password strength (minimum 8 characters)
			if len(passwordStr) < 8 {
				jsonTag := field.Tag.Get("json")
				fieldName := strings.Split(jsonTag, ",")[0]
				if fieldName == "" {
					fieldName = field.Name
				}
				invalidFields = append(invalidFields, fieldName)
			}
		}
	}

	if len(invalidFields) > 0 {
		return &PasswordValidationError{Fields: invalidFields}
	}

	return nil
}

// ValidationError represents validation errors.
type ValidationError struct {
	Fields []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " required"
}

// EmailValidationError represents email validation errors.
type EmailValidationError struct {
	Fields []string
}

func (e *EmailValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " must be valid email address(es)"
}

// UUIDValidationError represents UUID validation errors.
type UUIDValidationError struct {
	Fields []string
}

func (e *UUIDValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " must be valid UUID(s)"
}

// PasswordValidationError represents password validation errors.
type PasswordValidationError struct {
	Fields []string
}

func (e *PasswordValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " must be at least 8 characters"
}

// GetRequestBody retrieves the validated request body from the context.
func GetRequestBody[T any](r *http.Request) *T {
	if body := r.Context().Value(requestBodyKey); body != nil {
		if typedBody, ok := body.(*T); ok {
			return typedBody
		}
	}
	return nil
}
