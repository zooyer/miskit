package buntdb

import (
	"context"
	"time"

	"github.com/tidwall/buntdb"
	"github.com/zooyer/miskit/imdb"
)

type buntConn struct {
	db *buntdb.DB
}

type buntDriver int

func (b buntDriver) Open(args string) (_ imdb.Conn, err error) {
	var c buntConn

	if c.db, err = buntdb.Open(args); err != nil {
		return
	}

	return &c, nil
}

func (c buntConn) Get(ctx context.Context, key string) (value string, err error) {
	if err = c.db.View(func(tx *buntdb.Tx) error {
		if value, err = tx.Get(key); err != nil && err != buntdb.ErrNotFound {
			return err
		}
		return nil
	}); err != nil {
		return
	}

	return
}

func (c buntConn) Set(ctx context.Context, key, value string) (err error) {
	if err = c.db.Update(func(tx *buntdb.Tx) error {
		if _, _, err = tx.Set(key, value, nil); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return
	}

	return
}

func (c buntConn) SetEx(ctx context.Context, key, value string, ttl time.Duration) (err error) {
	if err = c.db.Update(func(tx *buntdb.Tx) error {
		var options = buntdb.SetOptions{
			Expires: ttl > 0,
			TTL:     ttl,
		}
		if _, _, err = tx.Set(key, value, &options); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return
	}

	return
}

func (c buntConn) Del(ctx context.Context, key string) (err error) {
	if err = c.db.Update(func(tx *buntdb.Tx) error {
		if _, err = tx.Delete(key); err != nil && err != buntdb.ErrNotFound {
			return err
		}
		return nil
	}); err != nil {
		return
	}

	return
}

func init() {
	imdb.Register("buntdb", new(buntDriver))
}
