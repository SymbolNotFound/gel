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
// github:SymbolNotFound/ggdl/go/lexer/reader.go

package lexer

import (
	"io"
	"unicode"
)

// Public interface for reading a stream of tokens, sending them to a channel.
// See also ReadAll(reader) which provides a simpler interface for full reads.
type TokenReader interface {
	// Reads the next token, sending it to output, returning error or nil.  If an
	// io.EOF error was encountered it is returned here as well.
	NextToken() error

	// Read/Receive-only channel for Token values sent as being read from the input.
	// Calling NextToken() or ReadAll() will produce tokens on this channel and one
	// of those methods will close the channel when it encounters EOF. An EOF token
	// is also produced as the last token on the channel, so consumers can listen
	// for it specifically or listen until channel close using `for ... := range`.
	TokenReceiver() <-chan Token
}

// Constructor function for a lexer-based token reader.
func NewTokenReader(input io.RuneReader, output chan Token) TokenReader {
	return &lexerState{input, output, NewCursor()}
}

// Repeatedly calls `NextToken()` until either the enf of file (EOF) is reached
// or until an error is returned when attempting to read the next token.  Unlike
// NextToken(), it does not forward the io.EOF error - if `EOF` is reached and
// no other errors are encountered, this method returns `nil`.
func ReadAll(reader TokenReader) error {
	for {
		err := reader.NextToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Backing store for the TokenReader's state.  Alternative implementations of
// TokenReader are possible (e.g., reading from a token buffer, generating for
// token macros, mocking in tests, extending with modules, etc.) so the naming
// indicates the particular use of this state:
// a Lexer producing tokens (from input RuneReader, to output Token channel).
type lexerState struct {
	input  io.RuneReader
	output chan Token
	cursor Cursor
}

// Simply returns the already-created channel.
func (reader *lexerState) TokenReceiver() <-chan Token {
	return reader.output
}

// Checks the cursor's error/EOF status and closes the channel if an error is
// found.  Sends an EOF token before closing, when an io.EOF is encountered.
// If a non-nil value is passed as argument, that is sent before sending EOF.
func (reader *lexerState) consumeEOF(token *Token) {
	if !reader.cursor.HasError() {
		if token != nil && *token != EOF {
			reader.output <- *token
		}
		if reader.cursor.IsEOF() {
			reader.output <- EOF
		}
		close(reader.output)
	}
}

// Reads the next token, sending it to the Token chan, possibly closing it.
func (reader *lexerState) NextToken() error {
	if reader.cursor.HasError() {
		return reader.cursor.ErrorValue()
	}

	var r rune
	reader.cursor, r = reader.cursor.FirstRune(reader.input)
	if reader.cursor.HasError() {
		reader.consumeEOF(nil)
		return reader.cursor.ErrorValue()
	}

	var token Token
	switch {
	case unicode.IsPunct(r), unicode.IsSymbol(r):
		token = reader.readOperator()
	case unicode.IsDigit(r):
		token = reader.readNumber()
	case unicode.IsLetter(r):
		token = reader.readKeywordOrIdent()
	default:
		token = reader.consumeUnexpectedToken()
	}

	reader.consumeEOF(&token)
	return reader.cursor.ErrorValue()
}

// Consumes what remains in the cursor's buffer as an UnexpectedToken{...}.
func (reader *lexerState) consumeUnexpectedToken() Token {
	pos := reader.cursor.Pos()
	var image string
	reader.cursor, image = reader.cursor.ConsumeAll()
	return UnexpectedToken(image, pos)
}
