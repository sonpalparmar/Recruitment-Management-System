package routes

import (
	"github.com/GolangAssignment/internal/config"
	"github.com/GolangAssignment/internal/controllers"
	"github.com/GolangAssignment/internal/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg config.Config) {
	// Initialize controllers with dependencies
	authController := controllers.NewAuthController(db, cfg)
	adminController := controllers.NewAdminController(db)
	jobController := controllers.NewJobController(db)
	applicantController := controllers.NewApplicantController(db, cfg)

	// Public routes
	router.POST("/signup", authController.SignUp)
	router.POST("/login", authController.Login)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware(cfg.JWTSecret))

	// Applicant-specific routes
	protected.POST("/uploadResume", middlewares.RoleMiddleware("Applicant"), applicantController.UploadResume)
	protected.GET("/jobs", jobController.GetJobs)
	protected.GET("/jobs/apply", middlewares.RoleMiddleware("Applicant"), jobController.ApplyJob)

	// Admin-specific routes
	admin := protected.Group("/admin")
	admin.Use(middlewares.RoleMiddleware("Admin"))
	{
		admin.POST("/job", adminController.CreateJob)
		admin.GET("/job/:job_id", adminController.GetJob)
		admin.GET("/applicants", adminController.GetAllApplicants)
		admin.GET("/applicant/:applicant_id", adminController.GetApplicantData)
	}
}
