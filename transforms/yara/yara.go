package yara

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	goyara "github.com/hillu/go-yara"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/transforms"
	"github.com/honeytrap/yara-parser/data"
	"github.com/honeytrap/yara-parser/grammar"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("filters/yarautils")

// Fetches and loads rules from a specification, a file, or an URL
func LoadRules(source string) ([]byte, error) {
	// Raw input
	if strings.Contains(source, "condition:") {
		return []byte(source), nil
	}
	// URL/File input
	u, err := url.Parse(source)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "":
		fallthrough
	case "file":
		content, err := ioutil.ReadFile(u.Path)
		return []byte(content), err
	case "http":
		fallthrough
	case "https":
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("Got HTTP status code %d (expected 200)", resp.StatusCode)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		return body, err
	default:
		return nil, fmt.Errorf("Unknown path scheme %s", u.Scheme)
	}
}

func helper(node interface{}) []string {
	switch t := node.(type) {
	case data.Expression:
		return findUnknownIdentifiers(node.(data.Expression))
	case string:
		return []string{node.(string)}
	case data.Keyword, data.RawString, data.StringCount, int64, bool, nil:
		return []string{}
	default:
		log.Errorf("Unknown AST type %#v\n", t)
		return []string{}
	}
}

func findUnknownIdentifiers(tree data.Expression) []string {
	return append(helper(tree.Left), helper(tree.Right)...)
}

var baseVariables = []string{} // Do not add "payload"!

type Compiler struct {
	compiler *goyara.Compiler

	allowedVariables []string
}

func NewCompiler() (Compiler, error) {
	c, err := goyara.NewCompiler()
	if err != nil {
		return Compiler{}, err
	}
	return Compiler{c, []string{}}, nil
}

// Stubs unknown variables
func (c *Compiler) AddString(rules string) error {
	ruleset, err := grammar.Parse(bytes.NewReader([]byte(rules)), os.Stdout)
	if err != nil {
		return err
	}
	c.allowedVariables = baseVariables
	for _, rule := range ruleset.Rules {
		unknowns := findUnknownIdentifiers(rule.Condition)
		c.allowedVariables = append(c.allowedVariables, unknowns...)
		for _, u := range unknowns {
			log.Debugf("Patching unknown identifier %s", u)
			err := c.compiler.DefineVariable(u, "")
			if err != nil {
				return err
			}
		}

	}
	return c.compiler.AddString(rules, "ht-config")
}

func (c *Compiler) AddRulesFrom(source string) error {
	body, err := LoadRules(source)
	if err != nil {
		return err
	}
	err = c.AddString(string(body))
	return err
}

type Matcher struct {
	rules *goyara.Rules

	allowedVariables []string
}

func NewMatcher(c Compiler) (Matcher, error) {
	rules, err := c.compiler.GetRules()
	if err != nil {
		return Matcher{}, err
	}
	return Matcher{rules, c.allowedVariables}, nil
}

func NewMatcherFrom(rules string) (Matcher, error) {
	c, err := NewCompiler()
	if err != nil {
		return Matcher{}, err
	}
	err = c.AddRulesFrom(rules)
	if err != nil {
		return Matcher{}, err
	}
	return NewMatcher(c)
}

func (m Matcher) GetMatches(e event.Event) ([]goyara.MatchRule, error) {
	for _, name := range m.allowedVariables {
		err := m.rules.DefineVariable(name, e.Get(strings.Replace(name, "__", ".", -1)))
		if err != nil {
			return nil, err
		}
	}
	// If the event doesn't contain a payload, an empty one will be used
	payload := []byte(e.Get("payload"))
	matches, err := m.rules.ScanMem(payload, 0, 30*time.Second)
	if err != nil {
		panic(err)
		return nil, err
	}
	return matches, nil
}

func (m Matcher) Match(e event.Event) (bool, error) {
	matches, err := m.GetMatches(e)
	return len(matches) > 0, err
}

// Like Match, but panics if an error occurs
func (m Matcher) MustMatch(e event.Event) bool {
	matches, err := m.GetMatches(e)
	if err != nil {
		panic(err)
	}
	return len(matches) > 0
}

func Yara(source string) transforms.TransformFunc {
	c, err := NewCompiler()
	if err != nil {
		panic(err)
	}
	err = c.AddRulesFrom(source)
	if err != nil {
		panic(err)
	}
	m, err := NewMatcher(c)
	if err != nil {
		panic(err)
	}
	return func(e event.Event) []event.Event {
		if (!m.MustMatch(e)) {
			return []event.Event{}
		}
		return []event.Event{e}
	}
}
