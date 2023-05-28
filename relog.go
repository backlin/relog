package relog

import (
	"fmt"
	"io"
)

type Logger struct {
	w         io.Writer
	entries   []entry
	cursorPos int
}

func NewLogger(w io.Writer) *Logger {
	return &Logger{w: w}
}

func (l *Logger) Close() error {
	return l.moveTo(len(l.entries))
}

type entry struct {
	id   string
	line string
}

// Log a message. If `id` is seen before then update that line, otherwise write at the end.
func (l *Logger) Log(id, msg string) error {
	for i, e := range l.entries {
		if e.id == id {
			return l.rewrite(i, msg)
		}
	}
	return l.append(id, msg)
}

const (
	esc     = "\033"
	newline = "\n"
)

func (l *Logger) moveTo(entry int) error {
	steps := entry - l.cursorPos
	var lineChange string
	if steps < 0 {
		lineChange = fmt.Sprintf("%s[%dA", esc, -steps)
	} else if steps > 0 {
		lineChange = fmt.Sprintf("%s[%dB", esc, steps)
	} else {
		return nil
	}
	if _, err := l.w.Write([]byte(lineChange)); err != nil {
		return err
	}
	l.cursorPos = entry
	return nil
}

func (l *Logger) rewrite(i int, line string) error {
	if err := l.moveTo(i); err != nil {
		return err
	}
	_, err := l.w.Write([]byte(fmt.Sprintf("\r%s[2K%s%s", esc, line, newline)))
	if err != nil {
		return err
	}
	l.entries[i].line = line
	l.cursorPos++ // because of the newline
	return nil
}

func (l *Logger) append(id, line string) error {
	l.entries = append(l.entries, entry{id, ""})
	return l.rewrite(len(l.entries)-1, line)
}
