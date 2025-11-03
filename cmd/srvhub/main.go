package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	appName    = "ctrlsrv"
	appVersion = "0.1.0"
)

var (
	configPath = flag.String("config", "", "Path to config file (defaults to config.yaml or config.example.yaml)")
	noGUI      = flag.Bool("no-gui", false, "Run without GUI (headless mode)")
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

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run GUI or headless
	if *noGUI {
		log.Println("Running in headless mode (no GUI)")
		<-sigChan
		log.Println("Shutting down...")
	} else {
		// Check if running in X session
		if os.Getenv("DISPLAY") == "" {
			log.Println("No DISPLAY found, running headless")
			<-sigChan
			log.Println("Shutting down...")
		} else {
			// Launch GUI
			runGUI(cfg, sigChan)
		}
	}

	// Cleanup
	log.Println("Shutdown complete")
}

// runGUI starts the Fyne-based touch UI
func runGUI(cfg *Config, sigChan chan os.Signal) {
	a := app.New()
	w := a.NewWindow(fmt.Sprintf("%s v%s", appName, appVersion))

	// Status label
	statusLabel := widget.NewLabel("System: OK")
	statusLabel.Alignment = 1 // Center

	// Create big touch-friendly buttons
	content := container.NewVBox(
		// Header
		widget.NewLabel(appName),
		statusLabel,
		widget.NewSeparator(),

		// Main buttons
		makeBigButton("🖨️  Print Queue", func() {
			showPrintQueue(w, cfg)
		}),

		makeBigButton("📁  Files", func() {
			showFiles(w, cfg)
		}),

		makeBigButton("💾  Storage", func() {
			showStorage(w, cfg)
		}),

		makeBigButton("⚙️  Services", func() {
			showServices(w, cfg)
		}),

		widget.NewSeparator(),

		// Utility buttons
		makeBigButton("⌨️  Terminal", func() {
			openTerminal()
		}),
	)

	w.SetContent(content)

	// Start fullscreen
	w.SetFullScreen(true)

	// Keyboard shortcuts
	w.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyF11:
			// Toggle fullscreen
			w.SetFullScreen(!w.FullScreen())
		case fyne.KeyEscape:
			// Show menu
			showMainMenu(w, cfg)
		}
	})

	// Status update goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Update status
				status := getSystemStatus(cfg)
				statusLabel.SetText(status)
			case <-sigChan:
				// Shutdown signal received
				a.Quit()
				return
			}
		}
	}()

	// Run the UI (blocking)
	w.ShowAndRun()
}

// makeBigButton creates a touch-friendly button
func makeBigButton(label string, tapped func()) *widget.Button {
	btn := widget.NewButton(label, tapped)
	btn.Importance = widget.HighImportance
	return btn
}

// UI action handlers (stubs for now)

func showPrintQueue(w fyne.Window, cfg *Config) {
	// TODO: Implement print queue view
	dialog := widget.NewLabel("Print queue view coming soon!")
	w.SetContent(container.NewVBox(
		dialog,
		makeBigButton("← Back", func() {
			// Recreate main UI
			runGUI(cfg, make(chan os.Signal, 1))
		}),
	))
}

func showFiles(w fyne.Window, cfg *Config) {
	// TODO: Implement file browser
	dialog := widget.NewLabel("File browser coming soon!")
	w.SetContent(container.NewVBox(
		dialog,
		makeBigButton("← Back", func() {
			runGUI(cfg, make(chan os.Signal, 1))
		}),
	))
}

func showStorage(w fyne.Window, cfg *Config) {
	// TODO: Implement storage status view
	info := getStorageInfo(cfg)
	w.SetContent(container.NewVBox(
		widget.NewLabel("Storage Status"),
		widget.NewLabel(info),
		makeBigButton("← Back", func() {
			runGUI(cfg, make(chan os.Signal, 1))
		}),
	))
}

func showServices(w fyne.Window, cfg *Config) {
	// TODO: Implement services view
	dialog := widget.NewLabel("Services view coming soon!")
	w.SetContent(container.NewVBox(
		dialog,
		makeBigButton("← Back", func() {
			runGUI(cfg, make(chan os.Signal, 1))
		}),
	))
}

func showMainMenu(w fyne.Window, cfg *Config) {
	// TODO: Implement main menu with settings
	dialog := widget.NewLabel("Main menu coming soon!")
	w.SetContent(container.NewVBox(
		dialog,
		makeBigButton("← Back", func() {
			runGUI(cfg, make(chan os.Signal, 1))
		}),
	))
}

func openTerminal() {
	// Try to open terminal
	// This will work when running in X session
	log.Println("Opening terminal...")
	// TODO: Implement terminal launch
}

// Helper functions (stubs for now)

func getSystemStatus(cfg *Config) string {
	// TODO: Check actual system status
	if err := checkStorage(cfg.Storage.Path); err != nil {
		return "⚠️  STORAGE ERROR"
	}
	return "✅ System: OK"
}

func getStorageInfo(cfg *Config) string {
	// TODO: Get actual storage stats
	return fmt.Sprintf("Path: %s\nStatus: Mounted\nUsage: Unknown", cfg.Storage.Path)
}
