package parser

type GrammarSpec struct {
	rules []RuleSpec
}

type RuleSpec struct {
	name string
	choice Choice
	arrange PostProcessing
}

func NewGrammar() Grammar {
	return &grammar{[]EarleyRule{}, ""}
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
	rules []EarleyRule
	start string
}

type EarleyRule struct {
	name    string
	choices []Choice
}
	
type Choice struct {
	symbols []EarleySymbol
	arrange PostProcessing
}

type EarleySymbol interface {
	isSymbol()
}

func (LiteralMatcher) isSymbol() {}
func (PatternMatcher) isSymbol() {}
func (RuleMatcher)    isSymbol() {}

type LiteralMatcher struct {
	image string
}

type PatternMatcher struct {
	pattern string
}

type RuleMatcher struct {
	name string
}
