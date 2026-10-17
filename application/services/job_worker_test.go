package services_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/dl/encoding-video/application/services"
	"github.com/dl/encoding-video/domain"
	"github.com/stretchr/testify/require"
	"github.com/streadway/amqp"
)

type memoryVideoRepository struct {
	videos []*domain.Video
}

func (r *memoryVideoRepository) Insert(video *domain.Video) (*domain.Video, error) {
	r.videos = append(r.videos, video)
	return video, nil
}

func (r *memoryVideoRepository) Find(id string) (*domain.Video, error) {
	return nil, nil
}

// runJobWorker sends a single message to a worker and returns its result.
func runJobWorker(t *testing.T, jobService services.JobService, body string) services.JobWorkerResult {
	t.Helper()

	messages := make(chan amqp.Delivery, 1)
	results := make(chan services.JobWorkerResult, 1)

	go services.JobWorker(messages, results, jobService, domain.Job{}, 0)
	messages <- amqp.Delivery{Body: []byte(body)}
	close(messages)

	select {
	case result := <-results:
		return result
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for worker result")
		return services.JobWorkerResult{}
	}
}

func TestJobWorkerSuccess(t *testing.T) {
	videoRepo := &memoryVideoRepository{}
	jobRepo := &memoryJobRepository{}
	jobService := newTestJobService(t, jobRepo)
	jobService.VideoService.VideoRepository = videoRepo

	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})
	fakeMp4fragment(t, `touch "$2"`)
	fakeCommand(t, "mp4dash", `mkdir -p "$4" && echo segment > "$4/segment.m4s"`)

	result := runJobWorker(t, jobService, `{"resource_id": "resource-id", "file_path": "sample.mp4"}`)

	require.Nil(t, result.Error)
	require.Len(t, videoRepo.videos, 1)
	require.NotEmpty(t, result.Job.ID)
	require.Equal(t, "COMPLETED", result.Job.Status)
	require.Equal(t, "output-bucket", result.Job.OutputBucketPath)
	require.Equal(t, videoRepo.videos[0].ID, result.Job.Video.ID)
}
