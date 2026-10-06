package domain_test

import (
	"testing"
	"time"

	"github.com/dl/encoding-video/domain"
	"github.com/stretchr/testify/require"
)

func TestValidateIfVideoIsEmpty(t *testing.T) {
	video := domain.NewVideo()
	err := video.Validate()

	require.Error(t, err)
}

func TestVideoIsValid(t *testing.T) {
	video := domain.NewVideo()
	video.ID = "7f1d5a52-7f0a-4c3e-9a55-1b2f0c9d6e11"
	video.ResourceID = "a"
	video.FilePath = "path"

	require.NoError(t, video.Validate())
}

func TestVideoIdIsNotAUuid(t *testing.T) {
	video := domain.NewVideo()

	video.ID = "abc"
	video.ResourceID = "a"
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	err := video.Validate()
	require.Error(t, err)
}
