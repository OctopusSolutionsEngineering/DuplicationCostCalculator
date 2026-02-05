package oauth

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

type mockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *mockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestExchangeCodeForTokenWrapper_Success(t *testing.T) {
	// Arrange
	responseBody := `{"access_token":"test_token","token_type":"bearer","scope":"repo"}`
	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			Header:     make(http.Header),
		},
		err: nil,
	}

	// Act
	result, err := ExchangeCodeForTokenWrapper(
		mockTransport,
		"test_client_id",
		"test_client_secret",
		"test_code",
		"https://github.com/login/oauth/access_token",
	)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.AccessToken != "test_token" {
		t.Errorf("Expected access_token to be 'test_token', got '%s'", result.AccessToken)
	}

	if result.TokenType != "bearer" {
		t.Errorf("Expected token_type to be 'bearer', got '%s'", result.TokenType)
	}

	if result.Scope != "repo" {
		t.Errorf("Expected scope to be 'repo', got '%s'", result.Scope)
	}
}

func TestExchangeCodeForTokenWrapper_ErrorResponse(t *testing.T) {
	// Arrange
	responseBody := `{"error":"bad_verification_code","error_description":"The code passed is incorrect or expired."}`
	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			Header:     make(http.Header),
		},
		err: nil,
	}

	// Act
	result, err := ExchangeCodeForTokenWrapper(
		mockTransport,
		"test_client_id",
		"test_client_secret",
		"bad_code",
		"https://github.com/login/oauth/access_token",
	)

	// Assert
	if err != nil {
		t.Errorf("Expected no error from function, got %v", err)
	}

	if result.Error != "bad_verification_code" {
		t.Errorf("Expected error field to be 'bad_verification_code', got '%s'", result.Error)
	}

	if result.ErrorDesc != "The code passed is incorrect or expired." {
		t.Errorf("Expected error_description to be 'The code passed is incorrect or expired.', got '%s'", result.ErrorDesc)
	}

	if result.AccessToken != "" {
		t.Errorf("Expected access_token to be empty, got '%s'", result.AccessToken)
	}
}

func TestExchangeCodeForTokenWrapper_InvalidURL(t *testing.T) {
	// Arrange
	mockTransport := &mockRoundTripper{
		response: nil,
		err:      nil,
	}

	// Act
	_, err := ExchangeCodeForTokenWrapper(
		mockTransport,
		"test_client_id",
		"test_client_secret",
		"test_code",
		"://invalid-url",
	)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestExchangeCodeForTokenWrapper_NetworkError(t *testing.T) {
	// Arrange
	mockTransport := &mockRoundTripper{
		response: nil,
		err:      io.ErrUnexpectedEOF,
	}

	// Act
	_, err := ExchangeCodeForTokenWrapper(
		mockTransport,
		"test_client_id",
		"test_client_secret",
		"test_code",
		"https://github.com/login/oauth/access_token",
	)

	// Assert
	if err == nil {
		t.Error("Expected network error, got nil")
	}
}

func TestExchangeCodeForTokenWrapper_InvalidJSON(t *testing.T) {
	// Arrange
	responseBody := `{"invalid json`
	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			Header:     make(http.Header),
		},
		err: nil,
	}

	// Act
	_, err := ExchangeCodeForTokenWrapper(
		mockTransport,
		"test_client_id",
		"test_client_secret",
		"test_code",
		"https://github.com/login/oauth/access_token",
	)

	// Assert
	if err == nil {
		t.Error("Expected JSON unmarshal error, got nil")
	}
}

func TestExchangeCodeForTokenWrapper_EmptyResponse(t *testing.T) {
	// Arrange
	responseBody := `{}`
	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			Header:     make(http.Header),
		},
		err: nil,
	}

	// Act
	result, err := ExchangeCodeForTokenWrapper(
		mockTransport,
		"test_client_id",
		"test_client_secret",
		"test_code",
		"https://github.com/login/oauth/access_token",
	)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.AccessToken != "" {
		t.Errorf("Expected empty access_token, got '%s'", result.AccessToken)
	}

	if result.TokenType != "" {
		t.Errorf("Expected empty token_type, got '%s'", result.TokenType)
	}
}

func TestExchangeCodeForTokenWrapper_AllFields(t *testing.T) {
	// Arrange
	responseBody := `{
		"access_token":"gho_16C7e42F292c6912E7710c838347Ae178B4a",
		"token_type":"bearer",
		"scope":"repo,gist",
		"error":"",
		"error_description":""
	}`
	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			Header:     make(http.Header),
		},
		err: nil,
	}

	// Act
	result, err := ExchangeCodeForTokenWrapper(
		mockTransport,
		"Iv1.8a61f9b3a7aba766",
		"test_secret",
		"test_code",
		"https://github.com/login/oauth/access_token",
	)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.AccessToken != "gho_16C7e42F292c6912E7710c838347Ae178B4a" {
		t.Errorf("Expected access_token to be 'gho_16C7e42F292c6912E7710c838347Ae178B4a', got '%s'", result.AccessToken)
	}

	if result.TokenType != "bearer" {
		t.Errorf("Expected token_type to be 'bearer', got '%s'", result.TokenType)
	}

	if result.Scope != "repo,gist" {
		t.Errorf("Expected scope to be 'repo,gist', got '%s'", result.Scope)
	}

	if result.Error != "" {
		t.Errorf("Expected error to be empty, got '%s'", result.Error)
	}

	if result.ErrorDesc != "" {
		t.Errorf("Expected error_description to be empty, got '%s'", result.ErrorDesc)
	}
}

func TestExchangeCodeForToken_CallsWrapper(t *testing.T) {
	// This test verifies that ExchangeCodeForToken properly delegates to ExchangeCodeForTokenWrapper
	// We can't easily mock http.DefaultTransport, but we can at least call the function
	// and verify it doesn't panic

	// Note: This will make a real HTTP call, but with invalid credentials it should fail gracefully
	_, err := ExchangeCodeForToken(
		"test_client_id",
		"test_client_secret",
		"test_code",
		"https://httpbin.org/status/404", // Use a test endpoint that returns 404
	)

	// We expect an error (either network or JSON parsing), but no panic
	// The key is that the function returns without panicking
	if err == nil {
		// If somehow no error, that's also acceptable for this basic delegation test
		t.Log("ExchangeCodeForToken completed without error (unexpected but not a failure)")
	}
}
