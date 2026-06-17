package runlog

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type Runner struct {
	Log         *Logger
	DryRun      bool
	Env         map[string]string
	Transaction Transaction
}

func (r Runner) Run(ctx context.Context, workDir string, command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("command is empty")
	}
	if r.Log != nil {
		r.Log.PrintlnWithTransaction(r.Transaction, "Running command:", strings.Join(command, " "))
		r.Log.PrintlnWithTransaction(r.Transaction, "Working directory:", workDir)
	}
	if r.DryRun {
		if r.Log != nil {
			r.Log.PrintlnWithTransaction(r.Transaction, "dry-run: command execution skipped")
		}
		return nil
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), envPairs(r.Env)...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	var wg sync.WaitGroup
	pump := func(prefix string, s *bufio.Scanner) {
		defer wg.Done()
		for s.Scan() {
			if r.Log != nil {
				r.Log.PrintlnWithTransaction(r.Transaction, prefix+s.Text())
			}
		}
	}
	wg.Add(2)
	go pump("", bufio.NewScanner(stdout))
	go pump("", bufio.NewScanner(stderr))
	wg.Wait()
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %s: %w", strings.Join(command, " "), err)
	}
	return nil
}

func envPairs(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	pairs := make([]string, 0, len(env))
	for k, v := range env {
		if strings.TrimSpace(k) == "" {
			continue
		}
		pairs = append(pairs, k+"="+v)
	}
	return pairs
}
