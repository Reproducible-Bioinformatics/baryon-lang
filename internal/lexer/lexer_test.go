package lexer

import (
	"testing"
)

func collectTokens(l *Lexer) []Token {
	tokens := []Token{}
	for token := range l.Token() {
		tokens = append(tokens, token)
		if token.Type == TOKEN_EOF {
			break
		}
	}
	return tokens
}

func TestLexer(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: `(identifier 'a' "string" 123 45.67
;comment
)`,
			expected: []Token{
				{Type: TOKEN_LPAREN, Literal: "("},
				{Type: TOKEN_IDENTIFIER, Literal: "identifier"},
				{Type: TOKEN_CHARACTER, Literal: "a"},
				{Type: TOKEN_STRING, Literal: "string"},
				{Type: TOKEN_NUMBER, Literal: "123"},
				{Type: TOKEN_NUMBER, Literal: "45.67"},
				{Type: TOKEN_COMMENT, Literal: "comment"},
				{Type: TOKEN_RPAREN, Literal: ")"},
				{Type: TOKEN_EOF},
			},
		},
	}

	for _, tt := range tests {
		lexer := New(tt.input)
		result := collectTokens(lexer)

		if len(result) != len(tt.expected) {
			t.Fatalf("wrong number of tokens: expected %d, got %d", len(tt.expected), len(result))
		}

		for i, tok := range tt.expected {
			if result[i].Type != tok.Type || result[i].Literal != tok.Literal {
				t.Errorf("unexpected token at %d: expected %+v, got %+v", i, tok, result[i])
			}
		}
	}
}
