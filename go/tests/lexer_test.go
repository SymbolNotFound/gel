package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/SymbolNotFound/ggdl/go/lexer"
)

func TestTokenize(t *testing.T) {
	var input string = "(<= (next ?x fun)\n  (does ?x ?y) (fun ?p ?y))" +
		" ;; does arity oddity \"intended\n( fun 69 gøød ) \n"

	expected := []lexer.Token{
		lexer.ExpressionStart(lexer.NewTokenPos(1, 1).InSentence()),
		lexer.LeftDoubleArrow(lexer.NewTokenPos(1, 2).InSentence()),
		lexer.ExpressionStart(lexer.NewTokenPos(1, 5).InSentence()),
		lexer.KeywordAt("next", lexer.NewTokenPos(1, 6).InSentence()),
		lexer.QuestionMark(lexer.NewTokenPos(1, 11).InSentence()),
		lexer.Identifier("x", lexer.NewTokenPos(1, 12).InSentence()),
		lexer.Identifier("fun", lexer.NewTokenPos(1, 14).InSentence()),
		lexer.ExpressionEnd(lexer.NewTokenPos(1, 17).InSentence()),
		lexer.ExpressionStart(lexer.NewTokenPos(2, 3).InSentence()),
		lexer.KeywordAt("does", lexer.NewTokenPos(2, 4).InSentence()),
		lexer.QuestionMark(lexer.NewTokenPos(2, 9).InSentence()),
		lexer.Identifier("x", lexer.NewTokenPos(2, 10).InSentence()),
		lexer.QuestionMark(lexer.NewTokenPos(2, 12).InSentence()),
		lexer.Identifier("y", lexer.NewTokenPos(2, 13).InSentence()),
		lexer.ExpressionEnd(lexer.NewTokenPos(2, 14).InSentence()),
		lexer.ExpressionStart(lexer.NewTokenPos(2, 16).InSentence()),
		lexer.Identifier("fun", lexer.NewTokenPos(2, 17).InSentence()),
		lexer.QuestionMark(lexer.NewTokenPos(2, 21).InSentence()),
		lexer.Identifier("p", lexer.NewTokenPos(2, 22).InSentence()),
		lexer.QuestionMark(lexer.NewTokenPos(2, 24).InSentence()),
		lexer.Identifier("y", lexer.NewTokenPos(2, 25).InSentence()),
		lexer.ExpressionEnd(lexer.NewTokenPos(2, 26).InSentence()),
		lexer.ExpressionEnd(lexer.NewTokenPos(2, 27).InSentence()),
		lexer.LineComment(";; does arity oddity \"intended",
			lexer.NewTokenPos(2, 29).InComment()),
		lexer.ExpressionStart(lexer.NewTokenPos(3, 1).InSentence()),
		lexer.Identifier("fun", lexer.NewTokenPos(3, 3).InSentence()),
		lexer.Integer("69", lexer.NewTokenPos(3, 7).InSentence()),
		lexer.Identifier("gøød", lexer.NewTokenPos(3, 10).InSentence()),
		lexer.ExpressionEnd(lexer.NewTokenPos(3, 15).InSentence()),
	}

	tokens := tokenizeString(input)
	for i, tt := range expected {
		if len(tokens) <= i {
			t.Fatalf("Too few tokens received from parsing, expected %s", tt.Image())
			break
		}
		if tt.TypeString() != tokens[i].TypeString() {
			t.Fatalf("Got %s, expected %s for token type", tokens[i].TypeString(), tt.TypeString())
		}
		if tt.Image() != tokens[i].Image() {
			t.Fatalf("Got %s, expected %s for token image", tokens[i].Image(), tt.Image())
		}
		if tt.Line() != tokens[i].Line() || tt.Column() != tokens[i].Column() {
			t.Fatalf("token position mismatch %d, %d expected %d, %d", tokens[i].Line(), tokens[i].Column(), tt.Line(), tt.Column())
		}
	}
}

func tokenizeString(input string) []lexer.Token {
	stringReader := strings.NewReader(input)
	output := make(chan lexer.Token)

	reader := lexer.NewTokenReader(stringReader, output)
	tokens := make([]lexer.Token, 0, 32)
	go func() {
		for token := range reader.TokenReceiver() {
			tokens = append(tokens, token)
			fmt.Println(token)
		}
	}()

	err := lexer.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return tokens
}

type bogusReader struct{
	count *int
	send  error
}
var bogusreadererrorstring string = "dowhatnow?"
func (br bogusReader) ReadRune() (rune, int, error) {
	if *br.count > 1 {
		(*br.count)--
		return 'a', 1, nil
	}
	return '(', 1,  br.send
}

// Test io.Reader that errors before EOF
func TestTokenize_Errors(t *testing.T) {
	var i int = 3
	br := bogusReader{&i, fmt.Errorf(bogusreadererrorstring)}
	output := make(chan lexer.Token)
	reader := lexer.NewTokenReader(br, output)
	tokens := make([]lexer.Token, 0, 32)
	go func() {
		for token := range reader.TokenReceiver() {
			tokens = append(tokens, token)
			fmt.Println(token)
		}
	}()

	err := lexer.ReadAll(reader)
	if err == nil {
		t.Errorf("Reader that returns nil should route its error through ReadAll")
	}
	if err != nil && err.Error() != bogusreadererrorstring {
		t.Errorf("Got a different error than expected from bogusReader: %s", err)
	}
}
