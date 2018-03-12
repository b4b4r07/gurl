package command

import (
	"fmt"
	"os"
	"os/exec"
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

	Newline bool
}

// Do makes HTTP request
func (r *Request) Do() (err error) {
	stdout := capture(func() {
		err = Run(r.makeCommand(), r.Env)
	})
	if r.Newline {
		stdout = strings.TrimRight(stdout, "\n") + "\n"
	}
	fmt.Fprint(os.Stdout, stdout)
	return
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
