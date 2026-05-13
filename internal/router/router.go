package router

import (
	"github.com/gin-gonic/gin"
	"github.com/muktiabdii/pdf-management-api/internal/controller"
	"github.com/muktiabdii/pdf-management-api/internal/repository"
	"github.com/muktiabdii/pdf-management-api/internal/service"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 10 << 20

	// dependency injection
	pdfRepo       := repository.NewPdfRepository(db)
	pdfService    := service.NewPdfService(pdfRepo)
	pdfController := controller.NewPdfController(pdfService)

	// routes
	api := r.Group("/api")
	{
		pdf := api.Group("/pdf")
		{
			pdf.POST("/generate",  pdfController.GeneratePdf)
			pdf.POST("/upload",    pdfController.UploadPdf)
			pdf.GET("/list",       pdfController.ListPdf)
			pdf.DELETE("/:id",     pdfController.DeletePdf)
		}
	}

	return r
}