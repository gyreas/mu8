package main

import (
	"math/rand"
	"os"
)

type Cpu struct {
	ip int
	i  uint16
	M  [MEMORY_SIZE]uint8

	R [16]uint8

	sp int
	S  [RETSTACK_SIZE]uint16
}

func (cpu *Cpu) fetch() uint16 {
	inst := uint16(cpu.M[cpu.ip])<<0x08 | uint16(cpu.M[cpu.ip+1])
	cpu.ip += 2
	return inst
}

func (cpu *Cpu) decode_execute(Mu8 *Mu8, inst uint16) {
	mem := cpu.M

	high := uint8(inst >> 0x08)
	low := uint8(inst & 0x00ff)
	nnn := inst & 0x0fff
	y := low >> 0x04
	x := high >> 0x04
	kk := low
	n := low & 0x0f

	switch high >> 0x04 {
	case 0x0:
		switch kk {
		case 0xe0:
			// clear_fb()
			logmsg("CLS\n")
		case 0xee:
			cpu.sp--
			cpu.ip = int(cpu.S[cpu.sp]) // goto next instruction after caller
			logmsg("RET -> 0x%.4x\n", cpu.ip)
		default:
			logmsg("SYS 0x%.04x\n", nnn)

		}
	case 0x1:
		cpu.ip = int(nnn)
		logmsg("JUMP -> 0x%.4x\n", cpu.ip)
		os.Exit(1)
	case 0x2:
		/* cpu.ip is the return address because `fetch()` */
		cpu.S[cpu.sp] = uint16(cpu.ip)
		cpu.sp += 1 //update the return S ptr

		pre_ip := cpu.ip
		cpu.ip = int(nnn)
		logmsg("CALL (0x%.4x) [0x%.4x]\n", cpu.ip, pre_ip)
	case 0x3:
		if cpu.R[x] == kk {
			cpu.ip += 2
		}
		logmsg("SKIP.EQ V%d, 0x%.4x\n", x, kk)
	case 0x4:
		if cpu.R[x] != kk {
			cpu.ip += 2
		}
		logmsg("SKIP.NEQ V%d, 0x%.4x\n", x, kk)
	case 0x5:
		if cpu.R[x] == cpu.R[y] {
			cpu.ip += 2
		}
		logmsg("SKIP.EQ V%d, V%d\n", x, y)
	case 0x6:
		cpu.R[x] = kk
		logmsg("LD V%d, 0x%.4x\n", x, kk)
	case 0x7:
		cpu.R[x] += kk
		logmsg("ADD V%d, 0x%.4x\n", x, kk)
	case 0x8:
		switch n {
		case 0:
			cpu.R[x] = cpu.R[y]
			logmsg("LD V%d, V%d\n", x, y)
		case 1:
			cpu.R[x] |= cpu.R[y]
			logmsg("OR V%d, V%d\n", x, y)
		case 2:
			cpu.R[x] &= cpu.R[y]
			logmsg("AND V%d, V%d\n", x, y)
		case 3:
			cpu.R[x] ^= cpu.R[y]
			logmsg("XOR V%d, V%d\n", x, y)
		case 4:
			sum := uint16(cpu.R[x]) + uint16(cpu.R[y])
			cpu.R[0x0f] = 0x00
			if sum > 0xff {
				cpu.R[0x0f] = 0x01
			}
			cpu.R[x] = uint8(sum & 0x00ff)
			logmsg("ADD.VF V%d, V%d\n", x, y)
		case 5:
			cpu.R[x] -= cpu.R[y]
			cpu.R[0x0f] = 0x00
			if cpu.R[x] > cpu.R[y] {
				cpu.R[0x0f] = 0x01
			}
			logmsg("SUB V%d, V%d\n", x, y)
		default:
			logmsg("error: unknown byte: [0x%.2x] (0x80)\n", low)
		}
	case 0x9:
		if cpu.R[x] != cpu.R[y] {
			cpu.ip += 2
		}
		logmsg("SKIP.NEQ V%d, V%d\n", x, kk)
	case 0xa:
		cpu.i = nnn
		logmsg("LD I, 0x%.4x\n", nnn)
	case 0xb:
		cpu.ip = int(nnn) + int(cpu.R[0x00])
		logmsg("JUMP V0, 0x%.4x\n", nnn)
	case 0xc:
		{
			b := make([]uint8, 10)
			_, err := rand.Read(b)
			if err != nil {
				logmsg("error: %s\n", err.Error())
				return
			}
			randb := b[4] /* 4 is the charm */
			cpu.R[x] = randb & kk
			logmsg("RAND V%d, %d\n", x, randb)
		}
	case 0xd:
		{
			cx := int(cpu.R[x])
			cy := int(cpu.R[y])
			// TODO: track where the I register is set
			// TODO: Set a flag to update the frame buffer here
			logmsg("DRAW V%d, V%d, %d bytes\n", cx, cy, n)
		}
	case 0xe:
		switch low {
		case 0x9e:
			logmsg("SKIP.KP V%d\n", x)
		case 0xa1:
			logmsg("SKIP.KNP V%d\n", x)
		default:
			logmsg("error: unknown byte: [0x%.2x] (0xe0)\n", low)
		}
	case 0xf:
		switch low {
		case 0x00:
			logmsg("STOP")
			return
		case 0x07:
			logmsg("LD V%d, DT\n", x)
		case 0x0a:
			logmsg("LD V%d, KEYPRESS\n", x)
		case 0x15:
			logmsg("LD DT,  V%d\n", x)
		case 0x18:
			logmsg("LD ST,  V%d\n", x)
		case 0x1e:
			cpu.i += uint16(cpu.R[x])
			logmsg("ADD I, V%d\n", x)
		case 0x29:
			cpu.i = FONT_DATA_BASE_ADDRESS + uint16(0x05*cpu.R[x])
			logmsg("LD F, V%d\n", x)
		case 0x33:
			dx := cpu.R[x]
			mem[cpu.i] = dx / 100
			mem[cpu.i+1] = (dx / 100) % 10
			mem[cpu.i+2] = dx / 100
			logmsg("BCD V%d\n", x)
		case 0x55:
			if int(x+1) != len(mem[cpu.i:][:x+1]) {
				logmsg(
					"error: unequal lengths: want: %d, got: %d\n",
					int(x+1), len(mem[cpu.i:][:x+1]))
				os.Exit(1)
			}
			copy(mem[cpu.i:][:x+1], cpu.R[:x+1])
			logmsg("LD [I:], V%d\n", x)
			cpu.i += uint16(x + 1)
		case 0x65:
			if int(x+1) != len(mem[cpu.i:][:x+1]) {
				logmsg(
					"error: unequal lengths: want: %d, got: %d\n",
					int(x+1), len(mem[cpu.i:][:x+1]))
				os.Exit(1)
			}

			copy(cpu.R[:x+1], mem[cpu.i:][:x+1])
			logmsg("LD V%d, [I:]\n", x)
			cpu.i += uint16(x + 1)
		default:
			logmsg("error: unknown byte: [0x%.2x] (0xf0)\n", low)
		}
	default:
		logmsg("error: unknown byte: [0x%.2x] (general)\n", low)
	}
}
