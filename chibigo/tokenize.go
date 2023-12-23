package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	TK_STR                      // String literals
)

type Token struct {
	kind TokenKind // Token kind
	next *Token    // Next token
	val  int       // If kind is TK_NUM, its value
	loc  int       // Token location
	len  int       // Token length
	ty   *Type     // Used if TK_STR
	str  string    // String literal contents including terminating '\0'
}

// Input filename
var currentFilename string

// Input string
var currentInput string

// Reports an error and exit.
func errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

// Reports an error message in the following format and exit.
//
// foo.go:10: x = y + 1;
//               ^ <error message here>
func verrorAt(loc int, format string, a ...interface{}) {
	line := loc
	for line > 0 && currentInput[line-1] != '\n' {
		line--
	}

	end := loc
	for end < len(currentInput) && currentInput[end] != '\n' {
		end++
	}

	// Get a line number.
	lineNo := 1
	for i := 0; i < line; i++ {
		if currentInput[i] == '\n' {
			lineNo++
		}
	}

	// Print out the line.
	lineContent := currentInput[line:end]
	content := fmt.Sprintf("%s:%d: %s\n", currentFilename, lineNo, lineContent)

	// Show the error message.
	pos := len(content) - (end - loc + 1)
	fmt.Fprintf(os.Stderr, content)
	fmt.Fprintf(os.Stderr, "%*s^ ", pos, "")
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

func consume(rest **Token, tok *Token, str string) bool {
	if equal(tok, str) {
		*rest = tok.next
		return true
	}
	*rest = tok
	return false
}

// Create a new token.
func newToken(kind TokenKind, start, punctLen int) *Token {
	tok := &Token{
		kind: kind,
		loc:  start,
		len:  punctLen,
	}
	return tok
}

func isPunct(idx int) bool {
	return string(currentInput[idx]) == "+" || string(currentInput[idx]) == "-" ||
		string(currentInput[idx]) == "*" || string(currentInput[idx]) == "/" ||
		string(currentInput[idx]) == "(" || string(currentInput[idx]) == ")" ||
		string(currentInput[idx]) == "<" || string(currentInput[idx]) == ">" ||
		string(currentInput[idx]) == ";" || string(currentInput[idx]) == "=" ||
		string(currentInput[idx]) == "{" || string(currentInput[idx]) == "}" ||
		string(currentInput[idx]) == "&" || string(currentInput[idx]) == "," ||
		string(currentInput[idx]) == "[" || string(currentInput[idx]) == "]"
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
	kw := []string{"return", "if", "else", "for", "int", "char", "var", "func"}
	for _, keyword := range kw {
		if equal(tok, keyword) {
			return true
		}
	}
	return false
}

func readStringLiteral(idx int) *Token {
	start := idx
	cur := start
	for ; string(currentInput[cur]) != "\""; cur++ {
		if string(currentInput[cur]) == "\n" {
			errorAt(start, "unclosed string literal")
		}
	}
	tok := newToken(TK_STR, start, cur-start)
	tok.ty = arrayOf(tyChar, cur-start)
	tok.str = string(currentInput[tok.loc : tok.loc+tok.len])
	return tok
}

func convertKeywords(tok *Token) {
	for t := tok; t.kind != TK_EOF; t = t.next {
		if isKeyword(t) {
			t.kind = TK_KEYWORD
		}
	}
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

// Tokenize `currentInput` and returns new tokens.

func tokenize(filename string, input string) (*Token, error) {
	currentFilename = filename
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
		// String literal
		if currentInput[idx] == '"' {
			idx++
			cur.next = readStringLiteral(idx)
			cur = cur.next
			idx += cur.len + 1
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

func readFile(path string) (string, error) {
	var buf []byte
	var err error

	if path == "-" {
		// Read from stdin if the given filename is "-".
		buf, err = ioutil.ReadAll(os.Stdin)
	} else {
		buf, err = ioutil.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("cannot open %s: %s", path, err)
		}
	}

	// Ensure the last line is properly terminated with '\n'.
	if len(buf) == 0 || buf[len(buf)-1] != '\n' {
		buf = append(buf, '\n')
	}

	return string(buf), nil
}

func tokenizeFile(path string) (*Token, error) {
	file, err := readFile(path)
	if err != nil {
		return nil, err
	}
	if token, err := tokenize(path, file); err != nil {
		return nil, err
	} else {
		return token, nil
	}
}
