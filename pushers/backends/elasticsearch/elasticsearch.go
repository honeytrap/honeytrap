package elasticsearch

import (
	"context"

	"net/http"
	"net/url"
	"time"

	"github.com/BurntSushi/toml"
	uuid "github.com/satori/go.uuid"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/op/go-logging"
)

var (
	_ = pushers.RegisterBackend("elasticsearch", NewWith)
)

var log = logging.MustGetLogger("honeytrap:channels:elasticsearch")

// SearchConfig defines a struct which holds configuration values for a SearchBackend.
type Config struct {
	Host *url.URL

	Index string
}

func (c *Config) UnmarshalTOML(p interface{}) error {
	data, _ := p.(map[string]interface{})

	if v, ok := data["host"]; !ok {
	} else if s, ok := v.(string); !ok {
	} else if u, err := url.Parse(s); err != nil {
		return err
	} else {
		c.Host = u
		c.Index = u.Path[1:]

		// remove path
		c.Host.Path = ""
	}

	return nil
}

type tomlURL struct {
	*url.URL
}

func (d *tomlURL) UnmarshalText(text []byte) error {
	var err error
	d.URL, err = url.Parse(string(text))
	return err
}

/*
u, err := url.Parse(config.Host)
if err != nil {
	return nil, err
}
*/

// SearchBackend defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type ElasticSearchBackend struct {
	Config

	client *http.Client
	ch     chan message.Event
}

// New returns a new instance of a SearchBackend.
func New(conf Config) *ElasticSearchBackend {
	ch := make(chan message.Event, 100)

	backend := ElasticSearchBackend{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 5,
			},
			Timeout: time.Duration(20) * time.Second,
		},
		ch:     ch,
		Config: conf,
	}

	go backend.run()

	return &backend
}

// NewWith defines a function to return a pushers.Backend which delivers
// new messages to a giving underline ElasticSearch API defined by the configuration
// retrieved from the giving toml.Primitive.
func NewWith(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var config Config

	if err := meta.PrimitiveDecode(data, &config); err != nil {
		return nil, err
	}

	return New(config), nil
}

func (hc ElasticSearchBackend) run() {
	log.Debug("Indexer started...")
	defer log.Debug("Indexer stopped...")

	// TODO: reconnect with elasticsearch, don't fail
	es, err := elastic.NewClient(elastic.SetURL(hc.Host.String()), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	bulk := es.Bulk()

	count := 0
	for {
		select {
		case doc := <-hc.ch:
			messageID := uuid.NewV4()

			bulk = bulk.Add(elastic.NewBulkIndexRequest().
				Index(hc.Index).
				Type("event").
				Id(messageID.String()).
				Doc(doc),
			)

			if bulk.NumberOfActions() < 10 {
				continue
			}
		case <-time.After(time.Second * 10):
		}

		if bulk.NumberOfActions() == 0 {
		} else if response, err := bulk.Do(context.Background()); err != nil {
			log.Errorf("Error indexing: %s", err.Error())
		} else {
			indexed := response.Indexed()
			count += len(indexed)

			log.Infof("Bulk indexing: %d total %d.\n", len(indexed), count)
		}
	}
}

// Send delivers the giving push messages into the internal elastic search endpoint.
func (hc ElasticSearchBackend) Send(message message.Event) {
	hc.ch <- message
}

/*
	// buffer
	log.Debugf("ElasticSearchBackend: Sending %d actions.", message)

	buf := new(bytes.Buffer)

	if message.Sensor == "honeytrap" && message.Category == "ping" {
		// ignore
		return
	}

	if err := json.NewEncoder(buf).Encode(message.Data); err != nil {
		log.Errorf("ElasticSearchBackend: Error encoding data: %s", err.Error())
		return
	}

	messageID := uuid.NewV4()
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/%s/%s", hc.config.Host, message.Sensor, message.Category, messageID.String()), buf)
	if err != nil {
		log.Errorf("ElasticSearchBackend: Error while preparing request: %s", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	if resp, err = hc.client.Do(req); err != nil {
		log.Errorf("ElasticSearchBackend: Error while sending messages: %s", err.Error())
		return
	}

	// for keep-alive
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Errorf("ElasticSearchBackend: Unexpected status code: %d", resp.StatusCode)
		return
	}

	log.Debugf("ElasticSearchBackend: Sent %d actions.", message)
}
*/
