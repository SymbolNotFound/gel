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
// github:SymbolNotFound/ggdl/go/lexer/fmt.go

package lexer

import "fmt"

// General implementation of the string conversion (i.e. for fmt interpolation).
// More specific Token types may override this String() function but the only
// operation that should make use of it is logging, for debugging & testing.
func (data Token) String() string {
	return fmt.Sprintf("%s <%s:%s>", data.TokenPos, data.TypeString(), data.Image())
}

// String conversion for the TokenPos value.  As a uint there was alerady a
// conversion available but the integer value obscures the actual position data.
func (data TokenPos) String() string {
	flag := data.flag()
	var flagchar byte = '?'
	switch flag {
	case kTOKENPOS_FLAG_COMMENT:
		flagchar = ';'
	case kTOKENPOS_FLAG_SENTENCE:
		flagchar = '.'
	case kTOKENPOS_FLAG_METAL:
		flagchar = '!'
	}

	return fmt.Sprintf("%c%d,%d", flagchar, data.Line(), data.Column())
}
