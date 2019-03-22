package yara

import "github.com/honeytrap/yara-parser/data"

/*
We need to walk the AST tree to figure out what identifiers are used by the code,
because Yara requires us to declare all variables before using them. Note that
findUnknownIdentifiers will also report private rule names; they are removed later
in the code.
*/

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