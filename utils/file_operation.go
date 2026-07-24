package utils

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	MaxImageSize = 20 * 1024 * 1024
	UploadRoots  = "uploads/"
	AvatarDir    = "image/avatar/"
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

func SaveAvatar(file *multipart.FileHeader, username string, context *gin.Context) (*string, error) {
	// Verify uploaded image
	extension, err := verifyUploadedImage(file)
	if err != nil {
		return nil, err
	}

	// Create directory if not exists
	os.MkdirAll(UploadRoots+AvatarDir, os.ModePerm)

	// Create image name and filepath
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), username, extension)
	path := GetAvatarPath(&filename)

	// Upload the image
	err = context.SaveUploadedFile(file, path)
	if err != nil {
		return nil, err
	}

	// Return the image's path
	return &path, nil
}

func RemoveImage(filepath *string) error {
	_, err := os.Stat(*filepath)
	if err == nil {
		err = os.Remove(*filepath)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetAvatarPath(filename *string) string {
	return filepath.Join(UploadRoots, AvatarDir, *filename)
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
