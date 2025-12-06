package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ccc - CleanClaudeConfig")
		fmt.Println("Usage: ccc <command>")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  clean     Clean up stale projects and orphaned data")
		fmt.Println("  list      List projects and their status")
		os.Exit(0)
	}

	fmt.Println("Not implemented yet")
	os.Exit(1)
}
