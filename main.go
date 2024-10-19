package main

import (
	"bufio"
	"crypto/rand"
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
	clear_fb()

	return Mu8{
		I:           0,
		Mem:         [2048]uint8{},
		Regs:        [16]uint8{},
		ReturnStack: [64]uint{},
		retptr:      0,
	}
}

const (
	FONT_DATA_BASE_ADDRESS uint16 = 0x50
	PROGRAM_ADDRESS_OFFSET uint   = 0x0200
)

func (mu8 *Mu8) loadRom(rom []uint8) {
	if len(rom)&0b100 != 0 {
		logger().Println("warning: possibly faulty ROM")
	}
	fmt.Printf("ROM size=0x%.4x\n", len(rom))

	copy(mu8.Mem[FONT_DATA_BASE_ADDRESS:][:len(ds)], ds)
	copy(mu8.Mem[PROGRAM_ADDRESS_OFFSET:], rom)
}

func (mu8 *Mu8) interpretRom() {
	logger := logger()
	program := mu8.Mem[:]

	var code uint8
	var ip uint
cycle:
	for ip = PROGRAM_ADDRESS_OFFSET; ip < uint(len(mu8.Mem)); ip += 2 {
		code = program[ip]
		h1 := code & 0x0f
		l := program[ip+1]
		l0 := l >> 0x04
		l1 := l & 0x0f

		switch code & 0xf0 {
		case 0x00:
			switch program[ip+1] {
			case 0x00:
				continue
			case 0xe0:
				clear_fb()
				logger.Println("Clear")
			case 0xee:
				if mu8.retptr > 0 {
					mu8.retptr--
				}
				ip = mu8.ReturnStack[mu8.retptr] + 2 // goto next instruction after caller
				logger.Printf("Return: 0x%.4x\n", ip)
			}
		case 0x10:
			ip = uint((uint16(h1) << 0x08) | uint16(l))
			logger.Printf("Jump: 0x%.4x\n", ip)
		case 0x20:
			mu8.ReturnStack[mu8.retptr] = ip
			mu8.retptr += 1 //update the return stack ptr

			pre_ip := ip
			ip = uint((uint16(h1) << 0x08) | uint16(l))
			logger.Printf("Call: 0x%.4x [0x%.4x]\n", ip, pre_ip)
		case 0x30:
			if mu8.Regs[h1] == l {
				ip += 2
			}
			logger.Printf("Skip =KK: 0x%.2x = 0x%.2x", mu8.Regs[h1], l)
		case 0x40:
			if mu8.Regs[h1] != l {
				ip += 2
			}
			logger.Printf("Skip !=KK: 0x%.2x != 0x%.2x", mu8.Regs[h1], l)
		case 0x50:
			if mu8.Regs[h1] == mu8.Regs[l0] {
				ip += 2
			}
			logger.Printf("Skip =VY: 0x%.2x = 0x%.2x\n", mu8.Regs[h1], mu8.Regs[l0])
		case 0x60:
			logger.Printf("Assign: %v 0x%.4x = 0x%.4x\n", mu8.Regs[:3], mu8.Regs[h1], l)
			mu8.Regs[h1] = l
		case 0x70:
			logger.Printf("Add: 0x%.4x += 0x%.4x\n", mu8.Regs[h1], l)
			mu8.Regs[h1] += l

		case 0x80:
			regX := h1
			regY := l0
			switch program[ip+1] & 1 {
			case 0x00:
				mu8.Regs[regX] = mu8.Regs[regY]
				logger.Println("Copy")
			case 0x01:
				mu8.Regs[regX] |= mu8.Regs[regY]
				logger.Println("Logical OR")
			case 0x02:
				mu8.Regs[regX] &= mu8.Regs[regY]
				logger.Println("Logical AND")
			case 0x03:
				mu8.Regs[regX] ^= mu8.Regs[regY]
				logger.Println("Logical XOR")
			case 0x04:
				if mu8.Regs[regX]+mu8.Regs[regY] > 0xff {
					mu8.Regs[0x0f] = 0x01
				} else {
					mu8.Regs[0x0f] = 0x00
				}
				mu8.Regs[regX] += mu8.Regs[regY]
				logger.Println("Add VY. Set overflow flag VF")
			case 0x05:
				if mu8.Regs[regX] < mu8.Regs[regY] {
					mu8.Regs[0x0f] = 0x00
				} else {
					mu8.Regs[0x0f] = 0x01
				}
				mu8.Regs[regX] -= mu8.Regs[regY]
				logger.Println("Subtract VY. Set carry flag ")
			default:
				fmt.Fprintf(os.Stderr, "error: unknown byte: [0x%.2x] (0x80)\n", program[ip+1])
			}
		case 0x90:
			if mu8.Regs[h1] != mu8.Regs[l0] {
				ip += 2
			}
			logger.Println("Skip next if VX != VY")
		case 0xa0:
			mu8.I = (uint16(h1) << 0x08) | uint16(l)
			logger.Printf("Set Mem Pointer: 0x%.4x\n", mu8.I)
		case 0xb0:
			ip = uint((uint16(h1)<<0x08)|uint16(l)) + uint(mu8.Regs[0x00])
			logger.Println("Jump to Mem Addr + V0")
		case 0xc0:
			{
				c := 10
				b := make([]uint8, c)
				_, err := rand.Read(b)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %s\n", err)
					return
				}
				mu8.Regs[h1] = b[4] /* 4 is the charm */ & l
				logger.Printf("Get random byte. AND with KK(0x%.4x): [0x%.4x]\n", l, mu8.Regs[h1])
			}
		case 0xd0:
			{
				x := int(mu8.Regs[h1])
				y := int(mu8.Regs[l0])
				n := l1

				// TODO: track where the I register is set
				sprite := mu8.Mem[mu8.I:][:n]
				changed_px := draw_sprite_at(sprite, x, y)

				mu8.Regs[0x0f] = 0x00
				if changed_px {
					mu8.Regs[0x0f] = 0x01
				}

				logger.Printf("Display byte pattern: I=0x%.4x, %v\n", mu8.I, mu8.Mem[mu8.I:][:n])
				draw_fb()
				// fmt.Println("--------------------------------")
			}
		case 0xe0:
			switch program[ip+1] {
			case 0x9e:
				{ // TODO: Implement keyboard input polling
					reader := bufio.NewReader(os.Stdin)
					b, err := reader.ReadByte()
					if err != nil {
						fmt.Fprintf(os.Stderr, "error: %s\n", err)
						os.Exit(1)
					}

					if ('0' <= b && b <= '9') || ('a' <= b && b <= 'f') {
						if mu8.Regs[h1] == b {
							ip += 2
						}
						fmt.Println("Input hex", b)
					} else {
						ip -= 2
					}
					logger.Printf("Skip keydown 0x%.2x = 0x%.2x\n", b, mu8.Regs[h1])
				}
			case 0xa1:
				{ // TODO: Implement keyboard input polling
					reader := bufio.NewReader(os.Stdin)
					b, err := reader.ReadByte()
					if err != nil {
						fmt.Fprintf(os.Stderr, "error: %s\n", err)
						os.Exit(1)
					}

					if ('0' <= b && b <= '9') || ('a' <= b && b <= 'f') {
						if mu8.Regs[h1] != b {
							ip += 2
						}
						fmt.Println("Input hex", b)
					} else {
						ip -= 2
					}
					logger.Printf("Skip keydown 0x%.2x != 0x%.2x\n", b, mu8.Regs[h1])
				}
			default:
				fmt.Fprintf(os.Stderr, "error: unknown byte: [0x%.2x] (0xe0)\n", program[ip+1])
			}
		case 0xf0:
			switch program[ip+1] & 0xff {
			case 0x00:
				logger.Println("Stop")
				break cycle

			case 0x07:
				logger.Println("Timer")
			case 0x0a:
				{ // TODO: Implement keyboard input polling
					reader := bufio.NewReader(os.Stdin)
					b, err := reader.ReadByte()
					if err != nil {
						fmt.Fprintf(os.Stderr, "error: %s\n", err)
						os.Exit(1)
					}

					if ('0' <= b && b <= '9') || ('a' <= b && b <= 'f') {
						mu8.Regs[h1] = b
						fmt.Println("Input hex", b)
					} else {
						ip -= 2
					}
				}
			case 0x15:
				logger.Println("Set Time")
			case 0x17:
				logger.Println("Set Pitch")
			case 0x18:
				logger.Println("Set Tone")
			case 0x1e:
				mu8.I += uint16(mu8.Regs[h1])
				logger.Println("Add to Mem Pointer")
			case 0x29:
				mu8.I = FONT_DATA_BASE_ADDRESS + uint16(0x05*mu8.Regs[h1])
				logger.Printf("Set Pointer to show VX: I=0x%.4x\n", mu8.I)
			case 0x33:
				{ // BCD
					b := uint16(mu8.Regs[h1])
					i := uint16(0)
					lower := uint16(100)

					logger.Println("Store 3-digit decimal")
					logger.Printf("[%.3d](I=0x%.4x) d=", b, mu8.I)
					for lower != 0 {
						logger.Printf("%d ", b/lower)
						program[mu8.I+i] = uint8(b / lower)
						b %= lower
						lower /= 10
						i++
					}
					logger.Println()
				}
			case 0x55:
				if int(h1+1) != len(mu8.Mem[mu8.I:][:h1+1]) {
					fmt.Fprintf(os.Stderr, "error: unequal lengths\n")
					os.Exit(1)
				}

				copy(program[mu8.I:][:h1+1], mu8.Regs[:h1+1])
				logger.Println("Store V0-VX at I:", mu8.Regs[:h1+1], program[mu8.I:][:h1+1])
				mu8.I += uint16(h1 + 1)
			case 0x65:
				if int(h1+1) != len(mu8.Mem[mu8.I:][:h1+1]) {
					fmt.Fprintf(os.Stderr, "error: unequal lengths\n")
					os.Exit(1)
				}

				copy(mu8.Regs[:h1+1], program[mu8.I:][:h1+1])
				logger.Println("Load V0-VX at I:", mu8.Regs[:h1+1], program[mu8.I:][:h1+1])
				mu8.I += uint16(h1 + 1)
			case 0x70:
				logger.Println("Send data in VX")
			case 0x71:
				logger.Println("Waits for received data into VX")
			case 0x72:
				logger.Println("Set baud rate")
			default:
				fmt.Fprintf(os.Stderr, "error: unknown byte: [0x%.2x] (0xf0)\n", program[ip+1])
			}
		default:
			fmt.Fprintf(os.Stderr, "error: unknown byte: [0x%.2x] (general)\n", program[ip+1])
		}
	}

	logger.Println()
	// fmt.Printf("I: 0x%.4x\n", mu8.I)
	// fmt.Printf("Regs: %v\n", mu8.Regs)
	// fmt.Printf("RetStack: %v\n", mu8.ReturnStack)
	// fmt.Printf("Retptr: %v\n", mu8.retptr)

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

	init_logger()
	mu8 := initMu8()

	logger().Println("Loading program from", rom)

	mu8.loadRom(buf)
	mu8.interpretRom()
	// draw_digits()

}

