package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/mail"
	"reflect"
	"strings"

	"github.com/LittleAksMax/bids-util/requests"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

// ApplyCORS configures CORS with an allow-all default or a restricted list of origins.
// Pass the ALLOWED_ORIGINS values from config. Use "*" to allow all.
func ApplyCORS(r chi.Router, allowedOrigins []string) {
	allowAll := false
	for _, o := range allowedOrigins {
		if strings.TrimSpace(o) == "*" {
			allowAll = true
			break
		}
	}

	c := cors.Handler(cors.Options{
		AllowedOrigins: func() []string {
			if allowAll {
				return []string{"*"}
			}
			return allowedOrigins
		}(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Auth-Claims", "X-Auth-Ts", "X-Auth-Sig"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	r.Use(c)
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
				requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{Success: false, Error: "invalid request body"})
				return
			}

			// Validate required fields
			if err := validateRequiredFields(&reqValue); err != nil {
				requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{Success: false, Error: err.Error()})
				return
			}

			// Validate email fields
			if err := validateEmails(&reqValue); err != nil {
				requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{Success: false, Error: err.Error()})
				return
			}

			// Validate UUID fields
			if err := validateUUIDs(&reqValue); err != nil {
				requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{Success: false, Error: err.Error()})
				return
			}

			// Validate password fields
			if err := validatePasswords(&reqValue); err != nil {
				requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{Success: false, Error: err.Error()})
				return
			}

			// Validate role fields
			if err := validateRoles(&reqValue); err != nil {
				requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{Success: false, Error: err.Error()})
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

// validateRoles checks if fields marked with validate:"role" contain allowed role values.
func validateRoles(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	var invalidFields []string
	allowed := map[string]struct{}{
		"user":  {},
		"admin": {},
	}

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Check if field has validate tag with "role"
		validateTag := field.Tag.Get("validate")
		if !strings.Contains(validateTag, "role") {
			continue
		}

		// Only validate string fields
		if fieldValue.Kind() == reflect.String {
			roleStr := strings.ToLower(strings.TrimSpace(fieldValue.String()))
			// Skip empty strings - let required validation handle that
			if roleStr == "" {
				continue
			}

			if _, ok := allowed[roleStr]; !ok {
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
		return &RoleValidationError{Fields: invalidFields}
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

// RoleValidationError represents role validation errors.
type RoleValidationError struct {
	Fields []string
}

func (e *RoleValidationError) Error() string {
	return strings.Join(e.Fields, ", ") + " must be one of: user, admin"
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
