package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/muktiabdii/pdf-management-api/internal/model"
	"github.com/muktiabdii/pdf-management-api/internal/repository"
	pdfgen "github.com/muktiabdii/pdf-management-api/pkg/pdf"
	"github.com/muktiabdii/pdf-management-api/pkg/storage"
)

// Maximum allowed file upload size: 10 MB.
const maxUploadSize = 10 << 20

// PdfService defines the interface for PDF business logic operations.
type PdfService interface {
	// GeneratePdf creates a new PDF document from the provided request data.
	GeneratePdf(ctx context.Context, req model.GeneratePdfRequest) (*model.PdfResponse, error)
	// UploadPdf handles validation and storage of user-uploaded PDF files.
	UploadPdf(ctx context.Context, fileHeader *multipart.FileHeader) (*model.PdfResponse, error)
	// ListPdf retrieves a paginated list of PDF files with optional status filtering.
	ListPdf(ctx context.Context, req model.ListPdfRequest) (*model.ListPdfResponse, error)
	// DeletePdf performs a soft-delete of a PDF file by its ID.
	DeletePdf(ctx context.Context, id uint) (*model.PdfResponse, error)
}

// pdfService implements the PdfService interface.
type pdfService struct {
	repo repository.PdfRepository
}

// NewPdfService creates a new instance of pdfService.
func NewPdfService(repo repository.PdfRepository) PdfService {
	return &pdfService{repo: repo}
}

// GeneratePdf generates a new PDF document from request data.
// It creates the PDF, uploads it to S3, and saves the metadata to the database.
func (s *pdfService) GeneratePdf(ctx context.Context, req model.GeneratePdfRequest) (*model.PdfResponse, error) {
	now := time.Now()

	// Create PDF data structure from request
	reportData := pdfgen.ReportData{
		Title:           req.Title,
		InstitutionName: req.InstitutionName,
		Address:         req.Address,
		Phone:           req.Phone,
		LogoURL:         req.LogoURL,
		Content:         req.Content,
		GeneratedAt:     now,
	}

	// Generate PDF from data
	generatedPdf, err := pdfgen.GenerateReport(reportData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	// Write PDF to buffer
	var buf bytes.Buffer
	if err := generatedPdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to write pdf to buffer: %w", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("report_%s_%s.pdf",
		now.Format("20060102"),
		uuid.New().String()[:8],
	)

	// Upload PDF to S3
	fileReader := bytes.NewReader(buf.Bytes())
	filepath, err := storage.UploadFileFromReader(ctx, filename, fileReader, "application/pdf", int64(buf.Len()))
	if err != nil {
		return nil, fmt.Errorf("failed to upload pdf to S3: %w", err)
	}

	// Save PDF metadata to database
	size := int64(buf.Len())
	record := &model.PdfFile{
		Filename: filename,
		Filepath: filepath,
		Size:     &size,
		Status:   model.StatusCreated,
	}

	if err := s.repo.Save(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to save pdf record: %w", err)
	}

	result := model.ToPdfResponse(*record)
	return &result, nil
}

// UploadPdf validates and uploads a user-provided PDF file.
// It checks file size, extension, and MIME type before uploading to S3 and saving to the database.
func (s *pdfService) UploadPdf(ctx context.Context, fileHeader *multipart.FileHeader) (*model.PdfResponse, error) {
	// Validate file size
	if fileHeader.Size > maxUploadSize {
		return nil, fmt.Errorf("FILE_TOO_LARGE")
	}

	// Validate file extension
	originalName := fileHeader.Filename
	if len(originalName) < 4 || originalName[len(originalName)-4:] != ".pdf" {
		return nil, fmt.Errorf("INVALID_EXTENSION")
	}

	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Validate MIME type by reading first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	mimeType := http.DetectContentType(buf[:n])
	if mimeType != "application/pdf" && mimeType != "application/octet-stream" {
		return nil, fmt.Errorf("INVALID_MIME_TYPE")
	}

	// Reset file pointer after reading MIME type
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Generate unique filename
	now := time.Now()
	filename := fmt.Sprintf("upload_%s_%s.pdf",
		now.Format("20060102"),
		uuid.New().String()[:8],
	)

	// Upload PDF to S3
	filepath, err := storage.UploadFile(ctx, filename, file, "application/pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to upload pdf to S3: %w", err)
	}

	// Save PDF metadata to database
	size := fileHeader.Size
	record := &model.PdfFile{
		Filename:     filename,
		OriginalName: &originalName,
		Filepath:     filepath,
		Size:         &size,
		Status:       model.StatusUploaded,
	}

	if err := s.repo.Save(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to save pdf record: %w", err)
	}

	result := model.ToPdfResponse(*record)
	return &result, nil
}

// ListPdf retrieves a paginated list of PDF files with optional status filtering.
func (s *pdfService) ListPdf(ctx context.Context, req model.ListPdfRequest) (*model.ListPdfResponse, error) {
	pdfs, total, err := s.repo.FindAll(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pdf list: %w", err)
	}

	// Apply default pagination values if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Convert entities to response DTOs
	responses := make([]model.PdfResponse, len(pdfs))
	for i, p := range pdfs {
		responses[i] = model.ToPdfResponse(p)
	}

	return &model.ListPdfResponse{
		Data: responses,
		Pagination: model.PaginationMeta{
			Page:  req.Page,
			Limit: req.Limit,
			Total: total,
		},
	}, nil
}

// DeletePdf performs a soft-delete of a PDF file by ID.
// This updates the file status to DELETED and sets the DeletedAt timestamp.
func (s *pdfService) DeletePdf(ctx context.Context, id uint) (*model.PdfResponse, error) {
	pdf, err := s.repo.SoftDelete(ctx, id)
	if err != nil {
		return nil, err // ErrNotFound or ErrAlreadyDeleted is bubbled up from repository
	}

	result := model.ToPdfResponse(*pdf)
	return &result, nil
}
