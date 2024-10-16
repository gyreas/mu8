package main

import (
	"fmt"
	"io/fs"
	"os"
)

func main() {
	argv := os.Args
	if len(argv) < 2 {
		fmt.Fprintf(os.Stderr, "usage: ./mu8 <rom>\n")
		os.Exit(1)
	}
	rom := argv[1]
	mu8_fs := os.DirFS(".")
	buf, err := fs.ReadFile(mu8_fs, rom)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
	if len(buf)&0b100 != 0 {
		fmt.Fprintf(os.Stderr, "error: possibly faulty ROM\n")
		os.Exit(1)
	}

	for i, b := range buf {
		fmt.Printf("%.2X", b)
		if i&1 == 1 {
			fmt.Printf(" ")
		}
	}
	fmt.Println()
}
