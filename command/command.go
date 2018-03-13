package command

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

type shell struct {
	command string
	env     map[string]string
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	newline bool
}

func (s *shell) Run() error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", s.command)
	} else {
		cmd = exec.Command("sh", "-c", s.command)
	}

	cmd.Stdin = s.stdin
	cmd.Stdout = s.stdout
	cmd.Stderr = s.stderr

	if s.newline {
		cmd.Stdout = &newlineWriter{s.stdout}
	}

	for k, v := range s.env {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
	}
	return cmd.Run()
}

type newlineWriter struct {
	w io.Writer
}

func (w *newlineWriter) Write(p []byte) (int, error) {
	n := len(p)
	if p[len(p)-1] != '\n' {
		p = append(p, []byte("\n")...)
	}
	_, err := w.w.Write(p)
	if err != nil {
		return 0, err
	}
	return n, nil
}
