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
// github:SymbolNotFound/ggdl/go/lexer/cursor_test.go

package lexer

import (
	"bufio"
	"io"
	"reflect"
	"strings"
	"testing"
)

func Test_cursorState_NextRune(t *testing.T) {
	type args struct {
		input      string
		skipSpaces bool
	}
	tests := []struct {
		name       string
		pos        TokenPos
		pending    []rune
		err        error
		args       args
		wantCursor Cursor
		wantRune   rune
	}{
		{"basic", NewTokenPos(1, 1), []rune{}, nil, args{"(", false},
			cursorState{NewTokenPos(1, 1), []rune{'('}, nil}, '('},
		{"basic skip", NewTokenPos(1, 1), []rune{}, nil, args{"(", true},
			cursorState{NewTokenPos(1, 1), []rune{'('}, nil}, '('},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := cursorState{
				pos:     tt.pos,
				pending: tt.pending,
				err:     tt.err,
			}
			input := bufio.NewReader(strings.NewReader(tt.args.input))
			gotCursor, gotRune := cursor.NextRune(input)
			if !reflect.DeepEqual(gotCursor, tt.wantCursor) {
				t.Errorf("cursorState.NextRune() got = %v, want %v", gotCursor, tt.wantCursor)
			}
			if gotRune != tt.wantRune {
				t.Errorf("cursorState.NextRune() got1 = %v, want %v", gotRune, tt.wantRune)
			}
		})
	}
}

func Test_cursorState_FirstRune(t *testing.T) {
	type fields struct {
		pos     TokenPos
		pending []rune
		err     error
	}
	type args struct {
		input io.RuneReader
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Cursor
		want1  rune
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := cursorState{
				pos:     tt.fields.pos,
				pending: tt.fields.pending,
				err:     tt.fields.err,
			}
			got, got1 := cursor.FirstRune(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cursorState.FirstRune() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("cursorState.FirstRune() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_cursorState_ConsumeAll(t *testing.T) {
	type fields struct {
		pos     TokenPos
		pending []rune
		err     error
	}
	tests := []struct {
		name   string
		fields fields
		want   Cursor
		want1  string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := cursorState{
				pos:     tt.fields.pos,
				pending: tt.fields.pending,
				err:     tt.fields.err,
			}
			got, got1 := cursor.ConsumeAll()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cursorState.ConsumeAll() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("cursorState.ConsumeAll() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_cursorState_ConsumeExceptFinal(t *testing.T) {
	type fields struct {
		pos     TokenPos
		pending []rune
		err     error
	}
	tests := []struct {
		name   string
		fields fields
		want   Cursor
		want1  string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := cursorState{
				pos:     tt.fields.pos,
				pending: tt.fields.pending,
				err:     tt.fields.err,
			}
			got, got1 := cursor.ConsumeExceptFinal()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cursorState.ConsumeExceptFinal() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("cursorState.ConsumeExceptFinal() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
