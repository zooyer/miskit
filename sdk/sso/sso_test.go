package sso

import (
	"context"
	"testing"
	"time"

	"github.com/zooyer/miskit/log"
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
		AppID:     6625877950,
		AppSecret: "ed9935282d224f062b7aa2e088434d7a",
		Addr:      "127.0.0.1:8800",
		Retry:     1,
		Logger:    logger,
		Timeout:   time.Second,
	}
	sso := New(option)

	resp, err := sso.Validate(context.Background(), "f543b6a0-98cf-411b-9b46-552139abbdc00")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}
