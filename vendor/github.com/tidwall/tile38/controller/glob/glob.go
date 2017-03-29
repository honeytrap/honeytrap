package glob

import "strings"

type Glob struct {
	Pattern string
	Desc    bool
	Limits  []string
	IsGlob  bool
}

func Match(pattern, name string) (matched bool, err error) {
	return wildcardMatch(pattern, name)
}

func IsGlob(pattern string) bool {
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '[', '*', '?':
			_, err := Match(pattern, "whatever")
			return err == nil
		}
	}
	return false
}

func Parse(pattern string, desc bool) *Glob {
	g := &Glob{Pattern: pattern, Desc: desc, Limits: []string{"", ""}}
	if strings.HasPrefix(pattern, "*") {
		g.IsGlob = true
		return g
	}
	if pattern == "" {
		g.IsGlob = false
		return g
	}
	n := 0
	isGlob := false
outer:
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '[', '*', '?':
			_, err := Match(pattern, "whatever")
			if err == nil {
				isGlob = true
			}
			break outer
		}
		n++
	}
	if n == 0 {
		g.Limits = []string{pattern, pattern}
		g.IsGlob = false
		return g
	}
	var a, b string
	if desc {
		a = pattern[:n]
		b = a
		if b[n-1] == 0x00 {
			for len(b) > 0 && b[len(b)-1] == 0x00 {
				if len(b) > 1 {
					if b[len(b)-2] == 0x00 {
						b = b[:len(b)-1]
					} else {
						b = string(append([]byte(b[:len(b)-2]), b[len(b)-2]-1, 0xFF))
					}
				} else {
					b = ""
				}
			}
		} else {
			b = string(append([]byte(b[:n-1]), b[n-1]-1))
		}
		if a[n-1] == 0xFF {
			a = string(append([]byte(a), 0x00))
		} else {
			a = string(append([]byte(a[:n-1]), a[n-1]+1))
		}
	} else {
		a = pattern[:n]
		if a[n-1] == 0xFF {
			b = string(append([]byte(a), 0x00))
		} else {
			b = string(append([]byte(a[:n-1]), a[n-1]+1))
		}
	}
	g.Limits = []string{a, b}
	g.IsGlob = isGlob
	return g
}
