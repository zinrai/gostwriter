package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/google/shlex"
	"gopkg.in/yaml.v3"
)

const version = "0.1.0"

// Config represents the root configuration structure
type Config struct {
	Defaults     map[string]interface{} `yaml:"defaults"`
	Environments map[string]Environment `yaml:"environments"`
}

// Environment represents an environment configuration
type Environment struct {
	Vars        map[string]interface{} `yaml:"vars"`
	GostCommand string                 `yaml:"gost_command"`
}

func main() {
	// Handle version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("gostwriter %s\n", version)
		os.Exit(0)
	}

	// Parse flags
	var env string
	var configFile string
	flag.StringVar(&env, "env", "", "Environment name (required)")
	flag.StringVar(&configFile, "config", "gost-config.yml", "Configuration file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gostwriter - A ghostwriter for your gost commands\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  gostwriter --env=ENVIRONMENT [--config=FILE]\n")
		fmt.Fprintf(os.Stderr, "  gostwriter --version\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  gostwriter --env=production\n")
		fmt.Fprintf(os.Stderr, "  gostwriter --env=staging --config=/path/to/config.yml\n")
		fmt.Fprintf(os.Stderr, "  gost $(gostwriter --env=production)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Validate required arguments
	if env == "" {
		fmt.Fprintf(os.Stderr, "Error: --env is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Generate command
	args, err := generateCommand(config, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Output the command (space-separated)
	fmt.Println(strings.Join(args, " "))
}

func loadConfig(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file '%s' not found", configFile)
		}
		return nil, fmt.Errorf("reading configuration: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	return &config, nil
}

func generateCommand(config *Config, envName string) ([]string, error) {
	// Find environment
	env, exists := config.Environments[envName]
	if !exists {
		availableEnvs := make([]string, 0, len(config.Environments))
		for name := range config.Environments {
			availableEnvs = append(availableEnvs, name)
		}
		return nil, fmt.Errorf("environment '%s' not found (available: %s)",
			envName, strings.Join(availableEnvs, ", "))
	}

	if env.GostCommand == "" {
		return nil, fmt.Errorf("environment '%s' has no gost_command defined", envName)
	}

	// Prepare template data
	data := make(map[string]interface{})

	// Add defaults
	if config.Defaults != nil {
		for k, v := range config.Defaults {
			data[k] = v
		}
	}

	// Add environment-specific vars (overrides defaults)
	if env.Vars != nil {
		for k, v := range env.Vars {
			data[k] = v
		}
	}

	// Parse and execute template
	tmpl, err := template.New("gost").Parse(env.GostCommand)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(buf.String())

	// Parse into arguments
	args := parseArguments(expanded)

	if len(args) == 0 {
		return nil, fmt.Errorf("no arguments generated from template")
	}

	return args, nil
}

func parseArguments(cmd string) []string {
	// Combine all lines into a single command string
	lines := strings.Split(cmd, "\n")
	var fullCmd strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			if fullCmd.Len() > 0 {
				fullCmd.WriteString(" ")
			}
			fullCmd.WriteString(line)
		}
	}

	// Use shlex to properly parse shell-like syntax
	args, err := shlex.Split(fullCmd.String())
	if err != nil {
		// Fallback to simple split if shlex fails
		fmt.Fprintf(os.Stderr, "Warning: Failed to parse arguments with shlex: %v\n", err)
		fmt.Fprintf(os.Stderr, "Falling back to simple split\n")
		return strings.Fields(fullCmd.String())
	}

	return args
}
