package main

import (
	"fmt"
)

const (
	BORDER_W = 1 /* Frame buffer padding */
	WIDTH    = 64 /* Frame buffer dimensions */ + 2*BORDER_W
	HEIGHT   = 32 /* Frame buffer dimensions */ + 2*BORDER_W
)

const (
	CELL_EMPTY          = " "
	CELL_FILLED         = "🮆"
	BORDER_HORI         = "─"
	BORDER_VERT         = "│"
	BORDER_LEFT_TOP     = "┌"
	BORDER_RIGHT_TOP    = "┐"
	BORDER_LEFT_BOTTOM  = "└"
	BORDER_RIGHT_BOTTOM = "┘"
)

var fb = [HEIGHT * WIDTH]byte{}

var ds = [][]uint8{
	{0xf0, 0x90, 0x90, 0x90, 0xf0},
	{0x20, 0x60, 0x20, 0x20, 0x70},
	{0xf0, 0x10, 0xf0, 0x80, 0xf0},
	{0xf0, 0x10, 0xf0, 0x10, 0xf0},
	{0x90, 0x90, 0xf0, 0x10, 0x10},
	{0xf0, 0x80, 0xf0, 0x10, 0xf0},
	{0xf0, 0x80, 0xf0, 0x90, 0xf0},
	{0xf0, 0x10, 0x20, 0x40, 0x40},
	{0xf0, 0x90, 0xf0, 0x90, 0xf0},
	{0xf0, 0x90, 0xf0, 0x10, 0xf0},
	{0xf0, 0x90, 0xf0, 0x90, 0x90},
	{0xe0, 0x90, 0xe0, 0x90, 0xe0},
	{0xf0, 0x80, 0x80, 0x80, 0xf0},
	{0xe0, 0x90, 0x90, 0x90, 0xe0},
	{0xf0, 0x80, 0xf0, 0x80, 0xf0},
	{0xf0, 0x80, 0xf0, 0x80, 0x80},
}

func clear_fb() {
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			fb[y*WIDTH+x] = ' '
		}
	}
	draw_border_fb()
}

func draw_border_fb() {
	for y := 0; y < HEIGHT; y++ {
		fb[y*WIDTH] = '|'
		fb[(y+1)*WIDTH-1] = '|'
	}
	for x := 0; x < WIDTH; x++ {
		fb[x] = '-'
		fb[(HEIGHT-1)*WIDTH+x] = '-'
	}
	fb[0] = '^'
	fb[WIDTH-1] = '+'
	fb[HEIGHT*WIDTH-1] = '>'
	fb[(HEIGHT-1)*WIDTH] = 'v'
}

func draw_fb() {
	for y := 0; y < HEIGHT; y++ {
		row := y * WIDTH
		for x := 0; x < WIDTH; x++ {
			switch fb[row+x] {
			case '*':
				fmt.Printf(CELL_FILLED)
			case '-':
				fmt.Printf(BORDER_HORI)
			case '|':
				fmt.Printf(BORDER_VERT)
			case '^':
				fmt.Printf(BORDER_LEFT_TOP)
			case '+':
				fmt.Printf(BORDER_RIGHT_TOP)
			case 'v':
				fmt.Printf(BORDER_LEFT_BOTTOM)
			case '>':
				fmt.Printf(BORDER_RIGHT_BOTTOM)
			default:
				fmt.Printf(CELL_EMPTY)
			}
		}
		fmt.Println()
	}
}

func draw_sprite_at(sprite []byte, x, y int) bool {
	changed_px := false
	for s := 0; s < len(sprite); s++ {
		dat := sprite[s]
		changed_px = draw_byte_at(dat, x, y+s+1)
	}
	return changed_px
}

func draw_byte_at(b byte, x, y int) bool {
	changed_px := false

	idx := y*WIDTH + x
	if x == 0 {
		idx += BORDER_W
	} else if x == WIDTH-1 {
		idx += 2 * BORDER_W
	}

	i := 0
	lower := uint8(1 << 7)
	for lower != 0 {
		if b&lower == 0 {
			if fb[idx+i] == '*' {
				changed_px = true
			}
			// fb[idx+i] = '.'
		} else {
			fb[idx+i] = '*'
		}
		i++
		lower >>= 1
	}
	return changed_px
}
