package config

import "time"

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
