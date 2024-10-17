package controllers

import (
	"net/http"

	"github.com/GolangAssignment/internal/models"
	"github.com/GolangAssignment/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminController struct {
	DB *gorm.DB
}

func NewAdminController(db *gorm.DB) *AdminController {
	return &AdminController{DB: db}
}

type CreateJobInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	CompanyName string `json:"company_name" binding:"required"`
}

func (ac *AdminController) CreateJob(c *gin.Context) {
	var input CreateJobInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User ID not found")
		return
	}

	job := models.Job{
		Title:       input.Title,
		Description: input.Description,
		CompanyName: input.CompanyName,
		PostedByID:  userID.(uint),
	}

	if err := ac.DB.Create(&job).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create job")
		return
	}

	utils.RespondWithSuccess(c, http.StatusCreated, gin.H{"message": "Job created successfully", "job_id": job.ID})
}

func (ac *AdminController) GetJob(c *gin.Context) {
	jobID := c.Param("job_id")
	var job models.Job
	if err := ac.DB.Preload("Applications").Preload("Applications.Applicant").First(&job, jobID).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Job not found")
		return
	}

	// Fetch applicants
	var applicants []models.User
	for _, application := range job.Applications {
		var applicant models.User
		if err := ac.DB.First(&applicant, application.ApplicantID).Error; err == nil {
			applicants = append(applicants, applicant)
		}
	}

	utils.RespondWithSuccess(c, http.StatusOK, gin.H{
		"job":        job,
		"applicants": applicants,
	})
}

func (ac *AdminController) GetAllApplicants(c *gin.Context) {
	var applicants []models.User
	if err := ac.DB.Where("user_type = ?", "Applicant").Find(&applicants).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch applicants")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"applicants": applicants})
}

func (ac *AdminController) GetApplicantData(c *gin.Context) {
	applicantID := c.Param("applicant_id")
	var profile models.Profile
	if err := ac.DB.Where("user_id = ?", applicantID).First(&profile).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Applicant profile not found")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"profile": profile})
}
