package ice

import (
	"context"
	"github.com/zooyer/miskit/log"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var (
		config = log.Config{
			Level: "DEBUG",
		}
		stdout    = log.NewStdoutRecorder(log.TextFormatter(true))
		logger, _ = log.New(config, nil)
	)
	logger.SetDefaultRecorder(stdout)

	var option = Option{
		AppID:     622660,
		AppSecret: "d566d332f9c78479baba83a6af7f6c45",
		Addr:      "192.168.1.10:8804",
		Retry:     1,
		Logger:    logger,
		Timeout:   time.Second,
	}
	ice := New(option)

	for i := 0; i < 20; i++ {
		id, err := ice.GenID(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		t.Log(id)
	}
}
