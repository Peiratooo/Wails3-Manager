package runlog

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunnerCapturesLongOutputLine(t *testing.T) {
	if os.Getenv("RUNLOG_HELPER_LONG_LINE") == "1" {
		fmt.Println(strings.Repeat("x", 128*1024))
		return
	}

	log := New(io.Discard)
	err := (Runner{
		Log: log,
		Env: map[string]string{"RUNLOG_HELPER_LONG_LINE": "1"},
	}).Run(context.Background(), "", []string{os.Args[0], "-test.run=TestRunnerCapturesLongOutputLine"})
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Join(log.Lines(), "\n")
	if !strings.Contains(lines, strings.Repeat("x", 128*1024)) {
		t.Fatalf("long output line was not captured")
	}
}
