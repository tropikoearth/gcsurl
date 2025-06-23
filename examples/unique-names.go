package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tropikoearth/gcsurl"
)

func main() {
	fmt.Println("🔗 GCS URL Generator - Unique Names Example")
	fmt.Println("===========================================")
	fmt.Println()

	ctx := context.Background()

	// Create generator
	generator, err := gcsurl.NewURLGeneratorWithBucket("documents-bucket")
	if err != nil {
		log.Fatalf("❌ Failed to create URL generator: %v", err)
	}

	// Example 1: Basic unique name generation
	fmt.Println("📁 Example 1: Basic Unique Name Generation")
	
	testFiles := []string{
		"document.pdf",
		"photo.jpg",
		"presentation.pptx",
	}
	
	for _, filename := range testFiles {
		upload, err := generator.GenerateSignedUploadURL(ctx, filename)
		if err != nil {
			log.Printf("❌ Failed to generate URL for %s: %v", filename, err)
			continue
		}
		
		fmt.Printf("📄 Original: %s\\n", upload.OriginalName)
		fmt.Printf("🔑 Generated: %s\\n", upload.GeneratedKey)
		fmt.Printf("🔗 URL: %s\\n", upload.UploadURL[:60]+"...")
		fmt.Printf("⏰ Expires: %s\\n", upload.ExpiresAt.Format("15:04:05"))
		fmt.Println()
	}

	// Example 2: Directory structure preservation
	fmt.Println("📂 Example 2: Directory Structure Preservation")
	
	nestedFiles := []string{
		"users/123/avatar.jpg",
		"reports/2025/january.xlsx",
		"projects/abc/documents/specification.pdf",
		"media/videos/intro.mp4",
	}
	
	for _, filename := range nestedFiles {
		upload, err := generator.GenerateSignedUploadURL(ctx, filename)
		if err != nil {
			log.Printf("❌ Failed to generate URL for %s: %v", filename, err)
			continue
		}
		
		fmt.Printf("📁 Input:     %s\\n", filename)
		fmt.Printf("🗂️  Generated: %s\\n", upload.GeneratedKey)
		
		// Show how directory is preserved
		originalDir := getDirectory(upload.OriginalName)
		generatedDir := getDirectory(upload.GeneratedKey)
		if originalDir == generatedDir {
			fmt.Printf("✅ Directory preserved: %s\\n", originalDir)
		} else {
			fmt.Printf("❌ Directory mismatch: %s → %s\\n", originalDir, generatedDir)
		}
		fmt.Println()
	}

	// Example 3: Simulating database storage
	fmt.Println("💾 Example 3: Database Storage Simulation")
	
	type Document struct {
		ID           string `json:"id"`
		FileKey      string `json:"fileKey"`      // Save this in database
		OriginalName string `json:"originalName"` // For user display
		BucketName   string `json:"bucketName"`
	}
	
	var documents []Document
	
	uploadFiles := []string{
		"contracts/supplier-agreement.pdf",
		"invoices/2025-001.pdf", 
		"photos/team-building.jpg",
	}
	
	for i, filename := range uploadFiles {
		upload, err := generator.GenerateSignedUploadURL(ctx, filename)
		if err != nil {
			log.Printf("❌ Failed to generate URL for %s: %v", filename, err)
			continue
		}
		
		// Simulate saving to database
		doc := Document{
			ID:           fmt.Sprintf("doc-%d", i+1),
			FileKey:      upload.GeneratedKey,  // Unique key for GCS operations
			OriginalName: upload.OriginalName,  // For showing to user
			BucketName:   generator.GetBucketName(),
		}
		documents = append(documents, doc)
		
		fmt.Printf("💾 Saved to DB:\\n")
		fmt.Printf("   ID: %s\\n", doc.ID)
		fmt.Printf("   FileKey: %s\\n", doc.FileKey)
		fmt.Printf("   OriginalName: %s\\n", doc.OriginalName)
		fmt.Printf("   Upload URL: %s\\n", upload.UploadURL[:50]+"...")
		fmt.Println()
	}

	// Example 4: Download using stored keys
	fmt.Println("📥 Example 4: Download Using Stored Keys")
	
	for _, doc := range documents {
		// Simulate getting document from database and generating download URL
		downloadURL, err := generator.GenerateSignedDownloadURL(ctx, doc.FileKey)
		if err != nil {
			log.Printf("❌ Failed to generate download URL for %s: %v", doc.ID, err)
			continue
		}
		
		fmt.Printf("📄 Document: %s (ID: %s)\\n", doc.OriginalName, doc.ID)
		fmt.Printf("🗂️  File Key: %s\\n", doc.FileKey)
		fmt.Printf("📥 Download URL: %s\\n", downloadURL[:50]+"...")
		fmt.Println()
	}

	// Example 5: Demonstrating collision avoidance
	fmt.Println("🚫 Example 5: Collision Avoidance")
	fmt.Println("   Uploading same filename multiple times:")
	
	duplicateFilename := "important-document.pdf"
	
	for i := 1; i <= 3; i++ {
		upload, err := generator.GenerateSignedUploadURL(ctx, duplicateFilename)
		if err != nil {
			log.Printf("❌ Failed to generate URL for attempt %d: %v", i, err)
			continue
		}
		
		fmt.Printf("🔄 Attempt %d:\\n", i)
		fmt.Printf("   Original: %s\\n", upload.OriginalName)
		fmt.Printf("   Generated: %s\\n", upload.GeneratedKey)
		fmt.Printf("   Unique ID: %s\\n", getUniqueID(upload.GeneratedKey))
		fmt.Println()
	}

	// Example 6: Using original names when needed
	fmt.Println("📝 Example 6: Using Original Names (No UUID)")
	fmt.Println("   Use case: Overwriting existing files")
	
	originalNameFile := "config/settings.json"
	
	// Method 1: With UUID (default)
	uploadWithUUID, err := generator.GenerateSignedUploadURL(ctx, originalNameFile)
	if err != nil {
		log.Printf("❌ Failed to generate UUID URL: %v", err)
	} else {
		fmt.Printf("🆔 With UUID: %s\\n", uploadWithUUID.GeneratedKey)
	}
	
	// Method 2: Without UUID (for overwrites)
	uploadOriginal, err := generator.GenerateSignedUploadURLWithOriginalName(ctx, originalNameFile)
	if err != nil {
		log.Printf("❌ Failed to generate original URL: %v", err)
	} else {
		fmt.Printf("📄 Original: %s\\n", uploadOriginal.GeneratedKey)
	}
	
	fmt.Println()
	fmt.Println("🎉 Example completed!")
	fmt.Println()
	fmt.Println("💡 Key Benefits:")
	fmt.Println("   • Zero file name collisions")
	fmt.Println("   • Directory structure preserved")
	fmt.Println("   • Original names available for UI")
	fmt.Println("   • Database-friendly unique keys")
	fmt.Println("   • Option to use original names when needed")
}

// Helper functions
func getDirectory(path string) string {
	if lastSlash := lastIndex(path, "/"); lastSlash != -1 {
		return path[:lastSlash]
	}
	return "." // No directory
}

func getUniqueID(generatedKey string) string {
	if lastSlash := lastIndex(generatedKey, "/"); lastSlash != -1 {
		filename := generatedKey[lastSlash+1:]
		if underscore := index(filename, "_"); underscore != -1 {
			return filename[:underscore]
		}
		return "no-uuid"
	}
	if underscore := index(generatedKey, "_"); underscore != -1 {
		return generatedKey[:underscore]
	}
	return "no-uuid"
}

func lastIndex(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func index(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}