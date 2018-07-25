package yara

// Crude set implementation (workaround for issue VirusTotal/yara#908)
type stringSet map[string]struct{}

func (s stringSet) Has(key string) bool {
	_, ok := s[key]
	return ok
}

func (s stringSet) Add(key string) {
	s[key] = struct{}{}
}

func (s stringSet) Remove(key string) {
	delete(s, key)
}

func (dst stringSet) Merge(src stringSet) stringSet {
	for key := range src {
		dst.Add(key)
	}
	return dst
}
