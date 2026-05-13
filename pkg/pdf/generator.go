package pdf

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type ReportData struct {
	Title           string
	InstitutionName string
	Address         string
	Phone           string
	LogoURL         string
	Content         string
	GeneratedAt     time.Time
}

func GenerateReport(data ReportData) (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// ── HEADER ────────────────────────────────────────────────────────────────

	// logo (opsional)
	if data.LogoURL != "" {
		if err := loadImageFromURL(pdf, data.LogoURL); err == nil {
			pdf.ImageOptions("logo.png", 15, 10, 25, 25, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		}
	}

	// nama institusi — tengah
	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(0, 12)
	pdf.CellFormat(210, 8, data.InstitutionName, "", 1, "C", false, 0, "")

	// alamat
	pdf.SetFont("Arial", "", 10)
	pdf.SetX(0)
	pdf.CellFormat(210, 5, data.Address, "", 1, "C", false, 0, "")

	// telepon
	pdf.SetX(0)
	pdf.CellFormat(210, 5, fmt.Sprintf("Telp: %s", data.Phone), "", 1, "C", false, 0, "")

	// garis pemisah header
	pdf.SetLineWidth(0.5)
	pdf.Line(15, 38, 195, 38)
	pdf.Ln(8)

	// ── CONTENT ───────────────────────────────────────────────────────────────

	// judul dokumen
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, data.Title, "", 1, "C", false, 0, "")

	// tanggal generate
	pdf.SetFont("Arial", "I", 10)
	pdf.CellFormat(0, 6,
		fmt.Sprintf("Tanggal: %s", data.GeneratedAt.Format("02 January 2006")),
		"", 1, "C", false, 0, "")
	pdf.Ln(6)

	// isi konten
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 7, data.Content, "", "L", false)

	// ── FOOTER ────────────────────────────────────────────────────────────────
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)

		// nomor halaman
		pdf.CellFormat(0, 5,
			fmt.Sprintf("Page %d of {nb}", pdf.PageNo()),
			"", 0, "L", false, 0, "")

		// timestamp generate
		pdf.CellFormat(0, 5,
			fmt.Sprintf("Generated at: %s", data.GeneratedAt.Format("02/01/2006 15:04:05")),
			"", 0, "R", false, 0, "")
	})
	pdf.AliasNbPages("{nb}")

	return pdf, nil
}

// loadImageFromURL — download logo dari URL lalu register ke gofpdf
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

	// register image ke gofpdf dari memory
	pdf.RegisterImageOptionsReader(
		"logo.png",
		gofpdf.ImageOptions{ImageType: "PNG"},
		newBytesReader(imageData),
	)

	return nil
}

// newBytesReader — wrapper agar []byte bisa dipakai sebagai io.Reader
type bytesReader struct {
	data []byte
	pos  int
}

func newBytesReader(data []byte) io.Reader {
	return &bytesReader{data: data}
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}