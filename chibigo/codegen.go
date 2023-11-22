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

// Compute the absolute address of a given node.
// It's an error if a given node does not reside in memory.

func genAddr(node *Node) {
	if node.kind == ND_VAR {
		///
		offset := (int(node.name[0]) - int('a') + 1) * 8
		fmt.Printf("  lea rax, %d[rbp]\n", -offset)
		return
	}

	errorf("not an lvalue")
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
	case ND_VAR:
		genAddr(node)
		fmt.Printf("  mov rax, [rax]\n")
		return
	case ND_ASSIGN:
		genAddr(node.lhs)
		push()
		genExpr(node.rhs)
		pop("rdi")
		fmt.Printf("  mov [rdi], rax\n")
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

func genStmt(node *Node) {
	if node.kind == ND_EXPR_STMT {
		genExpr(node.lhs)
		return
	}

	errorf("invalid statement")
	return
}

func codegen(node *Node) {
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".globl main\n")
	fmt.Printf("main:\n")

	fmt.Printf("  push rbp\n")
	fmt.Printf("  mov rbp, rsp\n")
	fmt.Printf("  sub rsp, 208\n")

	for n := node; n != nil; n = n.next {
		// Traverse the AST to emit assembly.
		genStmt(n)
		if depth != 0 {
			panic("Depth is not zero")
		}
	}

	fmt.Printf("  mov rsp, rbp\n")
	fmt.Printf("  pop rbp\n")
	fmt.Printf("  ret\n")
}
