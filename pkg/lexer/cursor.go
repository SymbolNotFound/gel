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
// github:SymbolNotFound/ggdl/go/lexer/cursor.go

package lexer

import (
	"io"
	"unicode"
)

// The Cursor represents a few properties of the lexer's state that are
// invariably coupled to each other -- token position, the runes ready to
// be integrated into the next token, and whether there is a pending rune
// waiting to be processed.  The token's next position should always be
// the current position plus the size of the pending rune, if there is a
// pending rune, but that relies on whether scanning can be done in LL(1)
// or (in some cases) LL(0) as with `(` and `)`.  It also smelled bad to be
// updating only part of the lexer state and another part that depended on
// it, while not doing so atomically.
//
// This, and its backing struct, are a solution to the above problems while
// also aiding the readability of the token-specific lexer code.  The coupled
// updates are done within the Advance and Consume methods, there is no
// redundant next pos or ambiguity about the contents of the pending image.
// In addition to that, the cursor is copy-on-write, all updates are conveyed
// by the return value of the updating method, and the implementing methods
// use by-value receivers so downcast-and-update has limited adverse effect.
//
// However, it assumes that it is the only reader on the provided input, and
// that its scan position is consistent between calls to Advance.  If there
// is need of multiple concurrent cursors on the same reader source, use new
// readers for each cursor or tee the source RunReader.  Rather than further
// complicate this code with management of byte offsets and seeks at each read,
// especially while this task of tokenizing byte streams is inherently single-
// threaded.  Calling code is expected to manage it, typically via lexerState.
type Cursor interface {
	// NextRune is called to extend the cursor by reading the next rune from
	// input.  Also updates the pending string except when skipping spaces, and
	// returns the updated cursor and the rune that was read.
	NextRune(input io.RuneReader) (Cursor, rune)
	// Similar to NextRune() but will read from the pending buffer if nonempty,
	// and read from input, populating pending, if pending was empty.  Implicitly
	// ignores leading spaces if needing to read from input.
	FirstRune(input io.RuneReader) (Cursor, rune)

	// Consumes all characters in the pending rune list, updating pos to match.
	ConsumeAll() (Cursor, string)
	// Same as ConsumeAll() except the last rune is left in the pending buffer.
	ConsumeExceptFinal() (Cursor, string)

	// Resets the TokenPos for this cursor to (0, 0, UNKNOWN).
	ResetPos() Cursor

	// The current position of the next Token that would be produced by consuming
	// the contents of this Cursor, whether or not anything is in pending buffer.
	Pos() TokenPos

	// Returns true if there is nothing pending in the cursor.
	IsEmpty() bool

	// Returns true if the last ReadRune call returned an error.
	HasError() bool
	// Returns `true` if the embedded error is io.EOF.
	IsEOF() bool
	// Returns the error (or nil) from the most recent read of input.  If an error
	// is encountered, it will persist through update methods and prohibit reads.
	//
	// Intentionally not extending `error` interface by naming this ErrorValue.
	ErrorValue() error
}

func NewCursor() Cursor {
	// Token position (0, 0) is used for unknown, and position (1, 1) is for the
	return cursorState{kTOKENPOS_ZERO, []rune{}, nil}
}

// Internal representation of the cursor state, its minimal required representation.
//
// Notably, the methods on this are by-value not by-pointer receiver.  There are
// copy-on-write intrinsics on the mutating methods (Next, First and the Consumes)
// and the remaining methods are readonly getters.
type cursorState struct {
	pos     TokenPos
	pending []rune
	err     error
}

func (cursor cursorState) Pos() TokenPos {
	return cursor.pos
}

// Arbitrary size for \t alignment.
const CURSOR_TAB_STOP = 4

// Returns the lines, columns offset for the runes in th pending buffer.
func offset(runes []rune) (uint, uint) {
	if len(runes) == 0 {
		return 0, 0
	}
	lines, cols := uint(0), uint(0)
	for _, r := range runes {
		if unicode.IsPrint(r) {
			cols += 1
		}
		if r == '\t' {
			cols = (cols + CURSOR_TAB_STOP)
			cols -= cols % CURSOR_TAB_STOP
		}
		if r == '\n' {
			lines, cols = lines+1, 1
		}
	}
	return lines, cols
}

func (cursor cursorState) IsEmpty() bool {
	return len(cursor.pending) == 0
}

func (cursor cursorState) IsEOF() bool {
	return cursor.err == io.EOF
}

func (cursor cursorState) HasError() bool {
	return cursor.err != nil
}

func (cursor cursorState) ErrorValue() error {
	return cursor.err
}

// NextRune is called to extend the cursor by reading the next rune from input.
func (cursor cursorState) NextRune(input io.RuneReader) (Cursor, rune) {
	if cursor.HasError() {
		return cursor, rune(0)
	}
	r, _, err := input.ReadRune()
	if err != nil {
		return cursorState{cursor.pos, cursor.pending, err}, r
	}

	return cursorState{cursor.pos, append(cursor.pending, r), err}, r
}

// FirstRune is called to get the first pending rune,
// filling in the pending buffer from input if needed.
func (cursor cursorState) FirstRune(input io.RuneReader) (Cursor, rune) {
	var r rune
	var next Cursor = cursor
	if cursor.Pos() == kTOKENPOS_ZERO {
		next = cursorState{cursor.Pos().NextAt(1, 1), cursor.pending, cursor.err}
	}
	if len(cursor.pending) > 0 {
		if len(cursor.pending) > 1 {
			// TODO: for safety, we should buffer additional entries from pending into
			// input, but currently we only ever have zero or one runes in the buffer.
			panic("unexpected FirstRune() call with more than one rune in pending buffer.")
		}
		r = cursor.pending[0]
	} else {
		next, r = next.NextRune(input)
	}

	if unicode.IsSpace(r) {
		for !next.HasError() && unicode.IsSpace(r) {
			next, r = next.NextRune(input)
		}
		if next.HasError() {
			next, _ = next.ConsumeAll()
			next.ResetPos()
		} else {
			next, _ = next.ConsumeExceptFinal()
		}
	}

	return next, r
}

func (cursor cursorState) ConsumeAll() (Cursor, string) {
	nextPos := cursor.pos.NextAt(offset(cursor.pending))
	return cursorState{nextPos, []rune{}, cursor.err}, string(cursor.pending)
}

func (cursor cursorState) ConsumeExceptFinal() (Cursor, string) {
	runeCount := len(cursor.pending)
	image, finalRune := cursor.pending[:runeCount-1], cursor.pending[runeCount-1]
	nextPos := cursor.pos.NextAt(offset(image))

	return cursorState{nextPos, []rune{finalRune}, cursor.err}, string(image)
}

func (cursor cursorState) ResetPos() Cursor {
	next := cursor
	next.pos = kTOKENPOS_ZERO
	return next
}