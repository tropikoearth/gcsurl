# GCS URL - Google Cloud Storage Signed URL Generator

A lightweight Go library for generating signed URLs for Google Cloud Storage upload and download operations. This library provides a simple interface for creating time-limited URLs without requiring database dependencies.

## Features

- ✅ Generate signed upload URLs for GCS
- ✅ Generate signed download URLs for GCS  
- ✅ Support for custom expiry times
- ✅ Multiple authentication methods (Service Account JSON, file, Workload Identity)
- ✅ Environment variable configuration
- ✅ No database dependencies
- ✅ Clean, simple API

## Installation

```bash
go get github.com/jhonymendonca/gcsurl
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jhonymendonca/gcsurl"
)

func main() {
    // Create URL generator (reads from environment variables)
    generator, err := gcsurl.NewURLGenerator()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Generate upload URL
    upload, err := generator.GenerateSignedUploadURL(ctx, "my-file.pdf")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Upload URL: %s\\n", upload.UploadURL)
    fmt.Printf("Expires at: %s\\n", upload.ExpiresAt)

    // Generate download URL
    downloadURL, err := generator.GenerateSignedDownloadURL(ctx, "my-file.pdf")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Download URL: %s\\n", downloadURL)
}
```

## Configuration

### Environment Variables

The library uses the following environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `GCS_BUCKET_NAME` | Default GCS bucket name (when not specified in constructor) | ❌ No** |
| `GCS_SERVICE_ACCOUNT_JSON` | Service account JSON as string | ❌ No* |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to service account JSON file | ❌ No* |
| `GCP_PROJECT_ID` | GCP Project ID | ❌ No |
| `GCS_DEFAULT_EXPIRY_MINUTES` | Default URL expiry time in minutes (default: 15) | ❌ No |

*At least one authentication method is required  
**Required only when using `NewURLGenerator()` or `NewURLGeneratorWithRestrictions()`. Other constructors accept bucket as parameter.

### Authentication Priority

The library attempts authentication in this order:

1. **Service Account JSON String** (`GCS_SERVICE_ACCOUNT_JSON`)
2. **Service Account JSON File** (`GOOGLE_APPLICATION_CREDENTIALS`)  
3. **Default Credentials** (Workload Identity, gcloud, etc.)

### Example Environment Setup

```bash
# Required
export GCS_BUCKET_NAME="my-storage-bucket"

# Option 1: JSON string (recommended for containers)
export GCS_SERVICE_ACCOUNT_JSON='{"type":"service_account","project_id":"my-project",...}'

# Option 2: JSON file path (good for local development)
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"

# Optional
export GCP_PROJECT_ID="my-gcp-project"
export GCS_DEFAULT_EXPIRY_MINUTES="30"  # Default: 15 minutes
```

## Usage Examples

### Basic Usage

```go
// Option 1: Using environment variable (requires GCS_BUCKET_NAME)
generator, err := gcsurl.NewURLGenerator()
if err != nil {
    log.Fatal(err)
}

// Option 2: Explicit bucket (recommended)
generator, err := gcsurl.NewURLGeneratorWithBucket("my-documents-bucket")
if err != nil {
    log.Fatal(err)
}

// Upload URL with unique name generation
upload, err := generator.GenerateSignedUploadURL(ctx, "documents/file.pdf")
// upload.GeneratedKey = "documents/a1b2c3d4_file.pdf" (save this in database)
// upload.OriginalName = "documents/file.pdf" (original user input)

// Download URL using the generated key from database
downloadURL, err := generator.GenerateSignedDownloadURL(ctx, upload.GeneratedKey)
```

### Multiple Buckets

```go
// Different buckets for different purposes
docsGenerator, err := gcsurl.NewURLGeneratorWithBucket("company-documents")
mediaGenerator, err := gcsurl.NewURLGeneratorWithBucket("user-media") 
reportsGenerator, err := gcsurl.NewURLGeneratorWithBucket("generated-reports")

// Each generator works with its specific bucket
docUpload, _ := docsGenerator.GenerateSignedUploadURL(ctx, "contract.pdf")
imageUpload, _ := mediaGenerator.GenerateSignedUploadURL(ctx, "photo.jpg")
reportUpload, _ := reportsGenerator.GenerateSignedUploadURL(ctx, "monthly-report.xlsx")
```

### Usage with Upload Restrictions & Unique Names

```go
// Option 1: Specific bucket + restrictions
restrictions := &gcsurl.UploadRestrictions{
    AllowMultiple:     true,
    AllowedExtensions: []string{".pdf", ".jpg", ".png"},
    MaxFileSizeMB:     10, // 10MB limit
}
generator, err := gcsurl.NewURLGeneratorWithBucketAndRestrictions("secure-documents", restrictions)

// Upload URL (restrictions automatically applied + unique name generated)
upload, err := generator.GenerateSignedUploadURL(ctx, "users/123/contract.pdf")

// Result:
// upload.UploadURL = "https://storage.googleapis.com/secure-documents/users/123/a1b2c3d4_contract.pdf?signature=..."
// upload.GeneratedKey = "users/123/a1b2c3d4_contract.pdf"  // Save this in database!
// upload.OriginalName = "users/123/contract.pdf"           // Original user input
// upload.ExpiresAt = time.Now().Add(15*time.Minute)

// The library will:
// 1. Generate unique filename while preserving directory structure
// 2. Validate file extension (.pdf is allowed)
// 3. Set appropriate Content-Type (application/pdf)
// 4. Apply size limits via Content-Length header
```

### Different Restrictions per Bucket

```go
// Documents bucket - only office files
docRestrictions := &gcsurl.UploadRestrictions{
    AllowedExtensions: []string{".pdf", ".doc", ".docx", ".xlsx"},
    MaxFileSizeMB:     20,
}
docsGen, _ := gcsurl.NewURLGeneratorWithBucketAndRestrictions("company-docs", docRestrictions)

// Media bucket - only images/videos
mediaRestrictions := &gcsurl.UploadRestrictions{
    AllowedExtensions: []string{".jpg", ".png", ".mp4", ".webm"},
    MaxFileSizeMB:     100,
}
mediaGen, _ := gcsurl.NewURLGeneratorWithBucketAndRestrictions("user-media", mediaRestrictions)

// Each generator applies its own restrictions
docUpload, _ := docsGen.GenerateSignedUploadURL(ctx, "report.pdf")    // ✅ Allowed
mediaUpload, _ := mediaGen.GenerateSignedUploadURL(ctx, "video.mp4")  // ✅ Allowed
invalidDoc, _ := docsGen.GenerateSignedUploadURL(ctx, "image.jpg")    // ❌ Error: not allowed
```

### Using Environment Variables for Restrictions

```go
// Read restrictions from environment variables
restrictions := gcsurl.NewUploadRestrictionsFromEnv()

var generator *gcsurl.URLGenerator
if restrictions != nil {
    generator, err = gcsurl.NewURLGeneratorWithRestrictions(restrictions)
} else {
    generator, err = gcsurl.NewURLGenerator()
}
```

Set environment variables:
```bash
export GCS_ALLOW_MULTIPLE_UPLOADS="true"
export GCS_ALLOWED_FILE_EXTENSIONS=".pdf,.jpg,.png"
export GCS_MAX_FILE_SIZE_MB="10"
```

### Advanced Usage

```go
// Custom configuration
config := gcsurl.Config{
    ProjectID:             "my-project",
    BucketName:            "my-bucket",
    ServiceAccountKeyPath: "/path/to/key.json",
}
generator, err := gcsurl.NewURLGeneratorWithConfig(config)

// Custom bucket and expiry
upload, err := generator.GenerateSignedUploadURLWithExpiry(
    ctx, 
    "different-bucket", 
    "file.pdf", 
    1*time.Hour, // 1 hour expiry
)

// Different bucket
downloadURL, err := generator.GenerateSignedDownloadURLWithBucket(
    ctx,
    "downloads-bucket", 
    "public/file.pdf",
)
```

### Docker Usage

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .

# Set environment variables
ENV GCS_BUCKET_NAME=my-bucket
ENV GCS_SERVICE_ACCOUNT_JSON='{"type":"service_account",...}'

CMD ["./main"]
```

### Kubernetes with Workload Identity

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    metadata:
      annotations:
        iam.gke.io/gcp-service-account: my-service-account@my-project.iam.gserviceaccount.com
    spec:
      serviceAccountName: my-k8s-service-account
      containers:
      - name: app
        image: my-app:latest
        env:
        - name: GCS_BUCKET_NAME
          value: "my-bucket"
        # No need to set GCS_SERVICE_ACCOUNT_JSON - Workload Identity handles it
```

## API Reference

### Types

```go
type URLGenerator struct {
    // Private fields
}

type DocumentUpload struct {
    UploadURL    string    `json:"uploadUrl"`    // The signed URL for upload
    ExpiresAt    time.Time `json:"expiresAt"`    // When the URL expires
    GeneratedKey string    `json:"generatedKey"` // Unique file path for storage (save this in database)
    OriginalName string    `json:"originalName"` // Original file name provided by user
}

type UploadRestrictions struct {
    AllowMultiple     bool     `json:"allowMultiple"`
    AllowedExtensions []string `json:"allowedExtensions"`
    MaxFileSizeMB     int64    `json:"maxFileSizeMB"`
    MaxFileSizeBytes  int64    `json:"maxFileSizeBytes"`
}

type Config struct {
    ProjectID             string
    BucketName            string  
    ServiceAccountKeyPath string
    DefaultExpiryMinutes  int
    UploadRestrictions    *UploadRestrictions
}
```

### Methods

#### Constructor Methods

```go
// Create with environment variables (requires GCS_BUCKET_NAME)
func NewURLGenerator() (*URLGenerator, error)

// Create with specific bucket (recommended)
func NewURLGeneratorWithBucket(bucketName string) (*URLGenerator, error)

// Create with bucket and restrictions
func NewURLGeneratorWithBucketAndRestrictions(bucketName string, restrictions *UploadRestrictions) (*URLGenerator, error)

// Create with environment bucket + restrictions
func NewURLGeneratorWithRestrictions(restrictions *UploadRestrictions) (*URLGenerator, error)

// Create with explicit config (flexible bucket hierarchy)
func NewURLGeneratorWithConfig(config Config) (*URLGenerator, error)

// Helper to read restrictions from environment variables
func NewUploadRestrictionsFromEnv() *UploadRestrictions
```

#### Upload URL Methods

```go
// Default bucket, default expiry (applies restrictions + generates unique name)
func (u *URLGenerator) GenerateSignedUploadURL(ctx context.Context, objectName string) (DocumentUpload, error)

// Custom bucket, default expiry (applies restrictions + generates unique name)
func (u *URLGenerator) GenerateSignedUploadURLWithBucket(ctx context.Context, bucketName, objectName string) (DocumentUpload, error)

// Use original name without UUID generation (for overwrites)
func (u *URLGenerator) GenerateSignedUploadURLWithOriginalName(ctx context.Context, objectName string) (DocumentUpload, error)

// Custom bucket and expiry (no restrictions, no unique name generation)
func (u *URLGenerator) GenerateSignedUploadURLWithExpiry(ctx context.Context, bucketName, objectName string, expiry time.Duration) (DocumentUpload, error)

// Validate upload against restrictions (manual validation)
func (u *URLGenerator) ValidateUpload(filename string) error
```

#### Download URL Methods

```go
// Default bucket, default expiry
func (u *URLGenerator) GenerateSignedDownloadURL(ctx context.Context, objectName string) (string, error)

// Custom bucket, default expiry
func (u *URLGenerator) GenerateSignedDownloadURLWithBucket(ctx context.Context, bucketName, objectName string) (string, error)

// Custom bucket and expiry
func (u *URLGenerator) GenerateSignedDownloadURLWithExpiry(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error)
```

#### Utility Methods

```go
// Create GCS client for advanced operations
func (u *URLGenerator) CreateStorageClient(ctx context.Context) (*storage.Client, error)

// Get configured bucket name
func (u *URLGenerator) GetBucketName() string

// Get configured project ID
func (u *URLGenerator) GetProjectID() string

// Get configured default expiry duration
func (u *URLGenerator) GetDefaultExpiry() time.Duration

// Get configured default expiry in minutes
func (u *URLGenerator) GetDefaultExpiryMinutes() int

// Get current upload restrictions
func (u *URLGenerator) GetUploadRestrictions() UploadRestrictions

// Check if upload restrictions are configured
func (u *URLGenerator) HasUploadRestrictions() bool
```

## Database Integration

Since the library now generates unique file names automatically, here's how to integrate with your database:

```go
type Document struct {
    ID           uuid.UUID `json:"id"`
    FileKey      string    `json:"fileKey"`      // Generated unique key from library
    OriginalName string    `json:"originalName"` // Original filename from user
    BucketName   string    `json:"bucketName"`
    ContentType  string    `json:"contentType"`
    FileSize     int64     `json:"fileSize"`
    UserID       uuid.UUID `json:"userId"`
    UploadedAt   time.Time `json:"uploadedAt"`
    CreatedAt    time.Time `json:"createdAt"`
}

// Upload flow
func HandleUpload(userID uuid.UUID, filename string) (*Document, error) {
    generator, _ := gcsurl.NewURLGeneratorWithBucket("documents")
    
    // Generate upload URL with unique name
    upload, err := generator.GenerateSignedUploadURL(ctx, filename)
    if err != nil {
        return nil, err
    }
    
    // Save to database with generated key
    doc := &Document{
        ID:           uuid.New(),
        FileKey:      upload.GeneratedKey,  // Use this for future GCS operations
        OriginalName: upload.OriginalName,  // Use this for user display
        BucketName:   generator.GetBucketName(),
        UserID:       userID,
        CreatedAt:    time.Now(),
    }
    
    db.Save(doc)
    
    return doc, nil
}

// Download flow
func HandleDownload(docID uuid.UUID) (string, error) {
    doc, _ := db.GetDocument(docID)
    
    generator, _ := gcsurl.NewURLGeneratorWithBucket(doc.BucketName)
    
    // Use the stored GeneratedKey for download
    downloadURL, err := generator.GenerateSignedDownloadURL(ctx, doc.FileKey)
    return downloadURL, err
}
```

### Unique Name Generation Examples

The library automatically generates unique names while preserving directory structure:

```go
// Input → Generated Key
"document.pdf" → "a1b2c3d4_document.pdf"
"photos/avatar.jpg" → "photos/e5f6g7h8_avatar.jpg"  
"users/123/files/contract.pdf" → "users/123/files/i9j0k1l2_contract.pdf"
"reports/2025/january.xlsx" → "reports/2025/m3n4o5p6_january.xlsx"
```

## Error Handling

The library returns descriptive errors for common issues:

```go
// Configuration errors
generator, err := gcsurl.NewURLGenerator()
if err != nil {
    log.Printf("Configuration error: %v", err)
    return
}

// Upload validation errors (when restrictions are configured)
upload, err := generator.GenerateSignedUploadURL(ctx, "script.js")
if err != nil {
    // Will return validation error if .js files are not allowed
    log.Printf("Upload validation failed: %v", err)
    return
}

// Manual validation
if err := generator.ValidateUpload("document.exe"); err != nil {
    log.Printf("File not allowed: %v", err)
    return
}
```

Common errors:
- Missing required environment variables
- Invalid service account JSON
- GCS API authentication failures
- Invalid bucket or object names
- **Upload validation failures** (file extension not allowed, size limits, etc.)

## Requirements

- Go 1.21 or higher
- Valid Google Cloud Platform project
- Service account with Storage Admin or Storage Object Admin role

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

For issues and questions:
- Create an issue on GitHub
- Check the [Google Cloud Storage documentation](https://cloud.google.com/storage/docs)
- Review the [Service Account setup guide](https://cloud.google.com/iam/docs/service-accounts)