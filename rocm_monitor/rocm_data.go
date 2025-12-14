package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GPU represents a single GPU device
type GPU struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Temperature float64 `json:"temperature"`
	Power       float64 `json:"power"`
	VRAMUsage   float64 `json:"vram_usage"`
	VRAMTotal   float64 `json:"vram_total"`
	GPUUsage    float64 `json:"gpu_usage"`
	FanSpeed    float64 `json:"fan_speed"`
	SCLKFreq    float64 `json:"sclk_freq"`    // System Clock MHz
	MCLKFreq    float64 `json:"mclk_freq"`    // Memory Clock MHz
}

// GPUStaticInfo holds static GPU information
type GPUStaticInfo struct {
	ID             int               `json:"id"`
	ProductName    string            `json:"product_name"`
	VendorName     string            `json:"vendor_name"`
	SerialNumber   string            `json:"serial_number"`
	UniqueID       string            `json:"unique_id"`
	FirmwareInfo   map[string]string `json:"firmware_info"`
	VRAMVendor     string            `json:"vram_vendor"`
	BusInfo        string            `json:"bus_info"`
}

// RocmData represents a monitoring snapshot
type RocmData struct {
	Timestamp time.Time `json:"timestamp"`
	GPUs      []GPU     `json:"gpus"`
	CPUUsage  float64   `json:"cpu_usage"`
}

// Parser handles rocm-smi output parsing
type Parser struct {
	// Pre-compiled regex patterns for better performance
	tempRegex        *regexp.Regexp
	powerRegex       *regexp.Regexp
	vramRegex        *regexp.Regexp
	gpuRegex         *regexp.Regexp
	fanRegex         *regexp.Regexp
	deviceIDRegex    *regexp.Regexp
	vramTotalRegex   *regexp.Regexp
	vramUsedRegex    *regexp.Regexp
	sclkRegex        *regexp.Regexp
	mclkRegex        *regexp.Regexp
}

// NewParser creates a new parser with pre-compiled regex patterns
func NewParser() *Parser {
	return &Parser{
		tempRegex:        regexp.MustCompile(`(\d+\.?\d*)\s*°C`),
		powerRegex:       regexp.MustCompile(`(\d+\.?\d*)\s*W`),
		vramRegex:        regexp.MustCompile(`(\d+\.?\d*)%\s+(\d+\.?\d*)%\s*$`), // VRAM% GPU% at end of line
		gpuRegex:         regexp.MustCompile(`(\d+\.?\d*)%\s*$`),                // GPU% at very end
		fanRegex:         regexp.MustCompile(`(\d+\.?\d*)%\s+auto`),           // Fan% before "auto"
		deviceIDRegex:    regexp.MustCompile(`^(\d+)\s+`),                      // Device ID at start of line
		vramTotalRegex:   regexp.MustCompile(`GPU\[(\d+)\]\s*:\s*VRAM Total Memory \(B\):\s*(\d+)`),
		vramUsedRegex:    regexp.MustCompile(`GPU\[(\d+)\]\s*:\s*VRAM Total Used Memory \(B\):\s*(\d+)`),
		sclkRegex:        regexp.MustCompile(`GPU\[\d+\]\s*:\s*sclk clock level:\s*\d+:\s*\((\d+)Mhz\)`), // SCLK frequency  
		mclkRegex:        regexp.MustCompile(`GPU\[\d+\]\s*:\s*mclk clock level:\s*\d+:\s*\((\d+)Mhz\)`), // MCLK frequency
	}
}

// ParseRocmSMIOutput parses the rocm-smi output and returns structured data
func (p *Parser) ParseRocmSMIOutput(output string) (*RocmData, error) {
	if output == "" {
		return nil, fmt.Errorf("empty rocm-smi output")
	}

	data := &RocmData{
		Timestamp: time.Now(),
		GPUs:      make([]GPU, 0),
	}

	// Parse detailed VRAM information first
	vramTotalMap := make(map[int]float64)
	vramUsedMap := make(map[int]float64)
	
	// Extract VRAM total and used bytes
	totalMatches := p.vramTotalRegex.FindAllStringSubmatch(output, -1)
	for _, match := range totalMatches {
		if len(match) > 2 {
			gpuID, _ := strconv.Atoi(match[1])
			totalBytes, _ := strconv.ParseFloat(match[2], 64)
			vramTotalMap[gpuID] = totalBytes / (1024 * 1024 * 1024) // Convert to GB
		}
	}
	
	usedMatches := p.vramUsedRegex.FindAllStringSubmatch(output, -1)
	for _, match := range usedMatches {
		if len(match) > 2 {
			gpuID, _ := strconv.Atoi(match[1])
			usedBytes, _ := strconv.ParseFloat(match[2], 64)
			vramUsedMap[gpuID] = usedBytes / (1024 * 1024 * 1024) // Convert to GB
		}
	}

	// Split output by GPU sections
	gpuSections := p.splitByGPU(output)
	
	for id, section := range gpuSections {
		gpu := GPU{
			ID: id,
		}

		// Parse temperature
		if matches := p.tempRegex.FindStringSubmatch(section); len(matches) > 1 {
			gpu.Temperature, _ = strconv.ParseFloat(matches[1], 64)
		}

		// Parse power
		if matches := p.powerRegex.FindStringSubmatch(section); len(matches) > 1 {
			gpu.Power, _ = strconv.ParseFloat(matches[1], 64)
		}

		// Parse VRAM% and GPU% from end of line
		if matches := p.vramRegex.FindStringSubmatch(section); len(matches) > 2 {
			gpu.VRAMUsage, _ = strconv.ParseFloat(matches[1], 64)
			gpu.GPUUsage, _ = strconv.ParseFloat(matches[2], 64)
		}

		// Parse fan speed
		if matches := p.fanRegex.FindStringSubmatch(section); len(matches) > 1 {
			gpu.FanSpeed, _ = strconv.ParseFloat(matches[1], 64)
		}

		// Parse SCLK frequency from entire output
		sclkPattern := regexp.MustCompile(fmt.Sprintf(`GPU\[%d\]\s*:\s*sclk clock level:\s*\d+:\s*\((\d+)Mhz\)`, id))
		if matches := sclkPattern.FindStringSubmatch(output); len(matches) > 1 {
			gpu.SCLKFreq, _ = strconv.ParseFloat(matches[1], 64)
		}

		// Parse MCLK frequency from entire output
		mclkPattern := regexp.MustCompile(fmt.Sprintf(`GPU\[%d\]\s*:\s*mclk clock level:\s*\d+:\s*\((\d+)Mhz\)`, id))
		if matches := mclkPattern.FindStringSubmatch(output); len(matches) > 1 {
			gpu.MCLKFreq, _ = strconv.ParseFloat(matches[1], 64)
		}

		// Use detailed VRAM information if available
		if total, exists := vramTotalMap[id]; exists {
			gpu.VRAMTotal = total
		}
		if used, exists := vramUsedMap[id]; exists {
			gpu.VRAMUsage = used
		}

		data.GPUs = append(data.GPUs, gpu)
	}

	if len(data.GPUs) == 0 {
		return nil, fmt.Errorf("no GPU data found in output")
	}

	return data, nil
}

// splitByGPU splits the output into sections per GPU
func (p *Parser) splitByGPU(output string) map[int]string {
	sections := make(map[int]string)
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		// Look for device lines (start with digit and have the GPU data)
		if matches := p.deviceIDRegex.FindStringSubmatch(line); len(matches) > 1 {
			gpuID, _ := strconv.Atoi(matches[1])
			sections[gpuID] = line
		}
	}

	// If no GPU sections found, treat entire output as GPU 0
	if len(sections) == 0 {
		sections[0] = output
	}

	return sections
}

// CPUStats holds CPU timing information
type CPUStats struct {
	Total float64
	Idle  float64
}

var lastCPUStats *CPUStats

// ReadCPUStats reads current CPU stats from /proc/stat
func ReadCPUStats() (*CPUStats, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/stat: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	// Read first line which contains overall CPU stats
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read /proc/stat")
	}
	
	line := scanner.Text()
	if !strings.HasPrefix(line, "cpu ") {
		return nil, fmt.Errorf("invalid /proc/stat format")
	}

	// Parse CPU times: cpu user nice system idle iowait irq softirq steal guest guest_nice
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return nil, fmt.Errorf("insufficient CPU stat fields")
	}

	var total, idle float64
	for i := 1; i < len(fields); i++ {
		val, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CPU stat field %d: %w", i, err)
		}
		total += val
		if i == 4 { // idle time is the 4th field (index 4)
			idle = val
		}
	}

	return &CPUStats{Total: total, Idle: idle}, nil
}

// GetCPUUsage calculates CPU usage percentage between two readings
func GetCPUUsage() (float64, error) {
	currentStats, err := ReadCPUStats()
	if err != nil {
		return 0, err
	}

	// If this is the first reading, store it and return 0
	if lastCPUStats == nil {
		lastCPUStats = currentStats
		return 0, nil
	}

	// Calculate differences
	totalDiff := currentStats.Total - lastCPUStats.Total
	idleDiff := currentStats.Idle - lastCPUStats.Idle

	// Store current stats for next calculation
	lastCPUStats = currentStats

	if totalDiff == 0 {
		return 0, nil
	}

	// CPU usage percentage = (totalDiff - idleDiff) / totalDiff * 100
	usage := (totalDiff - idleDiff) / totalDiff * 100
	return usage, nil
}

// GetGPUStaticInfo retrieves static GPU information
func GetGPUStaticInfo() ([]GPUStaticInfo, error) {
	var gpuInfos []GPUStaticInfo
	
	// Get firmware information
	fwInfo, err := getGPUFirmwareInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get firmware info: %w", err)
	}
	
	// Get product name
	productName, err := getGPUProductName()
	if err != nil {
		productName = "Unknown"
	}
	
	// Get serial number
	serialNumber, err := getGPUSerialNumber()
	if err != nil {
		serialNumber = "Unknown"
	}
	
	// Get unique ID
	uniqueID, err := getGPUUniqueID()
	if err != nil {
		uniqueID = "Unknown"
	}
	
	// Get VRAM vendor
	vramVendor, err := getGPUVRAMVendor()
	if err != nil {
		vramVendor = "Unknown"
	}
	
	// Get bus info
	busInfo, err := getGPUBusInfo()
	if err != nil {
		busInfo = "Unknown"
	}
	
	// For now, assume single GPU (ID 0)
	gpuInfo := GPUStaticInfo{
		ID:             0,
		ProductName:    productName,
		VendorName:     "AMD",
		SerialNumber:   serialNumber,
		UniqueID:       uniqueID,
		FirmwareInfo:   fwInfo,
		VRAMVendor:     vramVendor,
		BusInfo:        busInfo,
	}
	
	gpuInfos = append(gpuInfos, gpuInfo)
	return gpuInfos, nil
}

// Helper functions for getting static GPU information
func getGPUFirmwareInfo() (map[string]string, error) {
	cmd := exec.Command("rocm-smi", "--showfwinfo")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	fwInfo := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.Contains(line, "firmware version") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(strings.ReplaceAll(parts[0], "GPU[0]", ""))
				key = strings.TrimSpace(strings.ReplaceAll(key, "\t", ""))
				value := strings.TrimSpace(parts[1])
				fwInfo[key] = value
			}
		}
	}
	
	return fwInfo, nil
}

func getGPUProductName() (string, error) {
	cmd := exec.Command("rocm-smi", "--showproductname")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown", nil
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "GPU[0]") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				// Check for "Not supported" messages
				if strings.Contains(value, "Not supported") || strings.Contains(value, "get_") {
					return "Not Available", nil
				}
				return value, nil
			}
		}
	}
	
	return "Unknown", nil
}

func getGPUSerialNumber() (string, error) {
	cmd := exec.Command("rocm-smi", "--showserial")
	output, err := cmd.Output()
	if err != nil {
		return "Not Available", nil
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "GPU[0]") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				if strings.Contains(value, "Not supported") || strings.Contains(value, "get_") {
					return "Not Available", nil
				}
				return value, nil
			}
		}
	}
	
	return "Unknown", nil
}

func getGPUUniqueID() (string, error) {
	cmd := exec.Command("rocm-smi", "--showuniqueid")
	output, err := cmd.Output()
	if err != nil {
		return "Not Available", nil
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "GPU[0]") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				if strings.Contains(value, "Not supported") || strings.Contains(value, "get_") {
					return "Not Available", nil
				}
				return value, nil
			}
		}
	}
	
	return "Unknown", nil
}

func getGPUVRAMVendor() (string, error) {
	cmd := exec.Command("rocm-smi", "--showmemvendor")
	output, err := cmd.Output()
	if err != nil {
		return "Not Available", nil
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "GPU[0]") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				if strings.Contains(value, "Not supported") || strings.Contains(value, "get_") {
					return "Not Available", nil
				}
				return value, nil
			}
		}
	}
	
	return "Unknown", nil
}

func getGPUBusInfo() (string, error) {
	cmd := exec.Command("rocm-smi", "--showbus")
	output, err := cmd.Output()
	if err != nil {
		return "Not Available", nil
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "GPU[0]") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				if strings.Contains(value, "Not supported") || strings.Contains(value, "get_") {
					return "Not Available", nil
				}
				return value, nil
			}
		}
	}
	
	return "Unknown", nil
}

// Validate checks if the parsed data is valid
func (d *RocmData) Validate() error {
	if len(d.GPUs) == 0 {
		return fmt.Errorf("no GPU data available")
	}

	for i, gpu := range d.GPUs {
		// Basic sanity checks
		if gpu.Temperature < 0 || gpu.Temperature > 150 {
			return fmt.Errorf("invalid temperature for GPU %d: %.2f", i, gpu.Temperature)
		}
		if gpu.Power < 0 || gpu.Power > 1000 {
			return fmt.Errorf("invalid power for GPU %d: %.2f", i, gpu.Power)
		}
		if gpu.GPUUsage < 0 || gpu.GPUUsage > 100 {
			return fmt.Errorf("invalid GPU usage for GPU %d: %.2f", i, gpu.GPUUsage)
		}
	}

	// Validate CPU usage
	if d.CPUUsage < 0 || d.CPUUsage > 100 {
		return fmt.Errorf("invalid CPU usage: %.2f", d.CPUUsage)
	}

	return nil
}