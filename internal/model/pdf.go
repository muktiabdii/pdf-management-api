package model

import (
	"time"
)

// PdfStatus represents the status of a PDF file in the system.
type PdfStatus string

// Status constants for PDF files.
const (
	// StatusCreated indicates a PDF has been generated but not yet finalized.
	StatusCreated PdfStatus = "CREATED"
	// StatusUploaded indicates a PDF has been successfully uploaded to storage.
	StatusUploaded PdfStatus = "UPLOADED"
	// StatusDeleted indicates a PDF has been soft-deleted from the system.
	StatusDeleted PdfStatus = "DELETED"
)

// PdfFile represents the database entity for storing PDF file information.
// It maps to the "pdf_files" table in PostgreSQL.
type PdfFile struct {
	ID           uint       `gorm:"primaryKey;autoIncrement"          json:"id"`
	Filename     string     `gorm:"type:varchar(255);not null"        json:"filename"`
	OriginalName *string    `gorm:"type:varchar(255)"                 json:"original_name"`
	Filepath     string     `gorm:"type:varchar(500);not null"        json:"filepath"`
	Size         *int64     `gorm:"type:bigint"                       json:"size"`
	Status       PdfStatus  `gorm:"type:varchar(20);not null"         json:"status"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"                    json:"created_at"`
	UpdatedAt    *time.Time `gorm:"autoUpdateTime"                    json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index"                             json:"deleted_at,omitempty"`
}

// TableName specifies the database table name for the PdfFile entity.
func (PdfFile) TableName() string {
	return "pdf_files"
}

// GeneratePdfRequest is the request payload for generating a new PDF document.
type GeneratePdfRequest struct {
	// Title is the title of the PDF document (required)
	Title string `json:"title" binding:"required"`
	// InstitutionName is the name of the issuing institution (required)
	InstitutionName string `json:"institution_name" binding:"required"`
	// Address is the institution's address (required)
	Address string `json:"address" binding:"required"`
	// Phone is the institution's phone number (required)
	Phone string `json:"phone" binding:"required"`
	// LogoURL is an optional URL to the institution's logo image
	LogoURL string `json:"logo_url"`
	// Content is the main content/body of the PDF document (required)
	Content string `json:"content" binding:"required"`
}

// ListPdfRequest is the query parameters for fetching a list of PDF files.
type ListPdfRequest struct {
	// Status filters PDFs by their status (CREATED, UPLOADED, or DELETED)
	Status string `form:"status"`
	// Page is the page number for pagination (default: 1)
	Page int `form:"page"`
	// Limit is the number of records per page (default: 10)
	Limit int `form:"limit"`
}

// PdfResponse is the response payload containing PDF file information.
type PdfResponse struct {
	ID           uint       `json:"id"`
	Filename     string     `json:"filename"`
	OriginalName *string    `json:"original_name"`
	Filepath     string     `json:"filepath"`
	Size         *int64     `json:"size"`
	Status       PdfStatus  `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type PaginationMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type ListPdfResponse struct {
	Data       []PdfResponse  `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// BaseResponse is the standard response format for all API endpoints.
type BaseResponse struct {
	// Success indicates whether the request was successful
	Success bool `json:"success"`
	// Message contains a human-readable message about the result
	Message string `json:"message,omitempty"`
	// ErrorCode contains a machine-readable error code for programmatic handling
	ErrorCode string `json:"error_code,omitempty"`
	// Data contains the response payload (if any)
	Data any `json:"data,omitempty"`
}

// ToPdfResponse converts a PdfFile entity to a PdfResponse DTO.
func ToPdfResponse(p PdfFile) PdfResponse {
	return PdfResponse{
		ID:           p.ID,
		Filename:     p.Filename,
		OriginalName: p.OriginalName,
		Filepath:     p.Filepath,
		Size:         p.Size,
		Status:       p.Status,
		CreatedAt:    p.CreatedAt,
		DeletedAt:    p.DeletedAt,
	}
}
