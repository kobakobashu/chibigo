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
func equal(tok *Token, op string) bool {
	return bytes.Equal([]byte(currentInput[tok.loc:tok.loc+tok.len]), []byte(op))
}

// Ensure that the current token is `s`.
func skip(tok *Token, op string) *Token {
	if !equal(tok, op) {
		errorf("expected '%s", op)
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
		string(currentInput[idx]) == "<" || string(currentInput[idx]) == ">"
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

//
// Parser
//

type NodeKind int

const (
	ND_ADD NodeKind = iota // +
	ND_SUB                 // -
	ND_MUL                 // *
	ND_DIV                 // /
	ND_NUM                 // Integer
	ND_NEG                 // unary -
	ND_EQ                  // ==
	ND_NE                  // !=
	ND_LT                  // <
	ND_LE                  // <=
)

// AST node type

type Node struct {
	kind NodeKind
	lhs  *Node
	rhs  *Node
	val  int
}

func newNode(kind NodeKind) *Node {
	node := new(Node)
	node.kind = kind
	return node
}

func newBinary(kind NodeKind, lhs *Node, rhs *Node) *Node {
	node := newNode(kind)
	node.lhs = lhs
	node.rhs = rhs
	return node
}

func newUnary(kind NodeKind, expr *Node) *Node {
	node := newNode(kind)
	node.lhs = expr
	return node
}

func newNum(val int) *Node {
	node := newNode(ND_NUM)
	node.val = val
	return node
}

// expr = equality

func expr(rest **Token, tok *Token) *Node {
	return equality(rest, tok)
}

// equality = relational ("==" relational | "!=" relational)*

func equality(rest **Token, tok *Token) *Node {
	node := relational(&tok, tok)

	for {
		if equal(tok, "==") {
			node = newBinary(ND_EQ, node, relational(&tok, tok.next))
			continue
		}
		if equal(tok, "!=") {
			node = newBinary(ND_NE, node, relational(&tok, tok.next))
			continue
		}
		*rest = tok
		return node
	}
}

// relational = add ("<" add | "<=" add | ">" add | ">=" add)*

func relational(rest **Token, tok *Token) *Node {
	node := add(&tok, tok)

	for {
		if equal(tok, "<") {
			node = newBinary(ND_LT, node, add(&tok, tok.next))
			continue
		}
		if equal(tok, "<=") {
			node = newBinary(ND_LE, node, add(&tok, tok.next))
			continue
		}
		if equal(tok, ">") {
			node = newBinary(ND_LT, add(&tok, tok.next), node)
			continue
		}
		if equal(tok, ">=") {
			node = newBinary(ND_LE, add(&tok, tok.next), node)
			continue
		}
		*rest = tok
		return node
	}
}

// add = mul ("+" mul | "-" mul)*

func add(rest **Token, tok *Token) *Node {
	node := mul(&tok, tok)

	for {
		if equal(tok, "+") {
			node = newBinary(ND_ADD, node, mul(&tok, tok.next))
			continue
		}

		if equal(tok, "-") {
			node = newBinary(ND_SUB, node, mul(&tok, tok.next))
			continue
		}

		*rest = tok
		return node
	}
}

// mul = unary ("*" unary | "/" unary)*

func mul(rest **Token, tok *Token) *Node {
	node := unary(&tok, tok)

	for {
		if equal(tok, "*") {
			node = newBinary(ND_MUL, node, unary(&tok, tok.next))
			continue
		}

		if equal(tok, "/") {
			node = newBinary(ND_DIV, node, unary(&tok, tok.next))
			continue
		}

		*rest = tok
		return node
	}
}

// unary = ("+" | "-") unary
//       | primary

func unary(rest **Token, tok *Token) *Node {
	if equal(tok, "+") {
		return unary(rest, tok.next)
	}
	if equal(tok, "-") {
		return newUnary(ND_NEG, unary(rest, tok.next))
	}

	return primary(rest, tok)
}

// primary = "(" expr ")" | num

func primary(rest **Token, tok *Token) *Node {
	if equal(tok, "(") {
		node := expr(&tok, tok.next)
		*rest = skip(tok, ")")
		return node
	}

	if tok.kind == TK_NUM {
		node := newNum(tok.val)
		*rest = tok.next
		return node
	}

	errorTok(tok, "expected an expression")
	return nil
}

//
// Code generator
//

var depth int

func push() {
	fmt.Printf("  push rax\n")
	depth++
}

func pop(arg string) {
	fmt.Printf("  pop %s\n", arg)
	depth--
}

func cmp(cmd string) {
	fmt.Printf("  cmp rax, rdi\n")
	fmt.Printf("  %s al\n", cmd)
	fmt.Printf("  movzb rax, al\n")
}

func genExpr(node *Node) {
	switch node.kind {
	case ND_NUM:
		fmt.Printf("  mov rax, %d\n", node.val)
		return
	case ND_NEG:
		genExpr(node.lhs)
		fmt.Printf("  neg rax\n")
		return
	}

	genExpr(node.rhs)
	push()
	genExpr(node.lhs)
	pop("rdi")

	switch node.kind {
	case ND_ADD:
		fmt.Printf("  add rax, rdi\n")
		return
	case ND_SUB:
		fmt.Printf("  sub rax, rdi\n")
		return
	case ND_MUL:
		fmt.Printf("  imul rax, rdi\n")
		return
	case ND_DIV:
		fmt.Printf("  cqo\n")
		fmt.Printf("  idiv rdi\n")
		return
	case ND_EQ:
		cmp("sete")
		return
	case ND_NE:
		cmp("setne")
		return
	case ND_LT:
		cmp("setl")
		return
	case ND_LE:
		cmp("setle")
		return
	}

	errorf("invalid expression")
	return
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Invalid arguments number")
		return
	}

	currentInput = os.Args[1]
	tok, err := tokenize()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	node := expr(&tok, tok)

	if tok.kind != TK_EOF {
		errorTok(tok, "extra token")
	}

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".globl main\n")
	fmt.Printf("main:\n")

	// Traverse the AST to emit assembly.
	genExpr(node)
	fmt.Printf("  ret\n")

	if depth != 0 {
		panic("Depth is not zero")
	}
}
