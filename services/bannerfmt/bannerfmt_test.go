package bannerfmt

import (
	"fmt"
	"testing"
	"time"
)

func TestFormatter(t *testing.T) {
	templ := `{{.Host}} {{timefmt .Time ""}}`
	dt := time.Date(2018, time.January, 25, 11, 50, 0, 0, time.UTC)

	data := struct {
		Host string
		Time time.Time
	}{
		Host: "mail.example.org",
		Time: dt,
	}

	tpl, err := New(templ, data)
	if err != nil {
		t.Error(err)
	}

	want := "mail.example.org " + dt.Format(time.RFC3339)
	str := tpl.String()

	if str != want {
		t.Errorf("Strings don't match; want %s got %s", want, str)
	}
}

func TestSet(t *testing.T) {
	templ := `{{.}}`
	value1 := "TEST1"
	value2 := "TEST2"

	banner, err := New(templ, value1)
	if err != nil {
		t.Error(err.Error())
	}

	banner.Set(value2)

	got := banner.String()

	if got != value2 {
		t.Errorf("Set got [%s] but want [%s]", got, value2)
	}
}

func ExampleBannerTimeNow() {
	templ := `{{now ""}}`

	t, err := New(templ, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(t.String())
}

func ExampleBannerOwnTime() {
	templ := `Date and Time: {{timefmt .Time "Mon 2 Jan 2006 15:04:02"}} -- {{.Banner}}`

	tm := time.Date(2018, time.February, 11, 15, 40, 0, 0, time.UTC)

	bannerData := struct {
		Time   time.Time
		Banner string
	}{
		Time:   tm,
		Banner: "BANNER",
	}

	t, err := New(templ, bannerData)
	if err != nil {
		panic(err)
	}

	out := t.String()

	fmt.Println(out)
	//Output: Date and Time: Sun 11 Feb 2018 15:40:11 -- BANNER
}

func ExampleBannerStruct() {
	templ := `{{.Host}} {{.ProductName}} Ready`

	bannerData := struct {
		Host, ProductName string
	}{
		Host:        "banner.example.org",
		ProductName: "SMTP Server 2.3.0.1-2.0b",
	}

	t, err := New(templ, bannerData)
	if err != nil {
		panic(err)
	}

	out := t.String()

	fmt.Println(out)
	//Output: banner.example.org SMTP Server 2.3.0.1-2.0b Ready
}
