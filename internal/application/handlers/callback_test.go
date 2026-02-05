package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/encryption"
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/infrastructure/oauth"
	"github.com/gin-gonic/gin"
)

func getTestKey() string {
	// return 32 chars for testing purposes
	return "PtBryEFj9MPJRT6VLZzpmGrpyGrMsAVF"
}

func TestCallbackHandlerWrapped(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		code               string
		state              string
		clientID           string
		clientSecret       string
		tokenResponse      oauth.TokenResponse
		tokenError         error
		expectedStatusCode int
		expectedRedirect   string
		expectCookie       bool
		expectJSON         bool
	}{
		{
			name:         "successful callback without state",
			code:         "auth-code-123",
			state:        "",
			clientID:     "client-id",
			clientSecret: "client-secret",
			tokenResponse: oauth.TokenResponse{
				AccessToken: "ghp_accesstoken123",
			},
			tokenError:         nil,
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/repos",
			expectCookie:       true,
			expectJSON:         false,
		},
		{
			name:         "successful callback with state",
			code:         "auth-code-123",
			state:        "owner/repo1,owner/repo2",
			clientID:     "client-id",
			clientSecret: "client-secret",
			tokenResponse: oauth.TokenResponse{
				AccessToken: "ghp_accesstoken123",
			},
			tokenError:         nil,
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/repos?repos=owner/repo1,owner/repo2",
			expectCookie:       true,
			expectJSON:         false,
		},
		{
			name:               "missing authorization code",
			code:               "",
			state:              "",
			clientID:           "client-id",
			clientSecret:       "client-secret",
			tokenResponse:      oauth.TokenResponse{},
			tokenError:         nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedRedirect:   "",
			expectCookie:       false,
			expectJSON:         true,
		},
		{
			name:               "missing client ID",
			code:               "auth-code-123",
			state:              "",
			clientID:           "",
			clientSecret:       "client-secret",
			tokenResponse:      oauth.TokenResponse{},
			tokenError:         nil,
			expectedStatusCode: http.StatusInternalServerError,
			expectedRedirect:   "",
			expectCookie:       false,
			expectJSON:         true,
		},
		{
			name:               "missing client secret",
			code:               "auth-code-123",
			state:              "",
			clientID:           "client-id",
			clientSecret:       "",
			tokenResponse:      oauth.TokenResponse{},
			tokenError:         nil,
			expectedStatusCode: http.StatusInternalServerError,
			expectedRedirect:   "",
			expectCookie:       false,
			expectJSON:         true,
		},
		{
			name:               "token exchange error",
			code:               "auth-code-123",
			state:              "",
			clientID:           "client-id",
			clientSecret:       "client-secret",
			tokenResponse:      oauth.TokenResponse{},
			tokenError:         errors.New("network error"),
			expectedStatusCode: http.StatusBadRequest,
			expectedRedirect:   "",
			expectCookie:       false,
			expectJSON:         true,
		},
		{
			name:         "token response contains error",
			code:         "auth-code-123",
			state:        "",
			clientID:     "client-id",
			clientSecret: "client-secret",
			tokenResponse: oauth.TokenResponse{
				Error:     "invalid_grant",
				ErrorDesc: "The provided authorization grant is invalid",
			},
			tokenError:         nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedRedirect:   "",
			expectCookie:       false,
			expectJSON:         true,
		},
		{
			name:         "empty access token in response",
			code:         "auth-code-123",
			state:        "",
			clientID:     "client-id",
			clientSecret: "client-secret",
			tokenResponse: oauth.TokenResponse{
				AccessToken: "",
			},
			tokenError:         nil,
			expectedStatusCode: http.StatusInternalServerError,
			expectedRedirect:   "",
			expectCookie:       false,
			expectJSON:         true,
		},
		{
			name:         "state with special characters",
			code:         "auth-code-123",
			state:        "org-name/repo_name",
			clientID:     "client-id",
			clientSecret: "client-secret",
			tokenResponse: oauth.TokenResponse{
				AccessToken: "ghp_token",
			},
			tokenError:         nil,
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/repos?repos=org-name/repo_name",
			expectCookie:       true,
			expectJSON:         false,
		},
		{
			name:         "complex state parameter",
			code:         "auth-code-123",
			state:        "OctopusDeploy/OctopusDeploy,actions/checkout,docker/build-push-action",
			clientID:     "client-id",
			clientSecret: "client-secret",
			tokenResponse: oauth.TokenResponse{
				AccessToken: encryption.EncryptStringNoErr("ghp_longtoken123456789", getTestKey),
			},
			tokenError:         nil,
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/repos?repos=OctopusDeploy/OctopusDeploy,actions/checkout,docker/build-push-action",
			expectCookie:       true,
			expectJSON:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock functions
			mockOAuthExchange := func(clientID, clientSecret, code, url string) (oauth.TokenResponse, error) {
				// Verify parameters
				if clientID != tt.clientID {
					t.Errorf("OAuth exchange called with clientID %q, expected %q", clientID, tt.clientID)
				}
				if clientSecret != tt.clientSecret {
					t.Errorf("OAuth exchange called with clientSecret %q, expected %q", clientSecret, tt.clientSecret)
				}
				if code != tt.code {
					t.Errorf("OAuth exchange called with code %q, expected %q", code, tt.code)
				}
				return tt.tokenResponse, tt.tokenError
			}

			mockClientID := func() string {
				return tt.clientID
			}

			mockClientSecret := func() string {
				return tt.clientSecret
			}

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Build request URL with query parameters
			url := "/callback"
			if tt.code != "" || tt.state != "" {
				url += "?"
				if tt.code != "" {
					url += "code=" + tt.code
				}
				if tt.state != "" {
					if tt.code != "" {
						url += "&"
					}
					url += "state=" + tt.state
				}
			}

			req := httptest.NewRequest("GET", url, nil)
			c.Request = req

			// Call the handler
			CallbackHandlerWrapped(c, mockOAuthExchange, mockClientID, mockClientSecret, getTestKey)

			// Check status code
			if w.Code != tt.expectedStatusCode {
				t.Errorf("Status code = %d, expected %d", w.Code, tt.expectedStatusCode)
			}

			// Check redirect
			if tt.expectedRedirect != "" {
				location := w.Header().Get("Location")
				if location != tt.expectedRedirect {
					t.Errorf("Redirect location = %q, expected %q", location, tt.expectedRedirect)
				}
			}

			// Check cookie
			if tt.expectCookie {
				cookies := w.Result().Cookies()
				found := false
				for _, cookie := range cookies {
					if cookie.Name == "github_token" {
						found = true
						if value, err := encryption.DecryptStringWrapper(cookie.Value, getTestKey); err != nil || value != tt.tokenResponse.AccessToken {
							t.Errorf("Cookie value = %q, expected %q. err %v", value, tt.tokenResponse.AccessToken, err)
						}
						if cookie.MaxAge != 3600 {
							t.Errorf("Cookie MaxAge = %d, expected 3600", cookie.MaxAge)
						}
						if cookie.Path != "/" {
							t.Errorf("Cookie Path = %q, expected %q", cookie.Path, "/")
						}
						if !cookie.HttpOnly {
							t.Error("Cookie should be HttpOnly")
						}
						break
					}
				}
				if !found {
					t.Error("Expected github_token cookie to be set, but it was not found")
				}
			} else {
				cookies := w.Result().Cookies()
				for _, cookie := range cookies {
					if cookie.Name == "github_token" {
						t.Error("Did not expect github_token cookie to be set, but it was found")
					}
				}
			}

			// Check JSON response
			if tt.expectJSON {
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json; charset=utf-8" {
					t.Errorf("Content-Type = %q, expected JSON", contentType)
				}
			}
		})
	}
}

func TestCallbackHandlerWrappedOAuthExchangeNotCalled(t *testing.T) {
	// Test that OAuth exchange is not called when code is missing
	gin.SetMode(gin.TestMode)

	exchangeCalled := false
	mockOAuthExchange := func(clientID, clientSecret, code, url string) (oauth.TokenResponse, error) {
		exchangeCalled = true
		return oauth.TokenResponse{}, nil
	}

	mockClientID := func() string {
		return "client-id"
	}

	mockClientSecret := func() string {
		return "client-secret"
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/callback", nil) // No code parameter
	c.Request = req

	CallbackHandlerWrapped(c, mockOAuthExchange, mockClientID, mockClientSecret, getTestKey)

	if exchangeCalled {
		t.Error("OAuth exchange should not be called when code is missing")
	}
}

func TestCallbackHandlerWrappedClientSettingsCalled(t *testing.T) {
	// Test that client ID and secret settings are called
	gin.SetMode(gin.TestMode)

	clientIDCalled := false
	clientSecretCalled := false

	mockOAuthExchange := func(clientID, clientSecret, code, url string) (oauth.TokenResponse, error) {
		return oauth.TokenResponse{AccessToken: "token"}, nil
	}

	mockClientID := func() string {
		clientIDCalled = true
		return "client-id"
	}

	mockClientSecret := func() string {
		clientSecretCalled = true
		return "client-secret"
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/callback?code=test-code", nil)
	c.Request = req

	CallbackHandlerWrapped(c, mockOAuthExchange, mockClientID, mockClientSecret, getTestKey)

	if !clientIDCalled {
		t.Error("Client ID setting function was not called")
	}

	if !clientSecretCalled {
		t.Error("Client secret setting function was not called")
	}
}

func TestCallbackHandlerWrappedMultipleCalls(t *testing.T) {
	// Test that calling the handler multiple times produces consistent results
	gin.SetMode(gin.TestMode)

	mockOAuthExchange := func(clientID, clientSecret, code, url string) (oauth.TokenResponse, error) {
		return oauth.TokenResponse{AccessToken: "token123"}, nil
	}

	mockClientID := func() string {
		return "client-id"
	}

	mockClientSecret := func() string {
		return "client-secret"
	}

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest("GET", "/callback?code=test-code&state=owner/repo", nil)
		c.Request = req

		CallbackHandlerWrapped(c, mockOAuthExchange, mockClientID, mockClientSecret, getTestKey)

		if w.Code != http.StatusFound {
			t.Errorf("Call %d: status code = %d, expected %d", i+1, w.Code, http.StatusFound)
		}

		location := w.Header().Get("Location")
		expectedLocation := "/repos?repos=owner/repo"
		if location != expectedLocation {
			t.Errorf("Call %d: redirect location = %q, expected %q", i+1, location, expectedLocation)
		}
	}
}

func TestCallbackHandlerWrappedErrorMessages(t *testing.T) {
	// Test that error messages are properly returned
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		code           string
		clientID       string
		clientSecret   string
		tokenResponse  oauth.TokenResponse
		tokenError     error
		expectErrorKey bool
		expectDescKey  bool
	}{
		{
			name:           "missing code error message",
			code:           "",
			clientID:       "client-id",
			clientSecret:   "client-secret",
			expectErrorKey: true,
			expectDescKey:  false,
		},
		{
			name:           "missing client ID error message",
			code:           "code",
			clientID:       "",
			clientSecret:   "secret",
			expectErrorKey: true,
			expectDescKey:  false,
		},
		{
			name:         "token response error with description",
			code:         "code",
			clientID:     "client-id",
			clientSecret: "client-secret",
			tokenResponse: oauth.TokenResponse{
				Error:     "invalid_grant",
				ErrorDesc: "The code is expired",
			},
			expectErrorKey: true,
			expectDescKey:  true,
		},
		{
			name:         "empty access token error message",
			code:         "code",
			clientID:     "client-id",
			clientSecret: "client-secret",
			tokenResponse: oauth.TokenResponse{
				AccessToken: "",
			},
			expectErrorKey: true,
			expectDescKey:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOAuthExchange := func(clientID, clientSecret, code, url string) (oauth.TokenResponse, error) {
				return tt.tokenResponse, tt.tokenError
			}

			mockClientID := func() string {
				return tt.clientID
			}

			mockClientSecret := func() string {
				return tt.clientSecret
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			url := "/callback"
			if tt.code != "" {
				url += "?code=" + tt.code
			}
			req := httptest.NewRequest("GET", url, nil)
			c.Request = req

			CallbackHandlerWrapped(c, mockOAuthExchange, mockClientID, mockClientSecret, getTestKey)

			// Check that response is JSON with error
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, expected JSON", contentType)
			}

			// Check that response body contains expected keys
			body := w.Body.String()
			if tt.expectErrorKey && body == "" {
				t.Error("Expected error message in response body")
			}
		})
	}
}

func TestCallbackHandlerWrappedCookieProperties(t *testing.T) {
	// Test that cookie is set with correct properties
	gin.SetMode(gin.TestMode)

	mockOAuthExchange := func(clientID, clientSecret, code, url string) (oauth.TokenResponse, error) {
		return oauth.TokenResponse{AccessToken: "test-token-123"}, nil
	}

	mockClientID := func() string {
		return "client-id"
	}

	mockClientSecret := func() string {
		return "client-secret"
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/callback?code=test-code", nil)
	c.Request = req

	CallbackHandlerWrapped(c, mockOAuthExchange, mockClientID, mockClientSecret, getTestKey)

	// Check cookie properties
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies were set")
	}

	var tokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "github_token" {
			tokenCookie = cookie
			break
		}
	}

	if tokenCookie == nil {
		t.Fatal("github_token cookie was not found")
	}

	// Verify all cookie properties
	if value, _ := encryption.DecryptStringWrapper(tokenCookie.Value, getTestKey); value != "test-token-123" {
		t.Errorf("Cookie value = %q, expected %q", tokenCookie.Value, "test-token-123")
	}

	if tokenCookie.MaxAge != 3600 {
		t.Errorf("Cookie MaxAge = %d, expected 3600", tokenCookie.MaxAge)
	}

	if tokenCookie.Path != "/" {
		t.Errorf("Cookie Path = %q, expected %q", tokenCookie.Path, "/")
	}

	if !tokenCookie.HttpOnly {
		t.Error("Cookie should be HttpOnly")
	}

	// Secure is set to false in the code, but in production should be true
	if tokenCookie.Secure {
		t.Log("Note: Cookie Secure flag is true (good for production)")
	}
}
