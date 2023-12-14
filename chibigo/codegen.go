package main

import (
	"fmt"
)

//
// Code generator
//

var depth int
var argreg = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
var current_fn *Function

var counter = count()

func count() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}

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

// Round up `n` to the nearest multiple of `align`. For instance,
// align_to(5, 8) returns 8 and align_to(11, 8) returns 16.

func alignTo(n int, align int) int {
	return (n + align - 1) / align * align
}

// Compute the absolute address of a given node.
// It's an error if a given node does not reside in memory.

func genAddr(node *Node) {
	switch node.kind {
	case ND_VAR:
		fmt.Printf("  lea rax, %d[rbp]\n", node.vr.offset)
		return
	case ND_DEREF:
		genExpr(node.lhs)
		return
	}

	errorTok(node.tok, "not an lvalue")
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
	case ND_DEREF:
		genExpr(node.lhs)
		fmt.Printf("  mov rax, [rax]\n")
		return
	case ND_ADDR:
		genAddr(node.lhs)
		return
	case ND_ASSIGN:
		genAddr(node.lhs)
		push()
		genExpr(node.rhs)
		pop("rdi")
		fmt.Printf("  mov [rdi], rax\n")
		return
	case ND_FUNCALL:
		nargs := 0
		for arg := node.args; arg != nil; arg = arg.next {
			genExpr(arg)
			push()
			nargs++
		}
		for i := nargs - 1; i >= 0; i-- {
			pop(argreg[i])
		}
		fmt.Printf("  mov rax, 0\n")
		fmt.Printf("  call %s\n", node.funcname)
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

	errorTok(node.tok, "invalid expression")
	return
}

func genStmt(node *Node) {
	switch node.kind {
	case ND_IF:
		c := counter()
		genExpr(node.cond)
		fmt.Printf("  cmp rax, 0\n")
		fmt.Printf("  je  .L.else.%d\n", c)
		genStmt(node.then)
		fmt.Printf("  jmp .L.end.%d\n", c)
		fmt.Printf(".L.else.%d:\n", c)
		if node.els != nil {
			genStmt(node.els)
		}
		fmt.Printf(".L.end.%d:\n", c)
		return
	case ND_FOR:
		c := counter()
		if node.init != nil {
			genExpr(node.init)
		}
		fmt.Printf(".L.begin.%d:\n", c)
		if node.cond != nil {
			genExpr(node.cond)
			fmt.Printf("  cmp rax, 0\n")
			fmt.Printf("  je  .L.end.%d\n", c)
		}
		genStmt(node.then)
		if node.inc != nil {
			genExpr(node.inc)
		}
		fmt.Printf("  jmp .L.begin.%d\n", c)
		fmt.Printf(".L.end.%d:\n", c)
		return
	case ND_BLOCK:
		for n := node.body; n != nil; n = n.next {
			genStmt(n)
		}
		return
	case ND_RETURN:
		genExpr(node.lhs)
		fmt.Printf("  jmp .L.return.%s\n", current_fn.name)
		return
	case ND_EXPR_STMT:
		genExpr(node.lhs)
		return
	}

	errorTok(node.tok, "invalid statement")
	return
}

// Assign offsets to local variables.

func assignLvarOffsets(prog *Function) {
	for fn := prog; fn != nil; fn = fn.next {
		offset := 0
		for vr := fn.locals; vr != nil; vr = vr.next {
			offset += 8
			vr.offset = -offset
		}
		fn.stackSize = alignTo(offset, 16)
	}
}

func codegen(prog *Function) {
	assignLvarOffsets(prog)

	fmt.Printf(".intel_syntax noprefix\n")
	for fn := prog; fn != nil; fn = fn.next {
		fmt.Printf(".globl %s\n", fn.name)
		fmt.Printf("%s:\n", fn.name)
		current_fn = fn

		// Prologue
		fmt.Printf("  push rbp\n")
		fmt.Printf("  mov rbp, rsp\n")
		fmt.Printf("  sub rsp, %d\n", fn.stackSize)

		// Save passed-by-register arguments to the stack
		i := 0
		for vr := fn.params; vr != nil; vr = vr.next {
			fmt.Printf("  mov %d[rbp], %s\n", vr.offset, argreg[i])
			i++
		}

		// Emit code
		genStmt(fn.body)
		if depth != 0 {
			panic("Depth is not zero")
		}

		// Epilogue
		fmt.Printf(".L.return.%s:\n", fn.name)
		fmt.Printf("  mov rsp, rbp\n")
		fmt.Printf("  pop rbp\n")
		fmt.Printf("  ret\n")
	}
}
