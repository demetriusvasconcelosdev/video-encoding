package services_test

import (
	"testing"

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
