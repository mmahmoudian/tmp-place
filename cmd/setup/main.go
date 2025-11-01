package setup

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mmahmoudian/tmp-place/cmd/shared"
	"github.com/spf13/cobra"
)

// setupConfigWizard guides the user through a CLI wizard to create a config.json file.
func setupConfigWizard() (shared.Config, error) {
	reader := bufio.NewReader(os.Stdin)

	prompt := func(q string, def string) (string, error) {
		if def != "" {
			fmt.Printf("%s [%s]: ", q, def)
		} else {
			fmt.Printf("%s: ", q)
		}
		text, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		text = strings.TrimSpace(text)
		if text == "" {
			return def, nil
		}
		return text, nil
	}

	promptInt := func(q string, def int) (int, error) {
		for {
			defStr := ""
			if def != 0 {
				defStr = strconv.Itoa(def)
			}
			s, err := prompt(q, defStr)
			if err != nil {
				return 0, err
			}
			if s == "" && def != 0 {
				return def, nil
			}
			n, err := strconv.Atoi(s)
			if err != nil {
				fmt.Println("Please enter a valid integer.")
				continue
			}
			return n, nil
		}
	}

	promptInt64 := func(q string, def int64) (int64, error) {
		for {
			defStr := ""
			if def != 0 {
				defStr = strconv.FormatInt(def, 10)
			}
			s, err := prompt(q, defStr)
			if err != nil {
				return 0, err
			}
			if s == "" && def != 0 {
				return def, nil
			}
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				fmt.Println("Please enter a valid integer.")
				continue
			}
			return n, nil
		}
	}

	// Defaults
	const (
		defaultHost        = "127.0.0.1"
		defaultPort        = 8080
		defaultDBFile      = "tmp-place.db"
		defaultLogFile     = "server.log"
		defaultLogLevel    = "info"
		defaultUploadsPath = "uploads"
		// Defaults below are rough and can be adjusted later
		defaultMaxSizeBytes int64 = 50 * 1024 * 1024 // 50 MB
		defaultMaxTTL       int64 = 7 * 24 * 3600    // 7 days
	)

	fmt.Println("Let's set up your configuration. Press Enter to accept defaults in [brackets].")

	// Collect inputs
	host, err := prompt("Server host", defaultHost)
	if err != nil {
		return shared.Config{}, err
	}

	port, err := promptInt("Server port", defaultPort)
	if err != nil {
		return shared.Config{}, err
	}

	dbFile, err := prompt("Path to database file", defaultDBFile)
	if err != nil {
		return shared.Config{}, err
	}

	logFile, err := prompt("Path to log file", defaultLogFile)
	if err != nil {
		return shared.Config{}, err
	}

	// log level validation
	var logLevel string
	for {
		ll, err := prompt("Log level (debug, info, warn, error)", defaultLogLevel)
		if err != nil {
			return shared.Config{}, err
		}
		llLower := strings.ToLower(ll)
		switch llLower {
		case "debug", "info", "warn", "error":
			logLevel = llLower
		default:
			fmt.Println("Please choose one of: debug, info, warn, error.")
			continue
		}
		break
	}

	uploadsPath, err := prompt("Uploads directory", defaultUploadsPath)
	if err != nil {
		return shared.Config{}, err
	}

	maxSize, err := promptInt64("Max file size (bytes)", defaultMaxSizeBytes)
	if err != nil {
		return shared.Config{}, err
	}

	maxTTL, err := promptInt64("Max TTL (seconds)", defaultMaxTTL)
	if err != nil {
		return shared.Config{}, err
	}

	// Build config struct
	cfg := shared.Config{
		Server: shared.ServerConfig{
			Host: host,
			Port: port,
			Database: shared.DatabaseConfig{
				DatabaseFile: dbFile,
			},
			Logging: shared.LoggingConfig{
				LogFile:  logFile,
				LogLevel: logLevel,
			},
		},
		Uploads: shared.UploadsConfig{
			Path:          uploadsPath,
			MaxFileSize:   maxSize,
			MaxTTLSeconds: maxTTL,
		},
	}

	// Ensure uploads directory exists
	if err := setupUploadDirectory(cfg.Uploads.Path); err != nil {
		return shared.Config{}, fmt.Errorf("failed to create uploads directory: %w", err)
	}

	// Persist to config.json
	if err := shared.SaveConfig("config.json", cfg); err != nil {
		return shared.Config{}, fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Configuration saved to config.json.")
	return cfg, nil
}

// setupUploadDirectory creates the upload directory if it doesn't exist.
func setupUploadDirectory(path string) error {
	// create upload directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
		fmt.Println("Upload directory created at:", path)
	} else {
		fmt.Println("Upload directory already exists at:", path)
	}
	return nil
}

// SetupHandler handles setup-related commands.
func SetupHandler(cmd *cobra.Command, args []string) {
	// will turn false is any of the checks fail
	alreadyDone := true

	// check if config.json exists
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		// ask user in CLI if they want to create a config file
		if !shared.AskYesNo("Would you like to go through the config creation wizard?") {
			os.Exit(1)
		}
		fmt.Println("Starting config creation wizard...")
		if _, err := setupConfigWizard(); err != nil {
			fmt.Println("Error during setup:", err)
			os.Exit(1)
		}

		// since we had to create the config, setup is not already done
		alreadyDone = false
	} else {
		fmt.Println("Config file found.")
	}

	// Load configuration
	cfg, err := shared.LoadConfig("config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// check if database file exists
	if _, err := os.Stat(cfg.Server.Database.DatabaseFile); os.IsNotExist(err) {
		fmt.Println("Database file not found at:", cfg.Server.Database.DatabaseFile)

		// ask user if they want to create the database file
		if shared.AskYesNo("Would you like to create the database file now?") {
			// create sqlite database using the db_schema.sql file
			if err := CreateDatabase(cfg.Server.Database.DatabaseFile); err != nil {
				os.Exit(1)
			}
			fmt.Println("Database file created at:", cfg.Server.Database.DatabaseFile)
		} else {
			fmt.Println("Database file creation skipped. Please create it manually using the 'db_schema.sql' file.")
			os.Exit(1)
		}

		// since we had to create the config, setup is not already done
		alreadyDone = false
	} else {
		fmt.Println("Database file found.")
	}

	// check if upload directory exists
	if _, err := os.Stat(cfg.Uploads.Path); os.IsNotExist(err) {
		fmt.Println("Uploads directory not found at:", cfg.Uploads.Path)

		// ask user if they want to create the uploads directory
		if shared.AskYesNo("Would you like to create the uploads directory now based on the config?") {
			if err := setupUploadDirectory(cfg.Uploads.Path); err != nil {
				fmt.Println("Error creating uploads directory:", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Uploads directory creation skipped. Please create it manually.")
			os.Exit(1)
		}

		// since we had to create the config, setup is not already done
		alreadyDone = false
	} else {
		fmt.Println("Uploads directory found.")
	}

	// if everything was already setup and there was no need for intervention
	if alreadyDone {
		fmt.Println("Setup is already completed. No further action is required.")
		return
	}
}
