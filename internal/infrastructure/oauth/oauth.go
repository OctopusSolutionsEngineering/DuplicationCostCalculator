package oauth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func ExchangeCodeForToken(clientID string, clientSecret string, code string, url string) (TokenResponse, error) {
	return ExchangeCodeForTokenWrapper(http.DefaultTransport, clientID, clientSecret, code, url)
}

func ExchangeCodeForTokenWrapper(transport http.RoundTripper, clientID string, clientSecret string, code string, url string) (TokenResponse, error) {
	// Prepare request to exchange code for token
	tokenRequest := TokenRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Code:         code,
	}

	requestBody, _ := json.Marshal(tokenRequest)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return TokenResponse{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	tokenResponse := TokenResponse{}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return TokenResponse{}, err
	}

	return tokenResponse, nil
}
