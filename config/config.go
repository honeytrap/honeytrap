package config

import (
	"regexp"
	"time"

	"github.com/imdario/mergo"

	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/process"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:config")

var elapso = regexp.MustCompile(`(\d+)(\w+)`)

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

type (

	// HouseKeeper defines the settings for operation cleanup.
	HouseKeeper struct {
		Every Delay `toml:"every"`
	}

	// LxcConfig defines the settings for the lxc director.
	LxcConfig struct {
		Commands []process.Command       `toml:"commands"`
		Scripts  []process.ScriptProcess `toml:"scripts"`
	}

	// IOConfig defines the settings for the iodirector.
	IOConfig struct {
		ServiceAddr string                  `toml:"service_addr"`
		Commands    []process.Command       `toml:"commands"`
		Scripts     []process.ScriptProcess `toml:"scripts"`
	}

	// CowrieConfig defines the settings for director meta.
	CowrieConfig struct {
		SSHPort  string                  `toml:"ssh_port"`
		SSHAddr  string                  `toml:"ssh_addr"`
		Commands []process.Command       `toml:"commands"`
		Scripts  []process.ScriptProcess `toml:"scripts"`
	}

	// DirectorConfig defines the settings for all directors supported
	// by honeytrap.
	DirectorConfig struct {
		IOConfig IOConfig                `toml:"io_config"`
		Cowrie   CowrieConfig            `toml:"cowrie_config"`
		LXC      LxcConfig               `toml:"lxc_config"`
		Commands []process.Command       `toml:"commands"`
		Scripts  []process.ScriptProcess `toml:"scripts"`
	}

	// Delays sets the individual duration set for all ops.
	// TODO: rename to Timers
	Delays struct {
		PushDelay        Delay `toml:"push_every"`
		FreezeDelay      Delay `toml:"freeze_every"`
		StopDelay        Delay `toml:"stop_every"`
		HousekeeperDelay Delay `toml:"housekeeper_every"`
	}

	// Console defines the struct to contain the console logging level.
	Console struct {
		Level string `toml:"level"`
	}

	// Folders defines the data path for usage in container ops.
	Folders struct {
		Data string `toml:"data"`
	}

	// Config defines the central type where all configuration is umarhsalled to.
	Config struct {
		Token       string         `toml:"token"`
		Template    string         `toml:"template"`
		NetFilter   string         `toml:"net_filter"`
		Keys        string         `toml:"keys"`
		Director    string         `toml:"director"`
		Delays      Delays         `toml:"delays"`
		Folders     Folders        `toml:"folders"`
		HouseKeeper HouseKeeper    `toml:"housekeeper"`
		Directors   DirectorConfig `toml:"directors"`

		Backends map[string]interface{}   `toml:"backends"`
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

// WebConfig defines the configuration for the web access point.
type WebConfig struct {
	Port string `toml:"port"`
	Path string `toml:"path"`
}

// AgentConfig defines configuration for the agent server.
type AgentConfig struct {
	Port string `toml:"port"`
	TLS  struct {
		Enabled bool `toml:"enabled"`
	} `toml:"tls"`
}

// HTTPProxyConfig defines the config type for the http proxy server.
type HTTPProxyConfig struct {
	Port string `toml:"port"`
}

// SIPProxyConfig defines the configuration struct for the sip server.
type SIPProxyConfig struct {
	Port string `toml:"port"`
}

// SMTPProxyConfig defines the configuration for the SMTPProxyConfig.
type SMTPProxyConfig struct {
	Port string `toml:"port"`
	Host string `toml:"host"`
	TLS  struct {
		CertificateKey string `toml:"certificate_key"`
		Certificate    string `toml:"certificate"`
	} `toml:"tls"`
}

// Delay defines a duration type.
type Delay time.Duration

// Duration returns the type of the giving duration from the provided pointer.
func (t *Delay) Duration() time.Duration {
	return time.Duration(*t)
}

// UnmarshalText handles unmarshalling duration values from the provided slice.
func (t *Delay) UnmarshalText(text []byte) error {
	s := string(text)

	d, err := time.ParseDuration(s)
	if err != nil {
		log.Errorf("Error parsing duration (%s): %s", s, err.Error())
		return err
	}

	*t = Delay(d)
	return nil
}

// DefaultConfig defines the default Config to be used to set default values.
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
	Directors: DirectorConfig{
		Cowrie: CowrieConfig{
			SSHPort: "2222",
		},
	},
	Web: WebConfig{
		Port: ":3000",
		Path: "",
	},
	Agent: AgentConfig{
		Port: ":6887",
	},
}

// New returns a new instance of the config struct.
func New() (*Config, error) {
	c := DefaultConfig
	return &c, nil
}

// Load attempts to load the giving toml configuration file.
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

		var output io.Writer

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
