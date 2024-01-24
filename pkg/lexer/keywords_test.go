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
// github:SymbolNotFound/ggdl/go/lexer/keywords_test.go

package lexer

import (
	"reflect"
	"strings"
	"testing"
)

func Test_lexerState_readKeywordOrIdent(t *testing.T) {
	startPos := NewTokenPos(1, 2).InSentence()
	tests := []struct {
		input  string
		pos    TokenPos
		want   Token
	}{
		{"role", startPos, keywords["role"].At(startPos)},
		{"roles", startPos, Token{startPos, &identToken{"roles"}}},
		{" role", startPos, keywords["role"].At(startPos.NextCol())},
		{" roles", startPos, Token{startPos.NextCol(), &identToken{"roles"}}},
		{"srole", startPos, Token{startPos, &identToken{"srole"}}},
		{"init", startPos, KeywordAt("init", startPos)},
		{"input", startPos, keywords["input"].At(startPos)},
		{"legal", startPos, KeywordAt("legal", startPos)},
		{"next", startPos, keywords["next"].At(startPos)},
		{"does", startPos, keywords["does"].At(startPos)},
		{"true", startPos, keywords["true"].At(startPos)},
		{"or", startPos, keywords["or"].At(startPos)},
		{"and", startPos, keywords["and"].At(startPos)},
		{"not", startPos, keywords["not"].At(startPos)},
		{"goal", startPos, keywords["goal"].At(startPos)},
		{"terminal", startPos, keywords["terminal"].At(startPos)},
		{"p&?q", startPos, Token{startPos, &identToken{"p"}}},
		{"ps&?q", startPos, Token{startPos, &identToken{"ps"}}},
		{"sees", startPos, keywords["sees"].At(startPos)},
		{"random", startPos, keywords["random"].At(startPos)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			reader := &lexerState{
				input:  strings.NewReader(tt.input),
				output: make(chan Token),
				cursor: cursorState{tt.pos, []rune{}, nil},
			}
			if got := reader.readKeywordOrIdent(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lexerState.readKeywordOrIdent() = %v, want %v", got, tt.want)
			}
		})
	}
}
