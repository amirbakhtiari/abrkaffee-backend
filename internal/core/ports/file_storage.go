package ports

import (
	"context"
	"errors"
	"io"
)

var (
	ErrFileTooLarge        = errors.New("file exceeds maximum allowed size")
	ErrUnsupportedFileType = errors.New("unsupported file type")
)

type FileStorage interface {
	Save(ctx context.Context, file io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
}
