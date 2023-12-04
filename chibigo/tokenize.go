package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

//
// Tokenizer
//

type TokenKind int

const (
	TK_PUNCT   TokenKind = iota // Punctuators
	TK_IDENT                    // Identifiers
	TK_NUM                      // Numeric literals
	TK_EOF                      // End-of-file markers
	TK_KEYWORD                  // Keywords
)

type Token struct {
	kind TokenKind // Token kind
	next *Token    // Next token
	val  int       // If kind is TK_NUM, its value
	loc  int       // Token location
	len  int       // Token length
}

var currentInput string

// Reports an error and exit.
func errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func verrorAt(loc int, format string, a ...interface{}) {
	fmt.Fprintln(os.Stderr, string(currentInput))
	fmt.Fprintf(os.Stderr, "%*s^ ", loc, "")
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func errorAt(loc int, format string, a ...interface{}) {
	verrorAt(loc, format, a...)
}

func errorTok(tok *Token, format string, a ...interface{}) {
	verrorAt(tok.loc, format, a...)
}

// Consumes the current token if it matches "op".
func equal(tok *Token, op string) bool {
	return bytes.Equal([]byte(currentInput[tok.loc:tok.loc+tok.len]), []byte(op))
}

// Ensure that the current token is `s`.
func skip(tok *Token, op string) *Token {
	if !equal(tok, op) {
		errorTok(tok, "expected %s", op)
	}
	return tok.next
}

// Ensure that the current token is TK_NUM.
func getNumber(tok *Token) (int, error) {
	if tok.kind != TK_NUM {
		return 0, fmt.Errorf("expected a number")
	}
	return tok.val, nil
}

// Create a new token.
func newToken(kind TokenKind, start, punct_len int) *Token {
	tok := &Token{
		kind: kind,
		loc:  start,
		len:  punct_len,
	}
	return tok
}

func isPunct(idx int) bool {
	return string(currentInput[idx]) == "+" || string(currentInput[idx]) == "-" ||
		string(currentInput[idx]) == "*" || string(currentInput[idx]) == "/" ||
		string(currentInput[idx]) == "(" || string(currentInput[idx]) == ")" ||
		string(currentInput[idx]) == "<" || string(currentInput[idx]) == ">" ||
		string(currentInput[idx]) == ";" || string(currentInput[idx]) == "=" ||
		string(currentInput[idx]) == "{" || string(currentInput[idx]) == "}"
}

func startswith(p, q string) bool {
	return strings.HasPrefix(p, q)
}

func readPunct(idx int) int {
	p := string(currentInput[idx:min(len(currentInput), idx+2)])
	if startswith(p, "==") || startswith(p, "!=") ||
		startswith(p, "<=") || startswith(p, ">=") {
		return 2
	}
	if isPunct(idx) {
		return 1
	}
	return 0
}

// Returns true if c is valid as the first character of an identifier.

func isIdent1(idx int) bool {
	return 'a' <= currentInput[idx] && currentInput[idx] <= 'z' ||
		'A' <= currentInput[idx] && currentInput[idx] <= 'Z' ||
		currentInput[idx] == '_'
}

// Returns true if c is valid as a non-first character of an identifier.

func isIdent2(idx int) bool {
	return isIdent1(idx) || '0' <= currentInput[idx] && currentInput[idx] <= '9'
}

func isKeyword(tok *Token) bool {
	kw := []string{"return", "if", "else"}
	for _, keyword := range kw {
		if equal(tok, keyword) {
			return true
		}
	}
	return false
}

func convertKeywords(tok *Token) {
	for t := tok; t.kind != TK_EOF; t = t.next {
		if isKeyword(t) {
			t.kind = TK_KEYWORD
		}
	}
}

// Tokenize `currentInput` and returns new tokens.
func tokenize(input string) (*Token, error) {
	currentInput = input
	head := Token{}
	cur := &head

	var err error
	idx := 0
	for idx < len(currentInput) {
		if unicode.IsSpace(rune(currentInput[idx])) {
			idx += 1
			continue
		}
		if unicode.IsDigit(rune(currentInput[idx])) {
			cur.next = newToken(TK_NUM, idx, 0)
			cur = cur.next
			tmp := idx
			cur.val, idx, err = extractNum(idx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return nil, err
			}
			cur.len = idx - tmp
			continue
		}
		// Identifier or keyword
		if isIdent1(idx) {
			start := idx
			idx++
			for isIdent2(idx) {
				idx++
			}
			cur.next = newToken(TK_IDENT, start, idx-start)
			cur = cur.next
			continue
		}
		if punctLen := readPunct(idx); punctLen >= 1 {
			cur.next = newToken(TK_PUNCT, idx, punctLen)
			cur = cur.next
			idx += punctLen
			continue
		}
		errorAt(idx, "invalid token: %s", string(currentInput[idx]))
	}
	cur.next = newToken(TK_EOF, idx, 0)
	cur = cur.next
	convertKeywords(head.next)
	return head.next, nil
}

func extractNum(idx int) (int, int, error) {
	numericPart := 0
	for cur := idx; cur < len(currentInput); cur++ {
		nextChar := string(currentInput[cur])
		num, err := strconv.Atoi(nextChar)
		if err != nil {
			return numericPart, cur, nil
		}
		numericPart = numericPart*10 + num
	}
	return numericPart, len(currentInput), nil
}
