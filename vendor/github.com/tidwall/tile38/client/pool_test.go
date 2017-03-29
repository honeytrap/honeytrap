package client

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	pool, err := DialPool("localhost:9876")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var wg sync.WaitGroup
	wg.Add(25)
	for i := 0; i < 25; i++ {
		go func(i int) {
			defer func() {
				wg.Done()
			}()
			conn, err := pool.Get()
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			msg, err := conn.Do("PING")
			if err != nil {
				t.Fatal(err)
			}
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(msg), &m); err != nil {
				t.Fatal(err)
			}
			if ok1, ok2 := m["ok"].(bool); !ok1 || !ok2 {
				t.Fatal("not ok")
			}
			if pong, ok := m["ping"].(string); !ok || pong != "pong" {
				t.Fatal("not pong")
			}
			defer conn.Do(fmt.Sprintf("drop test:%d", i))
			msg, err = conn.Do(fmt.Sprintf("drop test:%d", i))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(msg), `{"ok":true`) {
				t.Fatal("expecting OK:TRUE response")
			}
			for j := 0; j < 100; j++ {
				lat, lon := rand.Float64()*180-90, rand.Float64()*360-180
				msg, err = conn.Do(fmt.Sprintf("set test:%d %d point %f %f", i, j, lat, lon))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(string(msg), `{"ok":true`) {
					t.Fatal("expecting OK:TRUE response")
				}
			}
		}(i)
	}
	wg.Wait()
}
