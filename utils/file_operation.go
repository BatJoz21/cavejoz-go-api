package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	MaxImageSize    = 20 * 1024 * 1024
	UploadRoots     = "uploads/"
	AvatarDir       = "avatar/"
	ImageContentDir = "content/image/"
)

var imageExtension = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

var imageMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

// randomFileStem returns an unguessable, filename-safe name with no caller
// input in it, so an upload path can never be steered outside its directory.
func randomFileStem() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func SaveAvatar(file *multipart.FileHeader, context *gin.Context) (*string, error) {
	// Verify uploaded image
	extension, err := verifyUploadedImage(file)
	if err != nil {
		return nil, err
	}

	// Create directory if not exists
	if err := os.MkdirAll(UploadRoots+AvatarDir, 0o755); err != nil {
		return nil, err
	}

	// Create image name and filepath
	stem, err := randomFileStem()
	if err != nil {
		return nil, err
	}
	filename := stem + extension
	path := GetAvatarPath(&filename)

	// Upload the image
	err = context.SaveUploadedFile(file, path)
	if err != nil {
		return nil, err
	}

	// Return the image's file name
	return &filename, nil
}

func SavePostContent(file *multipart.FileHeader, context *gin.Context) (*string, error) {
	// Verify uploaded image
	extension, err := verifyUploadedImage(file)
	if err != nil {
		return nil, err
	}

	// Create directory if not exists
	if err := os.MkdirAll(UploadRoots+ImageContentDir, 0o755); err != nil {
		return nil, err
	}

	// Create image name and filepath
	stem, err := randomFileStem()
	if err != nil {
		return nil, err
	}
	filename := stem + extension
	path := GetImageContentPath(&filename)

	// Upload the image
	err = context.SaveUploadedFile(file, path)
	if err != nil {
		return nil, err
	}

	// Return the image's file name
	return &filename, nil
}

func RemoveImage(filename *string, mode string) error {
	// Get filepath
	var filepath string
	switch mode {
	case "profile":
		filepath = GetAvatarPath(filename)
	case "content":
		filepath = GetImageContentPath(filename)
	}

	// Find the file
	_, err := os.Stat(filepath)
	if err == nil {
		// Remove the file
		err = os.Remove(filepath)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetAvatarPath(filename *string) string {
	return filepath.Join(UploadRoots, AvatarDir, *filename)
}

func GetDefaultAvatar() string {
	return filepath.Join("assets", "default_user.png")
}

func GetImageContentPath(filename *string) string {
	return filepath.Join(UploadRoots, ImageContentDir, *filename)
}

func verifyUploadedImage(file *multipart.FileHeader) (string, error) {
	// Checking image's size
	if file.Size > MaxImageSize {
		return "", errors.New("Image's size is too large")
	}

	// Checking image's extension
	extension := strings.ToLower(filepath.Ext(file.Filename))
	if !imageExtension[extension] {
		return "", errors.New("Image's type is not allowed")
	}

	// Open the file
	openFile, err := file.Open()
	if err != nil {
		return "", err
	}
	defer openFile.Close()

	// Checking MIME type
	buffer := make([]byte, 512)
	n, err := openFile.Read(buffer)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buffer[:n])
	if !imageMimeTypes[contentType] {
		return "", errors.New("Invalid content type: " + contentType)
	}

	// Return image's extension if success
	return extension, nil
}
