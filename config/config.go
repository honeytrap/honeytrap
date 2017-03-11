package config

import (
	"regexp"
	"time"

	"github.com/imdario/mergo"
	"github.com/op/go-logging"

	"github.com/BurntSushi/toml"
	"io"
	"os"
)

var log = logging.MustGetLogger("honeytrap:config")

var elapso = regexp.MustCompile(`(\d+)(\w+)`)

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

type (
	HouseKeeper struct {
		Every Delay `toml:"every"`
	}

	// TODO: rename to Timers
	Delays struct {
		PushDelay        Delay `toml:"push_every"`
		FreezeDelay      Delay `toml:"freeze_every"`
		StopDelay        Delay `toml:"stop_every"`
		HousekeeperDelay Delay `toml:"housekeeper_every"`
	}

	Console struct {
		Level string `toml:"level"`
	}

	Folders struct {
		Data string `toml:"data"`
	}

	Config struct {
		Token       string      `toml:"token"`
		Template    string      `toml:"template"`
		NetFilter   string      `toml:"net_filter"`
		Keys        string      `toml:"keys"`
		Delays      Delays      `toml:"delays"`
		Folders     Folders     `toml:"folders"`
		HouseKeeper HouseKeeper `toml:"housekeeper"`

		Channels []map[string]interface{} `toml:"channels"`

		Services []toml.Primitive `toml:"services"`

		Web   WebConfig   `toml:"web"`
		Agent AgentConfig `toml:"agent"`

		Providers []struct {
			LXC struct {
			} `toml:"lxc"`
		} `toml:"providers"`

		Logging []struct {
			Output string `toml:"output"`
			Level  string `toml:"level"`
		} `toml:"logging"`
	}
)

type WebConfig struct {
	Port string `toml:"port"`
}

type AgentConfig struct {
	Port string `toml:"port"`
	TLS  struct {
		Enabled bool `toml:"enabled"`
	} `toml:"tls"`
}

type HTTPProxyConfig struct {
	Port string `toml:"port"`
}

type SIPProxyConfig struct {
	Port string `toml:"port"`
}

type SMTPProxyConfig struct {
	Port string `toml:"port"`
	Host string `toml:"host"`
	TLS  struct {
		CertificateKey string `toml:"certificate_key"`
		Certificate    string `toml:"certificate"`
	} `toml:"tls"`
}

type Delay time.Duration

func (t *Delay) Duration() time.Duration {
	return time.Duration(*t)
}

func (t *Delay) UnmarshalText(text []byte) error {
	s := string(text)

	d, err := time.ParseDuration(s)
	if err != nil {
		log.Error("Error parsing duration (%s): %s", s, err.Error())
		return err
	}

	*t = Delay(d)
	return nil
}

var DefaultConfig = Config{
	Token:     "",
	Template:  "honeytrap",
	NetFilter: "",
	Delays: Delays{
		PushDelay:        Delay(10 * time.Second),
		FreezeDelay:      Delay(15 * time.Minute),
		StopDelay:        Delay(30 * time.Minute),
		HousekeeperDelay: Delay(1 * time.Minute),
	},
	Folders: Folders{},
	HouseKeeper: HouseKeeper{
		Every: Delay(60 * time.Second),
	},
	Web: WebConfig{
		Port: ":3000",
	},
	Agent: AgentConfig{
		Port: ":6887",
	},
}

func New() (*Config, error) {
	c := DefaultConfig
	return &c, nil
}

func (c *Config) Load(file string) error {
	conf := Config{}
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		return err
	}

	if err := mergo.MergeWithOverwrite(c, conf); err != nil {
		return err
	}

	logBackends := []logging.Backend{}
	for _, log := range conf.Logging {
		var err error

		var output io.Writer = os.Stdout

		switch log.Output {
		case "stdout":
			output = os.Stdout
		case "stderr":
			output = os.Stderr
		default:
			output, err = os.OpenFile(os.ExpandEnv(log.Output), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
		}

		if err != nil {
			panic(err)
		}

		backend := logging.NewLogBackend(output, "", 0)
		backendFormatter := logging.NewBackendFormatter(backend, format)
		backendLeveled := logging.AddModuleLevel(backendFormatter)

		level, err := logging.LogLevel(log.Level)
		if err != nil {
			panic(err)
		}

		backendLeveled.SetLevel(level, "")

		logBackends = append(logBackends, backendLeveled)
	}

	logging.SetBackend(logBackends...)

	return nil
}
