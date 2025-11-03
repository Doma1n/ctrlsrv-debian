package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"syscall"
)

// APIServer handles HTTP API requests
type APIServer struct {
	config *Config
	mux    *http.ServeMux
}

// NewAPIServer creates a new API server
func NewAPIServer(cfg *Config) *APIServer {
	s := &APIServer{
		config: cfg,
		mux:    http.NewServeMux(),
	}

	// Register routes
	s.mux.HandleFunc("/", s.handleRoot)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/storage", s.handleStorage)
	s.mux.HandleFunc("/api/printing/queues", s.handlePrintQueues)
	s.mux.HandleFunc("/api/services", s.handleServices)

	return s
}

// Start starts the API server
func (s *APIServer) Start() error {
	log.Printf("API server listening on %s", s.config.Server.ListenAddr)
	return http.ListenAndServe(s.config.Server.ListenAddr, s.mux)
}

// handleRoot serves the root page
func (s *APIServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>ctrlsrv</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 800px;
            margin: 40px auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .card {
            background: white;
            padding: 20px;
            margin: 20px 0;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 { color: #333; }
        .status { font-size: 24px; margin: 10px 0; }
        .ok { color: green; }
        .error { color: red; }
        a { color: #0066cc; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="card">
        <h1>ctrlsrv Control Panel</h1>
        <div class="status ok">✅ System Online</div>
    </div>
    <div class="card">
        <h2>API Endpoints</h2>
        <ul>
            <li><a href="/api/health">Health Check</a></li>
            <li><a href="/api/storage">Storage Status</a></li>
            <li><a href="/api/printing/queues">Print Queues</a></li>
            <li><a href="/api/services">Services Status</a></li>
        </ul>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// Response types

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Storage bool   `json:"storage_ok"`
}

type StorageResponse struct {
	Path      string  `json:"path"`
	Available bool    `json:"available"`
	Used      uint64  `json:"used_bytes"`
	Free      uint64  `json:"free_bytes"`
	Total     uint64  `json:"total_bytes"`
	UsedPct   float64 `json:"used_percent"`
}

type ServiceStatus struct {
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	Enabled bool   `json:"enabled"`
}

type ServicesResponse struct {
	Services []ServiceStatus `json:"services"`
}

// handleHealth returns system health status
func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	storageOK := checkStorage(s.config.Storage.Path) == nil

	response := HealthResponse{
		Status:  "ok",
		Version: appVersion,
		Storage: storageOK,
	}

	if !storageOK {
		response.Status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	jsonResponse(w, response)
}

// handleStorage returns storage information
func (s *APIServer) handleStorage(w http.ResponseWriter, r *http.Request) {
	available := checkStorage(s.config.Storage.Path) == nil

	response := StorageResponse{
		Path:      s.config.Storage.Path,
		Available: available,
	}

	if available {
		// Get disk usage stats
		var stat syscall.Statfs_t
		if err := syscall.Statfs(s.config.Storage.Path, &stat); err == nil {
			response.Total = stat.Blocks * uint64(stat.Bsize)
			response.Free = stat.Bfree * uint64(stat.Bsize)
			response.Used = response.Total - response.Free
			if response.Total > 0 {
				response.UsedPct = float64(response.Used) / float64(response.Total) * 100
			}
		}
	}

	jsonResponse(w, response)
}

// handlePrintQueues returns print queue status
func (s *APIServer) handlePrintQueues(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement CUPS integration
	response := map[string]interface{}{
		"queues": []map[string]interface{}{
			{
				"name":  s.config.CUPS.Printer,
				"jobs":  0,
				"state": "idle",
			},
		},
	}

	jsonResponse(w, response)
}

// handleServices returns service status
func (s *APIServer) handleServices(w http.ResponseWriter, r *http.Request) {
	services := []string{
		"cups",
		"smbd",
		"saned",
		"docker",
	}

	var statuses []ServiceStatus
	for _, name := range services {
		status := ServiceStatus{
			Name:    name,
			Active:  false, // TODO: Check actual service status
			Enabled: false,
		}
		statuses = append(statuses, status)
	}

	response := ServicesResponse{
		Services: statuses,
	}

	jsonResponse(w, response)
}

// jsonResponse writes a JSON response
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
