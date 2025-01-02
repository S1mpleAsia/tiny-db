package parse

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type lexer interface {
	MatchDelim(d rune) bool // ',' '='
	MatchIntConstant() bool
	MatchStringConstant() bool
	MatchKeyWord(word string) bool
	MatchIdentifier() bool // identifier (table_name, column_name, ...)

	EatDelim(d rune) error
	EatIntConstant() (int32, error)
	EatStringConstant() (string, error)
	EatKeyword(word string) error
	EatIdentifier() (string, error)
}

type TokenKind int

const (
	TOKEN_KIND_EOF TokenKind = iota + 1
	TOKEN_KIND_DELIMITER
	TOKEN_KIND_INTEGER
	TOKEN_KIND_STRING
	TOKEN_KIND_KEYWORD
	TOKEN_KIND_IDENTIFIER
)

type token struct {
	kind  TokenKind
	value string
}

const whiteSpaces string = " \t\r\n"

var keywords = map[string]struct{}{
	"select":  {},
	"from":    {},
	"where":   {},
	"and":     {},
	"insert":  {},
	"into":    {},
	"values":  {},
	"delete":  {},
	"update":  {},
	"set":     {},
	"create":  {},
	"table":   {},
	"int":     {},
	"varchar": {},
	"view":    {},
	"as":      {},
	"index":   {},
	"on":      {},
}

var _ lexer = (*Lexer)(nil)

type Lexer struct {
	input       string
	token       *token
	whiteSpaces string
	keywords    map[string]struct{}
}

func NewLexer(input string) (*Lexer, error) {
	l := &Lexer{
		input:       input,
		token:       nil,
		whiteSpaces: whiteSpaces,
		keywords:    keywords,
	}

	if err := l.nextToken(); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *Lexer) MatchDelim(d rune) bool {
	return l.token.kind == TOKEN_KIND_DELIMITER && l.token.value == string(d)
}

func (l *Lexer) MatchIntConstant() bool {
	return l.token.kind == TOKEN_KIND_INTEGER
}

func (l *Lexer) MatchStringConstant() bool {
	return l.token.kind == TOKEN_KIND_STRING
}

func (l *Lexer) MatchKeyWord(word string) bool {
	return l.token.kind == TOKEN_KIND_KEYWORD && l.token.value == word
}

func (l *Lexer) MatchIdentifier() bool {
	return l.token.kind == TOKEN_KIND_IDENTIFIER
}

func (l *Lexer) EatDelim(d rune) error {
	if !l.MatchDelim(d) {
		return NewBadSyntaxError(fmt.Sprintf("expected %c, got %q", d, l.token.value))
	}

	if err := l.nextToken(); err != nil {
		return err
	}

	return nil
}

func (l *Lexer) EatIntConstant() (int32, error) {
	if !l.MatchIntConstant() {
		return 0, NewBadSyntaxError(fmt.Sprintf("expected int, but got %q", l.token.value))
	}

	value, err := strconv.Atoi(l.token.value)
	if err != nil {
		return 0, err
	}

	if err := l.nextToken(); err != nil {
		return 0, err
	}

	return int32(value), nil
}

func (l *Lexer) EatStringConstant() (string, error) {
	if !l.MatchStringConstant() {
		return "", NewBadSyntaxError(fmt.Sprintf("expected string, but got %q", l.token.value))
	}

	value := l.token.value

	if err := l.nextToken(); err != nil {
		return "", err
	}

	return value, nil
}

func (l *Lexer) EatKeyword(word string) error {
	if !l.MatchKeyWord(word) {
		return NewBadSyntaxError(fmt.Sprintf("expected %q, but got %q", word, l.token.value))
	}

	if err := l.nextToken(); err != nil {
		return err
	}

	return nil
}

func (l *Lexer) EatIdentifier() (string, error) {
	if !l.MatchIdentifier() {
		return "", NewBadSyntaxError(fmt.Sprintf("expeceted indentifier, got %q", l.token.value))
	}

	value := l.token.value

	if err := l.nextToken(); err != nil {
		return "", err
	}

	return value, nil
}

func (l *Lexer) nextToken() error {
	l.input = strings.TrimLeft(l.input, l.whiteSpaces)

	if len(l.input) == 0 {
		l.token = &token{
			kind:  TOKEN_KIND_EOF,
			value: "",
		}

		return nil
	}

	switch r := l.input[0]; {
	case isDigit(r):
		return l.readInteger()
	case r == '\'':
		return l.readString()
	case isIdentifierStart(r):
		return l.readIdentifier()
	default:
		return l.readDelimiter()
	}
}

// Integer: [0-9]+
func (l *Lexer) readInteger() error {
	pos := 1

	for ; pos < len(l.input); pos++ {
		if !isDigit(l.input[pos]) {
			break
		}
	}

	l.token = &token{
		kind:  TOKEN_KIND_INTEGER,
		value: l.input[:pos],
	}

	l.input = l.input[pos:]
	return nil
}

// String: '.*'
func (l *Lexer) readString() error {
	pos := 1

	close := strings.IndexByte(l.input[pos:], '\'')
	if close == -1 {
		return NewBadSyntaxError("unterminated string")
	}

	pos += close
	content := l.input[1:pos]

	pos += 1
	l.token = &token{
		kind:  TOKEN_KIND_STRING,
		value: content,
	}

	l.input = l.input[pos:]
	return nil
}

// Identifier: [A-Z_a-z][0-9A-Z_a-z]*
func (l *Lexer) readIdentifier() error {
	// [A-Z_a-z]
	pos := 1

	for ; pos < len(l.input); pos++ {
		if !isIdentifierContinuation(l.input[pos]) {
			break
		}
	}

	word := strings.ToLower(l.input[0:pos])
	kind := TOKEN_KIND_IDENTIFIER

	if _, ok := l.keywords[word]; ok {
		kind = TOKEN_KIND_KEYWORD
	}

	l.token = &token{
		kind:  kind,
		value: word,
	}

	l.input = l.input[pos:]
	return nil
}

func (l *Lexer) readDelimiter() error {
	_, size := utf8.DecodeRuneInString(l.input)

	l.token = &token{
		kind:  TOKEN_KIND_DELIMITER,
		value: l.input[:size],
	}

	l.input = l.input[size:]
	return nil
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isIdentifierStart(c byte) bool {
	return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || c == '_'
}

func isIdentifierContinuation(c byte) bool {
	return isIdentifierStart(c) || isDigit(c)
}
