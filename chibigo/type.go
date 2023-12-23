package main

//
// Type
//

type TypeKind int

const (
	TY_CHAR TypeKind = iota
	TY_INT
	TY_PTR
	TY_FUNC
	TY_ARRAY
)

type Type struct {
	kind     TypeKind
	size     int
	base     *Type  // Pointer
	name     *Token // Declaration
	arrayLen int
	returnTy *Type
	params   *Type
	next     *Type
}

var tyChar = &Type{kind: TY_CHAR, size: 1}
var tyInt = &Type{kind: TY_INT, size: 8}

func isInteger(ty *Type) bool {
	return ty.kind == TY_INT || ty.kind == TY_CHAR
}

func copyType(ty *Type) *Type {
	ret := new(Type)
	*ret = *ty
	return ret
}

func pointerTo(base *Type) *Type {
	ty := new(Type)
	ty.kind = TY_PTR
	ty.size = 8
	ty.base = base
	return ty
}

func funcType(returnTy *Type) *Type {
	ty := new(Type)
	ty.kind = TY_FUNC
	ty.returnTy = returnTy
	return ty
}

func arrayOf(base *Type, len int) *Type {
	ty := new(Type)
	ty.kind = TY_ARRAY
	ty.size = base.size * len
	ty.base = base
	ty.arrayLen = len
	return ty
}

func addType(node *Node) {
	if node == nil || node.ty != nil {
		return
	}

	addType(node.lhs)
	addType(node.rhs)
	addType(node.cond)
	addType(node.then)
	addType(node.els)
	addType(node.init)
	addType(node.inc)

	for n := node.body; n != nil; n = n.next {
		addType(n)
	}
	for n := node.args; n != nil; n = n.next {
		addType(n)
	}

	switch node.kind {
	case ND_ADD, ND_SUB, ND_MUL, ND_DIV, ND_NEG:
		node.ty = node.lhs.ty
		return
	case ND_ASSIGN:
		if node.lhs.ty.kind == TY_ARRAY {
			errorTok(node.lhs.tok, "not an lvalue")
		}
		node.ty = node.lhs.ty
		return
	case ND_EQ, ND_NE, ND_LT, ND_LE, ND_NUM, ND_FUNCALL:
		node.ty = tyInt
		return
	case ND_VAR:
		node.ty = node.vr.ty
		return
	case ND_ADDR:
		if node.lhs.ty.kind == TY_ARRAY {
			node.ty = pointerTo(node.lhs.ty.base)
		} else {
			node.ty = pointerTo(node.lhs.ty)
		}
		return
	case ND_DEREF:
		if node.lhs.ty.base == nil {
			errorTok(node.tok, "invalid pointer dereference")
		}
		node.ty = node.lhs.ty.base
		return
	}
}
