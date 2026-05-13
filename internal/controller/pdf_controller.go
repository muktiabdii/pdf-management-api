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
		// Perbaikan: Gunakan BaseResponse statt gin.H
		resp := model.BaseResponse{Success: false}
		status := http.StatusInternalServerError

		switch err.Error() {
		case "FILE_TOO_LARGE":
			status = http.StatusRequestEntityTooLarge
			resp.ErrorCode = "FILE_TOO_LARGE"
			resp.Message = "File size exceeds maximum limit (10MB)"
		case "INVALID_EXTENSION":
			status = http.StatusBadRequest
			resp.ErrorCode = "INVALID_EXTENSION"
			resp.Message = "Only .pdf files are allowed"
		case "INVALID_MIME_TYPE":
			status = http.StatusBadRequest
			resp.ErrorCode = "INVALID_MIME_TYPE"
			resp.Message = "File must be a valid PDF (application/pdf)"
		default:
			resp.Message = "failed to upload pdf: " + err.Error()
		}
		ctx.JSON(status, resp)
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

	// Perbaikan: Ganti gin.H dengan BaseResponse agar konsisten
	ctx.JSON(http.StatusOK, model.BaseResponse{
		Success: true,
		Message: "PDF list fetched successfully",
		Data:    result, // result di sini biasanya berisi {data: [], pagination: {}}
	})
}

// DeletePdf — DELETE /api/pdf/:id
func (c *PdfController) DeletePdf(ctx *gin.Context) {
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
		resp := model.BaseResponse{Success: false}
		status := http.StatusInternalServerError

		switch {
		case errors.Is(err, repository.ErrNotFound):
			status = http.StatusNotFound
			resp.Message = "pdf file not found"
			resp.ErrorCode = "NOT_FOUND"
		case errors.Is(err, repository.ErrAlreadyDeleted):
			status = http.StatusConflict
			resp.Message = "pdf file already deleted"
			resp.ErrorCode = "ALREADY_DELETED"
		default:
			resp.Message = "failed to delete pdf: " + err.Error()
		}
		ctx.JSON(status, resp)
		return
	}

	ctx.JSON(http.StatusOK, model.BaseResponse{
		Success: true,
		Message: "PDF deleted successfully",
		Data:    result,
	})
}