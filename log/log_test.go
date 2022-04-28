package log

import (
	"context"
	"encoding/json"
	"github.com/zooyer/miskit/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func writer(path, name string) (writer zapcore.WriteSyncer, err error) {
	file, err := os.Create(filepath.Join(path, name))
	if err != nil {
		return
	}

	return zapcore.AddSync(file), nil
}

func encoder(debug bool) zapcore.Encoder {
	var config = zap.NewProductionEncoderConfig()
	if debug {
		config = zap.NewDevelopmentEncoderConfig()
		//configs.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	config.EncodeTime = timeEncoder

	//return zapcore.NewMapObjectEncoder()
	return zapcore.NewConsoleEncoder(config)
}

func TestLog(t *testing.T) {
	logger := zap.NewExample()
	logger, _ = zap.NewDevelopment()
	logger, _ = zap.NewProduction()

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	config.EncoderConfig.EncodeTime = timeEncoder
	//configs.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	//w, err := writer("./", "test.log")
	//if err != nil {
	//	t.Fatal(err)
	//}

	core := zapcore.NewCore(encoder(true), zapcore.AddSync(os.Stdout), zap.DebugLevel)
	logger = zap.New(core, zap.AddCaller())
	defer logger.Sync()

	sugar := logger.Sugar()
	sugar.Infow("failed to fetch URL", "rpc", "talos", "cost", 32.3, "code", 0)
}

func TestNew(t *testing.T) {
	var config = Config{
		Filename:   "",
		Output:     "stdout",
		Level:      "DEBUG",
		Align:      false,
		Interval:   0,
		//AutoClear:  false,
		//ClearHours: 0,
		//Separate:   false,
	}
	log, err := New(config, TextFormatter(true))
	if err != nil {
		t.Fatal(err)
	}

	var ctx = context.Background()
	traceInfo := trace.New(nil, "test")
	traceInfo.Tag = "test"
	traceInfo.Lang = "zh-CN"
	traceInfo.SpanID = "this is span id"
	traceInfo.Content = json.RawMessage(`{"key": "val"}`)
	ctx = trace.Set(ctx, traceInfo)

	log.Debug(ctx, "Hello")
}
