package controllers

import (
	"net/http"

	"github.com/GolangAssignment/internal/models"
	"github.com/GolangAssignment/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobController struct {
	DB *gorm.DB
}

func NewJobController(db *gorm.DB) *JobController {
	return &JobController{DB: db}
}

func (jc *JobController) GetJobs(c *gin.Context) {
	var jobs []models.Job
	if err := jc.DB.Find(&jobs).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch jobs")
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"jobs": jobs})
}

func (jc *JobController) ApplyJob(c *gin.Context) {
	jobID := c.Query("job_id")
	if jobID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "Job ID is required")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User ID not found")
		return
	}

	// Check if job exists
	var job models.Job
	if err := jc.DB.First(&job, jobID).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Job not found")
		return
	}

	// Check if already applied
	var application models.Application
	if err := jc.DB.Where("job_id = ? AND applicant_id = ?", job.ID, userID.(uint)).First(&application).Error; err == nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Already applied to this job")
		return
	}

	// Create application
	application = models.Application{
		JobID:       job.ID,
		ApplicantID: userID.(uint),
	}

	if err := jc.DB.Create(&application).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to apply for job")
		return
	}

	// Increment total applications
	jc.DB.Model(&job).Update("total_applications", job.TotalApplications+1)

	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Applied to job successfully"})
}
