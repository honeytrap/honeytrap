package banner

import (
	"strings"
	"text/template"
	"time"
)

type bannerT struct {
	tmpl *template.Template
}

func New(Host, Text string, DateTime bool) (*bannerT, error) {
	var tmpl strings.Builder

	if Host != "" {
		tmpl.WriteString(Host)
		tmpl.WriteString(" ")
	}

	if Text != "" {
		tmpl.WriteString(Text)
		tmpl.WriteString(" ")
	}

	if DateTime {
		tmpl.WriteString("{{datetime}}")
	}

	t, err := template.New("").Funcs(template.FuncMap{
		"datetime": func() string {
			return time.Now().Format(time.RFC3339)
		},
	}).Parse(tmpl.String())
	if err != nil {
		return nil, err
	}

	return &bannerT{
		tmpl: t,
	}, nil
}

func (t *bannerT) Banner() string {
	var parsed strings.Builder

	if err := t.tmpl.Execute(&parsed, ""); err != nil {
		return ""
	}

	return parsed.String()
}
