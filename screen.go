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

type EventKind uint8

const (
	EventQuit EventKind = iota
	EventClear
	EventResize
	EventSprite
)

type Event struct {
	Kind   EventKind
	Width  int
	Height int
	Sprite Sprite
}

type Display struct {
	screen  t.Screen
	fb      Fb
	fbpos   Vec2
	Echan   chan Event
	Key     chan uint8
	Collide chan uint8
}

func NewDisplay() Display {
	s, err := t.NewScreen()
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

	dp := Display{
		screen: s,

		fb:    NewFb(FB_WIDTH, FB_HEIGHT),
		fbpos: Vec2{BORDER_PAD, BORDER_PAD},

		Echan:   make(chan Event, 4),
		Key:     make(chan uint8, 1),
		Collide: make(chan uint8),
	}
	dp.drawBorder()

	return dp
}

// Flash a FrameBuffer to the underlying screen
func (dp *Display) renderFb() {
	s := dp.screen
	fb := dp.fb
	sw, sh := s.Size()
	style := t.StyleDefault.Foreground(t.ColorBlue).Background(t.ColorReset)

	idx := func(x, y int) int {
		return y*fb.w + x
	}

	for y := 0; y < fb.h; y++ {
		dy := dp.fbpos.y + y
		if dy == 0 {
			dy += BORDER_WIDTH
		}
		if dy == sh-1 {
			dy -= BORDER_WIDTH
		}

		for x := 0; x < fb.w; x++ {
			r := CELL_FILLED
			if fb.buf[idx(x, y)] == 0 {
				r = CELL_EMPTY
			}

			// respect the screen border
			dx := dp.fbpos.x + x
			if dx == 0 {
				dx += BORDER_WIDTH
			}
			if dx == sw-1 {
				dx -= BORDER_WIDTH
			}

			s.SetContent(dx, dy, r, nil, style)
		}
	}
}

// This is analogous to the Fetch-Execute-Cycle of the `Cpu`.
// It starts the window with content, and is only called if `isScreenInit` = true
//
// This is the driver/center of this abstraction, and must be a Goroutine to not block
func (dp *Display) StartRenderLoop() {
	sw, sh := dp.screen.Size()
	log.Printf("Screen: %dx%d\n", sw, sh)

	// spinner shit
	// spinnerStyles := []t.Style{
	// 	t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorRed),
	// 	t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorGreen),
	// 	t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorBlue),
	// 	t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorWhite),
	// }
	// lstyles := len(spinnerStyles)
	// spinners := `←↖↑↗→↘↓↙`
	// len_spinners := len(spinners)

	go dp.pollEvent()

renderloop:
	for {
		select {
		case ev := <-dp.Echan:
			switch ev.Kind {
			case EventQuit:
				log.Println("LoopQuit::")
				close(dp.Key)
				close(dp.Echan)
				break renderloop
			case EventClear:
				clear(dp.fb.buf)
				dp.renderFb()
				dp.screen.Show()
			case EventResize:
				dp.handleResize()
				sw, sh := ev.Width, ev.Height
				log.Printf("Resized to: %dx%d\n", sw, sh)
			case EventSprite:
				sprite := ev.Sprite
				dp.Collide <- dp.fb.DrawSpriteAt(sprite.data, Vec2{int(sprite.x), int(sprite.y)})
				dp.renderFb()
				dp.screen.Show()
			default:
			}

		default:
		}
		time.Sleep(8 * time.Millisecond)
	}

	log.Println("Done")
}

// This async'ly polls events from the window and other sources. It also emits
// events to certain sources
//
// Thus, it must be it's own Goroutine
func (dp *Display) pollEvent() {
	s := dp.screen

	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *t.EventResize:
			log.Printf("EventResize::")
			sw, sh := ev.Size()
			dp.Echan <- Event{Kind: EventResize, Width: sw, Height: sh}

		case *t.EventKey:
			log.Println("EventKey::")
			if ev.Key() == t.KeyEscape || ev.Key() == t.KeyCtrlC {
				log.Printf("stopping\n")
				dp.Echan <- Event{Kind: EventQuit}
				log.Printf("STOP\n")
				return
			} else if ev.Key() == t.KeyRune {
				logmsg("rune\n")
				key := uint8(ev.Rune())

				if '0' <= key && key <= '9' {
					logmsg("key 0-9: %c/%d\n", key, key-'0')
					key = key - '0'
				}
				if 'a' <= key && key <= 'f' {
					logmsg("key a-f: %c/%d\n", key, 0xa+key-'a')
					key = 0xa + key - 'a'
				}

				select {
				case dp.Key <- key:
				default:
				}
			}
		default:
		}
	}
}

func (dp *Display) handleResize() {
	log.Println("to resize")
	dp.screen.Clear()
	dp.drawBorder()
	dp.renderFb()
	dp.screen.Sync()
}

func (dp *Display) HandleQuit() {
	log.Println("Quit::")
	maybePanic := recover()
	dp.screen.Clear()
	dp.screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}

func (dp *Display) drawBorder() {
	s := dp.screen
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
