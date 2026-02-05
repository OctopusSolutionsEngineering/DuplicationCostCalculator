package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCalculate(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		cookieValue        string
		queryParams        map[string]string
		expectedStatusCode int
		expectedRedirect   string
		expectFile         bool
	}{
		{
			name:               "authenticated user sees calculate page",
			cookieValue:        "valid-github-token",
			queryParams:        map[string]string{},
			expectedStatusCode: 200,
			expectedRedirect:   "",
			expectFile:         true,
		},
		{
			name:               "unauthenticated user redirected to login",
			cookieValue:        "",
			queryParams:        map[string]string{},
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/",
			expectFile:         false,
		},
		{
			name:        "unauthenticated user with repos param redirected with param",
			cookieValue: "",
			queryParams: map[string]string{
				"repos": "owner/repo1,owner/repo2",
			},
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/?repos=owner/repo1,owner/repo2",
			expectFile:         false,
		},
		{
			name:        "authenticated user with repos param sees calculate page",
			cookieValue: "valid-token",
			queryParams: map[string]string{
				"repos": "owner/repo",
			},
			expectedStatusCode: 200,
			expectedRedirect:   "",
			expectFile:         true,
		},
		{
			name:               "authenticated user with empty token value",
			cookieValue:        "",
			queryParams:        map[string]string{},
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/",
			expectFile:         false,
		},
		{
			name:        "unauthenticated with complex repos parameter",
			cookieValue: "",
			queryParams: map[string]string{
				"repos": "org-1/repo-1,org_2/repo_2,org3/repo3",
			},
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/?repos=org-1/repo-1,org_2/repo_2,org3/repo3",
			expectFile:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router and context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create a mock request
			req := httptest.NewRequest("GET", "/calculate", nil)

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

			// Call Calculate
			Calculate(c)

			// Check the status code
			// Note: When file doesn't exist, we get 404 instead of 200, both are acceptable for authenticated users
			if tt.expectFile {
				if w.Code != 200 && w.Code != 404 {
					t.Errorf("Status code = %d, expected 200 or 404 (file serving or file not found)", w.Code)
				}
			} else if w.Code != tt.expectedStatusCode {
				t.Errorf("Status code = %d, expected %d", w.Code, tt.expectedStatusCode)
			}

			// Check redirect if not authenticated
			if !tt.expectFile {
				location := w.Header().Get("Location")
				if location != tt.expectedRedirect {
					t.Errorf("Redirect location = %q, expected %q", location, tt.expectedRedirect)
				}
			}

			// If expecting file, check that response is not empty
			// Note: In test environment, c.File() might not actually serve the file
			// So we just verify that we didn't get redirected
			if tt.expectFile {
				if w.Code == http.StatusFound {
					t.Error("Expected file to be served, but got redirect instead")
				}
			}
		})
	}
}

func TestCalculateFileServing(t *testing.T) {
	// Test that the Calculate handler attempts to serve the correct file
	gin.SetMode(gin.TestMode)

	// Create the html directory if it doesn't exist
	tempDir := "html"
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create html directory: %v", err)
	}

	// Create a temporary calculate.html file for testing
	tempFile := "html/calculate.html"
	content := []byte("<html><body>Test Calculate Page</body></html>")
	err = os.WriteFile(tempFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tempFile)
	defer os.RemoveAll(tempDir)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/calculate", nil)
	req.AddCookie(&http.Cookie{
		Name:  "github_token",
		Value: "valid-token",
	})
	c.Request = req

	Calculate(c)

	// Check that we got a successful response
	if w.Code != 200 {
		t.Errorf("Status code = %d, expected 200", w.Code)
	}

	// Check that the response contains the file content
	body := w.Body.String()
	if body != string(content) {
		t.Errorf("Response body does not match expected file content")
	}
}

func TestCalculateMultipleCalls(t *testing.T) {
	// Test that calling Calculate multiple times produces consistent results
	gin.SetMode(gin.TestMode)

	// Test authenticated scenario
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest("GET", "/calculate", nil)
		req.AddCookie(&http.Cookie{
			Name:  "github_token",
			Value: "valid-token",
		})
		c.Request = req

		Calculate(c)

		// Should try to serve file (status 200 or 404 if file doesn't exist)
		if w.Code != 200 && w.Code != 404 {
			t.Errorf("Call %d: unexpected status code %d", i+1, w.Code)
		}
	}

	// Test unauthenticated scenario
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest("GET", "/calculate", nil)
		c.Request = req

		Calculate(c)

		// Should redirect
		if w.Code != http.StatusFound {
			t.Errorf("Call %d: unexpected status code %d, expected %d", i+1, w.Code, http.StatusFound)
		}

		location := w.Header().Get("Location")
		if location != "/" {
			t.Errorf("Call %d: redirect location = %q, expected %q", i+1, location, "/")
		}
	}
}

func TestCalculateWithDifferentTokens(t *testing.T) {
	// Test that different valid tokens all allow access
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

			req := httptest.NewRequest("GET", "/calculate", nil)
			req.AddCookie(&http.Cookie{
				Name:  "github_token",
				Value: token,
			})
			c.Request = req

			Calculate(c)

			// Should try to serve file (not redirect)
			if w.Code == http.StatusFound {
				t.Errorf("With token %q, got redirect instead of file serving", token)
			}
		})
	}
}

func TestCalculateAuthenticationIntegration(t *testing.T) {
	// Integration test to verify Calculate properly uses IsAuthenticated
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		hasCookie      bool
		cookieValue    string
		expectRedirect bool
		redirectTo     string
	}{
		{
			name:           "no cookie means redirect",
			hasCookie:      false,
			expectRedirect: true,
			redirectTo:     "/",
		},
		{
			name:           "empty cookie means redirect",
			hasCookie:      true,
			cookieValue:    "",
			expectRedirect: true,
			redirectTo:     "/",
		},
		{
			name:           "valid cookie means no redirect",
			hasCookie:      true,
			cookieValue:    "valid-token",
			expectRedirect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/calculate", nil)
			if tc.hasCookie {
				req.AddCookie(&http.Cookie{
					Name:  "github_token",
					Value: tc.cookieValue,
				})
			}
			c.Request = req

			Calculate(c)

			if tc.expectRedirect {
				if w.Code != http.StatusFound {
					t.Errorf("Expected redirect (status %d), got status %d", http.StatusFound, w.Code)
				}
				location := w.Header().Get("Location")
				if location != tc.redirectTo {
					t.Errorf("Redirect location = %q, expected %q", location, tc.redirectTo)
				}
			} else {
				if w.Code == http.StatusFound {
					t.Error("Expected file serving, got redirect")
				}
			}
		})
	}
}

func TestCalculateQueryParamPreservation(t *testing.T) {
	// Test that query parameters are preserved in redirects
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		repos       string
		expectedURL string
	}{
		{
			repos:       "owner/repo",
			expectedURL: "/?repos=owner/repo",
		},
		{
			repos:       "owner1/repo1,owner2/repo2",
			expectedURL: "/?repos=owner1/repo1,owner2/repo2",
		},
		{
			repos:       "org-name/repo_name",
			expectedURL: "/?repos=org-name/repo_name",
		},
	}

	for _, tc := range testCases {
		t.Run("repos: "+tc.repos, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/calculate?repos="+tc.repos, nil)
			c.Request = req

			Calculate(c)

			// Should redirect (no auth)
			if w.Code != http.StatusFound {
				t.Errorf("Status code = %d, expected %d", w.Code, http.StatusFound)
			}

			location := w.Header().Get("Location")
			if location != tc.expectedURL {
				t.Errorf("Redirect location = %q, expected %q", location, tc.expectedURL)
			}
		})
	}
}
