package ipp

type matcher func([]byte) (string, bool)

func Ps(buf []byte) (string, bool) {
	if len(buf) > 1 && buf[0] == 0x25 && buf[1] == 0x21 {
		return ".ps", true
	}

	return "", false
}

func Pdf(buf []byte) (string, bool) {
	if len(buf) > 3 &&
		buf[0] == 0x25 && buf[1] == 0x50 &&
		buf[2] == 0x44 && buf[3] == 0x46 {
		return ".pdf", true
	}

	return "", false
}

func Rtf(buf []byte) (string, bool) {
	if len(buf) > 4 &&
		buf[0] == 0x7B && buf[1] == 0x5C &&
		buf[2] == 0x72 && buf[3] == 0x74 &&
		buf[4] == 0x66 {
		return ".rtf", true
	}

	return "", false
}

func extension(doc []byte) string {
	matchers := []matcher{
		Ps,
		Pdf,
		Rtf,
	}

	for _, f := range matchers {
		if ext, ok := f(doc); ok {
			return ext
		}
	}

	return ".octet-stream"
}
