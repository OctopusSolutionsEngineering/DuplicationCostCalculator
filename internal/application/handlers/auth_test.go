package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireAuth(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		cookieValue        string
		queryParams        map[string]string
		expectedResult     bool
		expectedRedirect   string
		expectedStatusCode int
	}{
		{
			name:               "authenticated with valid token",
			cookieValue:        "valid-github-token",
			queryParams:        map[string]string{},
			expectedResult:     true,
			expectedRedirect:   "",
			expectedStatusCode: 0,
		},
		{
			name:               "not authenticated without token",
			cookieValue:        "",
			queryParams:        map[string]string{},
			expectedResult:     false,
			expectedRedirect:   "/",
			expectedStatusCode: http.StatusFound,
		},
		{
			name:        "not authenticated with repos query param",
			cookieValue: "",
			queryParams: map[string]string{
				"repos": "owner/repo1,owner/repo2",
			},
			expectedResult:     false,
			expectedRedirect:   "/?repos=owner/repo1,owner/repo2",
			expectedStatusCode: http.StatusFound,
		},
		{
			name:        "authenticated with repos query param (should not redirect)",
			cookieValue: "valid-token",
			queryParams: map[string]string{
				"repos": "owner/repo",
			},
			expectedResult:     true,
			expectedRedirect:   "",
			expectedStatusCode: 0,
		},
		{
			name:        "not authenticated with multiple query params",
			cookieValue: "",
			queryParams: map[string]string{
				"repos": "owner/repo1,owner/repo2,owner/repo3",
			},
			expectedResult:     false,
			expectedRedirect:   "/?repos=owner/repo1,owner/repo2,owner/repo3",
			expectedStatusCode: http.StatusFound,
		},
		{
			name:        "not authenticated with special characters in repos param",
			cookieValue: "",
			queryParams: map[string]string{
				"repos": "org-name/repo-name",
			},
			expectedResult:     false,
			expectedRedirect:   "/?repos=org-name/repo-name",
			expectedStatusCode: http.StatusFound,
		},
		{
			name:               "authenticated with empty cookie value",
			cookieValue:        "",
			queryParams:        map[string]string{},
			expectedResult:     false,
			expectedRedirect:   "/",
			expectedStatusCode: http.StatusFound,
		},
		{
			name:               "authenticated with whitespace token",
			cookieValue:        "   ",
			queryParams:        map[string]string{},
			expectedResult:     true,
			expectedRedirect:   "",
			expectedStatusCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router and context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create a mock request
			req := httptest.NewRequest("GET", "/test", nil)

			// Add query parameters
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Add cookie if provided
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  "github_token",
					Value: tt.cookieValue,
				})
			}

			c.Request = req

			// Call IsAuthenticated
			result := IsAuthenticated(c)

			// Check the result
			if result != tt.expectedResult {
				t.Errorf("IsAuthenticated() = %v, expected %v", result, tt.expectedResult)
			}

			// Check redirect if authentication failed
			if !tt.expectedResult {
				if w.Code != tt.expectedStatusCode {
					t.Errorf("Status code = %d, expected %d", w.Code, tt.expectedStatusCode)
				}

				location := w.Header().Get("Location")
				if location != tt.expectedRedirect {
					t.Errorf("Redirect location = %q, expected %q", location, tt.expectedRedirect)
				}
			}
		})
	}
}

func TestRequireAuthNoCookie(t *testing.T) {
	// Test the case where no cookie is set at all
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	result := IsAuthenticated(c)

	if result != false {
		t.Errorf("IsAuthenticated() with no cookie = %v, expected false", result)
	}

	if w.Code != http.StatusFound {
		t.Errorf("Status code = %d, expected %d", w.Code, http.StatusFound)
	}

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("Redirect location = %q, expected %q", location, "/")
	}
}

func TestRequireAuthPreservesReposParam(t *testing.T) {
	// Test that the repos parameter is correctly preserved in the redirect
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name        string
		reposParam  string
		expectedURL string
	}{
		{
			name:        "single repo",
			reposParam:  "owner/repo",
			expectedURL: "/?repos=owner/repo",
		},
		{
			name:        "multiple repos",
			reposParam:  "owner1/repo1,owner2/repo2",
			expectedURL: "/?repos=owner1/repo1,owner2/repo2",
		},
		{
			name:        "repos with special characters",
			reposParam:  "org-name/repo_name",
			expectedURL: "/?repos=org-name/repo_name",
		},
		{
			name:        "empty repos param redirects to root",
			reposParam:  "",
			expectedURL: "/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/test", nil)
			if tc.reposParam != "" {
				q := req.URL.Query()
				q.Add("repos", tc.reposParam)
				req.URL.RawQuery = q.Encode()
			}
			c.Request = req

			result := IsAuthenticated(c)

			if result != false {
				t.Errorf("IsAuthenticated() = %v, expected false", result)
			}

			location := w.Header().Get("Location")
			if location != tc.expectedURL {
				t.Errorf("Redirect location = %q, expected %q", location, tc.expectedURL)
			}
		})
	}
}

func TestRequireAuthWithValidToken(t *testing.T) {
	// Test various valid token scenarios
	gin.SetMode(gin.TestMode)

	tokens := []string{
		"ghp_validtoken123456789",
		"token-with-hyphens",
		"token_with_underscores",
		"MixedCaseToken123",
		"very-long-token-that-represents-a-real-github-personal-access-token",
	}

	for _, token := range tokens {
		t.Run("token: "+token, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/test", nil)
			req.AddCookie(&http.Cookie{
				Name:  "github_token",
				Value: token,
			})
			c.Request = req

			result := IsAuthenticated(c)

			if result != true {
				t.Errorf("IsAuthenticated() with token %q = false, expected true", token)
			}

			// Should not redirect
			if w.Code != 0 && w.Code != 200 {
				t.Errorf("Unexpected status code %d for authenticated request", w.Code)
			}
		})
	}
}

func TestRequireAuthConsistency(t *testing.T) {
	// Test that calling IsAuthenticated multiple times with the same context produces consistent results
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "github_token",
		Value: "valid-token",
	})
	c.Request = req

	result1 := IsAuthenticated(c)

	// Reset the writer for the second call
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	result2 := IsAuthenticated(c)

	// Reset again for the third call
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	result3 := IsAuthenticated(c)

	if result1 != result2 || result2 != result3 {
		t.Errorf("IsAuthenticated() produced inconsistent results: %v, %v, %v", result1, result2, result3)
	}

	if !result1 || !result2 || !result3 {
		t.Errorf("Expected all results to be true, got: %v, %v, %v", result1, result2, result3)
	}
}
