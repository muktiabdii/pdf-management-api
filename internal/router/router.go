package router

import (
	"github.com/gin-gonic/gin"
	"github.com/muktiabdii/pdf-management-api/internal/controller"
	"github.com/muktiabdii/pdf-management-api/internal/repository"
	"github.com/muktiabdii/pdf-management-api/internal/service"
	"gorm.io/gorm"
)

// Setup initializes and configures the HTTP router with all routes and middleware.
// It sets up dependency injection by creating repository, service, and controller instances,
// then registers all PDF-related API endpoints.
func Setup(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 10 << 20

	// Set up dependency injection: repository -> service -> controller
	pdfRepo := repository.NewPdfRepository(db)
	pdfService := service.NewPdfService(pdfRepo)
	pdfController := controller.NewPdfController(pdfService)

	// Define API routes
	api := r.Group("/api")
	{
		pdf := api.Group("/pdf")
		{
			// POST /api/pdf/generate - Generate a new PDF document
			pdf.POST("/generate", pdfController.GeneratePdf)
			// POST /api/pdf/upload - Upload an existing PDF file
			pdf.POST("/upload", pdfController.UploadPdf)
			// GET /api/pdf/list - Get paginated list of PDFs with optional status filter
			pdf.GET("/list", pdfController.ListPdf)
			// DELETE /api/pdf/:id - Soft-delete a PDF by ID
			pdf.DELETE("/:id", pdfController.DeletePdf)
		}
	}

	return r
}
