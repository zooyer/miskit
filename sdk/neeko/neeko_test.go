package neeko

import (
	"context"
	"testing"
)

func TestClient_SystemInfo(t *testing.T) {
	client := New()
	info, err := client.SystemInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}
