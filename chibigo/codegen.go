package main

import (
	"fmt"
)

func println(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

//
// Code generator
//

var depth int

var argreg8 = [...]string{"dil", "sil", "dl", "cl", "r8b", "r9b"}
var argreg64 = [...]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
var current_fn *Obj

var counter = count()

func count() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}

func push() {
	println("  push rax")
	depth++
}

func pop(arg string) {
	println("  pop %s", arg)
	depth--
}

func cmp(cmd string) {
	println("  cmp rax, rdi")
	println("  %s al", cmd)
	println("  movzb rax, al")
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
		if node.vr.isLocal == true {
			// Local variable
			println("  lea rax, %d[rbp]", node.vr.offset)
		} else {
			// Global variable
			println("  lea rax, [rip + %s]", node.vr.name)
		}
		return
	case ND_DEREF:
		genExpr(node.lhs)
		return
	}

	errorTok(node.tok, "not an lvalue")
}

// Load a value from where %rax is pointing to.

func load(ty *Type) {
	if ty != nil && ty.kind == TY_ARRAY {
		return
	}
	if ty != nil && ty.size == 1 {
		println("  movsx rax, byte ptr [rax]")
	} else {
		println("  mov rax, [rax]")
	}
}

// Store %rax to an address that the stack top is pointing to.

func store(ty *Type) {
	pop("rdi")
	if ty != nil && ty.size == 1 {
		println("  mov [rdi], al")
	} else {
		println("  mov [rdi], rax")
	}
}

func genExpr(node *Node) {
	switch node.kind {
	case ND_NUM:
		println("  mov rax, %d", node.val)
		return
	case ND_NEG:
		genExpr(node.lhs)
		println("  neg rax")
		return
	case ND_VAR:
		genAddr(node)
		load(node.ty)
		return
	case ND_DEREF:
		genExpr(node.lhs)
		load(node.ty)
		return
	case ND_ADDR:
		genAddr(node.lhs)
		return
	case ND_ASSIGN:
		genAddr(node.lhs)
		push()
		genExpr(node.rhs)
		store(node.ty)
		return
	case ND_FUNCALL:
		nargs := 0
		for arg := node.args; arg != nil; arg = arg.next {
			genExpr(arg)
			push()
			nargs++
		}
		for i := nargs - 1; i >= 0; i-- {
			pop(argreg64[i])
		}
		println("  mov rax, 0")
		println("  call %s", node.funcname)
		return
	}

	genExpr(node.rhs)
	push()
	genExpr(node.lhs)
	pop("rdi")

	switch node.kind {
	case ND_ADD:
		println("  add rax, rdi")
		return
	case ND_SUB:
		println("  sub rax, rdi")
		return
	case ND_MUL:
		println("  imul rax, rdi")
		return
	case ND_DIV:
		println("  cqo")
		println("  idiv rdi")
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
		println("  cmp rax, 0")
		println("  je  .L.else.%d", c)
		genStmt(node.then)
		println("  jmp .L.end.%d", c)
		println(".L.else.%d:", c)
		if node.els != nil {
			genStmt(node.els)
		}
		println(".L.end.%d:", c)
		return
	case ND_FOR:
		c := counter()
		if node.init != nil {
			genExpr(node.init)
		}
		println(".L.begin.%d:", c)
		if node.cond != nil {
			genExpr(node.cond)
			println("  cmp rax, 0")
			println("  je  .L.end.%d", c)
		}
		genStmt(node.then)
		if node.inc != nil {
			genExpr(node.inc)
		}
		println("  jmp .L.begin.%d", c)
		println(".L.end.%d:", c)
		return
	case ND_BLOCK:
		for n := node.body; n != nil; n = n.next {
			genStmt(n)
		}
		return
	case ND_RETURN:
		genExpr(node.lhs)
		println("  jmp .L.return.%s", current_fn.name)
		return
	case ND_EXPR_STMT:
		genExpr(node.lhs)
		return
	}

	errorTok(node.tok, "invalid statement")
	return
}

// Assign offsets to local variables.

func assignLvarOffsets(prog *Obj) {
	for fn := prog; fn != nil; fn = fn.next {
		if fn.isFunction == false {
			continue
		}
		offset := 0
		for vr := fn.locals; vr != nil; vr = vr.next {
			offset += vr.ty.size
			vr.offset = -offset
		}
		fn.stackSize = alignTo(offset, 16)
	}
}

func emitData(prog *Obj) {
	for vr := prog; vr != nil; vr = vr.next {
		if vr.isFunction {
			continue
		}

		println("  .data")
		println("  .globl %s", vr.name)
		println("%s:", vr.name)
		if vr.init != nil {
			if vr.init.rhs != nil && vr.init.rhs.vr != nil && vr.init.rhs.vr.initData != "" {
				for i := 0; i < vr.ty.size; i++ {
					println("  .byte %d", vr.init.rhs.vr.initData[i])
				}
			} else if vr.init.rhs.val != 0 {
				println("  .long %d", vr.init.rhs.val)
			}
		}
		if vr.initData != "" {
			for i := 0; i < vr.ty.size; i++ {
				println("  .byte %d", vr.initData[i])
			}
		} else {
			println("  .zero %d", vr.ty.size)
		}
	}
}

func emitText(prog *Obj) {
	assignLvarOffsets(prog)

	println(".intel_syntax noprefix")
	for fn := prog; fn != nil; fn = fn.next {
		if fn.isFunction == false {
			continue
		}

		println(".globl %s", fn.name)
		println(".text")
		println("%s:", fn.name)
		current_fn = fn

		// Prologue
		println("  push rbp")
		println("  mov rbp, rsp")
		println("  sub rsp, %d", fn.stackSize)

		// Save passed-by-register arguments to the stack
		i := 0
		for vr := fn.params; vr != nil; vr = vr.next {
			if vr != nil && vr.ty != nil && vr.ty.size == 1 {
				println("  mov %d[rbp], %s", vr.offset, argreg8[i])
			} else {
				println("  mov %d[rbp], %s", vr.offset, argreg64[i])
			}
			i++
		}

		// Emit code
		genStmt(fn.body)
		if depth != 0 {
			panic("Depth is not zero")
		}

		// Epilogue
		println(".L.return.%s:", fn.name)
		println("  mov rsp, rbp")
		println("  pop rbp")
		println("  ret")
	}
}

func codegen(prog *Obj) {
	assignLvarOffsets(prog)
	emitData(prog)
	emitText(prog)
}
