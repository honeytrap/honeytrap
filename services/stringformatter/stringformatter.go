package stringformatter

import (
	"strings"
	"text/template"
	"time"
)

type strFormat struct {
	templ *template.Template
}

func New(templ string) (*strFormat, error) {
	t, err := template.New("").Funcs(template.FuncMap{
		"timefmt": func(tm time.Time, fmt string) string {
			if fmt == "" {
				fmt = time.RFC3339
			}

			return tm.Format(fmt)
		},
	}).Parse(templ)
	if err != nil {
		return nil, err
	}

	return &strFormat{
		templ: t,
	}, nil
}

// template: `datetime: {{timefmt .Time "Jan 2 15:04:02 2006"}} Text: {{.VarText}}`
// will render a string like this: "datetime Mar 19 14:17:19 2018 Text: sometext"
func (s *strFormat) String(tm time.Time, vartext string) string {

	var parsed strings.Builder

	if err := s.templ.Execute(&parsed, struct {
		Time    time.Time
		VarText string
	}{
		Time:    tm,
		VarText: vartext,
	}); err != nil {
		return ""
	}

	return parsed.String()
}
