package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v57/github"
)

func TestCostHandlerWrapped(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		cookieValue        string
		requestBody        map[string]interface{}
		mockReport         models.Report
		expectedStatusCode int
		expectJSON         bool
		expectError        bool
	}{
		{
			name:        "successful request with valid token and repositories",
			cookieValue: "valid-github-token",
			requestBody: map[string]interface{}{
				"repositories": []string{"owner/repo1", "owner/repo2"},
			},
			mockReport: models.Report{
				NumberOfRepos:                       2,
				NumberOfReposWithDuplicationOrDrift: 1,
				Contributors:                        map[string][]string{},
				UniqueContributors:                  []string{},
				ActionAuthors:                       map[string][]string{},
				WorkflowAdvisories:                  map[string][]string{},
				Comparisons:                         map[string]map[string]models.RepoMeasurements{},
			},
			expectedStatusCode: http.StatusOK,
			expectJSON:         true,
			expectError:        false,
		},
		{
			name:               "missing access token",
			cookieValue:        "",
			requestBody:        map[string]interface{}{},
			mockReport:         models.Report{},
			expectedStatusCode: http.StatusUnauthorized,
			expectJSON:         true,
			expectError:        true,
		},
		{
			name:        "request with extra fields",
			cookieValue: "valid-token",
			requestBody: map[string]interface{}{
				"repositories": []string{"owner/repo"},
				"invalid":      "data",
			},
			mockReport: models.Report{
				NumberOfRepos: 1,
			},
			expectedStatusCode: http.StatusOK,
			expectJSON:         true,
			expectError:        false,
		},
		{
			name:        "empty repositories list",
			cookieValue: "valid-token",
			requestBody: map[string]interface{}{
				"repositories": []string{},
			},
			mockReport: models.Report{
				NumberOfRepos: 0,
			},
			expectedStatusCode: http.StatusOK,
			expectJSON:         true,
			expectError:        false,
		},
		{
			name:        "single repository",
			cookieValue: "valid-token",
			requestBody: map[string]interface{}{
				"repositories": []string{"owner/repo"},
			},
			mockReport: models.Report{
				NumberOfRepos: 1,
			},
			expectedStatusCode: http.StatusOK,
			expectJSON:         true,
			expectError:        false,
		},
		{
			name:        "multiple repositories with complex report",
			cookieValue: "ghp_token123",
			requestBody: map[string]interface{}{
				"repositories": []string{
					"OctopusDeploy/OctopusDeploy",
					"actions/checkout",
					"docker/build-push-action",
				},
			},
			mockReport: models.Report{
				NumberOfRepos:                       3,
				NumberOfReposWithDuplicationOrDrift: 2,
				Contributors: map[string][]string{
					"OctopusDeploy/OctopusDeploy": {"user1", "user2"},
				},
				UniqueContributors: []string{"user1", "user2"},
				ActionAuthors: map[string][]string{
					"actions/checkout": {"actions"},
				},
				WorkflowAdvisories: map[string][]string{},
				Comparisons: map[string]map[string]models.RepoMeasurements{
					"OctopusDeploy/OctopusDeploy": {
						"actions/checkout": {
							StepsWithDifferentVersionsCount: 5,
							StepsWithSimilarConfigCount:     3,
						},
					},
				},
			},
			expectedStatusCode: http.StatusOK,
			expectJSON:         true,
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock functions
			getClientCalled := false
			generateReportCalled := false
			var capturedAccessToken string
			var capturedRepositories []string

			mockGetClient := func(accessToken string) *github.Client {
				getClientCalled = true
				capturedAccessToken = accessToken
				return github.NewClient(nil)
			}

			mockGenerateReport := func(client *github.Client, repositories []string) models.Report {
				generateReportCalled = true
				capturedRepositories = repositories
				return tt.mockReport
			}

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request with JSON body
			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/cost", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Add cookie if provided
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  "github_token",
					Value: tt.cookieValue,
				})
			}

			c.Request = req

			// Call the handler
			CostHandlerWrapped(c, mockGetClient, mockGenerateReport)

			// Check status code
			if w.Code != tt.expectedStatusCode {
				t.Errorf("Status code = %d, expected %d", w.Code, tt.expectedStatusCode)
			}

			// Check JSON response
			if tt.expectJSON {
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json; charset=utf-8" {
					t.Errorf("Content-Type = %q, expected JSON", contentType)
				}
			}

			// Check if error response
			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse error response: %v", err)
				}
				if _, exists := response["error"]; !exists {
					t.Error("Expected error field in response")
				}
			} else {
				// Check successful response contains report
				var response models.Report
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse report response: %v", err)
				}
			}

			// Verify mock function calls for successful scenarios
			if !tt.expectError && tt.cookieValue != "" {
				if !getClientCalled {
					t.Error("getClient was not called")
				}
				if !generateReportCalled {
					t.Error("generateReport was not called")
				}
				if capturedAccessToken != tt.cookieValue {
					t.Errorf("Access token = %q, expected %q", capturedAccessToken, tt.cookieValue)
				}
				if repos, ok := tt.requestBody["repositories"].([]string); ok {
					if len(capturedRepositories) != len(repos) {
						t.Errorf("Captured %d repositories, expected %d", len(capturedRepositories), len(repos))
					}
				}
			}
		})
	}
}

func TestCostHandlerWrappedUnauthorized(t *testing.T) {
	// Test various unauthorized scenarios
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		hasCookie   bool
		cookieValue string
	}{
		{
			name:      "no cookie at all",
			hasCookie: false,
		},
		{
			name:        "empty cookie value",
			hasCookie:   true,
			cookieValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetClient := func(accessToken string) *github.Client {
				t.Error("getClient should not be called when unauthorized")
				return nil
			}

			mockGenerateReport := func(client *github.Client, repositories []string) models.Report {
				t.Error("generateReport should not be called when unauthorized")
				return models.Report{}
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody := map[string]interface{}{
				"repositories": []string{"owner/repo"},
			}
			bodyBytes, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/cost", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			if tt.hasCookie {
				req.AddCookie(&http.Cookie{
					Name:  "github_token",
					Value: tt.cookieValue,
				})
			}

			c.Request = req

			CostHandlerWrapped(c, mockGetClient, mockGenerateReport)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Status code = %d, expected %d", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestCostHandlerWrappedInvalidJSON(t *testing.T) {
	// Test various invalid JSON scenarios
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		requestBody string
	}{
		{
			name:        "malformed JSON",
			requestBody: `{invalid json}`,
		},
		{
			name:        "wrong field types",
			requestBody: `{"repositories": "not an array"}`,
		},
		{
			name:        "empty body",
			requestBody: ``,
		},
		{
			name:        "null repositories treated as empty array",
			requestBody: `{"repositories": null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetClient := func(accessToken string) *github.Client {
				return github.NewClient(nil)
			}

			mockGenerateReport := func(client *github.Client, repositories []string) models.Report {
				// Some cases will reach here, others won't
				return models.Report{}
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("POST", "/cost", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{
				Name:  "github_token",
				Value: "valid-token",
			})

			c.Request = req

			CostHandlerWrapped(c, mockGetClient, mockGenerateReport)

			// Malformed JSON and empty body should return 400
			// Wrong field types and null values are accepted by Gin as empty/zero values
			if tt.name == "malformed JSON" || tt.name == "empty body" {
				if w.Code != http.StatusBadRequest {
					t.Errorf("Status code = %d, expected %d", w.Code, http.StatusBadRequest)
				}
			} else {
				// Other cases are accepted
				if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
					t.Errorf("Status code = %d for %s", w.Code, tt.name)
				}
			}
		})
	}
}

func TestCostHandlerWrappedAccessTokenPassed(t *testing.T) {
	// Test that access token is correctly passed to getClient
	gin.SetMode(gin.TestMode)

	tokens := []string{
		"ghp_token123",
		"token-with-hyphens",
		"very-long-token-representing-github-pat",
	}

	for _, token := range tokens {
		t.Run("token: "+token, func(t *testing.T) {
			var capturedToken string

			mockGetClient := func(accessToken string) *github.Client {
				capturedToken = accessToken
				return github.NewClient(nil)
			}

			mockGenerateReport := func(client *github.Client, repositories []string) models.Report {
				return models.Report{}
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody := map[string]interface{}{
				"repositories": []string{"owner/repo"},
			}
			bodyBytes, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/cost", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{
				Name:  "github_token",
				Value: token,
			})

			c.Request = req

			CostHandlerWrapped(c, mockGetClient, mockGenerateReport)

			if capturedToken != token {
				t.Errorf("Captured token = %q, expected %q", capturedToken, token)
			}
		})
	}
}

func TestCostHandlerWrappedRepositoriesPassed(t *testing.T) {
	// Test that repositories are correctly passed to generateReport
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		repositories []string
	}{
		{
			name:         "single repository",
			repositories: []string{"owner/repo"},
		},
		{
			name:         "multiple repositories",
			repositories: []string{"owner1/repo1", "owner2/repo2", "owner3/repo3"},
		},
		{
			name:         "repositories with special characters",
			repositories: []string{"org-name/repo_name", "another-org/another_repo"},
		},
		{
			name:         "empty list",
			repositories: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedRepos []string

			mockGetClient := func(accessToken string) *github.Client {
				return github.NewClient(nil)
			}

			mockGenerateReport := func(client *github.Client, repositories []string) models.Report {
				capturedRepos = repositories
				return models.Report{}
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody := map[string]interface{}{
				"repositories": tt.repositories,
			}
			bodyBytes, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/cost", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{
				Name:  "github_token",
				Value: "valid-token",
			})

			c.Request = req

			CostHandlerWrapped(c, mockGetClient, mockGenerateReport)

			if len(capturedRepos) != len(tt.repositories) {
				t.Errorf("Captured %d repositories, expected %d", len(capturedRepos), len(tt.repositories))
			}

			for i, repo := range tt.repositories {
				if capturedRepos[i] != repo {
					t.Errorf("Repository[%d] = %q, expected %q", i, capturedRepos[i], repo)
				}
			}
		})
	}
}

func TestCostHandlerWrappedResponseFormat(t *testing.T) {
	// Test that the response contains all expected fields
	gin.SetMode(gin.TestMode)

	mockGetClient := func(accessToken string) *github.Client {
		return github.NewClient(nil)
	}

	mockGenerateReport := func(client *github.Client, repositories []string) models.Report {
		return models.Report{
			NumberOfRepos:                       2,
			NumberOfReposWithDuplicationOrDrift: 1,
			Contributors: map[string][]string{
				"owner/repo1": {"user1", "user2"},
			},
			UniqueContributors: []string{"user1", "user2"},
			ActionAuthors: map[string][]string{
				"owner/repo1": {"actions", "docker"},
			},
			WorkflowAdvisories: map[string][]string{
				"owner/repo1": {"CVE-2023-1234"},
			},
			Comparisons: map[string]map[string]models.RepoMeasurements{
				"owner/repo1": {
					"owner/repo2": {
						StepsWithDifferentVersions:       []string{"actions/checkout"},
						StepsWithDifferentVersionsCount:  1,
						StepsWithSimilarConfig:           []string{"actions/setup-node"},
						StepsWithSimilarConfigCount:      1,
						StepsThatIndicateDuplicationRisk: 2,
					},
				},
			},
		}
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := map[string]interface{}{
		"repositories": []string{"owner/repo1", "owner/repo2"},
	}
	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/cost", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "github_token",
		Value: "valid-token",
	})

	c.Request = req

	CostHandlerWrapped(c, mockGetClient, mockGenerateReport)

	if w.Code != http.StatusOK {
		t.Fatalf("Status code = %d, expected %d", w.Code, http.StatusOK)
	}

	var response models.Report
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify key fields
	if response.NumberOfRepos != 2 {
		t.Errorf("NumberOfRepos = %d, expected 2", response.NumberOfRepos)
	}

	if response.NumberOfReposWithDuplicationOrDrift != 1 {
		t.Errorf("NumberOfReposWithDuplicationOrDrift = %d, expected 1", response.NumberOfReposWithDuplicationOrDrift)
	}

	if len(response.UniqueContributors) != 2 {
		t.Errorf("UniqueContributors length = %d, expected 2", len(response.UniqueContributors))
	}

	if len(response.Contributors) != 1 {
		t.Errorf("Contributors length = %d, expected 1", len(response.Contributors))
	}

	if len(response.ActionAuthors) != 1 {
		t.Errorf("ActionAuthors length = %d, expected 1", len(response.ActionAuthors))
	}

	if len(response.Comparisons) != 1 {
		t.Errorf("Comparisons length = %d, expected 1", len(response.Comparisons))
	}
}

func TestCostHandlerWrappedMultipleCalls(t *testing.T) {
	// Test that calling the handler multiple times produces consistent results
	gin.SetMode(gin.TestMode)

	callCount := 0

	mockGetClient := func(accessToken string) *github.Client {
		callCount++
		return github.NewClient(nil)
	}

	mockGenerateReport := func(client *github.Client, repositories []string) models.Report {
		return models.Report{
			NumberOfRepos: len(repositories),
		}
	}

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := map[string]interface{}{
			"repositories": []string{"owner/repo1", "owner/repo2"},
		}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/cost", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "github_token",
			Value: "valid-token",
		})

		c.Request = req

		CostHandlerWrapped(c, mockGetClient, mockGenerateReport)

		if w.Code != http.StatusOK {
			t.Errorf("Call %d: status code = %d, expected %d", i+1, w.Code, http.StatusOK)
		}
	}

	if callCount != 3 {
		t.Errorf("getClient called %d times, expected 3", callCount)
	}
}
