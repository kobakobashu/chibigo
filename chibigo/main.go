package main

import (
	"fmt"
	"os"
	"strconv"
)

func extractNum(s string, idx int) (int, int, error) {
	numericPart := 0
	for cur := idx; cur < len(s); cur++ {
		nextChar := string(s[cur])
		num, err := strconv.Atoi(nextChar)
		if err != nil {
			return numericPart, cur, nil
		}
		numericPart = numericPart*10 + num
	}
	return numericPart, len(s), nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Invalid arguments number")
		return
	}
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".globl main\n")
	fmt.Printf("main:\n")

	idx := 0
	str := os.Args[1]

	num, newIdx, err := extractNum(str, idx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	idx = newIdx

	fmt.Printf("  mov rax, %d\n", num)

	for idx < len(str) {
		op := string(str[idx])
		if op == "+" || op == "-" {
			idx++
			num, idx, err = extractNum(str, idx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
			if op == "+" {
				fmt.Printf("  add rax, %d\n", num)
			} else {
				fmt.Printf("  sub rax, %d\n", num)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Found invalid symbol: %s\n", op)
			return
		}
	}

	fmt.Printf("  ret\n")
}
