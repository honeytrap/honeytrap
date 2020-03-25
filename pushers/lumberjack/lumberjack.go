package lumberjack

import (
	"crypto/tls"
	"github.com/cenkalti/backoff/v4"
	lumberClient "github.com/elastic/go-lumber/client/v2"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
	"net"
	"time"
)

var (
	_ = pushers.Register("lumberjack", New)
)

var log = logging.MustGetLogger("channels/lumberjack")

type Backend struct {
	Config
	ch chan event.Event
}

func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan event.Event, 100)

	c := Backend{
		ch: ch,
	}

	for _, optionFn := range options {
		_ = optionFn(&c)
	}

	go c.run()

	return &c, nil
}

func (b Backend) run() {
	var s []interface{}

	bo := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         time.Minute,
		MaxElapsedTime:      0,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	_ = backoff.RetryNotify(func() error {
		conn, err := net.Dial("tcp", b.URL)
		if err != nil {
			return err
		}

		if b.Secure {
			host, _, _ := net.SplitHostPort(b.URL)
			tlsConn := tls.Client(conn, &tls.Config{
				ServerName: host,
			})

			if err := tlsConn.Handshake(); err != nil {
				return err
			}

			conn = tlsConn
		}

		// Create new lumberjack client for protocol encoding
		client, err := lumberClient.NewWithConn(conn, lumberClient.CompressionLevel(b.CompressionLevel))
		if err != nil {
			return err
		}

		// Create new Synchronous client for sending Events to ingestion point.
		cl, err := lumberClient.NewSyncClientWith(client)
		if err != nil {
			return err
		}

		for {
			// Collect Honeypot data and save it to an interface slice.
			select {
			case evt, _ := <-b.ch:
				category := evt.Get("category")

				//Ignore heartbeats. These won't be sent to the endpoint.
				if category == "heartbeat" {
					continue
				}

				evt.Store("@metadata", map[string]interface{}{
					"beat": b.Index,
				})

				evt.Store("beat", map[string]interface{}{
					"name": "honeytrap",
				})

				s = append(s, evt)

				if len(s) < 10 {
					continue
				}

			// Send event slice to ingestion point.
			case <-time.After(time.Second * time.Duration(b.Interval)):
			}

			c, err := cl.Send(s)
			if err != nil {
				return err
			}

			if len(s) != c {
				log.Errorf("not all messages were sent. Received: %d, Transmitted %d", len(s), c)
			}

			s = s[:0]
		}
	}, bo, func(err error, duration time.Duration) {
		log.Errorf("Error %s, retrying in %v", err.Error(), duration)
	})
}

func (b Backend) Send(message event.Event) {
	select {
	case b.ch <- message:
	default:
		log.Errorf("Could not send more messages, channel full")
	}
}
