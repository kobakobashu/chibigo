package main

//
// Type
//

type TypeKind int

const (
	TY_INT TypeKind = iota
	TY_PTR
)

type Type struct {
	kind TypeKind
	base *Type  // Pointer
	name *Token // Declaration
}

var tyInt = &Type{kind: TY_INT}

func isInteger(ty *Type) bool {
	return ty.kind == TY_INT
}

func pointerTo(base *Type) *Type {
	ty := new(Type)
	ty.kind = TY_PTR
	ty.base = base
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

	switch node.kind {
	case ND_ADD, ND_SUB, ND_MUL, ND_DIV, ND_NEG, ND_ASSIGN:
		node.ty = node.lhs.ty
		return
	case ND_EQ, ND_NE, ND_LT, ND_LE, ND_NUM, ND_FUNCALL:
		node.ty = tyInt
		return
	case ND_VAR:
		node.ty = node.vr.ty
		return
	case ND_ADDR:
		node.ty = pointerTo(node.lhs.ty)
		return
	case ND_DEREF:
		if node.lhs.ty.kind != TY_PTR {
			errorTok(node.tok, "invalid pointer dereference")
		}
		node.ty = node.lhs.ty.base
		return
	}
}
