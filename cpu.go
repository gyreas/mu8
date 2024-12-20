package main

import (
	"math/rand"
)

type Sprite struct {
	x, y uint8
	data []byte
}

type Cpu struct {
	Ip int
	i  uint16
	M  [MEMORY_SIZE]uint8

	sprite  Sprite
	key     uint8
	has_key bool
	clear   bool

	delay uint8
	R     [16]uint8

	sp int
	S  [RETSTACK_SIZE]uint16
}

func NewCpu() Cpu {
	cpu := Cpu{}
	cpu.i = 0
	cpu.Ip = PROGRAM_ADDRESS_OFFSET
	copy(cpu.M[FONT_DATA_OFFSET:][:len(Digits)], Digits)
	cpu.sprite.data = nil

	return cpu
}

func (cpu *Cpu) Fetch() uint16 {
	inst := uint16(cpu.M[cpu.Ip])<<0x08 | uint16(cpu.M[cpu.Ip+1])
	cpu.Ip += 2
	return inst
}

func (cpu *Cpu) Cycle(Mu8 *Mu8, collide, key <-chan uint8, echan chan<- Event) {
	// see if any key was pressed
	// TODO: make a small deadline for this instead of instantaneous
	{
		select {
		case k := <-key: // got a key
			// now, the key must correspond to its index in `cpu.key`
			logmsg("[cycle]: got key: %d\n", k)
			cpu.key = k
			cpu.has_key = true
		default:
			cpu.has_key = false
		}
	}

	inst := cpu.Fetch()
	logmsg("%.4x | [%.4x] \n", cpu.Ip-2, inst)

	cpu.DecodeExecute(Mu8, inst)

	if cpu.clear {
		echan <- Event{Kind: EventClear}
		cpu.clear = false
	}

	if cpu.sprite.data != nil {
		echan <- Event{Kind: EventSprite, Sprite: cpu.sprite}
		cpu.R[0x0f] = <-collide

		logmsg("Rendered FB from Cpu\n")

		// invalidate the data
		cpu.sprite.data = nil
	}

	if cpu.delay > 0 {
		cpu.delay--
	}
}

const CHARM = 0x44

func (cpu *Cpu) DecodeExecute(Mu8 *Mu8, inst uint16) {
	mem := &cpu.M

	high := uint8(inst >> 0x08)
	low := uint8(inst & 0x00ff)
	nnn := inst & 0x0fff
	x := high & 0x0f
	y := low >> 0x04
	kk := low
	n := low & 0x0f

	switch high >> 0x04 {
	case 0x0:
		switch kk {
		case 0xe0:
			cpu.clear = true
			logmsg("CLS\n")
		case 0xee:
			cpu.sp--
			cpu.Ip = int(cpu.S[cpu.sp]) // goto next instruction after caller
			logmsg("RET -> 0x%.4x\n", cpu.Ip)
		default:
			logmsg("SYS 0x%.04x\n", nnn)

		}
	case 0x1:
		cpu.Ip = int(nnn)
		logmsg("JUMP -> 0x%.4x\n", cpu.Ip)
		// os.Exit(1)
	case 0x2:
		/* cpu.Ip is the return address because `fetch()` */
		cpu.S[cpu.sp] = uint16(cpu.Ip)
		cpu.sp += 1 //update the return S ptr

		pre_ip := cpu.Ip
		cpu.Ip = int(nnn)
		logmsg("CALL (0x%.4x) [0x%.4x]\n", cpu.Ip, pre_ip)
	case 0x3:
		if cpu.R[x] == kk {
			cpu.Ip += 2
		}
		logmsg("SKIP.EQ V%d, 0x%.4x\n", x, kk)
	case 0x4:
		if cpu.R[x] != kk {
			cpu.Ip += 2
		}
		logmsg("SKIP.NEQ V%d, 0x%.4x\n", x, kk)
	case 0x5:
		if cpu.R[x] == cpu.R[y] {
			cpu.Ip += 2
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
			cpu.Ip += 2
		}
		logmsg("SKIP.NEQ V%d, V%d\n", x, kk)
	case 0xa:
		cpu.i = nnn
		logmsg("LD I, 0x%.4x\n", nnn)
	case 0xb:
		cpu.Ip = int(nnn) + int(cpu.R[0x00])
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

			logmsg("DRAW (0x%.4x) V%.1x(%d), V%.1x(%d), %d bytes\n", cpu.i, x, cx, y, cy, n)

			cpu.sprite.x = uint8(cx)
			cpu.sprite.y = uint8(cy)
			cpu.sprite.data = cpu.M[cpu.i:][:n]

			logmsg("Sprite: %v\n", cpu.sprite)
		}
	case 0xe:
		switch low {
		case 0x9e:
			logmsg("SKIP.KP V%d\n", x)
			if cpu.R[x] == cpu.key {
				cpu.Ip += 2
			}
		case 0xa1:
			logmsg("SKIP.KNP V%d\n", x)
			if cpu.R[x] != cpu.key {
				cpu.Ip += 2
			}
		default:
			logmsg("error: unknown byte: [0x%.2x] (0xe0)\n", low)
		}
	case 0xf:
		switch low {
		case 0x00:
			logmsg("STOP")
			return
		case 0x07:
			logmsg("LD V%d, DT(%d)\n", x, cpu.delay)
			cpu.R[x] = cpu.delay
		case 0x0a:
			if cpu.has_key {
				logmsg("LD V%d, KEYPRESS: %x\n", x, cpu.key)
				cpu.R[x] = cpu.key
				cpu.has_key = false
			} else {
				// spin on this instruction until a key is pressed
				cpu.Ip -= 2
			}
		case 0x15:
			logmsg("LD DT(%d),  V%d\n", cpu.delay, x)
			cpu.delay = cpu.R[x]
		case 0x18:
			logmsg("LD ST,  V%d\n", x)
		case 0x1e:
			cpu.i += uint16(cpu.R[x])
			logmsg("ADD I, V%d\n", x)
		case 0x29:
			cpu.i = FONT_DATA_OFFSET + uint16(0x05*cpu.R[x])
			logmsg("LD F, V%d: %x\n", x, cpu.R[x])
		case 0x33:
			dx := cpu.R[x]
			mem[cpu.i] = dx / 100
			mem[cpu.i+1] = (dx / 10) % 10
			mem[cpu.i+2] = dx % 10
			logmsg("BCD (0x%.4x)%v V%d\n", cpu.i, mem[cpu.i:][:3], x)
		case 0x55:
			copy(mem[cpu.i:][:x+1], cpu.R[:x+1])
			logmsg("LD [I:], V%d: %v\n", x, cpu.R[:x+1])
			cpu.i += uint16(x + 1)
		case 0x65:
			logmsg("LD V%d, [I(0x%.4x):]: %v/%v\n", x, cpu.i, cpu.R[:x+1], mem[cpu.i:][:x+1])
			copy(cpu.R[:x+1], mem[cpu.i:][:x+1])
			cpu.i += uint16(x + 1)
		default:
			logmsg("error: unknown byte: [0x%.2x] (0xf0)\n", low)
		}
	default:
		logmsg("error: unknown byte: [0x%.2x] (general)\n", low)
	}
}
