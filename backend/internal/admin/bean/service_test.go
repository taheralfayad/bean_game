package bean

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
)

type fakeStore struct {
	insertName     string
	insertImageKey *string
	deleteID       uint16
	deleteImageKey string
	deleteCalled   bool
}

func (f *fakeStore) InsertBean(ctx context.Context, name string, imageKey *string) (*Bean, error) {
	f.insertName = name
	f.insertImageKey = imageKey
	return &Bean{Id: 7, Name: name, ImageKey: imageKey}, nil
}

func (f *fakeStore) DeleteBean(ctx context.Context, id uint16) (string, error) {
	f.deleteCalled = true
	f.deleteID = id
	return f.deleteImageKey, nil
}

type fakeBlobs struct {
	putKey         string
	putSize        int64
	putContentType string
	putData        []byte
	putCalls       int
	deletedKeys    []string
}

func (f *fakeBlobs) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	f.putCalls++
	f.putKey = key
	f.putSize = size
	f.putContentType = contentType
	f.putData = data
	return nil
}

func (f *fakeBlobs) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *fakeBlobs) Delete(ctx context.Context, key string) error {
	f.deletedKeys = append(f.deletedKeys, key)
	return nil
}

var pngData = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00, 0x00, 0x00, 0x0d}

func TestCreateBean(t *testing.T) {
	store := &fakeStore{}
	blobStore := &fakeBlobs{}
	service := NewService(store, blobStore, 1<<20)

	created, err := service.Create(context.Background(), "coffee", Image{
		Reader: bytes.NewReader(pngData),
		Size:   int64(len(pngData)),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.Id != 7 {
		t.Errorf("Id = %d, want 7", created.Id)
	}
	if created.Name != "coffee" {
		t.Errorf("Name = %q, want %q", created.Name, "coffee")
	}
	if created.ImageKey == nil {
		t.Fatal("ImageKey = nil, want key")
	}
	if store.insertName != "coffee" {
		t.Errorf("inserted name = %q, want %q", store.insertName, "coffee")
	}
	if store.insertImageKey == nil || *store.insertImageKey != *created.ImageKey {
		t.Errorf("inserted image key = %v, want %q", store.insertImageKey, *created.ImageKey)
	}

	if !regexp.MustCompile(`^images/[0-9a-f-]{36}\.png$`).MatchString(blobStore.putKey) {
		t.Errorf("blob key = %q, want images/<uuid>.png", blobStore.putKey)
	}
	if blobStore.putKey != *created.ImageKey {
		t.Errorf("blob key = %q, want %q", blobStore.putKey, *created.ImageKey)
	}
	if blobStore.putContentType != "image/png" {
		t.Errorf("content type = %q, want image/png", blobStore.putContentType)
	}
	if blobStore.putSize != int64(len(pngData)) {
		t.Errorf("size = %d, want %d", blobStore.putSize, len(pngData))
	}
	if !bytes.Equal(blobStore.putData, pngData) {
		t.Errorf("stored data = %v, want %v", blobStore.putData, pngData)
	}
}

func TestCreateBeanSupportedImageTypes(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		wantExt     string
		wantContent string
	}{
		{
			name:        "jpeg",
			data:        []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01},
			wantExt:     ".jpg",
			wantContent: "image/jpeg",
		},
		{
			name:        "gif",
			data:        []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00"),
			wantExt:     ".gif",
			wantContent: "image/gif",
		},
		{
			name:        "webp",
			data:        []byte("RIFF\x1a\x00\x00\x00WEBPVP8 \x0e\x00\x00\x00"),
			wantExt:     ".webp",
			wantContent: "image/webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{}
			blobStore := &fakeBlobs{}
			service := NewService(store, blobStore, 1<<20)

			created, err := service.Create(context.Background(), "bean", Image{
				Reader: bytes.NewReader(tt.data),
				Size:   int64(len(tt.data)),
			})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			if !strings.HasSuffix(blobStore.putKey, tt.wantExt) {
				t.Errorf("blob key = %q, want suffix %q", blobStore.putKey, tt.wantExt)
			}
			if blobStore.putKey != *created.ImageKey {
				t.Errorf("blob key = %q, want %q", blobStore.putKey, *created.ImageKey)
			}
			if blobStore.putContentType != tt.wantContent {
				t.Errorf("content type = %q, want %q", blobStore.putContentType, tt.wantContent)
			}
		})
	}
}

func TestDeleteBeanWithImage(t *testing.T) {
	store := &fakeStore{deleteImageKey: "images/abc.png"}
	blobStore := &fakeBlobs{}
	service := NewService(store, blobStore, 1<<20)

	if err := service.Delete(context.Background(), 42); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if !store.deleteCalled {
		t.Fatal("DeleteBean was not called")
	}
	if store.deleteID != 42 {
		t.Errorf("deleted id = %d, want 42", store.deleteID)
	}
	if len(blobStore.deletedKeys) != 1 || blobStore.deletedKeys[0] != "images/abc.png" {
		t.Errorf("deleted blob keys = %v, want [images/abc.png]", blobStore.deletedKeys)
	}
}

func TestDeleteBeanWithoutImage(t *testing.T) {
	store := &fakeStore{}
	blobStore := &fakeBlobs{}
	service := NewService(store, blobStore, 1<<20)

	if err := service.Delete(context.Background(), 42); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if !store.deleteCalled {
		t.Fatal("DeleteBean was not called")
	}
	if len(blobStore.deletedKeys) != 0 {
		t.Errorf("deleted blob keys = %v, want none", blobStore.deletedKeys)
	}
}
