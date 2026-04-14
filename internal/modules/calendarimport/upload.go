package calendarimport

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func BindImportICSCommand(c *gin.Context, userID uuid.UUID) (ImportICSCommand, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxICSFileSizeBytes)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return ImportICSCommand{}, mapUploadError(err)
	}
	if fileHeader.Size > MaxICSFileSizeBytes {
		return ImportICSCommand{}, fmt.Errorf("file is too large")
	}

	content, err := readUploadedFile(fileHeader)
	if err != nil {
		return ImportICSCommand{}, mapUploadError(err)
	}

	return ImportICSCommand{
		UserID:           userID,
		OriginalFilename: fileHeader.Filename,
		Content:          content,
	}, nil
}

func readUploadedFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file")
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func mapUploadError(err error) error {
	if err == nil {
		return nil
	}

	message := err.Error()
	switch {
	case strings.Contains(message, "request body too large"):
		return fmt.Errorf("file is too large")
	case strings.Contains(message, "no such file"), strings.Contains(message, "missing form body"), strings.Contains(message, "http: no such file"):
		return fmt.Errorf("file is required")
	default:
		return err
	}
}
