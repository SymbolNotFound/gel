// Copyright (c) 2023 Symbol Not Found L.L.C.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// github:SymbolNotFound/ggdl/go/lexer/symbols.go

package lexer

import (
	"unicode"
)

// Symbol tokens always have the same image, they can share a common instance.
type SymbolToken struct{}

const (
	RUNE_OPEN_PAREN    = '('
	RUNE_CLOSE_PAREN   = ')'
	RUNE_QUESTION_MARK = '?'
	RUNE_COMMENT_SEMI  = ';'

	RUNE_BEGIN_ARROW_LD = '<'
	IMAGE_ARROW_LD      = "<="
)

// Using the pending rune and (optionally) additional runes from input,
// reads a Symbol token.  This method only reads when an additional rune
// is expected, so any EOF errors propagated will not include a valid
// Token (but may include an IllegalToken token).
func (reader *lexerState) readOperator() Token {
	var r rune
	reader.cursor, r = reader.cursor.FirstRune(reader.input)
	var token Token

	switch r {
	case RUNE_OPEN_PAREN:
		cursor := reader.cursor
		reader.cursor, _ = cursor.ConsumeAll()
		token = ExpressionStart(cursor.Pos())

	case RUNE_CLOSE_PAREN:
		cursor := reader.cursor
		reader.cursor, _ = cursor.ConsumeAll()
		token = ExpressionEnd(cursor.Pos())

	case RUNE_QUESTION_MARK:
		cursor := reader.cursor
		reader.cursor, _ = cursor.ConsumeAll()
		token = QuestionMark(cursor.Pos())

	case RUNE_BEGIN_ARROW_LD: // looking for two runes, <=
		reader.cursor, r = reader.cursor.NextRune(reader.input)
		cursor := reader.cursor
		var image string
		if reader.cursor.HasError() {
			// In GDL the '<' symbol by itself means nothing,
			// return it as an IllegalToken if we see any error, including EOF.
			reader.cursor, image = cursor.ConsumeAll()
			token = UnexpectedToken(image, cursor.Pos())
		} else {
			if r == rune(IMAGE_ARROW_LD[1]) {
				reader.cursor, _ = cursor.ConsumeAll()
				token = LeftDoubleArrow(cursor.Pos())
			} else {
				if unicode.IsSpace(r) {
					reader.cursor, image = cursor.ConsumeExceptFinal()
				} else {
					reader.cursor, image = cursor.ConsumeAll()
				}
				token = UnexpectedToken(image, cursor.Pos())
			}
		}

	case RUNE_COMMENT_SEMI:
		cursor := reader.cursor
		for !reader.cursor.HasError() && r != '\n' {
			reader.cursor, r = reader.cursor.NextRune(reader.input)
		}
		var image string
		if reader.cursor.HasError() {
			reader.cursor, image = reader.cursor.ConsumeAll()
		} else {
			// No errors, exclude the last pending character, it is a '\n' newline.
			reader.cursor, image = reader.cursor.ConsumeExceptFinal()
			reader.cursor, _ = reader.cursor.ConsumeAll()
		}
		token = LineComment(image, cursor.Pos())

	default:
		cursor := reader.cursor
		var image string
		reader.cursor, image = cursor.ConsumeAll()
		token = UnexpectedToken(string(image), cursor.Pos())
	}

	return token
}

// Token EXPR_START = "("
type ExprStartToken struct{ SymbolToken }

// Begins all expressions, the main structural denotation in GDL syntax.
func ExpressionStart(pos TokenPos) Token {
	return Token{pos, &EXPR_START}
}
func (tok ExprStartToken) TypeString() string { return "OPEN_PAREN" }
func (tok ExprStartToken) Image() string      { return string(RUNE_OPEN_PAREN) }

var EXPR_START ExprStartToken

// Token EXPR_END = ")"
type ExprEndToken struct{ SymbolToken }

// Indicates the end of expressions and sub-expressions within a sentence.
func ExpressionEnd(pos TokenPos) Token {
	return Token{pos, &EXPR_END}
}
func (tok ExprEndToken) TypeString() string { return "CLOSE_PAREN" }
func (tok ExprEndToken) Image() string      { return string(RUNE_CLOSE_PAREN) }

var EXPR_END ExprEndToken

// Token ARROW_LD = "<="
type LDArrowToken struct{ SymbolToken }

// Used in constructing relations.
func LeftDoubleArrow(pos TokenPos) Token {
	return Token{pos, &ARROW_LD}
}
func (tok LDArrowToken) TypeString() string { return "ARROW_LD" }
func (tok LDArrowToken) Image() string      { return IMAGE_ARROW_LD }

var ARROW_LD LDArrowToken

// Token QUE_MARK = "?"
type QMarkToken struct{ SymbolToken }

// Used in the production rule for Variable terms.
func QuestionMark(pos TokenPos) Token {
	return Token{pos, &QUE_MARK}
}
func (tok QMarkToken) TypeString() string { return "QUE_MARK" }
func (tok QMarkToken) Image() string      { return string(RUNE_QUESTION_MARK) }

var QUE_MARK QMarkToken

// Read the pending runes and into reader.input as needed, to read an Integer.
func (reader *lexerState) readNumber() Token {
	var r rune
	reader.cursor, r = reader.cursor.FirstRune(reader.input)

	// This method should only be called if the first digit rune has already been
	// peeked at or if the grammar would require the next token to be an Integer,
	// so return an UnexpectedToken if that is not the case.
	if !unicode.IsDigit(r) {
		return reader.consumeUnexpectedToken()
	}

	for !reader.cursor.HasError() && unicode.IsDigit(r) {
		reader.cursor, r = reader.cursor.NextRune(reader.input)
	}
	cursor := reader.cursor
	var image string
	if cursor.HasError() {
		// Produce the Integer token and maintain the error status in the cursor.
		reader.cursor, image = cursor.ConsumeAll()
		return Integer(image, cursor.Pos())
	}

	reader.cursor, image = cursor.ConsumeExceptFinal()
	return Integer(image, cursor.Pos())
}

// A token representing a sequence of digits (an integer numeral).
type integerToken struct{ image string }

// More complex numeric types can be constructed from sequences of unsigned
// integers and punctuation.  This also keeps the tokenizer state management
// simpler by defining negatives, floats, etc. in terms of production rule
// semantics.  GDL and GDL-II both only assume integer constants in [0-100].
func Integer(image string, pos TokenPos) Token {
	return Token{pos, &integerToken{image}}
}
func (num *integerToken) TypeString() string { return "INTEGER" }
func (num *integerToken) Image() string      { return num.image }
