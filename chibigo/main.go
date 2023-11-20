package main

import (
	"fmt"
	"os"
)

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
