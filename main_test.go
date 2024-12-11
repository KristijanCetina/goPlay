package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock the database functions for testing
func mockValidateAPIKey(key string) bool {
	// Simulate valid and invalid API keys
	if key == "valid-key" {
		return true
	}
	return false
}

// Test the handler function
func TestHandler(t *testing.T) {
	// Replace the validateAPIKey function with the mock
	// validateAPIKey = mockValidateAPIKey()

	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid API Key",
			apiKey:         "valid-key",
			expectedStatus: http.StatusOK,
			expectedBody:   "{\"message\":\"Hello, base route!\"}\n",
		},
		{
			name:           "Invalid API Key",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: Invalid API key\n",
		},
		{
			name:           "Missing API Key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: Invalid API key\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request
			req := httptest.NewRequest(http.MethodGet, "/apiroute", nil)
			if tt.apiKey != "" {
				q := req.URL.Query()
				q.Add("api_key", tt.apiKey)
				req.URL.RawQuery = q.Encode()
			}

			// Create a ResponseRecorder to capture the response
			rr := httptest.NewRecorder()

			// Call the handler function
			handler(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check the response body
			if body := rr.Body.String(); body != tt.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v", body, tt.expectedBody)
			}
		})
	}
}
