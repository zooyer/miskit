package pool

import (
	"context"
	"sync"
	"time"
)

type Entry interface {
	Ping() error
	Close() error
}

type entry struct {
	entry Entry
	time  time.Time
}

type Pool struct {
	min   int
	max   int
	num   int
	idle  time.Duration
	new   func() (entry Entry, err error)
	entry chan entry
	mutex sync.Mutex
}

func New(min, max int, idle time.Duration, new func() (entry Entry, err error)) *Pool {
	var pool = &Pool{
		min:   min,
		max:   max,
		idle:  idle,
		new:   new,
		entry: make(chan entry, max),
	}

	return pool
}

func (p *Pool) ping(e entry) bool {
	return e.entry.Ping() == nil
}

func (p *Pool) expired(e entry) bool {
	return time.Now().After(e.time.Add(p.idle))
}

func (p *Pool) add(n int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.num += n
}

func (p *Pool) less(n int, then func()) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.num+len(p.entry) < n {
		then()
		return true
	}

	return false
}

func (p *Pool) Get(ctx context.Context) (entry Entry, err error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case e := <-p.entry:
			if p.expired(e) || !p.ping(e) {
				_ = e.entry.Close()
				continue
			}
			p.add(1)
			return e.entry, nil
		default:
			if p.less(p.max, func() {
				if entry, err = p.new(); err == nil {
					p.num++
				}
			}) {
				return
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case e := <-p.entry:
				p.add(1)
				return e.entry, nil
			}
		}
	}
}

func (p *Pool) Put(e Entry) (err error) {
	defer p.add(-1)

	if err = e.Ping(); err != nil {
		return e.Close()
	}

	select {
	case p.entry <- entry{entry: e, time: time.Now()}:
	default:
		return e.Close()
	}

	return
}

func (p *Pool) Len() int {
	return len(p.entry)
}
