package oauth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func ExchangeCodeForToken(clientID string, clientSecret string, code string, url string) (TokenResponse, error) {
	// Prepare request to exchange code for token
	requestBody, _ := json.Marshal(map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return TokenResponse{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
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
