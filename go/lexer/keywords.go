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
// github:SymbolNotFound/ggdl/go/lexer/keywords.go

package lexer

import "unicode"

func (reader *lexerState) readKeywordOrIdent() Token {
	var r rune
	reader.cursor, r = reader.cursor.FirstRune(reader.input)
	for !reader.cursor.HasError() && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
		reader.cursor, r = reader.cursor.NextRune(reader.input)
	}
	cursor := reader.cursor
	var image string
	if cursor.HasError() {
		reader.cursor, image = cursor.ConsumeAll()
	} else {
		reader.cursor, image = cursor.ConsumeExceptFinal()
	}
	var token Token = KeywordAt(image, cursor.Pos())

	return token
}

// All keywords are given the KEYWORD token type.
type KeywordToken struct {
	 image string
}

// Satisfies the requirement for TokenType interface.
func (tok KeywordToken) TypeString() string {
  return "KEYWORD"
}

// Satisfies the requirement for TokenType interface.
func (tok KeywordToken) Image() string {
	return tok.image
}

// Constructs a Token instance pointing to the singular KeywordToken instance
// for the specific keyword.
func (tok KeywordToken) At(pos TokenPos) Token {
	return Token{pos, tok}
}

var keywords map[string]KeywordToken = make(map[string]KeywordToken)

func defineKeyword(image string) { keywords[image] = KeywordToken{image} }
func init() {
	// Special relations defined for the semantics of GDL.
	defineKeyword("role")
	defineKeyword("legal")
	defineKeyword("next")
	defineKeyword("does")
	defineKeyword("goal")
	defineKeyword("terminal")

	// Additional special relations for GDL-II.
	defineKeyword("sees")
	defineKeyword("random")

	// These could be inferred or defined in terms of other rules.
	defineKeyword("init")
	defineKeyword("input")

	// Some boolean relations that GDL assumes existence of.
	defineKeyword("true")
	defineKeyword("or")
	defineKeyword("and")
	defineKeyword("not")
}

func KeywordAt(image string, pos TokenPos) Token {
	if keyword, ok := keywords[image]; ok {
		return keyword.At(pos)
	}
	return Identifier(image, pos)
}