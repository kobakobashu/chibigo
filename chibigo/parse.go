package main

//
// Parser
//

type NodeKind int

const (
	ND_ADD       NodeKind = iota // +
	ND_SUB                       // -
	ND_MUL                       // *
	ND_DIV                       // /
	ND_NUM                       // Integer
	ND_NEG                       // unary -
	ND_EQ                        // ==
	ND_NE                        // !=
	ND_LT                        // <
	ND_LE                        // <=
	ND_EXPR_STMT                 // Expression statement
)

// AST node type

type Node struct {
	kind NodeKind
	next *Node
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

// stmt = expr-stmt

func stmt(rest **Token, tok *Token) *Node {
	return exprStmt(rest, tok)
}

// expr-stmt = expr ";"

func exprStmt(rest **Token, tok *Token) *Node {
	node := newUnary(ND_EXPR_STMT, expr(&tok, tok))
	*rest = skip(tok, ";")
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

// program = stmt*

func parse(tok *Token) *Node {
	head := Node{}
	cur := &head
	for tok.kind != TK_EOF {
		cur.next = stmt(&tok, tok)
		cur = cur.next
	}

	return head.next
}
