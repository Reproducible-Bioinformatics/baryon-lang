package lexer

import (
	"iter"
	"strings"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number
	column       int  // current column number
}

type TokenType int

const (
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF
	TOKEN_LPAREN     // (
	TOKEN_RPAREN     // )
	TOKEN_IDENTIFIER // bala, enrichment_analysis, string, _, etc.
	TOKEN_STRING     // "string content"
	TOKEN_NUMBER     // 123, 45.67
	TOKEN_CHARACTER  // 'a' - Note: example uses "character" as type, not literal
	TOKEN_COMMENT    // ; comment
)

var tokenStrings = [...]string{
	TOKEN_ILLEGAL:    "ILLEGAL",
	TOKEN_EOF:        "EOF",
	TOKEN_LPAREN:     "LPAREN",
	TOKEN_RPAREN:     "RPAREN",
	TOKEN_IDENTIFIER: "IDENTIFIER",
	TOKEN_STRING:     "STRING",
	TOKEN_NUMBER:     "NUMBER",
	TOKEN_CHARACTER:  "CHARACTER",
	TOKEN_COMMENT:    "COMMENT",
}

func (tt TokenType) String() string {
	if tt >= 0 && int(tt) < len(tokenStrings) {
		return tokenStrings[tt]
	}
	return "UNKNOWN"
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Creates a new Lexer.
func New(input string) *Lexer {
	lexer := &Lexer{
		input: input,
		line:  1,
	}
	lexer.readChar() // Initialize ch, position, readPosition, column
	return lexer
}

// readChar reads the next character and advances the position.
func (l *Lexer) readChar() {
	startColumn := l.column
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0 // Reset column after newline
	} else {
		// Only increment column if it's not a newline
		l.column++
	}
	// Handle potential '\r\n' - if we just read \r, peek for \n
	if l.ch == '\r' && l.peekChar() == '\n' {
		l.readChar() // Consume the \n, readChar handles line/col update
	} else if l.ch == '\r' { // Handle standalone \r as newline
		l.line++
		l.column = 0
	}

	// If column reset due to newline, ensure it starts at 1 for the next char
	if startColumn > 0 && l.column == 0 {
		l.column = 1
	} else if startColumn == 0 && l.column == 0 && l.ch != 0 {
		// Initial character or after newline
		l.column = 1
	}
}

// peekChar looks ahead without consuming the character.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// skipWhitespace skips spaces, tabs, and newlines/carriage returns.
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readString reads a string literal enclosed in double or single quotes.
// It handles basic escape sequences for the quote character itself.
func (l *Lexer) readString(quoteType byte) string {
	position := l.position + 1 // Skip the opening quote
	var sb strings.Builder
	for {
		prevCh := l.ch
		l.readChar()
		if l.ch == quoteType {
			// Check for escaped quote
			if prevCh == '\\' {
				// This means we have an escaped quote, continue reading
				currentContent := sb.String()
				if len(currentContent) > 0 {
					sb.Reset()
					sb.WriteString(currentContent[:len(currentContent)-1])
				}
				sb.WriteByte(quoteType) // Add the actual quote char
				continue
			}
			// End of string found
			break
		}
		if l.ch == 0 { // EOF before closing quote
			break
		}
		sb.WriteByte(l.ch)
		// if we see \, peek next. If it's a quote, skip \
		// TODO: More complex escapes (\n, \t etc.) are not handled here but could be added.
		if l.ch == '\\' && l.peekChar() == quoteType {
			// Read the escaped quote in the next iteration
			l.readChar()
		}
	}
	// TODO: remove the escape characters (\).
	// A more robust implementation would build the string char by char, handling escapes.
	str := l.input[position:l.position]
	// TODO: Add proper escape sequence processing if needed. For now, return raw content.
	if l.ch == quoteType {
		l.readChar() // Consume the closing quote
	}

	return str
}

// readComment reads from ';' to the end of the line.
func (l *Lexer) readComment() string {
	position := l.position + 1 // Skip the semicolon
	for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position] // Excludes the newline
}

// readIdentifier reads a sequence of letters, digits, or underscores.
func (l *Lexer) readIdentifier() string {
	position := l.position
	// Allow leading underscore
	if isLetter(l.ch) || l.ch == '_' {
		l.readChar()
		for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}
	// Backtrack one char because the loop reads one past the identifier
	l.readPosition--
	l.position--
	if l.input[l.position] == '\n' { // Correct column if backtrack crossed newline
		l.line--
		// Calculate previous line's length (tricky, maybe store last line length?)
		// For simplicity, reset column - less accurate but avoids complexity
		l.column = 0 // Less accurate, but simpler
	} else {
		l.column--
	}
	l.ch = l.input[l.position] // Restore char

	ident := l.input[position:l.readPosition]

	// Advance again for the next token read
	l.readChar()
	return ident
}

// readNumber reads an integer or floating-point number.
func (l *Lexer) readNumber() string {
	position := l.position
	hasDot := false
	for isDigit(l.ch) || (l.ch == '.' && !hasDot) {
		if l.ch == '.' {
			hasDot = true
		}
		l.readChar()
	}
	// Backtrack one char
	l.readPosition--
	l.position--
	if l.input[l.position] == '\n' {
		l.line--
		l.column = 0 // Simpler column handling
	} else {
		l.column--
	}
	l.ch = l.input[l.position]

	numStr := l.input[position:l.readPosition]

	// Advance again
	l.readChar()
	return numStr
}

// Token generates the sequence of tokens.
func (l *Lexer) Token() iter.Seq[Token] {
	return func(yield func(Token) bool) {
		for {
			l.skipWhitespace()

			tok := Token{Line: l.line, Column: l.column} // Capture position before consuming char

			currentChar := l.ch // Character that determines the token type

			// Handle single-character tokens first
			switch currentChar {
			case '(':
				tok.Type, tok.Literal = TOKEN_LPAREN, "("
				l.readChar() // Consume '('
			case ')':
				tok.Type, tok.Literal = TOKEN_RPAREN, ")"
				l.readChar() // Consume ')'
			case '"':
				tok.Type = TOKEN_STRING
				// readString consumes the closing quote
				tok.Literal = l.readString('"')
			case '\'':
				tok.Type = TOKEN_CHARACTER
				// readString consumes the closing quote
				tok.Literal = l.readString('\'')
			case ';':
				tok.Type = TOKEN_COMMENT
				// readComment consumes until newline
				tok.Literal = l.readComment()
				// Do not consume the newline itself here, let skipWhitespace handle it
			case 0:
				tok.Type, tok.Literal = TOKEN_EOF, ""
				// Don't consume EOF
			default:
				// Multi-character tokens
				if isLetter(currentChar) || currentChar == '_' {
					// readIdentifier consumes the identifier chars + 1 extra
					tok.Type, tok.Literal = TOKEN_IDENTIFIER, l.readIdentifier()
				} else if isDigit(currentChar) {
					// readNumber consumes the number chars + 1 extra
					tok.Type, tok.Literal = TOKEN_NUMBER, l.readNumber()
				} else {
					// Unrecognized character
					tok.Type, tok.Literal = TOKEN_ILLEGAL, string(currentChar)
					l.readChar() // Consume the illegal character
				}
			}

			// Yield the token
			if !yield(tok) || tok.Type == TOKEN_EOF {
				break // Stop iteration if yield returns false or EOF is reached
			}
		}
	}
}

// Helper functions (keep as before)
func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
