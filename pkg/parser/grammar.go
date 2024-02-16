package parser

import "io"

// I normally wouldn't do this but I'll fully type the grammar when its
// dependencies are better defined.  This allows us to easily catch all
// references to `any`, implying that we intend to remove.
type any interface{}

type Lexer interface {
	Restart(chunk string)

	Next() (EarleyToken, error)

	Save() any

	Has(toktype string) bool
}

type EarleyToken interface {
	Value() any
}

type EarleyRule struct {
	name    string
	symbols []EarleySymbol
	arrange PostProcessing
}

type EarleySymbol interface {
	Test(io.Reader) (string, int)
}

type Literal struct {
	image string
}

func (Literal) Test(input io.Reader) (string, int) {
	// TODO evaluate the contents of a reader
	return "", 0
}

type Matcher struct {
	pattern string
}

func (Matcher) Test(input io.Reader) (string, int) {
	// TODO evaluate the contents of reader for this pattern (see regexp pkg)
	return "", 0
}

type NonTerm struct {
	name string
}

func (NonTerm) Test(input io.Reader) (string, int) {
	// TODO evaluate the contents of a reader
	return "", 0
}

type PostProcessing interface {
	isPostProc()
}
type Nothing struct{}

func (Nothing) isPostProc() {}

type ListProjection struct {
	values []PostProcessing
}

func (lp ListProjection) isPostProc() {}

func projlist(values ...PostProcessing) ListProjection {
	return ListProjection{
		values: values,
	}
}

type DictProjection struct {
	name  string
	attrs []attr
}

func (DictProjection) isPostProc() {}

type attr interface {
	isDictKV()
}

type keyval struct {
	key   string
	value PostProcessing
}

func (keyval) isDictKV() {}

func projdict(name string, attrs ...attr) DictProjection {
	return DictProjection{
		name:  name,
		attrs: attrs,
	}
}

type ref struct {
	refID int
}

func (ref) isPostProc() {}

type expand struct {
	ref ref
}

func (expand) isPostProc() {}
func (expand) isDictKV()   {}

var rules []EarleyRule

func init() {
	// Convenience constructor for literal matches.
	l := func(image string) EarleySymbol {
		return Literal{
			image: image,
		}
	}
	// Convenience constructor for pattern matching.
	p := func(pattern string) EarleySymbol {
		return Matcher{
			pattern: pattern,
		}
	}
	// Convenience constructor for Nonterminal references.
	s := func(name string) EarleySymbol {
		return NonTerm{
			name: name,
		}
	}
	// Symbol list constructor.
	def := func(symbols ...EarleySymbol) []EarleySymbol {
		return symbols
	}
	// Projects just the first element in the list (a common postprocessor).
	id := ref{0}

	var countRules int = 31
	rules := make([]EarleyRule, countRules)

	rules[0] = EarleyRule{"input", def(s("_"), s("grammar"), s("_")), ref{1}}
	rules[1] = EarleyRule{"grammar", def(s("production")), nil}
	rules[2] = EarleyRule{"grammar",
		def(s("grammar"), s("_"), s("production")),
		projlist(expand{ref{0}}, ref{2})}

	rules[3] = EarleyRule{"_$1", def(s("__")), id}
	rules[4] = EarleyRule{"_$1", def(), Nothing{}}
	rules[5] = EarleyRule{"_", def(s("_$1")), nil}

	rules[6] = EarleyRule{"__$1", def(s("__")), id}
	rules[7] = EarleyRule{"__$1", def(), Nothing{}}
	rules[8] = EarleyRule{"__", def(s("__$1"), s("SPACING")), nil}
	rules[9] = EarleyRule{"__", def(s("__$1"), s("COMMENT")),
		projlist(expand{ref{0}}, projdict("Comment", keyval{"text", ref{1}}))}

	rules[10] = EarleyRule{"production",
		def(s("WORD"), s("_"), l("::="), s("_"), s("pattern_body")),
		projdict("Pattern", keyval{"name", ref{0}}, expand{ref{4}})}
	rules[11] = EarleyRule{"production",
		def(s("WORD"), s("_"), l("::="), s("_"), s("rule_body")),
		projdict("Rule", keyval{"name", ref{0}}, keyval{"choices", ref{4}})}

	rules[12] = EarleyRule{"pattern_body", def(s("PATTERN")),
		projdict("Pattern", keyval{"pattern", ref{0}})}

	rules[13] = EarleyRule{"rule_body", def(s("parse_alt")), nil}
	rules[14] = EarleyRule{"rule_body",
		def(s("rule_body"), s("_"), l("|"), s("_"), s("parse_choice")),
		projlist(expand{ref{0}}, ref{4})}

	rules[15] = EarleyRule{"parse_choice", def(s("rule_expr")),
		projdict("Choice", keyval{"tokens", ref{0}})}

	rules[16] = EarleyRule{"rule_expr", def(s("rule_atom")), nil}

	rules[17] = EarleyRule{"rule_expr",
		def(s("rule_expr"), s("_"), s("rule_atom")),
		projlist(expand{ref{0}}, ref{2})}

	rules[18] = EarleyRule{"rule_atom", def(s("rule_matcher")),
		projdict("Term", keyval{"match", ref{0}})}

	rules[19] = EarleyRule{"rule_atom", def(s("rule_matcher"), s("KLEENE_MOD")),
		projdict("Term", keyval{"match", ref{0}}, keyval{"kleene", ref{1}})}

	rules[20] = EarleyRule{"rule_atom$1", def(s("KLEENE_MOD")),
		projdict("Term", keyval{"kleene", ref{0}})}
	rules[21] = EarleyRule{"rule_atom$1", def(), Nothing{}}
	rules[22] = EarleyRule{"rule_atom",
		def(l("("), s("_"), s("rule_body"), s("_"), l(")"), s("_"),
			s("rule_atom$1")),
		projdict("Subexpr", keyval{"tokens", ref{2}}, expand{ref{6}})}

	rules[23] = EarleyRule{"rule_matcher", def(s("WORD")),
		projdict("Matcher", keyval{"nonterm", ref{0}})}
	rules[24] = EarleyRule{"rule_matcher", def(s("STRING")),
		projdict("Matcher", keyval{"literal", ref{0}})}
	rules[25] = EarleyRule{"rule_matcher", def(s("CHARCLASS")),
		projdict("Matcher", keyval{"charset", ref{0}})}

	rules[26] = EarleyRule{"SPACING", def(p("\\s+")), id}
	rules[27] = EarleyRule{"COMMENT", def(p("\\(\\*([^*]+|\\*+[^)])*\\*+\\)")), id}
	rules[28] = EarleyRule{"WORD", def(p("[A-Z_a-z][A-Z_a-z0-9]*")), id}
	rules[29] = EarleyRule{"STRING",
		def(p("(\\\\[\"bfnrt\\/\\\\]|\\u[a-fA-F0-9]{4}|[^\"\\\\\\n])*")), id}
	rules[30] = EarleyRule{"CHARCLASS", def(p("\\[(?:\\\\.|[^\\\\\\n])+?\\]")), id}
	rules[31] = EarleyRule{"PATTERN", def(p("/(?:\\\\.|[^\\\\\\n])+?/m?")), id}
	rules[32] = EarleyRule{"KLEENE_MOD", def(p("[?*+]")), id}
}
