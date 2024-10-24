package main

import (
	t "github.com/gdamore/tcell/v2"
	"log"
	"time"
)

const (
	BORDER_WIDTH = 1
	BORDER_PAD   = 2 * BORDER_WIDTH
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

type Display struct {
	fb     Fb
	fbpos  Vec2
	screen t.Screen
	sEvent chan t.Event
	spin   chan rune
	stop   chan struct{}
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

	D := Display{
		screen: s,
		fb:     NewFb(FB_WIDTH, FB_HEIGHT),
		fbpos:  Vec2{1, 1},
		spin:   make(chan rune),
		sEvent: make(chan t.Event),
		stop:   make(chan struct{}),
	}

	// Initial draw
	{
		defStyle := t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorReset)
		D.fb.drawDigit(4, D.fbpos)

		s.SetStyle(defStyle)
		s.EnableMouse()
		s.EnablePaste()
		s.Clear()

		D.fb.renderToScreen(s, D.fbpos)
		drawBorder(s)
	}

	go spinner(&D)
	go func(D *Display) {
		for {
			select {
			case <-D.stop:
				close(D.stop)
				close(D.sEvent)
				return
			default:
				D.sEvent <- s.PollEvent()
			}
		}
	}(&D)

mainLoop:
	for {
		D.screen.Show()

		if processEvent(&D) {
			break mainLoop
		}
	}
}

func spinner(D *Display) {
	ticker := time.NewTicker(1 * time.Millisecond)
	spinnerStyles := []t.Style{
		t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorRed),
		t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorGreen),
		t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorBlue),
		t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorWhite),
	}

	for {
		for i, c := range `-\|/` {
			select {
			case D.spin <- c:
			case <-D.stop:
				close(D.spin)
				return
			default:
				D.screen.SetContent(2, 1, c, nil, spinnerStyles[i])
				D.screen.Show()
			}

			ticker.Reset(111 * time.Millisecond)
			for {
				<-ticker.C
				ticker.Stop()
				break
			}
		}
	}
}

func processEvent(D *Display) bool {
	s := D.screen
	buf := D.fb
	fbpos := D.fbpos

	ev := <-D.sEvent
	switch ev := ev.(type) {
	case *t.EventResize:
		s.Clear()
		drawBorder(s)
		buf.renderToScreen(s, fbpos)
		s.Sync()
	case *t.EventKey:
		if ev.Key() == t.KeyEscape || ev.Key() == t.KeyCtrlC {
			log.Printf("STOP\n")
			D.stop <- struct{}{}
			return true
		} else if ev.Key() == t.KeyCtrlL {
			s.Sync()
		}
	}
	return false
}
