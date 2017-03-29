package client

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const dialTimeout = time.Second * 3
const pingTimeout = time.Second

// Pool represents a pool of tile38 connections.
type Pool struct {
	mu     sync.Mutex
	conns  []*Conn
	addr   string
	closed bool
}

// DialPool creates a new pool with 5 initial connections to the specified tile38 server.
func DialPool(addr string) (*Pool, error) {
	pool := &Pool{
		addr: addr,
	}
	// create some connections. 5 is a good start
	var tconns []*Conn
	for i := 0; i < 5; i++ {
		conn, err := pool.Get()
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("unable to fill pool: %s", err)
		}
		tconns = append(tconns, conn)
	}
	pool.conns = tconns
	return pool, nil
}

// Close releases the resources used by the pool.
func (pool *Pool) Close() error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if pool.closed {
		return errors.New("pool closed")
	}
	pool.closed = true
	for _, conn := range pool.conns {
		conn.pool = nil
		conn.Close()
	}
	pool.conns = nil
	return nil
}

// Get borrows a connection. When the connection closes, the application returns it to the pool.
func (pool *Pool) Get() (*Conn, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	for len(pool.conns) != 0 {
		i := rand.Int() % len(pool.conns)
		conn := pool.conns[i]
		pool.conns = append(pool.conns[:i], pool.conns[i+1:]...)
		// Ping to test on borrow.
		conn.SetDeadline(time.Now().Add(pingTimeout))
		if _, err := conn.Do("PING"); err != nil {
			conn.pool = nil
			conn.Close()
			continue
		}
		conn.SetDeadline(time.Time{})
		return conn, nil
	}
	conn, err := DialTimeout(pool.addr, dialTimeout)
	if err != nil {
		return nil, err
	}
	conn.pool = pool
	return conn, nil
}

func (pool *Pool) put(conn *Conn) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if pool.closed {
		return errors.New("pool closed")
	}
	conn.SetDeadline(time.Time{})
	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})
	pool.conns = append(pool.conns, conn)
	return nil
}
