package runlog

import (
	"bufio"
	"context"
	"fmt"
	"io"
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

func (runner Runner) Run(ctx context.Context, workDir string, command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("command is empty")
	}
	if runner.Log != nil {
		runner.Log.PrintlnWithTransaction(runner.Transaction, "Running command:", strings.Join(command, " "))
		runner.Log.PrintlnWithTransaction(runner.Transaction, "Working directory:", workDir)
	}
	if runner.DryRun {
		if runner.Log != nil {
			runner.Log.PrintlnWithTransaction(runner.Transaction, "dry-run: command execution skipped")
		}
		return nil
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), envPairs(runner.Env)...)
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
	readErrs := make(chan error, 2)
	pump := func(prefix string, source io.Reader) {
		defer wg.Done()
		reader := bufio.NewReader(source)
		for {
			line, err := reader.ReadString('\n')
			if line != "" {
				line = strings.TrimRight(line, "\r\n")
				if runner.Log != nil {
					runner.Log.PrintlnWithTransaction(runner.Transaction, prefix+line)
				}
			}
			if err == nil {
				continue
			}
			if err != io.EOF {
				readErrs <- err
			}
			return
		}
	}
	wg.Add(2)
	go pump("", stdout)
	go pump("", stderr)
	wg.Wait()
	close(readErrs)
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %s: %w", strings.Join(command, " "), err)
	}
	for readErr := range readErrs {
		if readErr != nil {
			return fmt.Errorf("read command output: %w", readErr)
		}
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
