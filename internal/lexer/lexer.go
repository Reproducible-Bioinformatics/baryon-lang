package lexer

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

type TokenType int

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF
	TOKEN_LPAREN     // (
	TOKEN_RPAREN     // )
	TOKEN_IDENTIFIER // bala, enrichment_analysis, etc.
	TOKEN_STRING     // "string content"
	TOKEN_NUMBER     // 123, 45.67
	TOKEN_CHARACTER  // 'a'
	TOKEN_COMMENT    // ; comment
)

// Creates a new Lexer.
func New(input string) *Lexer {
	return &Lexer{
		input:        input,
		position:     0,
		readPosition: 0,
		ch:           0,
		line:         1,
		column:       0,
	}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition++
	l.column++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readString() string {
	position := l.position + 1 // Skip the opening quote
	for {
		l.readChar()
		if l.ch == '"' || l.ch == '\'' || l.ch == 0 {
			break
		}
		// Handle escape sequences
		if l.ch == '\\' && (l.peekChar() == '"' || l.peekChar() == '\'') {
			l.readChar()
		}
	}
	if l.ch == '"' || l.ch == '\'' {
		l.readChar()
		return l.input[position : l.position-1]
	}
	return l.input[position:l.position]
}

func (l *Lexer) readComment() string {
	position := l.position + 1 // Skip the semicolon
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) Token(yield func(Token) bool) {
	l.skipWhitespace()

	tok := Token{Line: l.line, Column: l.column}

	switch l.ch {
	case '(':
		tok.Type, tok.Literal = TOKEN_LPAREN, "("
	case ')':
		tok.Type, tok.Literal = TOKEN_RPAREN, ")"
	case '"':
		tok.Type, tok.Literal = TOKEN_STRING, l.readString()
	case '\'':
		tok.Type, tok.Literal = TOKEN_CHARACTER, l.readString()
	case ';':
		tok.Type, tok.Literal = TOKEN_COMMENT, l.readComment()
	case 0:
		tok.Type, tok.Literal = TOKEN_EOF, ""
	default:
		if isLetter(l.ch) {
			tok.Type, tok.Literal = TOKEN_IDENTIFIER, l.readIdentifier()
		} else if isDigit(l.ch) {
			tok.Type, tok.Literal = TOKEN_NUMBER, l.readNumber()
		} else {
			tok.Type, tok.Literal = TOKEN_ILLEGAL, string(l.ch)
		}
	}

	l.readChar()

	yield(tok)
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
