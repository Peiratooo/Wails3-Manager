// Package runlog contains the small logging buffer and command runner used by
// Wails services. It deliberately avoids business concepts so services can share
// it without importing each other.
package runlog

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	mu         sync.Mutex
	writers    []io.Writer
	lines      []string
	recordLogs bool
	OnLine     func(Line)
}

type Transaction struct {
	ID    string
	Type  string
	Title string
}

type Line struct {
	Text             string
	TransactionID    string
	TransactionType  string
	TransactionTitle string
}

func New(writers ...io.Writer) *Logger {
	if len(writers) == 0 {
		writers = []io.Writer{os.Stdout}
	}
	return &Logger{writers: writers, lines: []string{}, recordLogs: true}
}

func (l *Logger) Lines() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.lines)
}

func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = []string{}
}

func (l *Logger) SetRecordLogs(recordLogs bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.recordLogs = recordLogs
}

func (l *Logger) PrintlnWithTransaction(tx Transaction, args ...any) {
	l.write(tx, fmt.Sprintln(args...))
}

func (l *Logger) write(tx Transaction, s string) {
	s = strings.TrimRight(s, "\n")
	line := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), s)
	l.mu.Lock()
	recordLogs := l.recordLogs
	if recordLogs {
		l.lines = append(l.lines, line)
	}
	for _, w := range l.writers {
		_, _ = fmt.Fprintln(w, line)
	}
	on := l.OnLine
	l.mu.Unlock()
	if recordLogs && on != nil {
		on(Line{
			Text:             line,
			TransactionID:    tx.ID,
			TransactionType:  tx.Type,
			TransactionTitle: tx.Title,
		})
	}
}
