// Package gcsurl provides utilities for generating Google Cloud Storage signed URLs
// for uploading and downloading files without database dependencies.
package gcsurl

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// URLGenerator provides methods for generating signed URLs for Google Cloud Storage
type URLGenerator struct {
	serviceAccountKeyPath string
	svcAccount            *ServiceAccount
	serviceAccountJSON    []byte
	projectID             string
	bucketName            string
	defaultExpiry         time.Duration
	uploadRestrictions    UploadRestrictions
}

// ServiceAccount holds GCP service account credentials
type ServiceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}

// DocumentUpload contains the signed upload URL and expiration time
type DocumentUpload struct {
	UploadURL    string    `json:"uploadUrl"`    // The signed URL for upload
	ExpiresAt    time.Time `json:"expiresAt"`    // When the URL expires
	GeneratedKey string    `json:"generatedKey"` // Unique file path for storage
	OriginalName string    `json:"originalName"` // Original file name provided by user
}

// UploadRestrictions holds upload validation rules
type UploadRestrictions struct {
	AllowMultiple     bool     `json:"allowMultiple"`
	AllowedExtensions []string `json:"allowedExtensions"`
	MaxFileSizeMB     int64    `json:"maxFileSizeMB"`
	MaxFileSizeBytes  int64    `json:"maxFileSizeBytes"`
}

// Config holds configuration for the URLGenerator
type Config struct {
	ProjectID             string
	BucketName            string
	ServiceAccountKeyPath string
	DefaultExpiryMinutes  int
	UploadRestrictions    *UploadRestrictions
}

// NewURLGenerator creates a new URLGenerator instance
// It reads configuration from environment variables:
// - GCS_BUCKET_NAME: The default GCS bucket name (optional - will error if not provided)
// - GCP_PROJECT_ID: The GCP project ID (optional)
// - GCS_SERVICE_ACCOUNT_JSON: Service account JSON as string (preferred)
// - GOOGLE_APPLICATION_CREDENTIALS: Path to service account JSON file (fallback)
// - GCS_DEFAULT_EXPIRY_MINUTES: Default expiry time in minutes (default: 15)
func NewURLGenerator() (*URLGenerator, error) {
	return NewURLGeneratorWithRestrictions(nil)
}

// NewURLGeneratorWithBucket creates a new URLGenerator with a specific bucket
func NewURLGeneratorWithBucket(bucketName string) (*URLGenerator, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name cannot be empty")
	}
	config := Config{BucketName: bucketName}
	return NewURLGeneratorWithConfig(config)
}

// NewURLGeneratorWithBucketAndRestrictions creates a new URLGenerator with bucket and restrictions
func NewURLGeneratorWithBucketAndRestrictions(bucketName string, restrictions *UploadRestrictions) (*URLGenerator, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name cannot be empty")
	}
	config := Config{
		BucketName:         bucketName,
		UploadRestrictions: restrictions,
	}
	return NewURLGeneratorWithConfig(config)
}

// NewURLGeneratorWithRestrictions creates a new URLGenerator with upload restrictions
func NewURLGeneratorWithRestrictions(restrictions *UploadRestrictions) (*URLGenerator, error) {
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("GCS_BUCKET_NAME environment variable is required when using NewURLGenerator() or NewURLGeneratorWithRestrictions(). Use NewURLGeneratorWithBucket() or NewURLGeneratorWithConfig() to specify bucket explicitly")
	}

	projectID := os.Getenv("GCP_PROJECT_ID")

	// Parse default expiry from environment variable
	defaultExpiry := 15 * time.Minute // Default to 15 minutes
	if expiryStr := os.Getenv("GCS_DEFAULT_EXPIRY_MINUTES"); expiryStr != "" {
		if minutes, err := strconv.Atoi(expiryStr); err == nil && minutes > 0 {
			defaultExpiry = time.Duration(minutes) * time.Minute
		}
	}

	// Use provided restrictions or empty restrictions
	uploadRestrictions := UploadRestrictions{AllowMultiple: true} // Default: allow multiple
	if restrictions != nil {
		uploadRestrictions = *restrictions
	}
	
	var svcAccount *ServiceAccount
	var svcAccountJSON []byte
	var serviceAccountKeyPath string

	// Try to load service account from environment variable first (JSON string)
	if env := os.Getenv("GCS_SERVICE_ACCOUNT_JSON"); env != "" {
		var sa ServiceAccount
		if err := json.Unmarshal([]byte(env), &sa); err != nil {
			return nil, fmt.Errorf("failed to parse GCS_SERVICE_ACCOUNT_JSON: %w", err)
		}
		svcAccount = &sa
		svcAccountJSON = []byte(env)
	} else if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		// Fallback to file-based credentials
		data, err := os.ReadFile(credPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read service account file %s: %w", credPath, err)
		}
		var sa ServiceAccount
		if err := json.Unmarshal(data, &sa); err != nil {
			return nil, fmt.Errorf("failed to parse service account file %s: %w", credPath, err)
		}
		svcAccount = &sa
		svcAccountJSON = data
		serviceAccountKeyPath = credPath
	}

	// For Workload Identity or default credentials, svcAccount can be nil
	// The client will use default credentials automatically

	return &URLGenerator{
		serviceAccountKeyPath: serviceAccountKeyPath,
		svcAccount:            svcAccount,
		serviceAccountJSON:    svcAccountJSON,
		projectID:             projectID,
		bucketName:            bucketName,
		defaultExpiry:         defaultExpiry,
		uploadRestrictions:    uploadRestrictions,
	}, nil
}

// NewURLGeneratorWithConfig creates a new URLGenerator with the provided config
func NewURLGeneratorWithConfig(config Config) (*URLGenerator, error) {
	// Bucket name hierarchy: Config.BucketName > GCS_BUCKET_NAME env var > Error
	bucketName := config.BucketName
	if bucketName == "" {
		bucketName = os.Getenv("GCS_BUCKET_NAME")
	}
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name is required. Provide via Config.BucketName or GCS_BUCKET_NAME environment variable")
	}

	// Parse default expiry from config or environment variable
	defaultExpiry := 15 * time.Minute // Default to 15 minutes
	if config.DefaultExpiryMinutes > 0 {
		defaultExpiry = time.Duration(config.DefaultExpiryMinutes) * time.Minute
	} else if expiryStr := os.Getenv("GCS_DEFAULT_EXPIRY_MINUTES"); expiryStr != "" {
		if minutes, err := strconv.Atoi(expiryStr); err == nil && minutes > 0 {
			defaultExpiry = time.Duration(minutes) * time.Minute
		}
	}

	// Use restrictions from config or default
	uploadRestrictions := UploadRestrictions{AllowMultiple: true} // Default: allow multiple
	if config.UploadRestrictions != nil {
		uploadRestrictions = *config.UploadRestrictions
	}

	var svcAccount *ServiceAccount
	var svcAccountJSON []byte

	// Try to load service account from environment variable first
	if env := os.Getenv("GCS_SERVICE_ACCOUNT_JSON"); env != "" {
		var sa ServiceAccount
		if err := json.Unmarshal([]byte(env), &sa); err != nil {
			return nil, fmt.Errorf("failed to parse GCS_SERVICE_ACCOUNT_JSON: %w", err)
		}
		svcAccount = &sa
		svcAccountJSON = []byte(env)
	} else if config.ServiceAccountKeyPath != "" {
		data, err := os.ReadFile(config.ServiceAccountKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read service account file %s: %w", config.ServiceAccountKeyPath, err)
		}
		var sa ServiceAccount
		if err := json.Unmarshal(data, &sa); err != nil {
			return nil, fmt.Errorf("failed to parse service account file %s: %w", config.ServiceAccountKeyPath, err)
		}
		svcAccount = &sa
		svcAccountJSON = data
	}

	return &URLGenerator{
		serviceAccountKeyPath: config.ServiceAccountKeyPath,
		svcAccount:            svcAccount,
		serviceAccountJSON:    svcAccountJSON,
		projectID:             config.ProjectID,
		bucketName:            bucketName,
		defaultExpiry:         defaultExpiry,
		uploadRestrictions:    uploadRestrictions,
	}, nil
}

// getServiceAccount returns the loaded service account or an error
func (u *URLGenerator) getServiceAccount() (ServiceAccount, error) {
	if u.svcAccount != nil {
		return *u.svcAccount, nil
	}
	return ServiceAccount{}, fmt.Errorf("service account not loaded - configure GCS_SERVICE_ACCOUNT_JSON or GOOGLE_APPLICATION_CREDENTIALS")
}

// GenerateSignedUploadURL generates a signed URL for uploading a file to GCS with unique naming
// The URL expires after the configured default time. If upload restrictions are configured,
// they will be automatically applied (validation + content-type detection + size limits).
// Automatically generates unique object names to prevent collisions while preserving directory structure.
func (u *URLGenerator) GenerateSignedUploadURL(ctx context.Context, objectName string) (DocumentUpload, error) {
	// Generate unique object name
	uniqueObjectName, err := u.generateUniqueObjectName(objectName)
	if err != nil {
		return DocumentUpload{}, fmt.Errorf("failed to generate unique object name: %w", err)
	}

	// Apply validation if restrictions are configured
	if u.hasRestrictions() {
		if err := u.ValidateUpload(objectName); err != nil {
			return DocumentUpload{}, err
		}
		upload, err := u.generateUploadURLWithRestrictions(ctx, u.bucketName, uniqueObjectName, u.defaultExpiry)
		if err != nil {
			return DocumentUpload{}, err
		}
		upload.GeneratedKey = uniqueObjectName
		upload.OriginalName = objectName
		return upload, nil
	}

	// No restrictions - use simple generation
	upload, err := u.GenerateSignedUploadURLWithExpiry(ctx, u.bucketName, uniqueObjectName, u.defaultExpiry)
	if err != nil {
		return DocumentUpload{}, err
	}
	upload.GeneratedKey = uniqueObjectName
	upload.OriginalName = objectName
	return upload, nil
}

// GenerateSignedUploadURLWithBucket generates a signed URL for uploading to a specific bucket with unique naming
// If upload restrictions are configured, they will be automatically applied.
// Automatically generates unique object names to prevent collisions while preserving directory structure.
func (u *URLGenerator) GenerateSignedUploadURLWithBucket(ctx context.Context, bucketName, objectName string) (DocumentUpload, error) {
	// Generate unique object name
	uniqueObjectName, err := u.generateUniqueObjectName(objectName)
	if err != nil {
		return DocumentUpload{}, fmt.Errorf("failed to generate unique object name: %w", err)
	}

	// Apply validation if restrictions are configured
	if u.hasRestrictions() {
		if err := u.ValidateUpload(objectName); err != nil {
			return DocumentUpload{}, err
		}
		upload, err := u.generateUploadURLWithRestrictions(ctx, bucketName, uniqueObjectName, u.defaultExpiry)
		if err != nil {
			return DocumentUpload{}, err
		}
		upload.GeneratedKey = uniqueObjectName
		upload.OriginalName = objectName
		return upload, nil
	}

	// No restrictions - use simple generation
	upload, err := u.GenerateSignedUploadURLWithExpiry(ctx, bucketName, uniqueObjectName, u.defaultExpiry)
	if err != nil {
		return DocumentUpload{}, err
	}
	upload.GeneratedKey = uniqueObjectName
	upload.OriginalName = objectName
	return upload, nil
}

// GenerateSignedUploadURLWithExpiry generates a signed URL for uploading with custom expiry
// This method does NOT generate unique names - it uses the exact objectName provided.
// Use this when you want to overwrite existing files or when you manage naming yourself.
func (u *URLGenerator) GenerateSignedUploadURLWithExpiry(ctx context.Context, bucketName, objectName string, expiry time.Duration) (DocumentUpload, error) {
	sa, err := u.getServiceAccount()
	if err != nil {
		return DocumentUpload{}, err
	}

	expires := time.Now().Add(expiry)
	opts := &storage.SignedURLOptions{
		Method:         "PUT",
		Expires:        expires,
		ContentType:    "application/octet-stream",
		GoogleAccessID: sa.ClientEmail,
		PrivateKey:     []byte(sa.PrivateKey),
	}

	signedURL, err := storage.SignedURL(bucketName, objectName, opts)
	if err != nil {
		return DocumentUpload{}, fmt.Errorf("failed to generate signed upload URL: %w", err)
	}

	return DocumentUpload{
		UploadURL:    signedURL,
		ExpiresAt:    expires,
		GeneratedKey: objectName, // Same as original when no unique naming
		OriginalName: objectName,
	}, nil
}

// GenerateSignedDownloadURL generates a signed URL for downloading a file from the default bucket
// The URL expires after the configured default time
func (u *URLGenerator) GenerateSignedDownloadURL(ctx context.Context, objectName string) (string, error) {
	return u.GenerateSignedDownloadURLWithExpiry(ctx, u.bucketName, objectName, u.defaultExpiry)
}

// GenerateSignedDownloadURLWithBucket generates a signed URL for downloading from a specific bucket
func (u *URLGenerator) GenerateSignedDownloadURLWithBucket(ctx context.Context, bucketName, objectName string) (string, error) {
	return u.GenerateSignedDownloadURLWithExpiry(ctx, bucketName, objectName, u.defaultExpiry)
}

// GenerateSignedDownloadURLWithExpiry generates a signed URL for downloading with custom expiry
func (u *URLGenerator) GenerateSignedDownloadURLWithExpiry(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	sa, err := u.getServiceAccount()
	if err != nil {
		return "", err
	}

	expires := time.Now().Add(expiry)
	opts := &storage.SignedURLOptions{
		Method:         "GET",
		Expires:        expires,
		GoogleAccessID: sa.ClientEmail,
		PrivateKey:     []byte(sa.PrivateKey),
	}

	signedURL, err := storage.SignedURL(bucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed download URL: %w", err)
	}

	return signedURL, nil
}

// CreateStorageClient creates a GCS client for advanced operations
// This can be useful if you need to perform additional GCS operations beyond URL generation
func (u *URLGenerator) CreateStorageClient(ctx context.Context) (*storage.Client, error) {
	if len(u.serviceAccountJSON) > 0 {
		client, err := storage.NewClient(ctx, option.WithCredentialsJSON(u.serviceAccountJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to create storage client with service account: %w", err)
		}
		return client, nil
	}

	// Fallback to default credentials (Workload Identity, etc.)
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client with default credentials: %w", err)
	}
	return client, nil
}

// GetBucketName returns the configured default bucket name
func (u *URLGenerator) GetBucketName() string {
	return u.bucketName
}

// GetProjectID returns the configured project ID
func (u *URLGenerator) GetProjectID() string {
	return u.projectID
}

// GetDefaultExpiry returns the configured default expiry duration
func (u *URLGenerator) GetDefaultExpiry() time.Duration {
	return u.defaultExpiry
}

// GetDefaultExpiryMinutes returns the configured default expiry in minutes
func (u *URLGenerator) GetDefaultExpiryMinutes() int {
	return int(u.defaultExpiry.Minutes())
}

// NewUploadRestrictionsFromEnv creates UploadRestrictions from environment variables
// Use this helper if you want to read restrictions from env vars
func NewUploadRestrictionsFromEnv() *UploadRestrictions {
	restrictions := &UploadRestrictions{
		AllowMultiple: true, // Default: allow multiple uploads
	}

	// Parse allow multiple uploads
	if allowMultiple := os.Getenv("GCS_ALLOW_MULTIPLE_UPLOADS"); allowMultiple != "" {
		restrictions.AllowMultiple = strings.ToLower(allowMultiple) == "true"
	}

	// Parse allowed file extensions
	if extensions := os.Getenv("GCS_ALLOWED_FILE_EXTENSIONS"); extensions != "" {
		extList := strings.Split(extensions, ",")
		for i, ext := range extList {
			ext = strings.TrimSpace(ext)
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			extList[i] = strings.ToLower(ext)
		}
		restrictions.AllowedExtensions = extList
	}

	// Parse max file size
	if maxSizeStr := os.Getenv("GCS_MAX_FILE_SIZE_MB"); maxSizeStr != "" {
		if maxSizeMB, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil && maxSizeMB > 0 {
			restrictions.MaxFileSizeMB = maxSizeMB
			restrictions.MaxFileSizeBytes = maxSizeMB * 1024 * 1024
		}
	}

	// Only return restrictions if at least one was configured
	if !restrictions.AllowMultiple || len(restrictions.AllowedExtensions) > 0 || restrictions.MaxFileSizeMB > 0 {
		return restrictions
	}
	return nil
}

// ValidateUpload validates if an upload meets the configured restrictions
func (u *URLGenerator) ValidateUpload(filename string) error {
	// Check file extension if restrictions are set
	if len(u.uploadRestrictions.AllowedExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(filename))
		allowed := false
		for _, allowedExt := range u.uploadRestrictions.AllowedExtensions {
			if ext == allowedExt {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file extension %s not allowed. Allowed extensions: %v", ext, u.uploadRestrictions.AllowedExtensions)
		}
	}

	return nil
}


// generateUploadURLWithRestrictions generates upload URL applying all restrictions
func (u *URLGenerator) generateUploadURLWithRestrictions(ctx context.Context, bucketName, objectName string, expiry time.Duration) (DocumentUpload, error) {
	sa, err := u.getServiceAccount()
	if err != nil {
		return DocumentUpload{}, err
	}

	expires := time.Now().Add(expiry)

	// Build headers based on restrictions
	var headers []string
	if u.uploadRestrictions.MaxFileSizeBytes > 0 {
		headers = append(headers, fmt.Sprintf("Content-Length:%d", u.uploadRestrictions.MaxFileSizeBytes))
	}

	// Determine content type based on file extension
	contentType := "application/octet-stream"
	if ext := strings.ToLower(filepath.Ext(objectName)); ext != "" {
		contentType = getContentTypeFromExtension(ext)
	}

	opts := &storage.SignedURLOptions{
		Method:         "PUT",
		Expires:        expires,
		ContentType:    contentType,
		Headers:        headers,
		GoogleAccessID: sa.ClientEmail,
		PrivateKey:     []byte(sa.PrivateKey),
	}

	signedURL, err := storage.SignedURL(bucketName, objectName, opts)
	if err != nil {
		return DocumentUpload{}, fmt.Errorf("failed to generate validated upload URL: %w", err)
	}

	return DocumentUpload{
		UploadURL: signedURL,
		ExpiresAt: expires,
		// GeneratedKey and OriginalName will be set by the calling function
	}, nil
}

// getContentTypeFromExtension returns the appropriate content type for a file extension
func getContentTypeFromExtension(ext string) string {
	contentTypes := map[string]string{
		".pdf":  "application/pdf",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}
	return "application/octet-stream"
}

// GetUploadRestrictions returns the current upload restrictions
func (u *URLGenerator) GetUploadRestrictions() UploadRestrictions {
	return u.uploadRestrictions
}

// hasRestrictions checks if any upload restrictions are configured
func (u *URLGenerator) hasRestrictions() bool {
	return len(u.uploadRestrictions.AllowedExtensions) > 0 || 
		u.uploadRestrictions.MaxFileSizeMB > 0 || 
		!u.uploadRestrictions.AllowMultiple
}

// HasUploadRestrictions returns true if upload restrictions are configured
func (u *URLGenerator) HasUploadRestrictions() bool {
	return u.hasRestrictions()
}

// generateUniqueObjectName generates a unique object name while preserving directory structure
// Input: "documents/file.pdf" -> Output: "documents/a1b2c3d4_file.pdf"
// Input: "file.pdf" -> Output: "a1b2c3d4_file.pdf"
func (u *URLGenerator) generateUniqueObjectName(originalPath string) (string, error) {
	// Generate UUID-like identifier (8 chars)
	uuid, err := generateShortUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}

	// Split path into directory and filename
	dir := filepath.Dir(originalPath)
	filename := filepath.Base(originalPath)

	// Split filename into name and extension
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Create unique filename: uuid_originalname.ext
	uniqueFilename := fmt.Sprintf("%s_%s%s", uuid, nameWithoutExt, ext)

	// Reconstruct full path
	if dir == "." {
		// No directory, just return unique filename
		return uniqueFilename, nil
	}

	// Combine directory with unique filename
	return filepath.Join(dir, uniqueFilename), nil
}

// generateShortUUID generates a short UUID-like string (8 characters)
func generateShortUUID() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

// GenerateSignedUploadURLWithOriginalName generates a signed URL using the original object name
// This method does NOT generate unique names - use this when you want to overwrite existing files
func (u *URLGenerator) GenerateSignedUploadURLWithOriginalName(ctx context.Context, objectName string) (DocumentUpload, error) {
	// Apply validation if restrictions are configured
	if u.hasRestrictions() {
		if err := u.ValidateUpload(objectName); err != nil {
			return DocumentUpload{}, err
		}
		upload, err := u.generateUploadURLWithRestrictions(ctx, u.bucketName, objectName, u.defaultExpiry)
		if err != nil {
			return DocumentUpload{}, err
		}
		upload.GeneratedKey = objectName
		upload.OriginalName = objectName
		return upload, nil
	}

	// No restrictions - use simple generation
	upload, err := u.GenerateSignedUploadURLWithExpiry(ctx, u.bucketName, objectName, u.defaultExpiry)
	if err != nil {
		return DocumentUpload{}, err
	}
	return upload, nil
}