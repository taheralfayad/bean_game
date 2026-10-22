package bean

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/taheralfayad/bean_game/backend/internal/blobs"
)

var (
	ErrNotFound         = errors.New("bean not found")
	ErrEmptyImage       = errors.New("image is empty")
	ErrImageTooLarge    = errors.New("image too large")
	ErrUnsupportedImage = errors.New("unsupported image type")
)

var extByMIME = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type Image struct {
	Reader io.ReadSeeker
	Size   int64
}

type beanStore interface {
	InsertBean(ctx context.Context, name string, imageKey *string) (*Bean, error)
	DeleteBean(ctx context.Context, id uint16) (string, error)
}

type Service struct {
	store    beanStore
	blobs    blobs.Storage
	maxBytes int64
}

func NewService(store beanStore, blobsClient blobs.Storage, maxBytes int64) *Service {
	return &Service{
		store:    store,
		blobs:    blobsClient,
		maxBytes: maxBytes,
	}
}

func (s *Service) Create(ctx context.Context, name string, image Image) (*Bean, error) {
	if image.Size <= 0 {
		return nil, ErrEmptyImage
	}
	if image.Size > s.maxBytes {
		return nil, ErrImageTooLarge
	}

	head := make([]byte, 512)
	n, err := io.ReadFull(image.Reader, head)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, fmt.Errorf("read image: %w", err)
	}

	contentType := http.DetectContentType(head[:n])
	ext, ok := extByMIME[contentType]
	if !ok {
		return nil, ErrUnsupportedImage
	}

	if _, err := image.Reader.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind image: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate image key: %w", err)
	}
	key := "images/" + id.String() + ext

	if err := s.blobs.Put(ctx, key, image.Reader, image.Size, contentType); err != nil {
		return nil, err
	}

	created, err := s.store.InsertBean(ctx, name, &key)
	if err != nil {
		if cleanupErr := s.blobs.Delete(ctx, key); cleanupErr != nil {
			return nil, errors.Join(err, fmt.Errorf("cleanup object: %w", cleanupErr))
		}
		return nil, err
	}

	return created, nil
}

func (s *Service) Delete(ctx context.Context, id uint16) error {
	imageKey, err := s.store.DeleteBean(ctx, id)
	if err != nil {
		return err
	}
	if imageKey == "" {
		return nil
	}
	return s.blobs.Delete(ctx, imageKey)
}
