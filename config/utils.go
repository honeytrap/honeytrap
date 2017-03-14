package config

import (
	"strconv"
	"time"
)

//ConvertToInt wraps the internal int coverter
func ConvertToInt(target string, def int) int {
	fo, err := strconv.Atoi(target)
	if err != nil {
		return def
	}
	return fo
}

// should become internal functions , config should return time.Duration
func MakeDuration(target string, def int) time.Duration {
	if !elapso.MatchString(target) {
		return time.Duration(def)
	}

	matchs := elapso.FindAllStringSubmatch(target, -1)

	if len(matchs) <= 0 {
		return time.Duration(def)
	}

	match := matchs[0]

	if len(match) < 3 {
		return time.Duration(def)
	}

	dur := time.Duration(ConvertToInt(match[1], def))

	mtype := match[2]

	switch mtype {
	case "s":
		log.Infof("Setting %d in seconds", dur)
		return dur * time.Second
	case "mcs":
		log.Infof("Setting %d in Microseconds", dur)
		return dur * time.Microsecond
	case "ns":
		log.Infof("Setting %d in Nanoseconds", dur)
		return dur * time.Nanosecond
	case "ms":
		log.Infof("Setting %d in Milliseconds", dur)
		return dur * time.Millisecond
	case "m":
		log.Infof("Setting %d in Minutes", dur)
		return dur * time.Minute
	case "h":
		log.Infof("Setting %d in Hours", dur)
		return dur * time.Hour
	default:
		log.Infof("Defaul %d to Seconds", dur)
		return time.Duration(dur) * time.Second
	}

}
