package repositories_test

import (
	"testing"
	"time"

	"github.com/dl/encoding-video/application/repositories"
	"github.com/dl/encoding-video/domain"
	"github.com/dl/encoding-video/framework/database"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestVideoRepositoryDbInsert(t *testing.T) {
	db := database.NewDbTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.ResourceID = "resource-id"
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repo := repositories.NewVideoRepository(db)
	_, err := repo.Insert(video)
	require.Nil(t, err)

	v, err := repo.Find(video.ID)
	require.Nil(t, err)
	require.Equal(t, video.ID, v.ID)
	require.Equal(t, video.ResourceID, v.ResourceID)
	require.Equal(t, video.FilePath, v.FilePath)
}
