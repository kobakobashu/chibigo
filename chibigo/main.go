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

	tok, err := tokenize(os.Args[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	prog := parse(tok)

	// Traverse the AST to emit assembly.
	codegen(prog)
}
