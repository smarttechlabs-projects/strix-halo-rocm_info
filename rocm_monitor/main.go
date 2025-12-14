package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	collector *Collector
	exporter  *Exporter
)

// Config holds application configuration
type Config struct {
	Port          int
	Interval      time.Duration
	MaxHistory    int
	AllowedOrigin string
	EnableMetrics bool
}

func main() {
	// Parse command line flags
	config := parseFlags()

	// Initialize collector with error handling
	collector = NewCollector(CollectorConfig{
		MaxHistory: config.MaxHistory,
		Interval:   config.Interval,
		ErrorCallback: func(err error) {
			log.Printf("Collector error: %v", err)
		},
	})

	// Initialize exporter
	exporter = NewExporter(collector)

	// Start data collection
	collector.Start()
	log.Printf("🚀 Started ROCm monitoring with interval: %v", config.Interval)

	// Setup HTTP routes
	setupRoutes(config)

	// Setup graceful shutdown
	setupGracefulShutdown()

	// Start HTTP server
	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("🔧 Server running on http://localhost%s", addr)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func parseFlags() Config {
	config := Config{}
	
	flag.IntVar(&config.Port, "port", 8080, "HTTP server port")
	flag.DurationVar(&config.Interval, "interval", 5*time.Second, "Collection interval")
	flag.IntVar(&config.MaxHistory, "history", 1000, "Maximum history size")
	flag.StringVar(&config.AllowedOrigin, "cors", "*", "CORS allowed origin")
	flag.BoolVar(&config.EnableMetrics, "metrics", false, "Enable Prometheus metrics endpoint")
	
	flag.Parse()
	
	return config
}

func setupRoutes(config Config) {
	// API routes
	http.HandleFunc("/api/stats", withCORS(statsHandler, config.AllowedOrigin))
	http.HandleFunc("/api/latest", withCORS(latestHandler, config.AllowedOrigin))
	http.HandleFunc("/api/gpuinfo", withCORS(gpuInfoHandler, config.AllowedOrigin))
	http.HandleFunc("/api/export.csv", withCORS(exportCSVHandler, config.AllowedOrigin))
	http.HandleFunc("/api/export.json", withCORS(exportJSONHandler, config.AllowedOrigin))
	http.HandleFunc("/api/config", withCORS(configHandler, config.AllowedOrigin))
	http.HandleFunc("/api/health", withCORS(healthHandler, config.AllowedOrigin))
	http.HandleFunc("/api/rocm-test", withCORS(rocmTestHandler, config.AllowedOrigin))
	
	// Prometheus metrics endpoint
	if config.EnableMetrics {
		http.HandleFunc("/metrics", prometheusHandler)
		log.Println("📊 Prometheus metrics enabled at /metrics")
	}
	
	// Static files
	http.Handle("/", http.FileServer(http.Dir("./static")))
}

func withCORS(handler http.HandlerFunc, allowedOrigin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		origin := r.Header.Get("Origin")
		if allowedOrigin == "*" || origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		handler(w, r)
	}
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for time window
	query := r.URL.Query()
	if windowStr := query.Get("window"); windowStr != "" {
		duration, err := time.ParseDuration(windowStr)
		if err == nil {
			// Get windowed data and return as simple array
			history := collector.GetHistory()
			if len(history) == 0 {
				http.Error(w, "No data available", http.StatusNotFound)
				return
			}

			// Filter history by time window
			cutoff := time.Now().Add(-duration)
			var filtered []RocmData
			for _, data := range history {
				if data.Timestamp.After(cutoff) {
					filtered = append(filtered, data)
				}
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(filtered); err != nil {
				http.Error(w, "Failed to encode data", http.StatusInternalServerError)
			}
			return
		}
	}
	
	// Return full history
	history := collector.GetHistory()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}

func latestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := exporter.ExportLatestJSON(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func gpuInfoHandler(w http.ResponseWriter, r *http.Request) {
	gpuInfo, err := GetGPUStaticInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get GPU info: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(gpuInfo); err != nil {
		http.Error(w, "Failed to encode GPU info", http.StatusInternalServerError)
	}
}

func exportCSVHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=rocm_stats.csv")
	
	if err := exporter.ExportCSV(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func exportJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment;filename=rocm_stats.json")
	
	if err := exporter.ExportJSON(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Update configuration
		var update struct {
			Interval string `json:"interval"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		if update.Interval != "" {
			duration, err := time.ParseDuration(update.Interval)
			if err != nil {
				http.Error(w, "Invalid interval format", http.StatusBadRequest)
				return
			}
			
			collector.SetInterval(duration)
			log.Printf("Updated collection interval to: %v", duration)
		}
		
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Return current configuration
	stats := collector.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	latest, err := collector.GetLatest()
	health := struct {
		Status    string    `json:"status"`
		Timestamp time.Time `json:"timestamp"`
		GPUCount  int       `json:"gpu_count"`
		Error     string    `json:"error,omitempty"`
	}{
		Status:    "healthy",
		Timestamp: time.Now(),
	}
	
	if err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		health.GPUCount = len(latest.GPUs)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func prometheusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	if err := exporter.ExportPrometheus(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func setupGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		log.Println("🛑 Shutting down gracefully...")
		collector.Stop()
		os.Exit(0)
	}()
}

// Helper function to parse interval from request

func parseInterval(intervalStr string) (time.Duration, error) {
	// Handle common formats
	intervalStr = strings.ToLower(strings.TrimSpace(intervalStr))
	
	// Try to parse as integer seconds
	if seconds, err := strconv.Atoi(intervalStr); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}
	
	// Try to parse as duration string
	return time.ParseDuration(intervalStr)
}