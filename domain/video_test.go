package domain_test

import (
	"testing"

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
	video.ID = "1"
	video.ResourceID = "a"
	video.FilePath = "path"

	require.NoError(t, video.Validate())
}
