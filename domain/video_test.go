package domain_test

import (
	"encoder/domain"
	"testing"\
	"github.com/stretchr/testify/require"
)

func TestValidateIfVideoIsEmpty(t *testing.T){
 video := domain.NewVideo()
	err := video.validate()

	require.Error(t, err)
}
