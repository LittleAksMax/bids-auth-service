package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test structures for validation
type testRequiredStruct struct {
	Field1 string `json:"field1" validate:"required"`
	Field2 string `json:"field2" validate:"required"`
	Field3 string `json:"field3"` // Not required
}

type testEmailStruct struct {
	Email1 string `json:"email1" validate:"required,email"`
	Email2 string `json:"email2" validate:"email"` // Optional email
	Field3 string `json:"field3"`                  // Not an email
}

type testUUIDStruct struct {
	UserID   string `json:"user_id" validate:"required,uuid"`
	ParentID string `json:"parent_id" validate:"uuid"` // Optional UUID
	Name     string `json:"name"`                      // Not a UUID
}

type testCombinedStruct struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	UserID   string `json:"user_id" validate:"required,uuid"`
	Optional string `json:"optional"`
}

// TestValidateRequiredFields tests the required field validation
func TestValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		expectedError bool
		errorContains string
	}{
		{
			name: "all required fields present",
			input: &testRequiredStruct{
				Field1: "value1",
				Field2: "value2",
				Field3: "value3",
			},
			expectedError: false,
		},
		{
			name: "one required field missing",
			input: &testRequiredStruct{
				Field1: "value1",
				Field2: "",
				Field3: "value3",
			},
			expectedError: true,
			errorContains: "field2",
		},
		{
			name: "multiple required fields missing",
			input: &testRequiredStruct{
				Field1: "",
				Field2: "",
				Field3: "value3",
			},
			expectedError: true,
			errorContains: "required",
		},
		{
			name: "optional field can be empty",
			input: &testRequiredStruct{
				Field1: "value1",
				Field2: "value2",
				Field3: "",
			},
			expectedError: false,
		},
		{
			name:          "all fields empty",
			input:         &testRequiredStruct{},
			expectedError: true,
			errorContains: "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredFields(tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidateEmails tests the email validation
func TestValidateEmails(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		expectedError bool
		errorContains string
	}{
		{
			name: "valid emails",
			input: &testEmailStruct{
				Email1: "test@example.com",
				Email2: "user@domain.co.uk",
				Field3: "not-an-email",
			},
			expectedError: false,
		},
		{
			name: "invalid email format in required field",
			input: &testEmailStruct{
				Email1: "invalid-email",
				Email2: "valid@example.com",
				Field3: "anything",
			},
			expectedError: true,
			errorContains: "email1",
		},
		{
			name: "invalid email format in optional field",
			input: &testEmailStruct{
				Email1: "valid@example.com",
				Email2: "not-an-email",
				Field3: "anything",
			},
			expectedError: true,
			errorContains: "email2",
		},
		{
			name: "multiple invalid emails",
			input: &testEmailStruct{
				Email1: "invalid1",
				Email2: "invalid2",
				Field3: "anything",
			},
			expectedError: true,
			errorContains: "email address",
		},
		{
			name: "empty email in optional field - should pass",
			input: &testEmailStruct{
				Email1: "valid@example.com",
				Email2: "",
				Field3: "anything",
			},
			expectedError: false,
		},
		{
			name: "email with display name",
			input: &testEmailStruct{
				Email1: "John Doe <john@example.com>",
				Email2: "jane@example.com",
				Field3: "anything",
			},
			expectedError: false,
		},
		{
			name: "email with special characters",
			input: &testEmailStruct{
				Email1: "user+tag@example.com",
				Email2: "test.user@sub.domain.com",
				Field3: "anything",
			},
			expectedError: false,
		},
		{
			name: "email missing @ symbol",
			input: &testEmailStruct{
				Email1: "invalidemail.com",
				Email2: "",
				Field3: "anything",
			},
			expectedError: true,
			errorContains: "email1",
		},
		{
			name: "email missing domain",
			input: &testEmailStruct{
				Email1: "user@",
				Email2: "",
				Field3: "anything",
			},
			expectedError: true,
			errorContains: "email1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmails(tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidateUUIDs tests the UUID validation
func TestValidateUUIDs(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		expectedError bool
		errorContains string
	}{
		{
			name: "valid UUIDs",
			input: &testUUIDStruct{
				UserID:   "550e8400-e29b-41d4-a716-446655440000",
				ParentID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				Name:     "test-name",
			},
			expectedError: false,
		},
		{
			name: "invalid UUID format in required field",
			input: &testUUIDStruct{
				UserID:   "not-a-uuid",
				ParentID: "550e8400-e29b-41d4-a716-446655440000",
				Name:     "test",
			},
			expectedError: true,
			errorContains: "user_id",
		},
		{
			name: "invalid UUID format in optional field",
			input: &testUUIDStruct{
				UserID:   "550e8400-e29b-41d4-a716-446655440000",
				ParentID: "invalid-uuid",
				Name:     "test",
			},
			expectedError: true,
			errorContains: "parent_id",
		},
		{
			name: "multiple invalid UUIDs",
			input: &testUUIDStruct{
				UserID:   "invalid1",
				ParentID: "invalid2",
				Name:     "test",
			},
			expectedError: true,
			errorContains: "UUID",
		},
		{
			name: "empty UUID in optional field - should pass",
			input: &testUUIDStruct{
				UserID:   "550e8400-e29b-41d4-a716-446655440000",
				ParentID: "",
				Name:     "test",
			},
			expectedError: false,
		},
		{
			name: "UUID without hyphens - valid per google/uuid",
			input: &testUUIDStruct{
				UserID:   "550e8400e29b41d4a716446655440000",
				ParentID: "",
				Name:     "test",
			},
			expectedError: false,
		},
		{
			name: "UUID v4 format",
			input: &testUUIDStruct{
				UserID:   "f47ac10b-58cc-4372-a567-0e02b2c3d479",
				ParentID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
				Name:     "test",
			},
			expectedError: false,
		},
		{
			name: "UUID with wrong length",
			input: &testUUIDStruct{
				UserID:   "550e8400-e29b-41d4-a716-4466554400",
				ParentID: "",
				Name:     "test",
			},
			expectedError: true,
			errorContains: "user_id",
		},
		{
			name: "UUID with invalid characters",
			input: &testUUIDStruct{
				UserID:   "550e8400-e29b-41d4-a716-44665544000g",
				ParentID: "",
				Name:     "test",
			},
			expectedError: true,
			errorContains: "user_id",
		},
		{
			name: "all zeros UUID - should be valid",
			input: &testUUIDStruct{
				UserID:   "00000000-0000-0000-0000-000000000000",
				ParentID: "",
				Name:     "test",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUUIDs(tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidateRequestMiddleware tests the complete validation middleware
func TestValidateRequestMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		errorContains  string
	}{
		{
			name: "valid combined request",
			requestBody: testCombinedStruct{
				Username: "john_doe",
				Email:    "john@example.com",
				UserID:   "550e8400-e29b-41d4-a716-446655440000",
				Optional: "anything",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing required field",
			requestBody: testCombinedStruct{
				Username: "",
				Email:    "john@example.com",
				UserID:   "550e8400-e29b-41d4-a716-446655440000",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "required",
		},
		{
			name: "invalid email",
			requestBody: testCombinedStruct{
				Username: "john_doe",
				Email:    "not-an-email",
				UserID:   "550e8400-e29b-41d4-a716-446655440000",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "email",
		},
		{
			name: "invalid UUID",
			requestBody: testCombinedStruct{
				Username: "john_doe",
				Email:    "john@example.com",
				UserID:   "not-a-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "UUID",
		},
		{
			name:           "invalid JSON",
			requestBody:    "not-valid-json",
			expectedStatus: http.StatusBadRequest,
			errorContains:  "invalid request body",
		},
		{
			name: "multiple validation errors - required fails first",
			requestBody: testCombinedStruct{
				Username: "",
				Email:    "invalid-email",
				UserID:   "invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Create a test handler that will be called if validation passes
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Retrieve the validated body
				validated := GetRequestBody[testCombinedStruct](r)
				if validated == nil {
					t.Error("expected validated body but got nil")
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(apiResponse{Success: true, Data: "ok"})
			})

			// Apply the validation middleware
			handler := ValidateRequest[testCombinedStruct]()(testHandler)
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d but got %d", tt.expectedStatus, rr.Code)
			}

			// Check error message if expected
			if tt.errorContains != "" {
				var response apiResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if !contains(response.Error, tt.errorContains) {
					t.Errorf("response error %q does not contain %q", response.Error, tt.errorContains)
				}
			}
		})
	}
}

// TestGetRequestBody tests the context retrieval function
func TestGetRequestBody(t *testing.T) {
	t.Run("body exists in context", func(t *testing.T) {
		expected := &testRequiredStruct{
			Field1: "value1",
			Field2: "value2",
		}

		ctx := context.WithValue(context.Background(), requestBodyKey, expected)
		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)

		result := GetRequestBody[testRequiredStruct](req)
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.Field1 != expected.Field1 || result.Field2 != expected.Field2 {
			t.Errorf("expected %+v but got %+v", expected, result)
		}
	})

	t.Run("body does not exist in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		result := GetRequestBody[testRequiredStruct](req)
		if result != nil {
			t.Errorf("expected nil but got %+v", result)
		}
	})

	t.Run("wrong type in context", func(t *testing.T) {
		wrong := &testEmailStruct{Email1: "test@example.com"}
		ctx := context.WithValue(context.Background(), requestBodyKey, wrong)
		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)

		result := GetRequestBody[testRequiredStruct](req)
		if result != nil {
			t.Errorf("expected nil for wrong type but got %+v", result)
		}
	})
}

// TestValidationErrorTypes tests the error type implementations
func TestValidationErrorTypes(t *testing.T) {
	t.Run("ValidationError", func(t *testing.T) {
		err := &ValidationError{Fields: []string{"field1", "field2"}}
		expected := "field1, field2 required"
		if err.Error() != expected {
			t.Errorf("expected %q but got %q", expected, err.Error())
		}
	})

	t.Run("EmailValidationError", func(t *testing.T) {
		err := &EmailValidationError{Fields: []string{"email1", "email2"}}
		expected := "email1, email2 must be valid email address(es)"
		if err.Error() != expected {
			t.Errorf("expected %q but got %q", expected, err.Error())
		}
	})

	t.Run("UUIDValidationError", func(t *testing.T) {
		err := &UUIDValidationError{Fields: []string{"user_id", "parent_id"}}
		expected := "user_id, parent_id must be valid UUID(s)"
		if err.Error() != expected {
			t.Errorf("expected %q but got %q", expected, err.Error())
		}
	})

	t.Run("single field errors", func(t *testing.T) {
		err1 := &ValidationError{Fields: []string{"field1"}}
		if err1.Error() != "field1 required" {
			t.Errorf("unexpected error: %s", err1.Error())
		}

		err2 := &EmailValidationError{Fields: []string{"email"}}
		if err2.Error() != "email must be valid email address(es)" {
			t.Errorf("unexpected error: %s", err2.Error())
		}

		err3 := &UUIDValidationError{Fields: []string{"id"}}
		if err3.Error() != "id must be valid UUID(s)" {
			t.Errorf("unexpected error: %s", err3.Error())
		}
	})
}

// TestValidatePasswords tests password strength validation
func TestValidatePasswords(t *testing.T) {
	type testPasswordStruct struct {
		Password string `json:"password" validate:"required,password"`
		Optional string `json:"optional" validate:"password"` // Optional password field
		Name     string `json:"name"`                         // Not a password
	}

	tests := []struct {
		name          string
		input         interface{}
		expectedError bool
		errorContains string
	}{
		{
			name: "valid password - meets minimum length",
			input: &testPasswordStruct{
				Password: "password123",
			},
			expectedError: false,
		},
		{
			name: "valid password - exactly 8 characters",
			input: &testPasswordStruct{
				Password: "12345678",
			},
			expectedError: false,
		},
		{
			name: "invalid password - too short",
			input: &testPasswordStruct{
				Password: "pass123",
			},
			expectedError: true,
			errorContains: "password",
		},
		{
			name: "invalid password - empty string",
			input: &testPasswordStruct{
				Password: "",
			},
			expectedError: false, // Empty handled by required validation
		},
		{
			name: "optional password field empty - should pass",
			input: &testPasswordStruct{
				Password: "validpass123",
				Optional: "",
			},
			expectedError: false,
		},
		{
			name: "optional password field invalid - should fail",
			input: &testPasswordStruct{
				Password: "validpass123",
				Optional: "short",
			},
			expectedError: true,
			errorContains: "optional",
		},
		{
			name: "very long password - should pass",
			input: &testPasswordStruct{
				Password: "ThisIsAVeryLongPasswordThatShouldDefinitelyPass123!@#",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePasswords(tt.input)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("nil pointer - should panic", func(t *testing.T) {
		var nilStruct *testRequiredStruct
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for nil pointer but didn't panic")
			}
		}()
		// This should panic - nil pointer dereference in reflection
		_ = validateRequiredFields(nilStruct)
	})

	t.Run("empty struct with no validation tags", func(t *testing.T) {
		type noValidation struct {
			Field1 string `json:"field1"`
			Field2 string `json:"field2"`
		}
		s := &noValidation{}

		err1 := validateRequiredFields(s)
		if err1 != nil {
			t.Errorf("expected no error for no required tags, got: %v", err1)
		}

		err2 := validateEmails(s)
		if err2 != nil {
			t.Errorf("expected no error for no email tags, got: %v", err2)
		}

		err3 := validateUUIDs(s)
		if err3 != nil {
			t.Errorf("expected no error for no uuid tags, got: %v", err3)
		}
	})

	t.Run("struct with only whitespace in required field", func(t *testing.T) {
		s := &testRequiredStruct{
			Field1: "   ",
			Field2: "valid",
		}
		// Current implementation only checks for empty string, not whitespace
		err := validateRequiredFields(s)
		if err != nil {
			t.Logf("whitespace handling: %v", err)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
