package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ... (ERPServiceClient, CourseResponse, RegistrationResponse, ListCourses, CreateCourse, RegisterForCourse remain the same) ...

// --- NEW FUNCTION ---
// ListMyRegistrations calls the ERP service to get the logged-in user's registrations.
func (c *ERPServiceClient) ListMyRegistrations(token string) ([]RegistrationResponse, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/registrations/me", nil)
	req.Header.Add("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call erp service for myRegistrations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erp service returned an error for myRegistrations: %s - %s", resp.Status, string(bodyBytes))
	}

	var registrations []RegistrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&registrations); err != nil {
		return nil, fmt.Errorf("failed to decode myRegistrations response: %w", err)
	}
	return registrations, nil
}

// ... (existing code below) ...
type ERPServiceClient struct {
	BaseURL string
}

type CourseResponse struct {
	ID           string `json:"ID"`
	CourseCode   string `json:"CourseCode"`
	Name         string `json:"Name"`
	Description  string `json:"Description"`
	Credits      int    `json:"Credits"`
	InstructorID string `json:"InstructorID"`
}

type RegistrationResponse struct {
	UserID   string `json:"UserID"`
	CourseID string `json:"CourseID"`
	Semester string `json:"Semester"`
}

func (c *ERPServiceClient) ListCourses(token string) ([]CourseResponse, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/courses/", nil) // Added trailing slash for consistency
	req.Header.Add("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call erp service for courses: %w", err)
	}
	defer resp.Body.Close()
	var courses []CourseResponse
	if err := json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return nil, fmt.Errorf("failed to decode courses response: %w", err)
	}
	return courses, nil
}

func (c *ERPServiceClient) CreateCourse(adminToken string, courseData interface{}) (*CourseResponse, error) {
	createURL := c.BaseURL + "/courses"
	requestBody, err := json.Marshal(courseData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create course request: %w", err)
	}
	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for CreateCourse: %w", err)
	}
	req.Header.Add("Authorization", adminToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call erp service for CreateCourse: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erp service returned an error for CreateCourse: %s - %s", resp.Status, string(bodyBytes))
	}
	var courseResponse CourseResponse
	if err := json.NewDecoder(resp.Body).Decode(&courseResponse); err != nil {
		return nil, fmt.Errorf("failed to decode CreateCourse response: %w", err)
	}
	return &courseResponse, nil
}

func (c *ERPServiceClient) RegisterForCourse(token, courseID, semester string) (*RegistrationResponse, error) {
	requestBody, _ := json.Marshal(map[string]string{
		"courseId": courseID,
		"semester": semester,
	})
	req, _ := http.NewRequest("POST", c.BaseURL+"/registrations", bytes.NewBuffer(requestBody))
	req.Header.Add("Authorization", token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call erp service for registration: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erp service returned an error for registration: %s - %s", resp.Status, string(bodyBytes))
	}
	return &RegistrationResponse{
		UserID:   "",
		CourseID: courseID,
		Semester: semester,
	}, nil
}
