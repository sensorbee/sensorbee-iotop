package iotop

import (
	"fmt"
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
)

// editBox is a command input box to operate in running.
// ref: nsf/termbox-go/_demos/editbox.go
type editBox struct {
	text          []byte
	lineVOffset   int
	cursorBOffset int // cursor offset in bytes
	cursorVOffset int // visual cursor offset in termbox cells
	cursorCOffset int // cursor offset in unicode code points
}

const (
	preferredHorizontalThreshold = 5
	tabstopLength                = 8
)

func (eb *editBox) redrawAll(prefix string) {
	//termbox.Clear(iotopTerminalColor, iotopTerminalColor)

	tbprint(0, 0, iotopTerminalColor, iotopTerminalColor, prefix)
	prefixLen := len(prefix)
	w, _ := termbox.Size()

	eb.draw(prefixLen, 0, w-prefixLen, 1)
	termbox.SetCursor(prefixLen+eb.cursorX(), 0)

	termbox.Flush()
}

func (eb *editBox) reset() {
	eb.text = []byte{}
	eb.lineVOffset = 0
	eb.cursorBOffset = 0
	eb.cursorVOffset = 0
	eb.cursorCOffset = 0
	termbox.HideCursor()
}

func (eb *editBox) start(prefix string) (string, error) {
	eb.redrawAll(prefix)

	running := true
	for running {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				eb.reset()
				running = false
			case termbox.KeyEnter:
				running = false
			case termbox.KeyArrowLeft, termbox.KeyCtrlB:
				eb.moveCursorOneRuneBackward()
			case termbox.KeyArrowRight, termbox.KeyCtrlF:
				eb.moveCursorOneRuneForward()
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				eb.deleteRuneBackward()
			case termbox.KeyDelete, termbox.KeyCtrlD:
				eb.deleteRuneForward()
			case termbox.KeyTab:
				eb.insertRune('\t')
			case termbox.KeySpace:
				eb.insertRune(' ')
			case termbox.KeyCtrlK:
				eb.deleteTheRestOfTheLine()
			case termbox.KeyHome, termbox.KeyCtrlA:
				eb.moveCursorToBeginningOfTheLine()
			case termbox.KeyEnd, termbox.KeyCtrlE:
				eb.moveCursorToEndOfTheLine()
			default:
				if ev.Ch != 0 {
					eb.insertRune(ev.Ch)
				}
			}
		case termbox.EventError:
			return "", fmt.Errorf("Cannot get key events, %v", ev.Err)
		}
		eb.redrawAll(prefix)
	}
	return string(eb.text), nil
}

func (eb *editBox) draw(x, y, w, h int) {
	eb.adjustVOffset(w)

	fill(x, y, w, h, termbox.Cell{Ch: ' '})

	t := eb.text
	lx := 0
	tabstop := 0
	for {
		rx := lx - eb.lineVOffset
		if len(t) == 0 {
			break
		}

		if lx == tabstop {
			tabstop += tabstopLength
		}

		if rx >= w {
			termbox.SetCell(w+w-1, y, '→', iotopTerminalColor,
				iotopTerminalColor)
			break
		}

		r, size := utf8.DecodeRune(t)
		if r == '\t' {
			for ; lx < tabstop; lx++ {
				rx = lx - eb.lineVOffset
				if rx >= w {
					goto next
				}

				if rx >= 0 {
					termbox.SetCell(x+rx, y, ' ', iotopTerminalColor,
						iotopTerminalColor)
				}
			}
		} else {
			if rx >= 0 {
				termbox.SetCell(x+rx, y, r, iotopTerminalColor,
					iotopTerminalColor)
			}
			lx += runewidth.RuneWidth(r)
		}
	next:
		t = t[size:]
	}

	if eb.lineVOffset != 0 {
		termbox.SetCell(x, y, '←', iotopTerminalColor, iotopTerminalColor)
	}
}

// AdjustVOffset adjusts line visual offset to a proper value depending on width
func (eb *editBox) adjustVOffset(width int) {
	ht := preferredHorizontalThreshold
	maxHThreshold := (width - 1) / 2
	if ht > maxHThreshold {
		ht = maxHThreshold
	}

	threshold := width - 1
	if eb.lineVOffset != 0 {
		threshold = width - ht
	}
	if eb.cursorVOffset-eb.lineVOffset >= threshold {
		eb.lineVOffset = eb.cursorVOffset + (ht - width + 1)
	}

	if eb.lineVOffset != 0 && eb.cursorVOffset-eb.lineVOffset < ht {
		eb.lineVOffset = eb.cursorVOffset - ht
		if eb.lineVOffset < 0 {
			eb.lineVOffset = 0
		}
	}
}

func (eb *editBox) moveCursorTo(boffset int) {
	eb.cursorBOffset = boffset
	eb.cursorVOffset, eb.cursorCOffset = voffsetCOffset(eb.text, boffset)
}

func (eb *editBox) runeUnderCursor() (rune, int) {
	return utf8.DecodeRune(eb.text[eb.cursorBOffset:])
}

func (eb *editBox) runeBeforeCursor() (rune, int) {
	return utf8.DecodeLastRune(eb.text[:eb.cursorBOffset])
}

func (eb *editBox) moveCursorOneRuneBackward() {
	if eb.cursorBOffset == 0 {
		return
	}
	_, size := eb.runeBeforeCursor()
	eb.moveCursorTo(eb.cursorBOffset - size)
}

func (eb *editBox) moveCursorOneRuneForward() {
	if eb.cursorBOffset == len(eb.text) {
		return
	}
	_, size := eb.runeUnderCursor()
	eb.moveCursorTo(eb.cursorBOffset + size)
}

func (eb *editBox) moveCursorToBeginningOfTheLine() {
	eb.moveCursorTo(0)
}

func (eb *editBox) moveCursorToEndOfTheLine() {
	eb.moveCursorTo(len(eb.text))
}

func (eb *editBox) deleteRuneBackward() {
	if eb.cursorBOffset == 0 {
		return
	}

	eb.moveCursorOneRuneBackward()
	_, size := eb.runeUnderCursor()
	eb.text = byteSliceRemove(eb.text, eb.cursorBOffset, eb.cursorBOffset+size)
}

func (eb *editBox) deleteRuneForward() {
	if eb.cursorBOffset == len(eb.text) {
		return
	}
	_, size := eb.runeUnderCursor()
	eb.text = byteSliceRemove(eb.text, eb.cursorBOffset, eb.cursorBOffset+size)
}

func (eb *editBox) deleteTheRestOfTheLine() {
	eb.text = eb.text[:eb.cursorBOffset]
}

func (eb *editBox) insertRune(r rune) {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	eb.text = byteSliceInsert(eb.text, eb.cursorBOffset, buf[:n])
	eb.moveCursorOneRuneForward()
}

// cursorX points X axis
// Please, keep in mind that cursor depends on the value of line_voffset, which
// is being set on Draw() call, so.. call this method after Draw() one.
func (eb *editBox) cursorX() int {
	return eb.cursorVOffset - eb.lineVOffset
}

func fill(x, y, w, h int, cell termbox.Cell) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			termbox.SetCell(x+lx, y+ly, cell.Ch, cell.Fg, cell.Bg)
		}
	}
}

func runeAdvanceLen(r rune, pos int) int {
	if r == '\t' {
		return tabstopLength - pos%tabstopLength
	}
	return runewidth.RuneWidth(r)
}

func voffsetCOffset(text []byte, boffset int) (voffset, coffset int) {
	text = text[:boffset]
	for len(text) > 0 {
		r, size := utf8.DecodeRune(text)
		text = text[size:]
		coffset++
		voffset += runeAdvanceLen(r, voffset)
	}
	return
}

func byteSliceGrow(s []byte, desiredCap int) []byte {
	if cap(s) < desiredCap {
		ns := make([]byte, len(s), desiredCap)
		copy(ns, s)
		return ns
	}
	return s
}

func byteSliceRemove(text []byte, from, to int) []byte {
	size := to - from
	copy(text[from:], text[to:])
	text = text[:len(text)-size]
	return text
}

func byteSliceInsert(text []byte, offset int, what []byte) []byte {
	n := len(text) + len(what)
	text = byteSliceGrow(text, n)
	text = text[:n]
	copy(text[offset+len(what):], text[offset:])
	copy(text[offset:], what)
	return text
}
