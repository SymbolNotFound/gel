package parser

type l = Literal
type p = Matcher
type s = NonTerm
type kv = KeyValue
type ref = ItemProjection
type str = StringProjection
type get = PropertyGetter

// Variable-length symbol list constructor.
func def(symbols ...EarleySymbol) []EarleySymbol {
	return symbols
}

// Variable-length list projection constructor.
func lproj(values ...PostProcessing) ListProjection {
	return ListProjection{
		values: values,
	}
}

// Variable-size record projection constructor.
func rproj(name string, attrs ...attr) RecordProjection {
	return RecordProjection{
		name:  name,
		attrs: attrs,
	}
}

// Shorthand for the first handful of references and the typical expand-first.
var all = ref{0}
var first = ref{1}
var second = ref{2}
var third = ref{3}
var fourth = ref{4}
var fifth = ref{5}
var first_cat = ExpandList{first}

// A lightly-commented already-compiled 
func EarleyBNFGrammar() Grammar {
	var g grammar
	g.lexer = nil // TODO need to refactor TokenReader a bit, move the channel to the return of the .parse/.feed calls
	for _, rule := range []EarleyRule{
		// Starting state, a grammar is a sequence of spacing-delimited productions.
		{"input", def(s{"_"}, s{"grammar"}, s{"_"}), second},
		{"grammar", def(s{"production"}), all},
		{"grammar", def(s{"grammar"}, s{"_"}, s{"production"}),
			lproj(first_cat, third)},

		// Spacing, optional `_` and at-least-one `__`.  Comments are captured here
		// but effectively skipped because no other production rule is including the
		// contents of spacing states.
		{"_", def(s{"__$1"}), all},
		{"__$1", def(s{"__"}), first},
		{"__$1", def(), Nothing{}},
		{"__", def(s{"__$1"}, s{"SPACING"}), lproj(ExpandList{first})},
		{"__", def(s{"__$1"}, s{"COMMENT"}),
			lproj(first_cat, rproj("Comment", kv{"text", second}))},

		// Patterns may have post-processing defined, but only simple references.
		{"production",
			def(s{"WORD"}, s{"_"}, l{"::="}, s{"_"}, s{"pattern_body"}),
			rproj("Pattern", kv{"name", first}, ExpandRecord{fifth})},
		{"pattern_body", def(s{"PATTERN"}),
			rproj("Matcher", kv{"pattern", first})},
		{"pattern_body",
			def(s{"PATTERN"}, s{"_"}, l{"=>"}, s{"_"}, s{"postproc_ref"}),
			rproj("Matcher", kv{"pattern", first}, kv{"post", fifth})},

		// Rules may be repeated, or rules may have their alternate choices listed.
		{"production",
			def(s{"WORD"}, s{"_"}, l{"::="}, s{"_"}, s{"rule_body"}),
			rproj("Rule", kv{"name", first}, kv{"choices", fifth})},
		{"rule_body", def(s{"parse_choice"}), all},
		{"rule_body",
			def(s{"rule_body"}, s{"_"}, l{"|"}, s{"_"}, s{"parse_choice"}),
			lproj(first_cat, fifth)},

		// Each choice has its own post-production context.
		{"parse_choice", def(s{"rule_expr"}),
			rproj("Choice", kv{"tokens", first})},
		{"parse_choice",
			def(s{"rule_expr"}, s{"_"}, l{"=>"}, s{"_"}, s{"postproc_atom"}),
			rproj("Choice", kv{"tokens", first}, kv{"post", fifth})},

		// Each rule expression is a simple concatenation of rule_atom members.
		{"rule_expr", def(s{"rule_atom"}), all},
		{"rule_expr",
			def(s{"rule_expr"}, s{"_"}, s{"rule_atom"}),
			lproj(first_cat, third)},

		// Rule atoms are matchers or subexpressions (also composed of matchers).
		{"rule_atom", def(s{"rule_matcher"}), first},
		{"rule_atom", def(s{"rule_matcher"}, s{"KLEENE_MOD"}),
			rproj("Matcher", ExpandRecord{first}, kv{"kleene", second})},
		{"rule_atom", def(l{"("}, s{"_"}, s{"rule_body"}, s{"_"}, l{")"}),
			rproj("Expr", kv{"tokens", third})},
		{"rule_atom", def(l{"("}, s{"_"}, s{"rule_body"}, s{"_"}, l{")"},
			s{"_"}, s{"kleene_mod"}),
			rproj("Expr", kv{"tokens", third}, kv{"kleene", ref{7}})},
		{"rule_atom", def(l{"["}, s{"_"}, s{"rule_body"}, s{"_"}, l{"]"}),
			rproj("Expr", kv{"tokens", third}, kv{"kleene", str{"?"}})},
		{"rule_atom", def(l{"{"}, s{"_"}, s{"rule_body"}, s{"_"}, l{"}"}),
			rproj("Expr", kv{"tokens", third}, kv{"kleene", str{"*"}})},

		// Symbolic references, literal strings or character classes (e.g., [a-z])
		{"rule_matcher", def(s{"WORD"}),
			rproj("Matcher", kv{"nonterm", first})},
		{"rule_matcher", def(s{"STRING"}),
			rproj("Matcher", kv{"literal", first})},
		{"rule_matcher", def(s{"CHARCLASS"}),
			rproj("Matcher", kv{"pattern", first})},

		// Non-trivial token definitions.
		{"SPACING", def(p{"\\s+"}), all},
		{"COMMENT", def(p{"\\(\\*([^*]+|\\*+[^)])*\\*+\\)"}), all},
		{"WORD", def(p{"[A-Z_a-z][A-Z_a-z0-9]*"}), all},
		{"NUMBER", def(p{"0|[1-9][0-9]*"}), all},
		{"STRING",
			def(p{"\"(\\\\[\"bfnrt\\/\\\\]|\\u[a-fA-F0-9]{4}|[^\"\\\\\\n])*\""}),
			first},
		{"KLEENE_MOD", def(p{"[?*+]"}), all},
		// Only the simple `[` ... `]` form of character class is supported.
		{"CHARCLASS", def(p{"\\[(?:\\\\.|[^\\\\\\n])+?\\]"}), all},
		// For more complicated (regular) expressions, use pattern notation.
		{"PATTERN", def(p{"/(?:\\\\.|[^\\\\\\n])+?/m?"}), all},

		// Post-processing top-level constructions.
		{"postproc_atom", def(s{"postproc_prop"}), first},
		{"postproc_atom", def(s{"postproc_ref"}), first},
		{"postproc_atom", def(s{"postproc_list"}), first},
		{"postproc_atom", def(s{"postproc_record"}), first},

		// (state reference)
		{"postproc_ref", def(l{"\\"}, s{"NUMBER"}),
			rproj("ItemProjection", kv{"ref", second})},

		// (property accessor)
		{"postproc_prop", def(s{"postproc_ref"}, l{"."}, s{"WORD"}),
			rproj("PropertyGetter", kv{"ref", first}, kv{"name", third})},
		{"postproc_prop", def(s{"postproc_prop"}, l{"."}, s{"WORD"}),
			rproj("PropertyGetter", kv{"ref", first}, kv{"name", third})},
		{"postproc_prop", def(s{"postproc_ref"}, l{"."}, s{"NUMBER"}),
			rproj("PropertyGetter", kv{"ref", first}, kv{"name", third})},
		{"postproc_prop", def(s{"postproc_prop"}, l{"."}, s{"NUMBER"}),
			rproj("PropertyGetter", kv{"ref", first}, kv{"name", third})},

		// (list projection)
		{"postproc_list", def(l{"["}, s{"_"}, s{"postproc_items"}, s{"_"},
			l{","}, s{"_"}, l{"]"}),
			rproj("ListProjection", kv{"values", third})},
		{"postproc_list", def(l{"["}, s{"_"}, s{"postproc_items"}, s{"_"}, l{"]"}),
			rproj("ListProjection", kv{"values", third})},

		// (list items)
		{"postproc_items", def(s{"postproc_item"}), all},
		{"postproc_items",
			def(s{"postproc_items"}, s{"_"}, l{","}, s{"_"}, s{"postproc_item"}),
			lproj(first_cat, fifth)},

		{"postproc_item", def(s{"postproc_prop"}), first},
		{"postproc_item", def(s{"postproc_list"}), first},
		{"postproc_item", def(s{"postproc_record"}), first},
		{"postproc_item", def(s{"postproc_ref"}, l{"..."}),
			rproj("ExpandRecord", kv{"ref", get{first, "ref"}})},
		{"postproc_item", def(s{"postproc_ref"}), first},

		// (record projection)
		{"postproc_record",
			def(s{"WORD"}, l{"{"}, s{"_"}, s{"postproc_keyvals"},
				s{"_"}, l{","}, s{"_"}, l{"}"}),
			rproj("RecordProjection", kv{"name", first}, kv{"attrs", fourth})},
		{"postproc_record",
			def(s{"WORD"}, l{"{"}, s{"_"}, s{"postproc_keyvals"}, s{"_"}, l{"}"}),
			rproj("RecordProjection", kv{"name", first}, kv{"attrs", fourth})},

		// (key-value attributes)
		{"postproc_keyvals", def(s{"postproc_kv"}), all},
		{"postproc_keyvals",
			def(s{"postproc_keyvals"}, s{"_"}, l{","}, s{"_"}, s{"postproc_kv"}),
			lproj(first_cat, fifth)},
		{"postproc_kv", def(s{"WORD"}, s{"_"}, l{":"}, s{"_"}, s{"postproc_atom"}),
			rproj("kv", kv{"key", first}, kv{"value", fifth})},
		{"postproc_kv", def(s{"postproc_ref"}, l{"..."}),
			rproj("ExpandRecord", kv{"ref", get{first, "ref"}})},
	} {
		g.AddRule(rule)
	}

	return &g
}
