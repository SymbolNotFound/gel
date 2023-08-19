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
// github:SymbolNotFound/ggdl/go/lexer/symbols_test.go

package lexer

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestReadSymbol(t *testing.T) {
	var startTokenPos = NewTokenPos(3, 2).InSentence()
	tests := []struct {
		name        string
		pending     rune
		input       string
		want        Token
		wantPending string
		withEOF     bool
	}{
		{"open expression", '(', "(", ExpressionStart(startTokenPos), "", false},
		{"open expression then EOF", '(', "", ExpressionStart(startTokenPos), "", false},
		{"close expression", ')', "))", ExpressionEnd(startTokenPos), "", false},
		{"close expression then EOF", ')', "", ExpressionEnd(startTokenPos), "", false},
		{"question mark", '?', "", QuestionMark(startTokenPos), "", false},
		{"left double arrow", '<', "= ", LeftDoubleArrow(startTokenPos), "", false},
		{"left double arrow then EOF", '<', "=", LeftDoubleArrow(startTokenPos), "", false},
		{"left double arrow, partial", '<', "?", UnexpectedToken("<?", startTokenPos), "", false},
		{"left double arrow, partial with EOF", '<', "", UnexpectedToken("<", startTokenPos), "", true},
		{"left double arrow separated by space", '<', " =", UnexpectedToken("<", startTokenPos), " ", false},
		{"unexpected exclamation mark", '!', "", UnexpectedToken("!", startTokenPos), "", false},
	}

	for _, tt := range tests {
		// Creating the lexer directly to more easily set and read the internals.
		// See `plugg/tests` package for examples of using only the exported types.
		reader := lexerState{
			strings.NewReader(tt.input),
			make(chan Token),
			cursorState{
				startTokenPos,
				[]rune{ tt.pending },
				nil }}

		// Test the readUntilSymbol() procedure.
		t.Run(tt.name, func(t *testing.T) {
			tok := reader.readOperator()

			if (reader.cursor.HasError() && reader.cursor.IsEOF()) != tt.withEOF {
				t.Errorf("ConsumeSymbol() error = %v", reader.cursor.ErrorValue())
			}
			if (!reader.cursor.HasError() && tt.withEOF) {
				t.Errorf("ConsumeSymbol() should have EOF but has no error")
			}
			if !reflect.DeepEqual(tok, tt.want) {
				t.Errorf("ConsumeSymbol() returns %v, wanted %v", tok, tt.want)
			}
			var image string
			reader.cursor, image = reader.cursor.ConsumeAll()
			if !reflect.DeepEqual(image, string(tt.wantPending)) {
				t.Errorf("readSymbol() cursor image = %v, want %v",
					image, string(tt.wantPending))
			}
		})
	}
}

func TestReadNumber(t *testing.T) {
	var startTokenPos = NewTokenPos(1, 1).InSentence()
	tests := []struct {
		name        string
		pending     rune
		input       string
		want        Token
		wantPending []rune
		wantEOF     bool
	}{
		{"number then space", '1', "23 ",
			Integer("123", startTokenPos), []rune{' '}, false},
		{"number then EOF", '1', "23",
			Integer("123", startTokenPos), []rune{}, true},
		{"number then alpha", '1', "23a",
			Integer("123", startTokenPos), []rune{'a'}, false},
		{"single digit number", '1', "bc",
			Integer("1", startTokenPos), []rune{'b'}, false},
		{"nondigit pending value", 'a', "1c",
			UnexpectedToken("a", startTokenPos), []rune{}, false},
	}
	for _, tt := range tests {
		reader := lexerState{
			strings.NewReader(tt.input),
			make(chan Token),
			cursorState{
				startTokenPos,
				[]rune{tt.pending},
				nil,
			},
		}

		t.Run(tt.name, func(t *testing.T) {
			tok := reader.readNumber()
			err := reader.cursor.ErrorValue()
			if (err != nil) && err != io.EOF {
				t.Error("readNumber() error and not io.EOF")
				return
			}
			if err == io.EOF && !tt.wantEOF {
				t.Errorf("readNumber() error %s not io.EOF", err)
				return
			}
			fmt.Println(tt.want)
			if !reflect.DeepEqual(tok, tt.want) {
				t.Errorf("readNumber() = %v, want %v", tok, tt.want)
			}
			var image string
			reader.cursor, image = reader.cursor.ConsumeAll()
			if !reflect.DeepEqual(image, string(tt.wantPending)) {
				t.Errorf("readNumber() cursor image = %v, want %v",
					image, string(tt.wantPending))
			}
		})
	}
}
