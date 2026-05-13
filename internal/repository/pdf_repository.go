package repository

import (
	"context"
	"errors"
	"time"

	"github.com/muktiabdii/pdf-management-api/internal/model"
	"gorm.io/gorm"
)

// Error definitions for repository operations.
var (
	// ErrNotFound is returned when a PDF file is not found in the database.
	ErrNotFound = errors.New("pdf file not found")
	// ErrAlreadyDeleted is returned when attempting to delete an already deleted PDF file.
	ErrAlreadyDeleted = errors.New("pdf file already deleted")
)

// PdfRepository defines the interface for PDF file data access operations.
type PdfRepository interface {
	// Save inserts a new PDF file record into the database.
	Save(ctx context.Context, pdf *model.PdfFile) error
	// FindAll retrieves a paginated list of PDF files with optional status filtering.
	FindAll(ctx context.Context, req model.ListPdfRequest) ([]model.PdfFile, int64, error)
	// FindByID retrieves a single PDF file record by its ID.
	FindByID(ctx context.Context, id uint) (*model.PdfFile, error)
	// SoftDelete marks a PDF file as deleted without removing the database record.
	SoftDelete(ctx context.Context, id uint) (*model.PdfFile, error)
}

// pdfRepository implements the PdfRepository interface using GORM.
type pdfRepository struct {
	db *gorm.DB
}

// NewPdfRepository creates a new instance of pdfRepository.
func NewPdfRepository(db *gorm.DB) PdfRepository {
	return &pdfRepository{db: db}
}

// Save inserts a new PDF file record into the database.
func (r *pdfRepository) Save(ctx context.Context, pdf *model.PdfFile) error {
	return r.db.WithContext(ctx).Create(pdf).Error
}

// FindAll retrieves all PDF files from the database with optional status filtering and pagination.
// It returns the list of PDFs, total count, and any error encountered.
func (r *pdfRepository) FindAll(ctx context.Context, req model.ListPdfRequest) ([]model.PdfFile, int64, error) {
	var pdfs []model.PdfFile
	var total int64

	query := r.db.WithContext(ctx).Model(&model.PdfFile{})

	// Apply status filter if provided
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Get total count before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	offset := (req.Page - 1) * req.Limit

	err := query.
		Order("created_at DESC").
		Limit(req.Limit).
		Offset(offset).
		Find(&pdfs).Error

	return pdfs, total, err
}

// FindByID retrieves a single PDF file record by its ID.
// Returns ErrNotFound if the record does not exist.
func (r *pdfRepository) FindByID(ctx context.Context, id uint) (*model.PdfFile, error) {
	var pdf model.PdfFile

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&pdf).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	return &pdf, err
}

// SoftDelete marks a PDF file as deleted without removing the database record.
// It updates the status to DELETED and sets the DeletedAt timestamp.
// Returns ErrNotFound if the record does not exist or ErrAlreadyDeleted if already deleted.
func (r *pdfRepository) SoftDelete(ctx context.Context, id uint) (*model.PdfFile, error) {
	pdf, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err // ErrNotFound
	}

	if pdf.Status == model.StatusDeleted {
		return nil, ErrAlreadyDeleted
	}

	now := time.Now()
	pdf.Status = model.StatusDeleted
	pdf.DeletedAt = &now

	err = r.db.WithContext(ctx).
		Model(pdf).
		Updates(map[string]any{
			"status":     model.StatusDeleted,
			"deleted_at": now,
		}).Error

	return pdf, err
}
