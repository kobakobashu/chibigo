package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"unicode"
)

type TokenKind int

const (
	TK_PUNCT TokenKind = iota // Punctuators
	TK_NUM                    // Numeric literals
	TK_EOF                    // End-of-file markers
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
func equal(tok *Token, s string, op string) bool {
	return bytes.Equal([]byte(s[tok.loc:tok.loc+tok.len]), []byte(op))
}

// Ensure that the current token is `s`.
func skip(tok *Token, s string, op string) *Token {
	if !equal(tok, s, op) {
		errorf("expected '%s", s)
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
func newToken(kind TokenKind, start int) *Token {
	tok := &Token{
		kind: kind,
		loc:  start,
	}
	return tok
}

// Tokenize `currentInput` and returns new tokens.
func tokenize() (*Token, error) {
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
			cur.next = newToken(TK_NUM, idx)
			cur = cur.next
			tmp := idx
			cur.val, idx, err = extractNum(currentInput, idx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return nil, err
			}
			cur.len = idx - tmp
			continue
		}
		if string(currentInput[idx]) == "+" || string(currentInput[idx]) == "-" {
			cur.next = newToken(TK_PUNCT, idx)
			cur = cur.next
			cur.len = 1
			idx++
			continue
		}
		errorAt(idx, "invalid token: %s", string(currentInput[idx]))
	}
	cur.next = newToken(TK_EOF, idx)
	cur = cur.next
	cur.len = 0
	return head.next, nil
}

func extractNum(s string, idx int) (int, int, error) {
	numericPart := 0
	for cur := idx; cur < len(s); cur++ {
		nextChar := string(s[cur])
		num, err := strconv.Atoi(nextChar)
		if err != nil {
			return numericPart, cur, nil
		}
		numericPart = numericPart*10 + num
	}
	return numericPart, len(s), nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Invalid arguments number")
		return
	}
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".globl main\n")
	fmt.Printf("main:\n")

	currentInput = os.Args[1]
	tok, err := tokenize()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	num, err := getNumber(tok)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	fmt.Printf("  mov rax, %d\n", num)
	tok = tok.next

	for tok.kind != TK_EOF {
		if equal(tok, currentInput, "+") {
			num, err := getNumber(tok.next)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
			fmt.Printf("  add rax, %d\n", num)
			tok = tok.next.next
			continue
		}
		if equal(tok, currentInput, "-") {
			num, err := getNumber(tok.next)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
			fmt.Printf("  sub rax, %d\n", num)
			tok = tok.next.next
			continue
		}
		fmt.Fprintf(os.Stderr, "Found invalid symbol: %s\n", string(currentInput[tok.loc]))
		return
	}

	fmt.Printf("  ret\n")
}
