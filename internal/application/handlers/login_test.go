package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLogin(t *testing.T) {
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
			name:               "authenticated user redirected to repos",
			cookieValue:        "valid-github-token",
			queryParams:        map[string]string{},
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/repos",
			expectFile:         false,
		},
		{
			name:               "unauthenticated user sees login page",
			cookieValue:        "",
			queryParams:        map[string]string{},
			expectedStatusCode: 200,
			expectedRedirect:   "",
			expectFile:         true,
		},
		{
			name:        "authenticated user with repos param redirected with param",
			cookieValue: "valid-token",
			queryParams: map[string]string{
				"repos": "owner/repo1,owner/repo2",
			},
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/repos?repos=owner/repo1,owner/repo2",
			expectFile:         false,
		},
		{
			name:        "unauthenticated user with repos param sees login page",
			cookieValue: "",
			queryParams: map[string]string{
				"repos": "owner/repo",
			},
			expectedStatusCode: 200,
			expectedRedirect:   "",
			expectFile:         true,
		},
		{
			name:               "empty cookie shows login page",
			cookieValue:        "",
			queryParams:        map[string]string{},
			expectedStatusCode: 200,
			expectedRedirect:   "",
			expectFile:         true,
		},
		{
			name:        "authenticated with complex repos parameter",
			cookieValue: "valid-token",
			queryParams: map[string]string{
				"repos": "org-1/repo-1,org_2/repo_2,org3/repo3",
			},
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/repos?repos=org-1/repo-1,org_2/repo_2,org3/repo3",
			expectFile:         false,
		},
		{
			name:        "authenticated with special characters in repos",
			cookieValue: "token123",
			queryParams: map[string]string{
				"repos": "OctopusDeploy/OctopusDeploy",
			},
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   "/repos?repos=OctopusDeploy/OctopusDeploy",
			expectFile:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router and context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create a mock request
			req := httptest.NewRequest("GET", "/", nil)

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

			// Call Login
			Login(c)

			// Check the status code
			// Note: When file doesn't exist, we get 404 instead of 200, both are acceptable for unauthenticated users
			if tt.expectFile {
				if w.Code != 200 && w.Code != 404 {
					t.Errorf("Status code = %d, expected 200 or 404 (file serving or file not found)", w.Code)
				}
			} else if w.Code != tt.expectedStatusCode {
				t.Errorf("Status code = %d, expected %d", w.Code, tt.expectedStatusCode)
			}

			// Check redirect if authenticated
			if !tt.expectFile {
				location := w.Header().Get("Location")
				if location != tt.expectedRedirect {
					t.Errorf("Redirect location = %q, expected %q", location, tt.expectedRedirect)
				}
			}

			// If expecting file, check that we didn't get redirected
			if tt.expectFile {
				if w.Code == http.StatusFound {
					t.Error("Expected file to be served, but got redirect instead")
				}
			}
		})
	}
}

func TestLoginFileServing(t *testing.T) {
	// Test that the Login handler attempts to serve the correct file
	gin.SetMode(gin.TestMode)

	// Create a temporary index.html file for testing
	tempFile := "index.html"
	content := []byte("<html><body>Test Login Page</body></html>")
	err := os.WriteFile(tempFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tempFile)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/", nil)
	// No cookie - should serve login page
	c.Request = req

	Login(c)

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

func TestLoginAuthenticatedRedirect(t *testing.T) {
	// Test that authenticated users are always redirected, never shown the login page
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name        string
		token       string
		repos       string
		expectedURL string
	}{
		{
			name:        "authenticated without repos",
			token:       "valid-token",
			repos:       "",
			expectedURL: "/repos",
		},
		{
			name:        "authenticated with single repo",
			token:       "valid-token",
			repos:       "owner/repo",
			expectedURL: "/repos?repos=owner/repo",
		},
		{
			name:        "authenticated with multiple repos",
			token:       "valid-token",
			repos:       "owner1/repo1,owner2/repo2",
			expectedURL: "/repos?repos=owner1/repo1,owner2/repo2",
		},
		{
			name:        "authenticated with complex repos",
			token:       "ghp_token123",
			repos:       "org-name/repo_name,another-org/another_repo",
			expectedURL: "/repos?repos=org-name/repo_name,another-org/another_repo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/", nil)
			if tc.repos != "" {
				q := req.URL.Query()
				q.Add("repos", tc.repos)
				req.URL.RawQuery = q.Encode()
			}
			req.AddCookie(&http.Cookie{
				Name:  "github_token",
				Value: tc.token,
			})
			c.Request = req

			Login(c)

			// Should redirect
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

func TestLoginUnauthenticatedShowsPage(t *testing.T) {
	// Test that unauthenticated users are shown the login page, not redirected
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name      string
		hasCookie bool
		repos     string
	}{
		{
			name:      "no cookie, no repos",
			hasCookie: false,
			repos:     "",
		},
		{
			name:      "no cookie, with repos",
			hasCookie: false,
			repos:     "owner/repo",
		},
		{
			name:      "empty cookie, no repos",
			hasCookie: true,
			repos:     "",
		},
		{
			name:      "empty cookie, with repos",
			hasCookie: true,
			repos:     "owner/repo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/", nil)
			if tc.repos != "" {
				q := req.URL.Query()
				q.Add("repos", tc.repos)
				req.URL.RawQuery = q.Encode()
			}
			if tc.hasCookie {
				req.AddCookie(&http.Cookie{
					Name:  "github_token",
					Value: "", // Empty cookie value
				})
			}
			c.Request = req

			Login(c)

			// Should try to serve file (not redirect)
			if w.Code == http.StatusFound {
				t.Error("Expected file serving, got redirect instead")
			}

			// Should get 200 or 404 (file exists or not)
			if w.Code != 200 && w.Code != 404 {
				t.Errorf("Status code = %d, expected 200 or 404", w.Code)
			}
		})
	}
}

func TestLoginWithDifferentTokens(t *testing.T) {
	// Test that different valid tokens all trigger redirect
	gin.SetMode(gin.TestMode)

	tokens := []string{
		"ghp_validtoken123456789",
		"token-with-hyphens",
		"token_with_underscores",
		"MixedCaseToken123",
		"very-long-token-that-represents-a-real-github-personal-access-token",
		" ", // Even whitespace-only token is considered valid
	}

	for _, token := range tokens {
		t.Run("token: "+token, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/", nil)
			req.AddCookie(&http.Cookie{
				Name:  "github_token",
				Value: token,
			})
			c.Request = req

			Login(c)

			// Should redirect (not serve file)
			if w.Code != http.StatusFound {
				t.Errorf("With token %q, got status %d instead of redirect", token, w.Code)
			}

			location := w.Header().Get("Location")
			if location != "/repos" {
				t.Errorf("With token %q, redirect location = %q, expected %q", token, location, "/repos")
			}
		})
	}
}

func TestLoginQueryParamPreservation(t *testing.T) {
	// Test that repos query parameter is preserved for authenticated users
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		repos       string
		expectedURL string
	}{
		{
			repos:       "owner/repo",
			expectedURL: "/repos?repos=owner/repo",
		},
		{
			repos:       "owner1/repo1,owner2/repo2",
			expectedURL: "/repos?repos=owner1/repo1,owner2/repo2",
		},
		{
			repos:       "org-name/repo_name",
			expectedURL: "/repos?repos=org-name/repo_name",
		},
		{
			repos:       "actions/checkout,docker/build-push-action",
			expectedURL: "/repos?repos=actions/checkout,docker/build-push-action",
		},
	}

	for _, tc := range testCases {
		t.Run("repos: "+tc.repos, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/?repos="+tc.repos, nil)
			req.AddCookie(&http.Cookie{
				Name:  "github_token",
				Value: "valid-token",
			})
			c.Request = req

			Login(c)

			// Should redirect
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

func TestLoginNoCookie(t *testing.T) {
	// Test the case where no cookie is set at all
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/", nil)
	c.Request = req

	Login(c)

	// Should try to serve file (not redirect)
	if w.Code == http.StatusFound {
		t.Error("Expected file serving, got redirect instead")
	}

	// Should get 200 or 404
	if w.Code != 200 && w.Code != 404 {
		t.Errorf("Status code = %d, expected 200 or 404", w.Code)
	}
}

func TestLoginConsistency(t *testing.T) {
	// Test that calling Login multiple times with the same input produces consistent results
	gin.SetMode(gin.TestMode)

	// Test authenticated scenario
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "github_token",
			Value: "valid-token",
		})
		c.Request = req

		Login(c)

		if w.Code != http.StatusFound {
			t.Errorf("Call %d: status code = %d, expected %d", i+1, w.Code, http.StatusFound)
		}

		location := w.Header().Get("Location")
		if location != "/repos" {
			t.Errorf("Call %d: redirect location = %q, expected %q", i+1, location, "/repos")
		}
	}

	// Test unauthenticated scenario
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest("GET", "/", nil)
		c.Request = req

		Login(c)

		// Should try to serve file
		if w.Code == http.StatusFound {
			t.Errorf("Call %d: unexpected redirect", i+1)
		}
	}
}

func TestLoginReposParamIgnoredWhenUnauthenticated(t *testing.T) {
	// Test that repos parameter doesn't affect behavior for unauthenticated users
	gin.SetMode(gin.TestMode)

	testCases := []string{
		"",
		"owner/repo",
		"owner1/repo1,owner2/repo2,owner3/repo3",
		"very-long-list/of,repos/here,and/there",
	}

	for _, repos := range testCases {
		t.Run("repos: "+repos, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/", nil)
			if repos != "" {
				q := req.URL.Query()
				q.Add("repos", repos)
				req.URL.RawQuery = q.Encode()
			}
			c.Request = req

			Login(c)

			// Should always try to serve file for unauthenticated users
			if w.Code == http.StatusFound {
				t.Errorf("Got redirect for unauthenticated user with repos=%q", repos)
			}

			// Should get 200 or 404
			if w.Code != 200 && w.Code != 404 {
				t.Errorf("Status code = %d, expected 200 or 404", w.Code)
			}
		})
	}
}
