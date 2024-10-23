package main

import (
	t "github.com/gdamore/tcell/v2"
	"log"
)

type Vec2 struct {
	x, y int
}

const (
	FB_WIDTH  = 64
	FB_HEIGHT = 32

	CELL_EMPTY          = t.RuneBullet
	CELL_FILLED         = t.RuneBlock
	BORDER_HORI         = t.RuneHLine
	BORDER_VERT         = t.RuneVLine
	BORDER_LEFT_TOP     = t.RuneULCorner
	BORDER_RIGHT_TOP    = t.RuneURCorner
	BORDER_LEFT_BOTTOM  = t.RuneLLCorner
	BORDER_RIGHT_BOTTOM = t.RuneLRCorner
)

var (
	Digits = []uint8{
		0xf0, 0x90, 0x90, 0x90, 0xf0, // 0x0
		0x20, 0x60, 0x20, 0x20, 0x70, // 0x1
		0xf0, 0x10, 0xf0, 0x80, 0xf0, // 0x2
		0xf0, 0x10, 0xf0, 0x10, 0xf0, // 0x3
		0x90, 0x90, 0xf0, 0x10, 0x10, // 0x4
		0xf0, 0x80, 0xf0, 0x10, 0xf0, // 0x5
		0xf0, 0x80, 0xf0, 0x90, 0xf0, // 0x6
		0xf0, 0x10, 0x20, 0x40, 0x40, // 0x7
		0xf0, 0x90, 0xf0, 0x90, 0xf0, // 0x8
		0xf0, 0x90, 0xf0, 0x10, 0xf0, // 0x9
		0xf0, 0x90, 0xf0, 0x90, 0x90, // 0xa
		0xe0, 0x90, 0xe0, 0x90, 0xe0, // 0xb
		0xf0, 0x80, 0x80, 0x80, 0xf0, // 0xc
		0xe0, 0x90, 0x90, 0x90, 0xe0, // 0xd
		0xf0, 0x80, 0xf0, 0x80, 0xf0, // 0xe
		0xf0, 0x80, 0xf0, 0x80, 0x80, // 0xf
	}
)

type Fb struct {
	buf []uint8
	w   int
	h   int
}

func NewFb(w, h int) Fb {
	return Fb{
		buf: make([]uint8, w*h),
		w:   w,
		h:   h,
	}
}

func (buf *Fb) drawSpriteAt(sprite []byte, ori Vec2) bool {
	x := ori.x
	y := ori.y
	collision := false
	for _, b := range sprite {
		// render the bytes starting from the first one
		j := 0
		mask := uint8(1 << 7)
		for mask != 0 {
			i := (x % buf.w) + (y%buf.h)*buf.w
			old_b := buf.buf[i]
			new_b := (b & mask) >> (7 - j)
			collision = (old_b ^ new_b) == 1

			buf.buf[i] = old_b ^ new_b

			j++
			x++
			mask >>= 1
		}
		x = ori.x
		y++
	}

	return collision
}

/* Draws the given digit (a number not a byte) into the specified buffer */
func (buf *Fb) drawDigit(d uint8, ori Vec2) {
	if (0x0 <= d && d <= 0x9) || (0xa <= d && d <= 0xf) {
		buf.drawSpriteAt(Digits[d*5:][:5], ori)
		return
	}
	log.Fatalf("error: '%d' is not a digit\n", d)
}

func (buf *Fb) drawDigits(ori Vec2) {
	x := ori.x
	y := ori.y
	i := 0
	var d uint8
	for d = 0; d < uint8(0x10); d++ {
		if i == 7 {
			i = 0
			x = ori.x
			y += 8
		}
		buf.drawDigit(d, Vec2{x, y})
		i++
		x += 8
	}
}

func (buf *Fb) renderToScreen(s t.Screen, ori Vec2) {
	sw, sh := s.Size()
	style := t.StyleDefault.Foreground(t.ColorSnow).Background(t.ColorReset)

	for y := 0; y < buf.h; y++ {
		for x := 0; x < buf.w; x++ {
			r := CELL_FILLED
			if buf.buf[y*buf.w+x] == 0 {
				r = CELL_EMPTY
			}

			// respect the screen border
			dx := ori.x + x
			if dx == 0 {
				dx += BORDER_WIDTH
			}
			if dx == sw-1 {
				dx -= BORDER_WIDTH
			}

			dy := ori.y + y
			if dy == 0 {
				dy += BORDER_WIDTH
			}
			if dy == sh-1 {
				dy -= BORDER_WIDTH
			}

			s.SetContent(dx, dy, r, nil, style)
		}
	}
}
