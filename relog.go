package relog

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

type Logger struct {
	w       io.Writer
	mu      *sync.Mutex
	entries []entry
}

func NewLogger(w io.Writer) *Logger {
	return &Logger{
		w:  w,
		mu: new(sync.Mutex),
	}
}

// Log a message. If `id` is seen before then update that line, otherwise write at the end.
func (l *Logger) Log(id, msg string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !strings.HasSuffix(msg, newline) {
		msg += newline
	}
	want := entry{
		id:  id,
		msg: msg,
	}

	idx, linesFromEnd, got := l.lookup(id)
	if idx == notFound {
		return l.append(want)
	}

	return l.update(idx, linesFromEnd, got, want)
}

type entry struct {
	id   string
	msg  string
	done bool
}

func (e entry) countLines() int {
	return strings.Count(e.msg, newline)
}

func (e entry) multiline() bool {
	return e.countLines() > 1
}

const (
	maxNEntries = 1000

	// https://www.xfree86.org/current/ctlseqs.html
	esc               = "\033"
	newline           = "\n"
	eraseLine         = esc + "[2K"
	eraseDisplayBelow = esc + "[J"

	notFound = -1
)

func (l *Logger) lookup(id string) (idx int, linesFromEnd int, got entry) {
	n := 0
	for i := len(l.entries) - 1; i >= 0; i-- {
		e := l.entries[i]
		n += e.countLines()
		if e.id == id {
			return i, n, e
		}
	}
	return notFound, 0, entry{}
}

func (l *Logger) update(i, linesFromEnd int, got, want entry) error {
	l.entries[i].msg = want.msg

	moveUp := fmt.Sprintf("%s[%dA", esc, linesFromEnd)

	if !got.multiline() && !want.multiline() {
		// Replace the line in-place
		var moveDown string
		if linesFromEnd > 1 {
			moveDown = fmt.Sprintf("%s[%dB", esc, linesFromEnd-1)
		}
		if _, err := l.w.Write([]byte(moveUp + eraseLine + want.msg + moveDown)); err != nil {
			return err
		}
		return nil
	}

	// Rewrite all messages from index to end
	if _, err := l.w.Write([]byte(moveUp + eraseDisplayBelow)); err != nil {
		return err
	}
	for _, e := range l.entries[i:] {
		if _, err := l.w.Write([]byte(e.msg)); err != nil {
			return err
		}
	}

	return nil
}

func (l *Logger) append(e entry) error {
	if _, err := l.w.Write([]byte(e.msg)); err != nil {
		return err
	}
	l.entries = append(l.entries, e)
	if len(l.entries) > maxNEntries {
		l.entries = l.entries[1:]
	}
	return nil
}
