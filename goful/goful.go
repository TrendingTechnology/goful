// Package goful is a CUI file manager.
package goful

import (
	"github.com/anmitsu/goful/filer"
	"github.com/anmitsu/goful/infobar"
	"github.com/anmitsu/goful/menu"
	"github.com/anmitsu/goful/message"
	"github.com/anmitsu/goful/progress"
	"github.com/anmitsu/goful/widget"
	"github.com/nsf/termbox-go"
)

// Goful is a CUI file manager.
type Goful struct {
	*filer.Filer
	shell     func(cmd string) []string
	terminal  func(cmd string) []string
	next      widget.Widget
	event     chan termbox.Event
	interrupt chan int
	callback  chan func()
	task      chan int
	exit      bool
}

// New creates a new goful client based recording a previous state.
func New(path string) *Goful {
	width, height := termbox.Size()
	goful := &Goful{
		Filer:     filer.NewFromState(path, 0, 0, width, height-2),
		shell:     nil,
		terminal:  nil,
		next:      nil,
		event:     make(chan termbox.Event, 20),
		interrupt: make(chan int, 2),
		callback:  make(chan func()),
		task:      make(chan int, 1),
		exit:      false,
	}
	return goful
}

// ConfigShell sets a function that returns a shell name and options.
func (g *Goful) ConfigShell(config func(cmd string) []string) {
	g.shell = config
}

// ConfigTerminal sets a function that returns a terminal name and options.
func (g *Goful) ConfigTerminal(config func(cmd string) []string) {
	g.terminal = config
}

// ConfigFiler sets a keymap function for the filer.
func (g *Goful) ConfigFiler(f func(*Goful) widget.Keymap) {
	g.MergeKeymap(f(g))
}

// Next returns a next widget for drawing and input.
func (g *Goful) Next() widget.Widget { return g.next }

// Disconnect references to a next widget for exiting.
func (g *Goful) Disconnect() { g.next = nil }

// Resize all widgets.
func (g *Goful) Resize(x, y, width, height int) {
	offset := 0
	if !progress.IsFinished() {
		offset = 2
	}
	g.Filer.Resize(x, y, width, height-2-offset)
	if wid := g.Next(); wid != nil {
		wid.Resize(x, y, width, height-2-offset)
	}
	message.Resize(0, height-2, width, 1)
	infobar.Resize(0, height-1, width, 1)
	progress.Resize(0, height-4, width, 1)
}

// Draw all widgets.
func (g *Goful) Draw() {
	g.Filer.Draw()
	if g.Next() != nil {
		g.Next().Draw()
	}
	message.Draw()
	infobar.Draw(g.File())
	progress.Draw()
}

// Input to a current widget.
func (g *Goful) Input(key string) {
	if g.Next() != nil {
		g.Next().Input(key)
	} else {
		g.Filer.Input(key)
	}
}

// Menu runs a menu mode.
func (g *Goful) Menu(name string) {
	m, err := menu.New(name, g)
	if err != nil {
		message.Error(err)
		return
	}
	g.next = m
}

// Run the goful client.
func (g *Goful) Run() {
	message.Info("Welcome to goful")

	go func() {
		for {
			g.event <- termbox.PollEvent()
		}
	}()

	g.Workspace().ReloadAll()
	g.Draw()
	widget.Flush()

	for !g.exit {
		select {
		case ev := <-g.event:
			g.eventHandler(ev)
		case <-g.interrupt:
			<-g.interrupt
		case callback := <-g.callback:
			callback()
		}
		g.Draw()
		widget.Flush()
	}
}

func (g *Goful) syncCallback(callback func()) {
	g.callback <- callback
}

func (g *Goful) eventHandler(ev termbox.Event) {
	if key := widget.EventToString(&ev); key != "" {
		if key == "resize" {
			width, height := ev.Width, ev.Height
			g.Resize(0, 0, width, height)
		} else {
			g.Input(key)
		}
	}
}
