package service

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"time"
	"io"
	"net/http"

	"github.com/google/uuid"
	pdfgen "github.com/muktiabdii/pdf-management-api/pkg/pdf"
	"github.com/muktiabdii/pdf-management-api/internal/model"
	"github.com/muktiabdii/pdf-management-api/internal/repository"
	"github.com/muktiabdii/pdf-management-api/pkg/storage"
)

const maxUploadSize = 10 << 20 // 10 MB

type PdfService interface {
	GeneratePdf(ctx context.Context, req model.GeneratePdfRequest) (*model.PdfResponse, error)
	UploadPdf(ctx context.Context, fileHeader *multipart.FileHeader) (*model.PdfResponse, error)
	ListPdf(ctx context.Context, req model.ListPdfRequest) (*model.ListPdfResponse, error)
	DeletePdf(ctx context.Context, id uint) (*model.PdfResponse, error)
}

type pdfService struct {
	repo repository.PdfRepository
}

func NewPdfService(repo repository.PdfRepository) PdfService {
	return &pdfService{repo: repo}
}

// GeneratePdf — generate PDF dari request, upload ke S3, simpan ke DB
func (s *pdfService) GeneratePdf(ctx context.Context, req model.GeneratePdfRequest) (*model.PdfResponse, error) {
	now := time.Now()

	// 1. generate PDF
	reportData := pdfgen.ReportData{
		Title:           req.Title,
		InstitutionName: req.InstitutionName,
		Address:         req.Address,
		Phone:           req.Phone,
		LogoURL:         req.LogoURL,
		Content:         req.Content,
		GeneratedAt:     now,
	}

	generatedPdf, err := pdfgen.GenerateReport(reportData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	// 2. tulis PDF ke buffer
	var buf bytes.Buffer
	if err := generatedPdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to write pdf to buffer: %w", err)
	}

	// 3. buat nama file unik
	filename := fmt.Sprintf("report_%s_%s.pdf",
		now.Format("20060102"),
		uuid.New().String()[:8],
	)

	// 4. upload ke S3
	fileReader := bytes.NewReader(buf.Bytes())
	filepath, err := storage.UploadFileFromReader(ctx, filename, fileReader, "application/pdf", int64(buf.Len()))
	if err != nil {
		return nil, fmt.Errorf("failed to upload pdf to S3: %w", err)
	}

	// 5. simpan record ke DB
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

// UploadPdf — validasi & upload file dari user ke S3, simpan ke DB
func (s *pdfService) UploadPdf(ctx context.Context, fileHeader *multipart.FileHeader) (*model.PdfResponse, error) {
	// 1. validasi ukuran
	if fileHeader.Size > maxUploadSize {
		return nil, fmt.Errorf("FILE_TOO_LARGE")
	}

	// 2. validasi ekstensi
	originalName := fileHeader.Filename
	if len(originalName) < 4 || originalName[len(originalName)-4:] != ".pdf" {
		return nil, fmt.Errorf("INVALID_EXTENSION")
	}

	// 3. buka file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// 4. validasi MIME type (baca 512 byte pertama)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	mimeType := http.DetectContentType(buf[:n])
	if mimeType != "application/pdf" && mimeType != "application/octet-stream" {
		return nil, fmt.Errorf("INVALID_MIME_TYPE")
	}

	// 5. reset pointer setelah baca 512 byte
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// 6. buat nama file unik
	now := time.Now()
	filename := fmt.Sprintf("upload_%s_%s.pdf",
		now.Format("20060102"),
		uuid.New().String()[:8],
	)

	// 7. upload ke S3
	filepath, err := storage.UploadFile(ctx, filename, file, "application/pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to upload pdf to S3: %w", err)
	}

	// 8. simpan record ke DB
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

// ListPdf — ambil daftar PDF dengan filter & pagination
func (s *pdfService) ListPdf(ctx context.Context, req model.ListPdfRequest) (*model.ListPdfResponse, error) {
	pdfs, total, err := s.repo.FindAll(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pdf list: %w", err)
	}

	// default pagination value
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

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

// DeletePdf — soft delete PDF berdasarkan id
func (s *pdfService) DeletePdf(ctx context.Context, id uint) (*model.PdfResponse, error) {
	pdf, err := s.repo.SoftDelete(ctx, id)
	if err != nil {
		return nil, err // ErrNotFound / ErrAlreadyDeleted langsung di-bubble up
	}

	result := model.ToPdfResponse(*pdf)
	return &result, nil
}