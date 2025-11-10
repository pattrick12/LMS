package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ... (AuthServiceClient, AuthResponse, UserResponse, Login, GetCurrentUser, GetUserByID remain the same) ...

// --- NEW FUNCTION ---
// CreateUser calls the auth service's admin endpoint to create a new user.
func (c *AuthServiceClient) CreateUser(adminToken string, userData interface{}) (*UserResponse, error) {
	createURL := fmt.Sprintf("%s/admin/users", c.BaseURL)
	requestBody, err := json.Marshal(userData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create user request: %w", err)
	}

	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for CreateUser: %w", err)
	}
	req.Header.Add("Authorization", adminToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service for CreateUser: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned an error for CreateUser: %s - %s", resp.Status, string(bodyBytes))
	}

	var responseData struct {
		UserID string `json:"userID"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode CreateUser response: %w", err)
	}

	// The create user endpoint doesn't return the full user, so we fetch it.
	return c.GetUserByID(adminToken, responseData.UserID)
}

// ... existing code below ...
type AuthServiceClient struct {
	BaseURL string
}

type AuthResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type UserResponse struct {
	ID       string `json:"ID"`
	Email    string `json:"Email"`
	Role     string `json:"Role"`
	FullName string `json:"FullName"`
}

func (c *AuthServiceClient) Login(email, password string) (*AuthResponse, error) {
	loginURL := fmt.Sprintf("%s/login", c.BaseURL)
	requestBody, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}
	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned an error: %s - %s", resp.Status, string(bodyBytes))
	}
	var authResponse AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}
	return &authResponse, nil
}

func (c *AuthServiceClient) GetCurrentUser(token string) (*UserResponse, error) {
	meURL := fmt.Sprintf("%s/users/me", c.BaseURL)
	req, err := http.NewRequest("GET", meURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service for current user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned an error for current user: %s - %s", resp.Status, string(bodyBytes))
	}
	var responseData struct {
		User UserResponse `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}
	return &responseData.User, nil
}

func (c *AuthServiceClient) GetUserByID(adminToken, userID string) (*UserResponse, error) {
	userURL := fmt.Sprintf("%s/admin/users/%s", c.BaseURL, userID)
	req, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for GetUserByID: %w", err)
	}
	req.Header.Add("Authorization", adminToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service for GetUserByID: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned an error for GetUserByID: %s", resp.Status)
	}
	var responseData struct {
		User UserResponse `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode user response from GetUserByID: %w", err)
	}
	return &responseData.User, nil
}
