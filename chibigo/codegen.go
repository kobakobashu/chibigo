package main

import (
	"fmt"
)

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