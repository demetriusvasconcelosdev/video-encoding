package services_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

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

func TestVideoUploadUploadObjectStorageError(t *testing.T) {
	storagePath := t.TempDir()
	t.Setenv("localStoragePath", storagePath)
	writeFiles(t, storagePath, "segment.ts")

	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	require.Nil(t, err)
	defer client.Close()

	upload := services.NewVideoUpload()
	upload.OutputBucket = "test-bucket"

	err = upload.UploadObject(filepath.Join(storagePath, "segment.ts"), client, ctx)

	require.NotNil(t, err)
}

func TestVideoUploadProcessUpload(t *testing.T) {
	storagePath := t.TempDir()
	t.Setenv("localStoragePath", storagePath)
	writeFiles(t, storagePath, "video/a.ts", "video/b.ts", "video/nested/c.ts")

	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})

	upload := services.NewVideoUpload()
	upload.VideoPath = filepath.Join(storagePath, "video")
	upload.OutputBucket = "test-bucket"

	done := make(chan string)
	errCh := make(chan error, 1)
	go func() { errCh <- upload.ProcessUpload(1, done) }()

	select {
	case msg := <-done:
		require.Equal(t, "upload completed", msg)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for upload to finish")
	}

	require.Nil(t, <-errCh)
	require.Len(t, upload.Paths, 3)
	require.Empty(t, upload.Errors)
}

func TestVideoUploadProcessUploadInvalidVideoPath(t *testing.T) {
	upload := services.NewVideoUpload()
	upload.VideoPath = filepath.Join(t.TempDir(), "does-not-exist")

	err := upload.ProcessUpload(1, make(chan string))

	require.NotNil(t, err)
}

func TestVideoUploadProcessUploadConcurrent(t *testing.T) {
	storagePath := t.TempDir()
	t.Setenv("localStoragePath", storagePath)
	writeFiles(t, storagePath, "video/a.ts", "video/b.ts", "video/c.ts", "video/d.ts", "video/e.ts")

	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})

	upload := services.NewVideoUpload()
	upload.VideoPath = filepath.Join(storagePath, "video")
	upload.OutputBucket = "test-bucket"

	done := make(chan string)
	errCh := make(chan error, 1)
	go func() { errCh <- upload.ProcessUpload(3, done) }()

	select {
	case msg := <-done:
		require.Equal(t, "upload completed", msg)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for upload to finish")
	}

	require.Nil(t, <-errCh)
	require.Len(t, upload.Paths, 5)
	require.Empty(t, upload.Errors)
}
