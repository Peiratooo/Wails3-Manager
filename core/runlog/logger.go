// Package runlog contains the small logging buffer and command runner used by
// Wails services. It deliberately avoids business concepts so services can share
// it without importing each other.
package runlog

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	mu      sync.Mutex
	writers []io.Writer
	lines   []string
	OnLine  func(string)
}

func New(writers ...io.Writer) *Logger {
	if len(writers) == 0 {
		writers = []io.Writer{os.Stdout}
	}
	return &Logger{writers: writers, lines: []string{}}
}

func (l *Logger) Lines() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, len(l.lines))
	copy(out, l.lines)
	return out
}

func (l *Logger) Since(cursor int) (int, []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(l.lines) {
		cursor = len(l.lines)
	}
	out := make([]string, len(l.lines[cursor:]))
	copy(out, l.lines[cursor:])
	return len(l.lines), out
}

func (l *Logger) Cursor() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.lines)
}

func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = []string{}
}

func (l *Logger) Println(args ...any)               { l.write(fmt.Sprintln(args...)) }
func (l *Logger) Printf(format string, args ...any) { l.write(fmt.Sprintf(format, args...)) }
func (l *Logger) Section(title string)              { l.write("\n== " + title + " ==\n") }

func (l *Logger) write(s string) {
	s = strings.TrimRight(s, "\n")
	line := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), s)
	l.mu.Lock()
	l.lines = append(l.lines, line)
	for _, w := range l.writers {
		_, _ = fmt.Fprintln(w, line)
	}
	on := l.OnLine
	l.mu.Unlock()
	if on != nil {
		on(line)
	}
}
