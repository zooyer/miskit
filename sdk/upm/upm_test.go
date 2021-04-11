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
		AppID:     19511,
		AppSecret: "e41e3df18ff53f208aa1789c61261f20",
		Addr:      "127.0.0.1:8805",
		Retry:     1,
		Logger:    logger,
		Timeout:   time.Second,
	}
	upm := New(option)

	cond := map[string]interface{}{
		"path":   "/upm/v1/test",
		"method": "GET",
	}

	auth, args, err := upm.Auth(context.Background(), 996, cond)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(auth, args.JSONString())
}
