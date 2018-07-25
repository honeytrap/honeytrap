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

var log = logging.MustGetLogger("filters/yara")

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
	case "", "file":
		content, err := ioutil.ReadFile(u.Path)
		return []byte(content), err
	case "http", "https":
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

func helper(node interface{}) stringSet {
	switch v := node.(type) {
	case data.Expression:
		return findUnknownIdentifiers(v)
	case string:
		ret := make(stringSet)
		ret[v] = struct{}{}
		return ret
	case data.RegexPair, data.RawString, data.Keyword, data.StringCount, int64, bool, nil:
		return make(stringSet)
	default:
		log.Errorf("Unknown AST type %#v\n", v)
		return make(stringSet)
	}
}

func findUnknownIdentifiers(tree data.Expression) stringSet {
	return helper(tree.Left).Merge(helper(tree.Right))
}

type Compiler struct {
	compiler *goyara.Compiler

	allowedVariables stringSet
	ruleNames        stringSet
}

func NewCompiler() (Compiler, error) {
	c, err := goyara.NewCompiler()
	if err != nil {
		return Compiler{}, err
	}
	return Compiler{c, make(stringSet), make(stringSet)}, nil
}

// Stubs unknown variables
func (c *Compiler) AddString(rules string) error {
	ruleset, err := grammar.Parse(bytes.NewReader([]byte(rules)), os.Stdout)
	if err != nil {
		return err
	}
	c.allowedVariables = make(stringSet)
	for _, rule := range ruleset.Rules {
		c.ruleNames.Add(rule.Identifier)
	}
	for _, rule := range ruleset.Rules {
		unknowns := findUnknownIdentifiers(rule.Condition)
		for v := range unknowns {
			if c.ruleNames.Has(v) {
				// Defining a variable with the same name as a rule results in unexpected behaviour.
				// This can happen when conditions refer to private rules
				continue
			}
			if c.allowedVariables.Has(v) {
				// Variable was already defined.
				// (Defining a variable more than once results in unexpected behaviour: VirusTotal/yara#908)
				continue
			}
			c.allowedVariables.Add(v)
			log.Debugf("Patching unknown identifier %s", v)
			err := c.compiler.DefineVariable(v, "")
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

	allowedVariables stringSet
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
	for name := range m.allowedVariables {
		key := strings.Replace(name, "___", "-", -1)
		key = strings.Replace(key, "__", ".", -1)
		if !e.Has(key) {
			continue
		}
		log.Debugf("Define %s = %s", name, e.Get(key))
		err := m.rules.DefineVariable(name, e.Get(key))
		if err != nil {
			return nil, err
		}
	}
	// If the event doesn't contain a payload, an empty one will be used
	payload := []byte(e.Get("payload"))
	matches, err := m.rules.ScanMem(payload, 0, 30*time.Second)
	if err != nil {
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
	return func(state transforms.State, e event.Event, send func(event.Event)) {
		matches, err := m.GetMatches(e)
		if err != nil {
			panic(err.Error())
			log.Error(err.Error())
			return
		}
		for _, match := range matches {
			// Duplicate the event and add Yara metadata
			extendedEvt := event.New(
				event.MergeFrom(event.ToMap(e)),
				event.Custom("yara.rule", match.Rule),
				event.Custom("yara.tags", strings.Join(match.Tags, ",")),
			)
			send(extendedEvt)
		}
	}
}
