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
	Log    *Logger
	DryRun bool
}

func (r Runner) Run(ctx context.Context, workDir string, command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("命令为空")
	}
	if r.Log != nil {
		r.Log.Println("执行命令：", strings.Join(command, " "))
		r.Log.Println("工作目录：", workDir)
	}
	if r.DryRun {
		if r.Log != nil {
			r.Log.Println("dry-run：跳过命令执行")
		}
		return nil
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = workDir
	cmd.Env = os.Environ()
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
				r.Log.Println(prefix + s.Text())
			}
		}
	}
	wg.Add(2)
	go pump("", bufio.NewScanner(stdout))
	go pump("", bufio.NewScanner(stderr))
	wg.Wait()
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("命令执行失败：%s：%w", strings.Join(command, " "), err)
	}
	return nil
}
