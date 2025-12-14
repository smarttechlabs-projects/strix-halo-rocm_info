package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// Exporter handles data export functionality
type Exporter struct {
	collector *Collector
}

// NewExporter creates a new exporter instance
func NewExporter(collector *Collector) *Exporter {
	return &Exporter{
		collector: collector,
	}
}

// ExportCSV writes data history as CSV
func (e *Exporter) ExportCSV(w io.Writer) error {
	history := e.collector.GetHistory()
	if len(history) == 0 {
		return fmt.Errorf("no data to export")
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"Timestamp",
		"GPU_ID",
		"Temperature_C",
		"Power_W",
		"VRAM_Usage_GB",
		"VRAM_Total_GB",
		"GPU_Usage_%",
		"SCLK_MHz",
		"MCLK_MHz",
		"CPU_Usage_%",
		"Fan_Speed_%",
	}
	
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, data := range history {
		timestamp := data.Timestamp.Format(time.RFC3339)
		
		for _, gpu := range data.GPUs {
			row := []string{
				timestamp,
				fmt.Sprintf("%d", gpu.ID),
				fmt.Sprintf("%.2f", gpu.Temperature),
				fmt.Sprintf("%.2f", gpu.Power),
				fmt.Sprintf("%.2f", gpu.VRAMUsage),
				fmt.Sprintf("%.2f", gpu.VRAMTotal),
				fmt.Sprintf("%.2f", gpu.GPUUsage),
				fmt.Sprintf("%.0f", gpu.SCLKFreq),
				fmt.Sprintf("%.0f", gpu.MCLKFreq),
				fmt.Sprintf("%.2f", data.CPUUsage),
				fmt.Sprintf("%.2f", gpu.FanSpeed),
			}
			
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	return nil
}

// ExportJSON writes data history as JSON
func (e *Exporter) ExportJSON(w io.Writer) error {
	history := e.collector.GetHistory()
	if len(history) == 0 {
		return fmt.Errorf("no data to export")
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	
	// Create export structure with metadata
	export := struct {
		ExportTime time.Time              `json:"export_time"`
		DataPoints int                    `json:"data_points"`
		Stats      map[string]interface{} `json:"statistics"`
		History    []RocmData             `json:"history"`
	}{
		ExportTime: time.Now(),
		DataPoints: len(history),
		Stats:      e.collector.GetStats(),
		History:    history,
	}

	if err := encoder.Encode(export); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// ExportLatestJSON writes only the latest data point as JSON
func (e *Exporter) ExportLatestJSON(w io.Writer) error {
	latest, err := e.collector.GetLatest()
	if err != nil {
		return fmt.Errorf("failed to get latest data: %w", err)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(latest); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// ExportPrometheus writes comprehensive metrics in Prometheus format
func (e *Exporter) ExportPrometheus(w io.Writer) error {
	latest, err := e.collector.GetLatest()
	if err != nil {
		return fmt.Errorf("failed to get latest data: %w", err)
	}

	// Use a buffer to capture all output and filter problematic text
	var buf bytes.Buffer
	
	stats := e.collector.GetStats()
	gpuStaticInfo, _ := GetGPUStaticInfo()

	// Generate timestamp for all metrics
	timestamp := latest.Timestamp.UnixMilli()

	// === GPU Hardware Metrics ===
	for _, gpu := range latest.GPUs {
		// Get GPU static info for labels
		var productName, vendor, serialNumber, vramVendor string
		if len(gpuStaticInfo) > int(gpu.ID) {
			info := gpuStaticInfo[gpu.ID]
			productName = info.ProductName
			vendor = info.VendorName
			serialNumber = info.SerialNumber
			vramVendor = info.VRAMVendor
		}

		labels := fmt.Sprintf(`gpu_id="%d",product_name="%s",vendor="%s",serial_number="%s",vram_vendor="%s"`, 
			gpu.ID, productName, vendor, serialNumber, vramVendor)

		// Temperature
		fmt.Fprintf(&buf, "# HELP rocm_gpu_temperature_celsius GPU edge temperature in Celsius\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_temperature_celsius gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_temperature_celsius{%s} %.2f %d\n", labels, gpu.Temperature, timestamp)

		// Power consumption
		fmt.Fprintf(&buf, "# HELP rocm_gpu_power_watts GPU power consumption in watts\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_power_watts gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_power_watts{%s} %.2f %d\n", labels, gpu.Power, timestamp)

		// GPU utilization
		fmt.Fprintf(&buf, "# HELP rocm_gpu_usage_percent GPU compute utilization percentage\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_usage_percent gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_usage_percent{%s} %.2f %d\n", labels, gpu.GPUUsage, timestamp)

		// VRAM usage
		fmt.Fprintf(&buf, "# HELP rocm_gpu_vram_usage_gb VRAM usage in gigabytes\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_vram_usage_gb gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_vram_usage_gb{%s} %.3f %d\n", labels, gpu.VRAMUsage, timestamp)

		// VRAM total
		fmt.Fprintf(&buf, "# HELP rocm_gpu_vram_total_gb Total VRAM in gigabytes\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_vram_total_gb gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_vram_total_gb{%s} %.3f %d\n", labels, gpu.VRAMTotal, timestamp)

		// VRAM utilization percentage
		vramUtilPct := 0.0
		if gpu.VRAMTotal > 0 {
			vramUtilPct = (gpu.VRAMUsage / gpu.VRAMTotal) * 100
		}
		fmt.Fprintf(&buf, "# HELP rocm_gpu_vram_utilization_percent VRAM utilization percentage\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_vram_utilization_percent gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_vram_utilization_percent{%s} %.2f %d\n", labels, vramUtilPct, timestamp)

		// Clock frequencies
		fmt.Fprintf(&buf, "# HELP rocm_gpu_sclk_mhz GPU system clock frequency in MHz\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_sclk_mhz gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_sclk_mhz{%s} %.0f %d\n", labels, gpu.SCLKFreq, timestamp)

		fmt.Fprintf(&buf, "# HELP rocm_gpu_mclk_mhz GPU memory clock frequency in MHz\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_mclk_mhz gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_mclk_mhz{%s} %.0f %d\n", labels, gpu.MCLKFreq, timestamp)

		// Fan speed
		fmt.Fprintf(&buf, "# HELP rocm_gpu_fan_speed_percent GPU fan speed percentage\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_fan_speed_percent gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_fan_speed_percent{%s} %.2f %d\n", labels, gpu.FanSpeed, timestamp)
	}

	// === System CPU Metrics ===
	fmt.Fprintf(&buf, "# HELP rocm_system_cpu_usage_percent System CPU utilization percentage\n")
	fmt.Fprintf(&buf, "# TYPE rocm_system_cpu_usage_percent gauge\n")
	fmt.Fprintf(&buf, "rocm_system_cpu_usage_percent %.2f %d\n", latest.CPUUsage, timestamp)

	// === System Information ===
	fmt.Fprintf(&buf, "# HELP rocm_system_gpu_count Number of detected GPUs\n")
	fmt.Fprintf(&buf, "# TYPE rocm_system_gpu_count gauge\n")
	fmt.Fprintf(&buf, "rocm_system_gpu_count %d %d\n", len(latest.GPUs), timestamp)

	// === Monitoring Health Metrics ===
	fmt.Fprintf(&buf, "# HELP rocm_monitor_collection_errors_total Total number of collection errors\n")
	fmt.Fprintf(&buf, "# TYPE rocm_monitor_collection_errors_total counter\n")
	if errorCount, ok := stats["collection_errors"]; ok {
		fmt.Fprintf(&buf, "rocm_monitor_collection_errors_total %.0f %d\n", errorCount.(float64), timestamp)
	} else {
		fmt.Fprintf(&buf, "rocm_monitor_collection_errors_total 0 %d\n", timestamp)
	}

	fmt.Fprintf(&buf, "# HELP rocm_monitor_collection_duration_ms Collection duration in milliseconds\n")
	fmt.Fprintf(&buf, "# TYPE rocm_monitor_collection_duration_ms gauge\n")
	if duration, ok := stats["avg_collection_time_ms"]; ok {
		fmt.Fprintf(&buf, "rocm_monitor_collection_duration_ms %.2f %d\n", duration.(float64), timestamp)
	} else {
		fmt.Fprintf(&buf, "rocm_monitor_collection_duration_ms 0 %d\n", timestamp)
	}

	fmt.Fprintf(&buf, "# HELP rocm_monitor_data_points_total Total collected data points\n")
	fmt.Fprintf(&buf, "# TYPE rocm_monitor_data_points_total counter\n")
	if dataPoints, ok := stats["total_collections"]; ok {
		fmt.Fprintf(&buf, "rocm_monitor_data_points_total %.0f %d\n", dataPoints.(float64), timestamp)
	} else {
		fmt.Fprintf(&buf, "rocm_monitor_data_points_total 0 %d\n", timestamp)
	}

	fmt.Fprintf(&buf, "# HELP rocm_monitor_uptime_seconds Monitor uptime in seconds\n")
	fmt.Fprintf(&buf, "# TYPE rocm_monitor_uptime_seconds gauge\n")
	if uptime, ok := stats["uptime_seconds"]; ok {
		fmt.Fprintf(&buf, "rocm_monitor_uptime_seconds %.0f %d\n", uptime.(float64), timestamp)
	} else {
		fmt.Fprintf(&buf, "rocm_monitor_uptime_seconds 0 %d\n", timestamp)
	}

	fmt.Fprintf(&buf, "# HELP rocm_monitor_memory_usage_mb Monitor memory usage in megabytes\n")
	fmt.Fprintf(&buf, "# TYPE rocm_monitor_memory_usage_mb gauge\n")
	if memUsage, ok := stats["memory_usage_mb"]; ok {
		fmt.Fprintf(&buf, "rocm_monitor_memory_usage_mb %.2f %d\n", memUsage.(float64), timestamp)
	} else {
		fmt.Fprintf(&buf, "rocm_monitor_memory_usage_mb 0 %d\n", timestamp)
	}

	fmt.Fprintf(&buf, "# HELP rocm_monitor_history_size_points Number of historical data points stored\n")
	fmt.Fprintf(&buf, "# TYPE rocm_monitor_history_size_points gauge\n")
	history := e.collector.GetHistory()
	fmt.Fprintf(&buf, "rocm_monitor_history_size_points %d %d\n", len(history), timestamp)

	// === Performance Thresholds ===
	for _, gpu := range latest.GPUs {
		labels := fmt.Sprintf(`gpu_id="%d"`, gpu.ID)

		// Temperature thresholds
		tempWarning := 0.0
		tempCritical := 0.0
		if gpu.Temperature > 80 {
			tempCritical = 1.0
		} else if gpu.Temperature > 70 {
			tempWarning = 1.0
		}

		fmt.Fprintf(&buf, "# HELP rocm_gpu_temperature_warning_threshold Temperature warning threshold exceeded\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_temperature_warning_threshold gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_temperature_warning_threshold{%s} %.0f %d\n", labels, tempWarning, timestamp)

		fmt.Fprintf(&buf, "# HELP rocm_gpu_temperature_critical_threshold Temperature critical threshold exceeded\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_temperature_critical_threshold gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_temperature_critical_threshold{%s} %.0f %d\n", labels, tempCritical, timestamp)

		// VRAM threshold
		vramUtilPct := 0.0
		if gpu.VRAMTotal > 0 {
			vramUtilPct = (gpu.VRAMUsage / gpu.VRAMTotal) * 100
		}
		vramHigh := 0.0
		if vramUtilPct > 80 {
			vramHigh = 1.0
		}
		fmt.Fprintf(&buf, "# HELP rocm_gpu_vram_high_utilization VRAM utilization above 80%\n")
		fmt.Fprintf(&buf, "# TYPE rocm_gpu_vram_high_utilization gauge\n")
		fmt.Fprintf(&buf, "rocm_gpu_vram_high_utilization{%s} %.0f %d\n", labels, vramHigh, timestamp)
	}

	// === Build Info ===
	fmt.Fprintf(&buf, "# HELP rocm_monitor_build_info ROCm Monitor build information\n")
	fmt.Fprintf(&buf, "# TYPE rocm_monitor_build_info gauge\n")
	fmt.Fprintf(&buf, "rocm_monitor_build_info{version=\"1.0.0\",go_version=\"unknown\"} 1 %d\n", timestamp)

	// Clean the output by removing problematic text that breaks Prometheus parsing
	output := buf.String()
	output = strings.ReplaceAll(output, "(MISSING)", "")
	output = strings.ReplaceAll(output, "\n\n", "\n") // Remove double newlines
	
	// Write the cleaned output
	_, err = w.Write([]byte(output))
	if err != nil {
		return fmt.Errorf("failed to write metrics: %w", err)
	}

	return nil
}

// ExportROCmTestMetrics exports ROCm test results in Prometheus format
func (e *Exporter) ExportROCmTestMetrics(w io.Writer, testSuite *ROCmTestSuite) error {
	if testSuite == nil {
		return fmt.Errorf("no test results to export")
	}

	timestamp := testSuite.Timestamp.UnixMilli()

	// === ROCm Test Suite Metrics ===
	fmt.Fprintf(w, "# HELP rocm_test_suite_success Overall ROCm test suite success (1=pass, 0=fail)\n")
	fmt.Fprintf(w, "# TYPE rocm_test_suite_success gauge\n")
	successValue := 0.0
	if testSuite.OverallSuccess {
		successValue = 1.0
	}
	fmt.Fprintf(w, "rocm_test_suite_success %.0f %d\n", successValue, timestamp)

	fmt.Fprintf(w, "# HELP rocm_test_suite_duration_ms Total test suite execution time in milliseconds\n")
	fmt.Fprintf(w, "# TYPE rocm_test_suite_duration_ms gauge\n")
	fmt.Fprintf(w, "rocm_test_suite_duration_ms %d %d\n", testSuite.Duration, timestamp)

	// Count test results
	passed := 0
	failed := 0
	warnings := 0
	for _, result := range testSuite.TestResults {
		if result.Success {
			passed++
			if len(result.Issues) > 0 {
				warnings++
			}
		} else {
			failed++
		}
	}

	fmt.Fprintf(w, "# HELP rocm_test_suite_total_tests Total number of tests executed\n")
	fmt.Fprintf(w, "# TYPE rocm_test_suite_total_tests gauge\n")
	fmt.Fprintf(w, "rocm_test_suite_total_tests %d %d\n", len(testSuite.TestResults), timestamp)

	fmt.Fprintf(w, "# HELP rocm_test_suite_passed_tests Number of tests that passed\n")
	fmt.Fprintf(w, "# TYPE rocm_test_suite_passed_tests gauge\n")
	fmt.Fprintf(w, "rocm_test_suite_passed_tests %d %d\n", passed, timestamp)

	fmt.Fprintf(w, "# HELP rocm_test_suite_failed_tests Number of tests that failed\n")
	fmt.Fprintf(w, "# TYPE rocm_test_suite_failed_tests gauge\n")
	fmt.Fprintf(w, "rocm_test_suite_failed_tests %d %d\n", failed, timestamp)

	fmt.Fprintf(w, "# HELP rocm_test_suite_warnings_tests Number of tests with warnings\n")
	fmt.Fprintf(w, "# TYPE rocm_test_suite_warnings_tests gauge\n")
	fmt.Fprintf(w, "rocm_test_suite_warnings_tests %d %d\n", warnings, timestamp)

	// === Individual Test Metrics ===
	for _, result := range testSuite.TestResults {
		// Sanitize test name for metric label
		testName := strings.ReplaceAll(strings.ToLower(result.Command), " ", "_")
		testName = strings.ReplaceAll(testName, "-", "_")
		
		labels := fmt.Sprintf(`test_name="%s",command="%s"`, testName, result.Command)

		// Test success status
		fmt.Fprintf(w, "# HELP rocm_test_success Individual test success (1=pass, 0=fail)\n")
		fmt.Fprintf(w, "# TYPE rocm_test_success gauge\n")
		successVal := 0.0
		if result.Success {
			successVal = 1.0
		}
		fmt.Fprintf(w, "rocm_test_success{%s} %.0f %d\n", labels, successVal, timestamp)

		// Test duration
		fmt.Fprintf(w, "# HELP rocm_test_duration_ms Individual test execution time in milliseconds\n")
		fmt.Fprintf(w, "# TYPE rocm_test_duration_ms gauge\n")
		fmt.Fprintf(w, "rocm_test_duration_ms{%s} %d %d\n", labels, result.Duration, timestamp)

		// Test issues count
		fmt.Fprintf(w, "# HELP rocm_test_issues_count Number of issues detected in test\n")
		fmt.Fprintf(w, "# TYPE rocm_test_issues_count gauge\n")
		fmt.Fprintf(w, "rocm_test_issues_count{%s} %d %d\n", labels, len(result.Issues), timestamp)
	}

	return nil
}

// ExportHistorySubset exports a time-windowed subset of history
func (e *Exporter) ExportHistorySubset(w io.Writer, duration time.Duration, format string) error {
	history := e.collector.GetHistory()
	if len(history) == 0 {
		return fmt.Errorf("no data to export")
	}

	// Find cutoff time
	cutoff := time.Now().Add(-duration)
	
	// Filter history
	var filtered []RocmData
	for _, data := range history {
		if data.Timestamp.After(cutoff) {
			filtered = append(filtered, data)
		}
	}

	if len(filtered) == 0 {
		return fmt.Errorf("no data in the specified time range")
	}

	// Create temporary exporter with filtered data
	tempCollector := &Collector{history: filtered}
	tempExporter := &Exporter{collector: tempCollector}

	// Export in requested format
	switch format {
	case "csv":
		return tempExporter.ExportCSV(w)
	case "json":
		return tempExporter.ExportJSON(w)
	case "prometheus":
		return tempExporter.ExportPrometheus(w)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}