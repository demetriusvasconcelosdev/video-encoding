package repositories_test

import (
	"testing"
	"time"

	"github.com/dl/encoding-video/application/repositories"
	"github.com/dl/encoding-video/domain"
	"github.com/dl/encoding-video/framework/database"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func newTestJob(t *testing.T, db *gorm.DB) *domain.Job {
	t.Helper()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.ResourceID = "resource-id"
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	_, err := repositories.NewVideoRepository(db).Insert(video)
	require.Nil(t, err)

	job, err := domain.NewJob("output_path", "Pending", video)
	require.Nil(t, err)
	job.VideoID = video.ID

	return job
}

func TestJobRepositoryDbInsert(t *testing.T) {
	db := database.NewDbTest()
	defer db.Close()

	job := newTestJob(t, db)

	repo := repositories.NewJobRepository(db)
	_, err := repo.Insert(job)
	require.Nil(t, err)

	j, err := repo.Find(job.ID)
	require.Nil(t, err)
	require.Equal(t, job.ID, j.ID)
	require.Equal(t, job.VideoID, j.VideoID)
	require.Equal(t, job.OutputBucketPath, j.OutputBucketPath)
	require.Equal(t, job.Status, j.Status)
	require.NotNil(t, j.Video)
	require.Equal(t, job.VideoID, j.Video.ID)
}

func TestJobRepositoryDbFindNotFound(t *testing.T) {
	db := database.NewDbTest()
	defer db.Close()

	repo := repositories.NewJobRepository(db)

	j, err := repo.Find(uuid.NewV4().String())
	require.Nil(t, j)
	require.EqualError(t, err, "job does not exist")
}

func TestJobRepositoryDbUpdate(t *testing.T) {
	db := database.NewDbTest()
	defer db.Close()

	job := newTestJob(t, db)

	repo := repositories.NewJobRepository(db)
	_, err := repo.Insert(job)
	require.Nil(t, err)

	job.Status = "Complete"
	job.Error = "some error"
	_, err = repo.Update(job)
	require.Nil(t, err)

	j, err := repo.Find(job.ID)
	require.Nil(t, err)
	require.Equal(t, "Complete", j.Status)
	require.Equal(t, "some error", j.Error)
	require.Equal(t, job.ID, j.ID)
}
