package services_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/dl/encoding-video/application/services"
	"github.com/dl/encoding-video/domain"
	"github.com/stretchr/testify/require"
)

type memoryJobRepository struct {
	statuses []string
	failWith error
}

func (r *memoryJobRepository) Insert(job *domain.Job) (*domain.Job, error) {
	return job, nil
}

func (r *memoryJobRepository) Find(id string) (*domain.Job, error) {
	return nil, errors.New("not implemented")
}

func (r *memoryJobRepository) Update(job *domain.Job) (*domain.Job, error) {
	if r.failWith != nil {
		return nil, r.failWith
	}
	r.statuses = append(r.statuses, job.Status)
	return job, nil
}

func newTestJobService(t *testing.T, repo *memoryJobRepository) services.JobService {
	t.Helper()

	videoService, _ := newTestVideoService(t)
	t.Setenv("inputBucketName", "input-bucket")
	t.Setenv("outputBucketName", "output-bucket")
	t.Setenv("CONCURRENCY_UPLOAD", "2")

	job, err := domain.NewJob("output-bucket", "STARTING", videoService.Video)
	require.Nil(t, err)

	return services.JobService{
		Job:           job,
		JobRepository: repo,
		VideoService:  videoService,
	}
}

func TestJobServiceStart(t *testing.T) {
	repo := &memoryJobRepository{}
	jobService := newTestJobService(t, repo)

	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})
	fakeMp4fragment(t, `touch "$2"`)
	fakeCommand(t, "mp4dash", `mkdir -p "$4" && echo segment > "$4/segment.m4s"`)

	err := jobService.Start()

	require.Nil(t, err)
	require.Equal(t, []string{
		"DOWNLOADING", "FRAGMENTING", "ENCODING", "UPLOADING", "FINISHING", "COMPLETED",
	}, repo.statuses)
}
