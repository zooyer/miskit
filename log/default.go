package log

import "context"

var std *Logger

func init() {
	var config = Config{
		Level: "DEBUG",
	}
	stdout := NewStdoutRecorder(TextFormatter(true))
	stderr := NewStderrRecorder(TextFormatter(true))
	std, _ = New(config, nil)
	std.SetDefaultRecorder(stdout)
	std.SetRecorder("WARNING", stderr)
	std.SetRecorder("ERROR", stderr)
	std.SetRecorder("FATAL", stderr)
}

func Debug(v ...interface{}) {
	std.Debug(context.Background(), v...)
}

func Info(v ...interface{}) {
	std.Info(context.Background(), v...)
}

func Warning(v ...interface{}) {
	std.Warning(context.Background(), v...)
}

func Error(v ...interface{}) {
	std.Error(context.Background(), v...)
}

func Fatal(v ...interface{}) {
	std.Fatal(context.Background(), v...)
}

func D(v ...interface{}) {
	std.Debug(context.Background(), v...)
}

func I(v ...interface{}) {
	std.Info(context.Background(), v...)
}

func W(v ...interface{}) {
	std.Warning(context.Background(), v...)
}

func E(v ...interface{}) {
	std.Error(context.Background(), v...)
}

func F(v ...interface{}) {
	std.Fatal(context.Background(), v...)
}
