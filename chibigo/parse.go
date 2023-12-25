package main

import "fmt"

//
// Parser
//

// All local variable instances created during parsing are
// accumulated to this list.
var locals *Obj
var globals *Obj

var scope *Scope = new(Scope)

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
	ND_ADDR                      // unary &
	ND_DEREF                     // unary *
	ND_VAR                       // Variable
	ND_RETURN                    // "return"
	ND_BLOCK                     // { ... }
	ND_FUNCALL                   // Function call
	ND_IF                        // "if"
	ND_FOR                       // "for"
)

// AST node type

type Node struct {
	kind     NodeKind // Node kind
	next     *Node    // Next node
	ty       *Type    // Type, e.g. int or pointer to int
	tok      *Token   // Representative token
	lhs      *Node    // Left-hand side
	rhs      *Node    // Right-hand side
	vr       *Obj
	val      int    // Used if kind == ND_NUM
	body     *Node  // Block
	funcname string // Function call
	args     *Node  // Function args
	cond     *Node  // "if" statement
	then     *Node  // "if" statement
	els      *Node  // "if" statement
	init     *Node  // "for" statement
	inc      *Node  // "for" statement
}

type Obj struct {
	next       *Obj
	name       string // Variable name
	ty         *Type  // Type
	isLocal    bool   // local or global/function
	offset     int    // Local variable
	isFunction bool   // Global variable or function
	params     *Obj
	body       *Node
	locals     *Obj
	stackSize  int
	initData   string // Global variable
	init       *Node  // Global variable initial value
}

// Scope for local or global variables.

type VarScope struct {
	next  *VarScope
	name  string
	vrObj *Obj
}

// Represents a block scope.

type Scope struct {
	next *Scope
	vrs  *VarScope
}

func enterScope() {
	sc := new(Scope)
	sc.next = scope
	scope = sc
}

func leaveScope() {
	scope = scope.next
}

// Find a local variable by name.

func findVar(tok *Token) *Obj {
	for sc := scope; sc != nil; sc = sc.next {
		for sc2 := sc.vrs; sc2 != nil; sc2 = sc2.next {
			if equal(tok, sc2.name) {
				return sc2.vrObj
			}
		}
	}

	return nil
}

func newNode(kind NodeKind, tok *Token) *Node {
	node := new(Node)
	node.kind = kind
	node.tok = tok
	return node
}

func newBinary(kind NodeKind, lhs *Node, rhs *Node, tok *Token) *Node {
	node := newNode(kind, tok)
	node.lhs = lhs
	node.rhs = rhs
	return node
}

func newUnary(kind NodeKind, expr *Node, tok *Token) *Node {
	node := newNode(kind, tok)
	node.lhs = expr
	return node
}

func newNum(val int, tok *Token) *Node {
	node := newNode(ND_NUM, tok)
	node.val = val
	return node
}

func newVarNode(vr *Obj, tok *Token) *Node {
	node := newNode(ND_VAR, tok)
	node.vr = vr
	return node
}

func pushScope(name string, vr *Obj) *VarScope {
	sc := new(VarScope)
	sc.name = name
	sc.vrObj = vr
	sc.next = scope.vrs
	scope.vrs = sc
	return sc
}

func newVar(name string, ty *Type) *Obj {
	vr := new(Obj)
	vr.name = string(name)
	vr.ty = ty
	pushScope(name, vr)
	return vr
}

func newLvar(name string, ty *Type) *Obj {
	vr := newVar(name, ty)
	vr.isLocal = true
	vr.next = locals
	locals = vr
	return vr
}

func newGvar(name string, ty *Type) *Obj {
	vr := newVar(name, ty)
	vr.next = globals
	globals = vr
	return vr
}

var id int = 0

func newUniqueName() string {
	buf := fmt.Sprintf(".L..%d", id)
	return buf
}

func newAnonGvar(ty *Type) *Obj {
	return newGvar(newUniqueName(), ty)
}

func newStringLiteral(str string, ty *Type) *Obj {
	vr := newAnonGvar(ty)
	vr.initData = str
	return vr
}

func getIdent(tok *Token) string {
	if tok.kind != TK_IDENT {
		errorTok(tok, "expected an identifier")
	}
	return currentInput[tok.loc : tok.loc+tok.len]
}

// func-params = (param ("," param)*)? ")"
// param       = declspec declarator

func funcParams(rest **Token, tok *Token) *Type {
	tok = tok.next

	head := new(Type)
	cur := head

	for !equal(tok, ")") {
		if cur != head {
			tok = skip(tok, ",")
		}
		vr := newNode(ND_VAR, tok)
		ty := declarator(&tok, tok.next)
		ty.name = vr.tok
		cur.next = copyType(ty)
		cur = cur.next
	}
	*rest = tok.next
	return head.next
}

// type-suffix = ("(" func-params

func typeSuffix(rest **Token, tok *Token) *Type {
	if equal(tok, "(") == true {
		return funcParams(rest, tok)
	}

	*rest = tok
	return nil
}

// declspec = "int" || "char"

func declspec(rest **Token, tok *Token) *Type {
	if equal(tok, "char") {
		*rest = tok.next
		return tyChar
	}

	if equal(tok, "int") {
		*rest = tok.next
		return tyInt
	}

	errorTok(tok, "Found an unsupported specifier")
	return nil
}

// declarator = "[" num "]" declarator
//            | "*"* declspec

func declarator(rest **Token, tok *Token) *Type {
	if equal(tok, "[") {
		sz, err := getNumber(tok.next)
		if err != nil {
			fmt.Printf("Error: ", err)
			return nil
		}
		tok = skip(tok.next.next, "]")
		base := declarator(&tok, tok)
		*rest = tok
		return arrayOf(base, sz)
	}

	tmp := tok
	for equal(tmp, "*") {
		tmp = tmp.next
	}

	ty := declspec(&tmp, tmp)
	for consume(&tok, tok, "*") {
		ty = pointerTo(ty)
	}

	*rest = tmp
	return ty
}

// declaration = "var" ident (("," ident)?)* declarator ("=" expr ("," expr)?)*)? ";"

func declaration(rest **Token, tok *Token) *Node {
	vrs_head := storeIdentTemp(&tok, tok)
	ty := declarator(&tok, tok)
	head := new(Node)
	cur := head
	if consume(&tok, tok, "=") {
		for vr_cur := vrs_head; vr_cur != nil; vr_cur = vr_cur.next {
			vr := newLvar(getIdent(vr_cur.tok), ty)

			lhs := newVarNode(vr, ty.name)
			rhs := assign(&tok, tok)
			node := newBinary(ND_ASSIGN, lhs, rhs, tok)
			cur.next = newUnary(ND_EXPR_STMT, node, tok)
			cur = cur.next
			consume(&tok, tok, ",")
		}
	} else {
		for vr_cur := vrs_head; vr_cur != nil; vr_cur = vr_cur.next {
			newLvar(getIdent(vr_cur.tok), ty)
		}
	}

	node := newNode(ND_BLOCK, tok)
	node.body = head.next
	*rest = tok.next
	return node
}

// Returns true if a given token represents a type.

func isTypename(tok *Token) bool {
	return equal(tok, "char") || equal(tok, "int")
}

// stmt = "return" expr ";"
//      | "if" expr "{" stmt "}" ("else" "{" stmt "}")?
//      | "for" (expr ";" expr ";" expr)? "{" stmt "}"
//      | "for" expr "{" stmt "}"
//      | "{" compound-stmt
//      | expr-stmt

func stmt(rest **Token, tok *Token) *Node {
	if equal(tok, "return") {
		node := newNode(ND_RETURN, tok)
		node.lhs = expr(&tok, tok.next)
		*rest = skip(tok, ";")
		return node
	}
	if equal(tok, "if") {
		node := newNode(ND_IF, tok)
		node.cond = expr(&tok, tok.next)
		node.then = stmt(&tok, tok)
		if equal(tok, "else") {
			node.els = stmt(&tok, tok.next)
		}
		*rest = tok
		return node
	}
	if equal(tok, "for") {
		node := newNode(ND_FOR, tok)
		tok = tok.next
		if !equal(tok, "{") {
			node.init = expr(&tok, tok)
			if equal(tok, ";") {
				// for
				tok = skip(tok, ";")
				node.cond = expr(&tok, tok)
				tok = skip(tok, ";")
				node.inc = expr(&tok, tok)
			} else {
				// while
				node.cond = node.init
				node.init = nil
			}
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

// compound-stmt = (declaration | stmt)* "}"

func componentStmt(rest **Token, tok *Token) *Node {
	node := newNode(ND_BLOCK, tok)

	head := new(Node)
	cur := head
	enterScope()
	for !equal(tok, "}") {
		if equal(tok, "var") {
			cur.next = declaration(&tok, tok)
			cur = cur.next
		} else {
			cur.next = stmt(&tok, tok)
			cur = cur.next
		}
	}
	leaveScope()
	node.body = head.next
	*rest = tok.next
	return node
}

// expr-stmt = expr? ";"

func exprStmt(rest **Token, tok *Token) *Node {
	if equal(tok, ";") {
		*rest = tok.next
		return newNode(ND_BLOCK, tok)
	}

	node := newNode(ND_EXPR_STMT, tok)
	node.lhs = expr(&tok, tok)
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
		return newBinary(ND_ASSIGN, node, assign(rest, tok.next), tok)
	}
	*rest = tok
	return node
}

// equality = relational ("==" relational | "!=" relational)*

func equality(rest **Token, tok *Token) *Node {
	node := relational(&tok, tok)

	for {
		start := tok
		if equal(tok, "==") {
			node = newBinary(ND_EQ, node, relational(&tok, tok.next), start)
			continue
		}
		if equal(tok, "!=") {
			node = newBinary(ND_NE, node, relational(&tok, tok.next), start)
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
		start := tok
		if equal(tok, "<") {
			node = newBinary(ND_LT, node, add(&tok, tok.next), start)
			continue
		}
		if equal(tok, "<=") {
			node = newBinary(ND_LE, node, add(&tok, tok.next), start)
			continue
		}
		if equal(tok, ">") {
			node = newBinary(ND_LT, add(&tok, tok.next), node, start)
			continue
		}
		if equal(tok, ">=") {
			node = newBinary(ND_LE, add(&tok, tok.next), node, start)
			continue
		}
		*rest = tok
		return node
	}
}

// In Go, '+' and '-' operators are not overloaded to perform the pointer arithmetic

func newAdd(lhs *Node, rhs *Node, tok *Token) *Node {
	addType(lhs)
	addType(rhs)

	// num + num
	if isInteger(lhs.ty) && isInteger(rhs.ty) {
		return newBinary(ND_ADD, lhs, rhs, tok)
	}

	if lhs.ty.base != nil && rhs.ty.base != nil {
		errorTok(tok, "invalid operands")
	}

	// Canonicalize `num + ptr` to `ptr + num`.
	if lhs.ty.base == nil && rhs.ty.base != nil {
		tmp := lhs
		lhs = rhs
		rhs = tmp
	}

	// ptr + num
	rhs = newBinary(ND_MUL, rhs, newNum(lhs.ty.base.size, tok), tok)
	return newBinary(ND_ADD, lhs, rhs, tok)
}

func newSub(lhs *Node, rhs *Node, tok *Token) *Node {
	addType(lhs)
	addType(rhs)

	// num + num
	if isInteger(lhs.ty) && isInteger(rhs.ty) {
		return newBinary(ND_SUB, lhs, rhs, tok)
	}

	if lhs.ty.base != nil || rhs.ty.base != nil {
		errorTok(tok, "invalid operands: pointer arithmetic is not supported in Go")
	}

	errorTok(tok, "invalid operands")
	return nil
}

// add = mul ("+" mul | "-" mul)*

func add(rest **Token, tok *Token) *Node {
	node := mul(&tok, tok)

	for {
		start := tok
		if equal(tok, "+") {
			node = newAdd(node, mul(&tok, tok.next), start)
			continue
		}

		if equal(tok, "-") {
			node = newSub(node, mul(&tok, tok.next), start)
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
		start := tok
		if equal(tok, "*") {
			node = newBinary(ND_MUL, node, unary(&tok, tok.next), start)
			continue
		}

		if equal(tok, "/") {
			node = newBinary(ND_DIV, node, unary(&tok, tok.next), start)
			continue
		}

		*rest = tok
		return node
	}
}

// unary = ("+" | "-" | "*" | "&") unary
//       | postfix

func unary(rest **Token, tok *Token) *Node {
	if equal(tok, "+") {
		return unary(rest, tok.next)
	}
	if equal(tok, "-") {
		return newUnary(ND_NEG, unary(rest, tok.next), tok)
	}
	if equal(tok, "&") {
		return newUnary(ND_ADDR, unary(rest, tok.next), tok)
	}
	if equal(tok, "*") {
		return newUnary(ND_DEREF, unary(rest, tok.next), tok)
	}

	return postfix(rest, tok)
}

// postfix = primary ("[" expr "]")*

func postfix(rest **Token, tok *Token) *Node {
	node := primary(&tok, tok)

	for equal(tok, "[") {
		// x[y] is short for *(x+y)
		start := tok
		idx := expr(&tok, tok.next)
		tok = skip(tok, "]")
		node = newUnary(ND_DEREF, newAdd(node, idx, start), start)
	}
	*rest = tok
	return node
}

// funcall = ident "(" (assign ("," assign)*)? ")"

func funcall(rest **Token, tok *Token) *Node {
	start := tok
	tok = tok.next.next

	head := new(Node)
	cur := head

	for !equal(tok, ")") {
		if cur != head {
			tok = skip(tok, ",")
		}
		cur.next = assign(&tok, tok)
		cur = cur.next
	}

	*rest = skip(tok, ")")

	node := newNode(ND_FUNCALL, start)
	node.funcname = currentInput[start.loc : start.loc+start.len]
	node.args = head.next
	return node
}

// primary = "(" expr ")" | ident func-args? | str | num

func primary(rest **Token, tok *Token) *Node {
	if equal(tok, "(") {
		node := expr(&tok, tok.next)
		*rest = skip(tok, ")")
		return node
	}

	if tok.kind == TK_IDENT {
		// Function call
		if equal(tok.next, "(") {
			return funcall(rest, tok)
		}

		// Variable
		vr := findVar(tok)
		if vr == nil {
			errorTok(tok, "undefined variable")
		}
		*rest = tok.next
		return newVarNode(vr, tok)
	}

	if tok.kind == TK_STR {
		vr := newStringLiteral(tok.str, tok.ty)
		*rest = tok.next
		return newVarNode(vr, tok)
	}

	if tok.kind == TK_NUM {
		node := newNum(tok.val, tok)
		*rest = tok.next
		return node
	}

	errorTok(tok, "expected an expression")
	return nil
}

func createParamLvars(param *Type) {
	if param != nil {
		createParamLvars(param.next)
		newLvar(getIdent(param.name), param)
	}
}

// function = "func" ident type-suffix "{"

func function(rest **Token, tok *Token) *Token {
	if !equal(tok, "func") {
		errorTok(tok, "unexpected declaration is found")
	}
	tok = tok.next
	if tok.kind != TK_IDENT {
		errorTok(tok, "expected a variable name")
	}

	params := typeSuffix(rest, tok.next)
	returnTy := declarator(rest, *rest)
	ty := funcType(returnTy)
	ty.params = params
	ty.name = tok
	locals = nil
	fn := newGvar(getIdent(ty.name), ty)
	fn.isFunction = true
	enterScope()
	createParamLvars(ty.params)
	fn.params = locals

	tok = skip(*rest, "{")
	fn.body = componentStmt(&tok, tok)
	fn.locals = locals
	leaveScope()
	return tok
}

func storeIdentTemp(rest **Token, tok *Token) *Node {
	tok = skip(tok, "var")
	vrs_head := new(Node)
	vrs_cur := vrs_head
	for tok.kind == TK_IDENT {
		vrs := newNode(ND_VAR, tok)
		vrs_cur.next = vrs
		vrs_cur = vrs_cur.next
		tok = tok.next
		if isTypename(tok) || equal(tok, "*") || equal(tok, "[") {
			break
		}
		tok = skip(tok, ",")
	}
	*rest = tok
	return vrs_head.next
}

func globalVariable(tok *Token) *Token {
	vrs_head := storeIdentTemp(&tok, tok)
	ty := declarator(&tok, tok)
	if consume(&tok, tok, "=") {
		for vr_cur := vrs_head; vr_cur != nil; vr_cur = vr_cur.next {
			vr := newGvar(getIdent(vr_cur.tok), ty)
			lhs := newVarNode(vr, ty.name)
			rhs := assign(&tok, tok)
			node := newBinary(ND_ASSIGN, lhs, rhs, tok)
			vr.init = node
			consume(&tok, tok, ",")
		}
	} else {
		for vr_cur := vrs_head; vr_cur != nil; vr_cur = vr_cur.next {
			newGvar(getIdent(vr_cur.tok), ty)
		}
	}
	tok = skip(tok, ";")
	return tok
}

// program = (function-definition | global-variable)*

func parse(tok *Token) *Obj {
	globals = nil

	for tok.kind != TK_EOF {
		// Function
		if equal(tok, "func") {
			tok = function(&tok, tok)
			continue
		}

		// Global variable
		tok = globalVariable(tok)
	}

	return globals
}
