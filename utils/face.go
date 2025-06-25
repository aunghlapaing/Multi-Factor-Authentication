package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FaceData represents stored face data
type FaceData struct {
	UserID    int    `json:"user_id"`
	FaceImage string `json:"face_image"` // Base64 encoded image
}

const faceDataDir = "./data/faces"

// InitFaceStorage initializes the face data storage directory
func InitFaceStorage() error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(faceDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create face data directory: %v", err)
	}
	return nil
}

// SaveFaceData saves face data for a user
func SaveFaceData(userID int, faceImageBase64 string) error {
	// Clean the base64 string if it contains data URL prefix
	if strings.HasPrefix(faceImageBase64, "data:image") {
		parts := strings.Split(faceImageBase64, ",")
		if len(parts) == 2 {
			faceImageBase64 = parts[1]
		}
	}

	// Create face data object
	faceData := FaceData{
		UserID:    userID,
		FaceImage: faceImageBase64,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(faceData)
	if err != nil {
		return fmt.Errorf("failed to marshal face data: %v", err)
	}

	// Save to file
	filename := filepath.Join(faceDataDir, fmt.Sprintf("user_%d.json", userID))
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write face data file: %v", err)
	}

	return nil
}

// GetFaceData retrieves face data for a user
func GetFaceData(userID int) (*FaceData, error) {
	filename := filepath.Join(faceDataDir, fmt.Sprintf("user_%d.json", userID))
	
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("no face data found for user %d", userID)
	}

	// Read file
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read face data file: %v", err)
	}

	// Parse JSON
	var faceData FaceData
	if err := json.Unmarshal(jsonData, &faceData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal face data: %v", err)
	}

	return &faceData, nil
}

// HasFaceData checks if a user has face data stored
func HasFaceData(userID int) bool {
	filename := filepath.Join(faceDataDir, fmt.Sprintf("user_%d.json", userID))
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// DeleteFaceData deletes face data for a user
func DeleteFaceData(userID int) error {
	filename := filepath.Join(faceDataDir, fmt.Sprintf("user_%d.json", userID))
	
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil // Already doesn't exist
	}

	// Delete file
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete face data file: %v", err)
	}

	return nil
}
