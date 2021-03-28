package ssh

import (
	"context"
	"fmt"
	"github.com/pkg/sftp"
	"github.com/zooyer/miskit/pool"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"sync"
	"time"
)

type Remote struct {
	Addr     string
	Username string
	Password string
}

type PoolOption struct {
	MaxConn int
	MinConn int
}

type Pool struct {
	min    int
	max    int
	idle   time.Duration
	conn   map[string]chan *ssh.Client
	remote map[string]Remote
	errors []error

	wg    sync.WaitGroup
	mutex sync.Mutex
	close chan struct{}
}

func NewPool(min, max int, idle time.Duration) *Pool {
	var pool = Pool{
		min:    min,
		max:    max,
		idle:   idle,
		conn:   make(map[string]chan *ssh.Client),
		remote: make(map[string]Remote),
		close:  make(chan struct{}, 10),
	}

	//go pool.loop()

	return &pool
}

func NewPool2(min, max int, idle time.Duration) *Pool2 {
	return &Pool2{
		min:  min,
		max:  max,
		idle: idle,
		pool: make(map[string]*pool.Pool),
	}
}

func (p *Pool) Add(addr, user, password string) {
	p.remote[addr+user+password] = Remote{
		Addr:     addr,
		Username: user,
		Password: password,
	}
}

func (p *Pool) initOne(addr, user, password string) {
	var key = p.key(addr, user, password)
	if p.conn[key] == nil {
		p.conn[key] = make(chan *ssh.Client, p.max+p.min)
	}

	start := time.Now()
	if client, err := Client(user, password, addr); err == nil {
		p.conn[key] <- client
	}
	fmt.Println("ssh connect:", user, addr, time.Since(start))

	return
}

func (p *Pool) init(addr, user, password string) {
	var key = p.key(addr, user, password)
	if p.conn[key] == nil {
		p.conn[key] = make(chan *ssh.Client, p.max+p.min)
	}

	var wg sync.WaitGroup
	var count = p.min - len(p.conn[key])
	wg.Add(count)
	start := time.Now()
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			if client, err := Client(user, password, addr); err == nil {
				p.conn[key] <- client
			}
		}()
	}
	wg.Wait()
	fmt.Println("ssh connect:", user, addr, time.Since(start))

	return
}

func (p *Pool) key(addr, user, password string) string {
	return fmt.Sprintf("%s@%s:%s", user, addr, password)
}

func (p *Pool) getConn(addr, user, password string) (*ssh.Client, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var key = p.key(addr, user, password)
	if len(p.conn[key]) == 0 {
		p.init(addr, user, password)
	}
	if len(p.conn[key]) == 0 {
		return nil, fmt.Errorf("%s@%s no connection available", user, addr)
	}

	return <-p.conn[key], nil
}

func (p *Pool) putConn(client *ssh.Client, addr, user, password string) {
	if client == nil {
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var key = p.key(addr, user, password)

	p.conn[key] <- client
}

func (p *Pool) Session(addr, user, password string) (session *ssh.Session, err error) {
	client, err := p.getConn(addr, user, password)
	if err != nil {
		return
	}
	defer p.putConn(client, addr, user, password)

	return client.NewSession()
}

func (p *Pool) SftpClient(addr, user, password string) (client *sftp.Client, err error) {
	sshClient, err := p.getConn(addr, user, password)
	if err != nil {
		return
	}
	defer p.putConn(sshClient, addr, user, password)

	return sftp.NewClient(sshClient)
}

func (p *Pool) ScpReader(reader io.Reader, remote, password string, fn func(size int)) (err error) {
	user, addr, filename, err := parse(remote)
	if err != nil {
		return
	}

	client, err := p.SftpClient(addr, user, password)
	if err != nil {
		return
	}
	defer client.Close()

	return ScpReader(client, filename, newReader(reader, fn))
}

func (p *Pool) Scp(local, remote, password string, fn func(current, total int64)) (err error) {
	file, err := os.Open(local)
	if err != nil {
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return
	}

	var (
		total   = stat.Size()
		current int64
	)

	return p.ScpReader(file, remote, password, func(size int) {
		current += int64(size)
		fn(current, total)
	})
}

func (p *Pool) Command(remote, password, cmd string) (output string, err error) {
	user, addr, _, err := parse(remote)
	if err != nil {
		return
	}

	session, err := p.Session(addr, user, password)
	if err != nil {
		return
	}
	defer session.Close()

	return CommandSession(session, cmd)
}

func (p *Pool) loop() {
	for range p.close {
		for _, remote := range p.remote {
			p.mutex.Lock()
			p.init(remote.Addr, remote.Username, remote.Password)
			p.mutex.Unlock()
		}
		time.Sleep(time.Second)
	}
}

func (p *Pool) Init(remote ...Remote) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, remote := range remote {
		p.init(remote.Addr, remote.Username, remote.Password)
	}
}

func (p *Pool) Close() error {
	p.close <- struct{}{}
	p.wg.Wait()
	var err error
	for _, err := range p.errors {
		if err != nil {
			return err
		}
	}
	return err
}

type client struct {
	*ssh.Client
}

func (c *client) Ping() error {
	session, err := c.NewSession()
	if err != nil {
		return err
	}
	return session.Close()
}

type Pool2 struct {
	min   int
	max   int
	idle  time.Duration
	pool  map[string]*pool.Pool
	mutex sync.Mutex
}

func (p *Pool2) key(addr, username, password string) string {
	return fmt.Sprintf("%s@%s:%s", username, addr, password)
}

func (p *Pool2) get(addr, username, password string) (*ssh.Client, error) {
	var key = p.key(addr, username, password)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.pool[key] == nil {
		var factory = func() (entry pool.Entry, err error) {
			cli, err := Client(username, password, addr)
			if err != nil {
				return
			}
			return &client{Client: cli}, nil
		}

		p.pool[key] = pool.New(p.min, p.max, p.idle, factory)
	}

	var ctx = context.Background()

	cli, err := p.pool[key].Get(ctx)
	if err != nil {
		return nil, err
	}

	return cli.(*client).Client, nil
}

func (p *Pool2) put(client *client, addr, username, password string) (err error) {
	var key = p.key(addr, username, password)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err = p.pool[key].Put(client); err != nil {
		return
	}

	return
}

func (p *Pool2) Session(addr, username, password string) (session *ssh.Session, err error) {
	cli, err := p.get(addr, username, password)
	if err != nil {
		return
	}
	defer p.put(&client{Client: cli}, addr, username, password)

	return cli.NewSession()
}

func (p *Pool2) SftpClient(addr, username, password string) (*sftp.Client, error) {
	cli, err := p.get(addr, username, password)
	if err != nil {
		return nil, err
	}
	defer p.put(&client{Client: cli}, addr, username, password)

	return sftp.NewClient(cli)
}

func (p *Pool2) ScpReader(reader io.Reader, remote, password string, fn func(size int)) (err error) {
	user, addr, filename, err := parse(remote)
	if err != nil {
		return
	}

	client, err := p.SftpClient(addr, user, password)
	if err != nil {
		return
	}
	defer client.Close()

	return ScpReader(client, filename, newReader(reader, fn))
}

func (p *Pool2) Scp(local, remote, password string, fn func(current, total int64)) (err error) {
	file, err := os.Open(local)
	if err != nil {
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return
	}

	var (
		total   = stat.Size()
		current int64
	)

	return p.ScpReader(file, remote, password, func(size int) {
		current += int64(size)
		fn(current, total)
	})
}

func (p *Pool2) Command(remote, password, cmd string) (output string, err error) {
	user, addr, _, err := parse(remote)
	if err != nil {
		return
	}

	session, err := p.Session(addr, user, password)
	if err != nil {
		return
	}
	defer session.Close()

	return CommandSession(session, cmd)
}
