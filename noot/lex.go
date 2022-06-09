package main

import (
	"io"
	"bufio"
	"unicode"
)

type Token int

const (
	EOF Token = iota
	ILLEGAL
	IDENT
	INT
	SEMI // ;
	COMMA // ;

	// Infix ops
	ADD // +
	SUB // -
	MUL // *
	DIV // /

	ASSIGN // =

	LPAREN // (
	RPAREN // )
	LBRACE // {
	RBRACE // }
)

var tokens = []string{
	EOF:     "EOF",
	ILLEGAL: "ILLEGAL",
	IDENT:   "IDENT",

	INT:     "INT",
	SEMI:    ";",
	COMMA:    ",",

	// Infix ops
	ADD: "ADD",
	SUB: "SUB",
	MUL: "MUL",
	DIV: "DIV",

	ASSIGN: "=",

	LPAREN: "(",
	RPAREN: ")",
	LBRACE: "{",
	RBRACE: "}",
}

func (t Token) String() string {
	return tokens[t]
}

type Position struct {
	line   int
	column int
}

type Lexer struct {
	lastToken Token
	pos    Position
	reader *bufio.Reader
}

func NewLexer(reader io.Reader) *Lexer {
	return &Lexer{
		lastToken: ILLEGAL,
		pos:    Position{line: 1, column: 0},
		reader: bufio.NewReader(reader),
	}
}

// Lex scans the input for the next token. It returns the position of the token,
// the token's type, and the literal value.
func (l *Lexer) Lex() (Position, Token, string) {
	// keep looping until we return a token
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return l.pos, EOF, "EOF"
			}

			// at this point there isn't much we can do, and the compiler
			// should just return the raw error to the user
			panic(err)
		}

		// update the column to the position of the newly read in rune
		l.pos.column++

		switch r {
		case '\n':
			// Decide if we want to add semicolon
			if l.lastToken == IDENT || l.lastToken == RPAREN || l.lastToken == INT {
				l.resetPosition()
				return l.pos, SEMI, ";"
			}
			l.resetPosition()
		case ';':
			l.lastToken = SEMI
			return l.pos, SEMI, ";"
		case ',':
			l.lastToken = COMMA
			return l.pos, COMMA, ","
		case '+':
			l.lastToken = ADD
			return l.pos, ADD, "+"
		case '-':
			l.lastToken = SUB
			return l.pos, SUB, "-"
		case '*':
			l.lastToken = MUL
			return l.pos, MUL, "*"
		case '/':
			l.lastToken = DIV
			return l.pos, DIV, "/"
		case '=':
			l.lastToken = ASSIGN
			return l.pos, ASSIGN, "="
		case '(':
			l.lastToken = LPAREN
			return l.pos, LPAREN, "("
		case ')':
			l.lastToken = RPAREN
			return l.pos, RPAREN, ")"
		case '{':
			l.lastToken = LBRACE
			return l.pos, LBRACE, "}"
		case '}':
			l.lastToken = RBRACE
			return l.pos, RBRACE, "}"
		default:
			if unicode.IsSpace(r) {
				continue // nothing to do here, just move on
			} else if unicode.IsDigit(r) {
				// backup and let lexInt rescan the beginning of the int
				startPos := l.pos
				l.backup()
				lit := l.lexInt()
				l.lastToken = INT
				return startPos, INT, lit
			} else if unicode.IsLetter(r) {
				// backup and let lexIdent rescan the beginning of the ident
				startPos := l.pos
				l.backup()
				lit := l.lexIdent()

				l.lastToken = IDENT
				return startPos, IDENT, lit
			} else {
				l.lastToken = ILLEGAL
				return l.pos, ILLEGAL, string(r)
			}
		}
	}
}

func (l *Lexer) resetPosition() {
	l.pos.line++
	l.pos.column = 0
}

func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}

	l.pos.column--
}

// lexInt scans the input until the end of an integer and then returns the
// literal.
func (l *Lexer) lexInt() string {
	var lit string
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				// at the end of the int
				return lit
			}
		}

		l.pos.column++
		if unicode.IsDigit(r) {
			lit = lit + string(r)
		} else {
			// scanned something not in the integer
			l.backup()
			return lit
		}
	}
}

// lexIdent scans the input until the end of an identifier and then returns the
// literal.
func (l *Lexer) lexIdent() string {
	var lit string
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				// at the end of the identifier
				return lit
			}
		}

		l.pos.column++
		if unicode.IsLetter(r) {
			lit = lit + string(r)
		} else {
			// scanned something not in the identifier
			l.backup()
			return lit
		}
	}
}
