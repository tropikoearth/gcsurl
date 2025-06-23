package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tropikoearth/gcsurl"
)

func main() {
	fmt.Println("🔗 GCS URL Generator Example - Multiple Buckets & Restrictions")
	fmt.Println("================================================================")
	fmt.Println()

	ctx := context.Background()

	// Example 1: Multiple buckets for different purposes
	fmt.Println("📦 Example 1: Multiple Buckets for Different Services")
	
	// Documents service bucket
	docsGenerator, err := gcsurl.NewURLGeneratorWithBucket("company-documents")
	if err != nil {
		log.Printf("❌ Failed to create docs generator: %v", err)
	} else {
		fmt.Printf("✅ Documents generator created (bucket: %s)\\n", docsGenerator.GetBucketName())
	}

	// Media service bucket
	mediaGenerator, err := gcsurl.NewURLGeneratorWithBucket("user-media")
	if err != nil {
		log.Printf("❌ Failed to create media generator: %v", err)
	} else {
		fmt.Printf("✅ Media generator created (bucket: %s)\\n", mediaGenerator.GetBucketName())
	}

	// Reports service bucket
	reportsGenerator, err := gcsurl.NewURLGeneratorWithBucket("generated-reports")
	if err != nil {
		log.Printf("❌ Failed to create reports generator: %v", err)
	} else {
		fmt.Printf("✅ Reports generator created (bucket: %s)\\n", reportsGenerator.GetBucketName())
	}
	fmt.Println()

	// Example 2: Different restrictions per bucket
	fmt.Println("🛡️  Example 2: Different Upload Restrictions per Bucket")
	
	// Documents bucket - only office files, larger size limit
	docRestrictions := &gcsurl.UploadRestrictions{
		AllowMultiple:     true,
		AllowedExtensions: []string{".pdf", ".doc", ".docx", ".xlsx", ".pptx"},
		MaxFileSizeMB:     25, // 25MB for documents
	}
	restrictedDocsGen, err := gcsurl.NewURLGeneratorWithBucketAndRestrictions("secure-documents", docRestrictions)
	if err != nil {
		log.Printf("❌ Failed to create restricted docs generator: %v", err)
	} else {
		fmt.Printf("📋 Secure documents bucket configured:\\n")
		fmt.Printf("   Bucket: %s\\n", restrictedDocsGen.GetBucketName())
		fmt.Printf("   Restrictions: %t\\n", restrictedDocsGen.HasUploadRestrictions())
		restrictions := restrictedDocsGen.GetUploadRestrictions()
		fmt.Printf("   Allowed: %v\\n", restrictions.AllowedExtensions)
		fmt.Printf("   Max size: %d MB\\n", restrictions.MaxFileSizeMB)
	}

	// Media bucket - only images/videos, smaller size limit
	mediaRestrictions := &gcsurl.UploadRestrictions{
		AllowMultiple:     false, // Single file uploads only
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".mp4", ".webm"},
		MaxFileSizeMB:     50, // 50MB for media
	}
	restrictedMediaGen, err := gcsurl.NewURLGeneratorWithBucketAndRestrictions("user-uploads", mediaRestrictions)
	if err != nil {
		log.Printf("❌ Failed to create restricted media generator: %v", err)
	} else {
		fmt.Printf("🎨 User uploads bucket configured:\\n")
		fmt.Printf("   Bucket: %s\\n", restrictedMediaGen.GetBucketName())
		fmt.Printf("   Multiple uploads: %t\\n", restrictedMediaGen.GetUploadRestrictions().AllowMultiple)
		fmt.Printf("   Allowed: %v\\n", restrictedMediaGen.GetUploadRestrictions().AllowedExtensions)
		fmt.Printf("   Max size: %d MB\\n", restrictedMediaGen.GetUploadRestrictions().MaxFileSizeMB)
	}
	fmt.Println()

	// Example 3: Testing different file types across buckets
	fmt.Println("🧪 Example 3: Testing File Uploads Across Different Buckets")
	
	testFiles := []struct {
		name      string
		generator *gcsurl.URLGenerator
		label     string
	}{
		{"contract.pdf", restrictedDocsGen, "Secure Docs"},
		{"presentation.pptx", restrictedDocsGen, "Secure Docs"},
		{"photo.jpg", restrictedMediaGen, "User Uploads"},
		{"video.mp4", restrictedMediaGen, "User Uploads"},
		// These should fail
		{"image.png", restrictedDocsGen, "Secure Docs"}, // Image in docs bucket
		{"document.pdf", restrictedMediaGen, "User Uploads"}, // Document in media bucket
	}

	for _, test := range testFiles {
		if test.generator == nil {
			continue
		}
		
		upload, err := test.generator.GenerateSignedUploadURL(ctx, test.name)
		if err != nil {
			fmt.Printf("❌ %s (%s): %v\\n", test.name, test.label, err)
		} else {
			fmt.Printf("✅ %s (%s): URL generated\\n", test.name, test.label)
			fmt.Printf("   Expires: %s\\n", upload.ExpiresAt.Format("15:04:05"))
		}
	}
	fmt.Println()

	// Example 4: Using environment variables (fallback)
	fmt.Println("🌍 Example 4: Environment Variable Fallback")
	
	// Try to read from environment
	envGenerator, err := gcsurl.NewURLGenerator()
	if err != nil {
		fmt.Printf("ℹ️  NewURLGenerator() failed: %v\\n", err)
		fmt.Println("   💡 This is expected when GCS_BUCKET_NAME is not set")
		fmt.Println("   💡 Use NewURLGeneratorWithBucket() for explicit bucket")
	} else {
		fmt.Printf("✅ Environment generator created (bucket: %s)\\n", envGenerator.GetBucketName())
	}

	// Show how config hierarchy works
	fmt.Println("\\n⚙️  Example 4.1: Config Hierarchy (Config > Env > Error)")
	config := gcsurl.Config{
		BucketName: "config-specified-bucket",
		ProjectID:  "my-project",
	}
	configGenerator, err := gcsurl.NewURLGeneratorWithConfig(config)
	if err != nil {
		log.Printf("❌ Failed to create config generator: %v", err)
	} else {
		fmt.Printf("✅ Config generator uses: %s (from Config.BucketName)\\n", configGenerator.GetBucketName())
	}
	fmt.Println()

	// Example 5: Reading restrictions from environment
	fmt.Println("📋 Example 5: Reading Restrictions from Environment Variables")
	
	envRestrictions := gcsurl.NewUploadRestrictionsFromEnv()
	if envRestrictions != nil {
		fmt.Println("✅ Found restrictions in environment:")
		fmt.Printf("   Multiple uploads: %t\\n", envRestrictions.AllowMultiple)
		if len(envRestrictions.AllowedExtensions) > 0 {
			fmt.Printf("   Allowed extensions: %v\\n", envRestrictions.AllowedExtensions)
		}
		if envRestrictions.MaxFileSizeMB > 0 {
			fmt.Printf("   Max file size: %d MB\\n", envRestrictions.MaxFileSizeMB)
		}
		
		// Use with specific bucket
		envRestrictedGen, err := gcsurl.NewURLGeneratorWithBucketAndRestrictions("env-bucket", envRestrictions)
		if err != nil {
			log.Printf("❌ Failed to create env restricted generator: %v", err)
		} else {
			fmt.Printf("✅ Created generator with env restrictions for bucket: %s\\n", envRestrictedGen.GetBucketName())
		}
	} else {
		fmt.Println("ℹ️  No restrictions found in environment variables")
		fmt.Println("   Set these environment variables to test:")
		fmt.Println("   export GCS_ALLOW_MULTIPLE_UPLOADS=\"false\"")
		fmt.Println("   export GCS_ALLOWED_FILE_EXTENSIONS=\".pdf,.jpg\"")
		fmt.Println("   export GCS_MAX_FILE_SIZE_MB=\"10\"")
	}
	fmt.Println()

	// Example 6: Real-world microservice pattern
	fmt.Println("🏗️  Example 6: Microservice Pattern")
	fmt.Println("   Different services, different buckets, different rules:")

	// Document management service
	fmt.Println("\\n   📄 Document Management Service:")
	docService := createDocumentService()
	if docService != nil {
		testDocumentService(ctx, docService)
	}

	// User avatar service  
	fmt.Println("\\n   👤 User Avatar Service:")
	avatarService := createAvatarService()
	if avatarService != nil {
		testAvatarService(ctx, avatarService)
	}

	// Backup service
	fmt.Println("\\n   💾 Backup Service:")
	backupService := createBackupService()
	if backupService != nil {
		testBackupService(ctx, backupService)
	}

	fmt.Println()
	fmt.Println("🎉 Example completed!")
	fmt.Println()
	fmt.Println("💡 Key Takeaways:")
	fmt.Println("   • Use NewURLGeneratorWithBucket() for explicit bucket control")
	fmt.Println("   • Use NewURLGeneratorWithBucketAndRestrictions() for security")
	fmt.Println("   • Each generator instance is independent")
	fmt.Println("   • Perfect for microservices with different storage needs")
	fmt.Println("   • Environment variables provide fallback configuration")
}

// Simulated microservices

type DocumentService struct {
	generator *gcsurl.URLGenerator
}

func createDocumentService() *DocumentService {
	restrictions := &gcsurl.UploadRestrictions{
		AllowedExtensions: []string{".pdf", ".doc", ".docx", ".txt"},
		MaxFileSizeMB:     50,
		AllowMultiple:     true,
	}
	
	gen, err := gcsurl.NewURLGeneratorWithBucketAndRestrictions("company-documents", restrictions)
	if err != nil {
		log.Printf("❌ Failed to create document service: %v", err)
		return nil
	}
	
	return &DocumentService{generator: gen}
}

func testDocumentService(ctx context.Context, service *DocumentService) {
	files := []string{"contract.pdf", "image.jpg"} // Second should fail
	for _, file := range files {
		_, err := service.generator.GenerateSignedUploadURL(ctx, file)
		if err != nil {
			fmt.Printf("      ❌ %s: %v\\n", file, err)
		} else {
			fmt.Printf("      ✅ %s: Upload URL generated\\n", file)
		}
	}
}

type AvatarService struct {
	generator *gcsurl.URLGenerator
}

func createAvatarService() *AvatarService {
	restrictions := &gcsurl.UploadRestrictions{
		AllowedExtensions: []string{".jpg", ".jpeg", ".png"},
		MaxFileSizeMB:     5, // Small avatars
		AllowMultiple:     false, // One avatar at a time
	}
	
	gen, err := gcsurl.NewURLGeneratorWithBucketAndRestrictions("user-avatars", restrictions)
	if err != nil {
		log.Printf("❌ Failed to create avatar service: %v", err)
		return nil
	}
	
	return &AvatarService{generator: gen}
}

func testAvatarService(ctx context.Context, service *AvatarService) {
	files := []string{"avatar.jpg", "document.pdf"} // Second should fail
	for _, file := range files {
		_, err := service.generator.GenerateSignedUploadURL(ctx, file)
		if err != nil {
			fmt.Printf("      ❌ %s: %v\\n", file, err)
		} else {
			fmt.Printf("      ✅ %s: Upload URL generated\\n", file)
		}
	}
}

type BackupService struct {
	generator *gcsurl.URLGenerator
}

func createBackupService() *BackupService {
	// No restrictions for backup service - accepts any file type
	gen, err := gcsurl.NewURLGeneratorWithBucket("system-backups")
	if err != nil {
		log.Printf("❌ Failed to create backup service: %v", err)
		return nil
	}
	
	return &BackupService{generator: gen}
}

func testBackupService(ctx context.Context, service *BackupService) {
	files := []string{"database.sql", "config.json", "logs.tar.gz"}
	for _, file := range files {
		_, err := service.generator.GenerateSignedUploadURL(ctx, file)
		if err != nil {
			fmt.Printf("      ❌ %s: %v\\n", file, err)
		} else {
			fmt.Printf("      ✅ %s: Upload URL generated\\n", file)
		}
	}
}