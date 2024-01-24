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
// github:SymbolNotFound/ggdl/go/lexer/token_pos.go

package lexer

//	TokenPos encoded as 32-bit uint:
//
// .LLLLLLLLLLLLLLLLLLLLCCCCCCCCCCFF.
// :[++++++++++++++++++]            :
// :        20 bits LINE            :
// :                    [++++++++]  :
// :                10 bits COLUMN  :
// :                              []:
// :                    2 bits FLAGS:
// :                                :
// `10987654321098765432109876543210'
//
// Use Line(), Column() and Next*() methods to read and update values.
type TokenPos uint32

const (
	// The zero value is (unknown line, unknown column, FLAG_UNK).
	kTOKENPOS_ZERO = TokenPos(0)

	// Actual lines and columns are 1-indexed, with max value ~0 (i.e., *_MASK).
	// The bit sizes were chosen to give ~ 1000 columns and two flag bits, with
	// the remainder going to the line number.  columns should rarely exceed the
	// alloted size; if they do, following lines will still order correctly.
	kTOKENPOS_LINE_UNKNOWN   uint = 0
	kTOKENPOS_COLUMN_UNKNOWN uint = 0

	kTOKENPOS_LINE_OFFSET   = 12
	kTOKENPOS_LINE_MASK     = (1 << 20) - 1
	kTOKENPOS_COLUMN_OFFSET = 2
	kTOKENPOS_COLUMN_MASK   = (1 << 10) - 1

	kTOKENPOS_FLAG_MASK byte = 0b11
	// The position has unknown flag: recently created, or a newline encountered
	kTOKENPOS_FLAG_UNK byte = 0b00
	// The token associated with this position is a line or block comment.
	kTOKENPOS_FLAG_COMMENT byte = 0b01
	// The token associated with this position is part of a GDL sentence.
	kTOKENPOS_FLAG_SENTENCE byte = 0b10
	// The token associated with this position is in a metalanguage, not GDL.
	kTOKENPOS_FLAG_METAL byte = 0b11
)

func NewTokenPos(line, col uint) TokenPos {
	return rawTokenPos(line, col, kTOKENPOS_FLAG_UNK)
}

// Convenience method for package-internal use, caps line and col values at max.
func rawTokenPos(line, col uint, flag byte) TokenPos {
	if line >= kTOKENPOS_LINE_MASK {
		// Maintains TokenPos-as-int monotonicity w.r.t. actual byte ordering.
		line, col = kTOKENPOS_LINE_MASK, kTOKENPOS_COLUMN_UNKNOWN
	}
	if col >= kTOKENPOS_COLUMN_MASK {
		// Maintain line numbering when exceeding the possible column number.
		col = kTOKENPOS_COLUMN_MASK
	}
	line_bits := uint32(line << kTOKENPOS_LINE_OFFSET)
	col_bits := uint32(col << kTOKENPOS_COLUMN_OFFSET)
	flag_bits := uint32(flag & kTOKENPOS_FLAG_MASK)
	return TokenPos(line_bits | col_bits | flag_bits)
}

// Returns the 1-indexed line number of the position, zero means unknwno.
// Token embeds this from TokenPos interface to adopt ts Line() method.
func (pos TokenPos) Line() uint {
	return uint(pos>>kTOKENPOS_LINE_OFFSET) & kTOKENPOS_LINE_MASK
}

// Returns the 1-indexed column number of the position, zero means unknown.
// Token embeds this from TokenPos interface to adopt its Column() method.
func (pos TokenPos) Column() uint {
	return uint(pos>>kTOKENPOS_COLUMN_OFFSET) & kTOKENPOS_COLUMN_MASK
}

// Increments the position to its next line, resetting the column as well.
// Flag's current value is reset from comment mode, retained otherwise.
func (pos TokenPos) NextLine() TokenPos {
	flag := pos.flag()
	if flag == kTOKENPOS_FLAG_COMMENT {
		flag = kTOKENPOS_FLAG_UNK
	}
	return rawTokenPos(pos.Line()+1, 0, flag)
}

// Increments the column, keeping the current flag.
func (pos TokenPos) NextCol() TokenPos {
	return rawTokenPos(pos.Line(), pos.Column()+1, pos.flag())
}

// Increments by number of lines then by number of columns.
func (pos TokenPos) NextAt(lines, cols uint) TokenPos {
	var next TokenPos = pos
	for ; lines > 0; lines-- {
		next = next.NextLine()
	}
	for ; cols > 0; cols-- {
		next = next.NextCol()
	}
	return next
}

// Returns the current flag value, as a byte.
func (pos TokenPos) flag() byte {
	return byte(pos) & kTOKENPOS_FLAG_MASK
}

// Produces the same Token position, ensuring its flag is set to SENTENCE mode.
func (pos TokenPos) InSentence() TokenPos {
	if pos.flag() == kTOKENPOS_FLAG_SENTENCE {
		return pos
	}
	return pos - TokenPos(pos.flag()) +
		TokenPos(kTOKENPOS_FLAG_SENTENCE)
}

// Produces the same Token position, ensuring its flag is set to COMMENT mode.
func (pos TokenPos) InComment() TokenPos {
	if pos.flag() == kTOKENPOS_FLAG_COMMENT {
		return pos
	}
	return pos - TokenPos(pos.flag()) +
		TokenPos(kTOKENPOS_FLAG_COMMENT)
}

// Produces the same Token position, ensuring its flag is set to COMMENT mode.
func (pos TokenPos) InMetaBlock() TokenPos {
	if pos.flag() == kTOKENPOS_FLAG_METAL {
		return pos
	}
	return pos - TokenPos(pos.flag()) +
		TokenPos(kTOKENPOS_FLAG_METAL)
}

// Resets the flag value to unknown.
func (pos TokenPos) ResetFlag() TokenPos {
	if pos.flag() == kTOKENPOS_FLAG_UNK {
		return pos
	}
	return pos - TokenPos(pos.flag())
}

func IsSentence(pos TokenPos) bool {
	return pos.flag() == kTOKENPOS_FLAG_SENTENCE
}

func IsComment(pos TokenPos) bool {
	return pos.flag() == kTOKENPOS_FLAG_COMMENT
}

func IsMetaBlock(pos TokenPos) bool {
	return pos.flag() == kTOKENPOS_FLAG_METAL
}
