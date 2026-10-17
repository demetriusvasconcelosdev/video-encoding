package services_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/dl/encoding-video/application/services"
	"github.com/stretchr/testify/require"
)

func TestNewVideoUpload(t *testing.T) {
	upload := services.NewVideoUpload()

	require.NotNil(t, upload)
	require.Empty(t, upload.Paths)
	require.Empty(t, upload.VideoPath)
	require.Empty(t, upload.OutputBucket)
	require.Empty(t, upload.Errors)
}

func writeFiles(t *testing.T, root string, names ...string) {
	t.Helper()

	for _, name := range names {
		path := filepath.Join(root, name)
		require.Nil(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.Nil(t, os.WriteFile(path, []byte("content of "+name), 0o644))
	}
}

func TestVideoUploadUploadObject(t *testing.T) {
	storagePath := t.TempDir()
	t.Setenv("localStoragePath", storagePath)
	writeFiles(t, storagePath, "video/segment.ts")

	var mu sync.Mutex
	var requests []string
	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	require.Nil(t, err)
	defer client.Close()

	upload := services.NewVideoUpload()
	upload.OutputBucket = "test-bucket"

	err = upload.UploadObject(filepath.Join(storagePath, "video/segment.ts"), client, ctx)

	require.Nil(t, err)
	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, requests)
	for _, path := range requests {
		require.Contains(t, path, "test-bucket")
	}
}

func TestVideoUploadUploadObjectFileNotFound(t *testing.T) {
	storagePath := t.TempDir()
	t.Setenv("localStoragePath", storagePath)

	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	})

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	require.Nil(t, err)
	defer client.Close()

	upload := services.NewVideoUpload()
	upload.OutputBucket = "test-bucket"

	err = upload.UploadObject(filepath.Join(storagePath, "missing.ts"), client, ctx)

	require.NotNil(t, err)
	require.True(t, os.IsNotExist(err))
}
