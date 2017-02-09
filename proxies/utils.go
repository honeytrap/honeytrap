package proxies

// Wrap a byte slice paragraph for use in SMTP header
func wrap(sl []byte) []byte {
	length := 0
	for i := 0; i < len(sl); i++ {
		if length > 76 && sl[i] == ' ' {
			sl = append(sl, 0, 0)
			copy(sl[i+2:], sl[i:])
			sl[i] = '\r'
			sl[i+1] = '\n'
			sl[i+2] = '\t'
			i += 2
			length = 0
		}
		if sl[i] == '\n' {
			length = 0
		}
		length++
	}
	return sl
}
