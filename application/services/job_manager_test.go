package services_test

import (
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/dl/encoding-video/application/services"
	"github.com/dl/encoding-video/domain"
	"github.com/dl/encoding-video/framework/database"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

type notification struct {
	message    string
	exchange   string
	routingKey string
}

type fakeNotifier struct {
	notifications chan notification
}

func newFakeNotifier() *fakeNotifier {
	return &fakeNotifier{notifications: make(chan notification, 10)}
}

func (n *fakeNotifier) Notify(message string, contentType string, exchange string, routingKey string) error {
	n.notifications <- notification{message: message, exchange: exchange, routingKey: routingKey}
	return nil
}

type fakeAcknowledger struct {
	mu       sync.Mutex
	acks     int
	rejected int
}

func (a *fakeAcknowledger) Ack(tag uint64, multiple bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.acks++
	return nil
}

func (a *fakeAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error { return nil }

func (a *fakeAcknowledger) Reject(tag uint64, requeue bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rejected++
	return nil
}

func startJobManager(t *testing.T, notifier services.Notifier) (chan amqp.Delivery, func()) {
	t.Helper()

	t.Setenv("CONCURRENCY_WORKERS", "1")
	t.Setenv("RABBITMQ_NOTIFICATION_EX", "amq.direct")
	t.Setenv("RABBITMQ_NOTIFICATION_ROUTING_KEY", "jobs")

	messages := make(chan amqp.Delivery)
	results := make(chan services.JobWorkerResult)

	manager := services.NewJobManager(database.NewDbTest(), nil, results, messages)
	manager.RabbitMQ = notifier

	finished := make(chan struct{})
	go func() {
		manager.Start(nil)
		close(finished)
	}()

	stop := func() {
		close(messages)
		close(results)
		select {
		case <-finished:
		case <-time.After(5 * time.Second):
			t.Fatal("job manager did not stop")
		}
	}

	return messages, stop
}

func TestJobManagerNotifiesParseError(t *testing.T) {
	notifier := newFakeNotifier()
	messages, stop := startJobManager(t, notifier)
	defer stop()

	acker := &fakeAcknowledger{}
	messages <- amqp.Delivery{Acknowledger: acker, Body: []byte("not a json")}

	select {
	case n := <-notifier.notifications:
		var payload services.JobNotificationError
		require.Nil(t, json.Unmarshal([]byte(n.message), &payload))
		require.Equal(t, "not a json", payload.Message)
		require.NotEmpty(t, payload.Error)
		require.Equal(t, "amq.direct", n.exchange)
		require.Equal(t, "jobs", n.routingKey)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for notification")
	}
}

func TestJobManagerNotifiesSuccessAndAcks(t *testing.T) {
	storagePath := t.TempDir()
	t.Setenv("localStoragePath", storagePath)
	t.Setenv("inputBucketName", "input-bucket")
	t.Setenv("outputBucketName", "output-bucket")
	t.Setenv("CONCURRENCY_UPLOAD", "2")

	fakeStorage(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})
	fakeMp4fragment(t, `touch "$2"`)
	fakeCommand(t, "mp4dash", `mkdir -p "$4" && echo segment > "$4/segment.m4s"`)

	notifier := newFakeNotifier()
	messages, stop := startJobManager(t, notifier)
	defer stop()

	acker := &fakeAcknowledger{}
	messages <- amqp.Delivery{Acknowledger: acker, Body: []byte(`{"resource_id": "resource-id", "file_path": "sample.mp4"}`)}

	select {
	case n := <-notifier.notifications:
		var job domain.Job
		require.Nil(t, json.Unmarshal([]byte(n.message), &job))
		require.Equal(t, "COMPLETED", job.Status)
		require.Equal(t, "output-bucket", job.OutputBucketPath)
		require.Equal(t, "resource-id", job.Video.ResourceID)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for notification")
	}

	require.Eventually(t, func() bool {
		acker.mu.Lock()
		defer acker.mu.Unlock()
		return acker.acks == 1 && acker.rejected == 0
	}, 5*time.Second, 10*time.Millisecond)
}
