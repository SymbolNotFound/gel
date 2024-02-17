package parser

type EarleyRule struct {
	name    string
	symbols []EarleySymbol
	arrange PostProcessing
}

type EarleySymbol interface {
	isSymbol()
}

func (Literal) isSymbol() {}
func (Matcher) isSymbol() {}
func (NonTerm) isSymbol() {}

type Literal struct {
	image string
}

type Matcher struct {
	pattern string
}

type NonTerm struct {
	name string
}
