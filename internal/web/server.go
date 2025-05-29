package web

import (
    "encoding/json"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "time"
    
    "github.com/s1-callgen/internal/generator"
    "github.com/s1-callgen/internal/models"
)

type WebServer struct {
    config    *models.Config
    generator *generator.Generator
    templates *template.Template
}

func NewWebServer(config *models.Config, gen *generator.Generator) *WebServer {
    return &WebServer{
        config:    config,
        generator: gen,
    }
}

func (w *WebServer) Start() error {
    if !w.config.WebInterface.Enabled {
        return nil
    }
    
    // Setup routes
    http.HandleFunc("/", w.authMiddleware(w.handleDashboard))
    http.HandleFunc("/api/stats", w.authMiddleware(w.handleStats))
    http.HandleFunc("/api/config", w.authMiddleware(w.handleConfig))
    http.HandleFunc("/api/numbers", w.authMiddleware(w.handleNumbers))
    http.HandleFunc("/api/control", w.authMiddleware(w.handleControl))
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    
    addr := fmt.Sprintf(":%d", w.config.WebInterface.Port)
    log.Printf("[WEB] Starting web interface on http://localhost%s", addr)
    
    return http.ListenAndServe(addr, nil)
}

func (w *WebServer) authMiddleware(handler http.HandlerFunc) http.HandlerFunc {
    return func(rw http.ResponseWriter, r *http.Request) {
        user, pass, ok := r.BasicAuth()
        if !ok || user != w.config.WebInterface.Auth.Username || 
           pass != w.config.WebInterface.Auth.Password {
            rw.Header().Set("WWW-Authenticate", `Basic realm="S1 Call Generator"`)
            http.Error(rw, "Unauthorized", http.StatusUnauthorized)
            return
        }
        handler(rw, r)
    }
}

func (w *WebServer) handleDashboard(rw http.ResponseWriter, r *http.Request) {
    // Serve the dashboard HTML
    fmt.Fprintf(rw, dashboardHTML)
}

func (w *WebServer) handleStats(rw http.ResponseWriter, r *http.Request) {
    stats := w.generator.GetStatistics()
    
    rw.Header().Set("Content-Type", "application/json")
    json.NewEncoder(rw).Encode(stats)
}

func (w *WebServer) handleConfig(rw http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        rw.Header().Set("Content-Type", "application/json")
        json.NewEncoder(rw).Encode(w.config)
    case "POST":
        // Update configuration
        var newConfig models.Config
        if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
            http.Error(rw, err.Error(), http.StatusBadRequest)
            return
        }
        
        // Apply new configuration
        w.config = &newConfig
        rw.WriteHeader(http.StatusOK)
    }
}

func (w *WebServer) handleNumbers(rw http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST":
        // Handle CSV upload or manual entry
        if err := r.ParseMultipartForm(32 << 20); err != nil {
            http.Error(rw, err.Error(), http.StatusBadRequest)
            return
        }
        
        file, _, err := r.FormFile("csv")
        if err == nil {
            // Process CSV file
            defer file.Close()
            // Implementation for CSV processing
        } else {
            // Process manual entry
            numbers := r.FormValue("numbers")
            // Implementation for manual number processing
        }
        
        rw.WriteHeader(http.StatusOK)
    }
}

func (w *WebServer) handleControl(rw http.ResponseWriter, r *http.Request) {
    var req struct {
        Action string `json:"action"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(rw, err.Error(), http.StatusBadRequest)
        return
    }
    
    switch req.Action {
    case "start":
        go w.generator.Start()
    case "stop":
        w.generator.Stop()
    case "toggle_autopilot":
        w.config.Autopilot.Enabled = !w.config.Autopilot.Enabled
    }
    
    rw.WriteHeader(http.StatusOK)
}

const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>S1 Call Generator Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .card { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; }
        .stat-box { text-align: center; }
        .stat-value { font-size: 2em; font-weight: bold; color: #2196F3; }
        .stat-label { color: #666; margin-top: 5px; }
        .chart-container { height: 300px; margin-top: 20px; }
        .controls { display: flex; gap: 10px; margin-bottom: 20px; }
        button { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #2196F3; color: white; }
        .btn-danger { background: #f44336; color: white; }
        .btn-success { background: #4CAF50; color: white; }
        .status { display: inline-block; padding: 5px 10px; border-radius: 4px; }
        .status.active { background: #4CAF50; color: white; }
        .status.inactive { background: #f44336; color: white; }
        #realtimeChart { width: 100%; height: 300px; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="container">
        <h1>S1 Call Generator Dashboard</h1>
        
        <div class="card">
            <h2>Controls</h2>
            <div class="controls">
                <button class="btn-primary" onclick="startGenerator()">Start</button>
                <button class="btn-danger" onclick="stopGenerator()">Stop</button>
                <button class="btn-success" onclick="toggleAutopilot()">Toggle Autopilot</button>
                <span class="status" id="status">Inactive</span>
                <span class="status" id="autopilot-status">Autopilot: OFF</span>
           </div>
       </div>
       
       <div class="card">
           <h2>Real-time Statistics</h2>
           <div class="stats-grid">
               <div class="stat-box">
                   <div class="stat-value" id="total-calls">0</div>
                   <div class="stat-label">Total Calls</div>
               </div>
               <div class="stat-box">
                   <div class="stat-value" id="active-calls">0</div>
                   <div class="stat-label">Active Calls</div>
               </div>
               <div class="stat-box">
                   <div class="stat-value" id="success-rate">0%</div>
                   <div class="stat-label">ASR</div>
               </div>
               <div class="stat-box">
                   <div class="stat-value" id="cps">0</div>
                   <div class="stat-label">Calls/Second</div>
               </div>
               <div class="stat-box">
                   <div class="stat-value" id="avg-duration">0s</div>
                   <div class="stat-label">Avg Duration</div>
               </div>
           </div>
       </div>
       
       <div class="card">
           <h2>Call Traffic Pattern</h2>
           <canvas id="realtimeChart"></canvas>
       </div>
       
       <div class="card">
           <h2>Upload Numbers</h2>
           <form id="upload-form">
               <input type="file" id="csv-file" accept=".csv">
               <button type="submit" class="btn-primary">Upload CSV</button>
           </form>
           <p>Or enter numbers manually:</p>
           <textarea id="manual-numbers" rows="5" cols="50" placeholder="ANI,DNIS,Country,Carrier"></textarea>
           <button onclick="uploadManual()" class="btn-primary">Add Numbers</button>
       </div>
   </div>
   
   <script>
       let chart;
       let chartData = {
           labels: [],
           datasets: [{
               label: 'Simultaneous Calls',
               data: [],
               borderColor: 'rgb(75, 192, 192)',
               backgroundColor: 'rgba(75, 192, 192, 0.2)',
               tension: 0.1
           }, {
               label: 'Call Attempts',
               data: [],
               borderColor: 'rgb(255, 99, 132)',
               backgroundColor: 'rgba(255, 99, 132, 0.2)',
               tension: 0.1
           }]
       };
       
       function initChart() {
           const ctx = document.getElementById('realtimeChart').getContext('2d');
           chart = new Chart(ctx, {
               type: 'line',
               data: chartData,
               options: {
                   responsive: true,
                   maintainAspectRatio: false,
                   scales: {
                       y: {
                           beginAtZero: true
                       }
                   },
                   plugins: {
                       legend: {
                           display: true,
                           position: 'top'
                       }
                   }
               }
           });
       }
       
       function updateStats() {
           fetch('/api/stats', {
               headers: {
                   'Authorization': 'Basic ' + btoa('admin:admin')
               }
           })
           .then(response => response.json())
           .then(data => {
               document.getElementById('total-calls').textContent = data.total_calls;
               document.getElementById('active-calls').textContent = data.active_calls;
               document.getElementById('success-rate').textContent = data.current_asr.toFixed(1) + '%';
               document.getElementById('cps').textContent = data.current_cps.toFixed(2);
               document.getElementById('avg-duration').textContent = data.average_call_duration.toFixed(1) + 's';
               
               // Update chart
               const now = new Date().toLocaleTimeString();
               chartData.labels.push(now);
               chartData.datasets[0].data.push(data.active_calls);
               chartData.datasets[1].data.push(data.total_calls);
               
               // Keep only last 50 points
               if (chartData.labels.length > 50) {
                   chartData.labels.shift();
                   chartData.datasets[0].data.shift();
                   chartData.datasets[1].data.shift();
               }
               
               chart.update();
           });
       }
       
       function startGenerator() {
           fetch('/api/control', {
               method: 'POST',
               headers: {
                   'Authorization': 'Basic ' + btoa('admin:admin'),
                   'Content-Type': 'application/json'
               },
               body: JSON.stringify({action: 'start'})
           }).then(() => {
               document.getElementById('status').className = 'status active';
               document.getElementById('status').textContent = 'Active';
           });
       }
       
       function stopGenerator() {
           fetch('/api/control', {
               method: 'POST',
               headers: {
                   'Authorization': 'Basic ' + btoa('admin:admin'),
                   'Content-Type': 'application/json'
               },
               body: JSON.stringify({action: 'stop'})
           }).then(() => {
               document.getElementById('status').className = 'status inactive';
               document.getElementById('status').textContent = 'Inactive';
           });
       }
       
       function toggleAutopilot() {
           fetch('/api/control', {
               method: 'POST',
               headers: {
                   'Authorization': 'Basic ' + btoa('admin:admin'),
                   'Content-Type': 'application/json'
               },
               body: JSON.stringify({action: 'toggle_autopilot'})
           }).then(() => {
               const status = document.getElementById('autopilot-status');
               if (status.textContent.includes('OFF')) {
                   status.textContent = 'Autopilot: ON';
                   status.className = 'status active';
               } else {
                   status.textContent = 'Autopilot: OFF';
                   status.className = 'status inactive';
               }
           });
       }
       
       function uploadManual() {
           const numbers = document.getElementById('manual-numbers').value;
           // Implementation for manual upload
       }
       
       // Initialize
       initChart();
       setInterval(updateStats, 2000);
       
       // Handle file upload
       document.getElementById('upload-form').addEventListener('submit', function(e) {
           e.preventDefault();
           const fileInput = document.getElementById('csv-file');
           const file = fileInput.files[0];
           
           if (file) {
               const formData = new FormData();
               formData.append('csv', file);
               
               fetch('/api/numbers', {
                   method: 'POST',
                   headers: {
                       'Authorization': 'Basic ' + btoa('admin:admin')
                   },
                   body: formData
               }).then(() => {
                   alert('Numbers uploaded successfully!');
                   fileInput.value = '';
               });
           }
       });
   </script>
</body>
</html>
`
