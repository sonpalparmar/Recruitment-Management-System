package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type ResumeData struct {
	Education  string `json:"education"`
	Email      string `json:"email"`
	Experience string `json:"experience"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Skills     string `json:"skills"`
}

func ProcessResume(apiKey, filePath string) (*ResumeData, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Prepare the request
	url := "http://localhost:3000/resume_parser/upload"
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("resume", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("apikey", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// Parse JSON
	var apiResponse struct {
		Education []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"education"`
		Email      string `json:"email"`
		Experience []struct {
			Dates []string `json:"dates"`
			Name  string   `json:"name"`
			URL   string   `json:"url"`
		} `json:"experience"`
		Name   string   `json:"name"`
		Phone  string   `json:"phone"`
		Skills []string `json:"skills"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	// Convert to ResumeData
	resumeData := &ResumeData{
		Name:  apiResponse.Name,
		Email: apiResponse.Email,
		Phone: apiResponse.Phone,
		Skills: func() string {
			return stringifySlice(apiResponse.Skills)
		}(),
		Education:  stringifyEducation(apiResponse.Education),
		Experience: stringifyExperience(apiResponse.Experience),
	}

	return resumeData, nil
}

func stringifySlice(slice []string) string {
	return fmt.Sprintf("%v", slice)
}

func stringifyEducation(education []struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}) string {
	var names []string
	for _, edu := range education {
		names = append(names, edu.Name)
	}
	return fmt.Sprintf("%v", names)
}

func stringifyExperience(experience []struct {
	Dates []string `json:"dates"`
	Name  string   `json:"name"`
	URL   string   `json:"url"`
}) string {
	var experiences []string
	for _, exp := range experience {
		experiences = append(experiences, exp.Name+" ("+fmt.Sprintf("%v", exp.Dates)+")")
	}
	return fmt.Sprintf("%v", experiences)
}
