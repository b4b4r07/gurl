package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/b4b4r07/req/command"
	"github.com/b4b4r07/req/config"
	"github.com/b4b4r07/req/iap"

	homedir "github.com/mitchellh/go-homedir"
)

const helpText string = `Usage: req [OPTIONS] URL

Extended options:
  --list, --list-urls    List service URLs
  --edit, --edit-config  Edit config file
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: too few arguments\n")
		return 1
	}

	var cfg config.Config
	if err := cfg.LoadFile(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Fprint(os.Stderr, helpText)
		return 1
	case "--list-urls", "--list":
		fmt.Println(strings.Join(cfg.GetURLs(), "\n"))
		return 0
	case "--edit-config", "--edit":
		if err := cfg.Edit(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())
			return 1
		}
		return 0
	default:
		// Ignore other arguments
	}

	// The last argument is regarded as an URL
	url := args[len(args)-1]

	service := cfg.GetService(url)
	env := cfg.GetEnv(url)

	if len(cfg.DefaultRequestCommand) > 0 {
		command.DefaultRequestCommand = cfg.DefaultRequestCommand
	}
	req := &command.Request{
		Command:   service.Command,
		Args:      args[0 : len(args)-1],
		URL:       url,
		Env:       env,
		Processes: service.Processes,
	}

	if service.UseIAP {
		authHeader, err := getAuthHeader(env)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			return 1
		}
		req.AddHeader(authHeader)
	}

	if err := req.Do(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())
		return 1
	}

	return 0
}

func getAuthHeader(env map[string]string) (string, error) {
	credentials, _ := homedir.Expand(env[iap.GoogleApplicationCredentials])
	clientID := env[iap.ClientID]

	if credentials == "" {
		return "", fmt.Errorf("%s is missing", iap.GoogleApplicationCredentials)
	}
	if clientID == "" {
		return "", fmt.Errorf("%s is missing", iap.ClientID)
	}

	token, err := iap.GetToken(credentials, clientID)
	if err != nil {
		return "", err
	}

	authHeader := fmt.Sprintf("'Authorization: Bearer %s'", token)
	return authHeader, nil
}
