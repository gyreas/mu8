package main

import (
	t "github.com/gdamore/tcell/v2"
	"log"
	"sync"
	"time"
)

const (
	BORDER_WIDTH = 1
	BORDER_PAD   = 2 * BORDER_WIDTH
)

var (
	Ping   = struct{}{}
	Resize = struct{}{}
)

type Display struct {
	screen t.Screen
	smu    sync.Mutex
	fb     Fb
	fbpos  Vec2
	sEvent chan t.Event
	resize chan struct{}
	ping   chan struct{}
	quit   chan struct{}
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
		smu:    sync.Mutex{},

		fb:    NewFb(FB_WIDTH, FB_HEIGHT),
		fbpos: Vec2{BORDER_PAD, BORDER_PAD},

		sEvent: make(chan t.Event),
		ping:   make(chan struct{}),
		resize: make(chan struct{}, 1),
		quit:   make(chan struct{}, 1),
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
func (dp *Display) startRenderLoop() {
	sw, sh := dp.screen.Size()
	log.Printf("Screen: %dx%d\n", sw, sh)

	go dp.spinner()
	go dp.pollEvent()

	for {

		dp.screen.Show()
		select {
		case <-dp.quit:
			return
		case <-dp.ping:
			dp.smu.Lock()
			dp.renderFb()
			dp.smu.Unlock()
		case <-dp.resize:
			dp.smu.Lock()
			dp.handleResize()
			sw, sh := dp.screen.Size()
			log.Printf("Resized to: %dx%d\n", sw, sh)
			dp.smu.Unlock()
		default:
			if dp.handleEvent() {
				return
			}
		}
	}
}

// This async'ly polls events from the window and other sources. It also emits
// events to certain sources
//
// Thus, it must be it's own Goroutine
func (dp *Display) pollEvent() {
	for {
		select {
		case <-dp.quit:
			return
		default:
			dp.sEvent <- dp.screen.PollEvent()
		}
	}
}

func (dp *Display) handleResize() {
	dp.screen.Clear()
	// dp.screen.Sync()
	dp.drawBorder()
	dp.renderFb()
	dp.screen.Sync()
}

func (dp *Display) handleEvent() bool {
	s := dp.screen

	ev := <-dp.sEvent
	switch ev := ev.(type) {
	case *t.EventResize:
		dp.resize <- Resize
	case *t.EventKey:
		if ev.Key() == t.KeyEscape || ev.Key() == t.KeyCtrlC {
			log.Printf("stopping\n")
			dp.quit <- Ping
			log.Printf("STOP\n")
			return true
		} else if ev.Key() == t.KeyCtrlL {
			s.Sync()
		}
	}
	return false
}

func (dp *Display) handleQuit() {
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

// This merely displays a unadorned spinner
func (dp *Display) spinner() {
	spinnerStyles := []t.Style{
		t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorRed),
		t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorGreen),
		t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorBlue),
		t.StyleDefault.Background(t.ColorReset).Foreground(t.ColorWhite),
	}

	spinners := `←↖↑↗→↘↓↙`

	time.Sleep(44 * time.Millisecond)
	for {
		for i, c := range spinners {
			select {
			case <-dp.quit:
				return
			default:
				dp.smu.Lock()
				sw, _ := dp.screen.Size()
				dp.screen.SetContent(sw-4, 2, c, nil, spinnerStyles[i%len(spinnerStyles)])
				dp.screen.Show()
				dp.smu.Unlock()
			}

			time.Sleep(644 * time.Millisecond)
		}
	}
}
