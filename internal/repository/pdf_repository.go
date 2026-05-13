package repository

import (
	"context"
	"errors"
	"time"

	"github.com/muktiabdii/pdf-management-api/internal/model"
	"gorm.io/gorm"
)

var (
	ErrNotFound        = errors.New("pdf file not found")
	ErrAlreadyDeleted  = errors.New("pdf file already deleted")
)

type PdfRepository interface {
	Save(ctx context.Context, pdf *model.PdfFile) error
	FindAll(ctx context.Context, req model.ListPdfRequest) ([]model.PdfFile, int64, error)
	FindByID(ctx context.Context, id uint) (*model.PdfFile, error)
	SoftDelete(ctx context.Context, id uint) (*model.PdfFile, error)
}

type pdfRepository struct {
	db *gorm.DB
}

func NewPdfRepository(db *gorm.DB) PdfRepository {
	return &pdfRepository{db: db}
}

// Save — insert record baru ke tabel pdf_files
func (r *pdfRepository) Save(ctx context.Context, pdf *model.PdfFile) error {
	return r.db.WithContext(ctx).Create(pdf).Error
}

// FindAll — ambil semua record dengan filter status & pagination
func (r *pdfRepository) FindAll(ctx context.Context, req model.ListPdfRequest) ([]model.PdfFile, int64, error) {
	var pdfs  []model.PdfFile
	var total int64

	query := r.db.WithContext(ctx).Model(&model.PdfFile{})

	// filter by status jika ada
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// hitung total sebelum pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// default pagination
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

// FindByID — cari satu record berdasarkan id
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

// SoftDelete — update status jadi DELETED dan set deleted_at
func (r *pdfRepository) SoftDelete(ctx context.Context, id uint) (*model.PdfFile, error) {
	pdf, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err // ErrNotFound
	}

	if pdf.Status == model.StatusDeleted {
		return nil, ErrAlreadyDeleted
	}

	now := time.Now()
	pdf.Status    = model.StatusDeleted
	pdf.DeletedAt = &now

	err = r.db.WithContext(ctx).
		Model(pdf).
		Updates(map[string]any{
			"status":     model.StatusDeleted,
			"deleted_at": now,
		}).Error

	return pdf, err
}