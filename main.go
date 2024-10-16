package main

import (
	"fmt"
	"io/fs"
	"os"
)

type Mu8 struct {
	I uint16 /* Address register */

	/* Memory map:
	 * 0000  –  003F  Stack
	 * 0040  –  004C  Scratchpad
	 * 004D  –  00FF  Unused
	 * 0100  –  01F   Display
	 * 0200  –  0FFF  Program area
	 */
	Mem         [2048]uint8
	Regs        [16]uint8 /* V0-VF */
	ReturnStack [64]uint  /* Subroutine Call Stack */
	retptr      int       /* Call stack pointer */
}

func initMu8() Mu8 {
	return Mu8{
		I:           0,
		Mem:         [2048]uint8{},
		Regs:        [16]uint8{},
		ReturnStack: [64]uint{},
		retptr:      0,
	}
}

func (mu8 *Mu8) loadRom(rom []uint8) {
	if len(rom)&0b100 != 0 {
		fmt.Fprintf(os.Stderr, "warning: possibly faulty ROM\n")
	}
	copy(mu8.Mem[0x200:], rom)
}

func (mu8 *Mu8) interpretRom() {
	program := mu8.Mem[0x0200:]
	sz := uint(len(program))

	var code uint8
	var ip uint
cycle:
	for ip = 0; ip < sz; ip += 2 {
		code = program[ip]

		switch code & 0xf0 {
		case 0x00:
			switch program[ip+1] {
			case 0x00:
				continue
			case 0xe0:
				fmt.Println("Clear")
			case 0xee:
				if mu8.retptr > 0 {
					mu8.retptr--
				}
				ip = mu8.ReturnStack[mu8.retptr] + 2 // goto next instruction after caller
				fmt.Printf("Return: 0x%.4x\n", ip)
				break cycle
			}
		case 0x10:
			fmt.Println("Jump")
		case 0x20:
			h := uint16(code) << 0x08
			l := uint16(program[ip+1])

			mu8.ReturnStack[mu8.retptr] = ip
			mu8.retptr += 1

			ip = uint((h|l)&0x0fff) - 0x0200
			fmt.Printf("Call: 0x%.4x\n", ip)
		case 0x30:
			fmt.Println("Skip = KK")
		case 0x40:
			fmt.Println("Skip != KK")
		case 0x50:
			fmt.Println("Skip != VY")

		case 0x60:
			regX := code & 0x0f
			mu8.Regs[regX] = program[ip+1]
			fmt.Println("Assign")
		case 0x70:
			regX := code & 0x0f
			mu8.Regs[regX] += program[ip+1]
			fmt.Println("Add")

		case 0x80:
			switch program[ip+1] & 1 {
			case 0x00:
				regX := code & 0x0f
				regY := program[ip+1] >> 4
				mu8.Regs[regX] = mu8.Regs[regY]
				fmt.Println("Copy")
			case 0x01:
				regX := code & 0x0f
				regY := program[ip+1] >> 4
				mu8.Regs[regX] |= mu8.Regs[regY]
				fmt.Println("Logical OR")
			case 0x02:
				regX := code & 0x0f
				regY := program[ip+1] >> 4
				mu8.Regs[regX] &= mu8.Regs[regY]
				fmt.Println("Logical AND")
			case 0x03:
				regX := code & 0x0f
				regY := program[ip+1] >> 4
				mu8.Regs[regX] ^= mu8.Regs[regY]
				fmt.Println("Logical XOR")
			case 0x04:
				regX := code & 0x0f
				regY := program[ip+1] >> 4
				mu8.Regs[regX] += mu8.Regs[regY]
				if mu8.Regs[regX] > 0xff {
					mu8.Regs[0x0f] = 1
				}
				fmt.Println("Add VY. Set VF=1")
			case 0x05:
				regX := code & 0x0f
				regY := program[ip+1] >> 4
				mu8.Regs[regX] -= mu8.Regs[regY]
				if mu8.Regs[regX] < mu8.Regs[regY] {
					mu8.Regs[0x0f] = 0
				}
				fmt.Println("Subtract VY. Set VF=0")
			default:
				fmt.Fprintf(os.Stderr, "error: unknown byte: [0x%.2x] (0x80)\n", program[ip+1])
			}
		case 0x90:
			regX := code & 0x0f
			regY := program[ip+1] >> 4
			if mu8.Regs[regX] != mu8.Regs[regY] {
				ip += 2
			}
			fmt.Println("Skip next if VX != VY")
		case 0xa0:
			fmt.Println("Set Mem Pointer")
		case 0xb0:
			fmt.Println("Jump to")
		case 0xc0:
			fmt.Println("Get random byte. AND with KK")
		case 0xd0:
			fmt.Println("Display byte pattern")
		case 0xe0:
			switch program[ip+1] {
			case 0x9e:
				fmt.Println("Skip keydown = VX")
			case 0xa1:
				fmt.Println("Skip keydown != VX")
			default:
				fmt.Fprintf(os.Stderr, "error: unknown byte: [0x%.2x] (0xe0)\n", program[ip+1])
			}
		case 0xf0:
			switch program[ip+1] & 0xff {
			case 0x00:
				fmt.Println("Stop")
				break cycle

			case 0x07:
				fmt.Println("Timer")
			case 0x0a:
				fmt.Println("Input hex")
			case 0x15:
				fmt.Println("Set Time")
			case 0x17:
				fmt.Println("Set Pitch")
			case 0x18:
				fmt.Println("Set Tone")
			case 0x1e:
				fmt.Println("Add to Mem Pointer")
			case 0x29:
				fmt.Println("Set Pointer to show VX")
			case 0x33:
				fmt.Println("Store 3-digit decimal")
			case 0x55:
				fmt.Println("Store V0-VX at I")
			case 0x65:
				fmt.Println("Load V0-VX at I")
			case 0x70:
				fmt.Println("Send data in VX")
			case 0x71:
				fmt.Println("Waits for received data into VX")
			case 0x72:
				fmt.Println("Set baud rate")
			default:
				fmt.Fprintf(os.Stderr, "error: unknown byte: [0x%.2x] (0xf0)\n", program[ip+1])
			}
		default:
			fmt.Fprintf(os.Stderr, "error: unknown byte: [0x%.2x] (general)\n", program[ip+1])
		}
	}

	fmt.Printf("%v\n", mu8.Regs)
	fmt.Printf("%v\n", mu8.ReturnStack)
	fmt.Printf("%v\n", mu8.retptr)
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

	mu8 := initMu8()
	mu8.loadRom(buf)
	mu8.interpretRom()
}
