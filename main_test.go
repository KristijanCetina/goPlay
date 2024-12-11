package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mockable ValidateAPIKey function
var mockValidateAPIKey = validateAPIKey

// Mockable RateLimiter struct
type MockRateLimiter struct {
	allowFunc func(string) bool
}

func (m *MockRateLimiter) Allow(apiKey string) bool {
	return m.allowFunc(apiKey)
}

func TestAPIRoute(t *testing.T) {
	// Mock validateAPIKey function
	mockValidateAPIKey = func(key string) bool {
		return key == "valid-key"
	}

	// Mock rate limiter
	mockRateLimiter := &MockRateLimiter{
		allowFunc: func(apiKey string) bool {
			return apiKey != "rate-limited-key"
		},
	}

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
			expectedBody:   "Access Granted!",
		},
		{
			name:           "Invalid API Key",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: Invalid API key\n",
		},
		{
			name:           "No Key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: Invalid API key\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/apiroute?api_key="+tt.apiKey, nil)
			rr := httptest.NewRecorder()

			// Inject mocks into handler
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				key := r.URL.Query().Get("api_key")
				if !mockValidateAPIKey(key) {
					http.Error(w, "Unauthorized: Invalid API key", http.StatusUnauthorized)
					return
				}

				if !mockRateLimiter.Allow(key) {
					http.Error(w, "Too Many Requests: Rate limit exceeded", http.StatusTooManyRequests)
					return
				}

				w.Header().Set("Content-Type", "text/plain")
				w.Write([]byte("Access Granted!"))
			}).ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if body := rr.Body.String(); body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}
