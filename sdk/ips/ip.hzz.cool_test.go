package ips

import (
	"context"
	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/zrpc"
	"reflect"
	"testing"
	"time"
)

func TestClient_QueryIP(t *testing.T) {
	type fields struct {
		url    string
		option Option
		client *zrpc.Client
	}
	type args struct {
		ctx context.Context
		ip  string
	}
	var (
		config = log.Config{
			Level: "DEBUG",
		}
		stdout    = log.NewStdoutRecorder(log.TextFormatter(true))
		logger, _ = log.New(config, nil)
	)
	logger.SetDefaultRecorder(stdout)
	tests := []struct {
		name    string
		args    args
		want    *IP
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx: context.Background(),
				ip:  "111.204.182.91",
			},
			want: &IP{
				IP:       "111.204.182.91",
				Country:  "中国",
				Province: "北京",
				City:     "北京",
				County:   "",
				Region:   "亚洲",
				ISP:      "联通",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(Option{
				Retry:   1,
				Logger:  logger,
				Timeout: time.Second,
			})
			got, err := c.QueryIP(tt.args.ctx, tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("QueryIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}
