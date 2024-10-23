package main

import (
	"fmt"
	t "github.com/gdamore/tcell/v2"
	"log"
)

func drawText(s t.Screen, start, end Vec2, style t.Style, text string) {
	row := start.y
	col := start.x
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= end.x {
			row++
			col = start.x
		}
		if row > end.y {
			break
		}
	}
}

func drawBorder(s t.Screen) {
	w, h := s.Size()
	style := t.StyleDefault.Foreground(t.ColorDarkCyan).Background(t.ColorReset)

	for y := 0; y < h; y++ {
		s.SetContent(0, y, BORDER_VERT, nil, style)
		s.SetContent(w-1, y, BORDER_VERT, nil, style)
	}

	for x := 0; x < w; x++ {
		s.SetContent(x, 0, BORDER_HORI, nil, style)
		s.SetContent(x, h-1, BORDER_HORI, nil, style)
	}

	s.SetContent(0, 0, BORDER_LEFT_TOP, nil, style)
	s.SetContent(w-1, 0, BORDER_RIGHT_TOP, nil, style)
	s.SetContent(0, h-1, BORDER_LEFT_BOTTOM, nil, style)
	s.SetContent(w-1, h-1, BORDER_RIGHT_BOTTOM, nil, style)
}

func main() {
	defStyle := t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorReset)

	// Initialize screen
	s, err := t.NewScreen()
	{
		if err != nil {
			log.Fatalf("%+v", err)
		}
		if err := s.Init(); err != nil {
			log.Fatalf("%+v", err)
		}

		if sw, sh := s.Size(); sw < FB_WIDTH || sh < FB_HEIGHT {
			s.Clear()
			s.Fini()
			if sw < FB_WIDTH {
				log.Fatalf("screen width %d < %d for rendering\n", sw, FB_WIDTH)
			} else if sh < FB_HEIGHT {
				log.Fatalf("screen height %d < %d for rendering\n", sh, FB_HEIGHT)
			}
		}
	}

	buf := NewFb(FB_WIDTH, FB_HEIGHT)
	buf.drawDigits(Vec2{0, 0})

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		s.Clear()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	// Initial draw
	{
		s.SetStyle(defStyle)
		s.EnableMouse()
		s.EnablePaste()
		s.Clear()
		buf.renderToScreen(s, Vec2{1, 1})
		drawBorder(s)
	}

	for {
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *t.EventResize:
			s.Clear()
			drawBorder(s)
			buf.renderToScreen(s, Vec2{1, 1})
			s.Sync()
		case *t.EventKey:
			if ev.Key() == t.KeyEscape || ev.Key() == t.KeyCtrlC {
				return
			} else if ev.Key() == t.KeyCtrlL {
				s.Sync()
			}
		case *t.EventMouse:
			if ev.Buttons() == t.ButtonPrimary || ev.Buttons() == t.ButtonSecondary {
				mx, my := ev.Position()
				coord := fmt.Sprintf("(%d, %d)", mx, my)
			inner: // move the coord so that pointer is pointing ', '
				for i, c := range []rune(coord) {
					if c == ' ' {
						mx -= i
						break inner
					}
				}

				s.Clear()
				drawBorder(s)
				drawText(s, Vec2{mx, my}, Vec2{mx + len(coord), my}, defStyle, coord)
				buf.renderToScreen(s, Vec2{1, 1})
			}
		}
	}
}
