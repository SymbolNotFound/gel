# Copyright (c) 2023 Symbol Not Found L.L.C.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# gdl-kif.ne: grammar definition for prefix-notation format of GDL.


@{%
import "astutil.js";
%}

# Enable this to compile to TypeScript instead of JavaScript
# @preprocessor typescript

@{%
// Tokenizer definition provided via https://github.com/no-context/moo
const moo = require('moo')

let lexer = moo.compile({
  SPACE: {
    match: /\s+/,
    lineBreaks: true
  },
  COMMENT: {
    match: /;[^\n]*/,
    value: s => s.substring(1).trim()
  },
  '(': '(',
  ')': ')',
  L_INFER: '<=',
  QUESTION_MARK: '?',
  NUMBER: {
    match: /(0|-?[1-9][0-9]*)/,
    value: s => Number(s)
  },
  STRING: {
    match: /"(?:\\["bfnrt\/\\]|\\u[a-fA-F0-9]{4}|[^"\\\n])*"/,
    value: s => JSON.parse(s).slice(1, -1)
  },
  IDENT: {
    match: /[a-zA-Z]+/,
    type: moo.keywords({
      KEYWORD: [
          'role', 'base', 'input',
          'init', 'terminal', 'goal',
          'next', 'legal', 'does',
          'distinct', 'true', 'or', 'and', 'not'
          // GDL-II
          'sees', 'random'
          // GDL-III
          'knows'
      ]
  })},
})
%}  # end of lexer definition

@lexer lexer

# First production rule is the default production for the parser.
input -> sentence+ {% id %}

# Nearley doesn't provide the _ and __ shorthands when using a lexer,
# but they are much more convenient and readable than %SPACE:* or %SPACE
# and during matching it allows us to fold them all into `null` as well.

# optional space
_ -> null | %SPACE {% d => null %}
# mandatory space
__ -> %SPACE {% d => null %}

# This production rule handles multiple sentences surrounded by optional space.
sentence+ ->
  # continuation if there are multiple sentences
  sentence+ _ sentence {% d => [ ...d[0], d[2] ] %}
  # reading only a single sentence
  | _ sentence {% d => [d[0]] %}

  # consume space with newlines
  | sentence+ _ "\n" {% d => d[0] %}
  # end of input, ignore remaining spaces
  | _ {% d => [] %}

sentence ->
    role_defn {% id %}
  | init_rule {% id %}
  | base_rule {% id %}
  | inference {% id %}
  | relation  {% id %}

# Role definition, a role relation appearing at the global scope.
role_defn -> "(" _ "role" __ name _ ")" {%
  d => ({
    type: "role_def",
    name: d[4],
    start: start(d[0]),
    end: end(d[6])
  })
%}

# A role relation can refer to a ground term or a variable.  When it is in the
# global scope it must contain a ground term, but as a relation it may retain a
# reference to a free variable, usually to constrain an inference.
role_relation -> "(" _ "role" __ var_or_name _ ")" {%
    d => ({
      type: "role_rel",
      name: roleName(d[4]),
      start: start(d[0]),
      end: end(d[6])
    })
%}

# An init rule defines a relation that holds true at the start of the game.
init_rule -> "(" _ "init" __ ground_relation _ ")" {%
    d => ({
      type: "init_rule",
      body: d[4],
      start: start(d[0]),
      end: end(d[6])
    })
%}

# A base rule indicates a term or relation of the domain without asserting its
# presence in the knowledge base.  It helps the solver determine the limits of
# knowledge, but typically these base rules could be inferred by the game rules.
base_rule -> "(" _ "base" __ relation _ ")" {%
    d => ({
      type: "base_rule",
      body: d[4],
      start: start(d[0]),
      end: end(d[6])
    })
%}

# This is a kind of right-to-left inference within a stratified and universally
# quantified logic as defined by GDL.  The expression is true when either
# head_expression is true or any body expressions are not true, or both.
inference -> "(" _ %L_INFER __ head_expression (__ body_expression):* _ ")" {%
    d => ({
      type: "inference",
      head: d[4],
      body: d[5],
      start: startOfToken(d[0]),
      end: end(d[7])
    })
%}

# Only certain game-independent relations will be in the head of an inference.
head_expression ->
    input_relation {% id %}
  | next_relation  {% id %}
  | base_rule      {% id %}
  | relation       {% id %}

body_expression ->
    does_relation {% id %}
  | role_relation {% id %}
  | relation      {% id %}

# An input describes a possible player action.  These are referred to by the
# `legal` and `does` relations and typically could be inferred from them.
input_relation -> "(" _ "input" __ var_or_name __ relation _ ")" {%
    d => ({
      type: "input",
      role: roleName(d[4]),
      action: d[6],
      start: start(d[0]),
      end: end(d[8])
    })
%}

# The indicated relation will hold true in the next state if the next relation
# is itself true (derived by the inference rule it is part of).
next_relation -> "(" _ "next" __ relation _ ")" {%
    d => ({
      type: "next",
      body: d[4],
      start: start(d[0]),
      end: end(d[6])
    })
%}

# The indicated player has done the action, either directly or as a result of
# the Game Master choosing a random action for the player to take.  This
# relation will typically appear in the body of an inference for a `next` rule.
does_relation -> "(" _ "does" __ var_or_name __ relation _ ")" {%
    d => ({
      type: "does",
      role: roleName(d[4]),
      action: d[6],
      start: start(d[0]),
      end: end(d[8])
    })
%}

# Indicates that it is legal for a player to take the specified action.
legality_rule -> "(" _ "legal" __ var_or_name __ relation _ ")" {%
    d => ({
      type: "legal",
      role: roleName(d[4]),
      action: d[6],
      start: start(d[0]),
      end: end(d[8])
    })
%}

relation+ ->
    _ relation (null | relation+)
      {% d => [d[1], ...d[3]] %}
  | _ "\n" relation+
      {% d => d[2] %}
  | _ relation _ "\n" relation+ {%
         d => [d[1], ...d[4]]
      %}

relation ->
    logical_relation {% id %}
  | legality_rule    {% id %}
  | goal_function    {% id %}
  | ground_term      {% id %}

logical_relation -> "(" _ logical_nary_op __ relation+ _ ")" {%
    d => ({
      type: d[2],
      body: [d[4], ...d[6]],
      start: start(d[0]),
      end: end(d[8])
    })
%} | "(" _ logical_unary_op __ relation _ ")" {%
    d => ({
      type: d[2],
      body: d[4],
      start: start(d[0]),
      end: end(d[6])
    })
%}

logical_nary_op -> ( "distinct" |  "or" | "and" ) {% id %}
logical_unary_op -> ( "true" | "not" ) {% id %}

goal_function -> "(" _ "goal" __ var_or_name __ %NUMBER _ ")" {%
    d => ({
      type: "goal",
      role: roleName(d[4]),
      value: d[6],
      start: start(d[0]),
      end: end(d[8])
    })
%}

# Variables are indicated by a `?` prefix.
var -> "?" name {%
  d => ({
    type: "variable",
    name: d[1],
    start: start(d[0]),
    end: d[1].end
  })
%}

# Either a variable or a symbolic name.
var_or_name -> var {% id %} | name {% id %}

ground_relation -> "(" _ name (__ ground_term+ {% d => d[1] %}):? _ ")" {%
    d => ({
      type: "relation",
      grounded: true,
      relation: d[2],
      parameters: d[3],
      start: start(d[0]),
      end: end(d[5])
    })
%}

# Produces a list of terms, all of which do not contain any free variables.
ground_term+ ->
    ground_term+ __ ground_term {%
        d => [...d[0], d[2]]
    %}
  | ground_term _ {% d => [d[1]] %}

# A ground term is one that is non-variable.
ground_term ->
    %NUMBER         {% id %}
  | terminal        {% id %}
  | name            {% id %}
  | ground_relation {% id %}

# Special GDL keyword that resolves to true when the game has terminated.
terminal -> "terminal" {% id %}

# Names may be represented as a symbol or a string, and there are equivalent
# values across those representations so we unify them here.
name -> %IDENT {%
    d => newName(d[0].value, start(d[0]), end(d[0]))
%}
name -> %STRING {%
    d => newName(d[0].value, start(d[0]), end(d[0]))
%}