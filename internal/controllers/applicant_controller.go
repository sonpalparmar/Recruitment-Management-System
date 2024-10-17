package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/GolangAssignment/internal/config"
	"github.com/GolangAssignment/internal/models"
	"github.com/GolangAssignment/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
	"gorm.io/gorm"
)

// ParsedData represents the structured data extracted from the resume.
type ParsedData struct {
	Education  string `json:"education"`
	Email      string `json:"email"`
	Experience string `json:"experience"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Skills     string `json:"skills"`
}

// ApplicantController handles applicant-related operations.
type ApplicantController struct {
	DB  *gorm.DB
	Cfg config.Config
}

// NewApplicantController creates a new instance of ApplicantController.
func NewApplicantController(db *gorm.DB, cfg config.Config) *ApplicantController {
	return &ApplicantController{DB: db, Cfg: cfg}
}

// UploadResume handles the resume upload, extraction, parsing, and profile updating.
func (ac *ApplicantController) UploadResume(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User ID not found")
		return
	}

	file, header, err := c.Request.FormFile("resume")
	if err != nil {
		log.Printf("Error getting resume file: %v", err)
		utils.RespondWithError(c, http.StatusBadRequest, "Resume file is required")
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" && ext != ".docx" {
		utils.RespondWithError(c, http.StatusBadRequest, "Only PDF and DOCX formats are allowed")
		return
	}

	// Save file
	userIDInt := userID.(uint)
	fileName := fmt.Sprintf("%d_%s", userIDInt, header.Filename)
	filePath := filepath.Join("uploads/resumes", fileName)
	if err := c.SaveUploadedFile(header, filePath); err != nil {
		log.Printf("Error saving file for user %d: %v", userIDInt, err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to save resume")
		return
	}

	log.Printf("Resume uploaded for user %d: %s", userIDInt, filePath)

	// Parse resume
	extractedData, err := ac.parseResume(filePath)
	if err != nil {
		log.Printf("Error parsing resume for user %d: %v", userIDInt, err)
		utils.RespondWithError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to parse resume: %v", err))
		return
	}

	if extractedData == nil {
		log.Printf("No data extracted from resume for user %d", userIDInt)
		utils.RespondWithError(c, http.StatusInternalServerError, "No data extracted from resume")
		return
	}

	// Update or create profile
	profile := models.Profile{
		UserID:         userIDInt,
		ResumeFilePath: filePath,
		Skills:         extractedData.Skills,
		Education:      extractedData.Education,
		Experience:     extractedData.Experience,
		Name:           extractedData.Name,
		Email:          extractedData.Email,
		Phone:          extractedData.Phone,
	}

	if err := ac.DB.Save(&profile).Error; err != nil {
		log.Printf("Error saving profile for user %d: %v", userIDInt, err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to save profile")
		return
	}

	log.Printf("Profile updated for user %d", userIDInt)
	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Resume uploaded and processed successfully"})
}

// parseResume extracts text from the resume and sends it to Gemini API for parsing.
func (ac *ApplicantController) parseResume(filePath string) (*ParsedData, error) {
	// Determine the file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	var resumeText string
	var err error

	// Extract text based on file type
	switch ext {
	case ".pdf":
		resumeText, err = extractTextFromPDF(filePath)
		if err != nil {
			log.Printf("Error extracting text from PDF %s: %v", filePath, err)
			return nil, fmt.Errorf("failed to extract text from PDF: %v", err)
		}
	case ".docx":
		resumeText, err = extractTextFromDOCX(filePath)
		if err != nil {
			log.Printf("Error extracting text from DOCX %s: %v", filePath, err)
			return nil, fmt.Errorf("failed to extract text from DOCX: %v", err)
		}
	default:
		log.Printf("Unsupported file extension: %s for file %s", ext, filePath)
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	if resumeText == "" {
		log.Printf("No text extracted from file %s", filePath)
		return nil, fmt.Errorf("no text extracted from file")
	}

	// Prepare the prompt for Gemini
	prompt := fmt.Sprintf(`
Extract the following information from the resume text below:

- Name
- Email
- Phone
- Education
- Experience
- Skills

Provide the information in JSON format with the following structure:

{
	"name": "",
	"email": "",
	"phone": "",
	"education": "",
	"experience": "",
	"skills": ""
}

Resume Text:
%s
`, resumeText)

	// Send the prompt to Gemini-1.5-pro API
	parsedData, err := ac.sendToGeminiAPI(prompt)
	if err != nil {
		log.Printf("Error parsing resume with Gemini: %v", err)
		return nil, err
	}

	return parsedData, nil
}

// sendToGeminiAPI sends the prompt to the Gemini API and parses the response.
func (ac *ApplicantController) sendToGeminiAPI(prompt string) (*ParsedData, error) {
	// Define the Gemini API endpoint
	apiURL := "https://api.gemini.com/v1/models/gemini-1.5-pro/completions" // Replace with actual endpoint

	// Create the request payload
	payload := map[string]interface{}{
		"prompt":            prompt,
		"max_tokens":        500,
		"temperature":       0.3,
		"top_p":             1.0,
		"frequency_penalty": 0.0,
		"presence_penalty":  0.0,
		// Add other parameters as required by Gemini API
	}

	// Marshal the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling payload: %v", err)
		return nil, err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ac.Cfg.GeminiAPIKey)) // Ensure GeminiAPIKey is set in config

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request to Gemini API: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Gemini API response: %v", err)
		return nil, err
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		log.Printf("Gemini API returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("Gemini API error: %s", string(body))
	}

	// Parse the response (assuming Gemini returns JSON with the completion)
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		log.Printf("Error unmarshalling Gemini API response: %v", err)
		return nil, err
	}

	// Extract the completion text
	// The structure depends on Gemini's API response format
	// Adjust the following lines based on actual response
	choices, ok := apiResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Printf("Invalid Gemini API response format")
		return nil, fmt.Errorf("invalid Gemini API response format")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		log.Printf("Invalid Gemini API response choice format")
		return nil, fmt.Errorf("invalid Gemini API response choice format")
	}

	text, ok := choice["text"].(string)
	if !ok {
		log.Printf("Gemini API response does not contain text")
		return nil, fmt.Errorf("Gemini API response does not contain text")
	}

	// Unmarshal the text into ParsedData
	var parsedData ParsedData
	if err := json.Unmarshal([]byte(text), &parsedData); err != nil {
		log.Printf("Error unmarshalling ParsedData: %v", err)
		return nil, err
	}

	return &parsedData, nil
}

// extractTextFromPDF extracts plain text from a PDF file using the ledongthuc/pdf package.
func extractTextFromPDF(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening PDF file: %v", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return "", fmt.Errorf("error creating PDF reader: %v", err)
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("error getting number of pages: %v", err)
	}

	var textBuilder strings.Builder
	for i := 0; i < numPages; i++ {
		page, err := pdfReader.GetPage(i + 1)
		if err != nil {
			return "", fmt.Errorf("error getting page %d: %v", i+1, err)
		}

		ex, err := extractor.New(page)
		if err != nil {
			return "", fmt.Errorf("error creating extractor for page %d: %v", i+1, err)
		}

		text, err := ex.ExtractText()
		if err != nil {
			return "", fmt.Errorf("error extracting text from page %d: %v", i+1, err)
		}

		textBuilder.WriteString(text)
	}

	return textBuilder.String(), nil
}

// extractTextFromDOCX extracts plain text from a DOCX file using the gooxml package.
func extractTextFromDOCX(filePath string) (string, error) {
	doc, err := document.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening DOCX file: %v", err)
	}
	defer doc.Close()

	var buf bytes.Buffer
	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			buf.WriteString(run.Text())
			buf.WriteString(" ")
		}
		buf.WriteString("\n")
	}
	return buf.String(), nil
}

// joinStrings joins a slice of strings with the specified separator.
func joinStrings(items []string, separator string) string {
	return strings.Join(items, separator)
}
