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

[Unreleased]: https://github.com/jhonymendonca/gcsurl/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/jhonymendonca/gcsurl/releases/tag/v1.0.0