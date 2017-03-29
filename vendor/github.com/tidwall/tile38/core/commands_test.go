package core

import (
	"fmt"
	"sort"
	"testing"
)

func TestCommands(t *testing.T) {
	var names []string
	for name := range Commands {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		cmd := Commands[name]
		if cmd.Group == "server" {
			fmt.Printf("%v\n", cmd.String())
		}
	}

}
