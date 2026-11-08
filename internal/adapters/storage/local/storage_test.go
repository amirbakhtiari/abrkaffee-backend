package local

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestStorageSaveAndDelete(t *testing.T) {
	storage, err := NewStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	// PNG signature
	png := []byte{
		0x89, 'P', 'N', 'G',
		0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D,
	}

	key, err := storage.Save(
		context.Background(),
		bytes.NewReader(png),
		"image/png",
	)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if key == "" {
		t.Fatal("Save() returned empty key")
	}

	fullPath := filepath.Join(storage.Root(), filepath.FromSlash(key))
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("saved file does not exist: %v", err)
	}

	if err := storage.Delete(context.Background(), key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := os.Stat(fullPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file still exists after Delete(), stat error = %v", err)
	}
}

func TestStorageRejectsUnsupportedFile(t *testing.T) {
	storage, err := NewStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	_, err = storage.Save(
		context.Background(),
		bytes.NewReader([]byte("this is not an image")),
		"image/png",
	)

	if !errors.Is(err, ErrInvalidImageType) {
		t.Fatalf("Save() error = %v, want ErrInvalidImageType", err)
	}
}

func TestStorageRejectsOversizedFile(t *testing.T) {
	storage, err := NewStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	data := bytes.Repeat([]byte{'a'}, MaxImageSize+1)

	_, err = storage.Save(
		context.Background(),
		bytes.NewReader(data),
		"image/png",
	)

	if !errors.Is(err, ErrImageTooLarge) {
		t.Fatalf("Save() error = %v, want ErrImageTooLarge", err)
	}
}

func TestStorageRejectsPathTraversal(t *testing.T) {
	storage, err := NewStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Delete(context.Background(), "../outside.txt")
	if !errors.Is(err, ErrInvalidImageKey) {
		t.Fatalf("Delete() error = %v, want ErrInvalidImageKey", err)
	}
}
