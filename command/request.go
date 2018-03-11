package command

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	// DefaultRequestCommand is a default command
	DefaultRequestCommand = "curl"
)

// Request represents
type Request struct {
	Command   string
	Args      []string
	Headers   []string
	URL       string
	Env       map[string]string
	Processes []string
}

// Do requests
func (r *Request) Do() error {
	command := r.makeCommand()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	for k, v := range r.Env {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
	}
	return cmd.Run()
}

func (r *Request) makeCommand() string {
	command := r.Command
	if command == "" {
		command = DefaultRequestCommand
	}
	if _, err := exec.LookPath(command); err != nil {
		panic(err)
	}
	for _, header := range r.Headers {
		command += " -H " + header
	}
	for _, arg := range r.Args {
		command += " " + arg
	}
	command += " " + r.URL
	if len(r.Processes) > 0 {
		command += " | " + strings.Join(r.Processes, " | ")
	}
	return command
}

// AddHeader adds the header for requesting
func (r *Request) AddHeader(header string) {
	r.Headers = append(r.Headers, header)
}
