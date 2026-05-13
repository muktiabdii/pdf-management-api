package model

import (
	"time"
)

type PdfStatus string

const (
	StatusCreated  PdfStatus = "CREATED"
	StatusUploaded PdfStatus = "UPLOADED"
	StatusDeleted  PdfStatus = "DELETED"
)

// entity — representasi tabel pdf_files di database
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

func (PdfFile) TableName() string {
	return "pdf_files"
}

// DTO request

type GeneratePdfRequest struct {
	Title           string `json:"title"            binding:"required"`
	InstitutionName string `json:"institution_name" binding:"required"`
	Address         string `json:"address"          binding:"required"`
	Phone           string `json:"phone"            binding:"required"`
	LogoURL         string `json:"logo_url"`        // optional
	Content         string `json:"content"          binding:"required"`
}

type ListPdfRequest struct {
	Status string `form:"status"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

// DTO response

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

type BaseResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Data      any    `json:"data,omitempty"`
}

// helper — convert entity ke response DTO

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