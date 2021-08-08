package pool

import (
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"sync"
	"testing"
	"time"
)

type client struct {
	*ssh.Client
}

func (c *client) Ping() (err error) {
	session, err := c.NewSession()
	if err != nil {
		return err
	}
	return session.Close()
}

func newClient(network, addr string, config *ssh.ClientConfig) (entry Entry, err error) {
	c, err := ssh.Dial(network, addr, config)
	if err != nil {
		return
	}
	return &client{Client: c}, nil
}

func BenchmarkNew(b *testing.B) {
	var (
		ctx   = context.Background()
		mutex sync.Mutex
		count int
	)

	var pool *Pool
	pool = New(1, 30, time.Minute, func() (entry Entry, err error) {
		mutex.Lock()
		defer mutex.Unlock()
		count++
		if count > 30 {
			fmt.Println(pool.num, len(pool.entry))
		}
		return newClient("tcp", "127.0.0.1:22", &ssh.ClientConfig{
			Config:  ssh.Config{},
			User:    "zzy",
			Auth:    []ssh.AuthMethod{ssh.Password("386143717")},
			Timeout: time.Second * 5,
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		})
	})

	for i := 0; i < 10; i++ {
		const goroutine = 65
		start := time.Now()
		var wg sync.WaitGroup
		wg.Add(goroutine)
		for i := 0; i < goroutine; i++ {
			go func() {
				defer wg.Done()
				for i := 0; i < 5; i++ {
					entry, err := pool.Get(ctx)
					if err != nil {
						b.Fatal(err)
					}
					if err = pool.Put(entry); err != nil {
						b.Log(err)
					}
				}
			}()
		}
		wg.Wait()

		b.Log("len:", pool.Len())
		b.Log("num:", pool.num)
		b.Log("count:", count)
		entry, err := pool.Get(ctx)
		if err != nil {
			b.Fatal(err)
		}
		pool.Put(entry)
		b.Log(time.Since(start))
	}
}
