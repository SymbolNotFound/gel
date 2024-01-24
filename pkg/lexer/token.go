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
// github:SymbolNotFound/ggdl/go/lexer/token.go

package lexer

// TokenType intrinsically defines the subtype of a Token and provides identifying methods.
type TokenType interface {
	// Returns a string representation of the type of this token.
	TypeString() string
	// Returns a string representation of this token, its syntactic image.
	Image() string
}

// Represents a Token instance by its position in the source and its type.
// The TokenType is an embedded interface (see above) and may be initialized
// with state/context or reuse a shared instance for the many tokens that are
// universally identical within their type (e.g. keywords, operator symbols).
// TokenPos is a 32-bit uint composite value defined in [token_pos.go].
type Token struct {
	TokenPos
	TokenType
}

// Identifier is a catch-all token for alpha-num strings that are not keywords.
func Identifier(name string, pos TokenPos) Token {
	return Token{pos, &identToken{name}}
}
func (ident *identToken) TypeString() string { return "IDENT" }
func (ident *identToken) Image() string      { return ident.name }

type identToken struct{ name string }

// EOF token indicates the end of the token stream.
// As EOF is not in the document, its TokenPos is always zero.
var EOF = Token{kTOKENPOS_ZERO, &eofToken{}}

type eofToken struct{}

// It is unnecessary to check against this value's type because we can compare
// directly against the single EOF instance.  The method is implemented so that
// it can conform to TokenType for the Token instance.
func (eof *eofToken) TypeString() string { return "EOF" }

// The image value is chosen based on the conventional value of an EOF signal,
// though there is no ASCII encoding for an EOF (26 or Ctrl-D), and in some
// environments such as text read mode in *nix, a zero value may be used.
// The actual value chosen is arbitrary but it seems a non-zero value would
// be less problematic, or not as surprisingly so, as including that would
// invalidate its status as a text file.  Ultimately, though, this value should
// never appear in an output stream, and if it does appear it should be oddly.
func (eof *eofToken) Image() string { return "\u001A" }

// An unexpected token is used when a parse error is encountered, despite there
// being no read errors encountered (those are returned with the NextToken call).
// An example would be incomplete Unicode bytes or a string without end quotes.
// Illegal tokens retain the image of the scan up to and including the bad char.
func UnexpectedToken(image string, pos TokenPos) Token {
	return Token{TokenPos: pos, TokenType: &unexpectedToken{image}}
}
func (oops *unexpectedToken) TypeString() string { return "UNEXPECTED" }
func (oops *unexpectedToken) Image() string      { return oops.image }

type unexpectedToken struct{ image string }

// Line comments are any sequence of characters beginning with a semicolon and
// extending until the next newline rune '\n'.
func LineComment(image string, pos TokenPos) Token {
	return Token{TokenPos: pos, TokenType: &comment{image}}
}
func (semi *comment) TypeString() string { return "COMMENT" }
func (semi *comment) Image() string      { return semi.image }

type comment struct{ image string }
