package parser

import "github.com/SymbolNotFound/ggdl/pkg/lexer"

func NewGrammar() Grammar {
	return &grammar{nil, []EarleyRule{}, ""}
}

type Grammar interface {
	AddRule(rule EarleyRule)
}

func (g *grammar) AddRule(rule EarleyRule) {
	if len(g.rules) == 0 {
		g.start = rule.name
	}
	g.rules = append(g.rules, rule)
}

type grammar struct {
	lexer lexer.TokenReader
	rules []EarleyRule
	start string
}

