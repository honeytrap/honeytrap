package stringformatter

import (
	"testing"
	"time"
)

func TestFormatter(t *testing.T) {
	templ := `datetime: {{timefmt .Time ""}} Text: {{.VarText}}`
	dt := time.Date(2018, time.January, 25, 11, 50, 0, 0, time.UTC)
	text := "ABC"

	tpl, err := New(templ)
	if err != nil {
		t.Error(err)
	}

	want := "datetime: " + dt.Format(time.RFC3339) + " Text: " + text
	str := tpl.String(dt, text)

	if str != want {
		t.Errorf("Lenghts don't match; want %s got %s", want, str)
	}
}
