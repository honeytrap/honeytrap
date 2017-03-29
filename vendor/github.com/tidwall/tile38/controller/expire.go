package controller

import (
	"time"

	"github.com/tidwall/resp"
	"github.com/tidwall/tile38/controller/log"
	"github.com/tidwall/tile38/controller/server"
)

// clearAllExpires removes all items that are marked at expires.
func (c *Controller) clearAllExpires() {
	c.expires = make(map[string]map[string]time.Time)
}

// clearIDExpires will clear a single item from the expires list.
func (c *Controller) clearIDExpires(key, id string) int {
	m := c.expires[key]
	if m == nil {
		return 0
	}
	delete(m, id)
	if len(m) == 0 {
		delete(c.expires, key)
	}
	return 1
}

// clearKeyExpires will clear all items that are marked as expires from a single key.
func (c *Controller) clearKeyExpires(key string) {
	delete(c.expires, key)
}

// expireAt will mark an item as expires at a specific time.
func (c *Controller) expireAt(key, id string, at time.Time) {
	m := c.expires[key]
	if m == nil {
		m = make(map[string]time.Time)
		c.expires[key] = m
	}
	m[id] = at
}

// getExpires will return the when the item expires.
func (c *Controller) getExpires(key, id string) (at time.Time, ok bool) {
	m := c.expires[key]
	if m == nil {
		ok = false
		return
	}
	at, ok = m[id]
	return
}

// backgroundExpiring watches for when items must expire from the database.
// It's runs through every item that has been marked as expires five times
// per second.
func (c *Controller) backgroundExpiring() {
	const stop = 0
	const delay = 1
	const nodelay = 2
	for {
		op := func() int {
			c.mu.RLock()
			defer c.mu.RUnlock()
			if c.stopBackgroundExpiring {
				return stop
			}
			// Only excute for leaders. Followers should ignore.
			if c.config.FollowHost == "" {
				now := time.Now()
				for key, m := range c.expires {
					for id, at := range m {
						if now.After(at) {
							// issue a DEL command
							c.mu.RUnlock()
							c.mu.Lock()

							// double check because locks were swapped
							var del bool
							if m2, ok := c.expires[key]; ok {
								if at2, ok := m2[id]; ok {
									if now.After(at2) {
										del = true
									}
								}
							}
							if !del {
								return nodelay
							}
							c.statsExpired++
							msg := &server.Message{}
							msg.Values = resp.MultiBulkValue("del", key, id).Array()
							msg.Command = "del"
							_, d, err := c.cmdDel(msg)
							if err != nil {
								c.mu.Unlock()
								log.Fatal(err)
								continue
							}
							if err := c.writeAOF(resp.ArrayValue(msg.Values), &d); err != nil {
								c.mu.Unlock()
								log.Fatal(err)
								continue
							}
							c.mu.Unlock()
							c.mu.RLock()
							return nodelay
						}
					}
				}
			}
			return delay
		}()
		switch op {
		case stop:
			return
		case delay:
			time.Sleep(time.Millisecond * 100)
		case nodelay:
			time.Sleep(time.Microsecond)
		}
	}
}
