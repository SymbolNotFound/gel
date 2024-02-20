package parser

type l = LiteralMatcher
type p = PatternMatcher
type s = RuleMatcher
type kv = KeyValue
type str = StringProjection
type ref = ItemProjection
type get = PropertyGetter

// Variable-length symbol list constructor.
func spec(symbols ...EarleySymbol) []EarleySymbol {
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
	gs := GrammarSpec{ rules: []EarleyRule{
		// Starting state, a grammar is a sequence of spacing-delimited productions.
		{"input", spec(s{"_"}, s{"grammar"}, s{"_"}), rproj(
			"GrammarSpec", kv{"rules", second})},
		{"grammar", def(s{"production"}), all},
		{"grammar", def(s{"grammar"}, s{"__"}, s{"production"}),
			// Also captures top-level comments from __.
			lproj(first_cat, ExpandList{second}, third)},

		// Spacing, optional `_` and at-least-one `__`.  Comments are captured here
		// but will be skipped unless the enclosing rule maintains a reference.
		{"_", def(s{"__$1"}), all},
		{"__$1", def(s{"__"}), first},
		{"__$1", def(), Nothing{}},
		{"__", spec(s{"__$1"}, s{"SPACING"}), lproj(ExpandList{first})},
		{"__", spec(s{"__$1"}, s{"COMMENT"}),
				lproj(first_cat, rproj("Comment", kv{"text", second}))},

		// Patterns may have post-processing defined, but only simple references.
		{"production",
			spec(s{"WORD"}, s{"_"}, l{"::="}, s{"_"}, s{"pattern_body"}),
			rproj("EarleyRule", kv{"name", first},
			  kv{"choices", lproj(fifth)})},
		{"pattern_body",
		  spec(s{"PATTERN"}),
			rproj("Choice",
				kv{"symbols", lproj(rproj("PatternMatcher", kv{"pattern", first}))},
				kv{"arrange", all})},
		{"pattern_body",
		  spec(s{"PATTERN"}, s{"_"}, l{"=>"}, s{"_"}, s{"postproc_ref"}),
			rproj("Choice",
				kv{"symbols", lproj(rproj("PatternMatcher", kv{"pattern", first}))},
				kv{"arrange", fifth})},

		// Rules may be repeated, or rules may have their alternate choices listed.
		{"production",
			spec(s{"WORD"}, s{"_"}, l{"::="}, s{"_"}, s{"rule_body"}),
			rproj("EarleyRule", kv{"name", first}, kv{"choices", fifth})},
		{"rule_body", spec(s{"parse_choice"}), all},
		{"rule_body",
			spec(s{"rule_body"}, s{"_"}, l{"|"}, s{"_"}, s{"parse_choice"}),
			lproj(first_cat, fifth)},

		// Each choice has its own post-production context.
		{"parse_choice",
			spec(s{"rule_expr"}, s{"_"}, l{"=>"}, s{"_"}, s{"postproc_atom"}),
			rproj("Choice", kv{"tokens", first}, kv{"arrange", fifth})},
		// By default, all Earley state matches are projected.
		{"parse_choice", spec(s{"rule_expr"}),
			rproj("Choice", kv{"tokens", first}, kv{"arrange", all})},

		// Each rule expression is a simple concatenation of rule_atom members.
		{"rule_expr", spec(s{"rule_atom"}), all},
		{"rule_expr",
			spec(s{"rule_expr"}, s{"_"}, s{"rule_atom"}),
			lproj(first_cat, third)},

		// Rule atoms are matchers or subexpressions (also composed of matchers).
		{"rule_atom", spec(s{"rule_matcher"}), first},
		{"rule_atom", spec(s{"rule_matcher"}, s{"KLEENE_MOD"}),
			rproj("record", ExpandRecord{first}, kv{"kleene", second})},
		{"rule_atom", spec(l{"("}, s{"_"}, s{"rule_body"}, s{"_"}, l{")"}),
			rproj("Expr", kv{"tokens", third})},
		{"rule_atom", spec(l{"("}, s{"_"}, s{"rule_body"}, s{"_"}, l{")"},
			s{"_"}, s{"kleene_mod"}),
			rproj("Expr", kv{"tokens", third}, kv{"kleene", ref{7}})},
		{"rule_atom", spec(l{"["}, s{"_"}, s{"rule_body"}, s{"_"}, l{"]"}),
			rproj("Expr", kv{"tokens", third}, kv{"kleene", str{"?"}})},
		{"rule_atom", spec(l{"{"}, s{"_"}, s{"rule_body"}, s{"_"}, l{"}"}),
			rproj("Expr", kv{"tokens", third}, kv{"kleene", str{"*"}})},

		// Symbolic references, literal strings or character classes (e.g., [a-z])
		{"rule_matcher", spec(s{"WORD"}),
			rproj("RuleMatcher", kv{"name", first})},
		{"rule_matcher", spec(s{"STRING"}),
			rproj("LiteralMatcher", kv{"image", first})},
		{"rule_matcher", spec(s{"CHARCLASS"}),
			rproj("PatternMatcher", kv{"pattern", first})},

		// Non-trivial token definitions.
		{"SPACING", spec(p{"(?m:\\s+)"}), all},
		{"COMMENT", spec(p{"(?m:\\(\\*([^*]+|\\*+[^)])*\\*+\\))"}), all},
		{"WORD", spec(p{"[A-Z_a-z][A-Z_a-z0-9]*"}), all},
		{"NUMBER", spec(p{"0|[1-9][0-9]*"}), all},
		{"STRING",
			spec(p{"\"(\\\\[\"bfnrt\\/\\\\]|\\u[a-fA-F0-9]{4}|[^\"\\\\\\n])*\""}),
			first},
		{"KLEENE_MOD", spec(p{"[?*+]"}), all},
		// Only the simple `[` ... `]` form of character class is supported.
		{"CHARCLASS", spec(p{"\\[(?:\\\\.|[^\\\\\\n])+?\\]"}), all},
		// For more complicated (regular) expressions, use pattern notation.
		{"PATTERN", spec(p{"/(?:\\\\.|[^\\\\\\n])+?/m?"}), all},

		// Post-processing top-level constructions.
		{"postproc_atom", spec(s{"postproc_prop"}), first},
		{"postproc_atom", spec(s{"postproc_ref"}), first},
		{"postproc_atom", spec(s{"postproc_list"}), first},
		{"postproc_atom", spec(s{"postproc_record"}), first},

		// (state reference)
		{"postproc_ref", spec(l{"\\"}, s{"NUMBER"}),
			rproj("ItemProjection", kv{"ref", second})},

		// (property accessor)
		{"postproc_prop", spec(s{"postproc_ref"}, l{"."}, s{"WORD"}),
			rproj("PropertyGetter", kv{"ref", first}, kv{"name", third})},
		{"postproc_prop", spec(s{"postproc_prop"}, l{"."}, s{"WORD"}),
			rproj("PropertyGetter", kv{"ref", first}, kv{"name", third})},
		{"postproc_prop", spec(s{"postproc_ref"}, l{"."}, s{"NUMBER"}),
			rproj("ElementGetter", kv{"ref", first}, kv{"incex", third})},
		{"postproc_prop", spec(s{"postproc_prop"}, l{"."}, s{"NUMBER"}),
			rproj("ElementGetter", kv{"ref", first}, kv{"index", third})},

		// (list projection)
		{"postproc_list",
		  spec(l{"["}, s{"_"}, s{"postproc_items"}, s{"_"}, l{","}, s{"_"}, l{"]"}),
			rproj("ListProjection", kv{"values", third})},
		{"postproc_list",
		  spec(l{"["}, s{"_"}, s{"postproc_items"}, s{"_"}, l{"]"}),
			rproj("ListProjection", kv{"values", third})},

		// (list items)
		{"postproc_items", spec(s{"postproc_item"}), all},
		{"postproc_items",
			spec(s{"postproc_items"}, s{"_"}, l{","}, s{"_"}, s{"postproc_item"}),
			lproj(first_cat, fifth)},

		{"postproc_item", spec(s{"postproc_prop"}), first},
		{"postproc_item", spec(s{"postproc_list"}), first},
		{"postproc_item", spec(s{"postproc_record"}), first},
		{"postproc_item", spec(s{"postproc_ref"}, l{"..."}),
			rproj("ExpandList", kv{"ref", get{first, "ref"}})},
		{"postproc_item", spec(s{"postproc_ref"}), first},

		// (record projection)
		{"postproc_record",
			spec(s{"WORD"}, l{"{"}, s{"_"}, s{"postproc_keyvals"},
				s{"_"}, l{","}, s{"_"}, l{"}"}),
			rproj("RecordProjection", kv{"name", first}, kv{"attrs", fourth})},
		{"postproc_record",
			spec(s{"WORD"}, l{"{"}, s{"_"}, s{"postproc_keyvals"}, s{"_"}, l{"}"}),
			rproj("RecordProjection", kv{"name", first}, kv{"attrs", fourth})},

		// (key-value attributes)
		{"postproc_keyvals", spec(s{"postproc_kv"}), all},
		{"postproc_keyvals",
			spec(s{"postproc_keyvals"}, s{"_"}, l{","}, s{"_"}, s{"postproc_kv"}),
			lproj(first_cat, fifth)},
		{"postproc_kv",
		  spec(s{"kv_key"}, s{"_"}, l{":"}, s{"_"}, s{"kv_value"}),
			rproj("KeyValue", kv{"key", first}, kv{"value", fifth})},
		{"postproc_kv", spec(s{"postproc_ref"}, l{"..."}),
			rproj("ExpandRecord", kv{"ref", get{first, "ref"}})},
		{"kv_key", spec(s{"WORD"}), first},
		{"kv_key", spec(s{"STRING"}), first},
		{"kv_value", spec(s{"STRING"}), first},
		{"kv_value", spec(s{"postproc_atom"}), first},
	}}

	var g grammar
	for _, rule := range gs.rules {
		g.AddRule(rule)
	}
	return &g
}
