package widget

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/anmitsu/goful/look"
	"github.com/nsf/termbox-go"
)

func TestEventToString(t *testing.T) {
	if err := termbox.Init(); err != nil {
		t.Fatal(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputAlt)

	fmt.Println("Exit by q")
	for {
		ev := termbox.PollEvent()
		key := EventToString(&ev)
		switch key {
		case "q":
			return
		default:
			fmt.Printf("%c %d %s; ", ev.Ch, ev.Key, key)
		}
	}
}

func TestGauge(t *testing.T) {
	if err := termbox.Init(); err != nil {
		t.Fatal(err)
	}
	defer termbox.Close()

	look.Set("default")
	maxval := 200 * 1024 * 1024

	width, _ := termbox.Size()
	gauge := NewProgressGauge(0, 0, width/2, 1)
	gauge.Start(float64(maxval))
	ticker := time.NewTicker(10 * time.Millisecond)

	const n = 50 * 1024 * 1024 / 100 // 50Mb/s
	progress := 0
	for {
		progress += n
		if progress > maxval {
			gauge.Finish()
			break
		}
		gauge.Update(float64(n))
		gauge.Draw()
		Flush()
		<-ticker.C
	}
	termbox.PollEvent()
}

func TestListBox(t *testing.T) {
	if err := termbox.Init(); err != nil {
		t.Fatal(err)
	}
	defer termbox.Close()

	look.Set("default")
	x, y := 0, 0
	width, height := termbox.Size()
	lb := NewListBox(x, y, width/2, height/2, "test")
	lb2 := NewListBox(int(float64(width)*0.7), y, int(float64(width)*0.3), height/2/2, "test2")
	lb3 := NewListBox(x, height/2, width/2, height/2, "test3")
	lb4 := NewListBox(width/2, height/2, width/2, height/2, "test4")

	contents := []string{
		"Hello world.",
		strings.Repeat("Hello world. ", 10),
		"こんにちは世界。",
		strings.Repeat("こんにちは世界。", 10),
		"こんにちは○×□△の世界。",
		strings.Repeat("こんにちは○×□△の世界。", 10),
	}
	lb.AppendString(contents...)
	lb2.AppendString(contents...)
	lb3.AppendString(contents...)
	lb4.AppendString(contents...)

	highlight := [][]string{
		{"Hello world.", "world"},
		{strings.Repeat("Hello world. ", 10), "world"},
		{"こんにちは世界。", "世界"},
		{strings.Repeat("こんにちは世界。 ", 10), "世界"},
		{"こんにちは○×□△の世界。", "□△の"},
		{strings.Repeat("こんにちは○×□△の世界。 ", 10), "□△の"},
	}

	for _, s := range highlight {
		lb.AppendHighlightString(s[0], s[1])
		lb2.AppendHighlightString(s[0], s[1])
		lb3.AppendHighlightString(s[0], s[1])
		lb4.AppendHighlightString(s[0], s[1])
	}

	more := strings.Repeat("ABCあいう○×□△", 10)
	for _, content := range more {
		lb.AppendString(string(content))
		lb2.AppendString(string(content))
		lb3.AppendString(string(content))
		lb4.AppendString(string(content))
	}

	lb.SetCursor(5)
	lb2.SetCursor(9)
	lb3.SetCursor(40)
	lb4.SetCursor(80)

	lb.Draw()
	lb2.Draw()
	lb3.Draw()
	lb4.Draw()
	Flush()
	termbox.PollEvent()
}

func TestInsertBytes(t *testing.T) {
	for _, d := range []struct {
		s      []byte
		data   []byte
		offset int
		result []byte
	}{
		{
			[]byte("Hello world"),
			[]byte("my "),
			6,
			[]byte("Hello my world"),
		},
		{
			[]byte("こんにちは世界"),
			[]byte("私の"),
			15,
			[]byte("こんにちは私の世界"),
		},
		{
			[]byte("こんにちは△□の世界"),
			[]byte("○✕"),
			15,
			[]byte("こんにちは○✕△□の世界"),
		},
	} {
		s := InsertBytes(d.s, d.data, d.offset)
		if !bytes.Equal(s, d.result) {
			t.Errorf("InsertBytes(%q, %q, %q)=%q, want %q", d.s, d.data, d.offset, s, d.result)
		}
	}
}

func TestDeleteBytes(t *testing.T) {
	for _, d := range []struct {
		s      []byte
		offset int
		length int
		result []byte
	}{
		{
			[]byte("Hello my world"),
			6,
			3,
			[]byte("Hello world"),
		},
		{
			[]byte("こんにちは私の世界"),
			15,
			6,
			[]byte("こんにちは世界"),
		},
		{
			[]byte("こんにちは○✕△□の世界"),
			15,
			15,
			[]byte("こんにちは世界"),
		},
	} {
		s := DeleteBytes(d.s, d.offset, d.length)
		if !bytes.Equal(s, d.result) {
			t.Errorf("DeleteBytes(%q, %q, %q)=%q, want %q", d.s, d.offset, d.length, s, d.result)
		}
	}
}
