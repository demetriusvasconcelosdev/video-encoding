package services_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dl/encoding-video/application/services"
	"github.com/dl/encoding-video/domain"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func newTestVideoService(t *testing.T) (services.VideoService, string) {
	t.Helper()

	storagePath := t.TempDir()
	t.Setenv("localStoragePath", storagePath)

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.ResourceID = "resource-id"
	video.FilePath = "videos/sample.mp4"

	service := services.NewVideoService()
	service.Video = video

	return service, storagePath
}

// fakeStorage points the GCS client to a local HTTP server through STORAGE_EMULATOR_HOST.
func fakeStorage(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	t.Setenv("STORAGE_EMULATOR_HOST", server.URL)
}

// fakeMp4fragment puts an executable called mp4fragment first in PATH.
func fakeMp4fragment(t *testing.T, script string) {
	t.Helper()

	binDir := t.TempDir()
	err := os.WriteFile(filepath.Join(binDir, "mp4fragment"), []byte("#!/bin/sh\n"+script), 0o755)
	require.Nil(t, err)

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestVideoServiceDownload(t *testing.T) {
	service, storagePath := newTestVideoService(t)

	var requestedPath string
	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Write([]byte("video content"))
	})

	err := service.Download("bucket-test")
	require.Nil(t, err)
	require.True(t, strings.Contains(requestedPath, "bucket-test"))
	require.True(t, strings.Contains(requestedPath, "sample.mp4"))

	content, err := os.ReadFile(filepath.Join(storagePath, service.Video.ID+".mp4"))
	require.Nil(t, err)
	require.Equal(t, "video content", string(content))
}

func TestVideoServiceDownloadObjectNotFound(t *testing.T) {
	service, storagePath := newTestVideoService(t)

	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	err := service.Download("bucket-test")
	require.NotNil(t, err)

	_, err = os.Stat(filepath.Join(storagePath, service.Video.ID+".mp4"))
	require.True(t, os.IsNotExist(err))
}

func TestVideoServiceDownloadInvalidStoragePath(t *testing.T) {
	service, _ := newTestVideoService(t)
	t.Setenv("localStoragePath", filepath.Join(t.TempDir(), "missing-dir"))

	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("video content"))
	})

	err := service.Download("bucket-test")
	require.NotNil(t, err)
}

func TestVideoServiceFragment(t *testing.T) {
	service, storagePath := newTestVideoService(t)

	// Records the arguments and creates the target file, like the real binary.
	fakeMp4fragment(t, `echo "$1 $2" > "$2"`)

	err := service.Fragment()
	require.Nil(t, err)

	info, err := os.Stat(filepath.Join(storagePath, service.Video.ID))
	require.Nil(t, err)
	require.True(t, info.IsDir())

	source := storagePath + "/" + service.Video.ID + ".mp4"
	target := storagePath + "/" + service.Video.ID + ".frag"

	args, err := os.ReadFile(target)
	require.Nil(t, err)
	require.Equal(t, source+" "+target+"\n", string(args))
}

func TestVideoServiceFragmentCommandFails(t *testing.T) {
	service, _ := newTestVideoService(t)

	fakeMp4fragment(t, "echo 'cannot open input' >&2\nexit 1")

	err := service.Fragment()
	require.NotNil(t, err)
}

func TestVideoServiceFragmentDirectoryAlreadyExists(t *testing.T) {
	service, storagePath := newTestVideoService(t)

	err := os.Mkdir(filepath.Join(storagePath, service.Video.ID), os.ModePerm)
	require.Nil(t, err)

	fakeMp4fragment(t, "exit 0")

	err = service.Fragment()
	require.NotNil(t, err)
	require.True(t, os.IsExist(err))
}
