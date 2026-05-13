package pdf

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// ReportData contains all the information needed to generate a PDF report.
type ReportData struct {
	// Title is the main title of the report
	Title string
	// InstitutionName is the name of the institution issuing the report
	InstitutionName string
	// Address is the institution's address
	Address string
	// Phone is the institution's phone number
	Phone string
	// LogoURL is an optional URL to the institution's logo
	LogoURL string
	// Content is the main body/content of the report
	Content string
	// GeneratedAt is the timestamp when the report was generated
	GeneratedAt time.Time
}

// GenerateReport creates a PDF document from the provided report data.
// The PDF includes header (with optional logo), content, and footer with page numbers.
func GenerateReport(data ReportData) (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// PDF Header Section

	// Load and display logo if provided
	if data.LogoURL != "" {
		if err := loadImageFromURL(pdf, data.LogoURL); err == nil {
			pdf.ImageOptions("logo.png", 15, 10, 25, 25, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		}
	}

	// Display institution name (centered, bold, large)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(0, 12)
	pdf.CellFormat(210, 8, data.InstitutionName, "", 1, "C", false, 0, "")

	// Display address (centered)
	pdf.SetFont("Arial", "", 10)
	pdf.SetX(0)
	pdf.CellFormat(210, 5, data.Address, "", 1, "C", false, 0, "")

	// Display phone number (centered)
	pdf.SetX(0)
	pdf.CellFormat(210, 5, fmt.Sprintf("Telp: %s", data.Phone), "", 1, "C", false, 0, "")

	// Draw separator line
	pdf.SetLineWidth(0.5)
	pdf.Line(15, 38, 195, 38)
	pdf.Ln(8)

	// PDF Content Section

	// Display report title (centered, bold)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, data.Title, "", 1, "C", false, 0, "")

	// Display generation date (centered, italic)
	pdf.SetFont("Arial", "I", 10)
	pdf.CellFormat(0, 6,
		fmt.Sprintf("Date: %s", data.GeneratedAt.Format("02 January 2006")),
		"", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Display main content
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 7, data.Content, "", "L", false)

	// PDF Footer Section
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)

		// Display page number on the left
		pdf.CellFormat(0, 5,
			fmt.Sprintf("Page %d of {nb}", pdf.PageNo()),
			"", 0, "L", false, 0, "")

		// Display generation timestamp on the right
		pdf.CellFormat(0, 5,
			fmt.Sprintf("Generated at: %s", data.GeneratedAt.Format("02/01/2006 15:04:05")),
			"", 0, "R", false, 0, "")
	})
	pdf.AliasNbPages("{nb}")

	return pdf, nil
}

// loadImageFromURL downloads an image from the provided URL and registers it with the PDF.
// Returns an error if the URL is invalid, the request fails, or the image cannot be processed.
func loadImageFromURL(pdf *gofpdf.Fpdf, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch logo: status %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Register image with gofpdf from memory buffer
	pdf.RegisterImageOptionsReader(
		"logo.png",
		gofpdf.ImageOptions{ImageType: "PNG"},
		newBytesReader(imageData),
	)

	return nil
}

// bytesReader is an adapter that allows a byte slice to be used as an io.Reader.
type bytesReader struct {
	// data is the byte slice to be read
	data []byte
	// pos is the current read position in the byte slice
	pos int
}

// newBytesReader creates a new io.Reader from a byte slice.
func newBytesReader(data []byte) io.Reader {
	return &bytesReader{data: data}
}

// Read reads bytes from the bytesReader into the provided buffer.
// It implements the io.Reader interface.
func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
