package stringformatter

import (
	"fmt"
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
	str := tpl.Format(dt, text)

	if str != want {
		t.Errorf("Strings don't match; want %s got %s", want, str)
	}
}

func ExampleFormat() {
	templ := `Date and Time: {{timefmt .Time "Mon 2 Jan 2006 15:04:02"}} -- Some Text: {{.VarText}}`

	t, err := New(templ)
	if err != nil {
		fmt.Println(err)
	}

	tm := time.Date(2018, time.February, 11, 15, 40, 0, 0, time.UTC)

	out := t.Format(tm, "VARTEXT")

	fmt.Println(out)
	//Output: Date and Time: Sun 11 Feb 2018 15:40:11 -- Some Text: VARTEXT
}

func ExampleFormatText() {
	templ := `{{.VarText}}`

	t, err := New(templ)
	if err != nil {
		fmt.Println(err)
	}

	tm := time.Time{}
	out := t.Format(tm, "Some Text")

	fmt.Println(out)
	//Output: Some Text
}
