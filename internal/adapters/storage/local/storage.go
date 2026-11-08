package local

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)

const (
	MaxImageSize = 5 << 20 // 5 MiB
)

var (
	ErrInvalidImageType = ports.ErrUnsupportedFileType
	ErrImageTooLarge    = ports.ErrFileTooLarge
	ErrInvalidImageKey  = errors.New("invalid image key")
)

type Storage struct {
	root string
}

func NewStorage(root string) (*Storage, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("storage root is required")
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve storage root: %w", err)
	}

	if err := os.MkdirAll(absoluteRoot, 0755); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}

	return &Storage{root: absoluteRoot}, nil
}

func (s *Storage) Root() string {
	return s.root
}

func (s *Storage) Save(
	ctx context.Context,
	file io.Reader,
	_ string,
) (string, error) {
	if file == nil {
		return "", ErrInvalidImageType
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	data, err := io.ReadAll(io.LimitReader(file, MaxImageSize+1))
	if err != nil {
		return "", fmt.Errorf("read image: %w", err)
	}

	if len(data) == 0 {
		return "", ErrInvalidImageType
	}

	if len(data) > MaxImageSize {
		return "", ErrImageTooLarge
	}

	contentType := http.DetectContentType(data)

	var extension string
	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	default:
		return "", ErrInvalidImageType
	}

	if !validImageSignature(data, contentType) {
		return "", ErrInvalidImageType
	}

	randomName, err := randomFilename()
	if err != nil {
		return "", fmt.Errorf("generate image name: %w", err)
	}

	relativeKey := filepath.ToSlash(filepath.Join(
		"products",
		randomName+extension,
	))

	fullPath := filepath.Join(s.root, filepath.FromSlash(relativeKey))

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("create image directory: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("write image: %w", err)
	}

	return relativeKey, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	cleanKey := filepath.Clean(filepath.FromSlash(key))

	if key == "" ||
		filepath.IsAbs(cleanKey) ||
		cleanKey == "." ||
		cleanKey == ".." ||
		strings.HasPrefix(cleanKey, ".."+string(os.PathSeparator)) {
		return ErrInvalidImageKey
	}

	fullPath := filepath.Join(s.root, cleanKey)

	relative, err := filepath.Rel(s.root, fullPath)
	if err != nil {
		return ErrInvalidImageKey
	}

	if relative == ".." ||
		strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return ErrInvalidImageKey
	}

	err = os.Remove(fullPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete image: %w", err)
	}

	return nil
}

func randomFilename() (string, error) {
	var value [16]byte

	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(value[:]), nil
}

func validImageSignature(data []byte, contentType string) bool {
	switch contentType {
	case "image/jpeg":
		return len(data) >= 3 &&
			data[0] == 0xFF &&
			data[1] == 0xD8 &&
			data[2] == 0xFF

	case "image/png":
		return len(data) >= 8 &&
			string(data[:8]) == "\x89PNG\r\n\x1a\n"

	case "image/webp":
		return len(data) >= 12 &&
			string(data[:4]) == "RIFF" &&
			string(data[8:12]) == "WEBP"

	default:
		return false
	}
}
