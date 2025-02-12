package lexer

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

// The list of token types
const (
	EOF TokenType = iota
	ERROR

	// Keywords
	CLASS
	INHERITS
	ISVOID
	IF
	ELSE
	FI
	THEN
	LET
	IN
	WHILE
	CASE
	ESAC
	LOOP
	POOL
	NEW
	OF
	NOT
	SELF      // Add SELF token type
	SELF_TYPE // Add SELF_TYPE token type
	VOID      // Add VOID token type

	// Data types
	STR_CONST
	BOOL_CONST
	INT_CONST

	// Identifiers
	TYPEID
	OBJECTID

	// Operators
	ASSIGN // <-
	DARROW // =>
	LT     // <
	LE     // <=
	EQ     // =
	PLUS   // +
	MINUS  // -
	TIMES  // *
	DIVIDE // /
	LPAREN // (
	RPAREN // )
	LBRACE // {
	RBRACE // }
	SEMI   // ;
	COLON  // :
	COMMA  // ,
	DOT    // .
	AT     // @
	NEG    // ~
)

func (tt TokenType) String() string {
	return [...]string{"EOF", "ERROR", "CLASS", "INHERITS", "ISVOID", "IF", "ELSE", "FI", "THEN", "LET", "IN", "WHILE", "CASE", "ESAC", "LOOP", "POOL",
		"NEW", "OF", "NOT", "SELF", "SELF_TYPE", "VOID", // Add VOID to string mapping
		"STR_CONST", "BOOL_CONST", "INT_CONST", "TYPEID", "OBJECTID", "ASSIGN", "DARROW", "LT", "LE", "EQ", "PLUS", "MINUS", "TIMES",
		"DIVIDE", "LPAREN", "RPAREN", "LBRACE", "RBRACE", "SEMI", "COLON", "COMMA", "DOT", "AT", "NEG"}[tt]
}

// Token represents a lexical token with its type, value, and position.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Lexer is the lexical analyzer.
type Lexer struct {
	reader *bufio.Reader
	line   int
	column int
	char   rune
}

// NewLexer creates a new lexer from an io.Reader
func NewLexer(reader io.Reader) *Lexer {
	l := &Lexer{
		reader: bufio.NewReader(reader),
		line:   1,
		column: 0,
		char:   ' ',
	}
	return l
}

// readChar reads the next character from the input.
func (l *Lexer) readChar() {
	var err error
	l.char, _, err = l.reader.ReadRune()
	if err != nil {
		l.char = 0 // EOF
	}

	l.column++
	if l.char == '\n' {
		l.line++
		l.column = 0
	}
}

// peekChar returns the next character without advancing the stream.
func (l *Lexer) peekChar() rune {
	char, _, err := l.reader.ReadRune()
	if err != nil {
		return 0
	}
	l.reader.UnreadRune()
	return char
}

// skipWhiteSpace skips whitespace characters and comments.
func (l *Lexer) skipWhiteSpace() {
	for unicode.IsSpace(l.char) || l.char == '-' || l.char == '(' {
		if l.char == '-' && l.peekChar() == '-' {
			// Single line comment
			for l.char != '\n' && l.char != 0 {
				l.readChar()
			}
		} else if l.char == '(' && l.peekChar() == '*' {
			// Multi-line comment
			l.readChar() // consume '('
			l.readChar() // consume '*'
			l.skipMultiLineComment()
		} else if unicode.IsSpace(l.char) {
			l.readChar()
		} else {
			break
		}
	}
}

// skipMultiLineComment skips over multi-line comments, handling nested comments.
func (l *Lexer) skipMultiLineComment() {
	nesting := 1
	for nesting > 0 {
		l.readChar()
		if l.char == 0 {
			return // EOF
		} else if l.char == '(' && l.peekChar() == '*' {
			nesting++
			l.readChar() // consume '*'
		} else if l.char == '*' && l.peekChar() == ')' {
			nesting--
			l.readChar() // consume ')'
		}
	}
	l.readChar() // consume the final ')'
}

func (l *Lexer) readNumber() string {
	// startPos := l.column
	var sb strings.Builder
	for unicode.IsDigit(l.char) {
		sb.WriteRune(l.char)
		l.readChar()
	}
	return sb.String()
}

func isIdentifierStart(char rune) bool {
	return unicode.IsLetter(char) || char == '_'
}

func isIdentifierPart(char rune) bool {
	return isIdentifierStart(char) || unicode.IsDigit(char)
}

func (l *Lexer) readIdentifier() string {
	var sb strings.Builder
	for isIdentifierPart(l.char) {
		sb.WriteRune(l.char)
		l.readChar()
	}
	return sb.String()
}

func (l *Lexer) readString() (string, error) {
	var sb strings.Builder
	l.readChar()
	for l.char != '"' {
		if l.char == 0 {
			return "", fmt.Errorf("EOF in string constant")
		}
		if l.char == '\n' {
			return "", fmt.Errorf("unterminated string constant")
		}

		if l.char == '\\' {
			l.readChar()
			switch l.char {
			case 'b':
				sb.WriteRune('\b')
			case 't':
				sb.WriteRune('\t')
			case 'n':
				sb.WriteRune('\n')
			case 'f':
				sb.WriteRune('\f')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case '0':
				sb.WriteRune(0)
			default:
				sb.WriteRune(l.char)
			}
		} else {
			sb.WriteRune(l.char)
		}

		l.readChar()
	}

	l.readChar()
	return sb.String(), nil
}

func (l *Lexer) NextToken() Token {
	l.skipWhiteSpace()

	tok := Token{
		Line:   l.line,
		Column: l.column,
	}

	switch {
	// Handle number literals first
	case unicode.IsDigit(l.char):
		num := l.readNumber()
		if _, err := strconv.Atoi(num); err != nil {
			tok.Type = ERROR
			tok.Literal = "Number out of range"
		} else {
			tok.Type = INT_CONST
			tok.Literal = num
		}
		return tok
	case l.char == 0:
		tok.Type = EOF
		tok.Literal = ""
	case l.char == '(':
		tok.Type = LPAREN
		tok.Literal = "("
		l.readChar()
	case l.char == ')':
		tok.Type = RPAREN
		tok.Literal = ")"
		l.readChar()
	case l.char == '{':
		tok.Type = LBRACE
		tok.Literal = "{"
		l.readChar()
	case l.char == '}':
		tok.Type = RBRACE
		tok.Literal = "}"
		l.readChar()
	case l.char == ';':
		tok.Type = SEMI
		tok.Literal = ";"
		l.readChar()
	case l.char == ':':
		tok.Type = COLON
		tok.Literal = ":"
		l.readChar()
	case l.char == ',':
		tok.Type = COMMA
		tok.Literal = ","
		l.readChar()
	case l.char == '+':
		tok.Type = PLUS
		tok.Literal = "+"
		l.readChar()
	case l.char == '*':
		tok.Type = TIMES
		tok.Literal = "*"
		l.readChar()
	// This could be a comment or a subtraction
	case l.char == '-':
		tok.Type = MINUS
		tok.Literal = "-"
		l.readChar()
	case l.char == '/':
		tok.Type = DIVIDE
		tok.Literal = "/"
		l.readChar()
	case l.char == '~':
		tok.Type = NEG
		tok.Literal = "~"
		l.readChar()
	case l.char == '.':
		tok.Type = DOT
		tok.Literal = "."
		l.readChar()
	case l.char == '=':
		if l.peekChar() == '>' {
			tok.Type = DARROW
			tok.Literal = "=>"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = EQ
			tok.Literal = "="
			l.readChar()
		}
	// Could be LT, LE, or ASSIGN
	case l.char == '<':
		if l.peekChar() == '-' {
			tok.Type = ASSIGN
			tok.Literal = "<-"
			l.readChar()
			l.readChar()
		} else if l.peekChar() == '=' {
			tok.Type = LE
			tok.Literal = "<="
			l.readChar()
			l.readChar()
		} else {
			tok.Type = LT
			tok.Literal = "<"
			l.readChar()
		}
	case l.char == '"':
		str, err := l.readString()
		if err != nil {
			tok.Type = ERROR
			tok.Literal = err.Error()
		} else {
			tok.Type = STR_CONST
			tok.Literal = str
		}
	case isIdentifierStart(l.char):
		identifier := l.readIdentifier()
		tok.Literal = identifier
		switch strings.ToLower(identifier) {
		// Handle case-insensitive keywords
		case "class":
			tok.Type = CLASS
		case "if":
			tok.Type = IF
		case "fi":
			tok.Type = FI
		case "else":
			tok.Type = ELSE
		case "then":
			tok.Type = THEN
		case "case":
			tok.Type = CASE
		case "esac":
			tok.Type = ESAC
		case "while":
			tok.Type = WHILE
		case "loop":
			tok.Type = LOOP
		case "pool":
			tok.Type = POOL
		case "of":
			tok.Type = OF
		case "let":
			tok.Type = LET
		case "in":
			tok.Type = IN
		case "inherits":
			tok.Type = INHERITS
		case "isvoid":
			tok.Type = ISVOID
		case "new":
			tok.Type = NEW
		case "not":
			tok.Type = NOT
		case "void":
			tok.Type = VOID
		case "self":
			tok.Type = SELF // Correctly tokenizes "self" as SELF
		// Handle boolean constants (case-insensitive)
		case "true", "false":
			tok.Type = BOOL_CONST
		default:
			if unicode.IsUpper(rune(identifier[0])) {
				tok.Type = TYPEID
			} else {
				tok.Type = OBJECTID
			}
		}
	default:
		tok.Type = ERROR
		tok.Literal = fmt.Sprintf("Unexpected character: %c", l.char)
		l.readChar()
	}

	return tok
}

// GetOperatorType returns the TokenType for a given operator string
func GetOperatorType(op string) TokenType {
	switch op {
	case "+":
		return PLUS
	case "-":
		return MINUS
	case "*":
		return TIMES
	case "/":
		return DIVIDE
	case "<":
		return LT
	case "<=":
		return LE
	case "=":
		return EQ
	case "<-":
		return ASSIGN
	case "~":
		return NEG
	case ".":
		return DOT
	case "@":
		return AT
	default:
		return ERROR
	}
}
