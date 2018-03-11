package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	neturl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// DefaultConfigPath is default config path
var (
	DefaultConfigPath = "config.json"
	DefaultConfigDir  = "req"
)

// Config represents config about the services
type Config struct {
	DefaultRequestCommand string    `json:"default_request_command"`
	Services              []Service `json:"services"`

	path string
}

// Service represents about information of services
type Service struct {
	URL       string   `json:"url"`
	Command   string   `json:"command"`
	UseIAP    bool     `json:"use_iap"`
	Env       Env      `json:"env"`
	Processes []string `json:"processes"`
}

// Env is the environment variables of services
type Env map[string]interface{}

func getDir() (string, error) {
	var dir string

	switch runtime.GOOS {
	default:
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	case "windows":
		dir = os.Getenv("APPDATA")
		if dir == "" {
			dir = filepath.Join(os.Getenv("USERPROFILE"), "Application Data")
		}
	}
	dir = filepath.Join(dir, DefaultConfigDir)

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return dir, fmt.Errorf("cannot create directory: %v", err)
	}

	return dir, nil
}

// LoadFile binds the config file to the structure
func (cfg *Config) LoadFile() error {
	dir, err := getDir()
	if err != nil {
		return err
	}
	cfg.path = filepath.Join(dir, DefaultConfigPath)

	_, err = os.Stat(cfg.path)
	if err == nil {
		raw, _ := ioutil.ReadFile(cfg.path)
		if err := json.Unmarshal(raw, cfg); err != nil {
			return err
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}
	f, err := os.Create(cfg.path)
	if err != nil {
		return err
	}

	// Insert sample config map as a default
	if len(cfg.Services) == 0 {
		cfg.DefaultRequestCommand = "curl"
		cfg.Services = []Service{Service{
			URL:     "https://iap-protected-app-url",
			Command: "curl",
			UseIAP:  true,
			Env: Env{
				"GOOGLE_APPLICATION_CREDENTIALS": "/path/to/google-credentials.json",
				"CLIENT_ID":                      "sample.apps.googleusercontent.com",
			},
			Processes: []string{"jq ."},
		}}
	}

	return json.NewEncoder(f).Encode(cfg)
}

// GetEnv returns Env structure with specific URL
func (cfg *Config) GetEnv(url string) map[string]string {
	env := make(map[string]string)
	service := cfg.GetService(url)
	u1, _ := neturl.Parse(url)
	u2, _ := neturl.Parse(service.URL)
	if u1.Host == u2.Host {
		for k, v := range service.Env {
			env[k] = v.(string)
		}
		return env
	}
	return env
}

// GetURLs returns list of URL
func (cfg *Config) GetURLs() (list []string) {
	for _, service := range cfg.Services {
		if service.URL == "" {
			continue
		}
		list = append(list, service.URL)
	}
	return
}

// Edit edits the config file with EDITOR
func (cfg *Config) Edit() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	command := fmt.Sprintf("%s %s", editor, cfg.path)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// GetService returns Service structure with specific URL
func (cfg *Config) GetService(url string) Service {
	for _, service := range cfg.Services {
		if url == service.URL {
			return service
		}
	}
	return Service{}
}
