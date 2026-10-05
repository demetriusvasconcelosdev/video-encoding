package encoder

import (
	"testing"

	"github.com/dl/encoding-video/internal/config"
)

func TestRun(t *testing.T) {
	if err := Run(config.Config{}); err != nil {
		t.Fatal(err)
	}
}
