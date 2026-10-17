package services_test

import (
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

func TestJobWorkerInvalidJson(t *testing.T) {
	result := runJobWorker(t, services.JobService{}, "not a json")

	require.Error(t, result.Error)
	require.Empty(t, result.Job.ID)
	require.Equal(t, "not a json", string(result.Message.Body))
}

func TestJobWorkerInvalidVideo(t *testing.T) {
	result := runJobWorker(t, services.JobService{}, `{"resource_id": "", "file_path": ""}`)

	require.Error(t, result.Error)
	require.Empty(t, result.Job.ID)
}
