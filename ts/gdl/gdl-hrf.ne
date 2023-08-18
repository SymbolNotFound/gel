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
# gdl-hrf.ne: grammar definition for Human-Readable Format of GDL.


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
    // Human-readable format GDL typically uses % for comments.
    match: /%+([^\n]*)/,
    value: s => s.substring(1).trim()
  },
  L_INFER: ":-",
  IMPLICATION: "=>", BICONDITIONAL: "<=>",
  "|": "|",
  "&": "&",
  "~": "~",
  "(": "(",  ")": ")",
  ",": ",",
  ACTION: "==>",
  NUMBER: {
    // Tokenizing only as far as the integers
    match: /(0|-?[1-9][0-9]*)/,
    value: s => Number(s)
  },
  VAR: {
    match: /(?[A-Za-z]|[A-Z])[0-9A-Z_a-z]*/,
    value: s => (s[0] == "?") ? s[1:] : s
  },
  IDENT: {
    match: /[a-z][0-9A-Z_a-z]*/,
    type: moo.keywords({
      KEYWORD: [
          'role',
          'base',
          'input',
          'init',
          'true',
          'does',
          'next',
          'legal', 
          'goal',
          'terminal', 
          'not',
          'or',
          'distinct',
          // GDL-II
          'sees',
          'random',
          // GDL-III
          'knows'
      ]
  })}
});
%} # end of token definitions

@lexer lexer

#-----------+
# SENTENCES |
#===========*

# Valid input is one or more sentences.  Typically a valid game will require
# several sentences to be complete and playable -- it must provide complete
# definitions for `role`, `input`, `base`, and `init`.

# Multiple sentences.  Note, left-recursion is faster for Earley.
input -> input _ sentence {% d => [ ...d[0], d[2] ] %}

# A single remaining sentence with leading spaces.
input -> _ sentence {% d => [ d[1] ] %}

# Trailing spaces and blank lines.
input -> input _ {% d => [ ...d[0] ] %}

# A sentence is an object constant or an inference or a ground relation.
sentence ->
    role_defn {% id %}
  | init_rule {% id %}
  | base_rule {% id %}
  | inference {% id %}
  | relation  {% id %}

#-----------+
# INFERENCE |
#===========*

# An inference consists of a head (the conclusion) and its body (premises),
# zero or more propositions which &'ed together determine the head's truth.
inference -> head_relation _ %L_INFER _ body_conjunction {%
    s => newInference(s[0], s[4], s[0].start, s[4][len(s[4])-1].end)
%}

# Only certain game-independent relations will be in the head of an inference.
head_relation ->
    input_relation {% id %}
  | next_relation  {% id %}
  | base_rule      {% id %}
  | proposition    {% id %}

# Only certain game-independent relations will be in the body of an inference.
body_relation ->
    does_relation {% id %}
  | role_var      {% id %}
  | proposition   {% id %}

# Multiple relations in the inference body.
body_conjunction -> body_conjunction _ "&" _ body_relation {%
    s => [ ...s[0], s[4] ]
%}

# One (remaining) relation in the inference body.
body_conjunction -> _ body_relation {% s => [ s[1] ] %}

# Trailing spaces and newline.
body_conjunction -> body_conjunction _ {% s => [ ...s[0] ] %}

#--------------+
# PLAYER ROLES |
#==============*

# `role(a)` means that `a` is a role in the game.  It is a complete sentence.
role_defn -> "role" _ "(" _ name _ ")" {%
    s => newRole(s[4].name, start(s[0]), end(s[6]))
%}

# `role(?p)` is a relation binding the variable `?p` to a valid role.  It is
# used to constrain variables to a role type and with `distinct` to separate
# acting roles from their opponents or their pieces from their opponents'.
# It typically appears within relations in the body (premises) of an inference.
role_var -> "role" _ "(" _ var _ ")" {%
    s => newRoleVar(s[5].name, start(s[0]), end(s[7]))
%}

#-------+
# FACTS |
#=======*

# `base(p)` means that `p` is a base proposition in the game's knowledge base.
# This describes the domain but does not assert the existence of any relations.
# It may be part of an inference to further constrain its proposition's terms.
base_rule -> "base" _ "(" _ proposition _ ")" {%
    s => newBaseRule(s[4], start(s[0]), end(s[6]))
%}

# `init(p)` means that the proposition `p` is true in the initial state.  There
# must not be any free variables in the relation within `init`.
init_rule -> "init" _ "(" _ ground_relation _ ")" {%
    s => newInitRule(s[4], start(s[0]), end(s[6]))
%}

# `true(p)` means that the proposition `p` is true in the current state.
true_function -> "true" _ "(" _ proposition _ ")" {%
    s => newUnaryOp('true', s[4], start(s[0]), end(s[6]))
%}

# `next(p)` means that the proposition `p` is true in the next state.
# This will typically appear in the head of an inference and must depend on
# `true` and `does` relations in its premises.
next_relation -> "next" _ "(" _ proposition _ ")" {%
    s => newNext(s[4], start(s[0]), end(s[6]))
%}

#---------+
# ACTIONS |
#=========*

# `input(r, a)` means that `a` is an action for role `r`.
input_relation -> "input" _ "(" _ name _ "," _ proposition _ ")" {%
    s => newAction('input', s[4], s[8], start(s[0]), end(s[10]))
%}

# `legal(r, a)` means it is legal for role `r` to play `a` in the current state.
# It must be defined in terms of input `true` relation(s).
legal_relation -> "legal" _ "(" _ name _ "," _ proposition _ ")" {%
    s => newAction('legal', s[4], s[8], start(s[0]), end(s[10]))
%}

# `does(r, a)` means that player `r` performs action `a` in the current state.
does_relation -> "does" _ "(" _ name _ "," _ proposition _ ")" {%
    s => newAction('does', s[4], s[8], start(s[0]), end(s[10]))
%}

#------------+
# COMPARISON |
#============*

logical_relation ->
    true_function     {% id %}
  | not_function      {% id %}
  | or_function       {% id %}
  | distinct_function {% id %}

# `not(p)` evaluates to true if proposition is not in KB, false if it is.
not_function -> "not" _ "(" _ proposition _ ")" {%
    s => newUnaryOp('not', s[4], start(s[0]), end(s[6]))
%}

# `or(a, b, ...)` evaluates to true if any of its parameters return true.
or_function -> "or" _ "(" _ relation _ "," param_list ")" {%
    s => newFunction('or', [s[4], ...s[7]], start(s[0]), end(s[8]))
%}

# `distinct(a, b)` evaluates to true if `a` and `b` are syntactically inequal.
# For three or more parameters, it is true when there are no equal members.  If
# there is zero or only one parameter, it is always true and can be elided.
distinct_function -> "distinct" _ "(" _ proposition _ "," _ proposition _ ")" {%
    s => newBinaryOp('distinct', s[4], s[8], start(s[0]), end(s[10]))
%}

#------------+
# RESOLUTION |
#============*

# `goal(r, u)` means that the current state has utility `u` for player `r`. 
# The inference it is a part of must have premises defined in terms of `true`.
goal_relation -> "goal" _ "(" _ name _ "," _ %NUMBER _ ")" {%
    s => newGoal(s[4], s[8], start(s[0]), end(s[10]))
%}

# `terminal` evaluates to true when the current state is a terminal state.
# The inference it is a part of must have premises defined in terms of true,
# or it may appear as a premise of a `goal` relation.
terminal_term -> "terminal" {%
  s => newTerminal(s[0])
%}

#-----------------+
# GROUND RELATION |
#=================*

ground_relation -> sym_name _ "(" _ ground_term+ _ ")" {%
    s => newGroundRelation(s[0], s[4], start(s[0]), end(s[6]))
%}

# Produces a list of terms, all of which do not contain any free variables.
ground_term+ ->
    ground_term+ _ "," _ ground_term {%
        s => [...s[0], s[2]]
    %}
  | _ ground_term {% s => [s[1]] %}
  | ground_term+ _ {% s => [...s[0]] %}

# A ground term is one that is non-variable.
ground_term ->
    %NUMBER         {% id %}
  | terminal_term   {% id %}
  | name            {% id %}
  | ground_relation {% id %}

#---------+
# GENERAL |
#=========*

# optional space
_ -> null | %SPACE {% s => null %}

# A symbol is when quoted terms are avoided, 
sym_name -> %IDENT {%
    s => newName(s[0].value, start(s[0]), end(s[0]))
%}

# A name may be a quoted string or an unquoted symbolic name (function_name).
name -> sym_name {% id %}
name -> %STRING {%
    s => newName(s[0].value, start(s[0]), end(s[0]))
%}

# A variable has its own namespace and represents a substitution for an atom.
var -> %VAR {%
    s => newVar(s[0].value, start(s[0]), end(s[1]))
%}

# Some productions expect either a var or a name but only a single term.
var_or_name -> var {% id %} | name {% id %}

proposition -> name _ "(" _ parameters _ ")" {%
    s => newFunction(s[0].name, s[4], s[0].start, end(s[6]))
%}

parameters ->
    param_list {% id %}
  | null {% () => [] %}

param_list -> param_list _ "," _ relation {%
    s => [...s[0], s[4]]
%}

param_list -> param_list _ {%
    s => [...s[0]]
%}

param_list -> _ relation {%
    s => [s[1]]
%}

relation ->
    logical_relation {% id %}
  | legal_relation   {% id %}
  | goal_relation    {% id %}
  | proposition      {% id %}
