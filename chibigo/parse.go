package main

//
// Parser
//

// All local variable instances created during parsing are
// accumulated to this list.
var locals *Obj

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
	ND_ASSIGN                    // =
	ND_VAR                       // Variable
	ND_RETURN                    // "return"
	ND_BLOCK                     // { ... }
	ND_IF                        // "if"
	ND_FOR                       // "for"
)

// AST node type

type Node struct {
	kind NodeKind // Node kind
	next *Node    // Next node
	lhs  *Node    // Left-hand side
	rhs  *Node    // Right-hand side
	vr   *Obj
	val  int   // Used if kind == ND_NUM
	body *Node // Block
	cond *Node // "if" statement
	then *Node // "if" statement
	els  *Node // "if" statement
	init *Node // "for" statement
	inc  *Node // "for" statement
}

type Obj struct {
	next   *Obj
	name   string
	offset int
}

type Function struct {
	body      *Node
	locals    *Obj
	stackSize int
}

// Find a local variable by name.

func findVar(tok *Token) *Obj {
	for vr := locals; vr != nil; vr = vr.next {
		if len(vr.name) == tok.len && vr.name == string(currentInput[tok.loc:tok.loc+tok.len]) {
			return vr
		}
	}
	return nil
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

func newVarNode(vr *Obj) *Node {
	node := newNode(ND_VAR)
	node.vr = vr
	return node
}

func newLvar(name string) *Obj {
	vr := new(Obj)
	vr.name = string(name)
	vr.next = locals
	locals = vr
	return vr
}

// stmt = "return" expr ";"
//      | "if" expr "{" stmt "}" ("else" "{" stmt "}")?
//      | "for" (expr ";" expr ";" expr)? "{" stmt "}"
//      | "{" compound-stmt
//      | expr-stmt

func stmt(rest **Token, tok *Token) *Node {
	if equal(tok, "return") {
		node := newUnary(ND_RETURN, expr(&tok, tok.next))
		*rest = skip(tok, ";")
		return node
	}
	if equal(tok, "if") {
		node := newNode(ND_IF)
		node.cond = expr(&tok, tok.next)
		node.then = stmt(&tok, tok)
		if equal(tok, "else") {
			node.els = stmt(&tok, tok.next)
		}
		*rest = tok
		return node
	}
	if equal(tok, "for") {
		node := newNode(ND_FOR)
		tok = tok.next
		if !equal(tok, "{") {
			node.init = expr(&tok, tok)
			tok = skip(tok, ";")
			node.cond = expr(&tok, tok)
			tok = skip(tok, ";")
			node.inc = expr(&tok, tok)
		}
		node.then = stmt(&tok, tok)
		*rest = tok
		return node
	}
	if equal(tok, "{") {
		return componentStmt(rest, tok.next)
	}
	return exprStmt(rest, tok)
}

// compound-stmt = stmt* "}"

func componentStmt(rest **Token, tok *Token) *Node {
	head := new(Node)
	cur := head
	for !equal(tok, "}") {
		cur.next = stmt(&tok, tok)
		cur = cur.next
	}

	node := newNode(ND_BLOCK)
	node.body = head.next
	*rest = tok.next
	return node
}

// expr-stmt = expr? ";"

func exprStmt(rest **Token, tok *Token) *Node {
	if equal(tok, ";") {
		*rest = tok.next
		return newNode(ND_BLOCK)
	}

	node := newUnary(ND_EXPR_STMT, expr(&tok, tok))
	*rest = skip(tok, ";")
	return node
}

// expr = assign

func expr(rest **Token, tok *Token) *Node {
	return assign(rest, tok)
}

// assign = equality ("=" assign)?

func assign(rest **Token, tok *Token) *Node {
	node := equality(&tok, tok)
	if equal(tok, "=") {
		node = newBinary(ND_ASSIGN, node, assign(&tok, tok.next))
	}
	*rest = tok
	return node
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

// primary = "(" expr ")" | ident | num

func primary(rest **Token, tok *Token) *Node {
	if equal(tok, "(") {
		node := expr(&tok, tok.next)
		*rest = skip(tok, ")")
		return node
	}

	if tok.kind == TK_IDENT {
		vr := findVar(tok)
		if vr == nil {
			vr = newLvar(currentInput[tok.loc : tok.loc+tok.len])
		}
		*rest = tok.next
		return newVarNode(vr)
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

func parse(tok *Token) *Function {
	tok = skip(tok, "{")

	prog := new(Function)
	prog.body = componentStmt(&tok, tok)
	prog.locals = locals

	return prog
}
