package cache

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestSetClientAndClient(t *testing.T) {
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer func() {
		_ = client.Close()
		mini.Close()
		SetClient(nil)
	}()

	SetClient(client)
	if Client() != client {
		t.Fatal("expected Client() to return the configured client")
	}
}
