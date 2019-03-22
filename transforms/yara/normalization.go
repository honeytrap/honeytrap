package yara

import "strings"

/*
Yara doesn't accept identifiers with dots and hyphens, so we need to convert them
to something else. Current rules:

"." => "__"
"-" => "___"
*/

func normalize(name string) string {
	out := strings.Replace(name, "-", "___", -1)
	out = strings.Replace(out, ".", "__", -1)
	return out
}

func denormalize(name string) string {
	out := strings.Replace(name, "___", "-", -1)
	out = strings.Replace(out, "__", ".", -1)
	return out
}