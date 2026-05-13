# PDF Management API

## 📋 Deskripsi Project

**PDF Management API** adalah REST API yang dibangun dengan Go dan Gin Framework untuk mengelola file PDF. Aplikasi ini menyediakan fitur untuk:

- ✅ Generate PDF dari template atau data
- ✅ Upload file PDF ke cloud storage (AWS S3)
- ✅ Menyimpan metadata PDF di database
- ✅ Melihat daftar semua file PDF yang telah diupload
- ✅ Menghapus file PDF dari sistem

Semua data disimpan secara aman di PostgreSQL dan file disimpan di AWS S3 dengan tracking status yang akurat.

---

## 🛠️ Tech Stack

| Komponen | Teknologi | Versi |
|----------|-----------|-------|
| **Backend Framework** | Gin Web Framework | v1.12.0 |
| **Programming Language** | Go | 1.25.0 |
| **Database** | PostgreSQL | 15 |
| **ORM** | GORM | - |
| **Cloud Storage** | AWS S3 | v1.101.0 |
| **Environment Management** | godotenv | - |
| **Containerization** | Docker & Docker Compose | - |

### Dependencies Utama:
- `github.com/gin-gonic/gin` - Web Framework
- `gorm.io/gorm` - Object Relational Mapping
- `gorm.io/driver/postgres` - PostgreSQL Driver
- `github.com/aws/aws-sdk-go-v2` - AWS SDK
- `github.com/joho/godotenv` - Environment Variables

---

## 📦 Struktur Folder Project

```
pdf-management-api/
│
├── cmd/
│   └── api/
│       └── main.go                 # Entry point aplikasi
│
├── internal/
│   ├── controller/
│   │   └── pdf_controller.go       # HTTP request handlers
│   │
│   ├── model/
│   │   └── pdf.go                  # Data structures & database entities
│   │
│   ├── repository/
│   │   └── pdf_repository.go       # Database layer (CRUD operations)
│   │
│   ├── router/
│   │   └── router.go               # Route definitions
│   │
│   └── service/
│       └── pdf_service.go          # Business logic layer
│
├── pkg/
│   ├── database/
│   │   └── postgres.go             # PostgreSQL connection & migrations
│   │
│   ├── pdf/
│   │   └── generator.go            # PDF generation logic
│   │
│   └── storage/
│       └── s3.go                   # AWS S3 integration
│
├── .env                            # Environment variables (git ignored)
├── .env.example                    # Environment variables template
├── docker-compose.yml              # Docker Compose configuration
├── Dockerfile                      # Docker image configuration
├── go.mod                          # Go module dependencies
├── go.sum                          # Go dependencies lock file
└── Readme.md                       # Documentation (file ini)
```

### Penjelasan Struktur:

- **`cmd/`** - Command line applications (entry point)
- **`internal/`** - Private application code (tidak bisa diimport dari package lain)
  - **controller** - Layer yang menangani HTTP requests dan responses
  - **model** - Struktur data dan database entities
  - **repository** - Layer yang mengelola akses data ke database
  - **service** - Logika bisnis aplikasi
  - **router** - Definisi routing dan dependency injection
- **`pkg/`** - Reusable packages yang bisa digunakan di tempat lain
  - **database** - Konfigurasi database dan migrations
  - **pdf** - Utility untuk PDF generation
  - **storage** - Integrasi dengan cloud storage (S3)

---

## 🚀 Cara Instalasi dan Menjalankan Project

### Prerequisites
Sebelum memulai, pastikan Anda telah menginstall:
- [Go 1.25.0](https://golang.org/dl/) atau lebih tinggi
- [PostgreSQL 15](https://www.postgresql.org/download/) atau [Docker](https://www.docker.com/)
- [Git](https://git-scm.com/)

### Langkah 1: Clone Repository

```bash
git clone https://github.com/muktiabdii/pdf-management-api.git
cd pdf-management-api
```

### Langkah 2: Setup Environment Variables

Buat file `.env` dengan menggandakan file `.env.example`:

```bash
cp .env.example .env
```

Edit file `.env` dan sesuaikan konfigurasi:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=pdf_dev
DB_PASSWORD=pdf_dev_pass
DB_NAME=pdf_management_db

# AWS S3 Configuration
AWS_REGION=ap-southeast-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_S3_BUCKET=your_bucket_name

# Server Configuration
PORT=8080
```

### Langkah 3: Install Dependencies

```bash
go mod download
go mod tidy
```

### Langkah 4A: Menjalankan dengan Docker Compose (Recommended)

```bash
docker-compose up -d
```

Command ini akan:
- Membangun Docker image untuk aplikasi
- Menjalankan PostgreSQL database container
- Menjalankan API server pada port `8080`

Cek status container:
```bash
docker-compose ps
```

### Langkah 4B: Menjalankan Lokal (Tanpa Docker)

#### Setup Database PostgreSQL

Jika menggunakan PostgreSQL lokal, buat database terlebih dahulu:

```sql
CREATE DATABASE pdf_management_db;
CREATE USER pdf_dev WITH PASSWORD 'pdf_dev_pass';
ALTER ROLE pdf_dev SET client_encoding TO 'utf8';
GRANT ALL PRIVILEGES ON DATABASE pdf_management_db TO pdf_dev;
```

#### Jalankan Aplikasi

```bash
go run ./cmd/api/main.go
```

Output yang diharapkan:
```
no .env file found, reading from environment
server running on port 8080
```

### Langkah 5: Verifikasi Aplikasi Berjalan

```bash
curl http://localhost:8080/api/pdf/list
```

---

## 📚 API Endpoints

### 1. **Generate PDF**
```http
POST /api/pdf/generate
Content-Type: application/json

{
  "data": {...}
}
```

### 2. **Upload PDF**
```http
POST /api/pdf/upload
Content-Type: multipart/form-data

FormData:
- file: [PDF file]
```

### 3. **List PDF Files**
```http
GET /api/pdf/list
```

Response:
```json
[
  {
    "id": 1,
    "filename": "document_123.pdf",
    "original_name": "my_document.pdf",
    "filepath": "s3://bucket/document_123.pdf",
    "size": 102400,
    "status": "UPLOADED",
    "created_at": "2026-05-13T10:30:00Z"
  }
]
```

### 4. **Delete PDF**
```http
DELETE /api/pdf/{id}
```

---

## 🛑 Menghentikan Aplikasi

### Jika menggunakan Docker Compose:
```bash
docker-compose down
```

### Jika running lokal:
Tekan `Ctrl + C` di terminal

---

## 📝 Notes

- Pastikan AWS S3 bucket sudah ada sebelum menjalankan aplikasi
- Database migrations berjalan otomatis saat aplikasi startup
- File uploads memiliki maksimal size 10MB (bisa diubah di router.go)
- Semua timestamps disimpan dalam format UTC

---

## 📄 Lisensi

Project ini menggunakan lisensi [MIT](LICENSE).

---

## 👤 Author

**Mukti Abdi** - [GitHub](https://github.com/muktiabdii)

---

## 📞 Support

Untuk pertanyaan atau issue, silakan buat issue di repository ini atau hubungi maintainer.
