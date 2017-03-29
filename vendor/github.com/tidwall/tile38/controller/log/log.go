package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Default is the default log
var Default = New(os.Stderr, &Config{})

const (
	clear   = "\x1b[0m"
	bright  = "\x1b[1m"
	dim     = "\x1b[2m"
	black   = "\x1b[30m"
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
	white   = "\x1b[37m"
)

// Config is the log configuration
type Config struct {
	HideInfo   bool
	HideTime   bool
	HideNotice bool
	HideWarn   bool
	HideDebug  bool
	HideError  bool
	HideFatal  bool
	HideHTTP   bool
	NoColors   bool
}

// Log is log
type Log struct {
	mu  sync.RWMutex
	w   io.Writer
	ib  []byte
	ob  []byte
	cfg *Config
	l   *log.Logger
	st  time.Time
	tth time.Duration
}

// New creates a new Log and outputs to w.
func New(w io.Writer, cfg *Config) *Log {
	if cfg == nil {
		cfg = &Config{}
	}
	lc := &Log{w: w, cfg: cfg}
	lc.l = log.New(lc, "", log.LstdFlags)
	lc.st = time.Now()
	return lc
}

func (w *Log) format(b []byte) []byte {
	s := string(b)
	if strings.Contains(s, "!RESET_TIME!") {
		w.st = time.Now()
		return nil
	}
	var useRed bool
	if strings.Contains(s, "!RED!") {
		useRed = true
		s = strings.Replace(s, "!RED!", "", 1)
	}

	si := strings.Index(s, "[")
	if si != -1 {
		ei := strings.Index(s[si+1:], "]")
		if ei != -1 {
			tag := s[si+1 : ei+si+1]
			format, otag, color, hide := true, "", clear, false
			switch tag {
			case "TIME":
				otag, color, hide = "[TIME]", magenta, w.cfg.HideTime
			case "INFO":
				otag, color, hide = "[INFO]", cyan, w.cfg.HideInfo
			case "HTTP":
				otag, color, hide = "[HTTP]", bright+black, w.cfg.HideHTTP
			case "FATAL", "FATA":
				otag, color, hide = "[FATA]", red, w.cfg.HideFatal
			case "ERR", "ERRO", "ERROR":
				otag, color, hide = "[ERRO]", bright+red, w.cfg.HideError
			case "WARN":
				otag, color, hide = "[WARN]", yellow, w.cfg.HideWarn
			case "NOTI":
				otag, color, hide = "[NOTI]", bright, w.cfg.HideInfo
			case "DEBUG", "DEBU":
				otag, color, hide = "[DEBU]", magenta, w.cfg.HideDebug
			default:
				format = false
			}
			if format {
				if hide {
					s = ""
				} else if w.cfg.NoColors || color == clear {
					s = s[:si] + otag + s[ei+si+2:]
				} else {
					if useRed {
						color = red
					}
					s = s[:si] + color + otag + clear + s[ei+si+2:]
				}
			}
			if otag == "[TIME]" {
				st := time.Now()
				df := st.Sub(w.st)
				if df < w.tth {
					return nil
				}
				w.st = st
				var str string
				if w.cfg.NoColors {
					str = fmt.Sprintf(" %s", df)
				} else {
					str = fmt.Sprintf(" %s%s%s", bright+black, df, clear)
				}
				s = s[:len(s)-1] + str + s[len(s)-1:]
			}
			if !w.cfg.NoColors {
				s = strings.Replace(s, "[Leader]", green+"[Leader]"+clear, -1)
				s = strings.Replace(s, "[Follower]", red+"[Follower]"+clear, -1)
				s = strings.Replace(s, "[Candidate]", yellow+"[Candidate]"+clear, -1)
			}
		}
	}
	return []byte(s)
}

// Write writes data directly to the log
func (w *Log) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ib = append(w.ib, p...)
	for {
		idx := bytes.Index(w.ib, []byte{'\n'})
		if idx == -1 {
			break
		}
		w.ob = append(w.ob, w.format(w.ib[:idx+1])...)
		w.ib = w.ib[idx+1:]
	}
	if len(w.ob) > 0 {
		n, err := w.w.Write(w.ob)
		if err != nil {
			return len(p), err
		}
		w.ob = w.ob[n:]
	}
	return len(p), nil
}

func expand(v []interface{}) string {
	var b bytes.Buffer
	for i, v := range v {
		if i > 0 {
			b.WriteByte(' ')
		}
		switch v := v.(type) {
		default:
			b.WriteString(fmt.Sprintf("%v", v))
		case error:
			b.WriteString(v.Error())
		case fmt.Stringer:
			b.WriteString(v.String())
		}
	}
	return b.String()
}

// Print is equivlent to Info
func (w *Log) Print(v ...interface{}) {
	w.Info(expand(v))
}

// Printf is equivlent to Infof
func (w *Log) Printf(format string, args ...interface{}) {
	w.Infof(format, args...)
}

// Info prints variables with [INFO] tag
func (w *Log) Info(v ...interface{}) {
	w.l.Printf("[INFO] %s", expand(v))
}

// Infof prints format [INFO] tag
func (w *Log) Infof(format string, args ...interface{}) {
	w.Info(fmt.Sprintf(format, args...))
}

// Notice prints variables [NOTI] tag
func (w *Log) Notice(v ...interface{}) {
	w.l.Printf("[NOTI] %s", expand(v))
}

// Noticef prints format [NOTI] tag
func (w *Log) Noticef(format string, args ...interface{}) {
	w.Notice(fmt.Sprintf(format, args...))
}

// Warn prints variables [WARN] tag
func (w *Log) Warn(v ...interface{}) {
	w.l.Printf("[WARN] %s", expand(v))
}

// Warnf prints format [WARN] tag
func (w *Log) Warnf(format string, args ...interface{}) {
	w.Warn(fmt.Sprintf(format, args...))
}

// Debug prints variables [DEBU] tag
func (w *Log) Debug(v ...interface{}) {
	w.l.Printf("[DEBU] %s", expand(v))
}

// Debugf prints format [DEBU] tag
func (w *Log) Debugf(format string, args ...interface{}) {
	w.Debug(fmt.Sprintf(format, args...))
}

// Error prints variables [ERRO] tag
func (w *Log) Error(v ...interface{}) {
	w.l.Printf("[ERRO] %s", expand(v))
}

// Errorf prints format [ERRO] tag
func (w *Log) Errorf(format string, args ...interface{}) {
	w.Error(fmt.Sprintf(format, args...))
}

// Fatal prints variables [FATA] tag followed by an os.Exit(-1).
func (w *Log) Fatal(v ...interface{}) {
	w.l.Printf("[FATA] %s", expand(v))
	os.Exit(-1)
}

// Fatalf prints format [FATA] tag followed by an os.Exit(-1).
func (w *Log) Fatalf(format string, args ...interface{}) {
	w.Fatal(fmt.Sprintf(format, args...))
}

// HTTP prints variables [HTTP] tag
func (w *Log) HTTP(v ...interface{}) {
	w.l.Printf("[HTTP] %s", expand(v))
}

// HTTPf prints format [HTTP] tag
func (w *Log) HTTPf(format string, args ...interface{}) {
	w.HTTP(fmt.Sprintf(format, args...))
}

// Time prints variables [TIME] tag
func (w *Log) Time(v ...interface{}) {
	w.l.Printf("[TIME] %s", expand(v))
}

// Timef prints format [TIME] tag
func (w *Log) Timef(format string, args ...interface{}) {
	w.Time(fmt.Sprintf(format, args...))
}

// ResetTime reset the start time
func (w *Log) ResetTime() {
	w.mu.Lock()
	w.st = time.Now()
	w.mu.Unlock()
}

// TimeMinimum sets the minimum duration before the time elapsed text appears
func (w *Log) TimeMinimum(min time.Duration) {
	w.mu.Lock()
	w.tth = min
	w.mu.Unlock()
}

// ResetTime reset the start time
func ResetTime() { Default.ResetTime() }

// TimeMinimum sets the minimum duration before the time elapsed text appears
func TimeMinimum(min time.Duration) { Default.TimeMinimum(min) }

// Info prints variables [INFO] tag
func Info(v ...interface{}) { Default.Info(v...) }

// Infof prints format [INFO] tag
func Infof(format string, args ...interface{}) { Default.Infof(format, args...) }

// Notice prints variables [NOTI] tag
func Notice(v ...interface{}) { Default.Notice(v...) }

// Noticef prints format [NOTI] tag
func Noticef(format string, args ...interface{}) { Default.Noticef(format, args...) }

// Warn prints variables [WARN] tag
func Warn(v ...interface{}) { Default.Warn(v...) }

// Warnf prints format [WARN] tag
func Warnf(format string, args ...interface{}) { Default.Warnf(format, args...) }

// Debug prints variables [DEBU] tag
func Debug(v ...interface{}) { Default.Debug(v...) }

// Debugf prints format [DEBU] tag
func Debugf(format string, args ...interface{}) { Default.Debugf(format, args...) }

// Error prints variables [ERRO] tag
func Error(v ...interface{}) { Default.Error(v...) }

// Errorf prints format [ERRO] tag
func Errorf(format string, args ...interface{}) { Default.Errorf(format, args...) }

// Fatal prints variables [FATA] tag followed by an os.Exit(-1).
func Fatal(v ...interface{}) { Default.Fatal(v...) }

// Fatalf prints format [FATA] tag followed by an os.Exit(-1).
func Fatalf(format string, args ...interface{}) { Default.Fatalf(format, args...) }

// Print is equivlent to Info
func Print(v ...interface{}) { Default.Print(v...) }

// Printf is equivlent to Infof
func Printf(format string, args ...interface{}) { Default.Printf(format, args...) }

// HTTP prints variables [HTTP] tag
func HTTP(v ...interface{}) { Default.HTTP(v...) }

// HTTPf prints format [HTTP] tag
func HTTPf(format string, args ...interface{}) { Default.HTTPf(format, args...) }

// Time prints variables [TIME] tag
func Time(v ...interface{}) { Default.Time(v...) }

// Timef prints format [TIME] tag
func Timef(format string, args ...interface{}) { Default.Timef(format, args...) }
