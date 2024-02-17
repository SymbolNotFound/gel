package parser

import "github.com/SymbolNotFound/ggdl/pkg/lexer"

func NewGrammar() Grammar {
	return &grammar{nil, []EarleyRule{}, ""}
}

type Grammar interface {
	SetLexer(lexer lexer.TokenReader) error
	AddRule(rule EarleyRule)
}

// Convenience constructor for literal matches.
func l(image string) EarleySymbol {
	return Literal{
		image: image,
	}
}

// Convenience constructor for pattern matching.
func p(pattern string) EarleySymbol {
	return Matcher{
		pattern: pattern,
	}
}

// Convenience constructor for Nonterminal references.
func s(name string) EarleySymbol {
	return NonTerm{
		name: name,
	}
}

// Symbol list constructor.
func def(symbols ...EarleySymbol) []EarleySymbol {
	return symbols
}

// Projects just the first element in the list (a common postprocessor).
var first = ref{0}

func (g *grammar) SetLexer(lexer lexer.TokenReader) error {
	// TODO check
	// * arg not nil
	// * lexer not already set

	g.lexer = lexer
	return nil
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

func EarleyBNFGrammar() Grammar {
	var g grammar
	g.lexer = nil // TODO need to refactor TokenReader a bit, move the channel to the return of the .parse/.feed calls
	for _, rule := range []EarleyRule{
		{"input", def(s("_"), s("grammar"), s("_")), ref{1}},
		{"grammar", def(s("production")), Ident{}},
		{"grammar", def(s("grammar"), s("_"), s("production")),
			projlist(Expand{0}, ref{2})},

		{"_", def(s("__$1")), Ident{}},
		{"__$1", def(s("__")), first},
		{"__$1", def(), Nothing{}},
		{"__", def(s("__$1"), s("SPACING")), Nothing{}},
		{"__", def(s("__$1"), s("COMMENT")),
			projlist(Expand{0}, projdict("Comment", keyval{"text", ref{1}}))},

		{"production",
			def(s("WORD"), s("_"), l("::="), s("_"), s("pattern_body")),
			projdict("Pattern", keyval{"name", ref{0}}, Expand{4})},
		{"production",
			def(s("WORD"), s("_"), l("::="), s("_"), s("rule_body")),
			projdict("Rule", keyval{"name", ref{0}}, keyval{"choices", ref{4}})},

		{"pattern_body", def(s("PATTERN")),
			projdict("Matcher", keyval{"pattern", ref{0}})},
		{"pattern_body",
			def(s("PATTERN"), s("_"), l("=>"), s("_"), s("postproc_ref")),
			projdict("Matcher", keyval{"pattern", ref{0}}, keyval{"post", ref{4}})},

		{"rule_body", def(s("parse_choice")), Ident{}},
		{"rule_body",
			def(s("rule_body"), s("_"), l("|"), s("_"), s("parse_choice")),
			projlist(Expand{0}, ref{4})},

		{"parse_choice", def(s("rule_expr")),
			projdict("Choice", keyval{"tokens", ref{0}})},
		{"parse_choice",
			def(s("rule_expr"), s("_"), l("=>"), s("_"), s("postproc_atom")),
			projdict("Choice", keyval{"tokens", ref{0}}, keyval{"post", ref{4}})},

		{"rule_expr", def(s("rule_atom")), Ident{}},

		{"rule_expr",
			def(s("rule_expr"), s("_"), s("rule_atom")),
			projlist(Expand{0}, ref{2})},

		{"rule_atom", def(s("rule_matcher")), ref{0}},
		{"rule_atom", def(s("rule_matcher"), s("KLEENE_MOD")),
			projdict("Matcher", Expand{0}, keyval{"kleene", ref{1}})},

		{"rule_atom", def(l("("), s("_"), s("rule_body"), s("_"), l(")")),
			projdict("Expr", keyval{"tokens", ref{2}})},
		{"rule_atom", def(l("("), s("_"), s("rule_body"), s("_"), l(")"),
			s("_"), s("kleene_mod")),
			projdict("Expr", keyval{"tokens", ref{2}}, keyval{"kleene", ref{6}})},

		{"rule_atom", def(l("["), s("_"), s("rule_body"), s("_"), l("]"), s("_")),
			projdict("Expr",
				keyval{"tokens", ref{2}}, keyval{"kleene", StringProjection{"?"}})},
		{"rule_atom", def(l("{"), s("_"), s("rule_body"), s("_"), l("}"), s("_")),
			projdict("Expr",
				keyval{"tokens", ref{2}}, keyval{"kleene", StringProjection{"*"}})},

		{"rule_matcher", def(s("WORD")),
			projdict("Matcher", keyval{"nonterm", first})},
		{"rule_matcher", def(s("STRING")),
			projdict("Matcher", keyval{"literal", first})},
		{"rule_matcher", def(s("CHARCLASS")),
			projdict("Matcher", keyval{"charset", first})},

		{"SPACING", def(p("\\s+")), first},
		{"COMMENT", def(p("\\(\\*([^*]+|\\*+[^)])*\\*+\\)")), first},
		{"WORD", def(p("[A-Z_a-z][A-Z_a-z0-9]*")), first},
		{"NUMBER", def(p("0|[1-9][0-9]*")), first},
		{"STRING",
			def(p("\"(\\\\[\"bfnrt\\/\\\\]|\\u[a-fA-F0-9]{4}|[^\"\\\\\\n])*\"")),
			ref{1}},
		{"CHARCLASS", def(p("\\[(?:\\\\.|[^\\\\\\n])+?\\]")), first},
		{"PATTERN", def(p("/(?:\\\\.|[^\\\\\\n])+?/m?")), first},
		{"KLEENE_MOD", def(p("[?*+]")), first},

		{"postproc_atom", def(s("postproc_ref")), first},
		{"postproc_atom", def(s("postproc_list")), first},
		{"postproc_atom", def(s("postproc_dict")), first},
		{"postproc_ref", def(l("\\"), s("NUMBER")),
			projdict("ItemProjection", keyval{"Reference", ref{1}})},

		{"postproc_list", def(l("["), s("_"), s("postproc_items"), s("_"), l("]")),
			projdict("ListProjection", keyval{"values", ref{2}})},
		{"postproc_items", def(s("postproc_item")), Ident{}},
		{"postproc_items",
			def(s("postproc_items"), s("_"), l(","), s("_"), s("postproc_items")),
			projlist(Expand{0}, ref{4})},

		{"postproc_item", def(s("postproc_ref")), first},
		{"postproc_item", def(s("postproc_expand")), first},
		{"postproc_item", def(s("postproc_list")), first},
		{"postproc_item", def(s("postproc_dict")), first},

		{"postproc_expand", def(s("postproc_ref"), l("...")),
			projdict("Expand", keyval{"Reference", ref{0}})},

		{"trailing_comma", def(l(",")), Nothing{}},
		{"trailing_comma", def(), Nothing{}},
		{"postproc_dict",
			def(s("WORD"), l("{"), s("_"), s("postproc_keyvals"),
				s("_"), s("trailing_comma"), s("_"), l("}")),
			projdict("DictProjection", keyval{"name", ref{0}}, keyval{"attrs", ref{3}})},
		{"postproc_keyvals", def(s("postproc_kv")), Ident{}},
		{"postproc_keyvals",
			def(s("postproc_keyvals"), s("_"), l(","), s("_"), s("postproc_kv")),
			projlist(Expand{0}, ref{4})},
		{"postproc_kv", def(s("WORD"), s("_"), l(":"), s("_"), s("postproc_atom")),
			projdict("keyval", keyval{"name", ref{0}}, keyval{"value", ref{4}})},
		{"postproc_kv", def(s("postproc_expand")), first},
	} {
		g.AddRule(rule)
	}

	return &g
}
