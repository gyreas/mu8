package main

import (
	"fmt"
	"io/fs"
	"os"
)

type Mu8 struct {
	Mem  [2048]byte
	Regs [16]byte
}

func initMu8() Mu8 {
	return Mu8{
		Mem:  [2048]byte{},
		Regs: [16]byte{},
	}
}

func (mu8 *Mu8) loadRom(rom []byte) {
	copy(mu8.Mem[0x200:], rom)
}

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

	mu8 := initMu8()
	mu8.loadRom(buf)
}
