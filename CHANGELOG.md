# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nothing yet

### Changed
- Nothing yet

### Deprecated
- Nothing yet

### Removed
- Nothing yet

### Fixed
- Nothing yet

### Security
- Nothing yet

## [1.1.0] - 2025-06-22

### Added
- **Automatic Unique File Naming** - Prevents file collisions with UUID-based naming
- `GeneratedKey` field in `DocumentUpload` struct for unique file storage keys
- `OriginalName` field in `DocumentUpload` struct for preserving user input
- `generateUniqueObjectName()` function that preserves directory structure
- `GenerateSignedUploadURLWithOriginalName()` method for overwriting files
- Support for both unique naming (default) and original naming (optional)

### Changed
- **Breaking Change**: `DocumentUpload` struct now includes `GeneratedKey` and `OriginalName` fields
- Main upload methods now automatically generate unique object names
- `GenerateSignedUploadURLWithExpiry()` updated to populate new struct fields

### Technical Details
- UUID generation using crypto/rand for collision-free file naming
- Directory structure preservation in unique naming (e.g., `users/123/file.pdf` → `users/123/a1b2c3d4_file.pdf`)
- Backward compatible API - existing methods work with enhanced functionality

## [1.0.0] - 2025-06-22

### Added
- Initial release of gcsurl library
- Core functionality for generating Google Cloud Storage signed URLs
- Support for upload URL generation with configurable expiry
- Support for download URL generation with configurable expiry
- Multiple authentication methods:
  - Service Account JSON string via `GCS_SERVICE_ACCOUNT_JSON`
  - Service Account JSON file via `GOOGLE_APPLICATION_CREDENTIALS`
  - Default credentials (Workload Identity, gcloud, etc.)
- Environment variable configuration:
  - `GCS_BUCKET_NAME` - Required default bucket name
  - `GCP_PROJECT_ID` - Optional project ID
  - `GCS_DEFAULT_EXPIRY_MINUTES` - Optional default expiry time (default: 15 minutes)
- `URLGenerator` struct with clean public API
- `Config` struct for programmatic configuration
- Utility methods for accessing configuration:
  - `GetBucketName()`
  - `GetProjectID()`
  - `GetDefaultExpiry()`
  - `GetDefaultExpiryMinutes()`
- Support for custom buckets and expiry times
- Storage client creation for advanced GCS operations
- Comprehensive documentation and examples
- Example application demonstrating all features

### Technical Details
- Built with Go 1.21+ compatibility
- Uses official Google Cloud Storage Go SDK
- No database dependencies (extracted from tropiko-filemanager)
- Follows Go best practices and conventions
- Comprehensive error handling with descriptive messages

[Unreleased]: https://github.com/tropikoearth/gcsurl/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/tropikoearth/gcsurl/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/tropikoearth/gcsurl/releases/tag/v1.0.0