package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	// UI routes
	s.mux.HandleFunc("/", s.handleRoot)
	s.mux.HandleFunc("/printer", s.handlePrinterPage)
	s.mux.HandleFunc("/files", s.handleFilesPage)
	s.mux.HandleFunc("/storage", s.handleStoragePage)
	s.mux.HandleFunc("/services", s.handleServicesPage)

	// API routes
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/storage", s.handleStorageAPI)
	s.mux.HandleFunc("/api/printing/queues", s.handlePrintQueues)
	s.mux.HandleFunc("/api/services", s.handleServicesAPI)

	return s
}

// Start starts the API server
func (s *APIServer) Start() error {
	log.Printf("API server listening on %s", s.config.Server.ListenAddr)
	return http.ListenAndServe(s.config.Server.ListenAddr, s.mux)
}

// Common HTML template
func (s *APIServer) renderPage(w http.ResponseWriter, title, content string) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s - ctrlsrv</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            color: white;
        }
        .header {
            padding: 15px 20px;
            background: rgba(0,0,0,0.2);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h1 { font-size: 1.5em; }
        .back-btn {
            background: rgba(255,255,255,0.2);
            border: none;
            color: white;
            padding: 10px 20px;
            border-radius: 8px;
            font-size: 1em;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
        }
        .back-btn:active { transform: scale(0.95); }
        .container {
            padding: 20px;
            max-width: 1200px;
            margin: 0 auto;
        }
        .card {
            background: rgba(255,255,255,0.15);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 25px;
            margin-bottom: 20px;
            border: 2px solid rgba(255,255,255,0.1);
        }
        .card h2 { margin-bottom: 15px; font-size: 1.5em; }
        .btn {
            background: rgba(255,255,255,0.2);
            border: 2px solid rgba(255,255,255,0.3);
            color: white;
            padding: 15px 30px;
            border-radius: 10px;
            font-size: 1.1em;
            cursor: pointer;
            margin: 5px;
            display: inline-block;
            text-decoration: none;
        }
        .btn:active { transform: scale(0.95); background: rgba(255,255,255,0.3); }
        .status-ok { color: #4ade80; }
        .status-warn { color: #fbbf24; }
        .status-error { color: #ef4444; }
        table { width: 100%%; border-collapse: collapse; margin-top: 15px; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }
        th { font-weight: 600; }
    </style>
</head>
<body>
    <div class="header">
        <h1>%s</h1>
        <a href="/" class="back-btn">← Home</a>
    </div>
    <div class="container">
        %s
    </div>
</body>
</html>`, title, title, content)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// handleRoot serves the main dashboard
func (s *APIServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>ctrlsrv</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            height: 100vh;
            display: flex;
            flex-direction: column;
            color: white;
            overflow: hidden;
        }
        .header {
            padding: 20px;
            text-align: center;
            background: rgba(0,0,0,0.2);
        }
        .header h1 { font-size: 2.5em; margin-bottom: 10px; }
        .status { font-size: 1.3em; opacity: 0.9; }
        .container {
            flex: 1;
            padding: 20px;
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            align-content: start;
            overflow-y: auto;
        }
        .card {
            background: rgba(255,255,255,0.15);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 40px 20px;
            text-align: center;
            cursor: pointer;
            transition: all 0.2s ease;
            border: 2px solid rgba(255,255,255,0.1);
            min-height: 180px;
            display: flex;
            flex-direction: column;
            justify-content: center;
            text-decoration: none;
            color: white;
        }
        .card:active {
            transform: scale(0.95);
            background: rgba(255,255,255,0.25);
        }
        .card .icon { font-size: 4em; margin-bottom: 15px; }
        .card .label { font-size: 1.5em; font-weight: 600; }
        @media (max-width: 768px) {
            .container { grid-template-columns: repeat(2, 1fr); }
            .card { min-height: 150px; padding: 30px 15px; }
            .card .icon { font-size: 3em; }
            .card .label { font-size: 1.2em; }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>ctrlsrv</h1>
        <div class="status" id="status">✅ System Online</div>
    </div>
    
    <div class="container">
        <a href="/printer" class="card">
            <div class="icon">🖨️</div>
            <div class="label">Print Queue</div>
        </a>
        
        <a href="/files" class="card">
            <div class="icon">📁</div>
            <div class="label">Files</div>
        </a>
        
        <a href="/storage" class="card">
            <div class="icon">💾</div>
            <div class="label">Storage</div>
        </a>
        
        <a href="/services" class="card">
            <div class="icon">⚙️</div>
            <div class="label">Services</div>
        </a>
    </div>
    
    <script>
        // Update status every 5 seconds
        async function updateStatus() {
            try {
                const res = await fetch('/api/health');
                const data = await res.json();
                const status = document.getElementById('status');
                if (data.status === 'ok') {
                    status.textContent = '✅ System Online';
                } else {
                    status.textContent = '⚠️ System Degraded';
                }
            } catch(e) {
                document.getElementById('status').textContent = '❌ Connection Lost';
            }
        }
        
        updateStatus();
        setInterval(updateStatus, 5000);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// handlePrinterPage shows print queue
func (s *APIServer) handlePrinterPage(w http.ResponseWriter, r *http.Request) {
	content := fmt.Sprintf(`
		<div class="card">
			<h2>🖨️ Print Queue</h2>
			<p>Printer: %s</p>
			<table>
				<thead>
					<tr>
						<th>Job ID</th>
						<th>Document</th>
						<th>Status</th>
					</tr>
				</thead>
				<tbody id="queue">
					<tr><td colspan="3">Loading...</td></tr>
				</tbody>
			</table>
		</div>
		
		<div class="card">
			<h2>Print Drop</h2>
			<p>Drop PDF files into: <code>\\ctrlsrv\printdrop</code></p>
			<p>Files will be printed automatically</p>
		</div>
		
		<script>
		async function updateQueue() {
			const res = await fetch('/api/printing/queues');
			const data = await res.json();
			const tbody = document.getElementById('queue');
			if (data.queues && data.queues.length > 0) {
				const queue = data.queues[0];
				if (queue.jobs === 0) {
					tbody.innerHTML = '<tr><td colspan="3" class="status-ok">No jobs in queue</td></tr>';
				} else {
					tbody.innerHTML = '<tr><td colspan="3">' + queue.jobs + ' jobs pending</td></tr>';
				}
			}
		}
		updateQueue();
		setInterval(updateQueue, 3000);
		</script>
	`, s.config.CUPS.Printer)

	s.renderPage(w, "Print Queue", content)
}

// handleFilesPage shows file browser
func (s *APIServer) handleFilesPage(w http.ResponseWriter, r *http.Request) {
	content := `
		<div class="card">
			<h2>📁 File Browser</h2>
			<p>Access files via Samba:</p>
			<ul style="list-style: none; padding: 20px 0;">
				<li style="margin: 10px 0;">
					<a href="#" class="btn" style="width: 100%; text-align: left;">
						💾 \\ctrlsrv\storage
					</a>
				</li>
				<li style="margin: 10px 0;">
					<a href="#" class="btn" style="width: 100%; text-align: left;">
						🖨️ \\ctrlsrv\printdrop
					</a>
				</li>
			</ul>
		</div>
		
		<div class="card">
			<h2>Quick Actions</h2>
			<button class="btn" onclick="alert('Opening file manager...')">
				Open File Manager
			</button>
		</div>
	`

	s.renderPage(w, "Files", content)
}

// handleStoragePage shows storage info
func (s *APIServer) handleStoragePage(w http.ResponseWriter, r *http.Request) {
	content := `
		<div class="card">
			<h2>💾 Storage Status</h2>
			<div id="storage-info">Loading...</div>
		</div>
		
		<script>
		async function updateStorage() {
			try {
				const res = await fetch('/api/storage');
				const data = await res.json();
				const div = document.getElementById('storage-info');
				
				if (!data.available) {
					div.innerHTML = '<p class="status-error">⚠️ Storage not available!</p>';
					return;
				}
				
				const used = (data.used_bytes / 1024 / 1024 / 1024).toFixed(2);
				const total = (data.total_bytes / 1024 / 1024 / 1024).toFixed(2);
				const free = (data.free_bytes / 1024 / 1024 / 1024).toFixed(2);
				const pct = data.used_percent.toFixed(1);
				
				let statusClass = 'status-ok';
				if (pct > 90) statusClass = 'status-error';
				else if (pct > 80) statusClass = 'status-warn';
				
				div.innerHTML = '
	<table>
	<tr><td>Path:</td><td><code>\${data.path}</code></td></tr>
	<tr><td>Status:</td><td class="status-ok">✅ Mounted</td></tr>
	<tr><td>Used:</td><td>\${used} GB</td></tr>
	<tr><td>Free:</td><td>\${free} GB</td></tr>
	<tr><td>Total:</td><td>\${total} GB</td></tr>
	<tr><td>Usage:</td><td class="\${statusClass}">\${pct}%</td></tr>
	</table>
	\';
			} catch (e) {
				document.getElementById('storage-info').innerHTML = 
					'<p class="status-error">❌ Failed to load storage info</p>';
			}
		}
		updateStorage();
		setInterval(updateStorage, 10000);
		</script>
	`

	s.renderPage(w, "Storage", content)
}

// handleServicesPage shows service status
func (s *APIServer) handleServicesPage(w http.ResponseWriter, r *http.Request) {
	content := `
		<div class="card">
			<h2>⚙️ System Services</h2>
			<table id="services-table">
				<thead>
					<tr>
						<th>Service</th>
						<th>Status</th>
					</tr>
				</thead>
				<tbody>
					<tr><td colspan="2">Loading...</td></tr>
				</tbody>
			</table>
		</div>
		
		<script>
		async function updateServices() {
			try {
				const res = await fetch(\'/api/services\');
				const data = await res.json();
				const tbody = document.querySelector(\'#services-table tbody\');
				
				tbody.innerHTML = data.services.map(s => {
					const status = s.active ? 
						\'<span class="status-ok">✅ Active</span>\' : 
						\'<span class="status-error">❌ Inactive</span>\';
					return \'<tr><td>\${s.name}</td><td>\${status}</td></tr>\';
				}).join('');
			} catch (e) {
				document.querySelector('#services-table tbody').innerHTML = 
					'<tr><td colspan="2" class="status-error">Failed to load</td></tr>';
			}
		}
		updateServices();
		setInterval(updateServices, 5000);
		</script>
	`

	s.renderPage(w, "Services", content)
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

// API handlers
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

func (s *APIServer) handleStorageAPI(w http.ResponseWriter, r *http.Request) {
	available := checkStorage(s.config.Storage.Path) == nil

	response := StorageResponse{
		Path:      s.config.Storage.Path,
		Available: available,
	}

	if available {
		// Use the cross-platform getStorageUsage function
		used, free, total, err := getStorageUsage(s.config.Storage.Path)
		if err == nil {
			response.Total = total
			response.Free = free
			response.Used = used
			if response.Total > 0 {
				response.UsedPct = float64(response.Used) / float64(response.Total) * 100
			}
		}
	}

	jsonResponse(w, response)
}

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

func (s *APIServer) handleServicesAPI(w http.ResponseWriter, r *http.Request) {
	sm := NewServiceManager()

	// List of services to monitor
	services := []string{
		"cups",
		"smbd",
		"nmbd",
		"saned",
		"docker",
	}

	// Get status for all services
	statuses := sm.GetMultipleStatuses(services)

	response := ServicesResponse{
		Services: statuses,
	}

	jsonResponse(w, response)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
