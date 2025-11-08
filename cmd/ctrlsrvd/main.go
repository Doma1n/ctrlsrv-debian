package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const (
	appName    = "ctrlsrv"
	appVersion = "0.1.0"
)

var (
	configPath = flag.String("config", "../../config.yaml", "Path to config file (defaults to config.yaml or config.example.yaml)")
	noGUI      = flag.Bool("no-gui", false, "Run without opening browser (headless mode)")
	version    = flag.Bool("version", false, "Print version and exit")
)

func main() {
	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Printf("%s version %s\n", appName, appVersion)
		os.Exit(0)
	}

	// Setup logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("%s %s starting...", appName, appVersion)

	// Load configuration
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Check storage availability
	if err := checkStorage(cfg.Storage.Path); err != nil {
		log.Fatalf("Storage check failed: %v", err)
	}

	// Start API server in background
	apiServer := NewAPIServer(cfg)
	go func() {
		log.Printf("Starting HTTP API on %s", cfg.Server.ListenAddr)
		if err := apiServer.Start(); err != nil {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	// Start QUIC server in background
	if cfg.Server.QUICAddr != "" {
		quicServer := NewQUICServer(cfg)
		go func() {
			log.Printf("Starting QUIC server on %s", cfg.Server.QUICAddr)
			if err := quicServer.Start(); err != nil {
				log.Fatalf("QUIC server failed: %v", err)
			}
		}()
	}

	// Open browser in kiosk mode unless --no-gui or no DISPLAY
	if !*noGUI && os.Getenv("DISPLAY") != "" {
		go openKioskBrowser(cfg.Server.ListenAddr)
	} else if !*noGUI {
		log.Println("No DISPLAY found, running in headless mode")
	} else {
		log.Println("Running in headless mode (--no-gui)")
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}

// openKioskBrowser opens a browser in kiosk/fullscreen mode
func openKioskBrowser(addr string) {
	// Wait for server to start
	time.Sleep(2 * time.Second)

	url := fmt.Sprintf("http://%s", addr)
	if addr == "0.0.0.0:8080" || addr == ":8080" {
		url = "http://localhost:8080"
	}

	log.Printf("Opening kiosk browser for: %s", url)

	// Try browsers in order of preference (lightest first)
	browsers := [][]string{
		{"surf", "-F", url},                 // Surf (webkit, ~50MB)
		{"netsurf-gtk", "-f", url},          // NetSurf (~20MB)
		{"midori", "-e", "Fullscreen", url}, // Midori (webkit)
	}

	for _, browser := range browsers {
		if _, err := exec.LookPath(browser[0]); err == nil {
			log.Printf("Launching browser: %s", browser[0])
			cmd := exec.Command(browser[0], browser[1:]...)
			if err := cmd.Start(); err != nil {
				log.Printf("Failed to start %s: %v", browser[0], err)
				continue
			}
			return
		}
	}

	log.Printf("No suitable browser found. Access UI at: %s", url)
	log.Println("Install one of: surf, netsurf-gtk, midori")
}
