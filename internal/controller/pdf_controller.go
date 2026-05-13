package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/muktiabdii/pdf-management-api/internal/model"
	"github.com/muktiabdii/pdf-management-api/internal/repository"
	"github.com/muktiabdii/pdf-management-api/internal/service"
)

type PdfController struct {
	service service.PdfService
}

func NewPdfController(service service.PdfService) *PdfController {
	return &PdfController{service: service}
}

// GeneratePdf — POST /api/pdf/generate
func (c *PdfController) GeneratePdf(ctx *gin.Context) {
	var req model.GeneratePdfRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, model.BaseResponse{
			Success: false,
			Message: "invalid request body: " + err.Error(),
		})
		return
	}

	result, err := c.service.GeneratePdf(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.BaseResponse{
			Success: false,
			Message: "failed to generate pdf: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, model.BaseResponse{
		Success: true,
		Message: "PDF generated successfully",
		Data:    result,
	})
}

// UploadPdf — POST /api/pdf/upload
func (c *PdfController) UploadPdf(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.BaseResponse{
			Success: false,
			Message: "file is required",
		})
		return
	}

	result, err := c.service.UploadPdf(ctx.Request.Context(), fileHeader)
	if err != nil {
		switch err.Error() {
		case "FILE_TOO_LARGE":
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success":    false,
				"message":    "File size exceeds maximum limit (10MB)",
				"error_code": "FILE_TOO_LARGE",
			})
		case "INVALID_EXTENSION":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"message":    "Only .pdf files are allowed",
				"error_code": "INVALID_EXTENSION",
			})
		case "INVALID_MIME_TYPE":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"message":    "File must be a valid PDF (application/pdf)",
				"error_code": "INVALID_MIME_TYPE",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, model.BaseResponse{
				Success: false,
				Message: "failed to upload pdf: " + err.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusCreated, model.BaseResponse{
		Success: true,
		Message: "PDF uploaded successfully",
		Data:    result,
	})
}

// ListPdf — GET /api/pdf/list
func (c *PdfController) ListPdf(ctx *gin.Context) {
	var req model.ListPdfRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, model.BaseResponse{
			Success: false,
			Message: "invalid query parameters: " + err.Error(),
		})
		return
	}

	// validasi nilai status jika diisi
	if req.Status != "" {
		validStatus := map[string]bool{
			"CREATED":  true,
			"UPLOADED": true,
			"DELETED":  true,
		}
		if !validStatus[req.Status] {
			ctx.JSON(http.StatusBadRequest, model.BaseResponse{
				Success: false,
				Message: "invalid status value, must be CREATED, UPLOADED, or DELETED",
			})
			return
		}
	}

	result, err := c.service.ListPdf(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.BaseResponse{
			Success: false,
			Message: "failed to fetch pdf list: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// DeletePdf — DELETE /api/pdf/:id
func (c *PdfController) DeletePdf(ctx *gin.Context) {
	// parse id dari URL param
	idStr := ctx.Param("id")
	idInt, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.BaseResponse{
			Success: false,
			Message: "invalid id format",
		})
		return
	}

	result, err := c.service.DeletePdf(ctx.Request.Context(), uint(idInt))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			ctx.JSON(http.StatusNotFound, model.BaseResponse{
				Success: false,
				Message: "pdf file not found",
			})
		case errors.Is(err, repository.ErrAlreadyDeleted):
			ctx.JSON(http.StatusConflict, model.BaseResponse{
				Success: false,
				Message: "pdf file already deleted",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, model.BaseResponse{
				Success: false,
				Message: "failed to delete pdf: " + err.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, model.BaseResponse{
		Success: true,
		Message: "PDF deleted successfully",
		Data:    result,
	})
}